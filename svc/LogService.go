package svc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/dal"
	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host"
	"google.golang.org/grpc"
)

var (
	_logDAL    dal.ILogDAL
	_clientDAL dal.IClientDAL
	logRSPool  *sync.Pool
)

func init() {
	_logDAL = dal.NewLogDAL()
	_clientDAL = dal.NewClientDAL()
	logRSPool = &sync.Pool{
		New: func() any {
			return new(model.LogEntryResult)
		},
	}
}

type LogService struct{}

func (self *LogService) Write(ctx context.Context, in *model.WriteLogCommand, opts ...grpc.CallOption) (*model.LogEntryResult, error) {
	r := logRSPool.Get().(*model.LogEntryResult)
	defer func() {
		r.LogEntry = nil
		r.Message = ""
		logRSPool.Put(r)
	}()

	go func() {
		client, err := _clientDAL.GetClient(in.ClientID)
		if err != nil {
			slog.Error(err)
		} else if client == nil {
			slog.Warn(fmt.Sprintf("Client '%s' not found", in.ClientID))
		} else {
			// determine database and table
			createdOnUtc := time.UnixMilli(in.LogEntry.CreatedOnUtc)

			var dbName, tableName string
			switch client.DBPolicy {
			case 1: // By Year
				dbName = fmt.Sprintf("%s%s_%04d", core.LOG_DB_PREFIX, client.ID, createdOnUtc.Year())
				tableName = fmt.Sprintf("%02d", createdOnUtc.Month())
				break
			case 2: // By Month
				dbName = fmt.Sprintf("%s%s_%04d%02d", core.LOG_DB_PREFIX, client.ID, createdOnUtc.Year(), createdOnUtc.Month())
				tableName = fmt.Sprintf("%02d", createdOnUtc.Day())
				break
			case 3: // By Day
				dbName = fmt.Sprintf("%s%s_%04d%02d%02d", core.LOG_DB_PREFIX, client.ID, createdOnUtc.Year(), createdOnUtc.Month(), createdOnUtc.Day())
				tableName = fmt.Sprintf("%02d", createdOnUtc.Hour())
				break
			default:
				dbName = fmt.Sprintf("%s%s", core.LOG_DB_PREFIX, client.ID)
				tableName = fmt.Sprintf("%02d", createdOnUtc.Year())
			}
			// generate id
			in.LogEntry.ID = host.GenerateID()

			err = _logDAL.InsertLogEntry(dbName, tableName, in.LogEntry)
			u.LogError(err)
		}
	}()

	return r, nil
}
