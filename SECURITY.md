# Security Policy

## Supported versions

This project is in active development. Security fixes are applied to the
current `main` branch. There is no supported release series yet.

## Report a vulnerability

Please do **not** open a public issue containing an exploit, credential, or
other sensitive details. Send a private report to
[the maintainer](mailto:shivamkumar872000@gmail.com) with:

- A short description of the issue and its possible impact.
- Steps to reproduce using synthetic data.
- The affected commit or component, if known.
- A safe way to contact you for follow-up.

Avoid testing against infrastructure you do not own or control. Do not include
real secrets in a report. The maintainer will acknowledge the report when
possible, investigate it, and coordinate a fix and disclosure with you.

## Current deployment boundary

The Go API has no authentication or rate limiting. The local example binds
PostgreSQL to loopback, but the API itself must remain private until an
authenticated ingress and network restrictions are in place. No public AWS
deployment is documented or supported yet.
