# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, report vulnerabilities by emailing: **devtools.sync.oss@gmail.com**

Include the following information in your report:

- Type of vulnerability (e.g., command injection, SQL injection, authentication bypass)
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact assessment of the vulnerability

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
- **Assessment**: We will investigate and provide an initial assessment within 7 days
- **Resolution Timeline**:
  - **Critical vulnerabilities**: Patched within 48 hours
  - **High severity**: Patched within 7 days
  - **Medium severity**: Patched within 30 days
  - **Low severity**: Addressed in next regular release

### After Reporting

1. We will confirm the vulnerability and determine its impact
2. We will develop and test a fix
3. We will release a patch and publish a security advisory
4. We will credit you in the advisory (unless you prefer to remain anonymous)

## Security Response SLA

| Severity | Response Time | Patch Timeline |
|----------|---------------|----------------|
| Critical | 24 hours      | 48 hours       |
| High     | 48 hours      | 7 days         |
| Medium   | 7 days        | 30 days        |
| Low      | 14 days       | Next release   |

## Security Best Practices for Users

### Agent Deployment

- Always use HTTPS for server communication
- Restrict config file permissions to owner-only (`chmod 600`)
- Use environment variables for sensitive configuration
- Keep agents updated to the latest version
- Review extension lists before deployment

### Server Deployment

- Deploy behind a reverse proxy with TLS termination
- Use strong, unique API keys per group
- Enable rate limiting
- Monitor audit logs for suspicious activity
- Regularly rotate API keys and JWT secrets
- Use PostgreSQL in production (not SQLite)

### Dashboard Access

- Use strong passwords
- Enable session timeout
- Review active sessions regularly
- Use HTTPS exclusively

## Security Features

This project implements the following security measures:

- **Authentication**: JWT tokens with short expiry and refresh rotation
- **Authorization**: Group-based access control
- **Input Validation**: Strict validation on all inputs, especially extension IDs
- **Transport Security**: TLS 1.2+ enforced for all connections
- **Audit Logging**: Comprehensive logging of security-relevant events
- **Rate Limiting**: Tiered rate limiting (IP, API key, global)
- **Extension Integrity**: SHA256 checksum verification
- **Blocklist**: Extension blocking by ID, publisher, or pattern

## Dependency Management

- Dependencies are monitored via Dependabot
- Security scanning runs on every PR via gosec and semgrep
- Critical CVEs in dependencies are patched within 48 hours

## Scope

The following are in scope for security reports:

- devtools-sync agent
- devtools-sync server
- devtools-sync dashboard
- Official container images
- Official documentation that could lead to insecure configurations

The following are out of scope:

- Third-party integrations not maintained by this project
- Social engineering attacks
- Physical attacks
- Denial of service attacks

## Recognition

We appreciate security researchers who help keep devtools-sync secure. With your permission, we will acknowledge your contribution in:

- Security advisories
- Release notes
- CONTRIBUTORS.md

Thank you for helping keep devtools-sync and its users safe.
