# Yulia-Lingo — Feature Tracker

Legend: ✅ done · 🚧 stub/partial · ❌ not started

---

## Commands

| Command   | Status | Notes |
|-----------|--------|-------|
| `/start`  | ✅ | Greets user by name, shows persistent reply keyboard |
| `/help`   | ✅ | Sends help text |
| `/cancel` | ✅ | Context-aware: different message for active vs idle state |
| `/lang`   | ✅ | Inline keyboard, persists choice to DB, updates session |

---

## Irregular Verbs (`🔺 Irregular Verbs`)

| Feature | Status | Notes |
|---------|--------|-------|
| A–Z letter picker keyboard | ✅ | 26 buttons, 5 per row, active letter marked ✅ |
| Paginated verb list (10/page) | ✅ | Shows verb / past / past participle + Russian original |
| ◀ / ▶ navigation | ✅ | Page index in callback data |
| Back to letter picker | ✅ | |
| Data loaded from Excel on startup | ✅ | `nepravilnye-glagoly-295.xlsx`, drops & recreates table |

---

## My Word List (`🔺 My Word List`)

| Feature | Status | Notes |
|---------|--------|-------|
| Part-of-speech picker (6 categories) | ✅ | Shows word count per category |
| Active category marked ✅ | ✅ | |
| View words by category (paginated) | 🚧 | Handler shows "coming soon" — DB schema exists, no list view yet |
| Add word to list | 🚧 | Confirm flow exists in translate, but `AddWord` not implemented in repository |
| Remove word | ❌ | |
| Mark word as learned | ❌ | |

---

## Translation (default handler — any free-text input)

| Feature | Status | Notes |
|---------|--------|-------|
| Translate any English word | ✅ | Via MyMemory API, up to 5 results |
| Input validation (letters/hyphen/apostrophe, max 50 chars) | ✅ | |
| Retry on failure (3 attempts, backoff) | ✅ | |
| Save word to list — confirm flow | ✅ | Two-step: Save → Confirm / Cancel buttons |
| Save word to list — actual DB write | ❌ | `HandleWordConfirm` logs but does not call repository |
| SSRF protection (allowlisted hosts) | ✅ | |

---

## Language / i18n

| Feature | Status | Notes |
|---------|--------|-------|
| Russian UI | ✅ | Default language |
| English UI | ✅ | |
| Language persisted per user in DB | ✅ | `user_preferences` table |
| Language cached in session | ✅ | Loaded from DB on first access |

---

## Infrastructure

| Feature | Status | Notes |
|---------|--------|-------|
| PostgreSQL via pgxpool | ✅ | Connection pooling, health check on startup |
| Graceful shutdown (SIGTERM/SIGINT) | ✅ | Waits for in-flight updates |
| Concurrency limit (semaphore) | ✅ | `TELEGRAM_MAX_CONCURRENT_USERS`, default 100 |
| Per-update timeout (30s) | ✅ | |
| Panic recovery per update | ✅ | Never crashes polling loop |
| Structured JSON logging (slog) | ✅ | |
| `/` command menu registered on startup | ✅ | Per language via `SetMyCommands` |
| Docker / Docker Compose | ✅ | Multi-stage Dockerfile |
| Environment-based config | ✅ | Fails fast on missing required vars |

---

## Known Gaps (next work)

1. `HandleWordConfirm` — wire up `my_word_list.Repository.AddWord` (method doesn't exist yet)
2. Word list view — implement paginated list per POS category
3. Word removal from list
4. `my_word_list.Repository` missing `AddWord` and `RemoveWord` methods
