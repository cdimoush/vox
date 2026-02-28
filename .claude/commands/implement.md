# Implement the following plan

Follow the `Instructions` to implement the `Plan` then `Report` the completed work.

## Instructions

### Step 0: Check Beads (if available)

If `.beads/` directory exists, check for a matching Beads task:

```bash
export PATH="$PATH:/Users/conner/.local/bin"
if [ -d ".beads" ]; then
    bd ready
fi
```

If a task matches what you're implementing:
```bash
bd update <task-id> --status in_progress
```

If no matching task exists, proceed without Beads tracking.

### Step 1: Implement

- Read the plan carefully.
- THINK HARD about each step before implementing.
- Follow the step-by-step tasks in order.
- Run validation commands at the end.

### Step 2: Update Beads (if used)

If you marked a task as in_progress in Step 0:
```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd close <task-id>
bd ready  # Show what's next
```

## Plan

$ARGUMENTS

## Report

When complete, summarize your work:
- List files created or modified
- Brief description of changes made
- Results of validation commands
- Beads task updated (if applicable)
