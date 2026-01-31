package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const docxBlocksDeletePageSize = 200

func newDocsMarkdownCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markdown",
		Short: "Convert or overwrite Docx with Markdown/HTML",
	}

	cmd.AddCommand(newDocsMarkdownConvertCmd(state))
	cmd.AddCommand(newDocsMarkdownOverwriteCmd(state))
	return cmd
}

func newDocsMarkdownConvertCmd(state *appState) *cobra.Command {
	var contentType string
	var content string
	var contentFile string

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert Markdown/HTML to Docx blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			raw, err := readInput(content, contentFile, "content")
			if err != nil {
				return err
			}
			normalized, err := normalizeDocxContentType(contentType)
			if err != nil {
				return err
			}

			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			resp, err := state.SDK.ConvertDocxContent(context.Background(), token, normalized, raw)
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
	cmd.Flags().StringVar(&contentFile, "content-file", "", "path to file containing markdown/html content")
	return cmd
}

func newDocsMarkdownOverwriteCmd(state *appState) *cobra.Command {
	var documentID string
	var contentType string
	var content string
	var contentFile string

	cmd := &cobra.Command{
		Use:   "overwrite <doc-id>",
		Short: "Overwrite a Docx document with Markdown/HTML",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(documentID) == "" {
					return errors.New("doc-id is required")
				}
				return nil
			}
			if documentID != "" && documentID != args[0] {
				return errors.New("doc-id provided twice")
			}
			return cmd.Flags().Set("doc-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			raw, err := readInput(content, contentFile, "content")
			if err != nil {
				return err
			}
			normalized, err := normalizeDocxContentType(contentType)
			if err != nil {
				return err
			}

			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			convertResp, err := state.SDK.ConvertDocxContent(context.Background(), token, normalized, raw)
			if err != nil {
				return err
			}
			if convertResp == nil {
				return errors.New("convert returned empty response")
			}
			scrubDocxTableMergeInfo(convertResp.Blocks)

			deleted, err := clearDocxBlockChildren(context.Background(), state.SDK, token, documentID, documentID)
			if err != nil {
				return err
			}

			createBody := &larkdocx.CreateDocumentBlockDescendantReqBody{
				ChildrenId:  convertResp.FirstLevelBlockIds,
				Descendants: convertResp.Blocks,
			}
			created, err := state.SDK.CreateDocxBlockDescendant(
				context.Background(),
				token,
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

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID (or provide as positional argument)")
	cmd.Flags().StringVar(&contentType, "content-type", "markdown", "content type (markdown|html)")
	cmd.Flags().StringVar(&content, "content", "", "raw markdown/html content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "path to file containing markdown/html content")
	return cmd
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

func scrubDocxTableMergeInfo(blocks []*larkdocx.Block) {
	for _, block := range blocks {
		if block == nil || block.Table == nil || block.Table.Property == nil {
			continue
		}
		block.Table.Property.MergeInfo = nil
	}
}

func clearDocxBlockChildren(ctx context.Context, sdk *larksdk.Client, token, documentID, blockID string) (int, error) {
	if sdk == nil {
		return 0, errors.New("sdk client is required")
	}
	total := 0
	for {
		items, _, _, err := sdk.GetDocxBlockChildren(ctx, token, documentID, blockID, docxBlocksDeletePageSize, "", -1, false, "")
		if err != nil {
			return total, err
		}
		if len(items) == 0 {
			return total, nil
		}
		end := len(items)
		_, err = sdk.BatchDeleteDocxBlockChildren(ctx, token, documentID, blockID, 0, end, -1, "")
		if err != nil {
			return total, err
		}
		total += end
	}
}
