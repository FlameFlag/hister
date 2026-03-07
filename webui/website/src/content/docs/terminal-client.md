---
date: '2026-02-13T10:59:19+01:00'
draft: false
title: 'Terminal Client'
---

See also [the configuration documentation](configuration).

## Command-Line Usage

View all available commands:

```bash
./hister help
```

### Index a URL Manually

To manually index a specific URL:

```bash
./hister index https://example.com
```

## TUI (Terminal UI)

Hister provides a terminal-based user interface for searching your browsing history without leaving your terminal.

### Start the TUI

Run the search command without any arguments:

```bash
hister search
```

### TUI Features

- **Multi-tab interface**: Search, History, Rules, and Add tabs
- **Mouse support**: Scroll with mouse wheel, click to select, right-click for context menu
- **Theming**: Built-in color themes with interactive picker (press `ctrl+t`)
- **Settings overlay**: Edit keybindings interactively (press `ctrl+s`)
- **Context menu**: Right-click on results for quick actions (open, delete, prioritize)

### Tabs

- **Search** (Alt+1): Main search interface
- **History** (Alt+2): View your recent search history
- **Rules** (Alt+3): Manage blacklist, priority, and alias rules
- **Add** (Alt+4): Manually add URLs to the index

### TUI Keybindings

The TUI uses the following keybindings by default:

| Key           | Action          | Description                                    |
| ------------- | --------------- | ---------------------------------------------- |
| `ctrl+c`      | quit            | Exit the TUI                                   |
| `f1`          | toggle_help     | Show/hide keybindings help overlay             |
| `tab`, `esc`  | toggle_focus    | Switch between search input and results list   |
| `up`, `k`     | scroll_up       | Navigate up in results                         |
| `down`, `j`   | scroll_down     | Navigate down in results                       |
| `enter`       | open_result     | Open the selected result in your browser       |
| `ctrl+d`, `d` | delete_result   | Delete the selected result from the index      |
| `ctrl+t`, `t` | toggle_theme    | Open the interactive theme picker              |
| `ctrl+s`, `s` | toggle_settings | Open the keybinding editor overlay             |
| `ctrl+o`, `o` | toggle_sort     | Toggle domain-based sorting for search results |
| `alt+1`       | tab_search      | Switch to the Search tab                       |
| `alt+2`       | tab_history     | Switch to the History tab                      |
| `alt+3`       | tab_rules       | Switch to the Rules tab                        |
| `alt+4`       | tab_add         | Switch to the Add tab                          |

### Mouse Controls

- **Left-click**: Select results or open tabs
- **Right-click**: Open context menu (open, delete, prioritize)
- **Scroll wheel**: Navigate through results
- **Scrollbar drag**: Quick scroll through long result lists

### Customizing TUI

TUI settings are stored in a separate `tui.yaml` file alongside your main config file. This file is automatically created with default values when you first run `hister search`.

**TUI config location**: `~/.config/hister/tui.yaml`

#### tui.yaml Structure

```yaml
# Theme settings
dark_theme: 'dracula'
light_theme: 'gruvbox'
color_scheme: 'auto'
# themes_dir: "/path/to/custom/themes"  # optional

# TUI keybindings
hotkeys:
  ctrl+c: 'quit'
  ctrl+t: 'toggle_theme'
  ctrl+s: 'toggle_settings'
  ctrl+o: 'toggle_sort'
  alt+1: 'tab_search'
  alt+2: 'tab_history'
  alt+3: 'tab_rules'
  alt+4: 'tab_add'
  # ... and all other TUI keybindings
```

#### Available TUI Actions

- `quit` - Exit the TUI application
- `toggle_help` - Show/hide the help overlay
- `toggle_focus` - Switch between input and results views
- `scroll_up`/`scroll_down` - Move selection up/down
- `open_result` - Open selected URL in browser
- `delete_result` - Delete selected entry from index
- `toggle_theme` - Open theme picker
- `toggle_settings` - Open keybinding editor
- `toggle_sort` - Toggle sorting mode
- `tab_search`/`tab_history`/`tab_rules`/`tab_add` - Switch tabs

Note: After modifying `tui.yaml`, restart the `hister search` command to apply changes.
