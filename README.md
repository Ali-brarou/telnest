# telnest

`telnest` is another Telnet honeypot written in Go.

It focuses on observing and logging exploitation attempts related to
**CVE-2026-24061** and common Telnet‑based malware activity.

The service presents a realistic login prompt and interactive shell while
safely capturing credentials and commands without executing anything.

