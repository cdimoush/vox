# Commercial Dictation Products — UX Research

*Research bead: system_designer-d8s.2*
*Focus: overlay UX, recording indicators, hotkey design, auto-paste patterns*

---

## 1. macOS Built-in Dictation

**Activation:** Double-tap `Fn` key (default), or user-customized shortcut. A dedicated microphone key exists on newer keyboards.

**Recording Indicator:**
- A small floating **feedback window** appears near the cursor/insertion point.
- Shows a **microphone icon with a fluctuating loudness/waveform bar** — real-time volume visualization.
- An audio tone signals when the Mac is ready to listen.
- In macOS Sonoma+, this condensed to a small microphone icon that appears in-context.

**Text Insertion:**
- Text appears inline at the cursor position as you speak — live streaming transcription.
- Ambiguous words are underlined in **blue**; clicking shows alternative interpretations.
- Punctuation via voice commands ("exclamation mark", "new paragraph").

**Ending Dictation:** Press the shortcut again, hit Return, or click "Done" in the feedback window.

**Language Switching:** Click the language label in the feedback window to switch mid-session.

**macOS Tahoe (2025):** Introduced a translucent "Liquid Glass" overlay with real-time waveform visualization and confidence-level highlighting on transcribed words.

**Key UX Pattern:** Inline insertion at cursor, minimal floating UI anchored near cursor, no separate panel. Modal — no push-to-talk, it's always active until stopped.

---

## 2. Windows Speech Recognition

**Activation:** `Win+H` opens the voice typing panel (Windows 11); older WSR uses `Win+Ctrl+H` or manual start.

**Recording Indicator:**
- A **status bar/panel** floats (typically center-top of screen).
- Shows text: "Listening…", "Thinking…", "Did you say…?" (disambiguation screen).
- A **voice meter** displays live volume visualization.
- A **microphone glyph button** in the panel shows availability and recording state.

**States communicated:**
1. Listening (active input)
2. Thinking (processing)
3. Disambiguation ("Did you say X or Y?")
4. Sleeping/inactive

**Text Insertion:** Directly at the cursor in the focused window.

**Design Guidance (Microsoft):** Recommends developers use a command bar button with microphone glyph to show both availability and state; provide ongoing feedback to avoid "apparent lack of response."

**Key UX Pattern:** Floating top-center panel with explicit state labels. Three-state model (listening / thinking / confirming) with named screens for each. Separation of recording UI from the cursor location.

---

## 3. Dragon NaturallySpeaking (Nuance Dragon Professional)

**Activation:** Hotword ("Wake up Dragon") or hotkey combination.

**Recording Indicator:**
- The **DragonBar** — a persistent toolbar across the top of the screen — shows current recognition mode and status.
- A small **Dragon logo indicator** appears at the insertion point during active transcription.
- Words appear in a dictation box first, then get committed to the target application (or directly inline for trained apps).

**Visual Feedback:**
- The DragonBar shows a microphone icon with recording state.
- Processing is indicated by the Dragon logo pulsing at the cursor.
- Recognition confidence is not directly displayed; errors require voice corrections ("correct that", "scratch that").

**UX Characteristics:**
- Heavy UI footprint — persistent toolbar, not just on-demand overlay.
- Strong "dictation box" pattern for uncertain applications: type into a Dragon-managed floating text area, then transfer.
- Very mature command vocabulary for hands-free navigation: "click link", "mouse grid", "choose N".

**Key UX Pattern:** Persistent overlay toolbar (DragonBar) rather than ephemeral indicator. Dictation box as safe intermediate text stage. Voice-correction workflow, not click-to-fix.

---

## 4. Talon Voice

**Activation:** Always-on background process. Modes switched via voice commands:
- `"wake up"` → enables recognition
- `"go to sleep"` → disables
- `"dictation mode"` → plain text transcription
- `"command mode"` → voice command interpretation

**Recording Indicator:**
- **Talon HUD** (optional community plugin): a HUD panel in the corner of the screen (typically taskbar-adjacent) displaying current mode, recent commands recognized, and active scripts.
- **Mode Indicator** (optional community plugin): a small icon showing current mode state (command / dictation / sleep).
- Without HUD: no native visual indicator; users rely on learned audio cues and behavior.

**Design Philosophy:**
- Talon is infrastructure, not a finished product. The visual layer is composable via Python scripts.
- Non-intrusive by default: minimal UI, status comes through behavior (words appearing) rather than indicators.
- Very low latency — feedback comes from text appearing instantly.
- Modal operation is fundamental: Talon explicitly separates "commanding" vs "dictating" vs "sleeping" as distinct states.

**Key UX Pattern:** Mode-based activation (not push-to-talk or toggle). Visual state is optional and user-configured. The system is always partially listening; modes gate what happens with audio.

---

## 5. MacWhisper

**Activation:** Customizable global keyboard shortcut triggers the **Global Overlay**.

**Recording Indicator:**
- A floating overlay window appears above all apps ("always on top" by default).
- The window is persistent across app switches unless "Always on Top" is disabled.
- Recording indicator turns green while active, processes shown with color change.

**Features:**
- **Auto Start**: Recording begins automatically when Global opens.
- **Auto Copy**: Transcript automatically copied to clipboard after transcription completes.
- Text must then be manually pasted (clipboard-based workflow, not direct injection).

**Key UX Pattern:** Floating panel that persists over all apps. Clipboard as the integration layer (no direct text injection). Simple color-coded status.

---

## 6. Superwhisper

**Activation:** Global hotkey (default: single modifier key like `Fn` or `Right Cmd`). Configurable.

**Recording Window:**
- A **mini floating window** appears during recording — compact but visible.
- Displays a **live audio waveform** during recording.
- A **color-coded status dot**:
  - Yellow = model loading
  - Blue = processing/transcribing
  - Green = complete
- Shows the active mode name and its keyboard shortcut.
- A **context indicator** lights up when clipboard or selected text was captured within the 3-second pre-recording window.

**Interaction:**
- Hover over mini window → Stop button appears.
- Cancel for recordings under 30s is instant; longer recordings require confirmation.
- Right-click for quick settings access and history.

**Hotkey Modes:**
- **Toggle**: press once to start, press again to stop.
- **Push-to-talk**: hold to record, release to stop.
- **Both can share the same key** (short tap = toggle, hold = push-to-talk).
- Mouse button support: quick click = toggle, hold = push-to-talk.
- **Mode cycling**: hold modifier + tap key to cycle modes.

**Auto-paste:** Text is injected directly at the cursor location in the active app.

**Super Mode:** Context-aware — reads selected text, clipboard, and active app to adapt formatting and style. Intelligently routes to appropriate processing.

**User Feedback:** Some users flagged the overlay as "intrusive" (big window), leading to the mini window option. Settings surface described as overwhelming for new users.

**Key UX Pattern:** Color-coded tri-state indicator (loading/processing/done). Flexible hotkey that supports both toggle and push-to-talk semantics. Context awareness via 3-second pre-capture window. Direct cursor injection.

---

## 7. Wispr Flow

**Activation:** `Fn` key (Mac default) or `Ctrl+Win` (Windows). Mouse button support (middle-click, Mouse 4–10). A "Flow Bar" in screen center offers clickable activation.

**Recording Indicator:**
- **White waveform bars** animate in the interface when recording starts.
- Audio "ping" cue signals readiness.
- On Android: a **floating "Flow Bubble"** appears over all apps, draggable and edge-snapping, with a pulsing recording indicator and waveform during hold-to-dictate.
- On iOS: a "purple spinning glow ring" animates around the start button.

**Text Insertion:**
- Automatically pastes formatted text at the cursor location.
- Tracks which text field was last focused to target insertion correctly.
- Fallback: if auto-paste fails, prompts user to paste manually via shortcut.

**AI Processing:**
- Filler word removal, punctuation correction, tone adjustment.
- "Command Mode" for AI-directed editing via voice.

**Key UX Pattern:** Hold `Fn` = push-to-talk (most common). Waveform feedback during capture. AI-polished output, not raw transcription. Mobile uses a persistent draggable bubble (Android-style floating action button for voice).

---

## 8. TypeWhisper (macOS)

**Indicator Styles — Three Options:**
1. **Notch** — integrates into the MacBook notch area; minimal screen intrusion; shows profile badge during recording.
2. **Overlay** — floating panel displayed prominently.
3. **Minimal** — compact power-user status view.

All styles support optional **live transcript preview** (partial transcription visible as you speak).

**Hotkey Activation Modes:**
- Push-to-talk (hold key)
- Toggle (press once to start/stop)
- Hybrid (combines both)
- **Single modifier key** as hotkey (Command, Shift, Option, Control alone)

**Audio Cues:** Sound feedback for recording start, transcription success, and errors.

**Auto-paste:** Text injected directly into the active input field.

**Key UX Pattern:** Three visual modes for different user preferences — notch integration is the most "Apple-native" feel. Single modifier key is ergonomically convenient. Audio feedback loop (sound + visual).

---

## 9. Open-Source Whisper-Based Tools (Patterns Survey)

### whisper-writer (savbell)
- Background process, `Ctrl+Shift+Space` default.
- Small status window shows current stage (recording / transcribing). Can be hidden.
- Direct text injection into active window.

### wisper (Ubuntu — taraksh01)
- **Slim 300×60px transparent recording bar** at top of screen.
- Real-time audio waveform in the bar.
- `Shift+Space` hotkey (toggle).
- System tray: left-click = toggle recording, right-click = menu.
- Direct cursor text injection via `ydotool` (no clipboard intermediary).
- Glassmorphism visual style.

### vtt-voice2text
- Floating overlay with animated states and status text.
- **Draggable** — user can reposition the overlay.
- Animated state transitions.

### AudioWhisper (mazdak)
- Hotkey → record → auto-copy to clipboard.
- Menu bar app (tray-based).

### WhisperType (glinkot)
- **Traffic light tray icon** changes to indicate: idle / recording / transcribing.
- "Always on top" window option.

### open-whispr
- System tray icon changes color with state.
- Small overlay appears in screen corner.

---

## Cross-Product UX Pattern Analysis

### Recording Activation Patterns
| Pattern | Products | Ergonomics |
|---|---|---|
| Double-tap single key | macOS Dictation (Fn×2), Wispr Flow (Fn) | Low friction, single hand |
| Push-to-talk (hold) | Wispr Flow, Superwhisper, TypeWhisper | Precise control, hand occupied |
| Toggle (press once) | Superwhisper, TypeWhisper, whisper-writer | Hands-free after start |
| Hybrid (tap=toggle, hold=PTT) | Superwhisper, Wispr Flow | Best of both worlds |
| Always-on modal | Talon Voice | Zero-latency, no activation cost |
| Voice-activated | Dragon, Talon | Fully hands-free |

### Visual Indicator Patterns
| Style | Products | Trade-offs |
|---|---|---|
| Floating pill/bar (slim) | wisper, Windows WSR | Minimal footprint, anchored |
| Floating panel (large) | Superwhisper, MacWhisper, Dragon | Rich info, potentially intrusive |
| Near-cursor indicator | macOS Dictation, Dragon (logo at cursor) | Contextual, low distraction |
| Notch integration | TypeWhisper | Invisible to workflow |
| Menu bar / tray icon | AudioWhisper, WhisperType | Always-visible, no window clutter |
| Mode HUD | Talon Voice | Comprehensive, power-user |

### State Communication
- **Minimal (2 states):** Recording / Not recording — color change (red/green or on/off)
- **Standard (3 states):** Recording → Processing → Done — common in Whisper apps (yellow/blue/green)
- **Rich (4+ states):** Listening → Thinking → Did-you-say → Done — Windows WSR; includes disambiguation
- **Modal:** Sleep / Command / Dictation — Talon Voice; orthogonal to recording state

### Text Delivery Patterns
| Method | Products | Notes |
|---|---|---|
| Direct cursor injection | Wispr Flow, TypeWhisper, Superwhisper, wisper | Best UX, requires accessibility API |
| Clipboard + auto-paste | MacWhisper, AudioWhisper, many OSS tools | Universal but flashes clipboard |
| Streaming inline (live) | macOS Dictation | Text appears word-by-word in real-time |
| Dictation box → transfer | Dragon | Safe intermediate stage for unreliable apps |

### Hotkey Ergonomics
- **Single modifier key** (Fn, Right Cmd, Right Option alone) — favored by Wispr Flow, Superwhisper, TypeWhisper. Easiest to type; no chord needed.
- **Chord shortcuts** (Ctrl+Shift+Space) — safer for conflict avoidance, but harder to type quickly.
- **Double-tap** (Fn×2) — Apple's native approach; discoverable but slightly awkward.
- **Push-to-talk** on same key as toggle — hybrid approach from Superwhisper: tap for toggle, hold for PTT.

### Audio Feedback (Sound Cues)
Several tools complement visual indicators with audio:
- A "ping" or tone on start (Wispr Flow, macOS Dictation)
- Success sound on completion (TypeWhisper)
- Error sounds (TypeWhisper)

Sound cues reduce the need to watch the screen for state changes — important when eyes are on writing context.

---

## Key Takeaways for vox-ui Design

1. **Hybrid activation wins:** The Superwhisper pattern (tap = toggle, hold = push-to-talk, same key) is ergonomically superior. Users can choose their style without reconfiguring.

2. **Single modifier key is the gold standard:** `Fn`, `Right Cmd`, or `Right Option` alone — easiest to type mid-sentence. No chords.

3. **Overlay size is contentious:** Large overlays draw complaints ("intrusive"). The wisper 300×60px slim bar and TypeWhisper's notch mode represent the minimal end. A compact floating bar or near-cursor indicator is preferable.

4. **Three-state color coding is the norm:** Loading/yellow → Processing/blue → Done/green. This is now expected in Whisper-based tools.

5. **Direct cursor injection beats clipboard:** All best-in-class tools use accessibility APIs (xdotool, accessibility API on macOS) to inject text. Clipboard fallback is acceptable but second-class.

6. **Waveform is de rigueur:** Even minimal tools show a waveform or audio bars during recording. It confirms the mic is live — critical feedback.

7. **Sound cues complement visuals:** A start tone + success tone reduces reliance on visual attention. Low effort to implement, high UX value.

8. **Draggable overlays add power-user value:** vtt-voice2text and Wispr Flow's Android bubble let users position the indicator. For a desktop overlay, configurable position or snapping to a corner matters.

9. **macOS Dictation's near-cursor placement is elegant:** The indicator appearing near the text insertion point keeps user attention where the text will appear. Breaks from the "fixed corner" convention.

10. **Modal approaches (Talon) are powerful but require learning:** For a CLI-first tool like vox, the Talon model is instructive but overkill — stay with explicit hotkey activation.
