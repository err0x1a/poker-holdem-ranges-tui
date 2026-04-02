package ranges

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// HandEntry represents a hand in an action's hands list.
// Can be a plain string (100% frequency) or {hand, freq} for mixed strategies.
type HandEntry struct {
	Hand string `yaml:"hand"`
	Freq int    `yaml:"freq"` // 0 means 100% (normal hand)
}

// UnmarshalYAML handles both string ("AA") and mapping ({hand: AQo, freq: 50}) forms.
func (h *HandEntry) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		h.Hand = value.Value
		h.Freq = 0
		return nil
	}
	type rawHandEntry HandEntry
	var raw rawHandEntry
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*h = HandEntry(raw)
	return nil
}

// ActionDetail holds one action's contribution to a hand
type ActionDetail struct {
	Title     string
	Color     string
	Freq      int
	RaiseSize string
}

// RangeMeta contains only title and description for menu display
type RangeMeta struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	FilePath    string `yaml:"-"`
}

// Action represents a single action type with its color and hands
type Action struct {
	Name        string      `yaml:"name"`
	Title       string      `yaml:"title"`
	Color       string      `yaml:"color"`
	RaiseSize   string      `yaml:"raise_size,omitempty"` // e.g. "2.5x", "3bb", "all-in"
	Hands       []HandEntry `yaml:"hands"`
	AddHands    []string    `yaml:"add_hands"`
	RemoveHands []string    `yaml:"remove_hands"`
}

// SiderangeItem represents a single siderange reference
type SiderangeItem struct {
	Label string `yaml:"label"`
	File  string `yaml:"file"`
	Tab   string `yaml:"tab"`
}

// Sideranges holds a titled group of siderange references
type Sideranges struct {
	Title string          `yaml:"title"`
	Items []SiderangeItem `yaml:"items"`
}

// OppositeRef points to an opposite range file (and optionally a specific tab)
type OppositeRef struct {
	File string `yaml:"file"`
	Tab  string `yaml:"tab"`
}

// Label returns a display label like "filename.yaml [tab]"
func (r OppositeRef) Label() string {
	label := filepath.Base(r.File)
	if r.Tab != "" {
		label += " [" + r.Tab + "]"
	}
	return label
}

// TabRange represents a single tab with its own actions and details
type TabRange struct {
	Tab        string       `yaml:"tab"`
	Base       string       `yaml:"base"`
	Details    string       `yaml:"details"`
	Actions    []Action     `yaml:"actions"`
	Opposite   *OppositeRef `yaml:"opposite"`
	Sideranges *Sideranges  `yaml:"sideranges"`
}

// RangeFile represents the full YAML structure
type RangeFile struct {
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Details     string       `yaml:"details"`
	Actions     []Action     `yaml:"actions"`
	Tabs        []TabRange   `yaml:"tab_ranges"`
	TabStyle    string       `yaml:"tab_style"`
	Opposite    *OppositeRef `yaml:"opposite"`
	Sideranges  *Sideranges  `yaml:"sideranges"`
}

// HasTabs returns true if the file uses the multi-tab format
func (rf *RangeFile) HasTabs() bool {
	return len(rf.Tabs) > 0
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
			continue
		}
		meta.FilePath = file
		metas = append(metas, meta)
	}

	return metas, nil
}

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

		if tr.Details == "" {
			tabs[i].Details = base.Details
		}

		childActions := make(map[string]*Action, len(tr.Actions))
		for j := range tr.Actions {
			childActions[tr.Actions[j].Name] = &tr.Actions[j]
		}

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

			handMap := make(map[string]bool, len(baseAction.Hands))
			handEntries := make(map[string]HandEntry, len(baseAction.Hands))
			for _, he := range baseAction.Hands {
				handMap[he.Hand] = true
				handEntries[he.Hand] = he
			}
			for _, h := range child.RemoveHands {
				delete(handMap, h)
				delete(handEntries, h)
			}
			for _, h := range child.AddHands {
				handMap[h] = true
				handEntries[h] = HandEntry{Hand: h, Freq: 0}
			}

			for _, he := range baseAction.Hands {
				if handMap[he.Hand] {
					merged.Hands = append(merged.Hands, handEntries[he.Hand])
				}
			}
			for _, h := range child.AddHands {
				if handMap[h] {
					merged.Hands = append(merged.Hands, handEntries[h])
					delete(handMap, h)
				}
			}

			resolved = append(resolved, merged)
			delete(childActions, baseAction.Name)
		}

		for _, child := range tr.Actions {
			if _, isNew := childActions[child.Name]; isNew {
				hands := child.Hands
				if hands == nil {
					hands = stringsToHandEntries(child.AddHands)
				}
				resolved = append(resolved, Action{
					Name:  child.Name,
					Title: child.Title,
					Color: child.Color,
					Hands: hands,
				})
			}
		}

		// Deduplicate: only non-mixed hands (freq == 0) get last-action-wins.
		// Mixed hands (freq > 0) can appear in multiple actions.
		handOwner := make(map[string]string)
		for _, a := range resolved {
			for _, he := range a.Hands {
				if he.Freq == 0 {
					handOwner[he.Hand] = a.Name
				}
			}
		}
		for j, a := range resolved {
			var filtered []HandEntry
			for _, he := range a.Hands {
				if he.Freq > 0 || handOwner[he.Hand] == a.Name {
					filtered = append(filtered, he)
				}
			}
			resolved[j].Hands = filtered
		}

		tabs[i].Actions = resolved
	}
}

func stringsToHandEntries(names []string) []HandEntry {
	entries := make([]HandEntry, len(names))
	for i, n := range names {
		entries[i] = HandEntry{Hand: n}
	}
	return entries
}

// ActionsToHandDetails converts actions to a hand->details map.
// All hands get entries (normal hands at 100%, mixed at their freq).
// Details are sorted by freq descending so [0] is the dominant action.
func ActionsToHandDetails(actions []Action) map[string][]ActionDetail {
	handDetails := make(map[string][]ActionDetail)

	for _, action := range actions {
		for _, he := range action.Hands {
			freq := 100
			if he.Freq > 0 {
				freq = he.Freq
			}
			handDetails[he.Hand] = append(handDetails[he.Hand], ActionDetail{
				Title:     action.Title,
				Color:     action.Color,
				Freq:      freq,
				RaiseSize: action.RaiseSize,
			})
		}
	}

	for hand, details := range handDetails {
		if len(details) > 1 {
			sort.Slice(details, func(i, j int) bool {
				return details[i].Freq > details[j].Freq
			})
			handDetails[hand] = details
		}
	}

	return handDetails
}

// buildLegend returns actions that have at least one hand, for legend display.
func buildLegend(actions []Action) []Action {
	var legend []Action
	for _, a := range actions {
		if len(a.Hands) > 0 {
			legend = append(legend, a)
		}
	}
	return legend
}

// TabDisplayData holds precomputed display data for a single tab (exported for opposite loading)
type TabDisplayData = tabDisplayData

// LoadOppositeData loads the opposite range data given a base file path and an OppositeRef.
func LoadOppositeData(basePath string, ref OppositeRef) *tabDisplayData {
	return LoadOppositeDataCached(basePath, ref, nil)
}

// LoadOppositeDataCached is like LoadOppositeData but uses a cache to avoid re-reading files.
func LoadOppositeDataCached(basePath string, ref OppositeRef, fileCache map[string]*RangeFile) *tabDisplayData {
	dir := filepath.Dir(basePath)
	oppPath := filepath.Clean(filepath.Join(dir, ref.File))

	var rf *RangeFile
	if fileCache != nil {
		if cached, ok := fileCache[oppPath]; ok {
			rf = cached
		}
	}
	if rf == nil {
		var err error
		rf, err = LoadRangeFile(oppPath)
		if err != nil {
			return nil
		}
		if fileCache != nil {
			fileCache[oppPath] = rf
		}
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
		handDetails: ActionsToHandDetails(actions),
		legend:      buildLegend(actions),
		details:     details,
	}
}

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
