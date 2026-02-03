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

## Database SSL/TLS Configuration

### Overview

The server enforces encrypted database connections in production to protect sensitive data in transit. This prevents credentials, user data, and password hashes from being exposed to network sniffing.

### SSL Modes

PostgreSQL supports several SSL modes with different security guarantees:

| Mode | Encryption | Server Auth | Hostname Check | Production Use |
|------|-----------|-------------|----------------|----------------|
| `disable` | ❌ None | ❌ No | ❌ No | ❌ **Never** |
| `allow` | ⚠️ Maybe | ❌ No | ❌ No | ❌ **Never** |
| `prefer` | ⚠️ Maybe | ❌ No | ❌ No | ❌ **Never** |
| `require` | ✅ Yes | ❌ No | ❌ No | ⚠️ **Minimum** |
| `verify-ca` | ✅ Yes | ✅ Yes | ❌ No | ✅ **Good** |
| `verify-full` | ✅ Yes | ✅ Yes | ✅ Yes | ✅ **Best** |

### Production Requirements

In production mode, the server **only allows** these SSL modes:
- `sslmode=require` - Minimum acceptable (encrypts traffic but doesn't verify server identity)
- `sslmode=verify-ca` - Recommended (verifies server certificate is signed by trusted CA)
- `sslmode=verify-full` - Best practice (verifies CA and hostname match)

The server will **refuse to start** with:
- `sslmode=disable`
- `sslmode=allow`
- `sslmode=prefer`
- Missing `sslmode` parameter

### Development Mode

Development mode (`ENVIRONMENT=development` or `ENVIRONMENT=dev`) allows any SSL mode including `disable` for local testing convenience.

### Configuration Examples

**Local Development:**
```bash
# Minimal configuration for local PostgreSQL
export ENVIRONMENT=development
export DATABASE_URL='postgres://devtools:password@localhost:5432/devtools_sync?sslmode=disable'
```

**Production (Basic SSL):**
```bash
export ENVIRONMENT=production
export DATABASE_URL='postgres://user:password@db.example.com:5432/devtools_sync?sslmode=require'
```

**Production (Verified SSL - Recommended):**
```bash
export ENVIRONMENT=production
export DATABASE_URL='postgres://user:password@db.example.com:5432/devtools_sync?sslmode=verify-full&sslrootcert=/etc/ssl/certs/ca-bundle.crt'
```

### Cloud Provider SSL Setup

**AWS RDS:**
```bash
# RDS provides SSL certificates, use verify-full for best security
export DATABASE_URL='postgres://user:password@mydb.abc123.us-east-1.rds.amazonaws.com:5432/devtools_sync?sslmode=verify-full&sslrootcert=/path/to/rds-ca-bundle.crt'

# Download RDS CA bundle
wget https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem -O rds-ca-bundle.crt
```

**Google Cloud SQL:**
```bash
# Cloud SQL with SSL/TLS
export DATABASE_URL='postgres://user:password@10.1.2.3:5432/devtools_sync?sslmode=verify-ca&sslrootcert=/path/to/server-ca.pem&sslcert=/path/to/client-cert.pem&sslkey=/path/to/client-key.pem'

# Download Cloud SQL certificates via gcloud
gcloud sql ssl-certs create client-cert --instance=INSTANCE_NAME
gcloud sql ssl-certs describe client-cert --instance=INSTANCE_NAME
```

**Azure Database for PostgreSQL:**
```bash
# Azure with SSL verification
export DATABASE_URL='postgres://user@servername:password@servername.postgres.database.azure.com:5432/devtools_sync?sslmode=verify-full&sslrootcert=/path/to/BaltimoreCyberTrustRoot.crt.pem'

# Download Azure root certificate
wget https://www.digicert.com/CACerts/BaltimoreCyberTrustRoot.crt.pem
```

**Heroku Postgres:**
```bash
# Heroku provides SSL automatically
export DATABASE_URL='postgres://user:password@ec2-1-2-3-4.compute-1.amazonaws.com:5432/dbname?sslmode=require'
```

### Docker Compose Production

Update `docker-compose.yml` for production:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    command:
      - "postgres"
      - "-c"
      - "ssl=on"
      - "-c"
      - "ssl_cert_file=/etc/ssl/certs/server.crt"
      - "-c"
      - "ssl_key_file=/etc/ssl/private/server.key"
    volumes:
      - ./certs:/etc/ssl/certs
      - ./keys:/etc/ssl/private
    environment:
      POSTGRES_USER: devtools
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: devtools_sync

  server:
    environment:
      ENVIRONMENT: production
      DATABASE_URL: postgres://devtools:${POSTGRES_PASSWORD}@postgres:5432/devtools_sync?sslmode=verify-full&sslrootcert=/etc/ssl/certs/ca.crt
    volumes:
      - ./certs/ca.crt:/etc/ssl/certs/ca.crt:ro
```

### Generating Self-Signed Certificates (Development/Testing)

For testing SSL in non-production environments:

```bash
# Generate CA private key
openssl genrsa -out ca-key.pem 4096

# Generate CA certificate
openssl req -new -x509 -days 365 -key ca-key.pem -out ca.pem \
  -subj "/CN=DevTools Sync CA"

# Generate server private key
openssl genrsa -out server-key.pem 4096

# Generate server certificate signing request
openssl req -new -key server-key.pem -out server.csr \
  -subj "/CN=postgres"

# Sign server certificate with CA
openssl x509 -req -days 365 -in server.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out server.pem

# Set permissions
chmod 600 server-key.pem ca-key.pem
chmod 644 server.pem ca.pem
```

### Validation

The server validates the database URL on startup:

- ✅ **SSL Mode Check:** Verifies `sslmode` parameter value
- ✅ **Production Enforcement:** Rejects weak SSL modes in production
- ✅ **Clear Error Messages:** Provides actionable guidance on configuration errors
- ✅ **Development Flexibility:** Allows any mode in development for local testing

### Troubleshooting

**Error: "database SSL required in production"**
```bash
# Check your DATABASE_URL
echo $DATABASE_URL

# Ensure it includes sslmode=require (minimum) or verify-full (recommended)
export DATABASE_URL='postgres://user:pass@host:5432/db?sslmode=verify-full'
```

**Error: "certificate verify failed"**
```bash
# Ensure CA certificate path is correct
export DATABASE_URL='postgres://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/correct/path/to/ca.crt'

# Verify certificate file exists and is readable
ls -la /correct/path/to/ca.crt
```

**Error: "server certificate for 'hostname' does not match"**
```bash
# Use verify-ca instead of verify-full if hostname doesn't match certificate
export DATABASE_URL='postgres://user:pass@host:5432/db?sslmode=verify-ca&sslrootcert=/path/to/ca.crt'
```

### Security Best Practices

1. **Use verify-full in production** when possible for maximum security
2. **Rotate certificates** before expiration (set alerts for 30 days before)
3. **Store certificates securely** - use secrets management (AWS Secrets Manager, HashiCorp Vault, etc.)
4. **Use strong cipher suites** - PostgreSQL 12+ defaults are secure
5. **Monitor SSL connections** - log and alert on unencrypted connection attempts
6. **Document certificate locations** in deployment runbooks
7. **Test SSL configuration** in staging before production deployment
