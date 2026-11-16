# Security Policy

## Supported versions

Only the latest release receives security fixes.

| Version | Supported |
|---|---|
| 0.1.x | Yes |

## Reporting a vulnerability

Email **gaspar.mdp@gmail.com** with:

- a description of the issue,
- steps to reproduce,
- the affected version (`repora --version`),
- the operating system and Git version.

Do not open a public issue for security reports. You will receive an acknowledgement within
a few days.

## Scope

Repora runs on your machine and executes the local Git binary against your repositories.
The most relevant classes of issues are:

- command construction that could inject arguments into Git,
- configuration writes that could lose or corrupt data,
- path handling that could escape the intended repository,
- confirmation bypasses for destructive actions.

Out of scope:

- vulnerabilities in Git itself,
- issues that require an already-compromised local account,
- behavior of hooks or credential helpers configured by the user.

## What Repora never does

- It does not execute shell strings; every Git call is an argument vector.
- It does not store credentials, tokens, or remote passwords.
- It does not send telemetry or contact any service other than your configured Git remotes.
- It does not rewrite history, force-push, reset, or clean repositories.
