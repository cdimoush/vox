# Epic to Beads Bridge

Parse an epic README and create a matching Beads task graph with proper dependencies.

## Usage

```
/beads-from-epic <epic-directory>
```

## Example

```
/beads-from-epic specs/epic-beads-integration
```

## Instructions

You are converting an epic's spec structure into Beads tasks. Follow these steps carefully:

### 1. Validate Prerequisites

First, check that Beads is available:
```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd --version
```

If `bd` command fails, inform the user they need to install Beads first.

Check the epic directory exists and has a README.md:
```bash
ls $ARGUMENTS/README.md
```

### 2. Parse Epic README

Read the epic README at `$ARGUMENTS/README.md`.

Extract the following information:
- Epic title (from `# Epic:` heading)
- Phases (from `### Phase N:` sections under `## Execution Order`)
- For each phase:
  - Phase name and description
  - Specs listed (from numbered lists with `[Type: Title](./file.md)` links)
  - Execution order (sequential within phase)

### 3. Create Beads Tasks

For each spec found, create a Beads task:

```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd create "Phase N: Spec Title" --description "Spec file: $ARGUMENTS/spec-file.md"
```

Track the task IDs returned for dependency setup.

### 4. Set Up Dependencies

**Within-phase dependencies** (sequential order):
- Task 2 in a phase depends on Task 1
- Task 3 depends on Task 2
- etc.

**Cross-phase dependencies**:
- First task of Phase N depends on last task of Phase N-1
- This creates "phase gates" where one phase must complete before the next starts

Use:
```bash
bd dep add <dependent-task-id> <blocking-task-id>
```

### 5. Output Summary

After creating all tasks, output a summary:

```
📊 Epic to Beads Bridge Complete

Epic: [Epic Title]
Source: $ARGUMENTS/README.md

Created Tasks:
  Phase 1: [Phase Name] (X tasks)
    ✓ [task-id]: [Title]
    ✓ [task-id]: [Title] (depends on [prev-id])
    ...

  Phase 2: [Phase Name] (Y tasks)
    ✓ [task-id]: [Title] (depends on [last-phase-1-id])
    ...

Total: N tasks with M dependencies

Next steps:
  - Run `bd ready` to see available tasks
  - Run `bd show <id>` to view task details
  - Start with the first ready task
```

### 6. Verify

Run verification commands:
```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd list
bd ready
bd show <first-task-id>
```

## Notes

- If tasks for this epic already exist, warn the user to avoid duplicates
- Use `bd list` to check for existing tasks with similar names before creating
- If the epic structure is unclear, explain what was found and ask for clarification
- Dependencies should mirror the "Execute in order" structure in the README

## Error Handling

- Epic directory doesn't exist → Clear error message
- README.md missing → Clear error message
- No "Execution Order" section → Explain structure needed
- Beads not installed → Explain how to install
- `.beads/` not initialized → Suggest `bd init`
