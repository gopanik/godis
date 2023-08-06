package internal

import (
	"bufio"
	"fmt"
	"net"
	"sync"

	"golang.org/x/exp/slog"
)

type Server struct {
	addr     string
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	logger   *slog.Logger
}

func NewServer(addr string, logger *slog.Logger) *Server {
	return &Server{
		quit:   make(chan interface{}),
		addr:   addr,
		logger: logger,
	}
}

// ListenAndServe starts the server with the provided address.
// It will spin up a separate goroutine to accept connections.
func (s *Server) ListenAndServe() error {
	s.logger.Info("Godis is starting")
	listener, err := net.Listen("tcp", "localhost"+s.addr)
	if err != nil {
		return err
	}

	s.logger.Info("Server initialized")
	s.listener = listener
	s.wg.Add(1)

	go s.serve()
	return nil
}

func (s *Server) serve() {
	defer s.wg.Done()

	s.logger.Info("Ready to accept connections")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				s.logger.Error("Error accepting a connection", slog.String("err", err.Error()))
			}
		} else {
			s.wg.Add(1)
			go func() {
				s.handleConn(conn)
				s.wg.Done()
			}()
		}
	}
}

func (s *Server) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		cmd := scanner.Text()
		if cmd == "PING" {
			conn.Write([]byte("+PONG\r\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s', with args beginning with: \r\n", cmd)))
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Warn("Error reading from conn", slog.String("err", err.Error()))
	}

	conn.Close()
}

func (s *Server) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}
