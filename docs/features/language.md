# Feature: Language (`/lang`)

**Status:** Stable  
**Handler:** `internal/user_prefs/handler.go`  
**Repository:** `internal/user_prefs/repository.go`

---

## Overview

Lets the user switch the bot UI language between 🇷🇺 Russian and 🇬🇧 English. The choice is persisted to PostgreSQL and cached in the in-memory session. Switching language also re-registers the `/` command menu descriptions in the new language for that chat.

---

## User Stories

- As a user, I want to switch the bot language so all messages appear in my preferred language.
- As a returning user, I want my language choice remembered across restarts.

---

## Functional Requirements

1. `/lang` sends an inline keyboard with two buttons: `🇷🇺 Russian` and `🇬🇧 English`. The currently active language is marked with `✅ `.
2. On button tap (`LANG_{code}` callback):
   a. Validate the language code — return error if not `ru` or `en`.
   b. Update session immediately (`session.SetLanguage`).
   c. Persist to `user_preferences` via upsert (failure is warn-logged, not fatal).
   d. Call `SetMyCommands` scoped to the chat with descriptions in the new language.
   e. Edit the inline message to show confirmation + updated keyboard (active mark moved).
   f. Send a second new message with the updated reply keyboard in the new language.

---

## Non-Functional Requirements

- Language switch takes effect immediately for all subsequent messages in the session.
- DB failure on persist must not block the user — session is already updated.

---

## Bot Flow

```
User: /lang
Bot: "Choose language:" + [✅ 🇷🇺 Russian | 🇬🇧 English]

User taps: 🇬🇧 English
Bot (edit): "Language set!" + [🇷🇺 Russian | ✅ 🇬🇧 English]
Bot (new msg): "Language set!" + reply keyboard with EN labels
```

---

## State Transitions

No FSM state changes. Language is stored directly in session and DB.

---

## Error Messages

| Condition | Behaviour |
|-----------|-----------|
| Invalid lang code in callback | Returns error (logged), no message sent |
| DB persist fails | Warn log only — session already updated, user unaffected |
| `SetMyCommands` fails | Warn log only — not fatal |

---

## Acceptance Criteria

- [ ] Active language shown with `✅ ` prefix on the button.
- [ ] After switching, all subsequent bot messages use the new language.
- [ ] Language persists after bot restart (loaded from DB into session on first access).
- [ ] Reply keyboard button labels update to the new language.
- [ ] `/` command menu descriptions update to the new language for that chat.

---

## Data Model

Table: `user_preferences`

| Column | Type | Notes |
|--------|------|-------|
| `user_id` | `BIGINT PK` | Telegram user ID |
| `language` | `VARCHAR(10)` | `'ru'` or `'en'`, default `'ru'` |

Upsert on every language change (`ON CONFLICT DO UPDATE`).

---

## Security & Privacy

- `user_id` is the Telegram numeric ID — no name or contact data stored.
- Language code is validated against the `IsValid()` allowlist before use.

---

## Metrics & Observability

- `lang.persist_failed` — warn log when DB write fails.
- `lang.set_commands_failed` — warn log when `SetMyCommands` fails.

---

## Known Limitations

- Only two languages supported (`ru`, `en`). Adding a third requires code changes in `i18n`, `user_prefs`, and all message files.
- No auto-detection from Telegram's `LanguageCode` field on the user object.
- Default language is `ru` (hardcoded in `session.Lang()` fallback and DB column default).
- On language switch, two messages are sent (edit + new). The new message is redundant if the reply keyboard was already visible — minor UX noise.

---

## Relationships to Other Features

- Language is read by every handler via `session.Lang()`.
- `/start` also builds the reply keyboard — if language is switched after `/start`, the keyboard labels update only via the second message sent by `HandleLang`.

---

## Out of Scope

- Per-feature language overrides.
- More than two languages.

---

## Open Questions

- Should the default language be detected from `update.Message.From.LanguageCode` on first `/start`?

---

## Testing Notes

- `internal/i18n/message_source_test.go` covers message loading, not language switching logic.
- No tests for `HandleLang` callback flow.

---

## Changelog

| Date | Change |
|------|--------|
| 2025 | Initial spec |
