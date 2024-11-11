package mysql

const (
	OKPacket            = 0x00
	ComQuit             = 0x01
	ComInitDB           = 0x02
	ComQuery            = 0x03
	ComFieldList        = 0x04
	ComCreateDB         = 0x05
	ComDropDB           = 0x06
	ComRefresh          = 0x07
	ComShutdown         = 0x08
	ComStatistics       = 0x09
	ComProcessInfo      = 0x0A
	ComConnect          = 0x0B
	ComProcessKill      = 0x0C
	ComDebug            = 0x0D
	ComPing             = 0x0E
	ComTime             = 0x0F
	ComDelayedInsert    = 0x10
	ComChangeUser       = 0x11
	ComBinlogDump       = 0x12
	ComTableDump        = 0x13
	ComConnectOut       = 0x14
	ComRegisterSlave    = 0x15
	ComStmtPrepare      = 0x16
	ComStmtExecute      = 0x17
	ComStmtSendLongData = 0x18
	ComStmtClose        = 0x19
	ComStmtReset        = 0x1A
	ComSetOption        = 0x1B
	ComStmtFetch        = 0x1C
	ComDaemon           = 0x1D
	ComBinlogDumpGTID   = 0x1E
	ComResetConnection  = 0x1F
)

var commandNames = map[byte]string{
	OKPacket:            "OK",
	ComQuit:             "ComQuit",
	ComInitDB:           "ComInitDB",
	ComQuery:            "ComQuery",
	ComFieldList:        "ComFieldList",
	ComCreateDB:         "ComCreateDB",
	ComDropDB:           "ComDropDB",
	ComRefresh:          "ComRefresh",
	ComShutdown:         "ComShutdown",
	ComStatistics:       "ComStatistics",
	ComProcessInfo:      "ComProcessInfo",
	ComProcessKill:      "ComProcessKill",
	ComDebug:            "ComDebug",
	ComPing:             "ComPing",
	ComTime:             "ComTime",
	ComDelayedInsert:    "ComDelayedInsert",
	ComChangeUser:       "ComChangeUser",
	ComBinlogDump:       "ComBinlogDump",
	ComTableDump:        "ComTableDump",
	ComConnectOut:       "ComConnectOut",
	ComRegisterSlave:    "ComRegisterSlave",
	ComStmtPrepare:      "ComStmtPrepare",
	ComStmtExecute:      "ComStmtExecute",
	ComStmtSendLongData: "ComStmtSendLongData",
	ComStmtClose:        "ComStmtClose",
	ComStmtReset:        "ComStmtReset",
	ComSetOption:        "ComSetOption",
	ComStmtFetch:        "ComStmtFetch",
	ComDaemon:           "ComDaemon",
	ComBinlogDumpGTID:   "ComBinlogDumpGTID",
	ComResetConnection:  "ComResetConnection",
}
