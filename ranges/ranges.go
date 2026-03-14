package ranges

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Grid cell dimensions (border + padding + content).
// Used by both View() and HandleClick() to keep rendering and click detection in sync.
const (
	cellW = 7 // NormalBorder(2) + Padding(0,1)(2) + content(3)
	cellH = 3 // NormalBorder(2) + content(1)
)

// allCards is the pre-computed 13x13 hand grid (always the same).
var allCards = generateCards()

// tabDisplayData holds precomputed display data for a single tab
type tabDisplayData struct {
	handDetails map[string][]ActionDetail
	legend      []Action
	details     string
}

// Model represents the state of the ranges view.
type Model struct {
	handDetails map[string][]ActionDetail
	legend      []Action
	details     string
	tabIndex    int
	tabs        []TabRange
	tabCache    []tabDisplayData

	// Cursor navigation
	cursorRow    int
	cursorCol    int
	cursorActive bool

	// Opposite range toggle
	oppositeData    []*tabDisplayData
	showingOpposite bool
	oppositeLabel   string
	savedDisplay    *tabDisplayData
}

// New creates a new model with the generated poker hands.
func New() Model {
	return Model{}
}

// NewWithRange creates a model from actions and details
func NewWithRange(actions []Action, details string) Model {
	return Model{
		handDetails: ActionsToHandDetails(actions),
		legend:      buildLegend(actions),
		details:     details,
	}
}

// NewWithTabs creates a model with multiple tabs, selecting the first one
func NewWithTabs(tabs []TabRange) Model {
	cache := make([]tabDisplayData, len(tabs))
	for i, tr := range tabs {
		cache[i] = tabDisplayData{
			handDetails: ActionsToHandDetails(tr.Actions),
			legend:      buildLegend(tr.Actions),
			details:     tr.Details,
		}
	}

	return Model{
		handDetails: cache[0].handDetails,
		legend:      cache[0].legend,
		details:     cache[0].details,
		tabIndex:    0,
		tabs:        tabs,
		tabCache:    cache,
	}
}

// SetOppositeData sets the opposite range data and label for toggle display.
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
	m.handDetails = d.handDetails
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

// WantsKey returns true if the ranges model wants to handle this key,
// preventing it from reaching the list model.
func (m Model) WantsKey(key string) bool {
	switch key {
	case "h", "j", "k", "l", "up", "down", "left", "right":
		return true
	}
	return false
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "h", "left":
			m.cursorActive = true
			if m.cursorCol > 0 {
				m.cursorCol--
			}
			return m, nil
		case "j", "down":
			m.cursorActive = true
			if m.cursorRow < 12 {
				m.cursorRow++
			}
			return m, nil
		case "k", "up":
			m.cursorActive = true
			if m.cursorRow > 0 {
				m.cursorRow--
			}
			return m, nil
		case "l", "right":
			m.cursorActive = true
			if m.cursorCol < 12 {
				m.cursorCol++
			}
			return m, nil
		case "o":
			m.toggleOpposite()
			return m, nil
		case "tab":
			if len(m.tabs) > 0 && m.tabIndex < len(m.tabs)-1 {
				m.SetTabIndex(m.tabIndex + 1)
			}
			return m, nil
		case "shift+tab":
			if len(m.tabs) > 0 && m.tabIndex > 0 {
				m.SetTabIndex(m.tabIndex - 1)
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
			handDetails: m.handDetails,
			legend:      m.legend,
			details:     m.details,
		}
		m.applyDisplay(opp)
		m.showingOpposite = true
	}
}

// HandleClick handles a mouse click at the given coordinates relative to the grid view.
func (m *Model) HandleClick(x, y int) {
	gridOffsetY := 0
	if len(m.tabs) > 0 || m.currentOpposite() != nil {
		gridOffsetY = 1
	}

	row := (y - gridOffsetY) / cellH
	col := x / cellW

	if row >= 0 && row < 13 && col >= 0 && col < 13 {
		m.cursorRow = row
		m.cursorCol = col
		m.cursorActive = true
	}
}

// View renders the model's state.
func (m Model) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Align(lipgloss.Right).
		Padding(0, 1).
		Margin(0)

	grayStyle := baseStyle.
		BorderForeground(lipgloss.Color("#444444")).
		Foreground(lipgloss.Color("#444444"))

	var allRows []string
	for row := 0; row < 13; row++ {
		var renderedRow []string
		for col := 0; col < 13; col++ {
			card := allCards[row*13+col]
			hand := strings.TrimSpace(card)
			details := m.handDetails[hand]
			// Compute effective color once (with dimming if freq < 100%)
			var style lipgloss.Style
			var color string
			if len(details) > 0 {
				color = details[0].Color
				totalFreq := 0
				for _, d := range details {
					totalFreq += d.Freq
				}
				if totalFreq < 100 {
					color = dimColor(color, float64(totalFreq)/100.0)
				}
				style = baseStyle.
					BorderForeground(lipgloss.Color(color)).
					Foreground(lipgloss.Color(color))
			} else {
				style = grayStyle
			}

			if m.cursorActive && row == m.cursorRow && col == m.cursorCol {
				style = style.
					BorderStyle(lipgloss.ThickBorder()).
					BorderForeground(lipgloss.Color("#FFFFFF"))
			}

			// Underline mixed hands; skip leading space on pairs to avoid visual artifact
			content := card
			if len(details) > 1 {
				ul := lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color(color))
				if card[0] == ' ' {
					content = " " + ul.Render(card[1:])
				} else {
					content = ul.Render(card)
				}
			}

			renderedRow = append(renderedRow, style.Render(content))
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
				lipgloss.JoinHorizontal(lipgloss.Center, tabItems...),
				hintStyle.Render(" ⇥"),
			)
		} else {
			tabSelector = lipgloss.JoinHorizontal(lipgloss.Center, tabItems...)
		}
		if eyeIndicator != "" {
			tabSelector = lipgloss.JoinHorizontal(lipgloss.Center, tabSelector, eyeIndicator)
		}
		grid = lipgloss.JoinVertical(lipgloss.Left, tabSelector, grid)
	} else if eyeIndicator != "" {
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

	// Build right panel: hand details (when cursor active) or strategy details
	panelContent := m.buildDetailsPanel()
	if panelContent != "" {
		detailsStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(1, 2).
			MarginLeft(2)
		if len(m.tabs) > 0 {
			detailsStyle = detailsStyle.MarginTop(1)
		}
		detailsPanel := detailsStyle.Render(panelContent)
		return lipgloss.JoinHorizontal(lipgloss.Top, gridWithLegend, detailsPanel)
	}

	return gridWithLegend
}

// buildDetailsPanel returns the content for the right-side details panel.
// Shows hand action breakdown when cursor is on a hand, otherwise strategy details.
func (m Model) buildDetailsPanel() string {
	if m.cursorActive {
		hand := strings.TrimSpace(allCards[m.cursorRow*13+m.cursorCol])
		if details, ok := m.handDetails[hand]; ok {
			return m.renderHandDetails(hand, details)
		}
	}
	return m.details
}

// renderHandDetails formats a hand's action breakdown for the details panel
func (m Model) renderHandDetails(hand string, details []ActionDetail) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString(titleStyle.Render(hand))
	b.WriteString("\n\n")

	for _, d := range details {
		colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(d.Color))
		line := fmt.Sprintf("%s %-12s %d%%", colorStyle.Render("■"), d.Title, d.Freq)
		if d.RaiseSize != "" {
			line += fmt.Sprintf("  (%s)", d.RaiseSize)
		}
		b.WriteString(line + "\n")
	}

	if m.details != "" {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("─────────────────"))
		b.WriteString("\n\n")
		b.WriteString(m.details)
	}

	return b.String()
}

// generateCards creates the 13x13 poker hand grid.
func generateCards() []string {
	ranks := []string{"A", "K", "Q", "J", "T", "9", "8", "7", "6", "5", "4", "3", "2"}
	cards := make([]string, 0, 13*13)
	for i, rankI := range ranks {
		for j, rankJ := range ranks {
			var hand string
			if i == j {
				hand = " " + rankI + rankJ + ""
			} else if i < j {
				hand = rankI + rankJ + "s"
			} else {
				hand = rankJ + rankI + "o"
			}
			cards = append(cards, hand)
		}
	}
	return cards
}

// Generate returns the pre-computed hand grid (kept for backward compat).
func Generate() []string {
	return allCards
}

// dimColor scales a hex color (#RRGGBB) by factor (0.0–1.0) to simulate reduced opacity.
// Uses a gentle range: factor 0.0 maps to 60% brightness, factor 1.0 maps to 100%.
func dimColor(hex string, factor float64) string {
	if len(hex) != 7 || hex[0] != '#' {
		return hex
	}
	// Map factor from [0,1] to [0.35, 1.0]
	scaled := 0.35 + factor*0.65
	r, _ := strconv.ParseUint(hex[1:3], 16, 8)
	g, _ := strconv.ParseUint(hex[3:5], 16, 8)
	b, _ := strconv.ParseUint(hex[5:7], 16, 8)
	r = uint64(float64(r) * scaled)
	g = uint64(float64(g) * scaled)
	b = uint64(float64(b) * scaled)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
