package ranges

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

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

	// 30BB checks
	tab30 := rf.Tabs[1]
	for _, a := range tab30.Actions {
		switch a.Name {
		case "raise":
			if contains(a.Hands, "J2s") {
				t.Error("30BB raise should not contain J2s")
			}
		case "raise_high_freq":
			if !contains(a.Hands, "T4s") || !contains(a.Hands, "K6o") {
				t.Errorf("30BB raise_high_freq should contain T4s and K6o, got: %v", a.Hands)
			}
		case "raise_mid_freq":
			if !contains(a.Hands, "J3s") || !contains(a.Hands, "64s") {
				t.Errorf("30BB raise_mid_freq should contain J3s and 64s, got: %v", a.Hands)
			}
		}
	}

	// 17BB checks
	tab17 := rf.Tabs[2]
	for _, a := range tab17.Actions {
		switch a.Name {
		case "raise":
			if contains(a.Hands, "J3s") {
				t.Error("17BB raise should not contain J3s")
			}
			if contains(a.Hands, "T4s") {
				t.Error("17BB raise should not contain T4s")
			}
		case "raise_high_freq":
			if contains(a.Hands, "T4s") {
				t.Errorf("17BB raise_high_freq should NOT contain T4s, got: %v", a.Hands)
			}
			if contains(a.Hands, "K6o") {
				t.Errorf("17BB raise_high_freq should NOT contain K6o, got: %v", a.Hands)
			}
			if !contains(a.Hands, "K8o") {
				t.Errorf("17BB raise_high_freq should contain K8o, got: %v", a.Hands)
			}
		case "raise_mid_freq":
			if contains(a.Hands, "J3s") {
				t.Errorf("17BB raise_mid_freq should NOT contain J3s, got: %v", a.Hands)
			}
			if contains(a.Hands, "64s") {
				t.Errorf("17BB raise_mid_freq should NOT contain 64s, got: %v", a.Hands)
			}
		case "raise_low_freq":
			if !contains(a.Hands, "65s") || !contains(a.Hands, "J5s") {
				t.Errorf("17BB raise_low_freq should contain 65s and J5s, got: %v", a.Hands)
			}
		}
	}

	fmt.Println("\n=== 17BB resolved actions ===")
	for _, a := range tab17.Actions {
		fmt.Printf("  %s: %v\n", a.Name, a.Hands)
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
			fmt.Printf("  %s: %v\n", a.Name, a.Hands)
		}
	}

	// Check 30BB
	tab30 := rf.Tabs[1]
	found := false
	for _, a := range tab30.Actions {
		if a.Name == "raise_mid_freq" {
			found = true
			if !contains(a.Hands, "J3s") {
				t.Error("30BB raise_mid_freq should contain J3s")
			}
		}
	}
	if !found {
		t.Error("30BB should have raise_mid_freq action")
	}

	// Check 17BB - J3s and T4s should be REMOVED
	tab17 := rf.Tabs[2]
	for _, a := range tab17.Actions {
		if a.Name == "raise_mid_freq" && contains(a.Hands, "J3s") {
			t.Errorf("17BB raise_mid_freq should NOT contain J3s, got: %v", a.Hands)
		}
		if a.Name == "raise_high_freq" && contains(a.Hands, "T4s") {
			t.Errorf("17BB raise_high_freq should NOT contain T4s, got: %v", a.Hands)
		}
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
