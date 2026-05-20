#!/bin/bash

# Speckit Implementation Monitor
# Run this in a separate terminal pane to watch Claude's progress live

cd /Users/mfinlayson@bgrove.com/src/activtrak-admin-cli || exit 1

while true; do
    clear
    echo -e "\033[1;36m===================================================\033[0m"
    echo -e "\033[1;36m           SPECKIT IMPLEMENTATION DASHBOARD        \033[0m"
    echo -e "\033[1;36m===================================================\033[0m"
    echo -e "Time: $(date "+%H:%M:%S") | Branch: $(git branch --show-current)"
    echo ""

    # Calculate Progress
    # We look for commits that have "(005):" in them, excluding the docs commits we made manually
    TOTAL_TASKS=30
    COMPLETED=$(git log --oneline origin/main..HEAD | grep -i "(005):" | grep -vi "docs(005)" | wc -l | awk '{print $1}')
    
    # Simple progress bar
    PERCENT=$((COMPLETED * 100 / TOTAL_TASKS))
    BARS=$((PERCENT / 5))
    SPACES=$((20 - BARS))
    printf "Progress: [%-${BARS}s%-${SPACES}s] %d%% (%d/%d Tasks)\n" "$(printf '#%.0s' $(seq 1 $BARS))" "" "$PERCENT" "$COMPLETED" "$TOTAL_TASKS"
    echo ""

    echo -e "\033[1;33m[ Most Recent Commits ]\033[0m"
    git log -n 5 --pretty=format:"%C(yellow)%h%Creset %C(green)%ad%Creset %s" --date=short
    echo ""
    echo ""

    echo -e "\033[1;33m[ Uncommitted Work (Current Task) ]\033[0m"
    git status --short
    echo ""
    
    echo -e "\033[1;90mRefreshing every 5 seconds... Press Ctrl+C to exit.\033[0m"
    sleep 5
done
