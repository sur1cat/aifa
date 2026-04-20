# Code Review Checklist

Use this checklist when reviewing Pull Requests.

## General

- [ ] PR has clear title and description
- [ ] Changes match the linked task/issue
- [ ] No unrelated changes included
- [ ] Commits follow conventional commit format

## Code Quality

- [ ] Code follows conventions from [CLAUDE.md](/CLAUDE.md)
- [ ] No commented-out code
- [ ] No TODO comments without linked issues
- [ ] Functions are small and focused (< 50 lines)
- [ ] Variable/function names are descriptive
- [ ] No magic numbers (use constants)
- [ ] DRY - no duplicated code

## Go Specific

- [ ] Errors are wrapped with context: `fmt.Errorf("...: %w", err)`
- [ ] No ignored errors (except explicit `_ = ...`)
- [ ] Context passed as first parameter
- [ ] Interfaces are small (1-3 methods)
- [ ] No naked returns in long functions
- [ ] Defer used for cleanup
- [ ] Proper use of pointers vs values

## Swift Specific

- [ ] Proper use of `@State`, `@Binding`, `@Observable`
- [ ] Views are small and composable
- [ ] No force unwrapping (`!`) without justification
- [ ] Async/await used correctly
- [ ] Memory leaks avoided (weak references where needed)
- [ ] Accessibility labels present

## Security

- [ ] No hardcoded secrets or credentials
- [ ] Input validation on all endpoints
- [ ] Parameterized queries (no SQL injection)
- [ ] Auth checks on protected routes
- [ ] No sensitive data in logs
- [ ] HTTPS enforced (if applicable)

## Testing

- [ ] Tests written for new functionality
- [ ] Tests pass locally
- [ ] Edge cases covered
- [ ] Mocks used appropriately
- [ ] Test names are descriptive

## Performance

- [ ] No N+1 queries
- [ ] Appropriate indexes exist
- [ ] No memory leaks
- [ ] Large data sets handled efficiently
- [ ] Caching considered where appropriate

## API Changes

- [ ] OpenAPI spec updated
- [ ] Backwards compatible (or versioned)
- [ ] Error responses follow standard format
- [ ] Rate limiting considered

## Documentation

- [ ] README updated if needed
- [ ] API docs updated
- [ ] Inline comments for complex logic
- [ ] Breaking changes documented

## Final Checks

- [ ] I would be comfortable maintaining this code
- [ ] The code is simpler than before (or necessarily complex)
- [ ] No obvious bugs

---

## Review Response Template

```markdown
## Review Summary

**Status**: ✅ Approved / 🔄 Changes Requested / ❓ Questions

### What I Reviewed
- File1.go
- File2.swift

### Feedback

#### Must Fix
- [ ] Issue 1

#### Should Fix
- [ ] Issue 2

#### Suggestions
- Consider...

### Questions
- Why did you choose...?

### Positive Notes
- Great work on...
```
