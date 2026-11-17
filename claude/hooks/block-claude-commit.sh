#!/bin/zsh
# PreToolUse hook for Bash: blocks `git commit` commands that mention "claude"
# (case-insensitive) anywhere in the command, which includes the message.
# Exit 2 sends the reason on stderr back to Claude and cancels the command.
input=$(cat)
cmd=$(printf '%s' "$input" | /opt/homebrew/bin/jq -r '.tool_input.command // empty')

printf '%s' "$cmd" | grep -Eq '(^|[;&|([:space:]])git[[:space:]]+commit([[:space:]]|$)' || exit 0

if printf '%s' "$cmd" | grep -iq 'claude'; then
  echo "Blocked: git commit messages must not mention Claude. Remove the word from the commit message (including any Co-Authored-By line) and retry." >&2
  exit 2
fi
exit 0
