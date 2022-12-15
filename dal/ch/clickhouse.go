package ch

import (
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/jmoiron/sqlx"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/stask"
	"github.com/syncfuture/go/u"
)

var (
	_db        *sqlx.DB
	_wherePool = &sync.Pool{
		New: func() any {
			return new(strings.Builder)
		},
	}
	_dbLocker = new(sync.RWMutex)
	_parallel = stask.NewParallel()
)

func Init() {
	connStr := core.GrpcCP.GetString("ConnectionStrings.ClickHouse")

	var err error
	_db, err = sqlx.Connect("clickhouse", connStr)
	u.LogFatal(err)

	connMaxLifetime := core.GrpcCP.GetInt("DataAccess.ConnMaxLifetime")
	maxOpenConns := core.GrpcCP.GetInt("DataAccess.MaxOpenConns")
	maxIdleConns := core.GrpcCP.GetInt("DataAccess.MaxIdleConns")

	_db.SetConnMaxLifetime(time.Second * time.Duration(connMaxLifetime))
	_db.SetMaxOpenConns(maxOpenConns)
	_db.SetMaxIdleConns(maxIdleConns)
}

type ClickHouseDAL struct{}

func ensureDBTableExsits(err error, dbName, tableName string) error {
	var sql string
	if err != nil {
		if innerErr, ok := err.(*proto.Exception); ok {
			_dbLocker.Lock()
			defer _dbLocker.Unlock()
			if innerErr.Code == 81 { // 81: DB not exists
				// Create db
				sql = fmt.Sprintf(_SQL_CREATE_DB, dbName)
				_, err = _db.Exec(sql)
				if err != nil {
					return serr.WithStack(err)
				}

				// Create table
				sql = fmt.Sprintf(_SQL_CREATE_TABLE, dbName, tableName)
				_, err = _db.Exec(sql)
				if err != nil {
					return serr.WithStack(err)
				}
			} else if innerErr.Code == 60 { // 60: Table not exists
				// Create table only
				sql = fmt.Sprintf(_SQL_CREATE_TABLE, dbName, tableName)
				_, err = _db.Exec(sql)
				if err != nil {
					return serr.WithStack(err)
				}
			}
		} else {
			return serr.WithStack(err)
		}
	}

	return nil
}

func (self *ClickHouseDAL) InsertLogEntry(dbName, tableName string, logEntry *model.LogEntry) error {
	sqlStr := fmt.Sprintf(_SQL_INSERT, dbName, tableName)
	_dbLocker.RLock()
	_, err := _db.Exec(sqlStr,
		logEntry.ID,
		logEntry.TraceNo,
		logEntry.User,
		logEntry.Message,
		logEntry.Error,
		logEntry.StackTrace,
		logEntry.Payload,
		int32(logEntry.Level),
		logEntry.Flags,
		logEntry.CreatedOnUtc,
	)
	_dbLocker.RUnlock()
	if err != nil {
		err = ensureDBTableExsits(err, dbName, tableName) // Ensure db and table are exist
		if err != nil {
			return serr.WithStack(err)
		}

		// Retry
		_, err = _db.Exec(sqlStr,
			logEntry.ID,
			logEntry.TraceNo,
			logEntry.User,
			logEntry.Message,
			logEntry.Error,
			logEntry.StackTrace,
			logEntry.Payload,
			int32(logEntry.Level),
			logEntry.Flags,
			logEntry.CreatedOnUtc,
		)

		if err != nil {
			return serr.WithStack(err)
		}
	}

	return nil
}
func (self *ClickHouseDAL) GetLogEntry(query *model.LogEntryQuery) (*model.LogEntry, error) {
	r := new(model.LogEntry)

	sqlSel := fmt.Sprintf(_SQL_SELECT_ONE, query.DBName, query.TableName)

	_dbLocker.RLock()
	err := _db.Get(r, sqlSel, query.ID)
	_dbLocker.RUnlock()

	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
func (self *ClickHouseDAL) GetLogEntries(query *model.LogEntriesQuery) ([]*model.LogEntry, int64, error) {
	return nil, 0, nil
}
func (self *ClickHouseDAL) GetDatabases(clientID string) ([]string, error) {
	return nil, nil
}
func (self *ClickHouseDAL) GetTables(database string) ([]string, error) {
	return nil, nil
}
