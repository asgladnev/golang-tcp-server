package server

import (
	"context"
	"net"
	"sync"

	"github.com/asgladnev/golang-tcp-server/internal/config"
	"go.uber.org/zap"
)

type TCPServer struct {
	logger *zap.Logger
	cfg    config.ServerConfig
	wg     sync.WaitGroup
}

func New(cfg config.ServerConfig, logger *zap.Logger) *TCPServer {
	return &TCPServer{
		cfg:    cfg,
		logger: logger,
	}
}

func (srv *TCPServer) Wait() {
	srv.wg.Wait()
}

func (srv *TCPServer) StartServer(ctx context.Context, listenr net.Listener) {
	for {
		select {
		case <-ctx.Done(): // graceful shutdown
			srv.logger.Info("Server shutting down, stop accepting new connections")
			return
		default:
		}

		conn, err := listenr.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				srv.logger.Info("Listener closed, stop accepting")
				return
			}
			srv.logger.Error("Error accept connection", zap.Error(err))
			continue
		}
		client := &TCPClient{conn: conn}
		go srv.handleClient(ctx, client)
	}
}
