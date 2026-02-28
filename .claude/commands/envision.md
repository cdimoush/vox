---
description: Explore a software vision through tight/loose idea classification and 3-5 mixed concepts
argument-hint: <what you want the software to do, at a high level>
---

## Vision Input

```text
$ARGUMENTS
```

You **MUST** consider the vision input above before proceeding.

## Purpose

Turn a high-level software desire into a collaborative **vision.md** document. Unlike rigid implementation plans, this is exploratory. The goal is to surface what we know (tight ideas) and what we don't (loose ideas), then remix them into 3-5 concept variants the user and agent can discuss and refine together.

## Workflow

### 1. Understand the Vision

Read the user's input. Identify the core desire — what should this software *feel like* and *do* at a level anyone could understand. Don't jump to implementation. Restate the vision in plain language.

### 2. Research Context

- Read `brain/active_projects.md`, `brain/glossary.md`, and any relevant brain files
- Read the codebase (`README.md`, relevant source files, existing specs)
- Check `bd list --status=open` for related existing work
- Understand what already exists vs what's new

### 3. Extract Ideas

From the vision input and research, extract every distinct idea, requirement, or capability mentioned or implied. For each idea, classify it:

**Tight** — Ready for implementation. The what, why, and how are clear enough to write code.
- Has a clear scope
- No major unknowns
- Could be turned into a task or feature today

**Loose** — Needs answers first. The desire is clear but the path isn't.
- Has open questions that change the approach
- Depends on decisions not yet made
- Multiple valid interpretations exist

### 4. Surface Questions for Loose Ideas

For each loose idea, write 1-3 specific questions that would make it tight. These aren't rhetorical — they're genuine blockers. Frame them as decisions the user needs to make.

### 5. Compose 3-5 Concepts

Create 3-5 concept variants that mix-and-match the tight and loose ideas in different ways. Each concept is a coherent vision of the software — not a phased plan, but a *snapshot of what it could be*.

Concepts should differ meaningfully:
- Different trade-offs (simple vs powerful, fast vs thorough)
- Different scopes (MVP vs full vision)
- Different architectural bets (which loose ideas you resolve which way)
- Different user experiences

Each concept inherits some tight ideas and takes a position on some loose ones.

### 6. Create Beads for Planning Work

Use `bd create` to track the planning work itself:
- One issue for the vision document (this work)
- Issues for any research needed to tighten loose ideas
- Issues for follow-up decisions the user needs to make

Use dependencies (`bd dep add`) when research must happen before decisions.

### 7. Write vision.md

Write the document to `specs/<feature-dir>/vision.md` (use the current branch's spec directory if it exists, otherwise `specs/vision-<name>.md`).

## Vision Document Template

Write the following structure. Replace all `<placeholders>` with real content. Remove any sections that don't apply.

```md
# Vision: <name>

> <One sentence: what this software should do, in plain language anyone would understand>

## The Desire

<2-3 paragraphs explaining what the user wants. No jargon. Write it like you're explaining to a smart friend who doesn't code. What problem does this solve? What does success look like?>

## What Exists Today

<Brief inventory of current state. What's already built? What can we leverage?>

---

## Ideas

### Tight Ideas (Ready to Build)

<For each tight idea:>

#### T1: <idea name>
<What it is, why it's clear, what it would take.>

#### T2: <idea name>
...

### Loose Ideas (Needs Answers)

<For each loose idea:>

#### L1: <idea name>
<What we want but don't fully know how to get.>

**Open questions:**
1. <specific question that would make this tight>
2. <specific question>

#### L2: <idea name>
...

---

## Concepts

> Each concept is a coherent snapshot of what this software could be.
> They mix tight and loose ideas differently.

### Concept 1: <name — e.g. "The Simple Path">

**In a sentence:** <what this version of the software is>

**Includes:** T1, T3, L2 (resolved as: <position taken>)

**Skips:** L1, T4

**Character:** <1-2 words capturing the personality — e.g. "Fast and scrappy", "Polished and complete", "Ambitious bet">

<2-3 paragraphs describing this concept as if it already exists. What does the user experience? What's the workflow?>

**Trade-offs:**
- Gets you: <what you gain>
- Costs you: <what you give up>

---

### Concept 2: <name>
<same structure>

---

### Concept 3: <name>
<same structure>

---

<Concepts 4-5 if warranted>

## Concept Comparison

| | Concept 1 | Concept 2 | Concept 3 |
|---|---|---|---|
| Scope | | | |
| Complexity | | | |
| Time to something useful | | | |
| Loose ideas resolved | | | |
| Biggest bet | | | |

## Decisions Needed

<Numbered list of decisions the user needs to make. Reference the loose ideas and which concepts depend on which answers.>

1. **<Decision>**: <context>. Affects concepts <X, Y>.
2. ...

## Next Steps

<What happens after the user picks a direction or answers questions. Reference beads issues created.>

---

*Generated from vision input. This is a living document — refine it through conversation.*
```

## Key Rules

- Write for humans first. No implementation details unless they clarify the vision.
- Concepts are not phases. They're parallel possibilities, not sequential steps.
- Be honest about what's loose. Don't fake certainty.
- 3 concepts minimum, 5 maximum. Each must be genuinely different.
- Use `bd create` and `bd dep add` for trackable planning work, not TodoWrite.
- Don't implement anything. This is pure vision work.
