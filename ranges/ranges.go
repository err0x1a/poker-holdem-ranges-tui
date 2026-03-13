package ranges

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tabDisplayData holds precomputed display data for a single tab
type tabDisplayData struct {
	handColors map[string]string
	legend     []Action
	details    string
}

// Model represents the state of the ranges view.
type Model struct {
	cards      []string
	handColors map[string]string
	legend     []Action
	details    string
	tabIndex   int
	tabs       []TabRange
	tabCache   []tabDisplayData

	// Opposite range toggle
	oppositeData    []*tabDisplayData // one per tab (or single element for no-tab files)
	showingOpposite bool
	oppositeLabel   string
	savedDisplay    *tabDisplayData
}

// New creates a new model with the generated poker hands.
func New() Model {
	return Model{
		cards:      Generate(),
		handColors: make(map[string]string),
		legend:     nil,
		details:    "",
	}
}

// NewWithRange creates a model with a specific range loaded
func NewWithRange(handColors map[string]string, legend []Action, details string) Model {
	return Model{
		cards:      Generate(),
		handColors: handColors,
		legend:     legend,
		details:    details,
	}
}

// NewWithTabs creates a model with multiple tabs, selecting the first one
func NewWithTabs(tabs []TabRange) Model {
	cache := make([]tabDisplayData, len(tabs))
	for i, tr := range tabs {
		cache[i] = tabDisplayData{
			handColors: ActionsToHandColors(tr.Actions),
			legend:     filterEmptyActions(tr.Actions),
			details:    tr.Details,
		}
	}

	return Model{
		cards:      Generate(),
		handColors: cache[0].handColors,
		legend:     cache[0].legend,
		details:    cache[0].details,
		tabIndex:   0,
		tabs:       tabs,
		tabCache:   cache,
	}
}

// SetOppositeData sets the opposite range data and label for toggle display.
// For tab files, pass one *tabDisplayData per tab; for non-tab files, pass a single element.
func (m *Model) SetOppositeData(data []*tabDisplayData, label string) {
	m.oppositeData = data
	m.oppositeLabel = label
}

// HasTabSelector returns true if the model has a tab selector bar
func (m Model) HasTabSelector() bool {
	return len(m.tabs) > 0
}

// TabIndex returns the currently selected tab index
func (m Model) TabIndex() int {
	return m.tabIndex
}

// applyDisplay updates the model's display fields from a tabDisplayData
func (m *Model) applyDisplay(d *tabDisplayData) {
	m.handColors = d.handColors
	m.legend = d.legend
	m.details = d.details
}

// SetTabIndex sets the selected tab index and updates display data
func (m *Model) SetTabIndex(index int) {
	if index >= 0 && index < len(m.tabs) {
		m.showingOpposite = false
		m.savedDisplay = nil
		m.tabIndex = index
		m.applyDisplay(&m.tabCache[index])
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "o":
			m.toggleOpposite()
			return m, nil
		case "left":
			if len(m.tabs) > 0 && m.tabIndex > 0 {
				m.SetTabIndex(m.tabIndex - 1)
			}
			return m, nil
		case "right":
			if len(m.tabs) > 0 && m.tabIndex < len(m.tabs)-1 {
				m.SetTabIndex(m.tabIndex + 1)
			}
			return m, nil
		}
	}
	return m, nil
}

// currentOpposite returns the opposite data for the current tab, or nil
func (m Model) currentOpposite() *tabDisplayData {
	idx := 0
	if len(m.tabs) > 0 {
		idx = m.tabIndex
	}
	if idx < len(m.oppositeData) {
		return m.oppositeData[idx]
	}
	return nil
}

// toggleOpposite switches between the original and opposite range display
func (m *Model) toggleOpposite() {
	opp := m.currentOpposite()
	if opp == nil {
		return
	}

	if m.showingOpposite {
		if m.savedDisplay != nil {
			m.applyDisplay(m.savedDisplay)
			m.savedDisplay = nil
		}
		m.showingOpposite = false
	} else {
		m.savedDisplay = &tabDisplayData{
			handColors: m.handColors,
			legend:     m.legend,
			details:    m.details,
		}
		m.applyDisplay(opp)
		m.showingOpposite = true
	}
}

// ActionType represents the type of action for a hand
// Deprecated: Use dynamic colors from YAML instead
type ActionType int

// View renders the model's state.
func (m Model) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Align(lipgloss.Right).
		Padding(0, 1).
		Margin(0)

	// Default gray style for all hands
	grayStyle := baseStyle.
		BorderForeground(lipgloss.Color("#666666")).
		Foreground(lipgloss.Color("#666666"))

	var allRows []string
	for i := 0; i < len(m.cards); i += 13 {
		end := i + 13
		end = min(end, len(m.cards))
		rowCards := m.cards[i:end]

		var renderedRow []string
		for _, card := range rowCards {
			hand := strings.TrimSpace(card)
			color, hasColor := m.handColors[hand]
			var style lipgloss.Style
			if hasColor {
				style = baseStyle.
					BorderForeground(lipgloss.Color(color)).
					Foreground(lipgloss.Color(color))
			} else {
				style = grayStyle
			}
			renderedRow = append(renderedRow, style.Render(card))
		}
		allRows = append(allRows, lipgloss.JoinHorizontal(lipgloss.Top, renderedRow...))
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, allRows...)

	// Build opposite eye indicator
	var eyeIndicator string
	if m.currentOpposite() != nil {
		eyeStyle := lipgloss.NewStyle().Padding(0, 1)
		if m.showingOpposite {
			eyeStyle = eyeStyle.Foreground(lipgloss.Color("#FFD166")).Bold(true)
			eyeIndicator = eyeStyle.Render("👁")
		} else {
			eyeStyle = eyeStyle.Foreground(lipgloss.Color("#555555"))
			eyeIndicator = eyeStyle.Render("👁")
		}
	}

	// Add tab selector above grid if present
	if len(m.tabs) > 0 {
		selectedStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#4488FF")).
			Padding(0, 1)
		dimStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(0, 1)

		var tabItems []string
		for i, tr := range m.tabs {
			if i == m.tabIndex {
				tabItems = append(tabItems, selectedStyle.Render(tr.Tab))
			} else {
				tabItems = append(tabItems, dimStyle.Render(tr.Tab))
			}
		}
		var tabSelector string
		if len(m.tabs) > 1 {
			hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
			tabSelector = lipgloss.JoinHorizontal(lipgloss.Center,
				hintStyle.Render("← "),
				lipgloss.JoinHorizontal(lipgloss.Center, tabItems...),
				hintStyle.Render(" →"),
			)
		} else {
			tabSelector = lipgloss.JoinHorizontal(lipgloss.Center, tabItems...)
		}
		if eyeIndicator != "" {
			tabSelector = lipgloss.JoinHorizontal(lipgloss.Center, tabSelector, eyeIndicator)
		}
		grid = lipgloss.JoinVertical(lipgloss.Left, tabSelector, grid)
	} else if eyeIndicator != "" {
		// No tabs — put eye indicator on its own line above grid
		grid = lipgloss.JoinVertical(lipgloss.Right, eyeIndicator, grid)
	}

	// Build legend
	var gridWithLegend string
	if len(m.legend) > 0 {
		var legendItems []string
		for _, action := range m.legend {
			legendStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(action.Color)).
				Bold(true)
			legendItems = append(legendItems, legendStyle.Render("■ "+action.Title))
		}
		legend := lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(legendItems, "  "))
		gridWithLegend = lipgloss.JoinVertical(lipgloss.Left, grid, "", legend)
	} else {
		gridWithLegend = grid
	}

	// Add details panel on the right
	if m.details != "" {
		detailsStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(1, 2).
			MarginLeft(2)
		if len(m.tabs) > 0 {
			detailsStyle = detailsStyle.MarginTop(1)
		}
		detailsPanel := detailsStyle.Render(m.details)
		return lipgloss.JoinHorizontal(lipgloss.Top, gridWithLegend, detailsPanel)
	}

	return gridWithLegend
}

// Generate creates a slice of all 169 possible poker hands.
func Generate() []string {
	ranks := []string{"A", "K", "Q", "J", "T", "9", "8", "7", "6", "5", "4", "3", "2"}
	cards := make([]string, 0, 13*13)
	for i, rankI := range ranks {
		for j, rankJ := range ranks {
			var hand string
			if i == j {
				// Pocket pairs
				hand = " " + rankI + rankJ + ""
			} else if i < j {
				// Suited hands
				hand = rankI + rankJ + "s"
			} else {
				// Off-suit hands
				hand = rankJ + rankI + "o"
			}
			cards = append(cards, hand)
		}
	}
	return cards
}
