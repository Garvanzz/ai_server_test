package impl

import (
	"testing"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"

	"github.com/golang/protobuf/proto"
)

func TestMain(m *testing.M) {
	log.DefaultInit()
	m.Run()
}

type mockActivityConfig struct {
	activityID int64
	value      int
}

func (m mockActivityConfig) GetActivityId() int64 { return m.activityID }

func TestFindTypedConf(t *testing.T) {
	confs := map[int64]mockActivityConfig{
		1: {activityID: 100, value: 1},
		2: {activityID: 200, value: 2},
		3: {activityID: 100, value: 3},
	}

	conf, ok := FindTypedConf(100, confs, func(c mockActivityConfig) bool {
		return c.value == 3
	})
	if !ok {
		t.Fatal("expected matching config")
	}
	if conf.value != 3 {
		t.Fatalf("expected value 3, got %d", conf.value)
	}

	if _, ok = FindTypedConf(999, confs, nil); ok {
		t.Fatal("expected no config for unknown activity id")
	}
}

func TestKeyTypeMismatch(t *testing.T) {
	params := EventParams{"score": "bad-type"}
	if _, ok := Key[int32](params, "score"); ok {
		t.Fatal("expected type mismatch to fail")
	}
}

func TestSetProtoByTypeNilPayload(t *testing.T) {
	RegisterActivity("test_framework_activity", &ActivityDesc{
		NewHandler: func() IActivity { return &BaseActivity{} },
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.ConfigId = 1
		},
	})

	msg := new(proto_activity.ActivityData)
	SetProtoByType("test_framework_activity", msg, nil)
	if msg.ConfigId != 0 {
		t.Fatalf("expected nil payload to skip SetProto, got ConfigId=%d", msg.ConfigId)
	}
}
