package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/urfave/cli/v3"

	"stock-cli/pkg/config"
)

var isTerminalFn = func(fd uintptr) bool { return term.IsTerminal(fd) }

var configCmd = cli.Command{
	Name:  "config",
	Usage: "Manage Stock CLI configuration",
	Commands: []*cli.Command{
		&configSetCmd,
		&configShowCmd,
		&configPathCmd,
	},
}

var configSetCmd = cli.Command{
	Name:  "set",
	Usage: "Set Kiwoom REST API credentials interactively",
	Action: func(_ context.Context, _ *cli.Command) error {
		return runConfigSet()
	},
}

var configShowCmd = cli.Command{
	Name:  "show",
	Usage: "Show current Kiwoom credentials and their source",
	Action: func(_ context.Context, _ *cli.Command) error {
		return runConfigShow()
	},
}

var configPathCmd = cli.Command{
	Name:  "path",
	Usage: "Print the config file path",
	Action: func(_ context.Context, _ *cli.Command) error {
		return runConfigPath()
	},
}

func runConfigSet() error {
	if !isTerminalFn(os.Stdin.Fd()) {
		return fmt.Errorf("config set requires an interactive terminal (stdin is not a TTY)")
	}

	path, err := config.Path()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		fmt.Print("Existing credentials found. Overwrite? [y/N]: ")
		var answer string
		fmt.Fscanln(os.Stdin, &answer)
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	m, err := runConfigSetTUI()
	if err != nil {
		return err
	}
	if m.aborted {
		fmt.Fprintln(os.Stderr, "Aborted.")
		return cli.Exit("", 130)
	}

	creds := config.Credentials{
		AppKey:    strings.TrimSpace(m.inputs[0].Value()),
		SecretKey: strings.TrimSpace(m.inputs[1].Value()),
	}
	if err := config.Save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	expandedPath := path
	if home, err := os.UserHomeDir(); err == nil {
		expandedPath = strings.Replace(path, home, "~", 1)
	}
	fmt.Printf("Credentials saved to %s\n", expandedPath)
	return nil
}

func runConfigShow() error {
	creds, err := config.Load()
	if err != nil {
		return err
	}

	if creds.IsEmpty() {
		fmt.Fprintln(os.Stderr, "No credentials configured. Run 'stock config set' to set up your Kiwoom API keys, or set KIWOOM_APPKEY / KIWOOM_SECRETKEY environment variables.")
		return nil
	}

	path, _ := config.Path()
	if home, err := os.UserHomeDir(); err == nil {
		path = strings.Replace(path, home, "~", 1)
	}

	fmt.Printf("appkey:      %s  (source: %s)\n", maskKey(creds.AppKey), creds.AppKeySource)
	fmt.Printf("secretkey:   %s  (source: %s)\n", maskKey(creds.SecretKey), creds.SecretKeySource)
	fmt.Printf("config_file: %s\n", path)
	return nil
}

func runConfigPath() error {
	path, err := config.Path()
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

type configSetModel struct {
	inputs  [2]textinput.Model
	current int
	errMsg  string
	done    bool
	aborted bool
}

func newConfigSetModel() configSetModel {
	m := configSetModel{}

	for i := range m.inputs {
		t := textinput.New()
		if i < 2 {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		}
		m.inputs[i] = t
	}
	m.inputs[0].Placeholder = "app key"
	m.inputs[1].Placeholder = "secret key"
	m.inputs[0].Focus()
	return m
}

func (m configSetModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configSetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit
		case tea.KeyEnter:
			val := strings.TrimSpace(m.inputs[m.current].Value())
			if val == "" {
				if m.current == 0 {
					m.errMsg = "app key cannot be empty"
				} else {
					m.errMsg = "secret key cannot be empty"
				}
				return m, nil
			}
			m.errMsg = ""
			if m.current < len(m.inputs)-1 {
				m.inputs[m.current].Blur()
				m.current++
				m.inputs[m.current].Focus()
				return m, textinput.Blink
			}
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.inputs[m.current], cmd = m.inputs[m.current].Update(msg)
	return m, cmd
}

func (m configSetModel) View() string {
	var b strings.Builder
	b.WriteString("Configure Kiwoom REST API credentials\n\n")

	b.WriteString("App Key: ")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n")

	if m.current >= 1 {
		b.WriteString("Secret Key: ")
		b.WriteString(m.inputs[1].View())
		b.WriteString("\n")
	}

	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(m.errMsg)
		b.WriteString("\n")
	}

	return b.String()
}

func runConfigSetTUI() (configSetModel, error) {
	m := newConfigSetModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return configSetModel{}, fmt.Errorf("input error: %w", err)
	}
	return result.(configSetModel), nil
}
