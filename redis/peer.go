package main

import (
	"fmt"
	"log/slog"
	"net"
)

type Peer struct {
	kv    StorageEngine
	conn  net.Conn
	msgCh chan []byte
}

func NewPeer(conn net.Conn, kv StorageEngine) *Peer {
	return &Peer{
		conn: conn,
		kv:   kv,
	}
}

func (p *Peer) readLoop(errs chan error) {
	buf := make([]byte, 1024)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			errs <- fmt.Errorf("error from peer read: %w", err)
			close(errs)
		}
		if n == 0 {
			slog.Warn(
				"Eval raw message error",
				"warn",
				"empty message received from client",
			)
			continue
		}
		cmd, err := parseRawMessage(buf)
		if err != nil {
			errs <- fmt.Errorf("eval raw message error: %s", err)
		}
		val, err := cmd.Exec(p.kv)
		if err != nil {
			panic("TODO: must handle error")
		}
		if val != nil {
			if err := p.WriteVal(val); err != nil {
				slog.Warn(
					"failed to msg to client",
					"warn",
					err,
					"cmd",
					cmd.Stmt(),
				)
				p.Errorf("failed to msg to client, responding to cmd `%s` :%s", cmd.Stmt(), err)
			}
			continue
		}
		p.WriteOk()
	}
}

func (p *Peer) WriteOk() error {
	_, err := p.conn.Write([]byte("+OK\r\n"))
	return err
}

func (p *Peer) WriteVal(val any) error {
	_, err := p.conn.Write([]byte(fmt.Sprintf("+%v\r\n", val)))
	return err
}

func (p *Peer) Errorf(msg string, args ...any) error {
	errStr := fmt.Sprintf(msg, args...)
	_, err := p.conn.Write([]byte(fmt.Sprintf("-%s\r\n", errStr)))
	return err
}

func parseRawMessage(msg []byte) (Command, error) {
	cmd, err := parseCommand(msg)
	if err != nil {
		return nil, fmt.Errorf("parsing msg error: %w", err)
	}
	return cmd, nil
}
