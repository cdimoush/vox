---
description: Break a vision into ordered specs with dependencies
argument-hint: <vision-doc.md or epic description>
---

# Epic Planning

Create an epic document in specs/epic-*/README.md that breaks a vision into ordered specs using the `Epic Format` below.

## Instructions

- **Read brain context first**: `brain/active_projects.md`, `brain/technical_systems.md`, `brain/glossary.md`
- If a vision document is provided, read it first to understand the strategic context.
- If no vision document is provided, create a focused epic for the given scope.
- Break the chosen path into ordered feature and chore specs.
- Define execution order with explicit dependencies between specs.
- Include user testing breakpoints (🔄) to validate assumptions before proceeding.
- Create the document as `specs/epic-<name>/README.md`.
- Replace every `<placeholder>` with specific details.
- THINK HARD about dependencies and order. Specs should build on each other logically.
- Reference brain files for context on technical systems and project priorities.

## Relevant Files

- `README.md` - Project overview
- `specs/vision-*.md` - Vision documents to reference
- `specs/epic-*/` - Example epic structures
- `.claude/commands/feature.md` - Feature spec format
- `.claude/commands/chore.md` - Chore spec format
- `brain/active_projects.md` - Current project context
- `brain/technical_systems.md` - Technical architecture
- `brain/claudes_wants_this.md` - Strategic priorities

## Epic Format

```md
# Epic: <epic name>

## Epic Overview

<2-3 paragraphs describing the epic's purpose, scope, and expected outcome>

## Vision Context

**Source Vision**: [Vision: <name>](../vision-<name>.md)

### Key Vision Points

**Current Pain Points Addressed:**
<bullet points from the vision's pain points that this epic solves>

**Vision Goals:**
<numbered list of goals from the vision that this epic achieves>

**Architecture Decision (if applicable):**
<key architectural decisions that inform this epic's approach>

## Specs in This Epic

### Phase 1: <phase name>
- [ ] [Chore: <name>](./<filename>.md) - <brief description>
- [ ] [Feature: <name>](./<filename>.md) - <brief description>

### Phase 2: <phase name>
- [ ] [Feature: <name>](./<filename>.md) - <brief description>
- [ ] [Chore: <name>](./<filename>.md) - <brief description>

<add more phases as needed>

## Execution Order

### Phase 1: <phase name> (<estimated effort>)
**Goal**: <what this phase achieves>

Execute in order:
1. [Chore/Feature: <name>](./<filename>.md) - <why this is first>
2. [Chore/Feature: <name>](./<filename>.md) - <why this follows>

**Success Criteria**:
<bullet points defining what "done" looks like for this phase>

**🔄 USER BREAKPOINT #1**: <what to validate before proceeding>

---

### Phase 2: <phase name> (<estimated effort>)
**Goal**: <what this phase achieves>

<same structure as Phase 1>

---

<add more phases as needed>

## Path Dependencies Diagram

```
<ASCII diagram showing spec dependencies and critical path>

Phase 1
    ↓
Phase 2
    ├─ Spec A (must exist first)
    └─ Spec B (depends on A)
    ↓
Phase 3

Critical Path:
- <list the must-have sequence>
```

## Implementation Notes

### Cross-Cutting Concerns

**Architecture Decisions:**
<bullet points about key technical choices that affect multiple specs>

**Shared Dependencies:**
<bullet points about libraries, tools, or infrastructure used across specs>

**Testing Strategy:**
<bullet points about how to validate the epic as a whole>

**Rollout Plan:**
<bullet points about incremental adoption and rollback strategy>

### User Testing Breakpoints

This epic includes **<N> explicit user testing breakpoints** (marked with 🔄):
<numbered list summarizing each breakpoint>

Each breakpoint is a decision point: proceed to next phase, iterate on current phase, or stop if goals are met.

## Success Metrics

- [ ] <measurable outcome 1>
- [ ] <measurable outcome 2>
- [ ] <measurable outcome 3>

## Future Enhancements

Ideas that came up during planning but are out of scope for this epic:

<numbered list of future work items>
```

## Creating Individual Specs

After the epic README is complete, create individual spec files for each item:

- For features: Follow the format in `.claude/commands/feature.md`
- For chores: Follow the format in `.claude/commands/chore.md`

Name spec files as:
- `feature-<name>.md` for features
- `chore-<name>.md` for chores

Place all spec files in the `specs/epic-<name>/` directory alongside the README.

## Epic
$ARGUMENTS
