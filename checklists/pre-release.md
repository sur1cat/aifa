# Pre-Release Checklist

Use this checklist before each release.

## Code Complete

- [ ] All planned features implemented
- [ ] All tasks marked as done
- [ ] No open blockers
- [ ] Feature flags set correctly

## Testing

- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] Manual testing complete
- [ ] Edge cases tested
- [ ] Offline mode tested
- [ ] Different device sizes tested
- [ ] Dark mode tested
- [ ] Accessibility tested (VoiceOver)

## Security

- [ ] Security checklist completed
- [ ] No new vulnerabilities introduced
- [ ] Secrets rotated (if needed)
- [ ] API rate limiting verified

## Performance

- [ ] Load testing complete
- [ ] No memory leaks
- [ ] App launch time acceptable (< 1s)
- [ ] API response times acceptable (< 200ms p95)
- [ ] Database queries optimized

## Documentation

- [ ] README updated
- [ ] API docs current
- [ ] CHANGELOG updated
- [ ] Release notes written
- [ ] Known issues documented

## Backend

- [ ] Database migrations tested
- [ ] Rollback procedure tested
- [ ] Health checks working
- [ ] Logging configured
- [ ] Error tracking configured (Sentry)
- [ ] Backup before deploy

## iOS App

- [ ] Version number bumped
- [ ] Build number incremented
- [ ] App icons correct
- [ ] Launch screen correct
- [ ] Info.plist reviewed
- [ ] Privacy descriptions updated
- [ ] TestFlight build tested
- [ ] Screenshots updated (if needed)

## Deployment

- [ ] Staging environment tested
- [ ] Deployment procedure documented
- [ ] Rollback procedure documented
- [ ] Team notified of deploy window
- [ ] Monitoring dashboards ready

## Post-Release

- [ ] Deployment successful
- [ ] Health checks passing
- [ ] Smoke tests passing
- [ ] Error rates normal
- [ ] Latency normal
- [ ] Team notified of success

---

## Release Checklist Template

```markdown
## Release v1.X.X

**Date**: YYYY-MM-DD
**Release Manager**: @username

### Summary
Brief description of what's in this release.

### New Features
- Feature 1
- Feature 2

### Bug Fixes
- Fix 1
- Fix 2

### Breaking Changes
- None / List them

### Pre-Release
- [ ] Code complete
- [ ] Tests passing
- [ ] Documentation updated

### Deployment
- [ ] Staging deployed and tested
- [ ] Production deploy scheduled
- [ ] Rollback plan ready

### Post-Release
- [ ] Production deployed
- [ ] Smoke tests passed
- [ ] Monitoring normal
- [ ] Announcement sent

### Notes
Any additional notes.
```
