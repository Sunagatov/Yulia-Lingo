# Telegram Bot Best Practices — Yulia-Lingo

This project is a Go Telegram bot. Apply the following architectural best practices
derived from the Festiva project (Java/Spring) — adapted for Go and this domain.

---

## Architecture

- **Package by feature**, not by layer. Each domain concept (`irregular_verbs/`, `my_word_list/`, `translate/`) owns its models, repository, and service. Never mix domains.
- **Command handler = strategy + registry**. Define a handler interface with `Command() string` and `Handle(update)`. Register all handlers in a map at startup. No `if/else` chains in the router.
- **Stateful handlers** extend the base interface with `HandledStates() []BotState` and `HandleState(update)`. The router checks user state before matching by command text.
- **Routing priority order** (fixed, documented, tested): `/cancel` → exact command → reply-keyboard label → active state → default fallback.
- **FSM enum** for conversation states (`IDLE`, `WAITING_FOR_WORD`, `WAITING_FOR_CONFIRM`, etc.). Default is always `IDLE`. Every flow sets state on entry, calls `clearState()` on exit.
- **Per-user session object** in a `sync.Map` — one struct per user holding state + all pending data. Never use separate maps per field.

## Callbacks & UI

- **Callback routing by prefix** (`WORD_`, `CONFIRM_`, `LANG_`). Group into dispatch functions. Constants live on the handler that owns them.
- **Active state in UI** — mark current selection with `✅` prefix in button labels, computed at render time.
- **Contextual data in button labels** — show counts, pins, status directly in the button text.
- **Pagination** for any unbounded list. `PageSize` constant, `◀ ▶` nav buttons, page index in callback data.
- **Two-step confirmation** for all destructive actions. `CONFIRM_*` / `CANCEL_*` callback pair.
- **Persistent reply keyboard** on `/start`. Label → command map in the router.
- **Reusable `BuildText()`** on handlers so callbacks can regenerate the same text after filter/sort/page changes.

## i18n

- **Lang type** with a locale method. Never raw strings for language codes.
- **Constant message keys** — all keys as named constants, never inline strings.
- **JSON/properties files** per language. No silent server-locale fallback.
- **Bilingual structs/maps** — embed display labels per language directly on the type.
- **Missing key = key fallback** — message resolver returns the key itself if missing. Test this explicitly.

## Domain & Services

- **Rich domain types** — business logic (next occurrence, computed fields) belongs on the type, not in services or handlers.
- **Thin service layer** — no presentation logic (formatting, i18n). Only business rules, sorting, bulk queries.
- **Functional update helper** — find-mutate-save in one private function to avoid repetition.
- **Bulk queries** — never N+1. Load all needed data in one query before processing a list.
- **DB index** on the user partition key — every query filters by it.

## Observability

- **Structured logs**: `domain.action: key=value` format. No prose sentences.
- **Log levels**: DEBUG=per-user actions, INFO=lifecycle, WARN=bad input/recoverable, ERROR=unexpected failures.
- **Never log PII** (names, user content) at INFO or above.
- **Context propagation** in loops — attach userId to log context without passing it to every function.
- **Metrics on every update** — success + error path, processing time. Sanitise user strings before external systems.
- **Null Object pattern** for optional integrations (metrics, external APIs). No-op is the default, real impl is opt-in via config.

## Testing

- **Three layers**: unit (no DB, mock deps) → integration (real DB via Testcontainers/Docker) → end-to-end flow (command → state → callback → DB assertion).
- **Shared test helper** for i18n init — initialise message source once, reuse across all unit tests.
- **Table-driven tests** for entity boundary logic and i18n coverage (every key resolves in every language).
- **Test name as behaviour sentence**: `"birthday today → notification sent"`.
- **Test config** — separate config that disables bot startup, uses test DB, disables optional features.
- **Live smoke test script** — bash script that sends real Telegram messages and asserts on replies via `getUpdates`. Run before every release against a dedicated test bot token.

## Infrastructure

- **Multi-stage Dockerfile**: build → runtime. Minimal base image (`alpine` or `distroless`).
- **`.dockerignore`**: exclude `.env*`, test files, docs, `*.md`, IDE files, `docker-compose*.yml`, `.git/`.
- **Docker Compose profiles** — infra (DB) always available, bot behind `--profile bot`. `depends_on: condition: service_healthy`.
- **Three env files**: `.env` (local defaults, committed), `.env.prod` (production template, committed, values filled on server), `.gitignore` excludes actual secrets.
- **All config from env vars** with safe local defaults. Required values have no default and fail fast on startup.
- **Graceful startup/shutdown** — register bot and set command list on startup, close cleanly on SIGTERM.
- **`SetMyCommands`** on startup — registers the `/` command menu in Telegram UI. Failure is logged, not fatal.
- **TelegramClient as an injected dependency** — never constructed inline, always mockable.

## Error Handling

- **Top-level recover in update handler** — never let a panic crash the polling loop. Log + record metrics.
- **Per-user error isolation in scheduler** — one bad user never blocks others.
- **Narrow error types** — handle specific errors, not all errors.
- **Context-aware `/cancel`** — different message for active vs idle state. Always re-attach main menu.

## State

- **In-memory state is ephemeral** — document this trade-off explicitly. Lost on restart. Acceptable for single-instance. Migration path: Redis.
- **`lastNotifiedDate`** on user preferences — prevents duplicate scheduler notifications on the same day.
- **Language cached in session** — read from DB once on first access, written through on change.

## Code Quality

- **Intermediate struct for stream/loop computation** — compute derived values once, reuse. Avoid calling the same method multiple times per item.
- **Response factory** — centralise all message construction. One place for parse mode, chat ID wiring.
- **Keyboard builders as pure functions** — given inputs, return keyboard. No side effects.
- **Format constants** in one place — never duplicated across handlers.
- **Domain constants** (caps, intervals) on the service — not scattered across handlers.
