/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/imprt"
	"gomificator/internal/settings"
	"gomificator/internal/storage"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// autoimportCmd represents the autoimport command
var autoimportCmd = &cobra.Command{
	Use:   "autoimport",
	Short: "Auto-import newest JSON on interval",
	Long:  `Watches autoimport.path and, every interval, imports the newest JSON backup until you quit.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := settings.LoadConfig(nil)
		if err != nil {
			panic(err)
		}

		m := makeAutoImportModel(cfg)
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("autoimport failed:", err)
			os.Exit(1)
		}
	},
}

// --- Bubble Tea model for autoimport ---

type autoKeymap struct {
	quit key.Binding
	run  key.Binding
}

type autoModel struct {
	cfg       *settings.Config
	interval  time.Duration
	help      help.Model
	keymap    autoKeymap
	quitting  bool
	lastRunAt time.Time
	lastFile  string
	lastCount int
	lastErr   error
}

func makeAutoImportModel(cfg *settings.Config) tea.Model {
	return autoModel{
		cfg:      cfg,
		interval: cfg.AutoImport.Every,
		help:     help.New(),
		keymap: autoKeymap{
			quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
			run:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run now")),
		},
	}
}

type tickMsg time.Time
type importResultMsg struct {
	file  string
	count int
	err   error
	at    time.Time
}

func (m autoModel) Init() tea.Cmd {
	return tea.Batch(
		m.doImport(),
		tea.Tick(m.interval, func(t time.Time) tea.Msg { return tickMsg(t) }),
	)
}

func (m autoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.run):
			return m, m.doImport()
		}
	case tickMsg:
		// schedule import and next tick
		return m, tea.Batch(m.doImport(), tea.Tick(m.interval, func(t time.Time) tea.Msg { return tickMsg(t) }))
	case importResultMsg:
		m.lastRunAt = msg.at
		m.lastFile = msg.file
		m.lastCount = msg.count
		m.lastErr = msg.err
		return m, nil
	}
	return m, nil
}

func (m autoModel) View() string {
	s := "Autoimport running\n"
	s += fmt.Sprintf("Dir: %s\n", m.cfg.AutoImport.Path)
	s += fmt.Sprintf("Every: %s\n", m.interval)
	if !m.lastRunAt.IsZero() {
		if m.lastErr != nil {
			s += fmt.Sprintf("Last: %s ERROR: %v\n", m.lastRunAt.Format(time.RFC3339), m.lastErr)
		} else {
			s += fmt.Sprintf("Last: %s file=%s imported=%d\n", m.lastRunAt.Format(time.RFC3339), filepath.Base(m.lastFile), m.lastCount)
		}
	}
	if !m.quitting {
		s += "\n" + m.help.ShortHelpView([]key.Binding{m.keymap.run, m.keymap.quit})
	}
	return s
}

func (m autoModel) doImport() tea.Cmd {
	dir := m.cfg.AutoImport.Path
	pattern := filepath.Join(dir, "*.json")
	return func() tea.Msg {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return importResultMsg{err: fmt.Errorf("glob %s: %w", pattern, err), at: time.Now()}
		}
		if len(matches) == 0 {
			return importResultMsg{err: fmt.Errorf("no JSON files found in %s", dir), at: time.Now()}
		}
		var newestPath string
		var newestMod int64
		for _, p := range matches {
			if !strings.EqualFold(filepath.Ext(p), ".json") {
				continue
			}
			info, err := os.Stat(p)
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			mt := info.ModTime().UnixNano()
			if mt > newestMod || newestPath == "" {
				newestMod = mt
				newestPath = p
			}
		}
		if newestPath == "" {
			return importResultMsg{err: fmt.Errorf("no regular JSON files found in %s", dir), at: time.Now()}
		}

		f, err := os.Open(newestPath)
		if err != nil {
			return importResultMsg{err: fmt.Errorf("open newest file: %w", err), at: time.Now()}
		}
		defer f.Close()

		importer := imprt.NewImporterFromSuperProductivityBackupFile(f)
		timers, err := importer.Import()
		if err != nil {
			return importResultMsg{file: newestPath, err: err, at: time.Now()}
		}

		storageService, err := storage.NewSqlliteStorage()
		if err != nil {
			return importResultMsg{file: newestPath, err: err, at: time.Now()}
		}
		for _, timer := range timers {
			if _, err := storageService.TimersRepo.Save(timer); err != nil {
				return importResultMsg{file: newestPath, err: err, at: time.Now()}
			}
		}
		return importResultMsg{file: newestPath, count: len(timers), at: time.Now()}
	}
}

func init() {
	rootCmd.AddCommand(autoimportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// autoimportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// autoimportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
