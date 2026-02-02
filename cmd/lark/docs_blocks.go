package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const docxBlocksMaxPageSize = 200

func newDocsBlocksCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "Manage Docx blocks",
	}

	cmd.AddCommand(newDocsBlocksGetCmd(state))
	cmd.AddCommand(newDocsBlocksListCmd(state))
	cmd.AddCommand(newDocsBlocksUpdateCmd(state))
	cmd.AddCommand(newDocsBlocksBatchUpdateCmd(state))
	cmd.AddCommand(newDocsBlocksChildrenCmd(state))
	cmd.AddCommand(newDocsBlocksDescendantCmd(state))
	return cmd
}

func newDocsBlocksGetCmd(state *appState) *cobra.Command {
	var revisionID int
	var userIDType string

	cmd := &cobra.Command{
		Use:   "get <document-id> <block-id>",
		Short: "Get a Docx block",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			if strings.TrimSpace(args[1]) == "" {
				return argsUsageError(cmd, errors.New("block-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			documentID := strings.TrimSpace(args[0])
			blockID := strings.TrimSpace(args[1])
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			block, err := state.SDK.GetDocxBlock(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), documentID, blockID, revisionID, userIDType)
			if err != nil {
				return err
			}
			payload := map[string]any{"block": block}
			if block == nil {
				return state.Printer.Print(payload, "no block found")
			}
			text := tableTextRow(
				[]string{"block_id", "block_type", "text"},
				docxBlockRow(block),
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func newDocsBlocksListCmd(state *appState) *cobra.Command {
	var limit int
	var pageSize int
	var revisionID int
	var userIDType string

	cmd := &cobra.Command{
		Use:   "list <document-id>",
		Short: "List Docx blocks",
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
			if limit <= 0 {
				return flagUsage(cmd, "limit must be greater than 0")
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			documentID := strings.TrimSpace(args[0])
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			if pageSize <= 0 {
				if limit < docxBlocksMaxPageSize {
					pageSize = limit
				} else {
					pageSize = docxBlocksMaxPageSize
				}
			}

			blocks := make([]*larkdocx.Block, 0, limit)
			pageToken := ""
			for {
				items, nextToken, hasMore, err := state.SDK.ListDocxBlocks(
					cmd.Context(),
					token,
					larksdk.AccessTokenType(tokenTypeValue),
					documentID,
					pageSize,
					pageToken,
					revisionID,
					userIDType,
				)
				if err != nil {
					return err
				}
				blocks = append(blocks, items...)
				if len(blocks) >= limit {
					break
				}
				if !hasMore || nextToken == "" {
					break
				}
				pageToken = nextToken
			}
			if len(blocks) > limit {
				blocks = blocks[:limit]
			}
			payload := map[string]any{"blocks": blocks}
			text := docxBlocksTable(blocks, "no blocks found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 200, "max number of blocks to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size for list requests")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func newDocsBlocksUpdateCmd(state *appState) *cobra.Command {
	var bodyJSON string
	var bodyFile string
	var revisionID int
	var clientToken string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "update <document-id> <block-id>",
		Short: "Update a Docx block",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			if strings.TrimSpace(args[1]) == "" {
				return argsUsageError(cmd, errors.New("block-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			documentID := strings.TrimSpace(args[0])
			blockID := strings.TrimSpace(args[1])
			raw, err := readInput(bodyJSON, bodyFile, "body")
			if err != nil {
				return err
			}
			var update larkdocx.UpdateBlockRequest
			if err := json.Unmarshal([]byte(raw), &update); err != nil {
				return fmt.Errorf("body must be valid JSON: %w", err)
			}

			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			resp, err := state.SDK.PatchDocxBlock(
				cmd.Context(),
				token,
				larksdk.AccessTokenType(tokenTypeValue),
				documentID,
				blockID,
				&update,
				revisionID,
				clientToken,
				userIDType,
			)
			if err != nil {
				return err
			}

			payload := map[string]any{"response": resp}
			if resp == nil || resp.Block == nil {
				return state.Printer.Print(payload, "no block updated")
			}
			revision := ""
			if resp.DocumentRevisionId != nil {
				revision = fmt.Sprintf("%d", *resp.DocumentRevisionId)
			}
			row := append(docxBlockRow(resp.Block), revision)
			text := tableTextRow(
				[]string{"block_id", "block_type", "text", "revision_id"},
				row,
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&bodyJSON, "body-json", "", "JSON body for update request")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "path to file containing JSON body (or - for stdin)")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func newDocsBlocksBatchUpdateCmd(state *appState) *cobra.Command {
	var requestsJSON string
	var requestsFile string
	var revisionID int
	var clientToken string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "batch-update <document-id>",
		Short: "Batch update Docx blocks",
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
			documentID := strings.TrimSpace(args[0])
			raw, err := readInput(requestsJSON, requestsFile, "requests")
			if err != nil {
				return err
			}
			var requests []*larkdocx.UpdateBlockRequest
			if err := json.Unmarshal([]byte(raw), &requests); err != nil {
				return fmt.Errorf("requests must be a JSON array: %w", err)
			}
			if len(requests) == 0 {
				return usageErrorWithUsage(cmd, "requests must be a non-empty JSON array", "", cmd.UsageString())
			}

			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			resp, err := state.SDK.BatchUpdateDocxBlocks(
				cmd.Context(),
				token,
				larksdk.AccessTokenType(tokenTypeValue),
				documentID,
				requests,
				revisionID,
				clientToken,
				userIDType,
			)
			if err != nil {
				return err
			}

			payload := map[string]any{"response": resp}
			blocks := []*larkdocx.Block{}
			if resp != nil {
				blocks = resp.Blocks
			}
			text := docxBlocksTable(blocks, "no blocks updated")
			if resp != nil && resp.DocumentRevisionId != nil {
				text = fmt.Sprintf("revision_id: %d\n%s", *resp.DocumentRevisionId, text)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&requestsJSON, "requests-json", "", "JSON array of update requests")
	cmd.Flags().StringVar(&requestsFile, "requests-file", "", "path to file containing JSON array of update requests")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func newDocsBlocksChildrenCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "children",
		Short: "Manage Docx block children",
	}
	cmd.AddCommand(newDocsBlocksChildrenListCmd(state))
	cmd.AddCommand(newDocsBlocksChildrenCreateCmd(state))
	cmd.AddCommand(newDocsBlocksChildrenDeleteCmd(state))
	return cmd
}

func newDocsBlocksChildrenListCmd(state *appState) *cobra.Command {
	var limit int
	var pageSize int
	var revisionID int
	var withDescendants bool
	var userIDType string

	cmd := &cobra.Command{
		Use:   "list <document-id> <block-id>",
		Short: "List children of a Docx block",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			if strings.TrimSpace(args[1]) == "" {
				return argsUsageError(cmd, errors.New("block-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return flagUsage(cmd, "limit must be greater than 0")
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			documentID := strings.TrimSpace(args[0])
			blockID := strings.TrimSpace(args[1])
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			if pageSize <= 0 {
				if limit < docxBlocksMaxPageSize {
					pageSize = limit
				} else {
					pageSize = docxBlocksMaxPageSize
				}
			}

			blocks := make([]*larkdocx.Block, 0, limit)
			pageToken := ""
			for {
				items, nextToken, hasMore, err := state.SDK.GetDocxBlockChildren(
					cmd.Context(),
					token,
					larksdk.AccessTokenType(tokenTypeValue),
					documentID,
					blockID,
					pageSize,
					pageToken,
					revisionID,
					withDescendants,
					userIDType,
				)
				if err != nil {
					return err
				}
				blocks = append(blocks, items...)
				if len(blocks) >= limit {
					break
				}
				if !hasMore || nextToken == "" {
					break
				}
				pageToken = nextToken
			}
			if len(blocks) > limit {
				blocks = blocks[:limit]
			}

			payload := map[string]any{"blocks": blocks}
			text := docxBlocksTable(blocks, "no blocks found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 200, "max number of blocks to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size for list requests")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().BoolVar(&withDescendants, "with-descendants", false, "include descendant blocks")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func newDocsBlocksChildrenCreateCmd(state *appState) *cobra.Command {
	var bodyJSON string
	var bodyFile string
	var revisionID int
	var clientToken string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "create <document-id> <block-id>",
		Short: "Create children blocks",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			if strings.TrimSpace(args[1]) == "" {
				return argsUsageError(cmd, errors.New("block-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			documentID := strings.TrimSpace(args[0])
			blockID := strings.TrimSpace(args[1])
			raw, err := readInput(bodyJSON, bodyFile, "body")
			if err != nil {
				return err
			}
			var body larkdocx.CreateDocumentBlockChildrenReqBody
			if err := json.Unmarshal([]byte(raw), &body); err != nil {
				return fmt.Errorf("body must be valid JSON: %w", err)
			}

			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			resp, err := state.SDK.CreateDocxBlockChildren(
				cmd.Context(),
				token,
				larksdk.AccessTokenType(tokenTypeValue),
				documentID,
				blockID,
				&body,
				revisionID,
				clientToken,
				userIDType,
			)
			if err != nil {
				return err
			}
			payload := map[string]any{"response": resp}
			blocks := []*larkdocx.Block{}
			if resp != nil {
				blocks = resp.Children
			}
			text := docxBlocksTable(blocks, "no blocks created")
			if resp != nil && resp.DocumentRevisionId != nil {
				text = fmt.Sprintf("revision_id: %d\n%s", *resp.DocumentRevisionId, text)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&bodyJSON, "body-json", "", "JSON body for create request")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "path to file containing JSON body")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func newDocsBlocksChildrenDeleteCmd(state *appState) *cobra.Command {
	var startIndex int
	var endIndex int
	var revisionID int
	var clientToken string

	cmd := &cobra.Command{
		Use:   "delete <document-id> <block-id>",
		Short: "Delete children blocks by index range",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			if strings.TrimSpace(args[1]) == "" {
				return argsUsageError(cmd, errors.New("block-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if startIndex < 0 || endIndex < 0 {
				return usageErrorWithUsage(cmd, "start-index and end-index must be >= 0", "", cmd.UsageString())
			}
			if endIndex <= startIndex {
				return usageErrorWithUsage(cmd, "end-index must be greater than start-index", "", cmd.UsageString())
			}
			documentID := strings.TrimSpace(args[0])
			blockID := strings.TrimSpace(args[1])
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete block children for %s/%s", documentID, blockID)); err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			resp, err := state.SDK.BatchDeleteDocxBlockChildren(
				cmd.Context(),
				token,
				larksdk.AccessTokenType(tokenTypeValue),
				documentID,
				blockID,
				startIndex,
				endIndex,
				revisionID,
				clientToken,
			)
			if err != nil {
				return err
			}
			payload := map[string]any{"response": resp}
			revision := ""
			if resp != nil && resp.DocumentRevisionId != nil {
				revision = fmt.Sprintf("%d", *resp.DocumentRevisionId)
			}
			text := tableTextRow(
				[]string{"document_id", "block_id", "start_index", "end_index", "revision_id"},
				[]string{documentID, blockID, fmt.Sprintf("%d", startIndex), fmt.Sprintf("%d", endIndex), revision},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&startIndex, "start-index", -1, "start index (inclusive)")
	cmd.Flags().IntVar(&endIndex, "end-index", -1, "end index (exclusive)")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	return cmd
}

func newDocsBlocksDescendantCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "descendant",
		Short: "Manage Docx descendant blocks",
	}
	cmd.AddCommand(newDocsBlocksDescendantCreateCmd(state))
	return cmd
}

func newDocsBlocksDescendantCreateCmd(state *appState) *cobra.Command {
	var bodyJSON string
	var bodyFile string
	var revisionID int
	var clientToken string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "create <document-id> <block-id>",
		Short: "Create nested blocks",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("document-id is required"))
			}
			if strings.TrimSpace(args[1]) == "" {
				return argsUsageError(cmd, errors.New("block-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			documentID := strings.TrimSpace(args[0])
			blockID := strings.TrimSpace(args[1])
			raw, err := readInput(bodyJSON, bodyFile, "body")
			if err != nil {
				return err
			}
			var body larkdocx.CreateDocumentBlockDescendantReqBody
			if err := json.Unmarshal([]byte(raw), &body); err != nil {
				return fmt.Errorf("body must be valid JSON: %w", err)
			}

			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			resp, err := state.SDK.CreateDocxBlockDescendant(
				cmd.Context(),
				token,
				larksdk.AccessTokenType(tokenTypeValue),
				documentID,
				blockID,
				&body,
				revisionID,
				clientToken,
				userIDType,
			)
			if err != nil {
				return err
			}
			payload := map[string]any{"response": resp}
			rows := docxBlockRelationRows(nil)
			if resp != nil {
				rows = docxBlockRelationRows(resp.BlockIdRelations)
			}
			text := tableTextFromRows([]string{"temporary_block_id", "block_id"}, rows, "no block relations")
			if resp != nil && resp.DocumentRevisionId != nil {
				text = fmt.Sprintf("revision_id: %d\n%s", *resp.DocumentRevisionId, text)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&bodyJSON, "body-json", "", "JSON body for create request")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "path to file containing JSON body")
	cmd.Flags().IntVar(&revisionID, "revision-id", -1, "document revision id (-1 for latest)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	return cmd
}

func readInput(raw, path, label string) (string, error) {
	if path != "" {
		data, err := readInputFile(path)
		if err != nil {
			return "", fmt.Errorf("read %s file: %w", label, err)
		}
		raw = string(data)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%s is required", label)
	}
	return raw, nil
}

func docxBlockRow(block *larkdocx.Block) []string {
	return []string{docxBlockID(block), docxBlockType(block), strings.TrimSpace(docxBlockText(block))}
}

func docxBlocksTable(blocks []*larkdocx.Block, emptyText string) string {
	rows := make([][]string, 0, len(blocks))
	for _, block := range blocks {
		if block == nil {
			continue
		}
		rows = append(rows, docxBlockRow(block))
	}
	return tableTextFromRows([]string{"block_id", "block_type", "text"}, rows, emptyText)
}

func docxBlockID(block *larkdocx.Block) string {
	if block == nil || block.BlockId == nil {
		return ""
	}
	return *block.BlockId
}

func docxBlockType(block *larkdocx.Block) string {
	if block == nil || block.BlockType == nil {
		return ""
	}
	return fmt.Sprintf("%d", *block.BlockType)
}

func docxBlockRelationRows(relations []*larkdocx.BlockIdRelation) [][]string {
	if len(relations) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(relations))
	for _, rel := range relations {
		if rel == nil {
			continue
		}
		temp := ""
		if rel.TemporaryBlockId != nil {
			temp = *rel.TemporaryBlockId
		}
		id := ""
		if rel.BlockId != nil {
			id = *rel.BlockId
		}
		rows = append(rows, []string{temp, id})
	}
	return rows
}
