# Security Notes

## Dependabot Alerts

### golang.org/x/crypto Alerts (Current Status: Not Applicable)

As of 2026-01-31, there are three open Dependabot alerts for `golang.org/x/crypto`:

1. **Alert #2 (HIGH severity)** - DoS via Slow or Incomplete SSH Key Exchange
   - **Patched in:** v0.35.0
   - **Why not applied:**
     - Requires Go 1.23+ (project uses Go 1.21)
     - Affects SSH servers with file transfer protocols
     - We only use golang.org/x/crypto for bcrypt password hashing, not SSH
     - Vulnerability does not apply to our usage

2. **Alert #3 (MEDIUM severity)** - SSH-related vulnerability
   - **Patched in:** v0.45.0
   - **Why not applied:**
     - Requires Go 1.24+ (project uses Go 1.21)
     - SSH-specific vulnerability
     - We don't use SSH features from golang.org/x/crypto

3. **Alert #4 (MEDIUM severity)** - SSH-related vulnerability
   - **Patched in:** v0.45.0
   - **Why not applied:**
     - Requires Go 1.24+ (project uses Go 1.21)
     - SSH-specific vulnerability
     - We don't use SSH features from golang.org/x/crypto

### Current Usage of golang.org/x/crypto

Our application uses `golang.org/x/crypto` exclusively for:
- `golang.org/x/crypto/bcrypt` - Password hashing for user authentication

We do NOT use:
- SSH client or server functionality
- File transfer protocols
- Any SSH-related features

### Future Considerations

While these specific alerts don't affect our usage, upgrading to Go 1.23+ should be considered as part of regular dependency maintenance. This would allow us to:
- Stay current with the Go ecosystem
- Access newer crypto library features
- Reduce false-positive security alerts

## Authentication System Security

See `docs/plans/2026-01-30-authentication-design.md` for comprehensive security design including:
- JWT token management
- Password security (bcrypt cost factor 12)
- Rate limiting
- Audit logging
- Role-based access control
