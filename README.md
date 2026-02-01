# telnest

`telnest` is another Telnet honeypot written in Go.

It focuses on observing and logging exploitation attempts related to
**CVE-2026-24061** and common Telnet‑based malware activity.

The service presents a realistic login prompt and interactive shell while
safely capturing credentials and commands without executing anything.

### Example log output
```
2026/02/01 22:08:57 telnest listening on :23
2026/02/01 22:09:21 [192.168.1.1:58822] connected
2026/02/01 22:09:21 [192.168.1.1:58822] TELNET CVE-2026-24061 exploited
2026/02/01 22:09:21 [192.168.1.1:58822] got a shell
2026/02/01 22:09:40 [192.168.1.1:58822] Shell command "whoami"
2026/02/01 22:09:40 [192.168.1.1:58822] Shell command "ls"
2026/02/01 22:09:47 [192.168.1.1:58822] disconnected
```
