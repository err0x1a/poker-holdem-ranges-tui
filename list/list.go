package list

import (
	"github.com/charmbracelet/bubbles/key"
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
	list      list.Model
	hasTitle  bool
	itemSlotH int // height + spacing per item
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

	// Rebind navigation: arrows/hjkl go to grid, list uses ctrl+n/p and pgup/pgdown
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("ctrl+p"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("ctrl+n"))
	l.KeyMap.PrevPage = key.NewBinding(key.WithKeys("pgup"))
	l.KeyMap.NextPage = key.NewBinding(key.WithKeys("pgdown"))
	l.KeyMap.Quit = key.NewBinding(key.WithKeys()) // disable esc quit

	model := Model{
		list:      l,
		hasTitle:  title != "",
		itemSlotH: 3, // defaultHeight(2) + defaultSpacing(1)
	}
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

// HandleClick selects the list item at the given screen Y coordinate.
func (m *Model) HandleClick(y int) {
	// Header offset: title(2 lines) or no-title(0), plus filter bar area
	headerH := 0
	if m.hasTitle {
		headerH = 2 // title line + blank line
	}

	itemY := y - headerH
	if itemY < 0 {
		return
	}

	idx := m.list.Paginator.Page*m.list.Paginator.PerPage + itemY/m.itemSlotH
	if idx >= 0 && idx < len(m.list.Items()) {
		m.list.Select(idx)
	}
}

// IsFiltering returns true when the list filter input is active
func (m *Model) IsFiltering() bool {
	return m.list.FilterState() == list.Filtering
}

// ViewWidth returns the total rendered width of the list view
func (m *Model) ViewWidth() int {
	return m.list.Width() + 2
}

// SelectedItem returns the currently selected item
func (m *Model) SelectedItem() (title, desc, filePath string) {
	if i, ok := m.list.SelectedItem().(item); ok {
		return i.title, i.desc, i.filePath
	}
	return "", "", ""
}

// SelectByPath selects the item matching the given file path, returns true if found
func (m *Model) SelectByPath(filePath string) bool {
	for i, li := range m.list.Items() {
		if it, ok := li.(item); ok && it.filePath == filePath {
			m.list.Select(i)
			return true
		}
	}
	return false
}
