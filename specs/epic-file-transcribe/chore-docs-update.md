# Chore: Documentation Update for File Command

## Summary

Update README.md and the CLI usage string to document the new `vox file <path>` command.

## What to Change

### `README.md`

Add a new section after `### vox — Record and transcribe`:

```markdown
### `vox file <path>` — Transcribe an audio file

```bash
$ vox file memo.m4a
⠋ Transcribing...
"Refactor the sensor config to use YAML"
✓ Copied to clipboard
```

Accepts `.wav`, `.m4a`, `.mp3`, `.webm`. Does not require SoX.
```

Also add `file` to the shell aliases section if appropriate.

### `cmd/vox/main.go`

Update the usage string from:
```
Usage: vox [ls|cp|show|clear]
```
to:
```
Usage: vox [file|ls|cp|show|clear]
```

## Files to Modify

- `README.md` — Add usage section for `vox file`
- `cmd/vox/main.go` — Update usage string in the default case

## Acceptance Criteria

- [ ] README documents `vox file <path>` with example
- [ ] Usage string includes `file`
- [ ] No changes to vision.md (this is post-v1 work)
