package ranges

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the state of the ranges view.
type Model struct {
	cards      []string
	handColors map[string]string
	legend     []Action
	details    string
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

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// No updates to handle in this model yet
	return m, nil
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
				hand = " " + rankI + rankJ + ""
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
