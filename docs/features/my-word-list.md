# Feature: My Word List

> **Status:** `Stable`  
> **Command:** `/my_word_list`  
> **Handler:** `my_word_list/handler.go`

---

## 1. Overview

Personal vocabulary manager where users can save, browse, filter, and practice English words with translations. Each word has a confidence level (1-5 stars) that users can adjust based on their mastery. Supports pagination, filtering by letter/part of speech/confidence, and detailed word views with multiple meanings.

---

## 2. User Stories

- As a user, I want to save new words so that I can build my personal vocabulary.
- As a user, I want to browse my word list so that I can review what I've learned.
- As a user, I want to filter words by letter/confidence so that I can focus on specific areas.
- As a user, I want to rate my confidence so that I can track my progress.
- As a user, I want to see detailed meanings so that I understand word usage.
- As a user, I want to delete words so that I can keep my list relevant.

---

## 3. Functional Requirements

1. `/my_word_list` shows paginated list of user's words (8 per page).
2. Each word is an **inline keyboard button** showing: word, preposition (if any), translation preview.
3. Pagination buttons: ◀ Prev, Page X/Y (center), Next ▶.
4. Filter/sort buttons: 🔍 Search, Sort ↕, ⚙️ Filters.
5. Tapping a word button opens detail view with all meanings and translations.
6. Detail view shows: word, preposition, confidence stars, meanings grouped by part of speech.
7. Detail view buttons: ☆1-5 (rate confidence), 🗑 Delete, ⬅️ Back.
8. Confidence rating: tap any star button (1-5) to set confidence level.
9. Delete requires confirmation: "🗑 Delete" button, then confirm with "🗑 Delete" or "❌ Cancel".
10. Empty list shows: "_No words saved yet._ Send any English word to translate and tap 💾 to save it."
11. Header shows: "Page X/Y · Z words" with active filters displayed.

---

## 4. Non-Functional Requirements

- **i18n:** All messages in `en.json` and `ru.json`
- **Pagination:** 8 words per page
- **Database:** PostgreSQL with `words` and `word_meanings` tables
- **Uniqueness:** `UNIQUE (user_id, word, preposition)` constraint prevents duplicates
- **Performance:** List query < 200ms, detail query < 100ms
- **Limits:** No hard limit on word count per user

---

## 5. Bot Flow

### Happy Path - Browse List

```
User sends: /my_word_list
Bot: "📝 My Word List
Page 1/4 · 25 words

[Button: go after — присматривать]
[Button: take — брать]
[Button: run — бежать]
[Button: see — видеть]
[Button: get — получать]
[Button: give — давать]
[Button: make — делать]
[Button: know — знать]

◀ Prev | Page 1/4 | Next ▶
🔍 Search | ★ Weakest first ↕
⚙️ Filters"

User taps: [go after] button
Bot: "*go* `after`
★★★☆☆

_(phrasal verb)_ `after`
• присматривать
• ухаживать

☆1 | ☆2 | ☆3 | ☆4 | ☆5
🗑 Delete | ⬅️ Back to word list"

User taps: ☆5 button
Bot: [Updates stars to ★★★★★, shows "★ Confidence updated"]

User taps: ⬅️ Back
Bot: [Returns to word list page 1]
```

### Happy Path - Filter by Letter

```
User taps: 🔤 By Letter
Bot: "🔤 Choose first letter:

A | B | C | D | E | F | G | H | I
J | K | L | M | N | O | P | Q | R
S | T | U | V | W | X | Y | Z"

User taps: "G"
Bot: "📚 My Word List (3 words starting with G)
Page 1/1

1. **go** (verb) ⭐⭐⭐
2. **get** (verb) ⭐⭐
3. **give** (verb) ⭐⭐⭐⭐

🔄 Reset Filters"
```

### Happy Path - Delete Word

```
User taps: 🗑️ Delete (from detail view)
Bot: "⚠️ Delete **go**?

This action cannot be undone.

✅ Yes, delete | ❌ No, keep it"

User taps: ✅ Yes
Bot: "🗑️ **go** has been deleted from your word list."
[Returns to main list]
```

### Edge Cases

| Scenario | Trigger | Bot Response |
|---|---|---|
| Empty list | `/my_word_list` with 0 words | "📝 Your word list is empty. Use /translate to add words." |
| Last page | Tap ➡️ on last page | Button disabled (grayed out) |
| First page | Tap ⬅️ on first page | Button disabled (grayed out) |
| Filter returns 0 results | Filter by letter "X" | "🔍 No words found starting with X" |
| Delete cancelled | Tap ❌ No | Returns to detail view, word not deleted |

---

## 6. State Transitions

| From State | Event | To State |
|---|---|---|
| `IDLE` | `/my_word_list` | `IDLE` (inline keyboard) |
| `IDLE` | Tap word | `IDLE` (detail view) |
| `IDLE` | Tap ⭐ Rate | `IDLE` (rating keyboard) |
| `IDLE` | Tap 🗑️ Delete | `IDLE` (confirmation) |
| `IDLE` | Tap 🔤 Filter | `IDLE` (letter keyboard) |

> This feature uses inline keyboards — no session state changes.

---

## 7. Error Messages

| Scenario | Message Key | EN Text |
|---|---|---|
| Empty list | `word_list.empty` | "📝 Your word list is empty. Use /translate to add words." |
| List header | `word_list.header` | "📚 My Word List ({count} words)\nPage {page}/{totalPages}" |
| Detail view | `word_list.detail` | "📖 **{word}** ({partOfSpeech})\n\n**Meanings:**\n{meanings}\n\n⭐ Current confidence: {stars}" |
| Confidence updated | `word_list.confidence_updated` | "✅ Confidence updated to {stars} for **{word}**" |
| Delete confirmation | `word_list.delete_confirm` | "⚠️ Delete **{word}**?\n\nThis action cannot be undone." |
| Deleted | `word_list.deleted` | "🗑️ **{word}** has been deleted from your word list." |
| No results | `word_list.no_results` | "🔍 No words found {filter}" |

---

## 8. Acceptance Criteria

- [ ] Given user has 25 words, then list shows 8 words per page with 4 total pages.
- [ ] Given user is on page 1, then ⬅️ Previous button is disabled.
- [ ] Given user is on last page, then ➡️ Next button is disabled.
- [ ] Given user taps a word, then detail view shows all meanings and translations.
- [ ] Given user rates confidence as ⭐⭐⭐⭐, then word shows 4 stars in list.
- [ ] Given user deletes a word, then word is removed from database and list.
- [ ] Given user filters by letter "G", then only words starting with G are shown.
- [ ] Given user has 0 words, then empty state message is shown.
- [ ] Given word has multiple meanings, then all are shown in detail view.
- [ ] Given user taps 🔄 Reset, then all filters are cleared.

---

## 9. Data Model

**`words` table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | `SERIAL` | ✅ | Primary key |
| `user_id` | `BIGINT` | ✅ | Telegram user ID |
| `word` | `VARCHAR(255)` | ✅ | English word |
| `part_of_speech` | `VARCHAR(50)` | ✅ | noun, verb, adjective, etc. |
| `preposition` | `VARCHAR(100)` | ✅ | For phrasal verbs (e.g., "after" in "look after") |
| `translation` | `VARCHAR(255)` | ✅ | Primary translation |
| `confidence` | `SMALLINT` | ✅ | 1-5 stars, default 1 |
| `created_at` | `TIMESTAMPTZ` | ✅ | When word was added |

**`word_meanings` table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | `SERIAL` | ✅ | Primary key |
| `word_id` | `INT` | ✅ | Foreign key to `words.id` |
| `part_of_speech` | `VARCHAR(50)` | ✅ | Part of speech for this meaning |
| `translation` | `VARCHAR(255)` | ✅ | Translation for this meaning |

**Constraints:**
- `UNIQUE (user_id, word, preposition)` — prevents duplicate words
- `ON DELETE CASCADE` — deleting word deletes all meanings

---

## 10. Security & Privacy

- **Ownership:** All queries filtered by `user_id` — users can only see their own words.
- **Deletion:** Words and meanings deleted on user request (🗑️ button).
- **Exposure:** No sensitive fields. Translations are user-provided.
- **SQL Injection:** Parameterized queries prevent injection.

---

## 11. Metrics & Observability

| Event | Log Level | Key Fields |
|---|---|---|
| `word_list.viewed` | `DEBUG` | `userId`, `page`, `totalWords` |
| `word_list.detail_viewed` | `DEBUG` | `userId`, `word` |
| `word_list.confidence_updated` | `INFO` | `userId`, `word`, `oldConfidence`, `newConfidence` |
| `word_list.deleted` | `INFO` | `userId`, `word` |
| `word_list.filtered` | `DEBUG` | `userId`, `filterType`, `filterValue` |

---

## 12. Known Limitations

- **No search** — users must browse or filter by letter/confidence/part of speech.
- **No sorting options** — default sort is by confidence ASC, then alphabetically.
- **No bulk delete** — users must delete words one by one.
- **No export** — users cannot export their word list to CSV/JSON.
- **No import** — users cannot bulk import words (except via `/translate`).
- **Fixed page size** — 8 words per page cannot be changed.
- **No word notes** — users cannot add personal notes to words.

---

## 13. Relationships to Other Features

- Depends on: `translate` (adds words to list)
- Uses: `language` (for i18n messages)
- Independent of: `irregular_verbs`

---

## 14. Out of Scope

- Search functionality
- Custom sorting
- Bulk operations (delete, export, import)
- Word notes/examples
- Spaced repetition quiz
- Word statistics (most/least confident)

---

## 15. Open Questions

- [ ] Should we add search functionality?
- [ ] Should we add bulk delete?
- [ ] Should we add export to CSV?
- [ ] Should page size be configurable?
- [ ] Should we add word usage examples?

---

## 16. Testing Notes

| Test File | Coverage |
|---|---|
| `my_word_list/entity_test.go` | Entity model tests |
| `my_word_list/detail_test.go` | Detail view formatting |

**Not covered by tests:**
- Pagination logic
- Filter logic
- Confidence rating
- Delete confirmation flow
- Empty state handling

---

## 17. Changelog

| Date | Change |
|---|---|
| 2026-03-14 | Initial spec created |
| 2026-03-14 | Added 1,806 words for user 591545118 |
