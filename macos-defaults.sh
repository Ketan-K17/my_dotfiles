#!/usr/bin/env bash
# Keyboard-friendly macOS defaults. Safe to re-run.
# Undo any line with: defaults delete -g <key>

# Delay before a held key starts repeating (lower = shorter; macOS default is 25)
defaults write -g InitialKeyRepeat -int 15

# Speed of repeat once started (lower = faster)
defaults write -g KeyRepeat -int 2

# Full keyboard access: Tab reaches every control in dialogs, not just text boxes
defaults write -g AppleKeyboardUIMode -int 3

# Some settings only apply after logging out and back in.
