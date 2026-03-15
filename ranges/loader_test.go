package ranges

import (
	"fmt"
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

	fmt.Println("\n=== 17BB resolved actions ===")
	for _, a := range tab17.Actions {
		fmt.Printf("  %s: %v\n", a.Name, handNames(a.Hands))
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
		fmt.Printf("\n%s:\n", tr.Tab)
		for _, a := range tr.Actions {
			fmt.Printf("  %s: %v\n", a.Name, handNames(a.Hands))
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
