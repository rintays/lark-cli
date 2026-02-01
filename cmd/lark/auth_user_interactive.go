package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"

	"lark/internal/authregistry"
	"lark/internal/output"
)

type userOAuthSelectionMode int

const (
	userOAuthSelectServices userOAuthSelectionMode = iota
	userOAuthSelectScopes
)

var errUserOAuthCanceled = errors.New("login canceled")

type userOAuthInteractiveSelection struct {
	Mode     userOAuthSelectionMode
	Services []string
	Scopes   []string
}

func promptUserOAuthSelection(state *appState, account string) (userOAuthInteractiveSelection, error) {
	if state == nil {
		return userOAuthInteractiveSelection{}, errors.New("missing app state")
	}
	if state.Printer.JSON || !interactiveAuthAvailable() {
		return userOAuthInteractiveSelection{}, errors.New("interactive login requires a TTY; use --scopes or --services")
	}

	prevServices, prevScopes := previousUserOAuthSelections(state, account)
	defaultMode := userOAuthSelectServices
	if len(prevServices) == 0 && len(prevScopes) > 0 {
		defaultMode = userOAuthSelectScopes
	}

	modeIndex, canceled, err := runModeSelect(defaultMode)
	if err != nil {
		return userOAuthInteractiveSelection{}, err
	}
	if canceled {
		return userOAuthInteractiveSelection{}, errUserOAuthCanceled
	}

	if modeIndex == 0 {
		services, err := promptUserOAuthServices(prevServices)
		if err != nil {
			return userOAuthInteractiveSelection{}, err
		}
		return userOAuthInteractiveSelection{Mode: userOAuthSelectServices, Services: services}, nil
	}

	scopes, err := promptUserOAuthScopes(state, prevScopes)
	if err != nil {
		return userOAuthInteractiveSelection{}, err
	}
	return userOAuthInteractiveSelection{Mode: userOAuthSelectScopes, Scopes: scopes}, nil
}

func interactiveAuthAvailable() bool {
	if !output.AutoStyle(os.Stdout) {
		return false
	}
	fd := os.Stdin.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func previousUserOAuthSelections(state *appState, account string) ([]string, []string) {
	if state == nil || state.Config == nil {
		return nil, nil
	}
	acct, ok := loadUserAccount(state.Config, account)
	if !ok {
		return nil, nil
	}
	var services []string
	var scopes []string
	if acct.UserRefreshTokenPayload != nil {
		if len(acct.UserRefreshTokenPayload.Services) > 0 {
			services = normalizeServices(acct.UserRefreshTokenPayload.Services)
		}
		if strings.TrimSpace(acct.UserRefreshTokenPayload.Scopes) != "" {
			scopes = normalizeScopes(parseScopeList(acct.UserRefreshTokenPayload.Scopes))
		}
	}
	if len(scopes) == 0 && len(acct.UserScopes) > 0 {
		scopes = normalizeScopes(acct.UserScopes)
	}
	if len(scopes) == 0 && strings.TrimSpace(acct.UserAccessTokenScope) != "" {
		scopes = normalizeScopes(parseScopeList(acct.UserAccessTokenScope))
	}
	return services, scopes
}

func promptUserOAuthServices(previous []string) ([]string, error) {
	services := authregistry.ListUserOAuthServices()
	if len(services) == 0 {
		return nil, errors.New("no user OAuth services available")
	}

	selectedSet := make(map[string]struct{})
	defaults := previous
	if len(defaults) == 0 {
		defaults = authregistry.DefaultUserOAuthServices
	}
	for _, svc := range normalizeServices(defaults) {
		selectedSet[svc] = struct{}{}
	}

	items := make([]optionItem, 0, len(services))
	for _, svc := range services {
		_, selected := selectedSet[svc]
		items = append(items, optionItem{
			Label:    svc,
			Value:    svc,
			Selected: selected,
		})
	}

	selected, canceled, err := runMultiSelect("Select OAuth services", items, false)
	if err != nil {
		return nil, err
	}
	if canceled {
		return nil, errUserOAuthCanceled
	}
	if len(selected) == 0 {
		return nil, errors.New("no services selected")
	}
	return normalizeServices(selected), nil
}

func promptUserOAuthScopes(state *appState, previous []string) ([]string, error) {
	available := userOAuthAvailableScopes()
	defaults := previous
	if len(defaults) == 0 {
		if scopes, _, err := resolveUserOAuthScopes(state, userOAuthScopeOptions{}); err == nil && len(scopes) > 0 {
			defaults = scopes
		} else {
			defaults = []string{defaultUserOAuthScope}
		}
	}

	available = appendMissingScopes(available, defaults)
	selectedSet := make(map[string]struct{})
	for _, scope := range canonicalizeUserOAuthScopes(defaults) {
		selectedSet[scope] = struct{}{}
	}

	items := make([]optionItem, 0, len(available))
	for _, scope := range available {
		_, selected := selectedSet[scope]
		locked := scope == defaultUserOAuthScope
		items = append(items, optionItem{
			Label:    scope,
			Value:    scope,
			Selected: selected || locked,
			Locked:   locked,
			Tag:      "required",
		})
	}

	selected, canceled, err := runMultiSelect("Select OAuth scopes", items, true)
	if err != nil {
		return nil, err
	}
	if canceled {
		return nil, errUserOAuthCanceled
	}
	selected = ensureOfflineAccess(selected)
	return canonicalizeUserOAuthScopes(selected), nil
}

func appendMissingScopes(available []string, defaults []string) []string {
	seen := make(map[string]struct{}, len(available))
	for _, scope := range available {
		seen[scope] = struct{}{}
	}
	for _, scope := range normalizeScopes(defaults) {
		if _, ok := seen[scope]; ok {
			continue
		}
		available = append(available, scope)
		seen[scope] = struct{}{}
	}
	return available
}

func runModeSelect(defaultMode userOAuthSelectionMode) (int, bool, error) {
	options := []string{
		"Select by service (recommended)",
		"Select by scope",
	}
	model := &singleSelectModel{
		title:   "Choose OAuth selection mode",
		options: options,
		cursor:  modeIndexFor(defaultMode),
	}
	program := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
	result, err := program.Run()
	if err != nil {
		return 0, false, err
	}
	final, ok := result.(*singleSelectModel)
	if !ok {
		return 0, false, errors.New("unexpected selection result")
	}
	return final.cursor, final.canceled, nil
}

func runMultiSelect(title string, items []optionItem, allowEmpty bool) ([]string, bool, error) {
	model := &multiSelectModel{
		title:      title,
		items:      items,
		allowEmpty: allowEmpty,
	}
	program := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
	result, err := program.Run()
	if err != nil {
		return nil, false, err
	}
	final, ok := result.(*multiSelectModel)
	if !ok {
		return nil, false, errors.New("unexpected selection result")
	}
	return final.selectedValues(), final.canceled, nil
}

type optionItem struct {
	Label    string
	Value    string
	Selected bool
	Locked   bool
	Tag      string
}

type singleSelectModel struct {
	title    string
	options  []string
	cursor   int
	canceled bool
}

func (m *singleSelectModel) Init() tea.Cmd {
	return nil
}

func (m *singleSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *singleSelectModel) View() string {
	styles := newSelectorStyles()
	var b strings.Builder
	b.WriteString(styles.title.Render(m.title))
	b.WriteString("\n\n")
	for i, option := range m.options {
		cursor := " "
		line := option
		if i == m.cursor {
			cursor = styles.cursor.Render(">")
			line = styles.active.Render(option)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, line))
	}
	b.WriteString("\nenter confirm | q cancel")
	return b.String()
}

type multiSelectModel struct {
	title      string
	items      []optionItem
	cursor     int
	allowEmpty bool
	canceled   bool
	errMessage string
}

func (m *multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m *multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			m.toggleSelected()
		case "enter":
			if !m.allowEmpty && len(m.selectedValues()) == 0 {
				m.errMessage = "Select at least one item."
				return m, nil
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *multiSelectModel) View() string {
	styles := newSelectorStyles()
	var b strings.Builder
	b.WriteString(styles.title.Render(m.title))
	b.WriteString("\n\n")
	if m.errMessage != "" {
		b.WriteString(styles.hint.Render(m.errMessage))
		b.WriteString("\n\n")
	}
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = styles.cursor.Render(">")
		}
		box := "[ ]"
		if item.Selected {
			box = "[x]"
		}
		label := item.Label
		if item.Locked && item.Tag != "" {
			label = fmt.Sprintf("%s (%s)", label, item.Tag)
		}
		line := fmt.Sprintf("%s %s", box, label)
		if i == m.cursor {
			line = styles.active.Render(line)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, line))
	}
	b.WriteString("\nspace toggle | enter confirm | q cancel")
	return b.String()
}

func (m *multiSelectModel) toggleSelected() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}
	item := m.items[m.cursor]
	if item.Locked {
		return
	}
	item.Selected = !item.Selected
	m.items[m.cursor] = item
}

func (m *multiSelectModel) selectedValues() []string {
	selected := make([]string, 0, len(m.items))
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item.Value)
		}
	}
	return selected
}

type selectorStyles struct {
	title  lipgloss.Style
	cursor lipgloss.Style
	active lipgloss.Style
	hint   lipgloss.Style
}

func newSelectorStyles() selectorStyles {
	brand := output.BrandColor()
	return selectorStyles{
		title:  lipgloss.NewStyle().Bold(true).Foreground(brand),
		cursor: lipgloss.NewStyle().Bold(true).Foreground(brand),
		active: lipgloss.NewStyle().Bold(true),
		hint:   lipgloss.NewStyle().Foreground(brand),
	}
}

func modeIndexFor(mode userOAuthSelectionMode) int {
	if mode == userOAuthSelectScopes {
		return 1
	}
	return 0
}
