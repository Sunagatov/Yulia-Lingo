# Feature: Irregular Verbs

> **Status:** `Stable`  
> **Command:** `/irregular_verbs`  
> **Handler:** `irregular_verbs/handler.go`

---

## 1. Overview

Browsable reference list of 295 English irregular verbs organized alphabetically. Users can browse verbs by first letter, see all three forms (infinitive, past, past participle) with Russian translations, and navigate through paginated results.

---

## 2. User Stories

- As a user, I want to browse irregular verbs so that I can look up verb forms.
- As a user, I want to filter by letter so that I can quickly find specific verbs.
- As a user, I want to see Russian translations so that I understand verb meanings.
- As a user, I want pagination so that I can browse large lists easily.

---

## 3. Functional Requirements

1. `/irregular_verbs` shows letter keyboard with verb counts per letter (e.g., "A (15)").
2. Currently active letter is marked with ✅.
3. Tapping a letter shows paginated list of verbs (10 per page).
4. Each verb shows: infinitive → past → past participle with Russian translation.
5. Navigation buttons: ◀ Prev (X/Y) and Next (X/Y) ▶.
6. "⬅️ Back to letters" button returns to letter selection.
7. Page footer shows: "Page X / Y · Letter".
8. Only letters with verbs are shown in keyboard.

---

## 4. Non-Functional Requirements

- **i18n:** All labels in `en.json` and `ru.json`
- **Pagination:** 10 verbs per page
- **Data source:** 295 irregular verbs loaded from Excel file on startup
- **Performance:** Page load < 200ms
- **Keyboard layout:** 5 buttons per row for letters

---

## 5. Bot Flow

### Happy Path

```
User sends: /irregular_verbs
Bot: "📖 Irregular Verbs

Pick a letter to browse verbs:

✅ A (15) | B (12) | C (10) | D (8) | E (5)
F (12) | G (15) | H (8) | I (3) | J (1)
K (3) | L (8) | M (6) | N (1) | O (2)
P (10) | Q (1) | R (8) | S (25) | T (12)
U (2) | W (10) | Y (1)" [key: choose_letter]

User taps: "G (15)"
Bot: "📖 Irregular Verbs — G

**get** → got → got/gotten
  ↳ получать

**give** → gave → given
  ↳ давать

**go** → went → gone
  ↳ идти

[...7 more verbs...]

Page 1 / 2 · G

◀ G (1/2) | G (2/2) ▶
⬅️ Back to letters"

User taps: "G (2/2) ▶"
Bot: [Shows page 2 with remaining 5 verbs]

User taps: "⬅️ Back to letters"
Bot: [Returns to letter keyboard]
```

### Edge Cases

| Scenario | Trigger | Bot Response |
|---|---|---|
| Single page | Letter with ≤10 verbs | No navigation buttons, only "Back to letters" |
| Last page | Tap Next on last page | Button disabled (not shown) |
| First page | Tap Prev on first page | Button disabled (not shown) |

---

## 6. State Transitions

This feature uses inline keyboards — no session state changes.

| From State | Event | To State |
|---|---|---|
| `IDLE` | `/irregular_verbs` | `IDLE` (letter keyboard) |
| `IDLE` | Tap letter | `IDLE` (verb list page 1) |
| `IDLE` | Tap Next/Prev | `IDLE` (different page) |
| `IDLE` | Tap Back | `IDLE` (letter keyboard) |

---

## 7. Error Messages

| Scenario | Message Key | EN Text |
|---|---|---|
| Letter keyboard | `choose_letter` | "📖 *Irregular Verbs*\n\nPick a letter to browse verbs:" |
| Verb list title | `irregular_verbs_title` | "📖 Irregular Verbs — *{letter}*" |
| Verb row | `verb_row` | "**{infinitive}** → {past} → {pastParticiple}" |
| Translation | `verb_row_translation` | "\n  _↳ {translation}_" |
| Back button | `back_to_letters` | "⬅️ Back to letters" |
| Page footer | `page_footer` | "\n_{pageInfo} · {letter}_" |

---

## 8. Acceptance Criteria

- [ ] Given `/irregular_verbs` is sent, then bot shows letter keyboard with counts.
- [ ] Given letter "G" has 15 verbs, then button shows "G (15)".
- [ ] Given user taps "G", then bot shows first 10 verbs starting with G.
- [ ] Given page has 15 verbs, then 2 pages are shown (10 + 5).
- [ ] Given user is on page 1, then "Next" button shows "G (2/2) ▶".
- [ ] Given user is on page 2, then "Prev" button shows "◀ G (1/2)".
- [ ] Given user taps "Back to letters", then letter keyboard is shown.
- [ ] Given verb has translation, then translation is shown below verb forms.
- [ ] Given 295 verbs total, then all are accessible via letter navigation.

---

## 9. Data Model

**In-memory (loaded from Excel on startup):**

| Field | Type | Notes |
|---|---|---|
| `verb` | `string` | Infinitive form (e.g., "go") |
| `past` | `string` | Past tense (e.g., "went") |
| `pastParticiple` | `string` | Past participle (e.g., "gone") |
| `original` | `string` | Russian translation (e.g., "идти") |

> Source: `resource/nepravilnye-glagoly-295.xlsx`  
> Loaded once on application startup via `repository.go`

---

## 10. Security & Privacy

- **Ownership:** No user data stored — verbs are read-only reference data.
- **Deletion:** No data to delete.
- **Exposure:** No sensitive fields.

---

## 11. Metrics & Observability

| Event | Log Level | Key Fields |
|---|---|---|
| `irregular_verbs.viewed` | `DEBUG` | `userId` |
| `irregular_verbs.letter_selected` | `DEBUG` | `userId`, `letter` |
| `irregular_verbs.page_viewed` | `DEBUG` | `userId`, `letter`, `page` |

---

## 12. Known Limitations

- **No search** — users must browse by letter.
- **No favorites** — users cannot save favorite verbs.
- **No quiz mode** — only reference browsing, no practice.
- **Fixed page size** — 10 verbs per page cannot be changed.
- **Excel file required** — app won't start if file is missing or corrupted.
- **No verb conjugation** — only shows 3 irregular forms, not full conjugation.

---

## 13. Relationships to Other Features

- Independent feature — no dependencies
- Uses: `language` (for i18n messages)
- Does NOT integrate with: `my_word_list` (verbs cannot be saved)

---

## 14. Out of Scope

- Search functionality
- Quiz/practice mode
- Saving verbs to word list
- Full verb conjugation tables
- Verb usage examples
- Audio pronunciation

---

## 15. Open Questions

- [ ] Should we add search functionality?
- [ ] Should we add quiz mode?
- [ ] Should users be able to save verbs to their word list?
- [ ] Should we show verb usage examples?

---

## 16. Testing Notes

| Test File | Coverage |
|---|---|---|
| None | Feature not covered by automated tests |

**Not covered by tests:**
- Letter keyboard generation
- Pagination logic
- Verb list formatting
- Navigation button logic

---

## 17. Changelog

| Date | Change |
|---|---|
| 2026-03-14 | Initial spec created (corrected from quiz to browsable list) |
