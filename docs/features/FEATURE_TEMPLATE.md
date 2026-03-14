# Feature: [Feature Name]

> **Status:** `Draft` | `Review` | `Stable`  
> **Command(s):** `/command`  
> **Handler(s):** `handler.go`

---

## 1. Overview

One paragraph. What this feature does and why it exists.

---

## 2. User Stories

- As a user, I want to ... so that ...

---

## 3. Functional Requirements

1. ...
2. ...

> Each requirement must be testable and unambiguous.

---

## 4. Non-Functional Requirements

- **i18n:** All messages must exist in both `en.json` and `ru.json`
- **Validation:** List all input constraints
- **Limits:** Any caps
- **Performance:** Any latency expectations

---

## 5. Bot Flow

### Happy Path

```
User sends: /command
Bot replies: [message key: xxx]
```

### Edge Cases

| Scenario | Trigger | Bot Response |
|---|---|---|
| ... | ... | ... |

---

## 6. State Transitions

| From State | Event | To State |
|---|---|---|
| `IDLE` | `/command` | `WAITING_FOR_XXX` |

---

## 7. Error Messages

| Scenario | Message Key | EN Text |
|---|---|---|
| ... | `key_name` | "..." |

---

## 8. Acceptance Criteria

- [ ] Given ... when ... then ...

---

## 9. Data Model

| Field | Type | Required | Notes |
|---|---|---|---|
| `field` | `string` | ✅ | ... |

---

## 10. Security & Privacy

- **Ownership:** Data scoped per `user_id`
- **Deletion:** What happens on account deletion
- **Exposure:** Fields that should never be logged

---

## 11. Metrics & Observability

| Event | Log Level | Key Fields |
|---|---|---|
| `event.name` | `INFO` | `userId` |

---

## 12. Known Limitations

- Current technical constraints

---

## 13. Relationships to Other Features

- Depends on: [feature]
- Affects: [feature]

---

## 14. Out of Scope

- What this feature does NOT do

---

## 15. Open Questions

- [ ] Unresolved decisions

---

## 16. Testing Notes

| Test File | Coverage |
|---|---|
| `test.go` | ... |

---

## 17. Changelog

| Date | Change |
|---|---|
| 2026-03-14 | Initial spec |
