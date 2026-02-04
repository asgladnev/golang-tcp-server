package server

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/asgladnev/golang-tcp-server/internal/config"
	"go.uber.org/zap"
)

type TCPServer struct {
	logger *zap.Logger
	cfg    config.ServerConfig
	wg     sync.WaitGroup
}

type TCPClient struct {
	conn net.Conn
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

func (srv *TCPServer) handleClient(ctx context.Context, client *TCPClient) {
	defer client.conn.Close() // Гарантированно закрываем соединение при выходе из функции
	srv.wg.Add(1)
	defer srv.wg.Done()

	clientIP := client.conn.RemoteAddr().String()
	srv.logger.Info("New client connection", zap.String("client_ip", clientIP))

	srv.enableTCPKeepAlive(client.conn)

	if srv.cfg.ChunkSize <= 0 {
		srv.logger.Error("chunkSize must be > 0")
		return
	}
	srv.readClientData(ctx, client, clientIP)
}

func (srv *TCPServer) readClientData(ctx context.Context, client *TCPClient, clientIP string) {
	buffer := make([]byte, srv.cfg.ChunkSize) // Буфер для чтения данных чанками фиксированного размера
	var totalBuffer uint64 = 0                // Счётчик общего объёма принятых данных (в байтах)

	// Цикл до EOF (закрытие соединения) или исчерпания буфера
	for {
		// Проверяем контекст перед каждое операцией
		select {
		case <-ctx.Done():
			srv.logger.Info("Stopping client handler due to server shutdown", zap.String("client_ip", clientIP))
			return
		default:
		}
		// Защита от медленных клиентов
		if err := client.conn.SetReadDeadline(time.Now().Add(srv.cfg.ReadTimeout)); err != nil {
			srv.logger.Error("Failed to set read deadline", zap.String("client_ip", clientIP), zap.Error(err))
			return
		}
		// Читаем очередной чанк данных из TCP-соединения
		n, err := client.conn.Read(buffer)
		// Если прочитано больше нуля байт — обрабатываем данные
		if n > 0 {
			// Если буфер превышен -> рвем соединение
			if !srv.checkMaxBuffer(clientIP, totalBuffer, n) {
				return
			}

			totalBuffer += uint64(n) // Записываем чанк
			//fmt.Print(string(buffer[:n])) // Выводим сообщение
		}

		// Обработка ошибок
		if err != nil {
			// Клиент корректно завершил передачу (отправил FIN (EOF))
			if err == io.EOF {
				srv.logger.Debug("Received all data from client",
					zap.String("client_ip", clientIP),
					zap.Uint64("total_bytes", totalBuffer))
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				srv.logger.Warn("Client read timeout (slow client or disconnect)",
					zap.String("client_ip", clientIP))
			} else {
				srv.logger.Error("Connection error",
					zap.String("client_ip", clientIP),
					zap.Error(err))
			}
			return
		}
	}
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
