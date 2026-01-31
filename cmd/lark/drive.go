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
	}
	cmd.AddCommand(newDriveListCmd(state))
	cmd.AddCommand(newDriveSearchCmd(state))
	cmd.AddCommand(newDriveGetCmd(state))
	cmd.AddCommand(newDriveExportCmd(state))
	cmd.AddCommand(newDriveDownloadCmd(state))
	cmd.AddCommand(newDriveUploadCmd(state))
	cmd.AddCommand(newDriveURLsCmd(state))
	cmd.AddCommand(newDriveShareCmd(state))
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
				return errors.New("limit must be greater than 0")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			files := make([]larksdk.DriveFile, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxDrivePageSize {
					pageSize = maxDrivePageSize
				}
				result, err := state.SDK.ListDriveFiles(context.Background(), token, larksdk.ListDriveFilesRequest{
					FolderToken: folderID,
					PageSize:    pageSize,
					PageToken:   pageToken,
				})
				if err != nil {
					return err
				}
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
			payload := map[string]any{"files": files}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL))
			}
			text := "no files found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")
	return cmd
}

func newDriveSearchCmd(state *appState) *cobra.Command {
	var query string
	var folderID string
	var fileTypes []string
	var limit int

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Drive files by text",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			files := make([]larksdk.DriveFile, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxDrivePageSize {
					pageSize = maxDrivePageSize
				}
				result, err := state.SDK.SearchDriveFiles(context.Background(), token, larksdk.SearchDriveFilesRequest{
					Query:       query,
					FolderToken: folderID,
					FileTypes:   fileTypes,
					PageSize:    pageSize,
					PageToken:   pageToken,
				})
				if err != nil {
					return err
				}
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
			payload := map[string]any{"files": files}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL))
			}
			text := "no files found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search text")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token to scope search")
	cmd.Flags().StringArrayVar(&fileTypes, "type", nil, "filter by file type (docx|sheet|bitable|file|doc); repeatable")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func newDriveGetCmd(state *appState) *cobra.Command {
	var fileToken string

	cmd := &cobra.Command{
		Use:   "get <file-token>",
		Short: "Get Drive file metadata",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return cobra.MaximumNArgs(1)(cmd, args)
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				fileToken = args[0]
			}
			if fileToken == "" {
				return errors.New("file-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			file, err := state.SDK.GetDriveFileMetadata(context.Background(), token, larksdk.GetDriveFileRequest{
				FileToken: fileToken,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"file": file}
			text := fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (or provide as positional argument)")
	return cmd
}

func newDriveUploadCmd(state *appState) *cobra.Command {
	var filePath string
	var folderToken string
	var uploadName string

	cmd := &cobra.Command{
		Use:   "upload --file <path>",
		Short: "Upload a local file to Drive",
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
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if folderToken == "" {
				folderToken = "root"
			}
			result, err := state.SDK.UploadDriveFile(context.Background(), token, larksdk.UploadDriveFileRequest{
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
				meta, err := state.SDK.GetDriveFileMetadata(context.Background(), token, larksdk.GetDriveFileRequest{FileToken: fileInfo.Token})
				if err != nil {
					return err
				}
				fileInfo = meta
			}
			payload := map[string]any{
				"file_token": fileInfo.Token,
				"file":       fileInfo,
			}
			text := fileInfo.Token
			if fileInfo.URL != "" {
				text = fmt.Sprintf("%s\t%s\t%s\t%s", fileInfo.Token, fileInfo.Name, fileInfo.FileType, fileInfo.URL)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "path to local file")
	cmd.Flags().StringVar(&folderToken, "folder-token", "", "Drive folder token (default: root)")
	cmd.Flags().StringVar(&uploadName, "name", "", "override the uploaded file name")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newDriveDownloadCmd(state *appState) *cobra.Command {
	var fileToken string
	var outPath string

	cmd := &cobra.Command{
		Use:   "download --file-token <token> --out <path>",
		Short: "Download a Drive file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if info, err := os.Stat(outPath); err == nil && info.IsDir() {
				return fmt.Errorf("output path is a directory: %s", outPath)
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			reader, err := state.SDK.DownloadDriveFile(context.Background(), token, fileToken)
			if err != nil {
				return err
			}
			defer reader.Close()
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			written, err := io.Copy(outFile, reader)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"file_token":    fileToken,
				"output_path":   outPath,
				"bytes_written": written,
			}
			text := fmt.Sprintf("%s\t%s\t%d", fileToken, outPath, written)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token")
	cmd.Flags().StringVar(&outPath, "out", "", "output file path")
	_ = cmd.MarkFlagRequired("file-token")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDriveExportCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var format string
	var outPath string

	cmd := &cobra.Command{
		Use:   "export <file-token>",
		Short: "Export a Drive file",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return cobra.MaximumNArgs(1)(cmd, args)
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				fileToken = args[0]
			}
			if fileToken == "" {
				return errors.New("file-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if info, err := os.Stat(outPath); err == nil && info.IsDir() {
				return fmt.Errorf("output path is a directory: %s", outPath)
			}
			format = strings.ToLower(format)
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			ticket, err := state.SDK.CreateExportTask(context.Background(), token, larksdk.CreateExportTaskRequest{
				Token:         fileToken,
				Type:          fileType,
				FileExtension: format,
			})
			if err != nil {
				return err
			}
			result, err := pollExportTask(context.Background(), state.SDK, token, ticket)
			if err != nil {
				return err
			}
			reader, err := state.SDK.DownloadExportedFile(context.Background(), token, result.FileToken)
			if err != nil {
				return err
			}
			defer reader.Close()
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			written, err := io.Copy(outFile, reader)
			if err != nil {
				return err
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
			text := fmt.Sprintf("%s\t%s\t%d", fileToken, outPath, written)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (or provide as positional argument)")
	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (for example: docx, sheet, bitable)")
	cmd.Flags().StringVar(&format, "format", "", "export format (for example: pdf, docx, xlsx)")
	cmd.Flags().StringVar(&outPath, "out", "", "output file path")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("format")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDriveURLsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "urls <file-id> [file-id...]",
		Short: "Print web URLs for Drive file IDs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			files := make([]larksdk.DriveFile, 0, len(args))
			for _, fileID := range args {
				file, err := state.SDK.GetDriveFileMetadata(context.Background(), token, larksdk.GetDriveFileRequest{FileToken: fileID})
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
			return state.Printer.Print(payload, strings.Join(lines, "\n"))
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
			if len(args) > 1 {
				return cobra.MaximumNArgs(1)(cmd, args)
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				fileToken = args[0]
			}
			if fileToken == "" {
				return errors.New("file-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
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
			permission, err := state.SDK.UpdateDrivePermissionPublic(context.Background(), token, fileToken, fileType, req)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"permission": permission,
				"file_token": fileToken,
				"type":       fileType,
			}
			text := fmt.Sprintf("%s\t%s\t%s\t%t\t%t", fileToken, fileType, permission.LinkShareEntity, permission.ExternalAccess, permission.InviteExternal)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (or provide as positional argument)")
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
