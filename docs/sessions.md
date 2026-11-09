**Safari sessions** save a group of tabs under a name so you can reopen them later.

There are two ways to use them, and both work with the same files in `~/Desktop/sessions/`:

- **Apps:** `~/dotfiles/apps/save_session.app` and `~/dotfiles/apps/load_session.app`. Open them from Spotlight or the Dock.
- **Terminal:** `save_session` and `load_session`, from `~/dotfiles/bin`.

**Saving**

- `save_session work` saves the tabs in the front window as "work".
- `save_session -a work` saves every window. Each one reopens as its own window.
- With no name, it asks you for one. If the name already exists, it asks before replacing it.

**Loading**

- `load_session` opens an fzf picker with a preview of the URLs. Press Tab to pick several.
- `load_session work` opens it straight away. Tab-completion fills in session names.
  A partial name like `load_session wo` opens the picker already filtered, or opens the session directly if only one matches.
- `load_session -l` lists sessions, newest first. `-e name` opens one in `$EDITOR`, and `-d name` deletes it.

Each session file is plain text with one URL per line. A blank line starts a new window, and lines beginning with `#` are ignored, so you can edit the files by hand.

The first time each app or command runs, macOS asks for permission to control Safari. Allow it.
