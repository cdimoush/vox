# Feature Planning

Create a plan in specs/*.md to implement the `Feature` using the markdown `Plan Format` below.

## Instructions

- Research the codebase before planning. Start with `README.md`.
- Create the plan as `specs/<feature-name>.md`. Name it based on the `Feature`.
- Replace every `<placeholder>` with specific details.
- THINK HARD about what's needed. Keep solutions simple and focused.
- Follow existing patterns in the codebase.
- If you need a new library, use `uv add` and note it in the `Notes` section.

## Relevant Files

- `README.md` - Project overview
- `scripts/` - Python scripts (transcribe.py, etc.)
- `.claude/commands/` - Slash command definitions
- `specs/` - Example specs showing the format

## Plan Format

```md
# Feature: <feature name>

## Feature Description
<describe the feature, its purpose, and value to users>

## User Story
As a <type of user>
I want to <action/goal>
So that <benefit/value>

## Problem
<what problem does this feature solve?>

## Solution
<how will this feature solve the problem?>

## Relevant Files

### Files to Modify
<list existing files to change and why>

### New Files
<list new files to create and their purpose>

## Step by Step Tasks
<list tasks as h3 headers with bullet points. Order matters - start with foundational changes, then specific implementation. Last step should run Validation Commands.>

## Acceptance Criteria
<list specific criteria that must be met for the feature to be complete>

## Validation Commands
<list commands to verify the feature works. Focus on simple checks like:>
- `python scripts/script_name.py --help` - Verify script runs
- `cat .claude/commands/command_name.md` - Verify command content
- `ls -la expected_files` - Verify files exist

## Notes
<optional: additional context, future considerations, or dependencies>
```

## Feature
$ARGUMENTS
