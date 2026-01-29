package main 

import "net"

type HoneyTelnet struct {
	listener net.listener
	running bool
}

func NewHoneyTelnet() *HoneyTelnet {
	return &HoneyTelnet {

	}
}
