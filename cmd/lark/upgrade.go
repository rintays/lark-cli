package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"

	"lark/internal/updatecheck"
	"lark/internal/version"
)

type upgradeOptions struct {
	CheckOnly   bool
	Prerelease  bool
	Force       bool
	Source      string
	Yes         bool
	Interactive bool
}

func newUpgradeCmd(state *appState) *cobra.Command {
	var opts upgradeOptions

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade lark to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Interactive = true
			return upgradeToLatest(context.Background(), state, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.CheckOnly, "check", false, "only check for updates, do not install")
	cmd.Flags().BoolVar(&opts.Prerelease, "prerelease", false, "allow upgrading to prerelease versions")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "force reinstall even if already on latest")
	cmd.Flags().StringVar(&opts.Source, "source", "github", "update source (github)")
	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "do not prompt; assume yes")

	return cmd
}

func upgradeToLatest(ctx context.Context, state *appState, opts upgradeOptions) error {
	if opts.Source == "" {
		opts.Source = "github"
	}
	source := strings.ToLower(strings.TrimSpace(opts.Source))
	if source != "github" {
		return fmt.Errorf("unsupported source %q (only github is supported for now)", source)
	}

	current := version.Version
	if current == "" {
		current = "v0.0.0"
	}

	// Brew-managed installs should be upgraded via brew to avoid breaking brew state.
	if isBrewManaged() {
		latest, found, err := detectLatest(ctx, opts.Prerelease)
		if err != nil {
			return err
		}
		if !found || latest == "" {
			return errors.New("latest version not found")
		}
		if opts.CheckOnly {
			// creativeprojects/go-selfupdate returns version without leading "v".
			if strings.TrimPrefix(current, "v") == strings.TrimPrefix(latest, "v") {
				fmt.Fprintln(errWriter(state), "Already up to date.")
				return nil
			}
			fmt.Fprintf(errWriter(state), "Update available: %s -> v%s (run: brew upgrade rintays/tap/lark)\n", current, strings.TrimPrefix(latest, "v"))
			return nil
		}
		if !opts.Yes {
			fmt.Fprintln(errWriter(state), "Upgrading via Homebrew: brew upgrade rintays/tap/lark")
		}
		return runBrewUpgrade(ctx)
	}

	latest, found, err := detectLatest(ctx, opts.Prerelease)
	if err != nil {
		return err
	}
	if !found || latest == "" {
		return errors.New("latest version not found")
	}

	if opts.CheckOnly {
		if strings.TrimPrefix(current, "v") == strings.TrimPrefix(latest, "v") {
			fmt.Fprintln(errWriter(state), "Already up to date.")
			return nil
		}
		fmt.Fprintf(errWriter(state), "Update available: %s -> v%s\n", current, strings.TrimPrefix(latest, "v"))
		return nil
	}

	if !opts.Force {
		if level, ok := updatecheck.DiffLevel(current, latest); ok && level == "" {
			fmt.Fprintln(errWriter(state), "Already up to date.")
			return nil
		}
	}

	// Minor/major upgrades default to interactive confirmation unless --yes.
	if opts.Interactive && !opts.Yes {
		level, ok := updatecheck.DiffLevel(current, latest)
		if ok && (level == "minor" || level == "major") {
			m := newUpgradePromptModel(fmt.Sprintf("%s -> %s", current, latest), level)
			p := tea.NewProgram(m)
			res, err := p.Run()
			if err != nil {
				return err
			}
			if model, ok := res.(upgradePromptModel); ok {
				if model.choice == "skip" {
					fmt.Fprintln(errWriter(state), "Upgrade canceled.")
					return nil
				}
			}
		}
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	repo := selfupdate.ParseSlug("rintays/lark-cli")
	updated, err := selfupdate.UpdateCommand(ctx, exe, current, repo)
	if err != nil {
		return err
	}
	if updated == nil {
		fmt.Fprintln(errWriter(state), "No update applied.")
		return nil
	}
	fmt.Fprintf(errWriter(state), "Upgraded to %s\n", updated.Version())
	return nil
}

func detectLatest(ctx context.Context, prerelease bool) (string, bool, error) {
	repo := selfupdate.ParseSlug("rintays/lark-cli")
	if !prerelease {
		rel, found, err := selfupdate.DetectLatest(ctx, repo)
		if err != nil {
			return "", false, err
		}
		if !found || rel == nil {
			return "", false, nil
		}
		return rel.Version(), true, nil
	}

	src, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return "", false, err
	}
	rels, err := src.ListReleases(ctx, repo)
	if err != nil {
		return "", false, err
	}
	bestTag := ""
	bestVer := updatecheck.Semver{}
	bestOK := false
	for _, r := range rels {
		tag := strings.TrimSpace(r.GetTagName())
		if tag == "" {
			continue
		}
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}
		ver, ok := updatecheck.ParseSemver(tag)
		if !ok {
			continue
		}
		if !bestOK || updatecheck.Compare(bestVer, ver) < 0 {
			bestTag = tag
			bestVer = ver
			bestOK = true
		}
	}
	if bestTag == "" {
		return "", false, nil
	}
	return bestTag, true, nil
}

func runBrewUpgrade(ctx context.Context) error {
	brew, err := exec.LookPath("brew")
	if err != nil {
		return errors.New("brew not found")
	}
	c := exec.CommandContext(ctx, brew, "upgrade", "rintays/tap/lark")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func isBrewManaged() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, _ = filepath.EvalSymlinks(exe)
	exe = filepath.Clean(exe)
	// Typical paths:
	// /opt/homebrew/Cellar/lark/<ver>/bin/lark
	// /usr/local/Cellar/lark/<ver>/bin/lark
	return strings.Contains(exe, string(filepath.Separator)+"Cellar"+string(filepath.Separator)+"lark"+string(filepath.Separator))
}

// ---- bubbletea prompt ----

type upgradePromptModel struct {
	msg    string
	level  string
	choice string // "upgrade" | "skip"
}

func newUpgradePromptModel(msg, level string) upgradePromptModel {
	return upgradePromptModel{msg: msg, level: level, choice: "upgrade"}
}

func (m upgradePromptModel) Init() tea.Cmd { return nil }

func (m upgradePromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.choice = "skip"
			return m, tea.Quit
		case "y", "enter":
			m.choice = "upgrade"
			return m, tea.Quit
		case "n":
			m.choice = "skip"
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m upgradePromptModel) View() string {
	return fmt.Sprintf("New %s update available (%s).\nUpgrade now? [y/N]\n", strings.ToUpper(m.level), m.msg)
}
