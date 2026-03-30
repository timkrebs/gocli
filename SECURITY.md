# Security Policy

## Supported Versions

Only the latest released version of gocli receives security fixes.
Older versions will not receive patches.

| Version | Supported |
|---------|-----------|
| latest  | Y |
| < latest | N |

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report security issues privately using one of the following methods:

1. **GitHub Private Vulnerability Reporting** (preferred):
   Navigate to the [Security tab](https://github.com/timkrebs/gocli/security/advisories/new)
   of this repository and click "Report a vulnerability".

2. **Direct contact**:
   Send a message to the maintainer via the GitHub profile at
   https://github.com/timkrebs with the subject line
   `[SECURITY] gocli vulnerability report`.

### What to include

- A description of the vulnerability and its potential impact.
- Steps to reproduce, including a minimal code example if possible.
- Any known mitigations or workarounds.
- The Go version, OS, and gocli version you are using.

### Response timeline

- **Acknowledgement**: within 3 business days.
- **Initial assessment**: within 7 business days.
- **Fix / advisory**: timeline depends on severity; critical issues are
  prioritised for the next patch release.

We will credit reporters in the release notes unless you prefer to remain
anonymous.

## Scope

As a **Go library** (not a network service or standalone binary), gocli's
attack surface is limited to code paths triggered by the consuming
application. Relevant areas include:

- Argument parsing that could be exploited via crafted CLI arguments.
- Template injection via `CommandHelpTemplate`.
- Path traversal or command injection in autocomplete paths.
- Dependency vulnerabilities (tracked weekly via `govulncheck` in CI).
