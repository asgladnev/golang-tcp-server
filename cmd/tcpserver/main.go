package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asgladnev/golang-tcp-server/internal/config"
	"github.com/asgladnev/golang-tcp-server/internal/server"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.ServerConfig{
		Host:        "127.0.0.1",
		Port:        9000,
		MaxBuffer:   65535,
		ChunkSize:   1024,
		ReadTimeout: 60 * time.Second,
	}

	srv := server.New(cfg, logger)

	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		logger.Error("Error start server", zap.Error(err))
		return
	}
	defer listener.Close()

	logger.Info("Start server", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))

	// Создаём контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Отслеживаем сигналы ОС для завершения сервера
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
		cancel()         // уведомляем все горутины о завершении
		listener.Close() // закрываем listener чтобы остановить Accept()
	}()

	// Запуск сервера с контекстом
	srv.StartServer(ctx, listener)
	srv.Wait()
	logger.Info("Server stopped gracefully")
}
