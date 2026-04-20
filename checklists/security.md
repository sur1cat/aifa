# Security Checklist

Use this checklist to ensure security best practices.

## Authentication

- [ ] Passwords hashed with bcrypt (cost >= 12)
- [ ] JWT secrets are strong (256+ bits)
- [ ] JWT secrets stored in environment variables
- [ ] Access tokens expire quickly (15-60 min)
- [ ] Refresh tokens expire reasonably (7-30 days)
- [ ] Refresh tokens are rotated on use
- [ ] Failed login attempts are rate limited
- [ ] Account lockout after X failed attempts
- [ ] Session invalidation on logout

## Authorization

- [ ] Every endpoint checks authentication
- [ ] Users can only access their own data
- [ ] Role checks implemented (if applicable)
- [ ] No horizontal privilege escalation
- [ ] No vertical privilege escalation
- [ ] Resource ownership verified

## Input Validation

- [ ] All input validated (type, length, format)
- [ ] Validation on both client and server
- [ ] SQL injection prevented (parameterized queries)
- [ ] XSS prevented (output encoding)
- [ ] Path traversal prevented
- [ ] File upload restrictions (type, size)
- [ ] JSON/XML parsing limits set

## Data Protection

- [ ] HTTPS only in production
- [ ] Sensitive data encrypted at rest
- [ ] PII minimized (collect only what's needed)
- [ ] Data retention policy defined
- [ ] Soft delete for compliance
- [ ] Database backups encrypted

## API Security

- [ ] Rate limiting implemented
- [ ] CORS properly configured
- [ ] Request size limits set
- [ ] Timeout limits set
- [ ] Error messages don't leak info
- [ ] Versions deprecated securely

## Secrets Management

- [ ] No secrets in code
- [ ] No secrets in git history
- [ ] Environment variables for config
- [ ] Production secrets rotated regularly
- [ ] .env files in .gitignore
- [ ] Secrets differ per environment

## Logging & Monitoring

- [ ] Authentication events logged
- [ ] Authorization failures logged
- [ ] No sensitive data in logs (passwords, tokens)
- [ ] Logs are tamper-resistant
- [ ] Alerting for suspicious activity

## Dependencies

- [ ] Dependencies regularly updated
- [ ] Known vulnerabilities scanned
- [ ] Minimal dependencies used
- [ ] Dependencies from trusted sources

## Mobile (iOS) Specific

- [ ] Keychain used for sensitive storage
- [ ] No sensitive data in UserDefaults
- [ ] Certificate pinning (optional for MVP)
- [ ] Jailbreak detection (optional)
- [ ] Debug logs disabled in release
- [ ] No sensitive data in screenshots/backups

## Infrastructure

- [ ] Firewall configured
- [ ] Unnecessary ports closed
- [ ] SSH key authentication only
- [ ] Regular security updates
- [ ] Backup/restore tested

---

## Security Review Template

```markdown
## Security Review: [Feature/PR Name]

**Date**: YYYY-MM-DD
**Reviewer**: @username

### Scope
What was reviewed:
- Endpoints: /api/v1/...
- Files: handler/auth.go

### Findings

#### Critical
- None

#### High
- None

#### Medium
- [ ] Finding description

#### Low
- [ ] Finding description

### Recommendations
1. ...
2. ...

### Sign-off
- [ ] Security review complete
- [ ] All critical/high issues resolved
```
