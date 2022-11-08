package main

import (
	"context"
	"encoding/json"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/Lukiya/logs/svc"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/host"
	"github.com/syncfuture/host/sconsul"
	"github.com/syncfuture/host/sfasthttp"
	"github.com/syncfuture/host/sgrpc"
)

var (
	logService = new(svc.LogService)
)

func main() {
	grpcHost := sgrpc.NewGRPCServiceHost(core.GrpcCP)
	go func() {
		// Register to consul
		sconsul.RegisterServiceInfo(core.GrpcCP)

		// Register GRPC
		model.RegisterLogEntryServiceServer(grpcHost.GetGRPCServer(), logService)
		slog.Fatal(grpcHost.Run())
	}()

	webHost := sfasthttp.NewFHWebHost(core.WebCP)
	webHost.GET("/api/logs", getLogs)

	slog.Fatal(webHost.Run())
}

func getLogs(ctx host.IHttpContext) {
	query := new(model.LogEntriesQuery)
	ctx.ReadQuery(query)

	rs, err := logService.GetLogEntries(context.Background(), query)
	if !host.HandleErr(err, ctx) {
		jsonBytes, err := json.Marshal(rs)
		if !host.HandleErr(err, ctx) {
			ctx.WriteJsonBytes(jsonBytes)
		}
	}
}
