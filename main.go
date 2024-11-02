package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/hyperfyodor/ssosage_proto"
	"google.golang.org/grpc"
)

func main() {

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
