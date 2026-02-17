# Enforce Role Hierarchy for User Invite Creation Design

**Issue:** #60 — Enforce role hierarchy for user invite creation
**Date:** 2026-02-16
**Status:** Approved

## Problem

The invite handler validates that the target role is valid but does not check whether the inviting user has permission to grant that role. A manager could create admin invites.

## Design

Add `canInviteRole(inviterRole, targetRole string) bool` to `user_handlers.go`. Check it after role validation in `NewInviteHandler`. Return 403 if the inviter's role level is below the target role level.

Viewer blocking is handled by `RequireRole("manager")` middleware at the route level — no duplication in the handler.

### Role hierarchy

```
viewer(1) < manager(2) < admin(3)
inviterLevel >= targetLevel → allowed
```

### Tests

- Table-driven `TestCanInviteRole` covering all 9 role combinations
- `TestInviteHandler_ManagerCannotInviteAdmin` — 403
- `TestInviteHandler_AdminCanInviteAdmin` — 200
- `TestInviteHandler_ManagerCanInviteViewer` — 200
- `TestInviteHandler_ManagerCanInviteManager` — 200
