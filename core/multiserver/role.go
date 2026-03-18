package multiserver

import (
	"encoding/json"
	"fmt"

	"xfx/core/define"
)

type LoginTokenPayload struct {
	UID           string `json:"uid"`
	AccountID     int64  `json:"accountId"`
	RoleID        int64  `json:"roleId"`
	EntryServerID int    `json:"entryServerId"`
	LogicServerID int    `json:"logicServerId"`
	PlayerID      int64  `json:"playerId"`
	IssuedAt      int64  `json:"issuedAt"`
}

func EncodeLoginTokenPayload(payload LoginTokenPayload) string {
	b, err := json.Marshal(payload)
	if err != nil {
		return payload.UID
	}
	return string(b)
}

func DecodeLoginTokenPayload(raw string) LoginTokenPayload {
	if raw == "" {
		return LoginTokenPayload{}
	}
	var payload LoginTokenPayload
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && payload.UID != "" {
		return payload
	}
	return LoginTokenPayload{UID: raw}
}

func AccountRoleRedisKey(uid string, entryServerID int) string {
	return fmt.Sprintf("%s:%s:%d", define.AccountRole, uid, entryServerID)
}
