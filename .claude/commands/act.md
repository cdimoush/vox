---
allowed-tools: Bash(python:*), Bash(uv:*), Bash(mkdir:*), Bash(mv:*), Bash(date:*), Read, Write, Edit, Glob, Grep, WebFetch, WebSearch
description: Transcribe audio and act on the request
argument-hint: <audio-file-path> [--project=<project-name>]
---

# Act on Audio

Transcribe the provided audio file, generate an intelligent title, and execute the request spoken within it.

## Step 0: Load Brain Context

**IMPORTANT**: Before transcribing, read relevant brain files for context.

**Always include**:
- `brain/active_projects.md` - Current priorities and projects
- `brain/glossary.md` - Terminology and technical terms

**Include based on project flag or content**:
- If `--project=design-lab` or memo mentions Design Lab → `brain/technical_systems.md`
- If `--project=simulation-mandate` or mentions Simulation Mandate → `brain/technical_systems.md`, `brain/people.md`
- If strategic planning topics → `brain/claudes_wants_this.md`
- If mentions people (Nick, Shiva, Melissa) → `brain/people.md`

**Check for project flag**:
```bash
# Extract project flag from $ARGUMENTS if present
# Example: $ARGUMENTS = "audio.m4a --project=simulation-mandate"
```

If project flag is provided, use it to determine context. Otherwise, load context after transcription based on content analysis.

## Step 1: Transcribe the Audio

Run the transcription script:

```bash
uv run python scripts/transcribe.py $ARGUMENTS
```

Save the full transcription text for use in subsequent steps.

## Step 2: Generate Intelligent Title

Use the transcription to generate a memorable, descriptive title:

```bash
uv run python scripts/generate_title.py --text "{transcription_text}"
```

Save the generated title (e.g., `api-refactor-discussion`).

## Step 3: Create Timestamp

Get the current timestamp:

```bash
date +%Y-%m-%d_%H-%M-%S
```

Save the timestamp (e.g., `2026-01-08_14-30-22`).

## Step 4: Analyze the Transcription

Analyze the transcription to identify:

1. **Project Context**: What project(s) does this relate to?
   - Check against `brain/active_projects.md`
   - Look for mentions of: Design Lab, Simulation Mandate, Whisper, Hologram, Apollo 3
   - Load additional brain context if not already loaded (technical_systems.md, people.md)

2. **Request Type**: What is the user asking for?
   - **Summary**: User wants a summary of ideas, thoughts, or content they described
   - **Research**: User wants research on a topic, including sources and analysis
   - **Code**: User wants code, scripts, or technical implementation
   - **Planning**: User wants a plan, outline, or structured approach
   - **Other**: Any other actionable request

3. **Key Details**: Extract the main topics, requirements, constraints, and goals mentioned
   - Reference brain files for context (technical terms from glossary.md, people from people.md)

4. **Deliverables**: What output files should be created?

5. **Tags**: What tags apply? (see `brain/tags.md`)
   - Project tags (design-lab, simulation-mandate, whisper, etc.)
   - Technology tags (isaac-sim, ros2, etc.)
   - People tags (nick-cto, shiva-vp, etc.)
   - Process tags (strategic-planning, meta-work, etc.)

## Step 5: Create Output Directory

Create the output directory using the generated title and timestamp:

```bash
mkdir -p "output/{title}_{timestamp}/"
```

Example: `output/api-refactor-discussion_2026-01-08_14-30-22/`

Store the directory path for use in subsequent steps.

## Step 6: Create README.md

In the output directory, create a `README.md` with the following structure:

```markdown
# {Generated Title}

## Source
- **Audio File**: [original filename] (included in this directory)
- **Transcribed**: [timestamp]

## Transcription Summary
[2-3 sentence summary of what was spoken]

## Project Context
[Which project(s) this relates to - reference brain/active_projects.md]

## Request Identified
- **Type**: [Summary/Research/Code/Planning/Other]
- **Description**: [What the user is asking for]

## Deliverables
- [ ] [List of files that will be created]
- [ ] [Each file with brief description]

## Brain Context Used
[List which brain files were referenced for context]
- brain/active_projects.md
- brain/technical_systems.md (if applicable)
- brain/people.md (if applicable)
- etc.

## Full Transcription
<details>
<summary>Click to expand full transcription</summary>

[Full transcription text]

</details>

---

## Tags

`[tag1]` `[tag2]` `[tag3]` (see brain/tags.md)
```

## Step 7: Execute the Request

Based on the request type, create the appropriate deliverables in the output directory:

### For Summary Requests
- Create `summary.md` with structured summary including:
  - Key points
  - Main ideas
  - Action items (if any)
  - Conclusions

### For Research Requests
- Create `research.md` with:
  - Overview of the topic
  - Key findings (use WebSearch/WebFetch to gather current information)
  - Sources and references
  - Analysis and recommendations

### For Code Requests
- Create implementation files with appropriate extensions
- Include comments explaining the code
- Create a `usage.md` if the code needs documentation

### For Planning Requests
- Create `plan.md` with:
  - Goals and objectives
  - Step-by-step approach
  - Considerations and tradeoffs
  - Next steps

### For Other Requests
- Interpret the request and create appropriate output
- Document your interpretation in the README

## Step 8: Update README with Completion Status

After creating all deliverables, update the README.md:
- Check off completed deliverables
- Add any notes about the output
- Include suggestions for follow-up if relevant

## Step 9: Move Audio to Output Directory

Move the processed audio file into the output directory alongside the transcription:

```bash
mv "$ARGUMENTS" "output/{title}_{timestamp}/"
```

This keeps everything together - the audio, transcription, and deliverables are all in one self-contained directory.

## Important Notes

- Always create the output directory first before writing any files
- If the request is unclear or contains multiple distinct asks, focus on the primary request and note others in the README
- If research is requested, use WebSearch and WebFetch to gather current information
- Maintain a professional, organized structure in all output files
- Only move the audio after successful processing - if any step fails, leave the file in place for retry
