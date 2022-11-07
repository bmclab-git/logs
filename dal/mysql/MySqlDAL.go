package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/Lukiya/logs/core"
	"github.com/Lukiya/logs/model"
	"github.com/syncfuture/go/sdto"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/stask"
	"github.com/syncfuture/go/u"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	_clients     map[string]*model.LogClient
	_cacheLocker = new(sync.RWMutex)
	_parallel    = stask.NewParallel()
	_db          *sqlx.DB
)

func init() {
	connStr := core.MainCP.GetString("ConnectionStrings.MySql")
	var err error
	_db, err = sqlx.Connect("mysql", connStr)
	u.LogFatal(err)

	err = refreshCache()
	u.LogFatal(err)
}

type MySqlDAL struct {
}

// ************************************************************************************************

func refreshCache() error {
	var clients []*model.LogClient
	err := _db.Select(&clients, "SELECT * FROM `Clients`")
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clients = make(map[string]*model.LogClient, len(clients))
	for _, x := range clients {
		_clients[x.ID] = x
	}

	return nil
}

func (self *MySqlDAL) InsertClient(client *model.LogClient) error {
	_, err := _db.NamedExec("INSERT INTO `Clients`(`ID`,`DBPolicy`) VALUES(:ID, :DBPolicy)", client)
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clients[client.ID] = client
	return nil
}
func (self *MySqlDAL) GetClient(id string) (r *model.LogClient, err error) {
	var ok bool
	_cacheLocker.RLock()
	r, ok = _clients[id]
	_cacheLocker.RUnlock()
	if ok {
		return r, nil
	}

	r = new(model.LogClient)
	err = _db.Get(r, "SELECT * FROM `Clients` WHERE ID = ?", id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, serr.WithStack(err)
	}

	_cacheLocker.Lock()
	_clients[id] = r
	_cacheLocker.Unlock()

	return nil, nil
}
func (self *MySqlDAL) UpdateClient(client *model.LogClient) error {
	_, err := _db.Exec("UPDATE `Clients` SET `DBPolicy` = ?", client.DBPolicy)
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	_clients[client.ID] = client

	return nil
}
func (self *MySqlDAL) DeleteClient(id string) error {
	_, err := _db.Exec("DELETE FROM `Clients` WHERE `ID` = ?", id)
	if err != nil {
		return serr.WithStack(err)
	}

	_cacheLocker.Lock()
	defer _cacheLocker.Unlock()
	delete(_clients, id)

	return nil
}
func (self *MySqlDAL) GetClients(query *model.LogClientsQuery) ([]*model.LogClient, int64, error) {
	listSel := "SELECT * FROM `Clients`"
	countSel := "SELECT COUNT(0) FROM `Clients`"
	whe := ""

	// TODO:
	// if query.Keyword != "" {
	// 	whe=" WHERE `ID` = "
	// }

	ord := " ORDER BY "
	switch query.OrderBy {
	case 1:
		ord += "`DBPolicy` "
		break
	default:
		ord += "`ID` "
		break
	}
	if query.OrderDir == "" {
		query.OrderDir = "ASC"
	}
	ord += query.OrderDir

	start := (query.PageIndex - 1) * query.PageSize
	end := start + query.PageSize
	lim := fmt.Sprintf(" LIMIT %d, %d", start, end)

	listSQL := listSel + whe + ord + lim
	countSQL := countSel + whe

	chrs := _parallel.Invoke(
		func(ch chan *sdto.ChannelResultDTO) {
			chr := &sdto.ChannelResultDTO{Result: 0}
			defer func() {
				ch <- chr
			}()

			var r int64
			chr.Error = _db.Get(&r, countSQL)
			chr.Result = r
		},
		func(ch chan *sdto.ChannelResultDTO) {
			chr := new(sdto.ChannelResultDTO)
			defer func() {
				ch <- chr
			}()

			var r []*model.LogClient
			chr.Error = _db.Select(&r, listSQL)
			chr.Result = r
		},
	)

	err := u.JointErrors(chrs[0].Error, chrs[1].Error)
	if err != nil {
		return nil, 0, err
	}

	totalCount := chrs[0].Result.(int64)
	list := chrs[1].Result.([]*model.LogClient)
	return list, totalCount, nil
}
func (self *MySqlDAL) RefreshCache() error {
	return nil
}

// ************************************************************************************************

func ensureDBTableExsits(err error, dbName, tableName string) error {
	var sql string
	if err != nil {
		if innerErr, ok := err.(*mysql.MySQLError); ok && innerErr.Number == 1146 { // 1146: Table not exists
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

func (self *MySqlDAL) GetDatabases(clientID string) ([]string, error) {
	return nil, nil
}

func (self *MySqlDAL) InsertLogEntry(dbName, tableName string, logEntry *model.LogEntry) error {
	sql := fmt.Sprintf(_SQL_INSERT, dbName, tableName)
	_, err := _db.NamedExec(sql, logEntry)
	if err != nil {
		err = ensureDBTableExsits(err, dbName, tableName) // Ensure db and table are exist
		if err == nil {
			// No error, retry
			_, err = _db.NamedExec(sql, logEntry)
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
	r := new(model.LogEntry)
	sqlSel := fmt.Sprintf(_SQL_SELECT_ONE, query.DBName, query.TableName)
	err := _db.Get(r, sqlSel, query.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, serr.WithStack(err)
	}

	return r, nil
}

func (self *MySqlDAL) GetLogEntries(query *model.LogEntriesQuery) ([]*model.LogEntry, int64, error) {
	listSel := fmt.Sprintf("SELECT * FROM `%s`.`%s`", query.DBName, query.TableName)
	countSel := fmt.Sprintf("SELECT COUNT(0) FROM `%s`.`%s`", query.DBName, query.TableName)
	whe := ""

	// TODO:
	// if query.Keyword != "" {
	// 	whe=" WHERE `ID` = "
	// }

	ord := " ORDER BY `CreatedOnUtc` DESC"

	start := (query.PageIndex - 1) * query.PageSize
	end := start + query.PageSize
	lim := fmt.Sprintf(" LIMIT %d, %d", start, end)

	listSQL := listSel + whe + ord + lim
	countSQL := countSel + whe

	chrs := _parallel.Invoke(
		func(ch chan *sdto.ChannelResultDTO) {
			chr := &sdto.ChannelResultDTO{Result: 0}
			defer func() {
				ch <- chr
			}()

			var r int64
			chr.Error = _db.Get(&r, countSQL)
			chr.Result = r
		},
		func(ch chan *sdto.ChannelResultDTO) {
			chr := new(sdto.ChannelResultDTO)
			defer func() {
				ch <- chr
			}()

			var r []*model.LogEntry
			chr.Error = _db.Select(&r, listSQL)
			chr.Result = r
		},
	)

	err := u.JointErrors(chrs[0].Error, chrs[1].Error)
	if err != nil {
		return nil, 0, err
	}

	totalCount := chrs[0].Result.(int64)
	list := chrs[1].Result.([]*model.LogEntry)

	return list, totalCount, nil
}
