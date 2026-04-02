package ranges

import (
	"fmt"
	"path/filepath"
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
	sideranges  *Sideranges
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

	// Legend filtering
	hiddenActions map[string]bool // keyed by action title

	// Sideranges navigation
	sideranges         *Sideranges
	siderangeIndex     int
	siderangeFocused   bool
	showingSiderange   bool
	savedSiderangeDisp *tabDisplayData
	activeSiderangeIdx int // index of loaded siderange, -1 = none
	filePath           string
}

// New creates a new model with the generated poker hands.
func New() Model {
	return Model{}
}

// NewWithRange creates a model from actions and details
func NewWithRange(actions []Action, details string, sideranges *Sideranges) Model {
	return Model{
		handDetails: ActionsToHandDetails(actions),
		legend:      buildLegend(actions),
		details:     details,
		sideranges:  sideranges,
	}
}

// NewWithTabs creates a model with multiple tabs, selecting the first one
func NewWithTabs(tabs []TabRange, fileSideranges *Sideranges) Model {
	cache := make([]tabDisplayData, len(tabs))
	for i, tr := range tabs {
		sr := tr.Sideranges
		if sr == nil {
			sr = fileSideranges
		}
		cache[i] = tabDisplayData{
			handDetails: ActionsToHandDetails(tr.Actions),
			legend:      buildLegend(tr.Actions),
			details:     tr.Details,
			sideranges:  sr,
		}
	}

	return Model{
		handDetails: cache[0].handDetails,
		legend:      cache[0].legend,
		details:     cache[0].details,
		sideranges:  cache[0].sideranges,
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

// HiddenActions returns the current hidden actions filter.
func (m Model) HiddenActions() map[string]bool {
	return m.hiddenActions
}

// SetHiddenActions restores a previously saved hidden actions filter.
func (m *Model) SetHiddenActions(h map[string]bool) {
	m.hiddenActions = h
}

// HasTabSelector returns true if the model has a tab selector bar
func (m Model) HasTabSelector() bool {
	return len(m.tabs) > 0
}

// TabIndex returns the currently selected tab index
func (m Model) TabIndex() int {
	return m.tabIndex
}

// TabName returns the name of the currently selected tab
func (m Model) TabName() string {
	if m.tabIndex >= 0 && m.tabIndex < len(m.tabs) {
		return m.tabs[m.tabIndex].Tab
	}
	return ""
}

// applyDisplay updates the model's display fields from a tabDisplayData
func (m *Model) applyDisplay(d *tabDisplayData) {
	m.handDetails = d.handDetails
	m.legend = d.legend
	m.details = d.details
	m.sideranges = d.sideranges
}

// SetTabIndex sets the selected tab index and updates display data
func (m *Model) SetTabIndex(index int) {
	if index >= 0 && index < len(m.tabs) {
		m.showingOpposite = false
		m.savedDisplay = nil
		m.showingSiderange = false
		m.savedSiderangeDisp = nil
		m.activeSiderangeIdx = -1
		m.tabIndex = index
		m.applyDisplay(&m.tabCache[index])
		m.siderangeIndex = 0
		m.siderangeFocused = false
	}
}

// SetTabByName sets the selected tab by name, returns true if found
func (m *Model) SetTabByName(name string) bool {
	for i, tr := range m.tabs {
		if tr.Tab == name {
			m.SetTabIndex(i)
			return true
		}
	}
	return false
}

// SetFilePath stores the file path for resolving relative siderange paths
func (m *Model) SetFilePath(path string) {
	m.filePath = path
}

// WantsKey returns true if the ranges model wants to handle this key,
// preventing it from reaching the list model.
func (m Model) WantsKey(key string) bool {
	switch key {
	case "h", "j", "k", "l", "up", "down", "left", "right":
		return true
	case "s":
		return m.hasSideranges()
	case "enter":
		return m.siderangeFocused
	case "esc":
		return m.siderangeFocused || m.showingSiderange
	}
	return false
}

func (m Model) hasSideranges() bool {
	return m.sideranges != nil && len(m.sideranges.Items) > 0
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		key := keyMsg.String()

		// Siderange-focused input
		if m.siderangeFocused {
			switch key {
			case "j", "down":
				if m.siderangeIndex < len(m.sideranges.Items)-1 {
					m.siderangeIndex++
				}
				return m, nil
			case "k", "up":
				if m.siderangeIndex > 0 {
					m.siderangeIndex--
				}
				return m, nil
			case "enter":
				(&m).loadSiderange()
				m.siderangeFocused = false
				return m, nil
			case "esc":
				m.siderangeFocused = false
				return m, nil
			case "s":
				if m.showingSiderange {
					(&m).restoreSiderange()
				}
				m.siderangeFocused = false
				return m, nil
			}
			return m, nil
		}

		switch key {
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
		case "s":
			if m.hasSideranges() {
				m.siderangeFocused = true
				if !m.showingSiderange {
					m.siderangeIndex = 0
				}
			}
			return m, nil
		case "esc":
			if m.showingSiderange {
				(&m).restoreSiderange()
			}
			return m, nil
		case "ctrl+o":
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

// loadSiderange loads a siderange file's data into the grid, preserving tabs and sideranges panel
func (m *Model) loadSiderange() {
	if !m.hasSideranges() || m.siderangeIndex >= len(m.sideranges.Items) {
		return
	}
	item := m.sideranges.Items[m.siderangeIndex]
	file := item.File
	if m.filePath != "" && !filepath.IsAbs(file) {
		file = filepath.Clean(filepath.Join(filepath.Dir(m.filePath), file))
	}

	rf, err := LoadRangeFile(file)
	if err != nil {
		return
	}

	var actions []Action
	var details string
	if rf.HasTabs() {
		tab := findTab(rf.Tabs, item.Tab)
		if tab == nil {
			return
		}
		actions = tab.Actions
		details = tab.Details
	} else {
		actions = rf.Actions
		details = rf.Details
	}

	// Save current display only if not already showing a siderange
	if !m.showingSiderange {
		m.savedSiderangeDisp = &tabDisplayData{
			handDetails: m.handDetails,
			legend:      m.legend,
			details:     m.details,
			sideranges:  m.sideranges,
		}
	}

	m.handDetails = ActionsToHandDetails(actions)
	m.legend = buildLegend(actions)
	m.details = details
	m.showingSiderange = true
	m.activeSiderangeIdx = m.siderangeIndex
}

// restoreSiderange restores the display from before siderange was loaded
func (m *Model) restoreSiderange() {
	if m.savedSiderangeDisp != nil {
		m.applyDisplay(m.savedSiderangeDisp)
		m.savedSiderangeDisp = nil
	}
	m.showingSiderange = false
	m.activeSiderangeIdx = -1
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
			sideranges:  m.sideranges,
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

	// Check for siderange clicks in the right panel area
	gridW := 13 * cellW
	if m.hasSideranges() && x >= gridW {
		if m.handleSiderangeClick(y) {
			return
		}
	}

	// Check if click is on the legend row
	// Legend is at: gridOffsetY + 13*cellH (grid) + 1 (empty line)
	legendY := gridOffsetY + 13*cellH + 1
	if y == legendY && len(m.legend) > 0 {
		m.handleLegendClick(x)
		return
	}

	row := (y - gridOffsetY) / cellH
	col := x / cellW

	if row >= 0 && row < 13 && col >= 0 && col < 13 {
		m.cursorRow = row
		m.cursorCol = col
		m.cursorActive = true
	}
}

// handleSiderangeClick checks if a click hit a siderange item in the right panel.
// Returns true if a siderange was clicked.
func (m *Model) handleSiderangeClick(y int) bool {
	// Panel Y offset: border(1) + padding(1) = 2, plus marginTop(1) if tabs
	panelContentY := 2
	if len(m.tabs) > 0 {
		panelContentY = 3
	}

	// Count lines in the content before sideranges (details or hand details)
	var contentLines int
	if m.cursorActive {
		hand := strings.TrimSpace(allCards[m.cursorRow*13+m.cursorCol])
		if details, ok := m.handDetails[hand]; ok {
			contentLines = strings.Count(m.renderHandDetails(hand, details), "\n")
		}
	}
	if contentLines == 0 {
		contentLines = strings.Count(m.details, "\n")
	}

	// Siderange header: \n + separator + \n\n + title + \n\n = 5 lines
	const siderangeHeaderLines = 5
	firstItemLine := contentLines + siderangeHeaderLines
	clickedLine := y - panelContentY
	itemIdx := clickedLine - firstItemLine

	nItems := len(m.sideranges.Items)
	if itemIdx >= 0 && itemIdx < nItems {
		m.siderangeIndex = itemIdx
		m.loadSiderange()
		m.siderangeFocused = false
		return true
	}
	return false
}

// handleLegendClick toggles action visibility based on click position in the legend.
func (m *Model) handleLegendClick(x int) {
	// Reconstruct legend item positions: "■ Title  ■ Title  ■ Title"
	// Each item is "■ " + title, separated by "  " (2 spaces)
	pos := 0
	for _, action := range m.legend {
		itemLen := len([]rune("■ " + action.Title))
		if x >= pos && x < pos+itemLen {
			if m.hiddenActions == nil {
				m.hiddenActions = make(map[string]bool)
			}
			m.hiddenActions[action.Title] = !m.hiddenActions[action.Title]
			return
		}
		pos += itemLen + 2 // +2 for "  " separator
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
			allDetails := m.handDetails[hand]
			// Filter out hidden actions
			var details []ActionDetail
			for _, d := range allDetails {
				if !m.hiddenActions[d.Title] {
					details = append(details, d)
				}
			}
			isCursor := m.cursorActive && row == m.cursorRow && col == m.cursorCol

			if len(details) > 0 {
				color := details[0].Color
				totalFreq := 0
				for _, d := range details {
					totalFreq += d.Freq
				}

				// Use split border progression when not 100% or when hand has multiple actions
				if totalFreq < 100 || len(details) >= 2 {
					renderedRow = append(renderedRow, renderSplitBorderCell(card, details, totalFreq, isCursor))
				} else {
					style := baseStyle.
						BorderForeground(lipgloss.Color(color)).
						Foreground(lipgloss.Color(color))
					if isCursor {
						style = style.Background(lipgloss.Color("#333333"))
					}
					renderedRow = append(renderedRow, style.Render(card))
				}
			} else {
				style := grayStyle
				if isCursor {
					style = style.Background(lipgloss.Color("#333333"))
				}
				renderedRow = append(renderedRow, style.Render(card))
			}
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
		tabSelector := lipgloss.JoinHorizontal(lipgloss.Center, tabItems...)
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
			if m.hiddenActions[action.Title] {
				dimStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#555555")).
					Strikethrough(true)
				legendItems = append(legendItems, dimStyle.Render("■ "+action.Title))
			} else {
				legendStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(action.Color)).
					Bold(true)
				legendItems = append(legendItems, legendStyle.Render("■ "+action.Title))
			}
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
// Appends sideranges list below the main content when available.
func (m Model) buildDetailsPanel() string {
	var content string
	if m.cursorActive {
		hand := strings.TrimSpace(allCards[m.cursorRow*13+m.cursorCol])
		if details, ok := m.handDetails[hand]; ok {
			content = m.renderHandDetails(hand, details)
		}
	}
	if content == "" {
		content = m.details
	}

	if m.hasSideranges() {
		content += m.renderSideranges()
	}

	return content
}

// renderSideranges formats the sideranges list for the details panel
func (m Model) renderSideranges() string {
	var b strings.Builder
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4488FF")).Bold(true)

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("─────────────────"))
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render(m.sideranges.Title))
	b.WriteString("\n\n")

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	for i, item := range m.sideranges.Items {
		isActive := m.showingSiderange && i == m.activeSiderangeIdx
		isFocused := m.siderangeFocused && i == m.siderangeIndex
		switch {
		case isActive:
			b.WriteString(activeStyle.Render("  ▸ " + item.Label))
		case isFocused:
			b.WriteString(focusedStyle.Render("  ▸ " + item.Label))
		default:
			b.WriteString(normalStyle.Render("  ▸ " + item.Label))
		}
		b.WriteString("\n")
	}

	return b.String()
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

// renderSplitBorderCell renders a cell with a left-to-right color progression on
// the top and bottom borders based on the frequency split between actions.
func renderSplitBorderCell(card string, details []ActionDetail, totalFreq int, selected bool) string {
	const innerW = 5 // padding(1) + content(3) + padding(1)
	const foldColor = "#444444"

	foldFreq := 100 - totalFreq
	if foldFreq < 0 {
		foldFreq = 0
	}

	// Build segment list: actions + fold remainder
	type segment struct {
		color string
		chars int
	}
	segs := make([]segment, 0, len(details)+1)
	freqs := make([]int, 0, len(details)+1)
	for _, d := range details {
		segs = append(segs, segment{d.Color, 0})
		freqs = append(freqs, d.Freq)
	}
	if foldFreq > 0 {
		segs = append(segs, segment{foldColor, 0})
		freqs = append(freqs, foldFreq)
	}

	// Distribute innerW chars proportionally
	assigned := 0
	for i, f := range freqs {
		segs[i].chars = int(float64(innerW)*float64(f)/100.0 + 0.5)
		assigned += segs[i].chars
	}
	for assigned != innerW {
		maxIdx := 0
		for i := range freqs {
			if freqs[i] > freqs[maxIdx] {
				maxIdx = i
			}
		}
		if assigned > innerW {
			segs[maxIdx].chars--
			assigned--
		} else {
			segs[maxIdx].chars++
			assigned++
		}
	}

	// Build horizontal border string once, reuse for top and bottom
	var hb strings.Builder
	for _, seg := range segs {
		if seg.chars > 0 {
			s := lipgloss.NewStyle().Foreground(lipgloss.Color(seg.color))
			hb.WriteString(s.Render(strings.Repeat("─", seg.chars)))
		}
	}
	hBorder := hb.String()

	sf := lipgloss.NewStyle().Foreground(lipgloss.Color(segs[0].color))
	sl := lipgloss.NewStyle().Foreground(lipgloss.Color(segs[len(segs)-1].color))
	st := lipgloss.NewStyle().Foreground(lipgloss.Color(details[0].Color))

	top := sf.Render("┌") + hBorder + sl.Render("┐")
	pad := " "
	if selected {
		bg := lipgloss.Color("#333333")
		pad = lipgloss.NewStyle().Background(bg).Render(" ")
		st = st.Background(bg)
	}
	mid := sf.Render("│") + pad + st.Render(card) + pad + sl.Render("│")
	bot := sf.Render("└") + hBorder + sl.Render("┘")

	return top + "\n" + mid + "\n" + bot
}


