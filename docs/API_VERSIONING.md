# API Versioning Strategy

## Current Version

**v1** - `/api/v1/*`

All API endpoints are currently served under the `/api/v1` prefix.

## Versioning Approach

Aifa API uses **URL path versioning** as the primary versioning strategy.

### Why URL Path Versioning?

1. **Explicit & Clear**: Version is immediately visible in the URL
2. **Cache-friendly**: Different versions have different URLs, making CDN/proxy caching straightforward
3. **Easy Testing**: Can easily test different versions in browser/curl
4. **Client Simplicity**: iOS app configures base URL once

### Version Format

```
https://your-api-domain.com/api/v{major}/*
```

- **Major version** only (v1, v2, v3)
- No minor/patch versions in URL - those are backward-compatible changes

## Compatibility Policy

### Backward-Compatible Changes (No version bump)

These changes can be made to any version without breaking clients:

- Adding new endpoints
- Adding new optional request fields
- Adding new response fields
- Adding new enum values (clients should ignore unknown values)
- Deprecating endpoints (with warning headers)
- Performance improvements
- Bug fixes that don't change API contract

### Breaking Changes (Require new version)

These changes require a new API version:

- Removing endpoints
- Removing request/response fields
- Changing field types
- Changing authentication mechanism
- Changing error response format
- Renaming fields
- Changing required fields to optional (or vice versa)

## Deprecation Process

When deprecating an API version or endpoint:

1. **Announce** deprecation 6 months before removal
2. **Add headers** to deprecated responses:
   ```
   Deprecation: true
   Sunset: Sat, 01 Jan 2027 00:00:00 GMT
   Link: <https://your-api-domain.com/api/v2/habits>; rel="successor-version"
   ```
3. **Log usage** of deprecated endpoints
4. **Notify** active users via email/push
5. **Remove** after sunset date

## Version Lifecycle

| Version | Status | Released | Sunset |
|---------|--------|----------|--------|
| v1 | Active | 2024-01-01 | - |

### Status Definitions

- **Active**: Current recommended version
- **Deprecated**: Still functional but clients should migrate
- **Sunset**: No longer available

## Client Implementation

### iOS App

```swift
struct APIConfig {
    static let baseURL = "https://your-api-domain.com/api/v1"
}
```

When migrating to a new version:

1. Update `baseURL` to new version
2. Update request/response models as needed
3. Test thoroughly
4. Ship app update

### Handling Multiple Versions

If needed, support multiple versions simultaneously:

```swift
enum APIVersion: String {
    case v1 = "v1"
    case v2 = "v2"

    var baseURL: String {
        "https://your-api-domain.com/api/\(rawValue)"
    }
}
```

## Server Implementation

### Router Setup (Gin)

```go
func SetupRoutes(r *gin.Engine) {
    // Version 1
    v1 := r.Group("/api/v1")
    {
        v1.POST("/auth/google", authHandler.GoogleSignIn)
        v1.GET("/habits", habitHandler.ListHabits)
        // ... more routes
    }

    // Version 2 (when needed)
    // v2 := r.Group("/api/v2")
    // {
    //     v2.POST("/auth/google", authHandlerV2.GoogleSignIn)
    // }
}
```

### Feature Flags for Gradual Rollout

For major changes, use feature flags:

```go
func (h *Handler) GetHabits(c *gin.Context) {
    if config.FeatureEnabled("new_habit_response") {
        // New response format
    } else {
        // Old response format
    }
}
```

## Error Handling Across Versions

All versions use consistent error format:

```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human-readable message"
    }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| VALIDATION_ERROR | 400 | Invalid request data |
| UNAUTHORIZED | 401 | Missing/invalid auth |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server error |

## Monitoring

Track these metrics per version:

- Request count
- Error rate
- Response latency (p50, p95, p99)
- Active clients
- Deprecation warning triggers

## Best Practices

1. **Version early**: Start with v1, even if you don't plan v2 soon
2. **Document everything**: Keep OpenAPI spec updated
3. **Test compatibility**: Run integration tests against all active versions
4. **Communicate changes**: Maintain a changelog
5. **Support old versions**: Keep deprecated versions running for reasonable time
6. **Monitor adoption**: Track client version distribution

## Related Documentation

- [OpenAPI Specification](./openapi.yaml)
- [CLAUDE.md](../../CLAUDE.md) - Project overview
