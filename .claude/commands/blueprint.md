---
description: Vision to execution in one shot — blueprint, track, and optionally execute
argument-hint: <path-to-vision.md>
---

# Blueprint

Turn a vision document into a fully tracked, dependency-ordered execution plan — and optionally execute it.

## Input

```text
$ARGUMENTS
```

You **MUST** read and digest the vision document above before proceeding.

## Workflow

Execute these phases in order. Do not skip phases.

---

### Phase 1: Digest Vision

1. **Read the vision document** at the path provided in `$ARGUMENTS`
2. **Read brain context**:
   - `brain/active_projects.md`
   - `brain/glossary.md`
   - `brain/technical_systems.md` (if technical content)
   - `brain/people.md` (if people mentioned)
3. **Read the codebase**: `README.md`, `CLAUDE.md`, relevant source files referenced in or related to the vision
4. **Check existing work**:
   ```bash
   export PATH="$PATH:/Users/conner/.local/bin"
   bd list --status=open
   ```
5. **If the vision contains multiple concepts and none is chosen**: Use `AskUserQuestion` to ask which concept to pursue. Present each concept as an option with its trade-offs from the vision doc.
6. **Determine the spec directory**: Use the vision file's parent directory (e.g., `specs/003-delegation-workflow/`). The blueprint will be written here.

---

### Phase 2: Craft Blueprint

Use two sequential sub-agents to create and then review the blueprint.

#### Step 1: Writer Sub-Agent

Deploy a **general-purpose** sub-agent with the Agent tool to write the blueprint document.

**Writer prompt must include:**
- The full vision document content (paste it in)
- The chosen concept (if multiple existed)
- Summary of brain context and codebase state
- Existing beads that relate to this work
- The blueprint template (from the "Blueprint Document Template" section below)
- The target file path: `<spec-dir>/blueprint.md`
- Instruction: "Write the blueprint to the specified path. Follow the template exactly. Each work item must be concrete enough for a single agent session."

**Writer guidelines:**
- Each work item should modify ≤5-8 files
- Each work item should have ≤5 acceptance criteria
- Each work item represents ~1 focused agent session
- Phases should have clear goals and success criteria
- Dependencies must be explicit (both within-phase and cross-phase)
- Include file paths for every work item (files to create and modify)

#### Step 2: Reviewer Sub-Agent

After the writer completes, deploy a **general-purpose** sub-agent with the Agent tool to review the blueprint.

**Reviewer prompt must include:**
- Path to the blueprint just written
- The original vision document path
- Instruction to read both and produce structured feedback

**Reviewer evaluates:**
1. **Completeness**: Does the blueprint cover the full vision scope? Missing work items?
2. **Correctness**: Are dependencies right? Are file paths valid? Do acceptance criteria match descriptions?
3. **Feasibility**: Can each work item actually be completed in one session? Are items too large or too small?
4. **Quality**: Are descriptions clear enough for an agent to execute without ambiguity?

**Reviewer output format:**
```
## Blueprint Review

### Verdict: APPROVE | REVISE

### Issues Found
- [CRITICAL] ...
- [SUGGESTION] ...

### Missing Work Items
- ...

### Dependency Issues
- ...

### Recommended Changes
- ...
```

#### Step 3: Incorporate Feedback

If the reviewer returns `REVISE`:
- Read the reviewer's feedback
- Edit the blueprint to address all CRITICAL issues
- Address SUGGESTION items where they improve clarity
- You may re-run the reviewer if changes were substantial

If the reviewer returns `APPROVE`:
- Proceed to Phase 3

---

### Phase 3: User Approval

Present the blueprint summary to the user via `AskUserQuestion`.

**Question**: "Blueprint for [vision name] is ready with [N] work items across [M] phases. How would you like to proceed?"

**Options:**
1. **Approve and Execute** — Create beads tasks and begin parallel execution immediately
2. **Approve (plan only)** — Create beads tasks but don't execute. Work is tracked and ready for `/implement` or manual pickup
3. **Request Changes** — Describe what to change (iterate on blueprint)
4. **Abort** — Discard the blueprint

**If "Request Changes"**: Ask what to change, edit the blueprint, and return to Phase 3.

**If "Abort"**: Stop. Delete the blueprint file if desired.

**If "Approve and Execute" or "Approve (plan only)"**: Proceed to Phase 4.

---

### Phase 4: Convert to Beads

Create beads tasks for every work item in the blueprint.

1. **Create tasks** — For each work item, run:
   ```bash
   export PATH="$PATH:/Users/conner/.local/bin"
   bd create --title="WI-X.Y: <title>" --type=<type> --priority=2
   ```

   Use parallel sub-agents to create multiple tasks simultaneously when there are many items.

2. **Set up dependencies** — For each dependency relationship:
   ```bash
   bd dep add <dependent-id> <blocks-id>
   ```

   Dependencies to create:
   - **Within-phase sequential**: If WI-1.2 depends on WI-1.1, add that dep
   - **Cross-phase gates**: First item of Phase N depends on last item of Phase N-1
   - **Explicit deps**: Any `depends_on` listed in work item YAML

3. **Output task map** — Print a mapping:
   ```
   Blueprint → Beads Task Map

   Phase 1: <phase name>
     WI-1.1: <title> → beads-XXX
     WI-1.2: <title> → beads-YYY (depends on beads-XXX)

   Phase 2: <phase name>
     WI-2.1: <title> → beads-ZZZ (depends on beads-YYY)
     ...

   Total: N tasks, M dependencies
   ```

4. **Verify** — Run `bd ready` to confirm the first tasks are unblocked.

**If user chose "Approve (plan only)"**: Stop here. Report the task map and exit.

---

### Phase 5: Execute

Only if user chose "Approve and Execute" in Phase 3.

Execute the blueprint phase by phase, using spec-executor sub-agents.

#### For each phase:

1. **Identify ready work items** — Items whose dependencies are all complete
2. **Deploy sub-agents**:
   - **Parallel items**: Deploy ALL independent items in a SINGLE message using multiple Agent tool calls
   - **Serial items**: Deploy one at a time, waiting for completion before the next

   **Sub-agent prompt template** (use general-purpose agent):
   ```
   SPEC EXECUTOR MODE

   Blueprint: <path-to-blueprint.md>
   Work Item: WI-X.Y
   Beads ID: <beads-task-id>
   Beads Path: /Users/conner/.local/bin

   ---

   ## Execution Protocol

   ### Step 1: Claim Task
   ```bash
   export PATH="$PATH:/Users/conner/.local/bin"
   bd update <beads-id> --status in_progress
   ```

   ### Step 2: Read Context
   - Read the blueprint at the path above
   - Find work item WI-X.Y
   - Read all files listed in "Files to Create/Modify"
   - Read brain context: brain/active_projects.md, brain/glossary.md

   ### Step 3: Implement
   - Follow the work item description exactly
   - Create/modify only the files specified
   - Meet all acceptance criteria
   - Stay within scope — no gold-plating

   ### Step 4: Validate
   - Run through each acceptance criterion
   - Verify files exist and contain expected content
   - Run any test commands specified

   ### Step 5: Close Task
   ```bash
   export PATH="$PATH:/Users/conner/.local/bin"
   bd close <beads-id>
   ```

   ### Step 6: Report
   Return a structured report:

   ## Spec Execution Report

   **Work Item**: WI-X.Y: <title>
   **Beads ID**: <id>
   **Status**: SUCCESS | PARTIAL | FAILED

   ### Files Created
   - ...

   ### Files Modified
   - ...

   ### Acceptance Criteria Results
   | Criterion | Result |
   |-----------|--------|
   | ... | PASS/FAIL |

   ### Blockers
   <any issues, or "None">

   ### Notes
   <observations or follow-up items>
   ```

3. **Collect reports** — After all sub-agents for a phase complete:
   - Verify all returned SUCCESS
   - If any PARTIAL or FAILED: report to user, ask whether to continue or stop
   - Run `bd ready` to confirm next phase's items are unblocked

4. **User breakpoints** — If the blueprint marks a phase with a user breakpoint:
   - Use `AskUserQuestion`: "Phase [N] complete. [summary of what was done]. Ready to proceed to Phase [N+1]?"
   - Options: "Continue", "Review changes first", "Stop here"

5. **Final report** — After all phases complete:
   ```
   Blueprint Execution Complete

   Vision: <vision name>
   Blueprint: <blueprint path>

   Phase Results:
     Phase 1: <name> — ✓ All items complete
     Phase 2: <name> — ✓ All items complete
     ...

   Tasks Closed: N/N
   Files Created: X
   Files Modified: Y

   Run `bd list --status=open` to see remaining work.
   ```

---

## Blueprint Document Template

The writer sub-agent must produce a document following this structure exactly:

```md
# Blueprint: <name>

> <One sentence summary of what this blueprint delivers>

**Vision**: [<vision name>](<relative-path-to-vision.md>)
**Concept**: <chosen concept name, if applicable>

## Overview

<2-3 paragraphs: what gets built, why, and how it fits into the larger system>

## Existing State

<What already exists that this blueprint builds on. Reference specific files and beads.>

---

## Phase 1: <phase name>

**Goal**: <what this phase achieves>
**Success Criteria**: <how to verify the phase is done>

### Dependency Graph

```
WI-1.1 ──→ WI-1.2
  │
  └──→ WI-1.3 (parallel with WI-1.2)
```

### WI-1.1: <title>

```yaml
type: feature | chore | bug
complexity: S | M | L | XL
depends_on: []
parallel_with: []
```

**Description**: <what to build and why>

**Files to Create/Modify**:
- `path/to/new_file.py` — (create) description
- `path/to/existing.py` — (modify) what changes

**Acceptance Criteria**:
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

### WI-1.2: <title>

<same structure>

---

**🔄 USER BREAKPOINT** *(optional)*: <what to validate before proceeding>

---

## Phase 2: <phase name>

<same structure as Phase 1, with cross-phase deps noted in depends_on>

---

## Summary

| Phase | Items | Parallel | Serial | Complexity |
|-------|-------|----------|--------|------------|
| 1: <name> | N | X | Y | S/M/L |
| 2: <name> | N | X | Y | S/M/L |
| **Total** | **N** | | | |

## Risks & Mitigations

- **Risk**: <description> → **Mitigation**: <approach>

## Notes

<Any additional context, future work triggered by this blueprint, or cross-cutting concerns>
```

---

## Key Rules

- **One blueprint per vision concept**. If the vision has multiple concepts, the user must choose one first.
- **Work items are agent-sized**. Each WI should be completable in one focused session (≤8 files, ≤5 acceptance criteria).
- **Writer and reviewer are separate agents**. This prevents sunk-cost bias in the writer.
- **Execution is always opt-in**. Never auto-execute without explicit user approval.
- **Reuse spec-executor pattern**. Sub-agents follow the same protocol as `.claude/agents/spec-executor.md`.
- **Beads are the source of truth**. Once Phase 4 creates tasks, beads tracks all progress.
- **Parallel by default**. Items without dependencies should run concurrently. Only serialize when dependencies require it.
- **Brain context first**. Always read brain files before writing or executing anything.
