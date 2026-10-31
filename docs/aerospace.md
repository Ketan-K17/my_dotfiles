AeroSpace is a tiling window manager: it arranges your windows automatically so they never overlap, and you move, resize and switch between them with the keyboard. Config: `aerospace/aerospace.toml`. All shortcuts use `Alt` as the modifier.

**Workspaces**

A workspace is a named screen of windows. You see one at a time and jump between them with a key. Mine are `T` (Ghostty), `B` (Safari), `D` (Standard Notes), and `1`, `2`, `3` (general use; Calendar, Mail and Finder open on `1`).

- `Alt+T`, `Alt+B`, `Alt+D`, `Alt+1/2/3`: go to that workspace.
- `Alt+Tab`: jump back to the previous workspace.
- `Alt+Shift+<same key>`: send the focused window to that workspace.
- `Alt+Shift+Tab`: move the current workspace to the other monitor.

**Moving between windows in a workspace**

- `Alt+h/j/k/l`: focus the window to the left, below, above, or right.
- `Alt+Shift+h/j/k/l`: swap the focused window with its neighbour.

**Layout and size**

- `Alt+/`: tiles (windows side by side). `Alt+,`: accordion (windows stacked).
- `Alt+Shift+minus` / `Alt+Shift+equal`: shrink / grow the focused window.
- `Alt+Shift+F`: fullscreen the focused window (press again to restore).
- `Alt+Shift+Space`: toggle a window between tiled and floating. Use this if a window won't behave.

**Service mode (rarer commands)**

Press `Alt+Shift+;` to enter service mode, then one key:

- `Esc`: reload the config after editing it.
- `r`: reset the layout of the current workspace (fixes a messy one).
- `f`: toggle floating for the focused window.
- `Backspace`: close every window except the focused one.

**Good to know**

- Apps are pinned to workspaces by rules in the config, but only new windows follow them. An already-open window needs `Alt+Shift+<key>` once.
- System Settings, Raycast and Calculator always float instead of tiling.
- To pin another app, add an `[[on-window-detected]]` block with its bundle ID. Get the ID with `osascript -e 'id of app "App Name"'`.
- Edit `~/dotfiles/aerospace/aerospace.toml` (not the copy in `~/.config`), then reload with service mode `Esc`.
- If shortcuts do nothing, check that AeroSpace is enabled under System Settings → Privacy & Security → Accessibility.
- Raycast's own window management should stay disabled so the two don't conflict.
