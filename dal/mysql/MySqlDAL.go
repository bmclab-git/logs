package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/sconv"
	"github.com/syncfuture/go/sdto"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/stask"
	"github.com/syncfuture/go/u"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// const (
// 	_TIME_FORMAT = "2006-01-02T15:04:05Z"
// )

var (
	// _clientsMap map[string]*model.LogClient
	_wherePool = &sync.Pool{
		New: func() any {
			return new(strings.Builder)
		},
	}
	// _cacheLocker = new(sync.RWMutex)
	_dbLocker = new(sync.RWMutex)
	_parallel = stask.NewParallel()
	_db       *sqlx.DB
)

func Init() {
	connStr := core.GrpcCP.GetString("ConnectionStrings.MySql")
	var err error
	_db, err = sqlx.Connect("mysql", connStr)
	u.LogFatal(err)

	connMaxLifetime := core.GrpcCP.GetInt("DataAccess.ConnMaxLifetime")
	maxOpenConns := core.GrpcCP.GetInt("DataAccess.MaxOpenConns")
	maxIdleConns := core.GrpcCP.GetInt("DataAccess.MaxIdleConns")

	_db.SetConnMaxLifetime(time.Second * time.Duration(connMaxLifetime))
	_db.SetMaxOpenConns(maxOpenConns)
	_db.SetMaxIdleConns(maxIdleConns)

	// err = refreshCache()
	// u.LogFatal(err)
}

type MySqlDAL struct {
}

// ************************************************************************************************

// func refreshCache() error {
// 	var clients []*model.LogClient
// 	err := _db.Select(&clients, "SELECT * FROM `Clients`")
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	_clientsMap = make(map[string]*model.LogClient, len(clients))
// 	for _, x := range clients {
// 		_clientsMap[x.ID] = x
// 	}

// 	return nil
// }

// func (self *MySqlDAL) InsertClient(client *model.LogClient) error {
// 	_, err := _db.NamedExec("INSERT INTO `Clients`(`ID`,`DBPolicy`) VALUES(:ID, :DBPolicy)", client)
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	_clientsMap[client.ID] = client
// 	return nil
// }
// func (self *MySqlDAL) GetClient(id string) (r *model.LogClient, err error) {
// 	var ok bool
// 	_cacheLocker.RLock()
// 	r, ok = _clientsMap[id]
// 	_cacheLocker.RUnlock()
// 	if ok {
// 		return r, nil
// 	}

// 	r = new(model.LogClient)
// 	err = _db.Get(r, "SELECT * FROM `Clients` WHERE ID = ?", id)
// 	if err != nil && !errors.Is(err, sql.ErrNoRows) {
// 		return nil, serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	_clientsMap[id] = r
// 	_cacheLocker.Unlock()

// 	return r, nil
// }
// func (self *MySqlDAL) UpdateClient(client *model.LogClient) error {
// 	_, err := _db.Exec("UPDATE `Clients` SET `DBPolicy` = ?", client.DBPolicy)
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	_clientsMap[client.ID] = client

// 	return nil
// }
// func (self *MySqlDAL) DeleteClient(id string) error {
// 	_, err := _db.Exec("DELETE FROM `Clients` WHERE `ID` = ?", id)
// 	if err != nil {
// 		return serr.WithStack(err)
// 	}

// 	_cacheLocker.Lock()
// 	defer _cacheLocker.Unlock()
// 	delete(_clientsMap, id)

// 	return nil
// }
// func (self *MySqlDAL) GetClients(query *model.LogClientsQuery) ([]*model.LogClient, error) {
// 	listSel := "SELECT * FROM `Clients`"

// 	var r []*model.LogClient
// 	err := _db.Select(&r, listSel)
// 	if err != nil {
// 		return nil, serr.WithStack(err)
// 	}

// 	return r, nil
// }
// func (self *MySqlDAL) RefreshCache() error {
// 	return nil
// }

// ************************************************************************************************

func ensureDBTableExsits(err error, dbName, tableName string) error {
	var sql string
	if err != nil {
		if innerErr, ok := err.(*mysql.MySQLError); ok && innerErr.Number == 1146 { // 1146: Table not exists
			_dbLocker.Lock()
			defer _dbLocker.Unlock()
			// Select db
			sql = fmt.Sprintf(_SQL_USE_DB, dbName)
			_, err = _db.Exec(sql)
			if err != nil {
				if innerErr, ok := err.(*mysql.MySQLError); ok && innerErr.Number == 1049 { // 1049: Unknown db
					// Create db
					sql = fmt.Sprintf(_SQL_CREATE_DB, dbName)
					_, err = _db.Exec(sql)
					if err != nil {
						return serr.WithStack(err)
					}
					// Create count SP
					sql = fmt.Sprintf(_SQL_SP_COUNT, dbName)
					_, err = _db.Exec(sql)
					if err != nil {
						return serr.WithStack(err)
					}
					// Create page SP
					sql = fmt.Sprintf(_SQL_SP_PAGE, dbName)
					_, err = _db.Exec(sql)
					if err != nil {
						return serr.WithStack(err)
					}
				} else {
					return serr.WithStack(err)
				}
			}

			// Create table
			sql = fmt.Sprintf(_SQL_CREATE_TABLE, dbName, tableName, tableName, tableName, tableName, tableName, tableName, tableName)
			_, err = _db.Exec(sql)
			if err != nil {
				return serr.WithStack(err)
			}
		} else {
			return serr.WithStack(err)
		}
	}

	return nil
}

func (self *MySqlDAL) InsertLogEntry(dbName, tableName string, logEntry *model.LogEntry) error {
	sql := fmt.Sprintf(_SQL_INSERT, dbName, tableName)
	_dbLocker.RLock()
	_, err := _db.NamedExec(sql, logEntry)
	_dbLocker.RUnlock()

	if err != nil {
		err = ensureDBTableExsits(err, dbName, tableName) // Ensure db and table are exist
		if err == nil {
			// No error, retry
			_, err := _db.NamedExec(sql, logEntry)
			if err != nil {
				return serr.WithStack(err)
			}

		} else {
			return err
		}
	}

	return nil
}

func (self *MySqlDAL) GetLogEntry(query *model.LogEntryQuery) (*model.LogEntry, error) {
	_dbLocker.RLock()
	defer _dbLocker.RUnlock()

	r := new(model.LogEntry)
	sqlSel := fmt.Sprintf(_SQL_SELECT_ONE, query.DBName, query.TableName)
	err := _db.Get(r, sqlSel, query.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, serr.WithStack(err)
	}

	return r, nil
}

func (self *MySqlDAL) GetLogEntries(query *model.LogEntriesQuery) ([]*model.LogEntry, int64, error) {
	// Build where
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
		slog.Debug(t.UnixMilli(), ", ", t.Format(time.RFC3339))
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

	// slog.Debug(where.String())

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
			countSql := fmt.Sprintf("CALL `%s`.`SYSSP_GetTotalCount` (?,?)", query.DBName)
			table := fmt.Sprintf("`%s`.`%s`", query.DBName, query.TableName)
			chr.Error = _db.Get(&totalCount, countSql, table, where.String())
			chr.Error = serr.WithStack(chr.Error)
			chr.Result = totalCount
		},
		func(ch chan *sdto.ChannelResultDTO) {
			chr := new(sdto.ChannelResultDTO)
			defer func() {
				ch <- chr
			}()

			listSql := fmt.Sprintf("CALL `%s`.`SYSSP_GetPagedData` (?,?,?,?,?,?)", query.DBName)
			table := fmt.Sprintf("`%s`.`%s`", query.DBName, query.TableName)
			ord := "`CreatedOnUtc` DESC"

			var r []*model.LogEntry
			chr.Error = _db.Select(&r, listSql, query.PageSize, query.PageIndex, table, ord, "*", where.String())
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

func (self *MySqlDAL) GetDatabases(clientID string) ([]string, error) {
	sqlStr := "SELECT `schema_name` FROM information_schema.schemata WHERE SCHEMA_NAME LIKE ? ORDER BY `schema_name` DESC;"

	var r []string
	keyword := "LOG\\_" + clientID + "%"
	err := _db.Select(&r, sqlStr, keyword)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}

func (self *MySqlDAL) GetTables(database string) ([]string, error) {
	sqlStr := "SELECT `table_name` FROM information_schema.tables WHERE table_schema = ? AND table_type = 'base table' ORDER BY `table_name` DESC;"

	var r []string
	err := _db.Select(&r, sqlStr, database)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
