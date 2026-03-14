# Feature: Cancel (`/cancel`)

**Status:** Stable  
**Handler:** `internal/bot/handler_registry.go` → `handleCancel`

---

## Overview

Context-aware cancel command. If the user is in an active FSM state, it clears the state and confirms cancellation. If already idle, it tells the user there is nothing to cancel.

---

## User Stories

- As a user mid-flow (e.g. waiting for import input), I want `/cancel` to abort the operation cleanly.
- As an idle user, I want `/cancel` to give me a clear message rather than silently doing nothing.

---

## Functional Requirements

1. If `session.State() != IDLE`: send `MsgCancelled`, call `session.ClearState()`.
2. If `session.State() == IDLE`: send `MsgNothingToCancel`.
3. `/cancel` is checked first in the routing chain — it takes priority over all other handlers.

---

## Non-Functional Requirements

- No database writes.
- Must not clear pending word or pending import data from session (only state is cleared — pending data remains until overwritten).

---

## Bot Flow

```
# Active state
User: /cancel
Bot: "Operation cancelled."

# Idle
User: /cancel
Bot: "Nothing to cancel."
```

---

## State Transitions

| Before | After |
|--------|-------|
| Any non-IDLE | `IDLE` |
| `IDLE` | `IDLE` (unchanged) |

---

## Error Messages

| Condition | Message key |
|-----------|-------------|
| State was active | `MsgCancelled` |
| State was idle | `MsgNothingToCancel` |

---

## Acceptance Criteria

- [ ] `/cancel` during `WAITING_FOR_SEARCH` clears state and sends cancelled message.
- [ ] `/cancel` during `WAITING_FOR_IMPORT` clears state and sends cancelled message.
- [ ] `/cancel` when idle sends nothing-to-cancel message.
- [ ] State is `IDLE` after any `/cancel` call.

---

## Data Model

No reads or writes.

---

## Security & Privacy

No concerns.

---

## Metrics & Observability

No dedicated metrics.

---

## Known Limitations

- `session.pendingImport` and `session.pending` (pending word) are NOT cleared by `/cancel`. If the user cancels an import mid-flow and then triggers a new import, the old pending data is overwritten — not a bug, but worth noting.

---

## Relationships to Other Features

- Cancels `WAITING_FOR_SEARCH` state set by My Word List search.
- Cancels `WAITING_FOR_IMPORT` state set by the Import handler.

---

## Out of Scope

- Cancelling in-progress Telegram API calls.

---

## Open Questions

- Should `/cancel` also clear `pendingImport` to free memory?

---

## Testing Notes

No dedicated unit tests. Logic is simple enough to verify manually.

---

## Changelog

| Date | Change |
|------|--------|
| 2025 | Initial spec |
