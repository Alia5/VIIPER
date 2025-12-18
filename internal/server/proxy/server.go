package proxy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Alia5/VIIPER/internal/log"
)

type Server struct {
	listenAddr        string
	upstreamAddr      string
	connectionTimeout time.Duration
	logger            *slog.Logger
	rawLogger         log.RawLogger
	ln                net.Listener
}

func New(listenAddr, upstreamAddr string, connectionTimeout time.Duration, logger *slog.Logger, rawLogger log.RawLogger) *Server {
	return &Server{
		listenAddr:        listenAddr,
		upstreamAddr:      upstreamAddr,
		connectionTimeout: connectionTimeout,
		logger:            logger,
		rawLogger:         rawLogger,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.listenAddr, err)
	}
	s.ln = ln
	s.logger.Info("USB-IP proxy listening", "addr", s.listenAddr)

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(strings.ToLower(err.Error()), "use of closed network connection") {
				s.logger.Info("Proxy server stopped")
				return nil
			}
			s.logger.Error("Accept error", "error", err)
			continue
		}
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			if err := tcpConn.SetNoDelay(true); err != nil {
				s.logger.Warn("failed to set TCP_NODELAY", "error", err)
			}
		}
		s.logger.Info("Client connected", "remote", clientConn.RemoteAddr())
		go s.handleProxy(clientConn)
	}
}

func (s *Server) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Server) handleProxy(clientConn net.Conn) {
	defer clientConn.Close()

	upstreamConn, err := net.DialTimeout("tcp", s.upstreamAddr, s.connectionTimeout)
	if err != nil {
		s.logger.Error("Failed to connect to upstream", "upstream", s.upstreamAddr, "error", err)
		return
	}
	defer upstreamConn.Close()

	s.logger.Info("Proxying connection", "client", clientConn.RemoteAddr(), "upstream", upstreamConn.RemoteAddr())

	err = clientConn.SetDeadline(time.Now().Add(s.connectionTimeout))
	if err != nil {
		s.logger.Error("Failed to set client deadline", "error", err)
		return
	}
	err = upstreamConn.SetDeadline(time.Now().Add(s.connectionTimeout))
	if err != nil {
		s.logger.Error("Failed to set upstream deadline", "error", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		bytes, err := s.copyWithLogging(upstreamConn, clientConn, true)
		if err != nil && !isExpectedDisconnect(err) {
			s.logger.Debug("Client->Server copy error", "error", err)
		}
		s.logger.Debug("Client->Server stream ended", "bytes", bytes)
		halfClose(upstreamConn, true)
		halfClose(clientConn, false)
	}()

	go func() {
		defer wg.Done()
		bytes, err := s.copyWithLogging(clientConn, upstreamConn, false)
		if err != nil && !isExpectedDisconnect(err) {
			s.logger.Debug("Server->Client copy error", "error", err)
		}
		s.logger.Debug("Server->Client stream ended", "bytes", bytes)
		halfClose(clientConn, true)
		halfClose(upstreamConn, false)
	}()

	wg.Wait()
	s.logger.Info("Connection closed", "client", clientConn.RemoteAddr())
}

func (s *Server) copyWithLogging(dst net.Conn, src net.Conn, clientToServer bool) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	parser := NewParser(s.logger)
	firstPacket := true

	for {
		n, rerr := src.Read(buf)
		if n > 0 {
			s.rawLogger.Log(clientToServer, buf[:n])

			parser.Parse(buf[:n], clientToServer)

			if firstPacket {
				err := src.SetDeadline(time.Time{})
				if err != nil {
					s.logger.Error("Failed to clear source deadline", "error", err)
					return total, err
				}
				err = dst.SetDeadline(time.Time{})
				if err != nil {
					s.logger.Error("Failed to clear destination deadline", "error", err)
					return total, err
				}
				firstPacket = false
			}

			wn, werr := dst.Write(buf[:n])
			total += int64(wn)
			if werr != nil {
				return total, werr
			}
			if wn != n {
				return total, fmt.Errorf("short write: wrote %d of %d", wn, n)
			}
		}

		if rerr != nil {
			if ne, ok := rerr.(net.Error); ok && ne.Timeout() {
				continue
			}
			if rerr == io.EOF {
				return total, nil
			}
			return total, rerr
		}
	}
}

func halfClose(conn net.Conn, write bool) {
	if tc, ok := conn.(*net.TCPConn); ok {
		if write {
			_ = tc.CloseWrite()
		} else {
			_ = tc.CloseRead()
		}
	}
}

func isExpectedDisconnect(err error) bool {
	if err == nil || err == io.EOF {
		return true
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "connection reset") ||
		strings.Contains(e, "broken pipe") ||
		strings.Contains(e, "forcibly closed")
}
