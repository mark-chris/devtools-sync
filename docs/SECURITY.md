# Security Notes

## Dependabot Alerts

### golang.org/x/crypto Alerts (Current Status: Dismissed)

As of 2026-01-31, three Dependabot alerts for `golang.org/x/crypto` have been dismissed as not applicable:

1. **Alert #2 (HIGH severity)** - DoS via Slow or Incomplete SSH Key Exchange
   - **Patched in:** v0.35.0
   - **Status:** Dismissed (tolerable risk)
   - **Rationale:**
     - Affects SSH servers with file transfer protocols
     - We only use golang.org/x/crypto for bcrypt password hashing, not SSH
     - Vulnerability does not apply to our usage
   - **Note:** While the project now uses Go 1.24, which supports this patch version, the vulnerability remains inapplicable to our use case

2. **Alert #3 (MEDIUM severity)** - SSH-related vulnerability
   - **Patched in:** v0.45.0
   - **Status:** Dismissed (tolerable risk)
   - **Rationale:**
     - SSH-specific vulnerability (GSSAPI authentication)
     - We don't use SSH features from golang.org/x/crypto

3. **Alert #4 (MEDIUM severity)** - SSH-related vulnerability
   - **Patched in:** v0.45.0
   - **Status:** Dismissed (tolerable risk)
   - **Rationale:**
     - SSH Agent server vulnerability
     - We don't use SSH features from golang.org/x/crypto

### Current Usage of golang.org/x/crypto

Our application uses `golang.org/x/crypto` exclusively for:
- `golang.org/x/crypto/bcrypt` - Password hashing for user authentication

We do NOT use:
- SSH client or server functionality
- File transfer protocols
- Any SSH-related features

### Go Version

The project uses **Go 1.24** (upgraded 2026-01-31), which is the latest stable release and will be supported until Go 1.26 is released (expected February 2026).

While upgrading to newer `golang.org/x/crypto` versions is now technically possible, it remains unnecessary since:
- The vulnerabilities are SSH-specific and don't affect bcrypt usage
- The current version (v0.31.0) is secure for our use case
- Upgrading would only eliminate false-positive security alerts

## Authentication System Security

See `docs/plans/2026-01-30-authentication-design.md` for comprehensive security design including:
- JWT token management
- Password security (bcrypt cost factor 12)
- Rate limiting
- Audit logging
- Role-based access control
