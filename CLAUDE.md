# Using the Aveline CLI

The `aveline` binary is the canonical way to read and write the team
wiki. Prefer it over hitting the API directly — it handles auth, config,
and the envelope contract for you.

## Surface

All verbs are **flat** (one `aveline --help` lists every operation).
Discover the full set with `aveline --help`; learn any verb with
`aveline <verb> --help`.

Common chains:

```sh
# Inspect
aveline whoami
aveline list-docs --tag runbook --pinned
aveline get-doc <slug>
aveline list-comments <slug>

# Write
aveline create-doc --title "..." --blocks blocks.json --tag <tag>
aveline edit-doc <slug> --blocks blocks.json --intent "why"   # full replace
aveline edit-doc <slug> --ops ops.json --intent "why"         # or surgical ops
aveline create-comment <slug> --body "..."
aveline reply-comment <slug> <parent-id> --body "..."

# Versions / history
aveline list-versions <slug>
aveline get-version <slug> <n>
```

## Envelope contract

**Success** — flat, payload at top level. Stable success keys are
documented per-verb (typically server-generated ids you can't compute):

```json
{"ok": true, "slug": "...", "doc_id": "...", "version_id": "...", "version_number": 7}
```

**Error** — single nested `error` object on stderr, exit code non-zero:

```json
{"ok": false, "error": {"code": "slug_taken", "message": "...", "details": {...}}}
```

Branch on `error.code`, not on `error.message`. Codes are stable.

## Exit codes

| Code | Meaning | Action |
| ---- | ------- | ------ |
| 0 | Success | continue |
| 2 | Validation | inspect `details`, fix the call |
| 3 | Auth (`unauthorized` / `forbidden`) | re-login, or you lack permission |
| 4 | `not_found` | the slug/id doesn't exist |
| 1 | Network / other | retry once, then surface |

## Input forms for body-like flags

`--blocks`, `--ops`, `--dispositions`, `--body` all accept:

- A path to a file (`blocks.json`)
- `-` for stdin
- Raw JSON / text directly (`--body "Looks good"`)
- `@PATH` for comment bodies (`--body @comment.md`)

Prefer files for anything multi-line — escaping JSON inside shell strings
is brittle.

## Workspace scoping

Most verbs operate inside a workspace. Pick a default once:

```sh
aveline use-workspace stable-pod
```

Override per-call with `-w <slug>` / `--workspace <slug>`. Verbs that
operate on global resources (`whoami`, `list-workspaces`,
`create-workspace`, `heartbeat`) ignore this.

## When NOT to use the CLI

- Browsing comments visually with timestamps — open the LiveView at
  `/w/<slug>/d/<doc-slug>` instead.
- One-off heartbeats — `curl /api/heartbeat` is fine.

For everything else, `aveline` is the preferred interface.
