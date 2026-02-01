# Security Tools Evaluation for DevTools Sync

This document evaluates additional security tools that would be beneficial to add as GitHub Actions for the DevTools Sync project.

## Current Security Posture

### Existing Tools
| Tool | Purpose | Scope |
|------|---------|-------|
| Semgrep | SAST scanning | Go, JavaScript |
| golangci-lint | Go code linting | Agent, Server |
| ESLint | JavaScript linting | Dashboard |
| Dependabot | Dependency updates | Go modules, npm, GitHub Actions |
| CodeCov | Code coverage | Server, Dashboard |

### Tech Stack
- **Backend**: Go 1.22 (server, agent)
- **Frontend**: JavaScript/React 19 (dashboard)
- **Database**: PostgreSQL 16
- **Infrastructure**: Docker containers
- **Package Managers**: Go modules, npm

---

## Recommended Additional Security Tools

### 1. Go Vulnerability Scanning (govulncheck) - HIGH PRIORITY

**What it does**: Official Go tool that detects known vulnerabilities in Go dependencies by analyzing which vulnerable functions your code actually calls.

**Why it's needed**: While Dependabot updates dependencies, govulncheck provides deeper analysis by checking if your code paths actually reach vulnerable code. This reduces false positives and identifies real risks.

**Implementation**:
```yaml
# .github/workflows/govulncheck.yml
name: Go Vulnerability Check

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 6 * * 1'  # Weekly on Monday

jobs:
  govulncheck:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        directory: [agent, server]
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: '1.22'
      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - name: Run govulncheck
        working-directory: ${{ matrix.directory }}
        run: govulncheck ./...
```

**Effort**: Low
**Impact**: High

---

### 2. npm Audit - HIGH PRIORITY

**What it does**: Scans npm dependencies for known security vulnerabilities using the npm advisory database.

**Why it's needed**: Complements Dependabot by providing immediate feedback on PRs before dependencies are updated. Catches vulnerabilities that may be introduced by new dependencies.

**Implementation**:
```yaml
# Add to .github/workflows/dashboard.yml
- name: Security audit
  run: npm audit --audit-level=high
  continue-on-error: false
```

**Effort**: Very Low (single line addition)
**Impact**: High

---

### 3. Container Scanning with Trivy - MEDIUM PRIORITY

**What it does**: Scans Docker images for vulnerabilities in OS packages, application dependencies, and misconfigurations.

**Why it's needed**: The project uses Docker containers. Trivy can detect vulnerabilities in base images and installed packages that other tools miss.

**Implementation**:
```yaml
# .github/workflows/container-scan.yml
name: Container Security Scan

on:
  push:
    branches: [main]
    paths:
      - '**/Dockerfile*'
      - 'docker-compose*.yml'
  pull_request:
    paths:
      - '**/Dockerfile*'

jobs:
  trivy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - name: Build server image
        run: docker build -f server/Dockerfile.dev -t devtools-server:test ./server

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'devtools-server:test'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'

      - name: Upload Trivy scan results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'
```

**Effort**: Medium
**Impact**: Medium-High

---

### 4. CodeQL Analysis - MEDIUM PRIORITY

**What it does**: GitHub's semantic code analysis engine that performs deep data flow analysis to detect security vulnerabilities, bugs, and code quality issues.

**Why it's needed**: CodeQL can find complex vulnerabilities that pattern-based tools like Semgrep might miss, including SQL injection, XSS, path traversal, and authentication issues through data flow tracking.

**Implementation**:
```yaml
# .github/workflows/codeql.yml
name: CodeQL Analysis

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 4 * * 1'  # Weekly

jobs:
  analyze:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      contents: read
    strategy:
      matrix:
        language: [go, javascript]
    steps:
      - uses: actions/checkout@v6

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
```

**Effort**: Low
**Impact**: Medium-High

---

### 5. Secret Scanning with Gitleaks - MEDIUM PRIORITY

**What it does**: Scans git history and current code for accidentally committed secrets like API keys, passwords, and tokens.

**Why it's needed**: Prevents credential leaks. While GitHub has built-in secret scanning for public repos, Gitleaks provides more comprehensive detection and works in CI for immediate feedback.

**Implementation**:
```yaml
# .github/workflows/secrets.yml
name: Secret Scanning

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  gitleaks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Effort**: Very Low
**Impact**: High (prevents credential exposure)

---

### 6. SBOM Generation with Syft - LOW PRIORITY

**What it does**: Generates a Software Bill of Materials (SBOM) listing all dependencies, licenses, and versions in your project.

**Why it's needed**: Important for supply chain security compliance and vulnerability tracking. Becoming a requirement in many enterprise environments.

**Implementation**:
```yaml
# .github/workflows/sbom.yml
name: Generate SBOM

on:
  release:
    types: [published]
  workflow_dispatch:

jobs:
  sbom:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          format: spdx-json
          output-file: sbom.spdx.json

      - name: Upload SBOM
        uses: actions/upload-artifact@v4
        with:
          name: sbom
          path: sbom.spdx.json
```

**Effort**: Low
**Impact**: Medium (compliance/audit value)

---

### 7. Dependency License Scanning - LOW PRIORITY

**What it does**: Checks that all dependencies use licenses compatible with your project's licensing requirements.

**Why it's needed**: Prevents legal issues from incompatible open-source licenses being introduced into the codebase.

**Implementation**:
```yaml
# Add to existing workflows or create new
- name: Check Go licenses
  run: |
    go install github.com/google/go-licenses@latest
    go-licenses check ./...
  working-directory: server
```

**Effort**: Low
**Impact**: Low-Medium (legal compliance)

---

## Summary and Recommendations

### Priority Matrix

| Tool | Priority | Effort | Impact | Status |
|------|----------|--------|--------|--------|
| govulncheck | High | Low | High | ✅ **Implemented** |
| npm audit | High | Very Low | High | ✅ **Implemented** |
| Gitleaks | Medium | Very Low | High | ✅ **Implemented** |
| CodeQL | Medium | Low | Medium-High | ✅ **Implemented** |
| Trivy | Medium | Medium | Medium-High | ✅ **Implemented** |
| SBOM (Syft) | Low | Low | Medium | ✅ **Implemented** |
| License scanning | Low | Low | Low-Medium | ✅ **Implemented** |

### Immediate Actions (High Priority)
1. Add `govulncheck` workflow for Go vulnerability scanning
2. Add `npm audit` step to dashboard workflow
3. Add `gitleaks` workflow for secret scanning

### Near-term Actions (Medium Priority)
4. Add CodeQL analysis for deeper vulnerability detection
5. Add Trivy scanning when production Docker images are built

### Future Considerations
- DAST (Dynamic Application Security Testing) with OWASP ZAP when the application is deployed
- Fuzzing tests for the API endpoints
- Security-focused integration tests

---

## Current vs. Proposed Security Coverage

```
                    CURRENT                      PROPOSED
                    -------                      --------
Code Quality:       [golangci-lint, ESLint]     [same]
SAST:               [Semgrep]                   [Semgrep + CodeQL]
Dependency Updates: [Dependabot]               [same]
Vulnerability Scan: [none]                     [govulncheck, npm audit]
Container Scan:     [none]                     [Trivy]
Secret Detection:   [none]                     [Gitleaks]
SBOM:               [none]                     [Syft]
License Compliance: [none]                     [go-licenses]
```

This evaluation provides a roadmap for strengthening the security posture of the DevTools Sync project through automated CI/CD security checks.

---

## Implementation Status

**Last Updated:** 2026-02-01

All recommended security tools have been implemented as GitHub Actions workflows:

### Implemented Workflows

1. **`.github/workflows/govulncheck.yml`**
   - Scans Go code for known vulnerabilities
   - Runs on push, PR, and weekly schedule
   - Matrix strategy for agent and server

2. **`.github/workflows/dashboard.yml`** (updated)
   - Added `npm audit` step for dependency vulnerability scanning
   - Fails on high-severity vulnerabilities

3. **`.github/workflows/secrets.yml`**
   - Gitleaks secret scanning
   - Scans full git history
   - Runs on all pushes and PRs

4. **CodeQL Analysis** (GitHub Default Setup)
   - Enabled via repository settings (not custom workflow)
   - Automatic semantic code analysis for Go and JavaScript
   - Runs on pushes, PRs, and weekly schedule
   - Note: Default setup cannot coexist with custom workflows

5. **`.github/workflows/container-scan.yml`**
   - Trivy scanning for Docker images
   - Scans both server and dashboard containers
   - Results uploaded to GitHub Security tab

6. **`.github/workflows/sbom.yml`**
   - Generates SBOMs for all components
   - Runs on releases and monthly schedule
   - Attaches SBOMs to GitHub releases

7. **`.github/workflows/license-check.yml`**
   - Validates Go and npm dependency licenses
   - Blocks incompatible licenses (GPL, AGPL, LGPL)
   - Weekly checks

### Security Coverage

The project now has comprehensive automated security scanning:
- ✅ Static Application Security Testing (SAST): Semgrep, CodeQL
- ✅ Dependency Vulnerability Scanning: govulncheck, npm audit, Dependabot
- ✅ Container Security: Trivy
- ✅ Secret Detection: Gitleaks
- ✅ Supply Chain Security: SBOM generation, license compliance
- ✅ Code Quality: golangci-lint, ESLint
