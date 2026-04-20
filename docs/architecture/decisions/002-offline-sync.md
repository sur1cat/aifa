# ADR-002: Offline-First Sync Strategy

## Status
Accepted

## Context

HabitFlow needs to work reliably when users don't have internet connection because:
1. Users check habits in subway, elevators, rural areas
2. Quick interactions shouldn't wait for network
3. Data loss would break user trust
4. Sync complexity affects development speed

We need to decide:
1. How to store data locally
2. How to sync with server
3. How to resolve conflicts

## Decision

### Local Storage: SwiftData

Use **SwiftData** (Apple's new persistence framework) because:
- Native Swift integration
- Automatic CloudKit sync (future)
- Works with SwiftUI seamlessly
- Modern replacement for Core Data

### Sync Strategy: Offline-First with Optimistic Updates

```
User Action → Local Write → UI Update → Background Sync → Conflict Resolution
```

1. **All writes go to local first**
   - User sees immediate feedback
   - Works without network

2. **Background sync when online**
   - Sync pending changes in queue
   - Exponential backoff on failure

3. **Pull changes on app launch**
   - Fetch changes since last sync
   - Merge with local data

### Conflict Resolution: Last-Write-Wins

For simplicity in MVP, use **Last-Write-Wins (LWW)** based on server timestamp:

| Conflict Type | Resolution |
|---------------|------------|
| Same field modified | Server timestamp wins |
| Deleted on server | Delete locally |
| Created offline | Get server ID on sync |
| Completion conflict | Union (keep both) |

### Sync Protocol

```
GET /api/v1/sync?since={timestamp}
```

Response:
```json
{
  "habits": {
    "updated": [...],
    "deleted": ["uuid1", "uuid2"]
  },
  "completions": {
    "updated": [...],
    "deleted": [...]
  },
  "sync_token": "2024-01-15T10:30:00Z"
}
```

### Rejected Alternatives

| Option | Reason for Rejection |
|--------|---------------------|
| CRDTs | Too complex for MVP, overkill |
| Server-first | Bad UX without network |
| Manual sync button | Poor UX, users forget |
| Real-time (WebSocket) | Unnecessary for habits, adds complexity |

## Consequences

### Positive
- Great offline UX
- Simple to implement for MVP
- No data loss risk
- Fast UI responses

### Negative
- LWW can lose edits (rare for habits)
- Need to handle merge carefully
- More client-side code

### Risks
- Edge case bugs in sync logic
- User confusion if data "changes" after sync

## Implementation Notes

1. **Sync queue table** in SwiftData:
   ```swift
   @Model
   class PendingSync {
       var entityType: String
       var entityId: UUID
       var operation: String // create, update, delete
       var payload: Data
       var createdAt: Date
   }
   ```

2. **Sync on events**:
   - App becomes active
   - Network becomes available
   - After local write (debounced)

3. **Retry strategy**:
   - Immediate retry on transient errors
   - Exponential backoff: 1s, 2s, 4s, 8s, max 60s
   - Give up after 10 attempts, try again next session

## Future Improvements

- **Phase 2**: Conflict UI for rare cases
- **Phase 3**: Real-time sync for shared habits
- **Phase 4**: CloudKit as alternative backend
