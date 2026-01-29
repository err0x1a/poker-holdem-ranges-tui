# Poker Holdem Ranges TUI

A terminal-based viewer for your custom Texas Hold'em poker ranges. Create your own range files and access them instantly with fast search and visualization in a 13x13 hand matrix.

![Screenshot](screenshot.png)

[View demo](demo.gif)

## Features

- Visual 13x13 hand matrix with action-based coloring
- Range list with search/filter
- Details panel with strategic notes
- Color-coded action legend
- YAML-based configuration
- Responsive terminal interface

## Download

Go to [Releases](../../releases) and download the file for your system:

| System | File |
|--------|------|
| Linux | `phr-tui-linux` |
| macOS (M1/M2/M3) | `phr-tui-mac-m1` |
| macOS (Intel) | `phr-tui-mac-intel` |
| Windows | `phr-tui-windows.exe` |

### Linux / macOS

Open the Terminal and run:

```bash
# Make it executable
chmod +x phr-tui-*

# Run
./phr-tui-linux examples/
```

### Windows (PowerShell)

Open PowerShell and run:

```powershell
.\phr-tui-windows.exe examples\
```

### Windows (WSL)

Open WSL and run (uses Linux version):

```bash
chmod +x phr-tui-linux
./phr-tui-linux examples/
```

## Usage

Download the [examples](examples/) folder to test:

```bash
# Run with a directory (loads all .yaml files)
./phr-tui examples/

# Run with a specific file
./phr-tui examples/01_ep_first_in.yaml

# Run with custom title
./phr-tui --title "My Ranges" --title-color "#FF0000" examples/
```

### Controls

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate list |
| `/` | Search/filter |
| `q` or `Ctrl+C` | Quit |

## Range Configuration (YAML)

Each range is defined in a YAML file:

```yaml
title: "BTN vs CO 3-bet"
description: "BTN 3-bet range against CO open"
details: |
  Strategic notes:
  - 3-bet value with premium hands
  - 3-bet bluff with suited connectors
actions:
  - name: raise
    title: "3-bet Value"
    color: "#20bf55"
    hands:
      - AA
      - KK
      - QQ
      - AKs
      - AKo
```

### YAML Fields

| Field | Description |
|-------|-------------|
| `title` | Range name displayed in the TUI |
| `description` | Short description shown in the file list |
| `details` | Strategic notes shown in the details panel (use `\|` for multiline) |
| `actions` | List of actions with their hands |

### Action Fields

| Field | Description |
|-------|-------------|
| `name` | Internal identifier (not displayed) |
| `title` | Action name shown in the legend |
| `color` | Hex color for the hands (e.g. `#20bf55`) |
| `hands` | List of hands for this action |

### Hand Notation

- **Pairs**: `AA`, `KK`, `QQ`, ..., `22`
- **Suited**: `AKs`, `AQs`, `T9s` (same suit)
- **Offsuit**: `AKo`, `AQo`, `T9o` (different suits)

### Recommended Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Green | `#20bf55` | Raise, 3-bet value |
| White | `#FFFFFF` | Call, Limp |
| Yellow | `#FFD166` | Bluff, 3-bet bluff |
| Light Red | `#FF8A80` | All-in |


## Suggested Config Organization

Organize your ranges in separate folders by category. Number the files to control the display order in the TUI:

```
ranges/
├── first-in/                    # Opening ranges
│   ├── 01_ep_40bb.yaml
│   ├── 02_mp_40bb.yaml
│   ├── 03_co_40bb.yaml
│   ├── 04_btn_40bb.yaml
│   └── 05_sb_40bb.yaml
└── defense/                     # Defensive ranges
    ├── 01_bb_vs_sb.yaml
    ├── 02_bb_vs_btn.yaml
    └── 03_co_vs_ep_3bet.yaml
```

Open multiple terminal tabs or use tmux to view different categories side by side:

```bash
# Terminal 1 / tmux pane 1
./phr-tui ranges/first-in/

# Terminal 2 / tmux pane 2
./phr-tui ranges/defense/
```

## Technologies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Cobra](https://github.com/spf13/cobra) - CLI
- [VHS](https://github.com/charmbracelet/vhs) - Demo recording

## Contributing

Feature requests are welcome! Open an [issue](../../issues) to suggest new features or improvements.

If you find this tool useful, consider supporting the project:

[![Sponsor](https://img.shields.io/badge/Sponsor-GitHub-pink?logo=github)](https://github.com/sponsors/err0x1a)

## License

MIT
