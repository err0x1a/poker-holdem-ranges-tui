package ranges

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RangeMeta contains only title and description for menu display
type RangeMeta struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	FilePath    string `yaml:"-"`
}

// Action represents a single action type with its color and hands
type Action struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Color       string   `yaml:"color"`
	Hands       []string `yaml:"hands"`
	AddHands    []string `yaml:"add_hands"`
	RemoveHands []string `yaml:"remove_hands"`
}

// OppositeRef points to an opposite range file (and optionally a specific tab)
type OppositeRef struct {
	File string `yaml:"file"`
	Tab  string `yaml:"tab"`
}

// TabRange represents a single tab with its own actions and details
type TabRange struct {
	Tab      string       `yaml:"tab"`
	Base     string       `yaml:"base"`
	Details  string       `yaml:"details"`
	Actions  []Action     `yaml:"actions"`
	Opposite *OppositeRef `yaml:"opposite"`
}

// RangeFile represents the full YAML structure
type RangeFile struct {
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Details     string       `yaml:"details"`
	Actions     []Action     `yaml:"actions"`
	Tabs        []TabRange   `yaml:"tab_ranges"`
	Opposite    *OppositeRef `yaml:"opposite"`
}

// HasTabs returns true if the file uses the multi-tab format
func (rf *RangeFile) HasTabs() bool {
	return len(rf.Tabs) > 0
}

// HandAction stores color for a hand
type HandAction struct {
	Color string
}

// LoadRangeMetas loads only title/description from all YAML files in a directory
func LoadRangeMetas(dir string) ([]RangeMeta, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	metas := make([]RangeMeta, 0, len(files))
	for _, file := range files {
		meta, err := loadMeta(file)
		if err != nil {
			continue // skip invalid files
		}
		meta.FilePath = file
		metas = append(metas, meta)
	}

	return metas, nil
}

// loadMeta loads only title and description from a YAML file
func loadMeta(path string) (RangeMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RangeMeta{}, err
	}

	var meta RangeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return RangeMeta{}, err
	}

	return meta, nil
}

// LoadRangeMeta loads title/description from a single file (exported)
func LoadRangeMeta(path string) (RangeMeta, error) {
	return loadMeta(path)
}

// LoadRangeFile loads the complete range file (on-demand when selected)
func LoadRangeFile(path string) (*RangeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rf RangeFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, err
	}

	if rf.HasTabs() {
		resolveTabs(rf.Tabs)
	}

	return &rf, nil
}

// resolveTabs resolves base inheritance for tab ranges
func resolveTabs(tabs []TabRange) {
	byTab := make(map[string]int, len(tabs))
	for i, tr := range tabs {
		byTab[tr.Tab] = i
	}

	for i, tr := range tabs {
		if tr.Base == "" {
			continue
		}
		baseIdx, ok := byTab[tr.Base]
		if !ok {
			continue
		}
		base := &tabs[baseIdx]

		// Inherit details if empty
		if tr.Details == "" {
			tabs[i].Details = base.Details
		}

		// Build map of child actions by name
		childActions := make(map[string]*Action, len(tr.Actions))
		for j := range tr.Actions {
			childActions[tr.Actions[j].Name] = &tr.Actions[j]
		}

		// Start with base actions, apply overrides
		var resolved []Action
		for _, baseAction := range base.Actions {
			child, hasOverride := childActions[baseAction.Name]
			if !hasOverride {
				resolved = append(resolved, baseAction)
				continue
			}

			merged := Action{
				Name:  baseAction.Name,
				Title: baseAction.Title,
				Color: baseAction.Color,
			}
			if child.Title != "" {
				merged.Title = child.Title
			}
			if child.Color != "" {
				merged.Color = child.Color
			}

			// Start with base hands, remove then add
			hands := make(map[string]bool, len(baseAction.Hands))
			for _, h := range baseAction.Hands {
				hands[h] = true
			}
			for _, h := range child.RemoveHands {
				delete(hands, h)
			}
			for _, h := range child.AddHands {
				hands[h] = true
			}

			// Preserve order: base hands (minus removed), then added
			for _, h := range baseAction.Hands {
				if hands[h] {
					merged.Hands = append(merged.Hands, h)
				}
			}
			for _, h := range child.AddHands {
				if hands[h] {
					merged.Hands = append(merged.Hands, h)
					delete(hands, h) // avoid duplicates
				}
			}

			resolved = append(resolved, merged)
			delete(childActions, baseAction.Name)
		}

		// Add new actions not in base
		for _, child := range tr.Actions {
			if _, isNew := childActions[child.Name]; isNew {
				hands := child.Hands
				if hands == nil {
					hands = child.AddHands
				}
				resolved = append(resolved, Action{
					Name:  child.Name,
					Title: child.Title,
					Color: child.Color,
					Hands: hands,
				})
			}
		}

		// Deduplicate: a hand should only appear in one action.
		// Collect all hands per action, last action wins.
		handOwner := make(map[string]string)
		for _, a := range resolved {
			for _, h := range a.Hands {
				handOwner[h] = a.Name
			}
		}
		for j, a := range resolved {
			var filtered []string
			for _, h := range a.Hands {
				if handOwner[h] == a.Name {
					filtered = append(filtered, h)
				}
			}
			resolved[j].Hands = filtered
		}

		tabs[i].Actions = resolved
	}
}

// ActionsToHandColors converts a slice of actions to a map of hand -> color
func ActionsToHandColors(actions []Action) map[string]string {
	colors := make(map[string]string)
	for _, action := range actions {
		for _, hand := range action.Hands {
			colors[hand] = action.Color
		}
	}
	return colors
}

// ToHandColors converts a RangeFile to a map of hand -> color for rendering
func (rf *RangeFile) ToHandColors() map[string]string {
	return ActionsToHandColors(rf.Actions)
}

// GetLegend returns actions that have at least one hand
func (rf *RangeFile) GetLegend() []Action {
	return filterEmptyActions(rf.Actions)
}

// filterEmptyActions returns only actions with hands
func filterEmptyActions(actions []Action) []Action {
	var filtered []Action
	for _, a := range actions {
		if len(a.Hands) > 0 {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// TabDisplayData holds precomputed display data for a single tab (exported for opposite loading)
type TabDisplayData = tabDisplayData

// LoadOppositeData loads the opposite range data given a base file path and an OppositeRef.
// Returns nil if the opposite file cannot be loaded.
func LoadOppositeData(basePath string, ref OppositeRef) *tabDisplayData {
	dir := filepath.Dir(basePath)
	oppPath := filepath.Join(dir, ref.File)
	oppPath = filepath.Clean(oppPath)

	rf, err := LoadRangeFile(oppPath)
	if err != nil {
		return nil
	}

	var actions []Action
	var details string

	if rf.HasTabs() {
		tab := findTab(rf.Tabs, ref.Tab)
		if tab == nil {
			return nil
		}
		actions = tab.Actions
		details = tab.Details
	} else {
		actions = rf.Actions
		details = rf.Details
	}

	return &tabDisplayData{
		handColors: ActionsToHandColors(actions),
		legend:     filterEmptyActions(actions),
		details:    details,
	}
}

// findTab finds a tab by name, or returns the first tab if name is empty
func findTab(tabs []TabRange, name string) *TabRange {
	if len(tabs) == 0 {
		return nil
	}
	if name == "" {
		return &tabs[0]
	}
	for i := range tabs {
		if tabs[i].Tab == name {
			return &tabs[i]
		}
	}
	return nil
}
