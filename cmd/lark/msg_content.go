package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type messageContentOptions struct {
	MsgType     string
	Text        string
	Post        string
	Content     string
	ContentFile string
	ImageKey    string
	FileKey     string
	MediaKey    string
	UUID        string
}

func addMessageContentFlags(cmd *cobra.Command, opts *messageContentOptions) {
	cmd.Flags().StringVar(&opts.MsgType, "msg-type", "", "message type (text, post, image, file, audio, media, sticker, interactive, share_chat, share_user)")
	cmd.Flags().StringVar(&opts.Text, "text", "", "text content")
	cmd.Flags().StringVar(&opts.Post, "post", "", "post (rich text) JSON content")
	cmd.Flags().StringVar(&opts.Content, "content", "", "raw JSON content for msg_type")
	cmd.Flags().StringVar(&opts.ContentFile, "content-file", "", "path to file containing raw JSON content")
	cmd.Flags().StringVar(&opts.ImageKey, "image-key", "", "image key (msg_type=image)")
	cmd.Flags().StringVar(&opts.FileKey, "file-key", "", "file key (msg_type=file|audio|media)")
	cmd.Flags().StringVar(&opts.MediaKey, "media-key", "", "media key (msg_type=media)")
	cmd.Flags().StringVar(&opts.UUID, "uuid", "", "request UUID for idempotency")
}

func resolveMessageContent(opts messageContentOptions) (msgType string, content string, err error) {
	content = strings.TrimSpace(opts.Content)
	if opts.ContentFile != "" {
		data, readErr := os.ReadFile(opts.ContentFile)
		if readErr != nil {
			return "", "", fmt.Errorf("read content file: %w", readErr)
		}
		content = strings.TrimSpace(string(data))
	}

	sources := 0
	if strings.TrimSpace(opts.Text) != "" {
		sources++
	}
	if strings.TrimSpace(opts.Post) != "" {
		sources++
	}
	if strings.TrimSpace(opts.ImageKey) != "" {
		sources++
	}
	if strings.TrimSpace(opts.FileKey) != "" {
		sources++
	}
	if strings.TrimSpace(opts.MediaKey) != "" {
		sources++
	}
	if strings.TrimSpace(content) != "" {
		sources++
	}
	if sources == 0 {
		return "", "", errors.New("message content is required")
	}
	if sources > 1 {
		return "", "", errors.New("please provide only one of text/post/image-key/file-key/media-key/content")
	}

	msgType = strings.TrimSpace(opts.MsgType)
	switch {
	case strings.TrimSpace(opts.Text) != "":
		if msgType == "" {
			msgType = "text"
		}
		if msgType != "text" {
			return "", "", fmt.Errorf("text requires msg_type=text (got %q)", msgType)
		}
		raw, marshalErr := json.Marshal(map[string]string{"text": strings.TrimSpace(opts.Text)})
		if marshalErr != nil {
			return "", "", marshalErr
		}
		return msgType, string(raw), nil
	case strings.TrimSpace(opts.Post) != "":
		if msgType == "" {
			msgType = "post"
		}
		if msgType != "post" {
			return "", "", fmt.Errorf("post requires msg_type=post (got %q)", msgType)
		}
		return msgType, strings.TrimSpace(opts.Post), nil
	case strings.TrimSpace(opts.ImageKey) != "":
		if msgType == "" {
			msgType = "image"
		}
		if msgType != "image" {
			return "", "", fmt.Errorf("image-key requires msg_type=image (got %q)", msgType)
		}
		raw, marshalErr := json.Marshal(map[string]string{"image_key": strings.TrimSpace(opts.ImageKey)})
		if marshalErr != nil {
			return "", "", marshalErr
		}
		return msgType, string(raw), nil
	case strings.TrimSpace(opts.MediaKey) != "":
		if msgType == "" {
			msgType = "media"
		}
		if msgType != "media" {
			return "", "", fmt.Errorf("media-key requires msg_type=media (got %q)", msgType)
		}
		raw, marshalErr := json.Marshal(map[string]string{"file_key": strings.TrimSpace(opts.MediaKey)})
		if marshalErr != nil {
			return "", "", marshalErr
		}
		return msgType, string(raw), nil
	case strings.TrimSpace(opts.FileKey) != "":
		if msgType == "" {
			msgType = "file"
		}
		switch msgType {
		case "file", "audio", "media":
		default:
			return "", "", fmt.Errorf("file-key requires msg_type=file|audio|media (got %q)", msgType)
		}
		raw, marshalErr := json.Marshal(map[string]string{"file_key": strings.TrimSpace(opts.FileKey)})
		if marshalErr != nil {
			return "", "", marshalErr
		}
		return msgType, string(raw), nil
	default:
		if msgType == "" {
			return "", "", errors.New("msg_type is required when using content")
		}
		if strings.TrimSpace(content) == "" {
			return "", "", errors.New("content is required")
		}
		return msgType, content, nil
	}
}
