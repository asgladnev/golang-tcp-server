package server

import (
	"net"

	"go.uber.org/zap"
)

// Включаем TCP KeepAlive для соединений
func (srv *TCPServer) enableTCPKeepAlive(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetKeepAlive(true); err != nil {
			srv.logger.Debug("Failed to enable TCP KeepAlive", zap.Error(err))
		}
		if err := tcpConn.SetKeepAlivePeriod(srv.cfg.KeepAlive); err != nil {
			srv.logger.Error("Failed to set TCP KeepAlive period", zap.Error(err))
		}
	}
}

// Проверка, чтобы клиент не превысил максимальный буфер
func (srv *TCPServer) checkMaxBuffer(clientIP string, totalBuffer uint64, n int) bool {
	if srv.cfg.MaxBuffer > 0 && totalBuffer+uint64(n) > srv.cfg.MaxBuffer {
		srv.logger.Info("Client exceeded max buffer size",
			zap.String("client_ip", clientIP),
			zap.Uint64("received_bytes", totalBuffer),
			zap.Uint64("max_buffer", srv.cfg.MaxBuffer))
		return false
	}
	return true
}
