package main

import (
	"github.com/Lukiya/logs/core"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/sfasthttp"
	"github.com/syncfuture/host/sgrpc"
)

func main() {
	grpcHost := sgrpc.NewGRPCServiceHost(core.GrpcCP)
	go func() {
		slog.Fatal(grpcHost.Run())
	}()

	webHost := sfasthttp.NewFHWebHost(core.MainCP)

	slog.Fatal(webHost.Run())
}
