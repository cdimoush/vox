# Spec Executor Agent

A beads-aware sub-agent for implementing specs from epic directories. Optimized for parallel execution on independent specs.

## Purpose

Execute individual spec files (features, chores) with automatic beads task lifecycle management. Multiple instances can run concurrently on independent specs.

## Agent Capabilities

This agent has full privileges to:
- Read and analyze code files
- Create and modify files
- Run shell commands (builds, tests, scripts)
- Install dependencies via `uv add`
- Interact with beads CLI (`bd` commands)

## When to Use This Agent

Deploy `spec-executor` when:
- Implementing specs from an epic directory
- Working through beads-tracked tasks
- Multiple independent specs can run in parallel

Do NOT use when:
- Specs have dependencies (use serial execution instead)
- Exploratory work without a clear spec
- Quick single-file edits

## Execution Modes

### Parallel Execution (Preferred for Independent Specs)

When specs have no dependencies, deploy multiple agents in a SINGLE message:

```
Main agent sends ONE message with multiple Task tool calls:

[Task 1] description: "Implement chore-voice-commands"
         subagent_type: "general-purpose"
         prompt: <spec-executor prompt for chore-voice-commands.md>

[Task 2] description: "Implement chore-beads-commands"
         subagent_type: "general-purpose"
         prompt: <spec-executor prompt for chore-beads-commands.md>

[Task 3] description: "Implement chore-planning-commands"
         subagent_type: "general-purpose"
         prompt: <spec-executor prompt for chore-planning-commands.md>
```

All instances work concurrently. Main agent waits for all to complete.

### Serial Execution (For Dependent Specs)

When specs have dependencies, wait for each to complete:

```
1. Deploy spec-executor for Spec A
2. Wait for completion, verify success
3. Deploy spec-executor for Spec B (depends on A)
4. Continue...
```

## Invocation Template

Use this prompt template when invoking via Task tool:

```
SPEC EXECUTOR MODE

Spec: <path-to-spec-file>
Epic: <epic-directory-name>
Beads Path: /Users/conner/.local/bin

---

## Execution Protocol

### Phase 1: Beads Integration

Check if beads is tracking this work:

```bash
export PATH="$PATH:/Users/conner/.local/bin"
if [ -d ".beads" ]; then
    bd list --status open
fi
```

If a task matches this spec (title contains spec name or description references spec file):
- Note the task ID
- Mark it in progress:
```bash
bd update <task-id> --status in_progress
```

If no matching task, proceed without beads tracking.

### Phase 2: Implement Spec

1. Read the spec file completely
2. Read all files mentioned in spec's context sections
3. Follow step-by-step tasks in EXACT order
4. Do not skip steps or reorder
5. Create/modify files as specified

### Phase 3: Validate

Run all validation commands from the spec's acceptance criteria.
Note pass/fail status for each.

### Phase 4: Beads Closure

If a beads task was marked in_progress:
```bash
export PATH="$PATH:/Users/conner/.local/bin"
bd close <task-id>
```

### Phase 5: Return Report

Return a structured report in this exact format:

---
## Spec Execution Report

**Spec**: <spec-filename>
**Status**: SUCCESS | PARTIAL | FAILED

### Beads
- Task ID: <id or "not tracked">
- Status: closed | in_progress | n/a

### Files Created
- `path/to/file1` - description
- `path/to/file2` - description

### Files Modified
- `path/to/file3` - what changed

### Validation Results
| Check | Result |
|-------|--------|
| <validation-1> | PASS/FAIL |
| <validation-2> | PASS/FAIL |

### Blockers
<List any issues that prevented completion, or "None">

### Notes
<Any relevant observations or follow-up items>
---
```

## Determining Dependencies

Read the epic's README.md, specifically:
- "Execution Order" section
- "Path Dependencies Diagram" section

**Parallel indicators** (can run concurrently):
- "Can be done in parallel"
- Specs in same phase with no arrows between them
- Different feature areas with no shared files
- No "depends on" relationship

**Serial indicators** (must wait):
- "Execute in order: 1, 2, 3..."
- Arrows (→ or ↓) between specs
- "depends on", "requires", "after"
- One spec creates files another spec needs

## Error Handling

If the agent encounters issues:

**Missing dependencies**:
- Report which dependencies are needed
- Status: PARTIAL

**Unclear spec**:
- Report specific ambiguities
- Status: PARTIAL

**Validation failures**:
- Report failure details
- Attempt fix if obvious
- Status: FAILED if unfixable

**File conflicts**:
- Report conflicts
- Do not overwrite without clear spec instruction
- Status: PARTIAL

## Example: Epic Phase with Parallel Specs

For `epic-aura-incubation` Phase 2 (4 independent chores):

Main agent reads README, identifies Phase 2 specs can parallelize.

Main agent sends ONE message:

```
[Task 1]
description: "Implement voice commands"
subagent_type: "general-purpose"
prompt: |
  SPEC EXECUTOR MODE

  Spec: specs/epic-aura-incubation/chore-voice-commands.md
  Epic: epic-aura-incubation
  Beads Path: /Users/conner/.local/bin

  [rest of execution protocol...]

[Task 2]
description: "Implement beads commands"
subagent_type: "general-purpose"
prompt: |
  SPEC EXECUTOR MODE

  Spec: specs/epic-aura-incubation/chore-beads-commands.md
  Epic: epic-aura-incubation
  Beads Path: /Users/conner/.local/bin

  [rest of execution protocol...]

[Task 3]
description: "Implement planning commands"
subagent_type: "general-purpose"
prompt: |
  SPEC EXECUTOR MODE

  Spec: specs/epic-aura-incubation/chore-planning-commands.md
  Epic: epic-aura-incubation
  Beads Path: /Users/conner/.local/bin

  [rest of execution protocol...]

[Task 4]
description: "Implement execution commands"
subagent_type: "general-purpose"
prompt: |
  SPEC EXECUTOR MODE

  Spec: specs/epic-aura-incubation/chore-execution-commands.md
  Epic: epic-aura-incubation
  Beads Path: /Users/conner/.local/bin

  [rest of execution protocol...]
```

All 4 Opus instances execute concurrently. Main agent collects all reports, verifies all succeeded, then proceeds to Phase 3.

## Context Preservation

Each agent instance starts fresh. The spec file must contain all necessary context. If implementation requires knowledge from a previous spec:
1. That work must already be committed/written to files
2. The spec should reference those files explicitly
3. Do NOT run dependent specs in parallel

## Best Practices

- Always read the full spec before starting implementation
- Follow task order exactly as specified
- Run ALL validation commands, not just some
- Report blockers immediately in the report
- Keep implementation focused on spec scope (no gold-plating)
- Close beads tasks only after validation passes
