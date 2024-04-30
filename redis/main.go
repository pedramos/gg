package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"
)

const (
	defaultListenAddr = ":5001"
)

type Server struct {
	Config
	peers map[*Peer]bool
	ln    net.Listener

	addPeerCh chan *Peer

	quitCh chan struct{}

	isreadyCh chan struct{}

	msgCh chan []byte

	kv StorageEngine
}

type Config struct {
	ListenAddr string
}

func NewServer(cfg Config) (*Server, error) {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	storageEng, err := NewStorageEngine(KeyValEngineType)
	if err != nil {
		return nil, fmt.Errorf("creating storage engine: %w", err)
	}
	s := &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		isreadyCh: make(chan struct{}),
		msgCh:     make(chan []byte),
		kv:        storageEng,
	}

	return s, nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	slog.Info("server started", "addr", s.ln.Addr())
	s.isreadyCh <- struct{}{}

	go s.eventloop()
	return s.acceptLoop()
}

func (s *Server) WaitReady() error {
	select {
	case <-s.isreadyCh:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for server to be ready")
	}
}

func (s *Server) eventloop() {
	for {
		select {
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
		case <-s.quitCh:
			break
		default:
			continue
		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "msg", err)
			continue
		}
		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.kv)
	s.addPeerCh <- peer

	slog.Info("new peer connected", "peer", conn.RemoteAddr())
	errs := make(chan error)
	go peer.readLoop(errs)
	for err := range errs {
		slog.Error("Error from peer", "peer", peer.conn.RemoteAddr(), "error", err)
	}

}

func main() {
	server, err := NewServer(Config{})
	if err != nil {
		slog.Error("new server error", "msg", err)
		os.Exit(1)
	}
	log.Fatal(server.Start())
}
