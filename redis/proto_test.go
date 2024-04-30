package main

import "testing"

func TestProtocol(t *testing.T) {
	raw := []byte("*3\r\n$3\r\nset\r\n$6\r\nleader\r\n$7\r\nCharlie\r\n")

	cmd, err := parseCommand(raw)
	if err != nil {
		t.Errorf("failed to parse cmd: %s", err)
	}
	if cmd == nil {
		t.Errorf("failed to parse cmd: %s", err)
	}

}
