package main

import (
	"context"
	"embed"
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
	logService       = new(svc.LogService)
	logClientService = new(svc.LogClientService)
)

//go:embed wwwroot
var staticFiles embed.FS

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
	webHost.POST("/api/logs", getLogs)
	webHost.GET("/api/listData", getListData)

	// webHost.POST("/api/clients", getClients)
	// webHost.POST("/api/dbs", getDatabases)
	// webHost.POST("/api/tables", getTables)

	webHost.ServeEmbedFiles("/{filepath:*}", "wwwroot", staticFiles)

	slog.Fatal(webHost.Run())
}

func getLogs(ctx host.IHttpContext) {
	query := new(model.LogEntriesQuery)
	ctx.ReadJSON(query)

	if query.PageSize <= 0 || query.PageIndex < 1 || query.DBName == "" || query.TableName == "" {
		ctx.SetStatusCode(400)
		return
	}

	rs, err := logService.GetLogEntries(context.Background(), query)
	if !host.HandleErr(err, ctx) {
		jsonBytes, err := json.Marshal(rs)
		if !host.HandleErr(err, ctx) {
			ctx.WriteJsonBytes(jsonBytes)
		}
	}
}

type GetListDataResult struct {
	Client    string
	Database  string
	Table     string
	Clients   []string
	Databases []string
	Tables    []string
}

func getListData(ctx host.IHttpContext) {
	r := &GetListDataResult{
		Clients:   make([]string, 0),
		Databases: make([]string, 0),
		Tables:    make([]string, 0),
	}

	r.Client = ctx.GetFormString("client")
	r.Database = ctx.GetFormString("db")

	clientRS, err := logClientService.GetClients(context.Background(), &model.LogClientsQuery{})
	if host.HandleErr(err, ctx) {
		return
	}
	r.Clients = clientRS.LogClients

	if r.Client != "" {
		// Has client id provided
		dbRS, err := logClientService.GetDatabases(context.Background(), &model.DatabasesQuery{
			ClientID: r.Client,
		})
		if host.HandleErr(err, ctx) {
			return
		}
		r.Databases = dbRS.Databases

		if r.Database != "" {
			tabRS, err := logClientService.GetTables(context.Background(), &model.TablesQuery{
				Database: r.Database,
			})
			if host.HandleErr(err, ctx) {
				return
			}
			r.Tables = tabRS.Tables
		}
	}

	jsonBytes, err := json.Marshal(r)
	if !host.HandleErr(err, ctx) {
		ctx.WriteJsonBytes(jsonBytes)
	}
}

/*
func getClients(ctx host.IHttpContext) {
	query := new(model.LogClientsQuery)
	ctx.ReadJSON(query)

	clientRS, err := logClientService.GetClients(context.Background(), query)
	if host.HandleErr(err, ctx) {
		return
	}

	jsonBytes, err := json.Marshal(clientRS)
	if !host.HandleErr(err, ctx) {
		ctx.WriteJsonBytes(jsonBytes)
	}
}

func getDatabases(ctx host.IHttpContext) {
	query := new(model.DatabasesQuery)
	ctx.ReadJSON(query)

	rs, err := logClientService.GetDatabases(context.Background(), query)
	if !host.HandleErr(err, ctx) {
		jsonBytes, err := json.Marshal(rs)
		if !host.HandleErr(err, ctx) {
			ctx.WriteJsonBytes(jsonBytes)
		}
	}
}
func getTables(ctx host.IHttpContext) {
	query := new(model.TablesQuery)
	ctx.ReadJSON(query)

	rs, err := logClientService.GetTables(context.Background(), query)
	if !host.HandleErr(err, ctx) {
		jsonBytes, err := json.Marshal(rs)
		if !host.HandleErr(err, ctx) {
			ctx.WriteJsonBytes(jsonBytes)
		}
	}
}
*/
