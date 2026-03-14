# Changelog

All notable changes to this project will be documented in this file.

## [2.1.0] - 2026-03-14

### ✨ Features

- 🎨 Split border color progression — cell borders show a left-to-right color gradient proportional to each action's frequency. Remaining percentage (fold) is shown in gray
- 🔘 Legend click filtering — click on a legend item to hide/show that action across all ranges. Hidden actions are removed from the grid and border progression. Filter persists across range and tab switches
- 🖱️ Background highlight cursor — selected cell uses a subtle background highlight instead of thick border, preserving border color information

### 🔧 Changes

- Opposite range toggle keybinding changed from `o` to `Ctrl+O`
- Removed frequency-based color dimming (replaced by border progression)
- Removed underline indicator for mixed hands (replaced by border progression)
