# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build Go binary (requires CGO_ENABLED=1)
make build-web      # Build React frontend (cd web && npm ci && npm run build)
make build-all      # Build frontend + backend
make docker         # Build Docker image
make test           # Run all tests: go test ./...
make lint           # Run golangci-lint
go test ./internal/codec/...   # Run tests for a single package
go test -run TestDetect ./internal/codec/...  # Run a single test
```

CGO is required — the WebP, HEIC, AVIF, and JXL codecs use C libraries. On Linux, install `libwebp-dev` and `libjxl-dev` before building.

## Architecture

Pixshift is a Go monolith that serves as CLI, HTTP server, and MCP server from a single binary. The entry point is `cmd/pixshift/main.go` which dispatches to mode-specific files. Arguments are parsed manually in `args.go` (no CLI framework).

### Core conversion flow

```
File → Detect (magic bytes) → Decode → Transform → Encode → Inject Metadata → Output
```

This is implemented in `internal/pipeline/pipeline.go:Execute()`. The `Job` struct (`pipeline/job.go`) carries all parameters for a single conversion. Animated GIFs take a multi-frame path: `DecodeAll` → per-frame `transformImage` → `EncodeAll`.

### Dual-mode server (`internal/server/`)

The server runs in two modes depending on environment:

- **Simple mode** (no `DATABASE_URL`): Bearer token auth, basic rate limiting, `/convert` + `/formats` + `/health` only. Backward compatible with v0.5.0.
- **Full mode** (`DATABASE_URL` set): Postgres-backed auth, Stripe billing, embedded React SPA, hosted API with API key auth, Google OAuth.

Route layout in full mode:

| Prefix | Handler file | Auth |
|--------|-------------|------|
| `/api/v1/*` | `api_handler.go` | API key (`X-API-Key` header) |
| `/internal/auth/*` | `auth_handler.go` | None (public) |
| `/internal/billing/*` | `billing_handler.go` | Session cookie |
| `/internal/*` | `web_handler.go` | Session cookie |
| `/api/webhooks/stripe` | `billing_handler.go` | Stripe signature |
| `/*` | `spa.go` | None (static files) |

Key server files: `server.go` (struct, dual-mode routing), `middleware.go` (tier-aware rate limiting), `ratelimit.go` (sliding window with `AllowN(key, limit)`).

### Codec registry (`internal/codec/`)

All image formats implement `Decoder` and `Encoder` interfaces defined in `codec.go`. Optional interfaces: `AdvancedEncoder` (format-specific encoding options), `MultiFrameDecoder`/`MultiFrameEncoder` (animated images). Codecs register themselves in `DefaultRegistry()` in `registry.go`. Format detection uses magic bytes in `detect.go`, not file extensions.

**To add a new format:** create `internal/codec/yourformat.go` implementing both interfaces, add the format constant to `codec.go`, register it in `DefaultRegistry()`, add magic bytes to `detect.go`, and update `ParseFormat()`/`DefaultExtension()`/`IsSupportedExtension()`.

### Worker pool (`internal/pipeline/pool.go`)

Parallel processing via a channel-based worker pool. Configurable via `-j` flag, defaults to CPU count.

### Key packages

| Package | Purpose |
|---------|---------|
| `pipeline` | Core conversion: Job/Result structs, Execute flow, worker pool |
| `codec` | Format interfaces, registry, magic-byte detection, per-format codecs |
| `transform` | Crop (pixel/aspect/gravity), EXIF auto-rotate, text watermark |
| `resize` | CatmullRom interpolation resizing |
| `metadata` | EXIF extraction and injection |
| `rules` | YAML config engine (first-match-wins rule evaluation) |
| `preset` | Built-in presets: web, thumbnail, print, archive |
| `server` | Dual-mode HTTP server (simple + full), multi-handler architecture |
| `db` | Postgres data layer: users, sessions, API keys, conversions, daily usage |
| `auth` | Password hashing (bcrypt), API key gen/validation, session tokens, Google OAuth, HTTP middleware |
| `billing` | Stripe integration: checkout, portal, webhooks, tier definitions |
| `mcp` | MCP server (stdio) with convert, formats, analyze, compare tools |
| `watch` | fsnotify-based watch mode with configurable debounce, ignore patterns, retry |
| `dedup` | Duplicate detection via perceptual hashing (dHash + Hamming) |
| `ssim` | Structural similarity image comparison |
| `contact` | Contact sheet (JPEG thumbnail grid) generation |

### CLI structure (`cmd/pixshift/`)

| File | Contents |
|------|----------|
| `main.go` | Entry point, signal handling, preset application, mode dispatch |
| `args.go` | `options` struct, `parseArgs()`, `printUsage()` |
| `batch.go` | `runBatchMode()` — parallel batch conversion with progress bar |
| `rules_mode.go` | `runRulesMode()`, `runRulesWatch()` |
| `watch_mode.go` | `runWatchMode()` |
| `serve_mode.go` | `runServeMode()` — wires DB, auth, billing, OAuth, SPA when `DATABASE_URL` is set |
| `mcp_mode.go` | `runMCPMode()` — starts MCP server on stdio |
| `analysis.go` | `runTreeMode()`, `runSSIMMode()`, `runDedupMode()`, `runContactSheetMode()` |
| `stdin.go` | `runStdinMode()` |
| `helpers.go` | `buildJob()`, `collectFiles()`, `buildOutputPath()`, formatting helpers |

### Web frontend (`web/`)

React 19 + Vite 6 + TypeScript + Tailwind CSS 4. Builds to `web/dist/` which is embedded in the Go binary via `web/embed.go` (`//go:embed all:dist`). Build with `make build-web`. The SPA is served by `spa.go` with index.html fallback for client-side routing.

### Database (`migrations/`)

Postgres schema in `migrations/001_initial.sql`. Five tables: `users`, `sessions`, `api_keys`, `conversions`, `daily_usage`. Migrations run automatically on server startup via `db.Migrate()`.

### API key format

`pxs_` prefix + 32 random hex chars. Full key shown once at creation. Only SHA-256 hash stored in DB. Validated via `auth.HashAPIKey()` + DB lookup.

## Build-time version injection

Version, commit, and date are injected via ldflags — see `Makefile` and `internal/version/`. Don't hardcode version strings.
