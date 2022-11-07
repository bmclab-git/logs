package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/Lukiya/logs/svc"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/host"
	"github.com/syncfuture/host/sfasthttp"
	"github.com/syncfuture/host/sgrpc"
)

var (
	logService = new(svc.LogService)
)

func main() {
	grpcHost := sgrpc.NewGRPCServiceHost(core.GrpcCP)
	go func() {
		model.RegisterLogEntryServiceServer(grpcHost.GetGRPCServer(), logService)
		slog.Fatal(grpcHost.Run())
	}()

	webHost := sfasthttp.NewFHWebHost(core.MainCP)
	webHost.GET("/api/logs", getLogs)

	slog.Fatal(webHost.Run())
}

func getLogs(ctx host.IHttpContext) {
	query := new(model.LogEntriesQuery)
	ctx.ReadQuery(query)

	time.Sleep(time.Second * 1)

	rs, err := logService.GetLogEntries(context.Background(), query)
	if !host.HandleErr(err, ctx) {
		jsonBytes, err := json.Marshal(rs)
		if !host.HandleErr(err, ctx) {
			ctx.WriteJsonBytes(jsonBytes)
		}
	}
}
