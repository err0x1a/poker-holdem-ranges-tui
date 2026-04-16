package ranges

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestHandEntryUnmarshal(t *testing.T) {
	data := `
- AA
- hand: AQo
  freq: 50
- "77"
`
	var entries []HandEntry
	if err := yaml.Unmarshal([]byte(data), &entries); err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].Hand != "AA" || entries[0].Freq != 0 {
		t.Errorf("expected {AA, 0}, got {%s, %d}", entries[0].Hand, entries[0].Freq)
	}
	if entries[1].Hand != "AQo" || entries[1].Freq != 50 {
		t.Errorf("expected {AQo, 50}, got {%s, %d}", entries[1].Hand, entries[1].Freq)
	}
	if entries[2].Hand != "77" || entries[2].Freq != 0 {
		t.Errorf("expected {77, 0}, got {%s, %d}", entries[2].Hand, entries[2].Freq)
	}
}

func TestActionsToHandDetailsMixed(t *testing.T) {
	actions := []Action{
		{
			Name:  "allin",
			Title: "All-In",
			Color: "#FF8A80",
			Hands: []HandEntry{
				{Hand: "AA", Freq: 0},
				{Hand: "AQo", Freq: 50},
			},
		},
		{
			Name:  "call",
			Title: "Call",
			Color: "#FFFFFF",
			Hands: []HandEntry{
				{Hand: "KK", Freq: 0},
				{Hand: "AQo", Freq: 50},
			},
		},
	}

	hd := ActionsToHandDetails(actions)

	// Dominant color (details[0]) for normal hands
	if hd["AA"][0].Color != "#FF8A80" {
		t.Errorf("AA dominant color should be #FF8A80, got %s", hd["AA"][0].Color)
	}
	if hd["KK"][0].Color != "#FFFFFF" {
		t.Errorf("KK dominant color should be #FFFFFF, got %s", hd["KK"][0].Color)
	}

	// Mixed hand: dominant is highest freq (both 50%, first wins = All-In)
	if hd["AQo"][0].Color != "#FF8A80" {
		t.Errorf("AQo dominant color should be #FF8A80, got %s", hd["AQo"][0].Color)
	}

	// All hands have details
	if len(hd["AA"]) != 1 || hd["AA"][0].Freq != 100 {
		t.Errorf("AA should have 1 detail at 100%%, got %v", hd["AA"])
	}
	if len(hd["AQo"]) != 2 {
		t.Fatalf("AQo should have 2 details, got %d", len(hd["AQo"]))
	}
	if hd["AQo"][0].Freq != 50 {
		t.Errorf("first detail should be 50%%, got %d%%", hd["AQo"][0].Freq)
	}
	if hd["AQo"][1].Freq != 50 {
		t.Errorf("second detail should be 50%%, got %d%%", hd["AQo"][1].Freq)
	}
}

func TestMixedDetails(t *testing.T) {
	actions := []Action{
		{
			Name:  "raise",
			Title: "Raise",
			Color: "#20bf55",
			Hands: []HandEntry{
				{Hand: "AQo", Freq: 30},
			},
		},
		{
			Name:  "allin",
			Title: "All-In",
			Color: "#FF8A80",
			Hands: []HandEntry{
				{Hand: "AQo", Freq: 50},
			},
		},
		{
			Name:  "call",
			Title: "Call",
			Color: "#FFFFFF",
			Hands: []HandEntry{
				{Hand: "AQo", Freq: 20},
			},
		},
	}

	hd := ActionsToHandDetails(actions)

	details := hd["AQo"]
	if len(details) != 3 {
		t.Fatalf("expected 3 details for AQo, got %d", len(details))
	}

	// Should be sorted by freq desc
	if details[0].Freq != 50 {
		t.Errorf("first should be 50%%, got %d%%", details[0].Freq)
	}
	if details[1].Freq != 30 {
		t.Errorf("second should be 30%%, got %d%%", details[1].Freq)
	}
	if details[2].Freq != 20 {
		t.Errorf("third should be 20%%, got %d%%", details[2].Freq)
	}
}

func TestBuildLegend(t *testing.T) {
	actions := []Action{
		{Name: "raise", Title: "Raise", Color: "#20bf55", Hands: []HandEntry{{Hand: "AA"}}},
		{Name: "call", Title: "Call", Color: "#FFFFFF", Hands: []HandEntry{{Hand: "KK"}}},
		{Name: "empty", Title: "Empty", Color: "#000000", Hands: nil},
	}

	legend := buildLegend(actions)
	if len(legend) != 2 {
		t.Errorf("expected 2 legend entries (empty filtered out), got %d", len(legend))
	}
}

func TestResolveTabsWithHandEntry(t *testing.T) {
	data := `
title: "Test"
tab_ranges:
  - tab: "base"
    actions:
      - name: raise
        title: "Raise"
        color: "#20bf55"
        hands:
          - AA
          - KK
          - hand: AQo
            freq: 50
      - name: allin
        title: "All-In"
        color: "#FF8A80"
        hands:
          - hand: AQo
            freq: 50

  - tab: "child"
    base: "base"
    actions:
      - name: raise
        remove_hands: [KK]
        add_hands: [QQ]
`

	var rf RangeFile
	if err := yaml.Unmarshal([]byte(data), &rf); err != nil {
		t.Fatal(err)
	}

	resolveTabs(rf.Tabs)

	child := rf.Tabs[1]
	for _, a := range child.Actions {
		if a.Name == "raise" {
			if containsHand(a.Hands, "KK") {
				t.Error("child raise should not contain KK")
			}
			if !containsHand(a.Hands, "AA") {
				t.Error("child raise should contain AA")
			}
			if !containsHand(a.Hands, "QQ") {
				t.Error("child raise should contain QQ")
			}
			for _, he := range a.Hands {
				if he.Hand == "AQo" && he.Freq != 50 {
					t.Errorf("AQo should have freq 50, got %d", he.Freq)
				}
			}
		}
		if a.Name == "allin" {
			if !containsHand(a.Hands, "AQo") {
				t.Error("child allin should contain AQo")
			}
		}
	}
}

func TestResolveBTN(t *testing.T) {
	data := `
title: "BTN First In"
tab_ranges:
  - tab: "60BB"
    actions:
      - name: raise
        title: "Raise"
        color: "#20bf55"
        hands: [AA, KK, QQ, JJ, TT, 99, 88, 77, 66, 55, AKs, AQs, AJs, ATs, A9s, A8s, A7s, A6s, A5s, A4s, A3s, KQs, KJs, KTs, K9s, K8s, QJs, QTs, Q9s, JTs, J9s, T9s, T8s, 98s, AKo, AQo, AJo, KQo, ATo, KJo, QJo, A2s, K7s, K6s, Q8s, J8s, "44", "33", "22", K5s, K4s, 76s, T7s, 97s, 87s, A9o, QTo, KTo, JTo, K3s, K2s, Q7s, Q6s, Q5s, Q4s, J7s, J6s, T6s, 96s, 86s, 75s, 65s, 54s, A8o, A7o, K9o, T9o, Q9o, J9o, A5o, K8o, T8o, 98o, 85s, J5s, Q3s, Q2s, J4s, J3s, T5s, T4s, T3s, 95s, 74s, 64s, 53s, A6o, A4o, A3o, A2o, K7o, K6o, Q8o, J8o, 87o]
      - name: raise_high_freq
        title: "Alta frequência"
        color: "#f7971e"
        hands: [K5o, Q7o, 97o]
      - name: raise_mid_freq
        title: "Média frequência"
        color: "#fdd835"
        hands: [T7o]
      - name: raise_low_freq
        title: "Baixa frequência"
        color: "#fff9c4"
        hands: [J7o, 76o, 43s, 84s]

  - tab: "30BB"
    base: "60BB"
    actions:
      - name: raise
        remove_hands: [J2s, T3s, 74s, "22", 53s, 87o]
      - name: raise_high_freq
        remove_hands: [K5o, Q7o, 97o]
        add_hands: [T4s, K6o]
      - name: raise_mid_freq
        remove_hands: [T7o]
        add_hands: [J3s, 64s]
      - name: raise_low_freq
        remove_hands: [J7o, 76o, 43s, 84s]

  - tab: "17BB"
    base: "30BB"
    actions:
      - name: raise
        remove_hands: [K2s, Q2s, Q3s, J4s, T5s, 95s, 85s, 75s, 54s, K7o, Q8o, J8o, T8o, 98o, 65s, J5s, Q4s, K3s, K8o]
      - name: raise_high_freq
        remove_hands: [T4s, K6o]
        add_hands: [K8o]
      - name: raise_mid_freq
        remove_hands: [J3s, 64s]
      - name: raise_low_freq
        add_hands: [65s, J5s, Q4s, K3s]
`

	var rf RangeFile
	if err := yaml.Unmarshal([]byte(data), &rf); err != nil {
		t.Fatal(err)
	}

	resolveTabs(rf.Tabs)

	tab30 := rf.Tabs[1]
	for _, a := range tab30.Actions {
		switch a.Name {
		case "raise":
			if containsHand(a.Hands, "J2s") {
				t.Error("30BB raise should not contain J2s")
			}
		case "raise_high_freq":
			if !containsHand(a.Hands, "T4s") || !containsHand(a.Hands, "K6o") {
				t.Errorf("30BB raise_high_freq should contain T4s and K6o, got: %v", handNames(a.Hands))
			}
		case "raise_mid_freq":
			if !containsHand(a.Hands, "J3s") || !containsHand(a.Hands, "64s") {
				t.Errorf("30BB raise_mid_freq should contain J3s and 64s, got: %v", handNames(a.Hands))
			}
		}
	}

	tab17 := rf.Tabs[2]
	for _, a := range tab17.Actions {
		switch a.Name {
		case "raise":
			if containsHand(a.Hands, "J3s") {
				t.Error("17BB raise should not contain J3s")
			}
			if containsHand(a.Hands, "T4s") {
				t.Error("17BB raise should not contain T4s")
			}
		case "raise_high_freq":
			if containsHand(a.Hands, "T4s") {
				t.Errorf("17BB raise_high_freq should NOT contain T4s, got: %v", handNames(a.Hands))
			}
			if containsHand(a.Hands, "K6o") {
				t.Errorf("17BB raise_high_freq should NOT contain K6o, got: %v", handNames(a.Hands))
			}
			if !containsHand(a.Hands, "K8o") {
				t.Errorf("17BB raise_high_freq should contain K8o, got: %v", handNames(a.Hands))
			}
		case "raise_mid_freq":
			if containsHand(a.Hands, "J3s") {
				t.Errorf("17BB raise_mid_freq should NOT contain J3s, got: %v", handNames(a.Hands))
			}
			if containsHand(a.Hands, "64s") {
				t.Errorf("17BB raise_mid_freq should NOT contain 64s, got: %v", handNames(a.Hands))
			}
		case "raise_low_freq":
			if !containsHand(a.Hands, "65s") || !containsHand(a.Hands, "J5s") {
				t.Errorf("17BB raise_low_freq should contain 65s and J5s, got: %v", handNames(a.Hands))
			}
		}
	}

	for _, a := range tab17.Actions {
		t.Logf("  %s: %v", a.Name, handNames(a.Hands))
	}
}

func TestResolveTabsChained(t *testing.T) {
	data := `
title: "Test"
tab_ranges:
  - tab: "60BB+"
    actions:
      - name: raise
        title: "Raise"
        color: "#20bf55"
        hands: [AA, KK, QQ]
      - name: raise_high_freq
        title: "Alta freq"
        color: "#f7971e"
        hands: [JJ, TT]

  - tab: "30BB"
    base: "60BB+"
    actions:
      - name: raise
        add_hands: [JTs]
      - name: raise_high_freq
        add_hands: [T4s, K6o]
      - name: raise_mid_freq
        title: "Media freq"
        color: "#fdd835"
        add_hands: [J3s, 64s]

  - tab: "17BB"
    base: "30BB"
    actions:
      - name: raise_high_freq
        remove_hands: [T4s, K6o]
      - name: raise_mid_freq
        remove_hands: [J3s, 64s]
`

	var rf RangeFile
	if err := yaml.Unmarshal([]byte(data), &rf); err != nil {
		t.Fatal(err)
	}

	resolveTabs(rf.Tabs)

	for _, tr := range rf.Tabs {
		t.Logf("%s:", tr.Tab)
		for _, a := range tr.Actions {
			t.Logf("  %s: %v", a.Name, handNames(a.Hands))
		}
	}

	tab30 := rf.Tabs[1]
	found := false
	for _, a := range tab30.Actions {
		if a.Name == "raise_mid_freq" {
			found = true
			if !containsHand(a.Hands, "J3s") {
				t.Error("30BB raise_mid_freq should contain J3s")
			}
		}
	}
	if !found {
		t.Error("30BB should have raise_mid_freq action")
	}

	tab17 := rf.Tabs[2]
	for _, a := range tab17.Actions {
		if a.Name == "raise_mid_freq" && containsHand(a.Hands, "J3s") {
			t.Errorf("17BB raise_mid_freq should NOT contain J3s, got: %v", handNames(a.Hands))
		}
		if a.Name == "raise_high_freq" && containsHand(a.Hands, "T4s") {
			t.Errorf("17BB raise_high_freq should NOT contain T4s, got: %v", handNames(a.Hands))
		}
	}
}

func TestSiderangesParsing(t *testing.T) {
	data := `
title: "EP First In"
sideranges:
  title: "vs 3-bet"
  items:
    - label: "vs UTG1"
      file: "responses/vs_utg1.yaml"
      tab: "20BB"
    - label: "vs LJ"
      file: "responses/vs_lj.yaml"
actions:
  - name: raise
    title: "Raise"
    color: "#20bf55"
    hands: [AA]
`
	var rf RangeFile
	if err := yaml.Unmarshal([]byte(data), &rf); err != nil {
		t.Fatal(err)
	}

	if rf.Sideranges == nil {
		t.Fatal("expected sideranges to be parsed")
	}
	if rf.Sideranges.Title != "vs 3-bet" {
		t.Errorf("expected title 'vs 3-bet', got %q", rf.Sideranges.Title)
	}
	if len(rf.Sideranges.Items) != 2 {
		t.Fatalf("expected 2 siderange items, got %d", len(rf.Sideranges.Items))
	}
	if rf.Sideranges.Items[0].Label != "vs UTG1" {
		t.Errorf("expected label 'vs UTG1', got %q", rf.Sideranges.Items[0].Label)
	}
	if rf.Sideranges.Items[0].File != "responses/vs_utg1.yaml" {
		t.Errorf("expected file 'responses/vs_utg1.yaml', got %q", rf.Sideranges.Items[0].File)
	}
	if rf.Sideranges.Items[0].Tab != "20BB" {
		t.Errorf("expected tab '20BB', got %q", rf.Sideranges.Items[0].Tab)
	}
	if rf.Sideranges.Items[1].Tab != "" {
		t.Errorf("expected empty tab, got %q", rf.Sideranges.Items[1].Tab)
	}
}

func TestSiderangesPerTab(t *testing.T) {
	data := `
title: "EP First In"
sideranges:
  title: "file-level"
  items:
    - label: "fallback"
      file: "fallback.yaml"
tab_ranges:
  - tab: "40+"
    sideranges:
      title: "vs 3-bet 40+"
      items:
        - label: "vs UTG1 40+"
          file: "responses/vs_utg1_40.yaml"
    actions:
      - name: raise
        title: "Raise"
        color: "#20bf55"
        hands: [AA]
  - tab: "20BB"
    actions:
      - name: raise
        title: "Raise"
        color: "#20bf55"
        hands: [AA]
`
	var rf RangeFile
	if err := yaml.Unmarshal([]byte(data), &rf); err != nil {
		t.Fatal(err)
	}

	// Tab with its own sideranges should use them
	if rf.Tabs[0].Sideranges == nil {
		t.Fatal("tab 40+ should have sideranges")
	}
	if rf.Tabs[0].Sideranges.Title != "vs 3-bet 40+" {
		t.Errorf("expected 'vs 3-bet 40+', got %q", rf.Tabs[0].Sideranges.Title)
	}

	// Tab without sideranges should be nil (fallback handled by NewWithTabs)
	if rf.Tabs[1].Sideranges != nil {
		t.Error("tab 20BB should not have its own sideranges")
	}

	// File-level sideranges should exist
	if rf.Sideranges == nil {
		t.Fatal("file-level sideranges should exist")
	}
	if rf.Sideranges.Title != "file-level" {
		t.Errorf("expected 'file-level', got %q", rf.Sideranges.Title)
	}
}

func TestSiderangesFallbackInModel(t *testing.T) {
	fileSideranges := &Sideranges{
		Title: "file-level",
		Items: []SiderangeItem{{Label: "fallback", File: "f.yaml"}},
	}
	tabSideranges := &Sideranges{
		Title: "tab-level",
		Items: []SiderangeItem{{Label: "tab item", File: "t.yaml"}},
	}

	tabs := []TabRange{
		{
			Tab:        "40+",
			Sideranges: tabSideranges,
			Actions:    []Action{{Name: "r", Title: "R", Color: "#fff", Hands: []HandEntry{{Hand: "AA"}}}},
		},
		{
			Tab:     "20BB",
			Actions: []Action{{Name: "r", Title: "R", Color: "#fff", Hands: []HandEntry{{Hand: "AA"}}}},
		},
	}

	model := NewWithTabs(tabs, fileSideranges, "", "", nil)

	// First tab should use its own sideranges
	if model.sideranges == nil || model.sideranges.Title != "tab-level" {
		t.Errorf("tab 40+ should use tab-level sideranges, got %v", model.sideranges)
	}

	// Switch to second tab - should fallback to file-level
	model.SetTabIndex(1)
	if model.sideranges == nil || model.sideranges.Title != "file-level" {
		t.Errorf("tab 20BB should fallback to file-level sideranges, got %v", model.sideranges)
	}
}

func TestSiderangesOnlyOnMiddleTab(t *testing.T) {
	// Simulates: file with many tabs, sideranges only on "20BB" tab, no file-level sideranges
	tabs := []TabRange{
		{Tab: "100BB", Actions: []Action{{Name: "r", Title: "R", Color: "#fff", Hands: []HandEntry{{Hand: "AA"}}}}},
		{Tab: "80BB", Actions: []Action{{Name: "r", Title: "R", Color: "#fff", Hands: []HandEntry{{Hand: "AA"}}}}},
		{
			Tab: "20BB",
			Sideranges: &Sideranges{
				Title: "vs 3-bet",
				Items: []SiderangeItem{
					{Label: "vs UTG1", File: "vs3bet/01.yaml"},
					{Label: "vs LJ", File: "vs3bet/02.yaml"},
				},
			},
			Actions: []Action{{Name: "r", Title: "R", Color: "#fff", Hands: []HandEntry{{Hand: "AA"}}}},
		},
		{Tab: "17BB", Actions: []Action{{Name: "r", Title: "R", Color: "#fff", Hands: []HandEntry{{Hand: "AA"}}}}},
	}

	model := NewWithTabs(tabs, nil, "", "", nil) // no file-level sideranges

	// Tab 0 (100BB) - no sideranges
	if model.hasSideranges() {
		t.Error("100BB should not have sideranges")
	}

	// Switch to tab 2 (20BB) - should have sideranges
	model.SetTabIndex(2)
	if !model.hasSideranges() {
		t.Fatal("20BB should have sideranges")
	}
	if model.sideranges.Title != "vs 3-bet" {
		t.Errorf("expected title 'vs 3-bet', got %q", model.sideranges.Title)
	}
	if len(model.sideranges.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(model.sideranges.Items))
	}

	// Switch to tab 3 (17BB) - no sideranges
	model.SetTabIndex(3)
	if model.hasSideranges() {
		t.Error("17BB should not have sideranges")
	}

	// Switch back to tab 2 (20BB) - should still have sideranges
	model.SetTabIndex(2)
	if !model.hasSideranges() {
		t.Fatal("20BB should still have sideranges after switching back")
	}

	// Verify buildDetailsPanel includes sideranges
	panel := model.buildDetailsPanel()
	if panel == "" {
		t.Error("panel should not be empty when sideranges exist")
	}
	if !strings.Contains(panel, "vs 3-bet") {
		t.Errorf("panel should contain sideranges title, got: %s", panel)
	}
	if !strings.Contains(panel, "vs UTG1") {
		t.Errorf("panel should contain siderange items, got: %s", panel)
	}
}

func TestParseTabStacks(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect map[string]float64
	}{
		{
			name:  "format with hero and avg",
			input: "HJ 14BB | 10BB Avg | CO 4 | BTN 18 | SB 7.5 | BB 15",
			expect: map[string]float64{"HJ": 14, "CO": 4, "BTN": 18, "SB": 7.5, "BB": 15},
		},
		{
			name:  "format with all positions",
			input: "HJ 6BB | 10BB Avg | UTG 24 | UTG+1 7 | LJ 8 | CO 9 | BTN 5 | SB 9.5 | BB 10",
			expect: map[string]float64{"HJ": 6, "UTG": 24, "UTG+1": 7, "LJ": 8, "CO": 9, "BTN": 5, "SB": 9.5, "BB": 10},
		},
		{
			name:  "villains only format",
			input: "CO 22 | BTN 26 | SB 12.5 | BB 17 || UTG 24 | UTG+1 27 | LJ 16",
			expect: map[string]float64{"CO": 22, "BTN": 26, "SB": 12.5, "BB": 17, "UTG": 24, "UTG+1": 27, "LJ": 16},
		},
		{
			name:   "simple tabs without stacks",
			input:  "20BB",
			expect: map[string]float64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTabStacks(tt.input)
			for pos, expected := range tt.expect {
				if got, ok := result[pos]; !ok {
					t.Errorf("missing position %s", pos)
				} else if got != expected {
					t.Errorf("position %s: got %g, want %g", pos, got, expected)
				}
			}
		})
	}
}

func TestDetectHeroPosition(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"HJ 14BB | 10BB Avg | CO 4", "HJ"},
		{"UTG 32BB | 20BB Avg", "UTG"},
		{"20BB", ""},
		{"CO 17 | BTN 23", "CO"},
	}
	for _, tt := range tests {
		got := detectHeroPosition(tt.input)
		if got != tt.expect {
			t.Errorf("detectHeroPosition(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

func TestPositionsBehind(t *testing.T) {
	behind := PositionsBehind("HJ")
	if behind[0] != "CO" || behind[1] != "BTN" || behind[2] != "SB" || behind[3] != "BB" {
		t.Errorf("HJ behind: got %v, want [CO BTN SB BB ...]", behind)
	}
	behind = PositionsBehind("UTG")
	if behind[0] != "UTG+1" {
		t.Errorf("UTG behind: got %v, want [UTG+1 ...]", behind)
	}
}

func containsHand(entries []HandEntry, name string) bool {
	for _, e := range entries {
		if e.Hand == name {
			return true
		}
	}
	return false
}

func handNames(entries []HandEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Hand
	}
	return names
}
