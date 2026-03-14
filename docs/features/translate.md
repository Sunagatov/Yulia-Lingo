# Feature: Translate

> **Status:** `Stable`  
> **Command:** `/translate`  
> **Handler:** `translate/handler.go`

---

## 1. Overview

Instant word translation from English to Russian with multiple meanings, part of speech labels, and one-click save to personal word list. Uses free dictionary API with fallback to translation API.

---

## 2. User Stories

- As a user, I want to translate English words so that I can understand their meaning.
- As a user, I want to see multiple meanings so that I understand different contexts.
- As a user, I want to save words to my list so that I can review them later.
- As a user, I want to see part of speech so that I understand word usage.

---

## 3. Functional Requirements

1. `/translate` prompts user to enter an English word.
2. User sends a word (e.g., "run").
3. Bot fetches meanings from dictionary API (dictionaryapi.dev).
4. Bot shows: word, phonetic, all meanings grouped by part of speech.
5. Each meaning shows: part of speech, definition, translations.
6. "💾 Save to My List" button appears below translations.
7. Tapping save adds word to user's word list with all meanings.
8. If word already exists, shows "⚠️ Word already in your list."
9. If API fails, fallback to simple translation API (lingva.ml).
10. `/cancel` exits translation mode.

---

## 4. Non-Functional Requirements

- **i18n:** All messages in `en.json` and `ru.json`
- **Validation:** 
  - Word must be 1-50 characters
  - Only letters, hyphens, apostrophes allowed
  - No numbers or special characters
- **API:** 
  - Primary: dictionaryapi.dev (free, no key)
  - Fallback: lingva.ml (free, no key)
- **Timeout:** 10 seconds per API call
- **Performance:** Translation < 2 seconds
- **SSRF Protection:** Only whitelisted domains allowed

---

## 5. Bot Flow

### Happy Path

```
User sends: /translate
Bot: "🔍 **Translate**\n\nSend me an English word to translate." [key: translate.prompt]

User: "run"
Bot: "📖 **run** /rʌn/

**verb**
• move at a speed faster than a walk
  → бежать, бегать

• (of a liquid) flow
  → течь, литься

• be in charge of; manage
  → управлять, руководить

**noun**
• an act of running
  → бег, пробежка

• a continuous period
  → период, серия

💾 Save to My List"

User taps: 💾 Save
Bot: "✅ **run** saved to your word list with 5 meanings!" [key: translate.saved]
```

### Fallback to Simple Translation

```
User: "obscure"
[Dictionary API returns no results]
Bot: "📖 **obscure**

**Translation:**
• неясный, неизвестный, малоизвестный

💾 Save to My List"
```

### Edge Cases

| Scenario | Trigger | Bot Response |
|---|---|---|
| Word already saved | Save "run" twice | "⚠️ **run** is already in your word list." |
| Invalid word | "123" or "test@" | "⚠️ Please enter a valid English word (letters only, 1-50 characters)." |
| Word not found | "asdfghjkl" | "❌ Word not found. Please check spelling." |
| API timeout | Network issue | "⚠️ Translation service unavailable. Please try again." |
| Empty input | "" | "⚠️ Please enter a word to translate." |

---

## 6. State Transitions

| From State | Event | To State |
|---|---|---|
| `IDLE` | `/translate` | `TRANSLATE_WAITING_WORD` |
| `TRANSLATE_WAITING_WORD` | valid word | `IDLE` (shows translation) |
| `TRANSLATE_WAITING_WORD` | `/cancel` | `IDLE` |
| `IDLE` | Tap 💾 Save | `IDLE` (word saved) |

---

## 7. Error Messages

| Scenario | Message Key | EN Text |
|---|---|---|
| Prompt | `translate.prompt` | "🔍 **Translate**\n\nSend me an English word to translate." |
| Invalid word | `translate.invalid_word` | "⚠️ Please enter a valid English word (letters only, 1-50 characters)." |
| Not found | `translate.not_found` | "❌ Word not found. Please check spelling." |
| Saved | `translate.saved` | "✅ **{word}** saved to your word list with {count} meanings!" |
| Already exists | `translate.already_exists` | "⚠️ **{word}** is already in your word list." |
| API error | `translate.api_error` | "⚠️ Translation service unavailable. Please try again." |

---

## 8. Acceptance Criteria

- [ ] Given `/translate` is sent, then bot prompts for a word.
- [ ] Given valid word "run", then bot shows multiple meanings with translations.
- [ ] Given word has phonetic, then phonetic is shown (e.g., /rʌn/).
- [ ] Given word has 5 meanings, then all 5 are shown grouped by part of speech.
- [ ] Given user taps 💾 Save, then word is added to word list with all meanings.
- [ ] Given word already exists, then "already in list" message is shown.
- [ ] Given invalid word "123", then validation error is shown.
- [ ] Given dictionary API fails, then fallback to translation API.
- [ ] Given API timeout, then error message is shown.
- [ ] Given `/cancel`, then translation mode exits.

---

## 9. Data Model

**Saves to `words` and `word_meanings` tables** (see my-word-list.md for schema).

**API Response (dictionaryapi.dev):**

```json
{
  "word": "run",
  "phonetic": "/rʌn/",
  "meanings": [
    {
      "partOfSpeech": "verb",
      "definitions": [
        {
          "definition": "move at a speed faster than a walk",
          "example": "she ran across the road"
        }
      ]
    }
  ]
}
```

**Fallback API (lingva.ml):**

```json
{
  "translation": "бежать, бегать"
}
```

---

## 10. Security & Privacy

- **Ownership:** Saved words scoped to `user_id`.
- **SSRF Protection:** Only whitelisted domains (dictionaryapi.dev, lingva.ml) allowed.
- **Input Validation:** Strict regex prevents injection attacks.
- **No API Keys:** Both APIs are free and don't require authentication.
- **Exposure:** User input (words) logged at DEBUG level only.

---

## 11. Metrics & Observability

| Event | Log Level | Key Fields |
|---|---|---|
| `translate.requested` | `DEBUG` | `userId`, `word` |
| `translate.api.success` | `DEBUG` | `word`, `meaningsCount` |
| `translate.api.fallback` | `INFO` | `word` |
| `translate.api.failed` | `ERROR` | `word`, `error` |
| `translate.saved` | `INFO` | `userId`, `word`, `meaningsCount` |
| `translate.already_exists` | `DEBUG` | `userId`, `word` |
| `translate.invalid_word` | `DEBUG` | `userId`, `input` |

---

## 12. Known Limitations

- **English to Russian only** — no other language pairs supported.
- **Single word only** — cannot translate phrases or sentences.
- **No context** — translations don't consider sentence context.
- **API dependency** — relies on free third-party APIs (no SLA).
- **No offline mode** — requires internet connection.
- **No pronunciation audio** — only phonetic text shown.
- **No word history** — doesn't track translation history.

---

## 13. Relationships to Other Features

- Affects: `my-word-list` (adds words)
- Uses: `language` (for i18n messages)
- Independent of: `irregular_verbs`

---

## 14. Out of Scope

- Multi-language support (beyond EN→RU)
- Phrase/sentence translation
- Context-aware translation
- Pronunciation audio
- Translation history
- Offline mode
- Custom dictionaries

---

## 15. Open Questions

- [ ] Should we add more language pairs?
- [ ] Should we cache translations to reduce API calls?
- [ ] Should we add pronunciation audio?
- [ ] Should we track translation history?
- [ ] Should we support phrase translation?

---

## 16. Testing Notes

| Test File | Coverage |
|---|---|
| `translate/validate_test.go` | Input validation (regex, length) |
| `translate/dict_client_test.go` | Dictionary API client |

**Not covered by tests:**
- Fallback logic (dictionary → translation API)
- Save to word list flow
- Duplicate detection
- API timeout handling
- SSRF protection

---

## 17. Changelog

| Date | Change |
|---|---|
| 2026-03-14 | Initial spec created |
