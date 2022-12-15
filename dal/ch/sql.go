package ch

const (
	_SQL_USE_DB = "use `%s`;"

	_SQL_CREATE_DB = "CREATE DATABASE IF NOT EXISTS `%s`;"

	_SQL_CREATE_TABLE = `CREATE TABLE IF NOT EXISTS ` + "`%s`" + `.` + "`%s`" + `(
	  ` + "`ID`" + ` String,
	  ` + "`TraceNo`" + ` String,
	  ` + "`User`" + ` String,
	  ` + "`Message`" + ` String,
	  ` + "`Error`" + ` String,
	  ` + "`StackTrace`" + ` String,
	  ` + "`Payload`" + ` String,
	  ` + "`Level`" + ` Int32,
	  ` + "`Flags`" + ` Int64,
	  ` + "`CreatedOnUtc`" + ` Int64
	) ENGINE = MergeTree
	  PRIMARY KEY (` + "`ID`" + `)
	  ORDER BY (` + "`ID`" + `, ` + "`CreatedOnUtc`" + `);`

	_SQL_INSERT = "INSERT INTO `%s`.`%s` (*) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"

	_SQL_SELECT_ONE = "SELECT * FROM `%s`.`%s` WHERE ID = ? LIMIT 1"
)
