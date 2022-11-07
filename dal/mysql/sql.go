package mysql

var (
	_SQL_USE_DB       = "use `%s`;"
	_SQL_CREATE_DB    = "CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = utf8mb4 REPLICA_NUM = 1;"
	_SQL_CREATE_TABLE = `CREATE TABLE ` + "`%s`" + `.` + "`%s`" + `(
	  ` + "`ID`" + ` varchar(15) NOT NULL,
	  ` + "`TraceNo`" + ` varchar(50) DEFAULT NULL,
	  ` + "`User`" + ` varchar(100) DEFAULT NULL,
	  ` + "`Message`" + ` varchar(500) DEFAULT NULL,
	  ` + "`Error`" + ` varchar(500) DEFAULT NULL,
	  ` + "`StackTrace`" + ` text DEFAULT NULL,
	  ` + "`Payload`" + ` text DEFAULT NULL,
	  ` + "`Level`" + ` int(11) NOT NULL,
	  ` + "`CreatedOnUtc`" + ` bigint(20) NOT NULL,
	  PRIMARY KEY (` + "`ID`" + `),
	  KEY ` + "`%s_TraceNO_IDX`" + ` (` + "`TraceNo`" + `) BLOCK_SIZE 16384 LOCAL,
	  KEY ` + "`%s_User_IDX`" + ` (` + "`User`" + `) BLOCK_SIZE 16384 LOCAL,
	  KEY ` + "`%s_Message_IDX`" + ` (` + "`Message`" + `) BLOCK_SIZE 16384 LOCAL,
	  KEY ` + "`%s_Error_IDX`" + ` (` + "`Error`" + `) BLOCK_SIZE 16384 LOCAL,
	  KEY ` + "`%s_Level_IDX`" + ` (` + "`Level`" + `) BLOCK_SIZE 16384 LOCAL,
	  KEY ` + "`%s_CreatedOnUtc_IDX`" + ` (` + "`CreatedOnUtc`" + `) BLOCK_SIZE 16384 LOCAL
	) DEFAULT CHARSET = utf8mb4 ROW_FORMAT = DYNAMIC COMPRESSION = 'zstd_1.3.8' REPLICA_NUM = 1 BLOCK_SIZE = 16384 USE_BLOOM_FILTER = FALSE TABLET_SIZE = 134217728 PCTFREE = 0;`
	_SQL_INSERT = `INSERT INTO ` + "`%s`" + `.` + "`%s`" + `
	(ID, TraceNo, ` + "`User`" + `, Message, Error, StackTrace, Payload, ` + "`Level`" + `, CreatedOnUtc)
	VALUES(:ID, :TraceNo, :User, :Message, :Error, :StackTrace, :Payload, :Level, :CreatedOnUtc);`
	_SQL_SELECT_ONE = "SELECT * FROM `%s`.`%s` WHERE ID = ? LIMIT 1"
)
