package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"

	"lark/internal/larksdk"
)

const docxImageUploadMaxBytes int64 = 20 << 20

type docxImageUploadRecord struct {
	BlockID   string `json:"block_id"`
	ImageURL  string `json:"image_url,omitempty"`
	FileToken string `json:"file_token,omitempty"`
	Error     string `json:"error,omitempty"`
}

type docxImageUploadSummary struct {
	Total    int
	Uploaded int
	Failed   int
	Records  []docxImageUploadRecord
	TokenMap map[string]string
}

type docxImageSource struct {
	Reader   io.Reader
	Size     int64
	FileName string
	CloseFn  func() error
}

func (s docxImageSource) Close() {
	if s.CloseFn != nil {
		_ = s.CloseFn()
	}
}

func uploadDocxImageBlocks(ctx context.Context, sdk *larksdk.Client, token, documentID, contentFile string, resp *larkdocx.ConvertDocumentRespData) (docxImageUploadSummary, error) {
	summary := docxImageUploadSummary{}
	if sdk == nil {
		return summary, errors.New("sdk client is required")
	}
	if resp == nil || len(resp.BlockIdToImageUrls) == 0 {
		return summary, nil
	}
	if strings.TrimSpace(documentID) == "" {
		return summary, errors.New("document id is required")
	}

	blocksByID := make(map[string]*larkdocx.Block, len(resp.Blocks))
	for _, block := range resp.Blocks {
		id := docxBlockID(block)
		if id == "" {
			continue
		}
		blocksByID[id] = block
	}

	baseDir := ""
	if strings.TrimSpace(contentFile) != "" && strings.TrimSpace(contentFile) != "-" {
		baseDir = filepath.Dir(contentFile)
	}

	for _, entry := range resp.BlockIdToImageUrls {
		rec := docxImageUploadRecord{}
		if entry != nil {
			if entry.BlockId != nil {
				rec.BlockID = strings.TrimSpace(*entry.BlockId)
			}
			if entry.ImageUrl != nil {
				rec.ImageURL = strings.TrimSpace(*entry.ImageUrl)
			}
		}
		summary.Total++

		if rec.BlockID == "" {
			rec.Error = "missing block_id"
			summary.Records = append(summary.Records, rec)
			continue
		}
		if rec.ImageURL == "" {
			rec.Error = "missing image_url"
			summary.Records = append(summary.Records, rec)
			continue
		}
		block := blocksByID[rec.BlockID]
		if block == nil {
			rec.Error = "block not found"
			summary.Records = append(summary.Records, rec)
			continue
		}

		source, err := openDocxImageSource(ctx, rec.ImageURL, baseDir)
		if err != nil {
			rec.Error = err.Error()
			summary.Records = append(summary.Records, rec)
			continue
		}

		fileToken, err := uploadDocxImageSource(ctx, sdk, token, documentID, source)
		source.Close()
		if err != nil {
			rec.Error = err.Error()
			summary.Records = append(summary.Records, rec)
			continue
		}

		if block.Image == nil {
			block.Image = &larkdocx.Image{}
		}
		block.Image.Token = &fileToken
		rec.FileToken = fileToken
		summary.Uploaded++
		if summary.TokenMap == nil {
			summary.TokenMap = make(map[string]string)
		}
		summary.TokenMap[rec.BlockID] = fileToken
		summary.Records = append(summary.Records, rec)
	}

	summary.Failed = summary.Total - summary.Uploaded
	return summary, nil
}

func uploadDocxImageSource(ctx context.Context, sdk *larksdk.Client, token, documentID string, source docxImageSource) (string, error) {
	if sdk == nil {
		return "", errors.New("sdk client is required")
	}
	if source.Reader == nil {
		return "", errors.New("image source is required")
	}
	if strings.TrimSpace(source.FileName) == "" {
		return "", errors.New("image file name is required")
	}
	if source.Size <= 0 {
		return "", errors.New("image size is required")
	}
	if source.Size > docxImageUploadMaxBytes {
		return "", fmt.Errorf("image exceeds %d bytes", docxImageUploadMaxBytes)
	}

	result, err := sdk.UploadDriveMedia(ctx, token, larksdk.UploadDriveMediaRequest{
		FileName:   source.FileName,
		ParentType: larkdrive.ParentTypeUploadAllMediaDocxImage,
		ParentNode: documentID,
		Size:       source.Size,
		File:       source.Reader,
	})
	if err != nil {
		return "", err
	}
	return result.FileToken, nil
}

func openDocxImageSource(ctx context.Context, imageURL, baseDir string) (docxImageSource, error) {
	trimmed := strings.TrimSpace(imageURL)
	if trimmed == "" {
		return docxImageSource{}, errors.New("image url is required")
	}
	if strings.HasPrefix(trimmed, "data:") {
		return decodeDocxImageDataURI(trimmed)
	}

	parsed, err := url.Parse(trimmed)
	if err == nil && parsed.Scheme != "" {
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https":
			return fetchDocxImageURL(ctx, parsed)
		case "file":
			return openDocxImageFile(pathFromFileURL(parsed))
		default:
			return docxImageSource{}, fmt.Errorf("unsupported image url scheme: %s", parsed.Scheme)
		}
	}

	localPath := trimmed
	if baseDir != "" && !filepath.IsAbs(localPath) {
		localPath = filepath.Join(baseDir, localPath)
	}
	return openDocxImageFile(localPath)
}

func decodeDocxImageDataURI(raw string) (docxImageSource, error) {
	comma := strings.Index(raw, ",")
	if comma < 0 {
		return docxImageSource{}, errors.New("invalid data uri")
	}
	meta := raw[len("data:"):comma]
	payload := raw[comma+1:]
	mediaType := ""
	isBase64 := false
	if meta != "" {
		parts := strings.Split(meta, ";")
		mediaType = strings.TrimSpace(parts[0])
		for _, part := range parts[1:] {
			if strings.EqualFold(strings.TrimSpace(part), "base64") {
				isBase64 = true
				break
			}
		}
	}
	var data []byte
	if isBase64 {
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return docxImageSource{}, fmt.Errorf("decode base64 data uri: %w", err)
		}
		data = decoded
	} else {
		unescaped, err := url.PathUnescape(payload)
		if err != nil {
			return docxImageSource{}, fmt.Errorf("decode data uri: %w", err)
		}
		data = []byte(unescaped)
	}
	if int64(len(data)) > docxImageUploadMaxBytes {
		return docxImageSource{}, fmt.Errorf("image exceeds %d bytes", docxImageUploadMaxBytes)
	}
	name := "image"
	if ext := imageExtFromContentType(mediaType); ext != "" {
		name += ext
	}
	return docxImageSource{Reader: bytes.NewReader(data), Size: int64(len(data)), FileName: name}, nil
}

func fetchDocxImageURL(ctx context.Context, parsed *url.URL) (docxImageSource, error) {
	if parsed == nil || parsed.String() == "" {
		return docxImageSource{}, errors.New("image url is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return docxImageSource{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return docxImageSource{}, err
	}
	if resp.Body == nil {
		resp.Body.Close()
		return docxImageSource{}, errors.New("empty image response body")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		resp.Body.Close()
		return docxImageSource{}, fmt.Errorf("image url returned %s", resp.Status)
	}

	name := fileNameFromURL(parsed, resp.Header.Get("Content-Type"))
	size := resp.ContentLength
	if size > docxImageUploadMaxBytes {
		resp.Body.Close()
		return docxImageSource{}, fmt.Errorf("image exceeds %d bytes", docxImageUploadMaxBytes)
	}
	if size < 0 {
		data, err := io.ReadAll(io.LimitReader(resp.Body, docxImageUploadMaxBytes+1))
		resp.Body.Close()
		if err != nil {
			return docxImageSource{}, err
		}
		if int64(len(data)) > docxImageUploadMaxBytes {
			return docxImageSource{}, fmt.Errorf("image exceeds %d bytes", docxImageUploadMaxBytes)
		}
		return docxImageSource{Reader: bytes.NewReader(data), Size: int64(len(data)), FileName: name}, nil
	}

	return docxImageSource{Reader: resp.Body, Size: size, FileName: name, CloseFn: resp.Body.Close}, nil
}

func openDocxImageFile(localPath string) (docxImageSource, error) {
	trimmed := strings.TrimSpace(localPath)
	if trimmed == "" {
		return docxImageSource{}, errors.New("image path is required")
	}
	file, err := os.Open(trimmed)
	if err != nil {
		return docxImageSource{}, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return docxImageSource{}, err
	}
	size := info.Size()
	if size > docxImageUploadMaxBytes {
		file.Close()
		return docxImageSource{}, fmt.Errorf("image exceeds %d bytes", docxImageUploadMaxBytes)
	}
	name := info.Name()
	if strings.TrimSpace(name) == "" {
		name = "image"
	}
	return docxImageSource{Reader: file, Size: size, FileName: name, CloseFn: file.Close}, nil
}

func fileNameFromURL(parsed *url.URL, contentType string) string {
	name := ""
	if parsed != nil {
		name = path.Base(parsed.Path)
		if name == "." || name == "/" {
			name = ""
		}
	}
	return normalizeImageFileName(name, contentType)
}

func normalizeImageFileName(name, contentType string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "image"
	}
	if filepath.Ext(name) == "" {
		if ext := imageExtFromContentType(contentType); ext != "" {
			name += ext
		}
	}
	return name
}

func imageExtFromContentType(contentType string) string {
	base := strings.ToLower(strings.TrimSpace(contentType))
	if base == "" {
		return ""
	}
	if idx := strings.Index(base, ";"); idx >= 0 {
		base = strings.TrimSpace(base[:idx])
	}
	switch base {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/bmp":
		return ".bmp"
	case "image/tiff":
		return ".tiff"
	default:
		return ""
	}
}

func pathFromFileURL(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	if parsed.Host != "" && parsed.Host != "localhost" {
		return "//" + parsed.Host + parsed.Path
	}
	value := parsed.Path
	if unescaped, err := url.PathUnescape(value); err == nil {
		return unescaped
	}
	return value
}
