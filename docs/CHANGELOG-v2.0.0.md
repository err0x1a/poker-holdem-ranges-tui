# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0] - 2026-03-13

### ✨ Features

- 📑 Multi-tab support with base inheritance (`tab_ranges`, `base`, `add_hands`, `remove_hands`)
- 👁️ Opposite range toggle — press `o` to view the opponent's range (`opposite.file`, `opposite.tab`)
- 🔀 Mixed hands with frequency — same hand in multiple actions with `freq` field
- 🎯 Grid cursor navigation with `h/j/k/l`, arrows, and mouse click
- 📊 Hand details panel — action breakdown on cursor hover for each hand
- 🖱️ Mouse support for list and grid selection
- 💰 `raise_size` optional field on actions (e.g. `"2.5x"`)
- 🌗 Freq-based color dimming — mixed hands appear dimmed proportional to frequency
- 🔡 Underline indicator for mixed hands in the grid

### 📝 Docs

- Updated README with detailed docs for all new features
- New example files demonstrating tabs, inheritance, mixed hands, and opposite ranges
- Updated demo.gif showcasing new UI capabilities
- Expanded color palette recommendations

### 🎲 New Examples

- BTN First In (Multi-Stack) — tabs, inheritance, mixed hands, raise_size
- BB vs BTN Raise — opposite range reference
