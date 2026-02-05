package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/asgladnev/golang-tcp-server/internal/config"
	"github.com/asgladnev/golang-tcp-server/internal/server"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	//logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfgPath := "config.yaml"
	cfg, err := config.LoadFromFile(cfgPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	srv := server.New(*cfg, logger)

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
