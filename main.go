package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	"github.com/hyperfyodor/ssosage_proto"
	"google.golang.org/grpc"

	// Драйвер для выполнения миграций SQLite 3
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	m, err := migrate.New(
		"file://.",
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", "./ssosage.db", "migrations_ssosage"),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
		} else {
			panic(err)
		}
	}

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	storage, err := NewSqliteStorage("./ssosage.db")

	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()

	ssosage := NewSsosage(log, storage)

	myGrpcServerImpl := server{ssosage: ssosage}

	ssosage_proto.RegisterSsosageServer(grpcServer, &myGrpcServerImpl)

	go func() {

		lis, err := net.Listen("tcp", ":3333")
		if err != nil {
			panic(err)
		}

		log.Info("server listening at", "addr", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("failed to serve", slErr(err))

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
