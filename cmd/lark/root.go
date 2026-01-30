package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/config"
	"lark/internal/larkapi"
	"lark/internal/larksdk"
	"lark/internal/output"
)

type appState struct {
	ConfigPath string
	Config     *config.Config
	JSON       bool
	Verbose    bool
	Printer    output.Printer
	Client     *larkapi.Client
	SDK        *larksdk.Client
}

func newRootCmd() *cobra.Command {
	state := &appState{}
	cmd := &cobra.Command{
		Use:          "lark",
		Short:        "A Go CLI for Feishu/Lark inspired by gog",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if state.ConfigPath == "" {
				path, err := config.DefaultPath()
				if err != nil {
					return err
				}
				state.ConfigPath = path
			}
			cfg, err := config.Load(state.ConfigPath)
			if err != nil {
				return err
			}
			state.Config = cfg
			state.Printer = output.Printer{Writer: cmd.OutOrStdout(), JSON: state.JSON}
			state.Client = &larkapi.Client{
				BaseURL:   cfg.BaseURL,
				AppID:     cfg.AppID,
				AppSecret: cfg.AppSecret,
			}
			sdkClient, err := larksdk.New(cfg)
			if err == nil {
				state.SDK = sdkClient
			} else if state.Verbose {
				fmt.Fprintf(state.Printer.Writer, "SDK disabled: %v\n", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&state.ConfigPath, "config", "", "config path (default: ~/.config/lark/config.json)")
	cmd.PersistentFlags().BoolVar(&state.JSON, "json", false, "output JSON")
	cmd.PersistentFlags().BoolVar(&state.Verbose, "verbose", false, "verbose output")

	cmd.AddCommand(newVersionCmd(state))
	cmd.AddCommand(newAuthCmd(state))
	cmd.AddCommand(newWhoamiCmd(state))
	cmd.AddCommand(newMsgCmd(state))
	cmd.AddCommand(newChatsCmd(state))
	cmd.AddCommand(newUsersCmd(state))
	cmd.AddCommand(newDriveCmd(state))
	cmd.AddCommand(newDocsCmd(state))
	cmd.AddCommand(newSheetsCmd(state))
	cmd.AddCommand(newCalendarCmd(state))
	cmd.AddCommand(newMeetingCmd(state))
	cmd.AddCommand(newContactsCmd(state))
	cmd.AddCommand(newMailCmd(state))

	return cmd
}

func requireCredentials(cfg *config.Config) error {
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return errors.New("app_id and app_secret must be set in config")
	}
	return nil
}

func cachedTokenValid(cfg *config.Config, now time.Time) bool {
	if cfg.TenantAccessToken == "" || cfg.TenantAccessTokenExpiresAt == 0 {
		return false
	}
	return cfg.TenantAccessTokenExpiresAt > now.Add(60*time.Second).Unix()
}

func ensureTenantToken(ctx context.Context, state *appState) (string, error) {
	if err := requireCredentials(state.Config); err != nil {
		return "", err
	}
	if cachedTokenValid(state.Config, time.Now()) {
		return state.Config.TenantAccessToken, nil
	}
	if state.Verbose {
		fmt.Fprintln(state.Printer.Writer, "refreshing tenant access token")
	}
	token, expiresIn, err := state.Client.TenantAccessToken(ctx)
	if err != nil {
		return "", err
	}
	state.Config.TenantAccessToken = token
	state.Config.TenantAccessTokenExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
	if err := config.Save(state.ConfigPath, state.Config); err != nil {
		return "", err
	}
	return token, nil
}

func execute() int {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
