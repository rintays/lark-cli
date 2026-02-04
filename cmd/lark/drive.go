package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxDrivePageSize = 200

func newDriveCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drive",
		Short: "Manage Drive files",
		Long: `Drive is the file storage for docs, sheets, slides, and files.

- Files and folders are identified by tokens (file_token, folder_token).
- A folder contains files; default folder-id is the root.
- File types include docx, sheet, slide, mindnote, and file.`,
	}
	annotateAuthServices(cmd, "drive")
	cmd.AddCommand(newDriveListCmd(state))
	cmd.AddCommand(newDriveSearchCmd(state))
	cmd.AddCommand(newDriveInfoCmd(state))
	cmd.AddCommand(newDriveExportCmd(state))
	cmd.AddCommand(newDriveDownloadCmd(state))
	cmd.AddCommand(newDriveUploadCmd(state))
	cmd.AddCommand(newDriveURLsCmd(state))
	cmd.AddCommand(newDriveShareCmd(state))
	cmd.AddCommand(newDrivePermissionsCmd(state))
	return cmd
}

func newDriveListCmd(state *appState) *cobra.Command {
	var folderID string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files in a Drive folder",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return flagUsage(cmd, "limit must be greater than 0")
			}
			folderID = strings.TrimSpace(folderID)
			if strings.EqualFold(folderID, "root") {
				folderID = "0"
			}
			if strings.EqualFold(folderID, "root") {
				folderID = "0"
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			files := make([]larksdk.DriveFile, 0, limit)
			pageToken := ""
			nextPageToken := ""
			hasMore := false
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxDrivePageSize {
					pageSize = maxDrivePageSize
				}
				result, err := state.SDK.ListDriveFiles(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.ListDriveFilesRequest{
					FolderToken: folderID,
					PageSize:    pageSize,
					PageToken:   pageToken,
				})
				if err != nil {
					return err
				}
				nextPageToken = result.PageToken
				hasMore = result.HasMore
				files = append(files, result.Files...)
				if len(files) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(files)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(files) > limit {
				files = files[:limit]
			}
			payload := map[string]any{
				"files":           files,
				"has_more":        hasMore,
				"next_page_token": nextPageToken,
			}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL))
			}
			text := tableText([]string{"token", "name", "type", "url"}, lines, "no files found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")
	return cmd
}

func newDriveSearchCmd(state *appState) *cobra.Command {
	var query string
	var fileTypes []string
	var folderID string
	var limit int
	var pages int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Drive files by text",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			// Allow multi-word queries without requiring shell quoting.
			query = strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return argsUsageError(cmd, errors.New("query is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return flagUsage(cmd, "limit must be greater than 0")
			}
			if pages <= 0 {
				return flagUsage(cmd, "pages must be greater than 0")
			}
			ctx := cmd.Context()
			token, err := resolveDriveSearchToken(ctx, state)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}

			folderID = strings.TrimSpace(folderID)
			var files []larksdk.DriveFile
			if folderID != "" {
				files, err = searchDriveFilesInFolder(ctx, state, token, query, fileTypes, folderID, limit, pages)
			} else {
				files, err = docsSearchDriveFiles(ctx, state, token, "drive", query, fileTypes, limit, pages)
			}
			if err != nil {
				return err
			}

			payload := map[string]any{"files": files}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL))
			}
			text := tableText([]string{"token", "name", "type", "url"}, lines, "no files found")
			return state.Printer.Print(payload, text)
		},
	}
	annotateAuthServices(cmd, "drive", "search-docs")

	cmd.Flags().StringSliceVar(&fileTypes, "type", nil, "filter by doc type (docx|doc|sheet|slides|bitable|mindnote|file); repeatable or comma-separated")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token to scope the search")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")
	return cmd
}

func newDriveInfoCmd(state *appState) *cobra.Command {
	var fileToken string

	cmd := &cobra.Command{
		Use:   "info <file-token>",
		Short: "Show Drive file info",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(token)
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			fetcher, err := newDriveMetadataFetcher(ctx, state)
			if err != nil {
				return err
			}
			file, err := fetcher.get(ctx, fileToken)
			if err != nil {
				return err
			}
			payload := map[string]any{"file": file}
			text := tableTextRow(
				[]string{"token", "name", "type", "url"},
				[]string{file.Token, file.Name, file.FileType, file.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	return cmd
}

func newDriveUploadCmd(state *appState) *cobra.Command {
	var filePath string
	var folderToken string
	var uploadName string

	cmd := &cobra.Command{
		Use:   "upload <path>",
		Short: "Upload a local file to Drive",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) == 0 {
				if strings.TrimSpace(filePath) == "" {
					return argsUsageError(cmd, errors.New("file is required"))
				}
				return nil
			}
			if filePath != "" && filePath != args[0] {
				return argsUsageError(cmd, errors.New("file provided twice"))
			}
			return cmd.Flags().Set("file", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := os.Stat(filePath)
			if err != nil {
				return err
			}
			if info.IsDir() {
				return fmt.Errorf("file path is a directory: %s", filePath)
			}
			if uploadName == "" {
				uploadName = filepath.Base(filePath)
			}
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			if folderToken == "" || folderToken == "root" {
				// Lark/Feishu Drive root folder token is "0".
				folderToken = "0"
			}
			result, err := state.SDK.UploadDriveFile(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), larksdk.UploadDriveFileRequest{
				FileName:    uploadName,
				FolderToken: folderToken,
				Size:        info.Size(),
				File:        file,
			})
			if err != nil {
				return err
			}
			fileToken := result.FileToken
			fileInfo := result.File
			if fileInfo.Token == "" {
				fileInfo.Token = fileToken
			}
			if fileInfo.Token == "" {
				return errors.New("upload response missing file token")
			}
			if fileInfo.Name == "" || fileInfo.FileType == "" || fileInfo.URL == "" {
				meta, err := driveFileMetadataWithToken(cmd.Context(), state.SDK, tokenTypeValue, token, fileInfo.Token)
				if err != nil {
					return err
				}
				fileInfo = meta
			}
			payload := map[string]any{
				"file_token": fileInfo.Token,
				"file":       fileInfo,
			}
			text := tableTextRow(
				[]string{"token", "name", "type", "url"},
				[]string{fileInfo.Token, fileInfo.Name, fileInfo.FileType, fileInfo.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "path to local file (or provide as positional argument)")
	cmd.Flags().StringVar(&folderToken, "folder-id", "", "Drive folder token (default: root)")
	cmd.Flags().StringVar(&folderToken, "folder-token", "", "Drive folder token (deprecated; use --folder-id)")
	_ = cmd.Flags().MarkDeprecated("folder-token", "use --folder-id")
	cmd.Flags().StringVar(&uploadName, "name", "", "override the uploaded file name")
	return cmd
}

func newDriveDownloadCmd(state *appState) *cobra.Command {
	var fileToken string
	var outPath string

	cmd := &cobra.Command{
		Use:   "download <file-token> --out <path>",
		Short: "Download a Drive file",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(token)
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outPath = strings.TrimSpace(outPath)
			writeStdout := outPath == "-"
			outDir := ""
			if !writeStdout {
				if info, err := os.Stat(outPath); err == nil && info.IsDir() {
					outDir = outPath
				}
			}
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			download, err := state.SDK.DownloadDriveFile(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), fileToken)
			if err != nil {
				return err
			}
			defer download.Reader.Close()
			if !writeStdout && outDir != "" {
				fileName := strings.TrimSpace(download.FileName)
				if fileName == "" {
					req := larksdk.GetDriveFileRequest{FileToken: fileToken}
					var meta larksdk.DriveFile
					if tokenTypeValue == tokenTypeUser {
						meta, err = state.SDK.GetDriveFileMetadataWithUserToken(cmd.Context(), token, req)
					} else {
						meta, err = state.SDK.GetDriveFileMetadata(cmd.Context(), token, req)
					}
					if err == nil {
						fileName = strings.TrimSpace(meta.Name)
					}
				}
				fileName = strings.TrimSpace(fileName)
				if fileName != "" {
					fileName = filepath.Base(fileName)
				}
				if fileName == "" || fileName == "." || fileName == string(filepath.Separator) {
					fileName = fileToken
				}
				outPath = filepath.Join(outDir, fileName)
			}
			var out io.Writer
			var outFile *os.File
			if writeStdout {
				out = cmd.OutOrStdout()
			} else {
				outFile, err = os.Create(outPath)
				if err != nil {
					return err
				}
				defer outFile.Close()
				out = outFile
			}
			written, err := io.Copy(out, download.Reader)
			if err != nil {
				return err
			}
			if writeStdout {
				if state.Verbose {
					fmt.Fprintf(errWriter(state), "wrote %d bytes to stdout\n", written)
				}
				return nil
			}
			payload := map[string]any{
				"file_token":    fileToken,
				"output_path":   outPath,
				"bytes_written": written,
			}
			text := tableTextRow(
				[]string{"file_token", "output_path", "bytes_written"},
				[]string{fileToken, outPath, fmt.Sprintf("%d", written)},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&outPath, "out", "", "output file path or directory (or - for stdout)")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDriveExportCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var format string
	var outPath string

	cmd := &cobra.Command{
		Use:     "export <file-token> --type <type> --format <format> --out <path>",
		Short:   "Export a Drive file",
		Example: "  lark drive export <file-token> --type docx --format pdf --out ./doc.pdf",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, kind, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(token)
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			if kind != "" && !cmd.Flags().Changed("type") {
				_ = cmd.Flags().Set("type", kind)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outPath = strings.TrimSpace(outPath)
			writeStdout := outPath == "-"
			if !writeStdout {
				if info, err := os.Stat(outPath); err == nil && info.IsDir() {
					return fmt.Errorf("output path is a directory: %s", outPath)
				}
			}
			format = strings.ToLower(format)
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			ticket, err := state.SDK.CreateExportTask(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), larksdk.CreateExportTaskRequest{
				Token:         fileToken,
				Type:          fileType,
				FileExtension: format,
			})
			if err != nil {
				return err
			}
			result, err := pollExportTask(cmd.Context(), state.SDK, token, larksdk.AccessTokenType(tokenTypeValue), ticket, fileToken)
			if err != nil {
				return err
			}
			reader, err := state.SDK.DownloadExportedFile(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), result.FileToken)
			if err != nil {
				return err
			}
			defer reader.Close()
			var out io.Writer
			var outFile *os.File
			if writeStdout {
				out = cmd.OutOrStdout()
			} else {
				outFile, err = os.Create(outPath)
				if err != nil {
					return err
				}
				defer outFile.Close()
				out = outFile
			}
			written, err := io.Copy(out, reader)
			if err != nil {
				return err
			}
			if writeStdout {
				if state.Verbose {
					fmt.Fprintf(errWriter(state), "wrote %d bytes to stdout\n", written)
				}
				return nil
			}
			payload := map[string]any{
				"file_token":        fileToken,
				"type":              fileType,
				"format":            format,
				"export_file_token": result.FileToken,
				"file_name":         result.FileName,
				"output_path":       outPath,
				"bytes_written":     written,
			}
			text := tableTextRow(
				[]string{"file_token", "output_path", "bytes_written"},
				[]string{fileToken, outPath, fmt.Sprintf("%d", written)},
			)
			return state.Printer.Print(payload, text)
		},
	}
	annotateAuthServices(cmd, "drive-export")

	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (for example: docx, sheet, bitable)")
	cmd.Flags().StringVar(&format, "format", "", "export format (for example: pdf, docx, xlsx)")
	cmd.Flags().StringVar(&outPath, "out", "", "output file path (or - for stdout)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("format")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDriveURLsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "urls <file-token> [file-token...]",
		Short: "Print web URLs for Drive file tokens",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			fetcher, err := newDriveMetadataFetcher(ctx, state)
			if err != nil {
				return err
			}
			files := make([]larksdk.DriveFile, 0, len(args))
			for _, fileID := range args {
				token, _, err := parseResourceRef(fileID)
				if err != nil {
					return err
				}
				file, err := fetcher.get(ctx, token)
				if err != nil {
					return err
				}
				files = append(files, file)
			}
			payload := map[string]any{"files": files}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", file.Token, file.URL, file.Name))
			}
			text := tableText([]string{"token", "url", "name"}, lines, "no files found")
			return state.Printer.Print(payload, text)
		},
	}

	return cmd
}

func newDriveShareCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var linkShare string
	var externalAccess bool
	var inviteExternal bool
	var shareEntity string
	var securityEntity string
	var commentEntity string

	cmd := &cobra.Command{
		Use:   "share <file-token>",
		Short: "Update Drive file sharing permissions",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			refToken, refKind, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(refToken)
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			if refKind != "" && !cmd.Flags().Changed("type") {
				_ = cmd.Flags().Set("type", refKind)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			req := larksdk.UpdateDrivePermissionPublicRequest{
				LinkShareEntity: linkShare,
				ShareEntity:     shareEntity,
				SecurityEntity:  securityEntity,
				CommentEntity:   commentEntity,
			}
			if cmd.Flags().Changed("external-access") {
				req.ExternalAccess = &externalAccess
			}
			if cmd.Flags().Changed("invite-external") {
				req.InviteExternal = &inviteExternal
			}
			permission, err := state.SDK.UpdateDrivePermissionPublic(cmd.Context(), token, fileToken, fileType, req)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"permission": permission,
				"file_token": fileToken,
				"type":       fileType,
			}
			text := tableTextRow(
				[]string{"file_token", "type", "link_share", "external_access", "invite_external"},
				[]string{
					fileToken,
					fileType,
					permission.LinkShareEntity,
					fmt.Sprintf("%t", permission.ExternalAccess),
					fmt.Sprintf("%t", permission.InviteExternal),
				},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (for example: doc, docx, sheet, bitable, file)")
	cmd.Flags().StringVar(&linkShare, "link-share", "", "link share permission (for example: tenant_readable, anyone_readable)")
	cmd.Flags().BoolVar(&externalAccess, "external-access", false, "allow external access")
	cmd.Flags().BoolVar(&inviteExternal, "invite-external", false, "allow external invite")
	cmd.Flags().StringVar(&shareEntity, "share-entity", "", "share permission scope (for example: tenant_editable)")
	cmd.Flags().StringVar(&securityEntity, "security-entity", "", "security permission scope (for example: tenant_editable)")
	cmd.Flags().StringVar(&commentEntity, "comment-entity", "", "comment permission scope (for example: tenant_editable)")
	_ = cmd.MarkFlagRequired("type")
	cmd.MarkFlagsOneRequired("link-share", "external-access", "invite-external", "share-entity", "security-entity", "comment-entity")
	return cmd
}

type driveMetadataFetcher struct {
	state         *appState
	token         string
	tokenType     tokenType
	allowFallback bool
}

func newDriveMetadataFetcher(ctx context.Context, state *appState) (*driveMetadataFetcher, error) {
	if state.SDK == nil {
		return nil, errors.New("sdk client is required")
	}
	token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
	if err != nil {
		return nil, err
	}
	requested, err := parseTokenType(state.TokenType)
	if err != nil {
		return nil, err
	}
	return &driveMetadataFetcher{
		state:         state,
		token:         token,
		tokenType:     tokenType,
		allowFallback: requested == tokenTypeAuto,
	}, nil
}

func (f *driveMetadataFetcher) get(ctx context.Context, fileToken string) (larksdk.DriveFile, error) {
	file, err := driveFileMetadataWithToken(ctx, f.state.SDK, f.tokenType, f.token, fileToken)
	if err == nil || !f.allowFallback || f.tokenType != tokenTypeTenant {
		return file, err
	}
	userToken, userErr := resolveDriveSearchToken(ctx, f.state)
	if userErr != nil || userToken == "" {
		return file, err
	}
	if f.state.Verbose {
		fmt.Fprintln(errWriter(f.state), "retrying drive info with user access token")
	}
	fallbackFile, fallbackErr := driveFileMetadataWithToken(ctx, f.state.SDK, tokenTypeUser, userToken, fileToken)
	if fallbackErr != nil {
		return file, err
	}
	f.token = userToken
	f.tokenType = tokenTypeUser
	return fallbackFile, nil
}

func driveFileMetadataWithToken(ctx context.Context, sdk *larksdk.Client, accessType tokenType, token, fileToken string) (larksdk.DriveFile, error) {
	if sdk == nil {
		return larksdk.DriveFile{}, errors.New("sdk client is required")
	}
	switch accessType {
	case tokenTypeUser:
		return sdk.GetDriveFileMetadataWithUserToken(ctx, token, larksdk.GetDriveFileRequest{FileToken: fileToken})
	case tokenTypeTenant:
		return sdk.GetDriveFileMetadata(ctx, token, larksdk.GetDriveFileRequest{FileToken: fileToken})
	default:
		return larksdk.DriveFile{}, fmt.Errorf("unsupported token type %s", accessType)
	}
}
