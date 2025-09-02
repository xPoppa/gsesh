#! /usr/bin/fish
# Rewrite this script to go so you can keep sessions better!

set SESSION_FILE "/home/poppa/.local/share/ghostty-sessionizer"
if test (count $argv) -eq 1
    set selected $argv[1]
else
    set selected (find -L ~/boot.dev ~/go ~/haskell ~/frontendmasters ~/Projects ~/natalie ~/ -mindepth 1 -maxdepth 1 -type d | fzf)
end

if test -z "$selected"
    exit 0
end

set selected_name (basename "$selected" | tr . _)

set has_session (grep "$selected" "$SESSION_FILE" )
echo $has_session
if test -n "$has_session";
    echo "Re attach exit with exit 0"
    set pid (echo $has_session | cut -d ' ' -f 1)
    echo "HAS_SESSION PID: $pid"
    echo "WMCTL pid: $(wmctrl -lp | grep $pid | cut -d ' ' -f 1)"
    set window_pid (wmctrl -lp | grep $pid | cut -d ' ' -f 1)
    wmctrl -ia $window_pid
    exit 0
end

if test -z "$has_session"; 
    ghostty --working-directory="$selected" &
    set pid (ps -eo pid,command | grep "ghostty --working-directory=$selected" | grep -v "grep" | cut -d ' ' -f 3)
    echo "PID ECHO: $pid"
    printf "PID: %d\n" $pid
    printf "SELECTED: %s\n" $selected
    printf "%d | %s\n" $pid $selected >> $SESSION_FILE
    cat $SESSION_FILE
    exit 0
end
