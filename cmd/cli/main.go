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

		if _, err := tea.NewProgram(model, tea.WithMouseCellMotion()).Run(); err != nil {
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
	tabIndexByFile   map[string]int
}

// NewMainModel creates a new main model
func NewMainModel(files []string, title string, titleColor string) MainModel {
	return MainModel{
		listModel:      list.New(files, title, titleColor),
		rangesModel:    ranges.New(),
		tabIndexByFile: make(map[string]int),
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
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		// Route keys that the ranges model wants (h/j/k/l cursor, enter/space, esc)
		// directly to it, bypassing the list model to avoid conflicts.
		if m.rangesModel.WantsKey(key) {
			newModel, cmd := m.rangesModel.Update(msg)
			m.rangesModel = newModel.(ranges.Model)
			return m, cmd
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			listW := m.listModel.ViewWidth()
			if msg.X < listW {
				m.listModel.HandleClick(msg.Y)
			} else {
				gridX := msg.X - listW - 2 // 2 = spacer
				if gridX >= 0 {
					m.rangesModel.HandleClick(gridX, msg.Y)
				}
			}
		}
	}

	listCmd = m.listModel.Update(msg)

	_, _, filePath := m.listModel.SelectedItem()
	if filePath != "" && filePath != m.selectedFilePath {
		// Save current tab index before switching
		if m.selectedFilePath != "" && m.rangesModel.HasTabSelector() {
			m.tabIndexByFile[m.selectedFilePath] = m.rangesModel.TabIndex()
		}
		// Preserve hidden actions across range switches
		savedHidden := m.rangesModel.HiddenActions()
		m.selectedFilePath = filePath
		if rf, err := ranges.LoadRangeFile(filePath); err == nil {
			if rf.HasTabs() {
				m.rangesModel = ranges.NewWithTabs(rf.Tabs)
				if savedIndex, ok := m.tabIndexByFile[filePath]; ok {
					m.rangesModel.SetTabIndex(savedIndex)
				}
				// Load opposite data per tab
				oppData, oppLabel := loadTabOpposites(filePath, rf)
				if oppData != nil {
					m.rangesModel.SetOppositeData(oppData, oppLabel)
				}
			} else {
				m.rangesModel = ranges.NewWithRange(rf.Actions, rf.Details)
				// Load single opposite
				if rf.Opposite != nil {
					if opp := ranges.LoadOppositeData(filePath, *rf.Opposite); opp != nil {
						m.rangesModel.SetOppositeData([]*ranges.TabDisplayData{opp}, rf.Opposite.Label())
					}
				}
			}
			m.rangesModel.SetHiddenActions(savedHidden)
		}
	}

	var newRangesModel tea.Model
	newRangesModel, rangesCmd = m.rangesModel.Update(msg)
	m.rangesModel = newRangesModel.(ranges.Model)

	return m, tea.Batch(listCmd, rangesCmd)
}

func (m MainModel) View() string {
	spacer := "  "
	listView := m.listModel.View()
	if m.rangesModel.HasTabSelector() {
		listView = lipgloss.NewStyle().MarginTop(1).Render(listView)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, spacer, m.rangesModel.View())
}

// loadTabOpposites loads opposite data for each tab in a tabbed range file.
// Returns nil if no tab has an opposite reference. Uses the file-level opposite as fallback.
// Caches loaded opposite files by resolved path to avoid redundant disk reads.
func loadTabOpposites(filePath string, rf *ranges.RangeFile) ([]*ranges.TabDisplayData, string) {
	oppData := make([]*ranges.TabDisplayData, len(rf.Tabs))
	hasAny := false
	label := ""
	fileCache := make(map[string]*ranges.RangeFile)

	for i, tr := range rf.Tabs {
		ref := tr.Opposite
		if ref == nil {
			ref = rf.Opposite // fallback to file-level
		}
		if ref == nil {
			continue
		}
		opp := ranges.LoadOppositeDataCached(filePath, *ref, fileCache)
		if opp != nil {
			oppData[i] = opp
			hasAny = true
			if label == "" {
				label = ref.Label()
			}
		}
	}

	if !hasAny {
		return nil, ""
	}
	return oppData, label
}
