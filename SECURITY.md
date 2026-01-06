# Security Policy

## Overview

scmd takes security seriously. This document outlines the security measures in place and how to report vulnerabilities.

## Security Features

### Input Validation

#### Slash Command Names (P0 - Critical)
**Protection Against:** Path traversal, command injection

All slash command names and aliases are validated to prevent malicious input:

- **Allowed Characters:** Alphanumeric, dash (`-`), underscore (`_`)
- **Length Limit:** 1-50 characters
- **Rejected Patterns:**
  - Path separators: `/`, `\`, `..`
  - Shell metacharacters: `;`, `|`, `&`, `$`, `` ` ``, `(`, `)`, `{`, `}`, `<`, `>`, `\n`, `\x00`
  - Special characters: `@`, `.`, spaces, etc.

**Example:**
```go
// Valid command names
"explain"           ✅
"my-command"        ✅
"cmd_123"           ✅

// Invalid (rejected)
"../../../passwd"   ❌ Path traversal
"test;rm -rf /"     ❌ Command injection
"test$(whoami)"     ❌ Command substitution
```

**Implementation:** `internal/validation/validators.go` - `ValidateCommandName()`

**Validation Points:**
- `slash.Runner.Add()` - When adding new commands
- `slash.Runner.LoadConfig()` - When loading from YAML
- `slash.Runner.AddAlias()` - When adding aliases

#### Repository URLs (P1 - High)
**Protection Against:** SSRF, local file access, metadata endpoint access

All repository URLs are validated to prevent server-side request forgery and unauthorized access:

- **Allowed Schemes:** `http`, `https` only
- **Rejected Schemes:** `file://`, `javascript:`, `data:`, `ftp://`, etc.
- **SSRF Protection:**
  - Localhost: `localhost`, `127.0.0.1`, `::1`, `0.0.0.0`, `*.localhost`
  - Private IPs: `10.x.x.x`, `172.16-31.x.x`, `192.168.x.x`
  - Link-local: `169.254.x.x` (including AWS metadata `169.254.169.254`)
  - IPv6 private ranges

**Example:**
```go
// Valid repository URLs
"https://github.com/user/repo"           ✅
"http://example.com/commands"            ✅
"https://api.example.com:8080/v1"        ✅

// Invalid (rejected)
"file:///etc/passwd"                     ❌ File scheme
"http://localhost:8080"                  ❌ Localhost
"http://169.254.169.254/meta-data"       ❌ AWS metadata
"http://10.0.0.1/internal"               ❌ Private IP
"http://192.168.1.1/admin"               ❌ Private IP
```

**Implementation:** `internal/validation/validators.go` - `ValidateRepoURL()`

**Validation Points:**
- `repos.Manager.Add()` - When adding new repositories
- `repos.Manager.Load()` - When loading from disk

### Command Injection Protection

**Built-in Protections:**
- Mock backend for testing (no actual execution)
- Input sanitization in prompts
- Shell command isolation
- No direct shell execution from user input

**Status:** ✅ **PROTECTED** - All test vectors pass

### Path Traversal Protection

**File Operations:**
- Output files are sandboxed to working directory
- Config files restricted to `~/.scmd/`
- No absolute path access without explicit validation

**Status:** ✅ **PROTECTED** - All test vectors pass

### Additional Security Measures

#### File Permissions
- Output files: `0644` (not world-writable)
- Config files: `0644` (not world-writable)
- Directories: `0755`

#### Resource Limits
- Input size limits (tested up to 100MB)
- Rate limiting (tested 100 commands in 30s)
- Timeout enforcement

#### YAML Safety
- Protection against YAML bombs
- Malicious config rejection
- Safe unmarshaling

## Security Test Coverage

Comprehensive security test suite in `tests/security/`:

| Attack Vector | Status | Test Count |
|---------------|--------|------------|
| Command Injection (Prompt) | ✅ Protected | 6 vectors |
| Command Injection (Stdin) | ✅ Protected | 4 vectors |
| Path Traversal (Output) | ✅ Protected | 4 vectors |
| Path Traversal (Config) | ✅ Protected | 1 test |
| Slash Command Names | ✅ Protected | 4 vectors |
| Repository URLs | ✅ Protected | 5 vectors |
| SSRF | ✅ Protected | 6 endpoints |
| Environment Injection | ✅ Protected | 1 test |
| Oversized Input | ✅ Protected | 100MB |
| Null Bytes | ✅ Protected | 1 test |
| Control Characters | ✅ Protected | 1 test |
| YAML Bombs | ✅ Protected | 1 test |
| **TOTAL** | **21/21 PASS** | **35+ vectors** |

**Test Suite:** Run with `go test ./tests/security -v`

## Reporting Security Vulnerabilities

### Please Do Not
- Open public GitHub issues for security vulnerabilities
- Disclose vulnerabilities before a patch is available

### Please Do
1. **Email:** security@scmd.dev (or repository owner)
2. **Include:**
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)
3. **Expected Response Time:** Within 48 hours

### What to Expect
1. **Acknowledgment:** Within 48 hours
2. **Assessment:** Within 1 week
3. **Fix Timeline:**
   - Critical (P0): 24-48 hours
   - High (P1): 1 week
   - Medium (P2): 2 weeks
   - Low (P3): Next release
4. **Public Disclosure:** After patch is released (coordinated)

## Security Disclosure Policy

### Timeline
1. **Day 0:** Vulnerability reported
2. **Day 2:** Acknowledged and assessed
3. **Day 7:** Fix developed and tested
4. **Day 14:** Patch released
5. **Day 21:** Public disclosure (if critical)

### Credits
Security researchers who responsibly disclose vulnerabilities will be credited in:
- Security advisory
- CHANGELOG.md
- GitHub security advisories

## Recent Security Fixes

### January 2026
- **P0 Critical:** Fixed slash command name validation (path traversal + command injection)
- **P1 High:** Fixed repository URL validation (SSRF + file:// access)
- **Created:** Comprehensive validation package with 52+ test cases
- **Result:** 21/21 security tests passing

See: Commit `4244400` - "fix: critical security vulnerabilities - P0 & P1 fixes"

## Security Best Practices for Users

### 1. Slash Commands
- Only add commands from trusted sources
- Review slash command YAMLs before adding
- Use `scmd slash list` to audit configured commands
- Remove unused commands with `scmd slash remove`

### 2. Repositories
- Only add repositories from trusted sources
- Prefer HTTPS over HTTP
- Verify repository URLs before adding
- Regularly audit with `scmd repo list`

### 3. LLM Backends
- Use API keys with least privilege
- Rotate API keys regularly
- Don't commit API keys to git
- Use environment variables for secrets

### 4. File Operations
- Be cautious with `-o` output paths
- Review generated files before using
- Don't pipe untrusted input to scmd

## Security Audit History

| Date | Auditor | Scope | Result |
|------|---------|-------|--------|
| 2026-01 | Internal | Comprehensive test suite | 2 critical issues fixed |

## Compliance

### Supported

Versions of scmd in the security support window:
- **Latest stable release:** ✅ Fully supported
- **Previous minor version:** ✅ Critical fixes only
- **Older versions:** ❌ Not supported

### Update Recommendation

**Always use the latest version** for maximum security.

Check version: `scmd version`
Update: `brew upgrade scmd` (or your package manager)

## Contact

- **Security Issues:** security@scmd.dev
- **General Issues:** GitHub Issues
- **Pull Requests:** GitHub Pull Requests

---

**Last Updated:** January 2026
**Security Contact:** security@scmd.dev
**PGP Key:** (to be added)
