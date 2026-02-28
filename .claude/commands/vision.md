---
description: Create high-level strategy with multiple implementation paths
argument-hint: <vision description or topic>
---

# Vision Planning

Create a high-level strategy document in specs/vision-*.md that explores the `Vision` using the `Vision Format` below.

## Instructions

- **Read brain context first**: `brain/active_projects.md`, `brain/technical_systems.md`, `brain/glossary.md`
- Research the codebase before planning. Start with `README.md`.
- Analyze the current state and identify gaps between where we are and where we want to be.
- Propose multiple implementation paths with clear trade-offs for each.
- Paths can be independent (choose one), partially shared (common foundation), or sequential (path B depends on A).
- Include a recommendation but leave the final decision to the user.
- Create the document as `specs/vision-<name>.md`. Name it based on the `Vision`.
- Replace every `<placeholder>` with specific details.
- THINK HARD about the big picture. This is strategic planning, not tactical execution.
- Reference brain files for context on projects, technical systems, and terminology.

## Relevant Files

- `README.md` - Project overview
- `specs/` - Example specs and vision documents
- `docs/` - Architecture documentation
- `brain/active_projects.md` - Current project context
- `brain/technical_systems.md` - Technical architecture reference
- `brain/claudes_wants_this.md` - Strategic priorities

## Vision Format

```md
# Vision: <vision name>

## Vision Statement
<one paragraph describing the desired future state and why it matters>

## Current State Analysis

### What Exists Today
<bullet points describing current capabilities and infrastructure>

### Current Pain Points
<numbered list of specific problems users face today>

### The Gap
<describe the difference between current state and desired state>

## Vision Goals
<numbered list of high-level goals this vision aims to achieve>

## Implementation Paths

### Path 1: <path name>
**Approach**: <brief description of this approach>

**Effort**: <Low | Medium | High>
**Impact**: <Low | Medium | High>
**Risk**: <Low | Medium | High>

**Dependencies**: <what must exist before this path can start>

**High-Level Features**:
<numbered list of features this path would deliver>

**High-Level Chores**:
<numbered list of supporting work needed>

**Trade-offs**:
- ✅ <advantage>
- ✅ <advantage>
- ❌ <disadvantage>
- ❌ <disadvantage>

---

### Path 2: <path name>
<same structure as Path 1>

---

### Path 3: <path name> (if applicable)
<same structure as Path 1>

---

## Path Comparison Matrix

| Criteria | Path 1 | Path 2 | Path 3 |
|----------|--------|--------|--------|
| Time to Value | <Fast/Medium/Slow> | | |
| Technical Complexity | <Low/Medium/High> | | |
| User Impact | <Low/Medium/High> | | |
| Maintenance Burden | <Low/Medium/High> | | |
| Scalability | <Low/Medium/High> | | |

## Recommended Approach

<describe the recommended sequence of paths with rationale>

**Why This Order:**
<numbered list explaining the reasoning behind the recommendation>

## Path Dependencies Diagram

```
<ASCII diagram showing relationships between paths>
```

## Next Steps

<numbered list of immediate actions to take after approving this vision>

## Open Questions

<numbered list of questions that need user input, with recommendations where possible>

## Future Considerations

<numbered list of ideas that came up during planning but are out of scope>

---

## Research Sources

<optional: links to relevant documentation, articles, or prior art that informed this vision>
```

## Vision
$ARGUMENTS
