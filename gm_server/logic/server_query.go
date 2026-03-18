package logic

import (
	"fmt"
	"strings"

	"xorm.io/xorm"
)

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
	"exe_name",
	"exe_path",
}

func isMissingGameServerRuntimeColumn(err error) bool {
	if err == nil {
		return false
	}
	text := err.Error()
	if !strings.Contains(text, "Unknown column") {
		return false
	}
	for _, column := range []string{"manage_mode", "process_name", "start_command", "work_dir"} {
		if strings.Contains(text, fmt.Sprintf("'%s'", column)) {
			return true
		}
	}
	return false
}

func retryLegacyGameServerFind(primary func() error, fallback func() error) error {
	err := primary()
	if isMissingGameServerRuntimeColumn(err) {
		return fallback()
	}
	return err
}

func retryLegacyGameServerGet(primary func() (bool, error), fallback func() (bool, error)) (bool, error) {
	has, err := primary()
	if isMissingGameServerRuntimeColumn(err) {
		return fallback()
	}
	return has, err
}

func applyLegacyGameServerCols(session *xorm.Session) *xorm.Session {
	return session.Cols(legacyGameServerColumns...)
}
