package logic

import (
	"xorm.io/xorm"
)

// legacyGameServerColumns lists only the stable non-process columns, used when an
// older xorm driver needs an explicit column list. Since process fields were removed
// from game_server in migration 006, this list is now the full column set.
var legacyGameServerColumns = []string{
	"id",
	"channel",
	"group_id",
	"logic_server_id",
	"merge_state",
	"merge_time",
	"ip",
	"port",
	"main_server_http_url",
	"server_state",
	"open_server_time",
	"stop_server_time",
	"server_name",
}

func applyLegacyGameServerCols(session *xorm.Session) *xorm.Session {
	return session.Cols(legacyGameServerColumns...)
}
