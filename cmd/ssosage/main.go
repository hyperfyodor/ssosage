package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	config "ssosage/internal/config/ssosage"
	"ssosage/internal/hasher"
	"ssosage/internal/helpers"
	"ssosage/internal/server"
	service "ssosage/internal/services/ssosage"
	"ssosage/internal/storage/sqlite"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/hyperfyodor/ssosage_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	log := setupLogger(cfg.Env)
	storage, err := sqlite.New(cfg.StoragePath)

	if err != nil {
		panic("failed to create storage")
	}

	bcryptHasher := hasher.BcryptHasher{}
	ssosage := service.New(log, storage, storage, &bcryptHasher)

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
	))

	grpcHadnler := server.New(ssosage)

	ssosage_proto.RegisterSsosageServer(grpcServer, grpcHadnler)

	go func() {

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GrpcPort))
		if err != nil {
			panic(err)
		}

		log.Info("server listening at", "addr", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("failed to serve", helpers.SlErr(err))

			panic(err)
		}

	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	grpcServer.Stop()
	storage.Stop()
	log.Info("Stopped ;)")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
