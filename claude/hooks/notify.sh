#!/bin/zsh
# Usage: notify.sh <permission|done>  (hook JSON arrives on stdin)
input=$(cat)
project=$(basename "$(echo "$input" | /opt/homebrew/bin/jq -r '.cwd // empty')")

case "$1" in
  permission)
    msg=$(echo "$input" | /opt/homebrew/bin/jq -r '.message // "Claude needs your attention"')
    title="Claude Code — needs you"
    sound="Glass"
    ;;
  done)
    msg="Finished responding"
    title="Claude Code — done"
    sound="Hero"
    ;;
esac

afplay "/System/Library/Sounds/$sound.aiff" &

# macOS notification via terminal-notifier — shows even while Ghostty is focused.
# Clicking it raises Ghostty and, inside tmux, jumps to the pane Claude runs in.
# Project goes in the subtitle: terminal-notifier rejects messages starting with '['.
args=(-title "$title" -message "$msg" -activate com.mitchellh.ghostty)
[[ -n "$project" ]] && args+=(-subtitle "$project")
if [[ -n "$TMUX" && -n "$TMUX_PANE" ]]; then
  sock=${TMUX%%,*}
  t=/opt/homebrew/bin/tmux
  args+=(-group "claude-$TMUX_PANE"
         -execute "$t -S '$sock' switch-client -t '$TMUX_PANE'; $t -S '$sock' select-window -t '$TMUX_PANE'; $t -S '$sock' select-pane -t '$TMUX_PANE'")
fi
/opt/homebrew/bin/terminal-notifier "${args[@]}" >/dev/null 2>&1 &
exit 0
