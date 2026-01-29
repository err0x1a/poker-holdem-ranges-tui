package list

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"poker-holdem-ranges-tui/ranges"
)

type item struct {
	title, desc string
	filePath    string
}

func (i item) Title() string { return i.title }

func (i item) Description() string { return i.desc }

func (i item) FilterValue() string { return i.title }

func (i item) FilePath() string { return i.filePath }

type Model struct {
	list list.Model
}

func New(files []string, title string, titleColor string) *Model {
	items := make([]list.Item, 0, len(files))
	for _, file := range files {
		meta, err := ranges.LoadRangeMeta(file)
		if err != nil {
			continue
		}
		items = append(items, item{
			title:    meta.Title,
			desc:     meta.Description,
			filePath: file,
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		BorderForeground(lipgloss.Color("#0496ff")).
		Foreground(lipgloss.Color("#0496ff"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		BorderForeground(lipgloss.Color("#0496ff")).
		Foreground(lipgloss.Color("#7fc8f8"))

	l := list.New(items, delegate, 30, 20)
	if title != "" {
		l.Title = title
		if titleColor != "" {
			l.Styles.Title = l.Styles.Title.
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color(titleColor)).
				Bold(true).
				Padding(0, 1)
		}
	} else {
		l.SetShowTitle(false)
	}
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	model := Model{list: l}
	return &model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width-2, height-2)
}

func (m *Model) View() string {
	return lipgloss.NewStyle().Width(m.list.Width() + 2).Height(m.list.Height() + 2).Render(m.list.View())
}

// SelectedItem returns the currently selected item
func (m *Model) SelectedItem() (title, desc, filePath string) {
	if i, ok := m.list.SelectedItem().(item); ok {
		return i.title, i.desc, i.filePath
	}
	return "", "", ""
}
