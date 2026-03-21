# Changelog

All notable changes to this project will be documented in this file.

## [2.2.0] - 2026-03-21

### ✨ Features

- 🔗 Sideranges — Link related ranges (e.g. 3-bet responses) from a tab. A new `sideranges` YAML field adds a titled list of references in the details panel. Press `s` to navigate the list, `Enter` to load inline (tabs and panel stay), `Esc` to restore. Click support included
- 🖱️ Click on siderange items to load them directly

### 🔧 Changes

- `Esc` no longer quits the program — it now exits siderange view. Use `q` or `Ctrl+C` to quit
- Removed dead `abs` helper function
- Test debug output converted from `fmt.Println` to `t.Log`
