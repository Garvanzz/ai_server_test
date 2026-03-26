package model

type ServerGroup struct {
	Id        int64  `json:"id" xorm:"pk autoincr"`
	Name      string `json:"name" xorm:"varchar(64) notnull"`
	SortOrder int    `json:"sortOrder" xorm:"sort_order notnull"`
	GroupType int    `json:"groupType" xorm:"group_type notnull default 0"`
	IsVisible int    `json:"isVisible" xorm:"is_visible notnull default 1"`
}

type ServerItem struct {
	Id                int64  `json:"id" xorm:"pk autoincr"`
	Channel           int    `json:"channel" xorm:"notnull"`
	GroupId           int    `json:"groupId" xorm:"group_id notnull"`
	LogicServerId     int64  `json:"logicServerId" xorm:"logic_server_id notnull default 0"`
	MergeState        int    `json:"mergeState" xorm:"merge_state notnull default 0"`
	MergeTime         int64  `json:"mergeTime" xorm:"merge_time notnull default 0"`
	Ip                string `json:"ip" xorm:"varchar(64) notnull"`
	Port              int    `json:"port" xorm:"notnull"`
	MainServerHttpUrl string `json:"mainServerHttpUrl" xorm:"main_server_http_url"` // 大厅服 HTTP 地址，GM 转发用（如 http://ip:9505）
	ServerState       int    `json:"serverState" xorm:"server_state"`               // 0：正常 1：拥挤 2：爆满 3：维护 4：未开服 5：停服
	OpenServerTime    int64  `json:"openServerTime" xorm:"open_server_time"`
	StopServerTime    int64  `json:"stopServerTime" xorm:"stop_server_time"`
	ServerName        string `json:"serverName" xorm:"server_name varchar(64)"`
}

type MergePlan struct {
	Id              int64  `json:"id" xorm:"pk autoincr"`
	Name            string `json:"name"`
	TargetServerId  int    `json:"targetServerId" xorm:"target_server_id"`
	SourceServerIds []int  `json:"sourceServerIds" xorm:"json source_server_ids"`
	Status          int    `json:"status"`
	Operator        string `json:"operator"`
	StartTime       int64  `json:"startTime" xorm:"start_time"`
	EndTime         int64  `json:"endTime" xorm:"end_time"`
	RollbackTime    int64  `json:"rollbackTime" xorm:"rollback_time"`
	Remark          string `json:"remark"`
}

type MergeServerMap struct {
	Id             int64  `json:"id" xorm:"pk autoincr"`
	PlanId         int64  `json:"planId" xorm:"plan_id"`
	SourceServerId int    `json:"sourceServerId" xorm:"source_server_id"`
	TargetServerId int    `json:"targetServerId" xorm:"target_server_id"`
	State          int    `json:"state"`
	ErrMsg         string `json:"errMsg" xorm:"err_msg"`
}

type MergeConflictLog struct {
	Id           int64  `json:"id" xorm:"pk autoincr"`
	PlanId       int64  `json:"planId" xorm:"plan_id"`
	ServerId     int    `json:"serverId" xorm:"server_id"`
	ConflictType string `json:"conflictType" xorm:"conflict_type"`
	BizKey       string `json:"bizKey" xorm:"biz_key"`
	OldValue     string `json:"oldValue" xorm:"old_value"`
	NewValue     string `json:"newValue" xorm:"new_value"`
	Resolved     int    `json:"resolved"`
	CreatedAt    int64  `json:"createdAt" xorm:"created_at"`
}

// ServerProcess 进程管理表
// server_type: 1=login_server 2=main_server 3=game_server/battle
// server_ref_id: 关联 game_server.id（login 为 0）
type ServerProcess struct {
	Id              int64  `json:"id" xorm:"pk autoincr"`
	ServerType      int    `json:"serverType" xorm:"server_type"`
	ServerRefId     int64  `json:"serverRefId" xorm:"server_ref_id"`
	ServerName      string `json:"serverName" xorm:"server_name varchar(128)"`
	ManageMode      string `json:"manageMode" xorm:"manage_mode varchar(32)"`
	ProcessBinName  string `json:"processBinName" xorm:"process_bin_name varchar(128)"`
	StartCommand    string `json:"startCommand" xorm:"start_command varchar(512)"`
	WorkDir         string `json:"workDir" xorm:"work_dir varchar(512)"`
	HttpHealthUrl   string `json:"httpHealthUrl" xorm:"http_health_url varchar(256)"`
	BuildRepoUrl    string `json:"buildRepoUrl" xorm:"build_repo_url varchar(512)"`
	BuildSourceDir  string `json:"buildSourceDir" xorm:"build_source_dir varchar(512)"`
	BuildOutputDir  string `json:"buildOutputDir" xorm:"build_output_dir varchar(512)"`
	BuildOutputName string `json:"buildOutputName" xorm:"build_output_name varchar(128)"`
	SortOrder       int    `json:"sortOrder" xorm:"sort_order"`
	Remark          string `json:"remark" xorm:"remark varchar(512)"`
}
