package main

import (
	"context"
	"testing"
	"time"

	"github.com/hyperfyodor/ssosage_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	ssosageClient ssosage_proto.SsosageClient
}

func NewSuite(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Minute)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	grpcAddress := "0.0.0.0:3333"

	cc, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	ssosageClient := ssosage_proto.NewSsosageClient(cc)

	return ctx, &Suite{
		T:             t,
		ssosageClient: ssosageClient,
	}
}
