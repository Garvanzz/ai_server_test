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

func TestPlayerDataKeyHelpers(t *testing.T) {
	tests := []struct {
		actID    int64
		playerID int64
	}{
		{1, 42},
		{88, 99001},
		{999, 1234567},
	}

	for _, tt := range tests {
		key := encodePlayerDataKey(tt.actID, tt.playerID)
		actID, playerID := decodePlayerDataKey(key)
		if actID != tt.actID || playerID != tt.playerID {
			t.Fatalf("roundtrip mismatch: got actId=%d playerId=%d, want actId=%d playerId=%d", actID, playerID, tt.actID, tt.playerID)
		}
	}
}

// PurgeActivityPlayerData 依赖 db.GetEngine/Redis，单测不覆盖；仅保证 key 编码与 data 包可编译。
