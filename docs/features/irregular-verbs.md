# Feature: Irregular Verbs Practice

> **Status:** `Stable`  
> **Command:** `/irregular_verbs`  
> **Handler:** `irregular_verbs/handler.go`

---

## 1. Overview

Interactive quiz to practice English irregular verbs. Users are shown a verb in one form and must provide the other two forms. The system validates answers with fuzzy matching to handle typos and provides immediate feedback.

---

## 2. User Stories

- As a user, I want to practice irregular verbs so that I can improve my English grammar.
- As a user, I want to see my mistakes immediately so that I can learn from them.
- As a user, I want flexible answer validation so that minor typos don't count as wrong.

---

## 3. Functional Requirements

1. `/irregular_verbs` starts a quiz session with a random verb from the database.
2. Bot shows one form (infinitive, past, or past participle) and asks for the other two.
3. User provides answers separated by space, comma, or slash.
4. System validates with fuzzy matching (Levenshtein distance ≤ 2).
5. Correct answers show ✅ with all three forms.
6. Wrong answers show ❌ with correct forms and user's attempt.
7. After each answer, bot immediately shows next verb.
8. `/cancel` exits the quiz and returns to main menu.

---

## 4. Non-Functional Requirements

- **i18n:** All prompts and feedback in `en.json` and `ru.json`
- **Validation:** Answers must contain exactly 2 forms (split by space/comma/slash)
- **Fuzzy matching:** Levenshtein distance ≤ 2 for typo tolerance
- **Data source:** 295 irregular verbs loaded from Excel file on startup
- **Performance:** Answer validation < 100ms

---

## 5. Bot Flow

### Happy Path

```
User sends: /irregular_verbs
Bot: "🎯 Irregular Verbs Practice\n\nGive me the past and past participle of:\n**go**\n\nFormat: went, gone" [key: irregular_verbs.prompt]

User: "went, gone"
Bot: "✅ Correct!\ngo → went → gone\n\nNext verb:\n**see**" [key: irregular_verbs.correct]

User: "saw, seen"
Bot: "✅ Correct!\nsee → saw → seen\n\nNext verb:\n**take**"

User: /cancel
Bot: "❌ Quiz cancelled." [key: cancel.success]
```

### Edge Cases

| Scenario | Trigger | Bot Response |
|---|---|---|
| Wrong answer | "taked, taked" | "❌ Wrong!\n✅ take → took → taken\n❌ You: take → taked → taked" |
| Typo (distance ≤ 2) | "tok, taken" | "✅ Correct! (typo accepted)\ntake → took → taken" |
| Invalid format | "took" (only 1 form) | "⚠️ Please provide 2 forms separated by space, comma or slash" |
| Empty answer | "" | "⚠️ Please provide 2 forms" |

---

## 6. State Transitions

| From State | Event | To State |
|---|---|---|
| `IDLE` | `/irregular_verbs` | `IRREGULAR_VERBS_QUIZ` |
| `IRREGULAR_VERBS_QUIZ` | valid answer | `IRREGULAR_VERBS_QUIZ` (next verb) |
| `IRREGULAR_VERBS_QUIZ` | `/cancel` | `IDLE` |

---

## 7. Error Messages

| Scenario | Message Key | EN Text |
|---|---|---|
| Prompt | `irregular_verbs.prompt` | "🎯 Irregular Verbs Practice\n\nGive me the past and past participle of:\n**{verb}**" |
| Correct | `irregular_verbs.correct` | "✅ Correct!\n{infinitive} → {past} → {pastParticiple}" |
| Wrong | `irregular_verbs.wrong` | "❌ Wrong!\n✅ {infinitive} → {past} → {pastParticiple}\n❌ You: {infinitive} → {userPast} → {userPastParticiple}" |
| Invalid format | `irregular_verbs.invalid_format` | "⚠️ Please provide 2 forms separated by space, comma or slash" |

---

## 8. Acceptance Criteria

- [ ] Given `/irregular_verbs` is sent, then bot shows a random verb and asks for 2 forms.
- [ ] Given correct answer "went, gone" for "go", then bot shows ✅ and next verb.
- [ ] Given wrong answer "goed, goed" for "go", then bot shows ❌ with correct forms.
- [ ] Given answer with typo "tok, taken" (distance ≤ 2), then bot accepts as correct.
- [ ] Given answer with only 1 form, then bot shows invalid format error.
- [ ] Given `/cancel` during quiz, then quiz exits and returns to main menu.
- [ ] Given 295 verbs in database, then each quiz shows random selection.

---

## 9. Data Model

**In-memory (loaded from Excel on startup):**

| Field | Type | Notes |
|---|---|---|
| `infinitive` | `string` | Base form (e.g., "go") |
| `past` | `string` | Past tense (e.g., "went") |
| `pastParticiple` | `string` | Past participle (e.g., "gone") |
| `translation` | `string` | Russian translation (e.g., "идти") |

> Source: `resource/nepravilnye-glagoly-295.xlsx`

---

## 10. Security & Privacy

- **Ownership:** No user data stored — quiz is stateless except for session state.
- **Deletion:** No data to delete.
- **Exposure:** No sensitive fields.

---

## 11. Metrics & Observability

| Event | Log Level | Key Fields |
|---|---|---|
| `irregular_verbs.started` | `DEBUG` | `userId` |
| `irregular_verbs.answer.correct` | `DEBUG` | `userId`, `verb` |
| `irregular_verbs.answer.wrong` | `DEBUG` | `userId`, `verb`, `userAnswer` |
| `irregular_verbs.cancelled` | `DEBUG` | `userId` |

---

## 12. Known Limitations

- **No progress tracking** — quiz doesn't remember which verbs user has practiced.
- **No difficulty levels** — all 295 verbs have equal probability.
- **No spaced repetition** — doesn't prioritize verbs user struggles with.
- **Fuzzy matching threshold fixed** — Levenshtein distance ≤ 2 cannot be configured.
- **Excel file required** — app won't start if file is missing or corrupted.

---

## 13. Relationships to Other Features

- Independent feature — no dependencies
- Uses: `cancel` (to exit quiz)
- Uses: `language` (for i18n messages)

---

## 14. Out of Scope

- Progress tracking
- Difficulty levels
- Spaced repetition algorithm
- Custom verb lists
- Verb conjugation practice (only irregular verbs)

---

## 15. Open Questions

- [ ] Should we track user progress (correct/wrong per verb)?
- [ ] Should we add difficulty levels (common vs rare verbs)?
- [ ] Should fuzzy matching threshold be configurable?

---

## 16. Testing Notes

| Test File | Coverage |
|---|---|
| None | Feature not covered by automated tests |

**Not covered by tests:**
- Answer validation logic
- Fuzzy matching (Levenshtein distance)
- Random verb selection
- State transitions

---

## 17. Changelog

| Date | Change |
|---|---|
| 2026-03-14 | Initial spec created |
