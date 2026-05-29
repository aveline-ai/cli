# aveline

Command-line client for the Aveline API.

## Install

From source (Go 1.22+):

```
go install github.com/aveline-ai/cli/cmd/aveline@latest
```

Or build locally:

```
make build   # produces bin/aveline
```

## Quick start

```
aveline login --api-url http://localhost:4000   # paste avl_... token
aveline whoami
aveline workspace list
aveline workspace use stable-pod

aveline save --title "Oncall rotation" --tag oncall --pin
aveline list --pinned --tag oncall
aveline get oncall-rotation
aveline edit oncall-rotation --add-tag ops --unpin
aveline delete oncall-rotation
aveline restore oncall-rotation
```

Saved views:

```
aveline view create oncall --name "Oncall" --tag oncall
aveline view list
aveline view get oncall
aveline view edit oncall --add-tag ops
aveline view delete oncall
```

## Configuration

`aveline` reads `$XDG_CONFIG_HOME/aveline/config.toml`, falling back to
`~/.config/aveline/config.toml`. It is written with mode 0600.

```toml
api_url   = "https://app.aveline.ai"
token     = "avl_..."
workspace = "stable-pod"
```

Precedence for `api_url`: `--api-url` flag > `AVELINE_API_URL` env > config
file > default `https://app.aveline.ai`.

## Output

Every command takes `--json` and emits the raw API response. This is the
mode Claude uses to interoperate with `aveline`.

## Body input

`aveline save` and `aveline edit` accept `--body -` to read markdown from
stdin or `--body FILE` to read from disk.

## Errors

Server error envelopes (`{ "error": { "code", "message", ... } }`) are
rendered verbatim to stderr; the process exits non-zero.
