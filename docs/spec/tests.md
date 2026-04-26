# vox-core — test plan

> Companion to `vox-core.md`. Every assertion in the spec must trace to at least one test here.

## Three tiers, increasing cost

- **Unit** (no network, no audio, milliseconds): pure functions — FindAPIKey across all discovery paths, JSONL parser/encoder, exit-code mapping, CLI flag parsing, clipboard tool detection
- **Component** (no network, real fixtures): mocked Whisper client, real WAV/M4A/MP3 inputs, exercise chunking thresholds + concatenation + JSON output shapes
- **Live** (real Whisper, gated by `VOX_LIVE_TESTS=1` env var): commit a small audio fixture, run the real binary against the real OpenAI API, assert transcript is similar (not exact-match) to expected text. Default `go test ./...` never hits the network

## Coverage targets (every bullet → at least one test)

- Every CLI command happy-path: `vox file`, `vox ls -n 5`, `vox cp 1`, `vox show 1`, `vox clear` with stdin "y" / "n", `vox login` with stdin
- Every JSON contract round-tripped (encode → decode → field-by-field assert)
- Every exit code triggered intentionally: 1 = bad path; 2 = mocked API 5xx; 3 = no key anywhere; 0 = happy
- Every FindAPIKey discovery path: env, `~/.vox/config`, each shell profile in order, all-empty
- Every supported audio extension (smoke: file exists, format detected, transcribe attempted)
- Chunking boundary: just-under and just-over 8 minutes (mocked Whisper, real WAV duration probe)
- History: append → list → list with -n → list with --all → cp → show → clear → list-after-clear (file missing is not error)
- Stdin pipe path: `cat fixture.wav | vox file - --format=wav`
- Stderr/stdout discipline: capture both streams, assert transcript only on stdout, "Copied" only on stderr

## Fixtures (committed in repo)

- `testdata/audio/testaudio.m4a` — short deterministic phrase, ~few seconds. Used by both component and live tests
- `testdata/audio/testaudio.expected.txt` — the canonical transcript text. Used as the assertion target in the live test (similarity-matched, not exact-match)
- `testdata/audio/short_silence.wav` — ~1 second silence (tests empty-transcript handling) — TODO add later
- `testdata/audio/over_eight_min.wav` — synthetic; generated on-demand by a test helper rather than committed (file would be too large)
- `testdata/golden/<scenario>/{stdin,args,env,expected_stdout.json,expected_stderr.txt,expected_exit}` — black-box golden runner format. One directory per scenario. CLI args parameterized in `args`; env in `env`; expected outputs side by side
- `testdata/history/<scenario>/{before.jsonl,after.jsonl}` — for state-mutation tests

## Audio round-trip via real Whisper

- `testaudio.m4a` is committed once and used in two ways:
  - **Component**: a mocked Whisper client returns a canned response; assert vox's wiring (stdout, history append, clipboard write, exit code) reacts correctly
  - **Live**: `VOX_LIVE_TESTS=1 go test -run LiveAudioRoundTrip ./...` actually calls OpenAI; assert the transcript is "similar" to `testaudio.expected.txt` via normalized comparison (lowercase, strip punctuation, optional Levenshtein threshold)
- The live test is opt-in. Default `go test ./...` stays network-free

## Existing-test audit

- `cmd/vox/file_test.go` (160 LOC) → **rewrite** as golden-runner tests against the new `testdata/golden/` layout
- `cmd/vox/format_test.go` (65 LOC) → **keep** (small, focused, format detection logic)
- `config/config_test.go` (129 LOC) → **rewrite** to cover every FindAPIKey path explicitly
- `recorder/recorder_test.go` (114 LOC) → **keep with audit** — recorder integration tests skip if SoX missing, that pattern stays
- `transcribe/transcribe_test.go` (24 LOC) + `transcribe/chunk_test.go` (81 LOC) → **rewrite** to use mocked Whisper + chunking-boundary fixtures
- `clipboard/clipboard_test.go` (59 LOC) → **keep**
- `history/history_test.go` (156 LOC) → **rewrite** to use the `testdata/history/<scenario>` before/after format
- All `daemon/*_test.go` and `cmd/vox/{daemon,ui}_test.go` → **delete** in vox-1yh phase 2 (out of scope here)

## What we deliberately don't test

- Performance benchmarks
- Real OpenAI failure modes — mock them; OpenAI's actual reliability is theirs to test
- Cross-platform clipboard tools beyond detection (CI doesn't have all of them)
