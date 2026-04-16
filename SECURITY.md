# Security Policy

## Supported versions

`gone` is actively developed. Only the latest `main` branch and tagged releases are supported.

| Version | Supported |
|:--|:--|
| Latest `main` | ✅ |
| Older tags | ❌ |

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

If you discover a security issue in `gone`, email the maintainer directly via the contact form on [agustiny-dev.ar](http://agustiny-dev.ar). Include:

- A clear description of the vulnerability
- Steps to reproduce
- Affected versions or commits
- Any proof-of-concept (optional, but helpful)

You will receive an acknowledgment within **72 hours**. We'll investigate, prepare a fix, and coordinate disclosure with you before publishing.

## Scope

In-scope issues include:

- Command injection via scanned paths or shell RC lines
- Unsafe handling of `osascript` arguments
- Privilege escalation from user to root
- File deletion that bypasses macOS Trash (no Put Back possible)
- Sensitive data exposure in the operations log

Out-of-scope:

- Issues requiring physical access to an unlocked machine
- Social engineering attacks
- Vulnerabilities in upstream dependencies (please report those upstream)
- macOS itself

## Hall of fame

Contributors who report valid security issues will be credited here after coordinated disclosure.
