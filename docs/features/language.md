# Feature: Language

> **Status:** `Stable`  
> **Command:** `/language`  
> **Handler:** `user_prefs/handler.go`

---

## 1. Overview

Allows users to switch the bot's interface language between English and Russian. The selection is persisted in the database and applied immediately to all subsequent messages.

---

## 2. User Stories

- As a user, I want to switch the bot language so that I can use it in my preferred language.
- As a user, I want to see which language is currently active so that I know my selection.
- As a user, I want the change to apply immediately so that I don't have to restart.

---

## 3. Functional Requirements

1. `/language` shows a two-button keyboard: 🇬🇧 English and 🇷🇺 Русский.
2. The currently active language is marked with ✅.
3. Tapping a language saves it immediately to the database.
4. The confirmation message is shown in the **newly selected** language.
5. All subsequent bot messages use the newly selected language.
6. Default language is English for new users.

---

## 4. Non-Functional Requirements

- **i18n:** Button labels and messages in both `en.json` and `ru.json`
- **Persistence:** Language saved to `user_preferences` table
- **Supported languages:** English (`en`) and Russian (`ru`) only
- **Performance:** Language switch < 100ms

---

## 5. Bot Flow

### Happy Path

```
User sends: /language
Bot: "🌐 Choose your language:

✅ 🇬🇧 English
🇷🇺 Русский" [key: language.choose, current: EN]

User taps: 🇷🇺 Русский
Bot: "✅ Язык установлен: Русский 🇷🇺" [key: language.set, in RU]
[Keyboard refreshes:]
"🌐 Выберите язык:

🇬🇧 English
✅ 🇷🇺 Русский"

User taps: 🇬🇧 English
Bot: "✅ Language set to English 🇬🇧" [key: language.set, in EN]
[Keyboard refreshes:]
"🌐 Choose your language:

✅ 🇬🇧 English
🇷🇺 Русский"
```

### Edge Cases

| Scenario | Trigger | Bot Response |
|---|---|---|
| Tap same language | Tap ✅ English again | Saved again (no-op), confirmation shown |
| New user | First `/language` | Defaults to English, shows keyboard |

---

## 6. State Transitions

This feature is stateless — no session state changes occur. Uses inline keyboard callbacks.

---

## 7. Error Messages

| Scenario | Message Key | EN Text |
|---|---|---|
| Prompt | `language.choose` | "🌐 Choose your language:" |
| Language set | `language.set` | "✅ Language set to **{language}** {flag}" |

---

## 8. Acceptance Criteria

- [ ] Given `/language` is sent, then bot shows keyboard with EN and RU buttons.
- [ ] Given current language is EN, then EN button has ✅ and RU does not.
- [ ] Given current language is RU, then RU button has ✅ and EN does not.
- [ ] Given user taps RU, then language is saved as RU and confirmation is shown in Russian.
- [ ] Given user taps EN, then language is saved as EN and confirmation is shown in English.
- [ ] Given user taps already-active language, then it is saved again and confirmation shown.
- [ ] Given new user, then default language is English.

---

## 9. Data Model

**`user_preferences` table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `user_id` | `BIGINT` | ✅ | Primary key, Telegram user ID |
| `language` | `VARCHAR(2)` | ✅ | 'en' or 'ru', default 'en' |
| `created_at` | `TIMESTAMPTZ` | ✅ | When preference was created |
| `updated_at` | `TIMESTAMPTZ` | ✅ | When preference was last updated |

---

## 10. Security & Privacy

- **Ownership:** Language preference scoped to `user_id`.
- **Deletion:** Preference deleted if user deletes account (future feature).
- **Exposure:** No sensitive fields.

---

## 11. Metrics & Observability

| Event | Log Level | Key Fields |
|---|---|---|
| `language.changed` | `INFO` | `userId`, `oldLang`, `newLang` |
| `language.viewed` | `DEBUG` | `userId`, `currentLang` |

---

## 12. Known Limitations

- **Only EN and RU supported** — adding new languages requires code changes.
- **No auto-detection** — language not inferred from Telegram user locale.
- **No per-feature language** — all features use same language.

---

## 13. Relationships to Other Features

- Affects: all features (all messages use selected language)
- Independent feature — no dependencies

---

## 14. Out of Scope

- Auto-detection from Telegram locale
- Languages beyond EN and RU
- Per-feature language settings

---

## 15. Open Questions

- [ ] Should we auto-detect language from Telegram user locale on first `/start`?
- [ ] Should we add more languages (e.g., Spanish, German)?

---

## 16. Testing Notes

| Test File | Coverage |
|---|---|
| None | Feature not covered by automated tests |

**Not covered by tests:**
- Language save callback
- Keyboard refresh logic
- Default language for new users

---

## 17. Changelog

| Date | Change |
|---|---|
| 2026-03-14 | Initial spec created |
