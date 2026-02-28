# Land: Session End Cleanup

Cleanly end a work session with Beads status update and context for the next session.

## Usage

```
/land
```

## What This Does

1. **Review active work** - Show tasks currently marked in_progress
2. **Update status** - Mark completed tasks as done
3. **File discoveries** - Capture any new tasks discovered during session
4. **Prepare handoff** - Show what's ready for next session
5. **Sync and push** - Ensure all changes are committed and pushed

## Instructions

Follow these steps to land the plane:

### 1. Check Beads Availability

```bash
export PATH="$PATH:/Users/conner/.local/bin"
if [ -d ".beads" ]; then
    echo "Beads available"
else
    echo "Beads not initialized - skipping task tracking"
fi
```

If Beads not available, skip to git operations.

### 2. Review Active Tasks

```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd list --status in_progress
```

For each active task, determine:
- Was it completed? → Mark done
- Still in progress? → Leave as-is, add notes if helpful
- Blocked? → Update status, file blocker as new task

### 3. Complete Finished Work

For any tasks that were completed this session:
```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd close <task-id>
```

### 4. File Discoveries

If you discovered new work during this session that wasn't captured:
```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd create "Discovered: <description>"
```

Set dependencies if appropriate:
```bash
bd dep add <new-task-id> <related-task-id>
```

### 5. Show Session Summary

Output a summary of what was accomplished:

```
📋 Session Summary

Completed this session:
  ✓ [task-id]: [Title]
  ✓ [task-id]: [Title]

Still in progress:
  ⏳ [task-id]: [Title]
     Note: [where you left off]

Discovered work filed:
  • [task-id]: [Title]

Ready for next session:
  → [task-id]: [Title]
  → [task-id]: [Title]
```

### 6. Sync and Push (CRITICAL)

This step is MANDATORY. Work is not complete until pushed.

```bash
# Sync Beads
export PATH="$PATH:/Users/conner/.local/bin"
bd sync

# Git operations
git status
git add -A
git commit -m "Session work: [brief description]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
git pull --rebase
git push
```

If push fails, resolve and retry until successful.

### 7. Final Verification

```bash
git status
```

Must show "Your branch is up to date with 'origin/...'" or similar.

## Notes

- Always run `/land` before ending a session
- The git push is non-negotiable - stranded local work is lost work
- If you can't push (no remote, permissions), clearly communicate this to the user
- Summary helps the next session (human or agent) know where to start
