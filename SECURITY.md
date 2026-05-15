# Security Policy

## Supported Versions

Only the latest release on the `main` branch is supported.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Email **michael@mjashley.com** with:

- Description of the vulnerability
- Steps to reproduce
- Impact assessment (if known)

You will receive an acknowledgment within 48 hours.

## Scope

The following are in scope:

- Volume driver plugin security (file path traversal, unauthorized mount access)
- Docker socket interaction
- Privilege escalation via the plugin

The following are out of scope:

- Docker daemon vulnerabilities
- Host filesystem permissions outside plugin-managed paths

## Disclosure Policy

We follow coordinated disclosure. We will credit reporters in release notes unless anonymity is requested.
