---
allowed-tools: Bash(mkdir:*), Task, Glob, Skill
description: Process all audio files in queue directory
---

Process all audio files in the queue/ directory using parallel sub-agents that invoke `/act`.

## Instructions

1. **Ensure directories exist**: Create queue and output directories if missing:
   ```bash
   mkdir -p queue output
   ```

2. **Discover queue files**: Use Glob to find all audio files in `queue/` with pattern `queue/*.{m4a,mp3,wav,mp4,mpeg,mpga,webm}`

3. **Handle empty queue**: If no files found, display:
   ```
   Queue is empty. Add audio files to queue/ directory.
   ```

4. **Process files in parallel**: For each audio file, spawn a Task agent (subagent_type: "general-purpose", model: "sonnet") with this prompt:

   ```
   Process this audio file by invoking the /act command. Working directory: /Users/conner/dev/whisper

   Audio file: {full_path_to_audio_file}

   Use the Skill tool to invoke the "act" skill with the audio file path as the argument:
   - skill: "act"
   - args: "{full_path_to_audio_file}"

   The /act command will:
   1. Transcribe the audio
   2. Generate an intelligent title
   3. Create an output directory with title and timestamp
   4. Analyze the request and create deliverables
   5. Move the audio file into the output directory

   Report back: "SUCCESS: {filename} -> [output directory created]" or "FAILED: {filename} - {error}"
   ```

5. **Spawn all agents in parallel**: Use multiple Task tool calls in a single response block.

6. **Report summary** after all agents complete:
   ```
   Queue Processing Complete

   Total files: X
   Successfully processed: Y
   Failed: Z

   Output directories created in: output/
   (Each output directory contains the source audio + transcription + deliverables)
   ```

## Error Handling

- If a sub-agent fails, log the error but continue processing other files
- The /act command only moves audio on success - failed files remain in queue for retry
- If queue/ directory doesn't exist, create it and display empty queue message
