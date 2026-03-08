package data

import (
	"testing"
	"xfx/core/define"
)

func TestPlayerDataKeyEncoding(t *testing.T) {
	base := int64(define.ActivityPlayerDataBase)
	// key = actId*Base + playerId  =>  actId = key/Base
	tests := []struct {
		actId    int64
		playerId int64
	}{
		{1, 1},
		{1, 999},
		{2, 100},
		{100, 50000},
	}
	for _, tt := range tests {
		key := tt.actId*base + tt.playerId
		gotActId := key / base
		gotPlayerId := key % base
		if gotActId != tt.actId || gotPlayerId != tt.playerId {
			t.Errorf("actId=%d playerId=%d: key=%d => actId=%d playerId=%d", tt.actId, tt.playerId, key, gotActId, gotPlayerId)
		}
	}
}

// PurgeActivityPlayerData 依赖 db.GetEngine/Redis，单测不覆盖；仅保证 key 编码与 data 包可编译。
