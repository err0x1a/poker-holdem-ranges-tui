package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"poker-holdem-ranges-tui/list"
	"poker-holdem-ranges-tui/ranges"
)

var title string

var titleColor string

var rootCmd = &cobra.Command{
	Use:   "phr-tui [files or directories...]",
	Short: "A TUI for viewing poker ranges",
	Long:  `A terminal user interface for viewing and studying poker ranges.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		files := expandArgs(args)
		if len(files) == 0 {
			fmt.Println("No .yaml files found")
			os.Exit(1)
		}

		model := NewMainModel(files, title, titleColor)

		if _, err := tea.NewProgram(model).Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func expandArgs(args []string) []string {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			continue
		}
		if info.IsDir() {
			entries, err := os.ReadDir(arg)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
					files = append(files, filepath.Join(arg, entry.Name()))
				}
			}
		} else if strings.HasSuffix(arg, ".yaml") || strings.HasSuffix(arg, ".yml") {
			files = append(files, arg)
		}
	}
	return files
}

func init() {
	rootCmd.Flags().StringVarP(&title, "title", "t", "", "Title for the ranges list (optional)")
	rootCmd.Flags().StringVarP(&titleColor, "title-color", "c", "", "Color for the title (hex, e.g. #FF0000)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// MainModel is the main model for the application
type MainModel struct {
	listModel        *list.Model
	rangesModel      ranges.Model
	selectedFilePath string
}

// NewMainModel creates a new main model
func NewMainModel(files []string, title string, titleColor string) MainModel {
	return MainModel{
		listModel:   list.New(files, title, titleColor),
		rangesModel: ranges.New(),
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(m.listModel.Init(), m.rangesModel.Init())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var listCmd, rangesCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		listWidth := msg.Width / 4
		m.listModel.SetSize(listWidth, 43)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	listCmd = m.listModel.Update(msg)

	_, _, filePath := m.listModel.SelectedItem()
	if filePath != "" && filePath != m.selectedFilePath {
		m.selectedFilePath = filePath
		if rf, err := ranges.LoadRangeFile(filePath); err == nil {
			m.rangesModel = ranges.NewWithRange(rf.ToHandColors(), rf.GetLegend(), rf.Details)
		}
	}

	var newRangesModel tea.Model
	newRangesModel, rangesCmd = m.rangesModel.Update(msg)
	m.rangesModel = newRangesModel.(ranges.Model)

	return m, tea.Batch(listCmd, rangesCmd)
}

func (m MainModel) View() string {
	spacer := "  "
	return lipgloss.JoinHorizontal(lipgloss.Top, m.listModel.View(), spacer, m.rangesModel.View())
}
