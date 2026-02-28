# Feature Development Agent

A sub-agent for implementing individual features or chores from spec files. Use this agent to offload development work and keep the main agent context focused.

## Agent Capabilities

This agent has full privileges to:
- Read and analyze code files
- Create and modify files
- Run shell commands (builds, tests, scripts)
- Install dependencies via `uv add`
- Create directories and organize code

## When to Use This Agent

Deploy `feature-dev` when:
- Implementing a feature or chore from a spec file in `specs/`
- Working through an epic's individual specs
- The main agent needs to delegate focused implementation work

## Execution Mode Guidelines

### Serial Execution (One at a Time)
Use serial execution when specs have dependencies:
- Spec B requires code/files created by Spec A
- Later specs import from or extend earlier work
- A chore sets up infrastructure that features depend on

Example: "Setup directories" must complete before "Queue processor" can use them.

### Parallel Execution (Concurrent)
Use parallel execution when specs are independent:
- Different feature areas with no shared code
- Documentation updates for separate features
- Platform-specific implementations (macOS vs Ubuntu guides)

Example: "macOS hotkey guide" and "Ubuntu hotkey guide" can run simultaneously.

## Invocation

The main agent should invoke this agent using the Task tool:

```
Task tool with subagent_type="general-purpose"
prompt: "Implement the spec at specs/<path-to-spec>.md. Follow all step-by-step tasks in order. Run validation commands when complete. Report: files created/modified, changes made, validation results."
```

For parallel execution, send multiple Task tool calls in a single message.

## Agent Instructions

When invoked, this agent will:

1. **Read the spec file** provided in the prompt
2. **Read relevant existing files** referenced in the spec
3. **Follow step-by-step tasks** in exact order specified
4. **Implement each task** completely before moving to the next
5. **Run validation commands** at the end
6. **Report results** in structured format

## Expected Input

The agent expects a prompt containing:
- Path to a spec file (e.g., `specs/feature-queue-processor.md`)
- Any additional context needed (usually none - spec should be self-contained)

## Expected Output

The agent returns a structured report:

```
## Implementation Complete

### Files Created
- path/to/new/file.py - description

### Files Modified
- path/to/existing/file.py - what changed

### Changes Summary
Brief description of what was implemented.

### Validation Results
- Command 1: PASS/FAIL + output summary
- Command 2: PASS/FAIL + output summary

### Notes
Any issues encountered, assumptions made, or follow-up items.
```

## Error Handling

If the agent encounters blocking issues:
- Missing dependencies: Report which dependencies are needed
- Unclear spec: Report specific ambiguities
- Test failures: Report failure details and attempted fixes
- File conflicts: Report conflicts and recommended resolution

The main agent can then decide to fix issues and retry, or escalate to the user.

## Context Preservation

Each agent invocation starts fresh. The spec file should contain all necessary context. If implementation requires knowledge from a previous spec, that information should be:
1. Already committed to files the agent can read
2. Explicitly included in the spec's "Relevant Files" section
