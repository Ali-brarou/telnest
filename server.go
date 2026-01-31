package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	Addr string
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Println("telnest listening on", s.Addr)
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	defer func() {
		log.Printf("[%s] disconnected\n", c.RemoteAddr())
		c.Close()
	}()

	log.Printf("[%s] connected", c.RemoteAddr())

	tr := NewTelnetReader(c)

	// try to capture the first byte sent by the client.
	// so early negotiations (NEW-ENV) are processed
	err := tr.Prime(c, time.Second)
	if err != nil {
		return
	}
	/* if the attacker put a USER env then skip auth */
	skipAuth := false
	_, ok := tr.Env["USER"]
	if ok {
		skipAuth = true
		log.Printf(
			"[%s] TELNET CVE-2026-24061 exploited",
			c.RemoteAddr(),
		)
	}

	rw := bufio.NewReadWriter(
		bufio.NewReader(tr),
		bufio.NewWriter(c),
	)

	// send the banner
	rw.WriteString("\n\rLinux 6.17.9-arch1-1 (ab-device) (pts/5)\n\r\n\r")
	rw.Flush()

	user := "root"
	if !skipAuth {
		// send login prompt
		rw.WriteString("ab-device login: ")
		rw.Flush()
		c.SetDeadline(time.Now().Add(30 * time.Second))
		user, err = rw.ReadString('\n')
		if err != nil {
			return
		}

		// hide password client side
		rw.Write([]byte{IAC, WILL, ECHO})
		// send password prompt
		rw.WriteString("Password : ")
		rw.Flush()
		c.SetDeadline(time.Now().Add(30 * time.Second))
		pass, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		// show text client side
		rw.Write([]byte{IAC, WONT, ECHO, 0xa})
		rw.Flush()

		log.Printf(
			"[%s] TELNET auth attempt user=%q pass=%q env=%v",
			c.RemoteAddr(),
			strings.TrimSpace(user),
			strings.TrimSpace(pass),
			tr.Env,
		)
	}

	prompt := fmt.Sprintf("[%s@ab-device ~]# ", strings.TrimSpace(user))
	for {
		rw.WriteString(prompt)
		rw.Flush()

		c.SetDeadline(time.Now().Add(5 * time.Minute))
		cmd, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		log.Printf("[%s] Shell command %q\n", c.RemoteAddr(), strings.TrimSpace(cmd))

		rw.WriteString(fakeShell(cmd))
		rw.Flush()
	}

	/*
		time.Sleep(2 * time.Second)
		rw.WriteString("\r\nLogin incorrect\r\n")
		rw.Flush()
	*/
}
