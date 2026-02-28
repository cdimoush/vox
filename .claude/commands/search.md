# Search Memos

Search across all processed memos in `output/` directory by keyword.

## Usage

```
/search <query>
```

## Example

```
/search beads
/search "simulation mandate"
/search design lab
```

## Instructions

Search through all processed voice memos in the `output/` directory for the given query.

### 1. Validate Input

The search query is: `$ARGUMENTS`

If no query provided, inform the user:
```
Usage: /search <query>
Example: /search beads
```

### 2. Search Output Directory

Search through all README.md and .md files in output/ subdirectories:

```bash
# Search for the query in all markdown files under output/
grep -r -i -l "$ARGUMENTS" output/*/*.md 2>/dev/null || echo "No matches found"
```

### 3. Gather Results with Context

For each matching file, extract:
- The memo directory name (which contains title and date)
- A context snippet showing the match

```bash
# Get matches with context (2 lines before/after)
grep -r -i -B 2 -A 2 "$ARGUMENTS" output/*/*.md 2>/dev/null
```

### 4. Format Results

Present results in a readable format:

```
🔍 Search Results for: "<query>"

Found X matches in Y memos:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📁 beads-integration-vision_2026-01-17_20-27-21
   File: README.md

   "...context with the QUERY highlighted..."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📁 another-memo_2026-01-15_10-30-00
   File: vision.md

   "...another matching context..."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 5. Provide Navigation

After results, suggest next steps:
- "Read full memo: Read output/<memo-dir>/README.md"
- "Search again: /search <different-query>"

## Notes

- Search is case-insensitive
- Searches all .md files (README.md, vision.md, research.md, etc.)
- Shows context around matches for relevance
- If no matches, suggests trying different terms
- Query can be multiple words (use quotes for exact phrase)

## Error Handling

- No output/ directory: Inform user no memos exist yet
- No matches: Suggest alternative search terms
- Empty query: Show usage instructions
