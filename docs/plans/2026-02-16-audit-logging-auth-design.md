# Wire Audit Logging to Authentication Handlers Design

**Issue:** #61 — Wire audit logging to authentication handlers
**Date:** 2026-02-16
**Status:** Approved

## Problem

The audit log infrastructure exists (`auth/audit.go` — types, event constants, `AuditLogger` interface, `InMemoryAuditLogger`, helper functions) but is not connected to the authentication handlers. Security-relevant events are not recorded.

## Scope

Wire logging only. Admin query API and request ID correlation deferred to separate issues.

## Design

Add `auth.AuditLogger` as a nullable parameter to each handler factory, matching the existing `rateLimiter` injection pattern. Handlers call `logger.Log()` at appropriate points.

### Events logged

| Handler | Event | Details |
|---------|-------|---------|
| NewLoginHandler | `auth.login.success` | user_id, email, IP, User-Agent |
| NewLoginHandler | `auth.login.failure` | attempted email, reason, IP, User-Agent |
| NewRefreshHandler | `auth.refresh.success` | user_id, IP, User-Agent |
| NewRefreshHandler | `auth.refresh.failure` | reason, IP, User-Agent |
| NewLogoutHandler | `auth.logout` | user_id, IP, User-Agent |
| NewInviteHandler | `user.invite.created` | inviter_id, invitee email, role |
| NewAcceptInviteHandler | `user.invite.accepted` | new user_id, email |

### IP and User-Agent

- IP: reuse `middleware.GetClientIP(r)`
- User-Agent: `r.UserAgent()`

### Error handling

Audit logging failures are non-blocking. Handlers log errors with `log.Printf` but do not fail the request.

### Tests

- Update existing handler tests to pass `nil` as audit logger
- Add one test per handler verifying correct audit event logged via `InMemoryAuditLogger`
- Test login failure logging separately from success
