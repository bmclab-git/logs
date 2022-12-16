package ch

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/jmoiron/sqlx"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/sdto"
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
	where := _wherePool.Get().(*strings.Builder)
	defer func() {
		where.Reset()
		_wherePool.Put(where)
	}()

	if query.StartTime != "" {
		t, err := time.ParseInLocation(time.RFC3339, query.StartTime, time.UTC)
		if err != nil {
			return nil, 0, serr.WithStack(err)
		}
		where.WriteString(" AND `CreatedOnUtc` >= " + sconv.ToString(t.UnixMilli()))
	}
	if query.EndTime != "" {
		t, err := time.ParseInLocation(time.RFC3339, query.EndTime, time.UTC)
		t = t.Add(time.Hour * 24)
		if err != nil {
			return nil, 0, serr.WithStack(err)
		}
		where.WriteString(" AND `CreatedOnUtc` <= " + sconv.ToString(t.UnixMilli()))
		// slog.Debug(t.UnixMilli(), ", ", t.Format(time.RFC3339))
	}
	if query.Level >= 0 {
		where.WriteString(" AND `Level` = " + strconv.FormatInt(int64(query.Level), 10))
	}

	// TODO: Prevent sql injection
	if query.User != "" {
		likeSql := " AND `User` LIKE '"
		if query.Flags&1 == 1 { // Has flag, do left & right fuzzy search, other wise, only do right fuzzy search
			likeSql += "%"
		}
		likeSql += query.User + "%'"
		where.WriteString(likeSql)
	}
	if query.TraceNo != "" {
		likeSql := " AND `TraceNo` LIKE '"
		if query.Flags&1 == 1 { // Has flag, do left & right fuzzy search, other wise, only do right fuzzy search
			likeSql += "%"
		}
		likeSql += query.TraceNo + "%'"
		where.WriteString(likeSql)
	}
	if query.Message != "" {
		likeSql := " AND `Message` LIKE '"
		if query.Flags&1 == 1 { // Has flag, do left & right fuzzy search, other wise, only do right fuzzy search
			likeSql += "%"
		}
		likeSql += query.Message + "%'"
		where.WriteString(likeSql)
	}

	// where.WriteString(fmt.Sprintf("%s ORDER BY `CreatedOnUtc` DESC LIMIT %d, %d", selectSql, start, end))

	_dbLocker.RLock()
	defer _dbLocker.RUnlock()

	// now := time.Now()
	chrs := _parallel.Invoke(
		func(ch chan *sdto.ChannelResultDTO) {
			chr := new(sdto.ChannelResultDTO)
			defer func() {
				ch <- chr
			}()
			var totalCount int64
			// selectSql := fmt.Sprintf("SELECT * FROM `%s`.`%s` WHERE 0 = 0", query.DBName, query.TableName)
			countSql := fmt.Sprintf("SELECT COUNT(0) FROM `%s`.`%s` WHERE 0 = 0 %s", query.DBName, query.TableName, where.String())
			chr.Error = _db.Get(&totalCount, countSql)
			chr.Error = serr.WithStack(chr.Error)
			chr.Result = totalCount
		},
		func(ch chan *sdto.ChannelResultDTO) {
			chr := new(sdto.ChannelResultDTO)
			defer func() {
				ch <- chr
			}()

			start := (query.PageIndex - 1) * query.PageSize
			end := start + query.PageSize

			listSql := fmt.Sprintf("SELECT * FROM `%s`.`%s` WHERE 0 = 0 %s ORDER BY `CreatedOnUtc` DESC LIMIT %d, %d", query.DBName, query.TableName, where.String(), start, end)

			var r []*model.LogEntry
			chr.Error = _db.Select(&r, listSql)
			chr.Error = serr.WithStack(chr.Error)
			chr.Result = r
		},
	)
	// elapsed := time.Since(now)
	// slog.Debugf("GetLogEntries: %d ms", elapsed.Milliseconds())

	err := u.JointErrors(chrs[0].Error, chrs[1].Error)
	if err != nil {
		return nil, 0, err
	}

	totalCount := chrs[0].Result.(int64)
	list := chrs[1].Result.([]*model.LogEntry)
	return list, totalCount, nil
}
func (self *ClickHouseDAL) GetDatabases(clientID string) ([]string, error) {
	var r []string
	keyword := "LOG\\_" + clientID + "%"
	err := _db.Select(&r, _SQL_SELECT_DATABASES, keyword)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
func (self *ClickHouseDAL) GetTables(database string) ([]string, error) {
	var r []string
	err := _db.Select(&r, _SQL_SELECT_TABLES, database)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
