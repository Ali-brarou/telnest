package main

import (
	"fmt"
	"strings"
)

func fakeShell(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var out strings.Builder
	for _, cmd := range splitCommands(input) {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		out.WriteString(runCommand(cmd))
	}
	return out.String()
}

func splitCommands(s string) []string {
	var cmds []string
	var cur strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == '&' && i+1 < len(s) && s[i+1] == '&' {
			cmds = append(cmds, cur.String())
			cur.Reset()
			i++
			continue
		}
		if s[i] == ';' {
			cmds = append(cmds, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(s[i])
	}
	if cur.Len() > 0 {
		cmds = append(cmds, cur.String())
	}
	return cmds
}

func splitArgs(s string) []string {
	var args []string
	var cur strings.Builder
	var quote byte

	for i := 0; i < len(s); i++ {
		c := s[i]

		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}

		case c == '"' || c == '\'':
			quote = c

		case c == ' ' || c == '\t':
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}

		default:
			cur.WriteByte(c)
		}
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}


func runCommand(cmd string) string {
	fields := splitArgs(cmd)
	if len(fields) == 0 {
		return ""
	}

	switch fields[0] {
	case "echo":
		return handleEcho(fields[1:])

	case "uname":
		if len(fields) > 1 && fields[1] == "-a" {
			return "Linux ab-device 6.17.9-arch1-1 #1 SMP PREEMPT x86_64 GNU/Linux\n"
		}
		return "Linux\n"

	case "id":
		return "uid=0(root) gid=0(root) groups=0(root)\n"

	case "pwd":
		return "/root\n"

	case "whoami":
		return "root\n"

	case "hostname":
		return "ab-device\n"

	case "ls":
		return "bin  etc  home  lib  tmp  var\n"

	case "cd":
		return ""

	case "ps":
		return "PID TTY          TIME CMD\n1 pts/0    00:00:00 bash\n"

	case "ifconfig":
		return "eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500\n"

	case "shell", "linuxshell", "sh", "/bin/sh":
		return ""

	case "enable":
		return ""

	case "wget", "curl":
		return "Connecting...\n"

	case "ftpget", "ftpput":
		return ""

	case "nc", "netcat":
		return ""

	case "chmod", "chown", "mount", "umount":
		return ""

	case "rm", "cp", "mv", "touch", "mkdir":
		return ""

	case "/bin/busybox", "busybox":
		return "BusyBox v1.36.1 (built-in shell)\n"

	case "ping":
		if len(fields) < 2 {
			return bashErr("ping: missing host operand\n")
		}
		host := fields[1]
		return fmt.Sprintf(
			"PING %s (127.0.0.1): 56 data bytes\n--- %s ping statistics ---\n1 packets transmitted, 1 received, 0%% packet loss\n",
			host, host,
		)

	case "cat":
		if len(fields) < 2 {
			return bashErr("cat: missing operand\n")
		}

		path := fields[1]

		switch path {
		case "/proc/cpuinfo":
			return `processor	: 0
model name	: ARMv7 Processor rev 5 (v7l)
BogoMIPS	: 38.40
`

		case "/proc/meminfo":
			return `MemTotal:        128000 kB
MemFree:          64000 kB
Buffers:           8000 kB
Cached:           12000 kB
`

		case "/etc/passwd":
			return `root:x:0:0:root:/root:/bin/sh
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
`

		case "/etc/shadow":
			return bashErr("cat: /etc/shadow: Permission denied\n")

		case "/proc/version":
			return "Linux version 6.17.9-arch1-1 (gcc version 13.2.1)\n"

		default:
			// handle multiple files: cat a b c
			if len(fields) > 2 {
				return bashErr(fmt.Sprintf(
					"cat: %s: No such file or directory\n",
					strings.Join(fields[1:], " "),
				))
			}
			return bashErr(fmt.Sprintf(
				"cat: %s: No such file or directory\n",
				path,
			))
		}

	case "w", "who", "last":
		return "root     pts/0        2024-02-01 12:34 (10.0.0.1)\n"

	case "netstat", "ss":
		return "Active Internet connections (servers and established)\nProto Recv-Q Send-Q Local Address Foreign Address State\ntcp        0      0 0.0.0.0:22          0.0.0.0:*         LISTEN\n"

	case "find":
		if len(fields) > 1 && strings.Contains(fields[1], "passwd") {
			return "/etc/passwd\n"
		}
		return ".\n"

	case "grep":
		if len(fields) > 2 && fields[2] == "/etc/passwd" {
			return "root:x:0:0:root:/root:/bin/bash\n"
		}
		return ""

	case "python", "python3", "perl", "php":
		return "Python 3.11.2 (main, Feb  1 2024, 00:00:00) [GCC 13.2.1]\n"

	case "env", "printenv":
		return `USER=root
HOME=/root
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
SHELL=/bin/bash
PWD=/root
TERM=xterm-256color
`


	case "exit", "logout":
		return "logout\n"

	default:
		return bashErr(fmt.Sprintf("%s: command not found\n", fields[0]))
	}
}

func bashErr(s string) string {
	return "-bash: " + s
}

func handleEcho(args []string) string {
	interpretEscapes := false
	start := 0

	if len(args) > 0 && args[0] == "-e" {
		interpretEscapes = true
		start = 1
	}

	clean := make([]string, 0, len(args[start:]))
	for _, a := range args[start:] {
		if strings.HasPrefix(a, ">") || strings.Contains(a, "/dev/null") {
			break
		}
		clean = append(clean, a)
	}

	text := strings.Join(clean, " ")
	if interpretEscapes {
		text = unescape(text)
	}
	return text + "\n"
}

func unescape(s string) string {
	var out strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				out.WriteByte('\n')
			case 'r':
				out.WriteByte('\r')
			case 't':
				out.WriteByte('\t')
			case '\\':
				out.WriteByte('\\')

			case 'x':
				// hex escape: \xNN
				if i+2 < len(s) {
					hex := s[i+1 : i+3]
					if v, err := parseHexByte(hex); err == nil {
						out.WriteByte(v)
						i += 2
					} else {
						out.WriteString(`\x`)
					}
				} else {
					out.WriteString(`\x`)
				}

			default:
				out.WriteByte('\\')
				out.WriteByte(s[i])
			}
		} else {
			out.WriteByte(s[i])
		}
	}

	return out.String()
}

func parseHexByte(s string) (byte, error) {
	var v byte
	for i := 0; i < 2; i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v = v<<4 | (c - '0')
		case c >= 'a' && c <= 'f':
			v = v<<4 | (c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v = v<<4 | (c - 'A' + 10)
		default:
			return 0, fmt.Errorf("invalid hex")
		}
	}
	return v, nil
}
