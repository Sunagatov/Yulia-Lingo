# Feature: Start (`/start`)

**Status:** Stable  
**Handler:** `internal/bot/start_handler.go`

---

## Overview

Entry point for the bot. Greets the user by first + last name, sends a persistent reply keyboard, and immediately follows with an inline menu keyboard.

---

## User Stories

- As a new user, I want to be greeted and see all available actions so I can start learning immediately.
- As a returning user, I want `/start` to reset my keyboard if it was lost.

---

## Functional Requirements

1. Greet user by `FirstName [LastName]` from Telegram profile.
2. Send a persistent reply keyboard with 4 buttons: `🔺 Irregular Verbs`, `🔺 My Word List`, language label, menu label.
3. Immediately send an inline menu keyboard in a second message.
4. Clear any active FSM state (`session.ClearState()`).

---

## Non-Functional Requirements

- Two messages sent per `/start` — if the first fails, the second is not sent.
- No database writes on `/start`.

---

## Bot Flow

```
User: /start
Bot → Message 1: "Welcome, {Name}!" + reply keyboard
Bot → Message 2: menu text + inline keyboard
```

---

## State Transitions

| Before | After |
|--------|-------|
| Any | `IDLE` |

---

## Error Messages

| Condition | Behaviour |
|-----------|-----------|
| `update.Message.From == nil` | Returns error, no message sent |

---

## Acceptance Criteria

- [ ] Reply keyboard appears with correct labels in the user's language.
- [ ] Inline menu appears in the same session.
- [ ] Any prior FSM state is cleared.
- [ ] If `LastName` is empty, only `FirstName` is shown (no trailing space).

---

## Data Model

No reads or writes. Language is loaded from session (which was loaded from DB on first access).

---

## Security & Privacy

- No user data stored.
- `From == nil` guard prevents nil-pointer panic.

---

## Metrics & Observability

No dedicated metrics. Errors are returned to the update loop which logs them.

---

## Known Limitations

- If the first `Send` (reply keyboard) succeeds but the second (inline menu) fails, the user sees the keyboard but no menu — no rollback.
- Language is whatever is in session; if DB is unavailable at session creation, defaults to `ru`.

---

## Relationships to Other Features

- Sends the same inline keyboard as `/menu`.
- Reply keyboard buttons route to `irregular-verbs` and `my-word-list` handlers.

---

## Out of Scope

- Onboarding tutorial.
- Analytics on new vs returning users.

---

## Open Questions

- Should `/start` re-send the keyboard even when the user is mid-flow (e.g. waiting for import input)?

---

## Testing Notes

No unit tests for `StartHandler`. Covered implicitly by integration tests if any exist.

---

## Changelog

| Date | Change |
|------|--------|
| 2025 | Initial spec |
