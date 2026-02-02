package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const docxBlocksDeletePageSize = 200

func newDocsConvertCmd(state *appState) *cobra.Command {
	var contentType string
	var content string
	var contentFile string

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert Markdown/HTML to Docx blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			raw, err := readDocxContent(content, contentFile)
			if err != nil {
				return err
			}
			normalized, err := normalizeDocxContentType(contentType)
			if err != nil {
				return err
			}

			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			resp, err := state.SDK.ConvertDocxContent(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), normalized, raw)
			if err != nil {
				return err
			}
			payload := map[string]any{"response": resp}
			if resp == nil {
				return state.Printer.Print(payload, "no blocks converted")
			}
			text := fmt.Sprintf("first_level_blocks: %d\ntotal_blocks: %d", len(resp.FirstLevelBlockIds), len(resp.Blocks))
			if len(resp.BlockIdToImageUrls) > 0 {
				text = fmt.Sprintf("%s\nimage_blocks: %d", text, len(resp.BlockIdToImageUrls))
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&contentType, "content-type", "markdown", "content type (markdown|html)")
	cmd.Flags().StringVar(&content, "content", "", "raw markdown/html content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "path to file containing markdown/html content (or - for stdin)")
	return cmd
}

func newDocsOverwriteCmd(state *appState) *cobra.Command {
	var contentType string
	var content string
	var contentFile string

	cmd := &cobra.Command{
		Use:   "overwrite <document-id>",
		Short: "Overwrite a Docx document with Markdown/HTML",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			refToken, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			documentID := strings.TrimSpace(refToken)
			raw, err := readDocxContent(content, contentFile)
			if err != nil {
				return err
			}
			normalized, err := normalizeDocxContentType(contentType)
			if err != nil {
				return err
			}

			accessToken, accessTokenType, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			convertResp, err := state.SDK.ConvertDocxContent(cmd.Context(), accessToken, larksdk.AccessTokenType(accessTokenType), normalized, raw)
			if err != nil {
				return err
			}
			if convertResp == nil {
				return errors.New("convert returned empty response")
			}
			scrubDocxTableMergeInfo(convertResp.Blocks)

			deleted, err := clearDocxBlockChildren(cmd.Context(), state.SDK, accessToken, larksdk.AccessTokenType(accessTokenType), documentID, documentID)
			if err != nil {
				return err
			}

			createBody := &larkdocx.CreateDocumentBlockDescendantReqBody{
				ChildrenId:  convertResp.FirstLevelBlockIds,
				Descendants: convertResp.Blocks,
			}
			created, err := state.SDK.CreateDocxBlockDescendant(
				cmd.Context(),
				accessToken,
				larksdk.AccessTokenType(accessTokenType),
				documentID,
				documentID,
				createBody,
				-1,
				"",
				"",
			)
			if err != nil {
				return err
			}

			payload := map[string]any{
				"document_id":           documentID,
				"deleted_blocks":        deleted,
				"inserted_blocks":       len(convertResp.Blocks),
				"block_id_relations":    nil,
				"document_revision_id":  nil,
				"image_block_url_map":   convertResp.BlockIdToImageUrls,
				"first_level_block_ids": convertResp.FirstLevelBlockIds,
			}
			revision := ""
			if created != nil {
				payload["block_id_relations"] = created.BlockIdRelations
				if created.DocumentRevisionId != nil {
					payload["document_revision_id"] = *created.DocumentRevisionId
					revision = fmt.Sprintf("%d", *created.DocumentRevisionId)
				}
			}

			text := tableTextRow(
				[]string{"document_id", "deleted_blocks", "inserted_blocks", "revision_id"},
				[]string{documentID, fmt.Sprintf("%d", deleted), fmt.Sprintf("%d", len(convertResp.Blocks)), revision},
			)
			if len(convertResp.BlockIdToImageUrls) > 0 {
				text = fmt.Sprintf("%s\nimage_blocks: %d (upload and replace required)", text, len(convertResp.BlockIdToImageUrls))
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&contentType, "content-type", "markdown", "content type (markdown|html)")
	cmd.Flags().StringVar(&content, "content", "", "raw markdown/html content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "path to file containing markdown/html content")
	return cmd
}

func readDocxContent(raw, path string) (string, error) {
	if path != "" {
		data, err := readInputFile(path)
		if err != nil {
			return "", fmt.Errorf("read content file: %w", err)
		}
		raw = string(data)
	} else {
		raw = normalizeDocxContentEscapes(raw)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("content is required")
	}
	return raw, nil
}

func normalizeDocxContentType(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return larkdocx.ContentTypeMarkdown, nil
	}
	switch value {
	case "markdown", "md":
		return larkdocx.ContentTypeMarkdown, nil
	case "html", "htm":
		return larkdocx.ContentTypeHTML, nil
	default:
		return "", fmt.Errorf("unsupported content type: %s", raw)
	}
}

func normalizeDocxContentEscapes(raw string) string {
	if raw == "" {
		return raw
	}
	if strings.Contains(raw, "\n") || strings.Contains(raw, "\r") {
		return raw
	}
	if !strings.Contains(raw, "\\") {
		return raw
	}
	quoted := `"` + strings.ReplaceAll(raw, `"`, `\"`) + `"`
	unquoted, err := strconv.Unquote(quoted)
	if err != nil {
		return raw
	}
	return unquoted
}

func scrubDocxTableMergeInfo(blocks []*larkdocx.Block) {
	for _, block := range blocks {
		if block == nil || block.Table == nil || block.Table.Property == nil {
			continue
		}
		block.Table.Property.MergeInfo = nil
	}
}

func clearDocxBlockChildren(ctx context.Context, sdk *larksdk.Client, token string, tokenType larksdk.AccessTokenType, documentID, blockID string) (int, error) {
	if sdk == nil {
		return 0, errors.New("sdk client is required")
	}
	total := 0
	for {
		items, _, _, err := sdk.GetDocxBlockChildren(ctx, token, tokenType, documentID, blockID, docxBlocksDeletePageSize, "", -1, false, "")
		if err != nil {
			return total, err
		}
		if len(items) == 0 {
			return total, nil
		}
		end := len(items)
		_, err = sdk.BatchDeleteDocxBlockChildren(ctx, token, tokenType, documentID, blockID, 0, end, -1, "")
		if err != nil {
			return total, err
		}
		total += end
	}
}
