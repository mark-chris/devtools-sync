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

## JWT Secret Management

### Generating a Strong JWT Secret

The JWT secret must be at least 32 characters and cryptographically random. Use one of these methods:

**Option 1: OpenSSL (recommended)**
```bash
openssl rand -base64 48
```

**Option 2: /dev/urandom**
```bash
head -c 32 /dev/urandom | base64
```

**Option 3: Go**
```go
package main
import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
)
func main() {
    b := make([]byte, 32)
    rand.Read(b)
    fmt.Println(base64.StdEncoding.EncodeToString(b))
}
```

### Setting the JWT Secret

**Development:**
```bash
# Default secret is allowed in development mode
export ENVIRONMENT=development
export JWT_SECRET=local-dev-jwt-secret-not-for-production
```

**Production:**
```bash
# Generate and set a strong secret
export JWT_SECRET=$(openssl rand -base64 48)
```

**Docker/Kubernetes:**
```yaml
# Use secrets management
apiVersion: v1
kind: Secret
metadata:
  name: devtools-sync-secrets
type: Opaque
data:
  jwt-secret: <base64-encoded-secret>
```

### Secret Validation

The server validates the JWT secret on startup:

- ✅ **Minimum length:** 32 characters
- ✅ **Not a weak default:** Checks against known weak values
- ✅ **Environment-aware:** Dev mode allows defaults with warning
- ❌ **Production:** Server refuses to start with weak/default secrets

### Secret Rotation Procedure

JWT secret rotation requires a dual-key validation period to avoid downtime:

**Step 1: Prepare new secret**
```bash
# Generate new secret
NEW_SECRET=$(openssl rand -base64 48)
```

**Step 2: Add new secret to environment** (implementation pending - see #57)
```bash
# Future: Support dual-key validation
export JWT_SECRET=$OLD_SECRET
export JWT_SECRET_NEW=$NEW_SECRET
```

**Step 3: Deploy with dual validation**
- Server validates tokens signed with either old or new secret
- Wait for all old tokens to expire (15 minutes for access tokens)

**Step 4: Complete rotation**
```bash
# Switch to new secret only
export JWT_SECRET=$NEW_SECRET
unset JWT_SECRET_NEW
```

**Step 5: Revoke old tokens**
- Optionally revoke all refresh tokens to force re-authentication
- Update secret in secrets management system

### When to Rotate Secrets

Rotate JWT secrets when:
- **Suspected compromise:** Immediately
- **Employee departure:** Within 24 hours if they had access
- **Regular schedule:** Every 90 days (best practice)
- **After security incident:** As part of incident response
- **Compliance requirement:** As mandated by policy

### Emergency Rotation

In case of suspected secret compromise:

1. **Generate new secret immediately**
2. **Deploy with new secret** (downtime acceptable in emergency)
3. **Revoke all refresh tokens** from database
4. **Force all users to re-authenticate**
5. **Investigate scope of compromise**
6. **Update monitoring alerts**
