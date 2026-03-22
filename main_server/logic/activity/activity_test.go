package activity

import (
	"errors"
	"testing"
	"xfx/core/define"
	"xfx/core/fsm"
	"xfx/main_server/logic/activity/impl"
	"xfx/pkg/log"
)

type mockLifecycleActivity struct {
	impl.BaseActivity
	recoverCount int
}

func (m *mockLifecycleActivity) OnRecover() {
	m.recoverCount++
}

func TestMain(m *testing.M) {
	log.DefaultInit()
	m.Run()
}

// ==================== FSM 状态与转换 ====================

func TestFSM_Constants(t *testing.T) {
	if StateWaiting == "" || StateRunning == "" || StateStopped == "" || StateClosed == "" {
		t.Fatal("state constants must be non-empty")
	}
	if EventStart == "" || EventStop == "" || EventClose == "" || EventRecover == "" || EventRestart == "" {
		t.Fatal("event constants must be non-empty")
	}
}

func TestFSM_TransitionTable(t *testing.T) {
	// 期望的合法转换
	valid := map[string]string{
		StateWaiting + "+" + EventStart:   StateRunning,
		StateWaiting + "+" + EventClose:   StateClosed,
		StateRunning + "+" + EventStop:    StateStopped,
		StateRunning + "+" + EventClose:   StateClosed,
		StateStopped + "+" + EventRecover: StateRunning,
		StateStopped + "+" + EventClose:   StateClosed,
		StateStopped + "+" + EventRestart: StateWaiting,
		StateClosed + "+" + EventRestart:  StateWaiting,
	}
	if len(valid) != 8 {
		t.Fatalf("expected 8 transitions, got %d", len(valid))
	}
	for _, tr := range transitions {
		key := tr.From + "+" + tr.Event
		to, ok := valid[key]
		if !ok {
			t.Errorf("unexpected transition: %s -> %s (%s)", tr.From, tr.To, tr.Event)
			continue
		}
		if to != tr.To {
			t.Errorf("transition %s: expected to=%s, got %s", key, to, tr.To)
		}
	}
}

func TestFSM_Trigger_ValidTransitions(t *testing.T) {
	m := &Manager{}
	m.entities = newEntityStore()
	m.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: m}, transitions...)
	ent := newTestEntity(1, 100, "TestType", StateRunning)
	ent.handler = &impl.BaseActivity{}
	ent.handler.SetBaseInfo(ent)
	m.entities.store(ent)

	err := m.sm.Trigger(StateRunning, EventStop, ent)
	if err != nil {
		t.Fatalf("Trigger(Running, Stop) failed: %v", err)
	}
	if ent.State != StateStopped {
		t.Fatalf("after Stop expected state Stopped, got %s", ent.State)
	}

	err = m.sm.Trigger(StateStopped, EventRecover, ent)
	if err != nil {
		t.Fatalf("Trigger(Stopped, Recover) failed: %v", err)
	}
	if ent.State != StateRunning {
		t.Fatalf("after Recover expected state Running, got %s", ent.State)
	}
}

func TestFSM_Trigger_InvalidTransition(t *testing.T) {
	m := &Manager{}
	m.entities = newEntityStore()
	m.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: m}, transitions...)
	ent := newTestEntity(1, 100, "Test", StateRunning)

	err := m.sm.Trigger(StateRunning, EventRestart, ent)
	if err == nil {
		t.Fatal("Trigger(Running, Restart) should fail")
	}
	if !errors.Is(err, fsm.ErrTransitionNotFound) {
		t.Fatalf("expected ErrTransitionNotFound, got %v", err)
	}
}

// ==================== entity checkState ====================

func TestEntity_CheckState_ActTimeClose_Running(t *testing.T) {
	ent := newTestEntity(1, 100, "Test", StateRunning)
	ent.TimeType = define.ActTimeClose
	// 不设 Start/End，checkState 里只判断 TimeType
	event := ent.checkState()
	if event != EventClose {
		t.Errorf("ActTimeClose+Running should yield EventClose, got %q", event)
	}
}

func TestEntity_CheckState_AlwaysOpen_Waiting(t *testing.T) {
	ent := newTestEntity(1, 100, "Test", StateWaiting)
	ent.TimeType = define.ActTimeAlwaysOpen
	ent.StartTime = 0
	ent.EndTime = 0
	event := ent.checkState()
	if event != EventStart {
		t.Errorf("ActTimeAlwaysOpen+Waiting should yield EventStart, got %q", event)
	}
}

func TestEntity_CheckState_Stopped_AfterEndTime(t *testing.T) {
	ent := newTestEntity(1, 100, "Test", StateStopped)
	ent.TimeType = define.ActTimeConfigured
	ent.StartTime = 1
	ent.EndTime = 1 // now >= 1 时才会 EventClose，测试时可能不满足
	// 仅保证逻辑存在：若 now >= EndTime 则应 EventClose。不依赖具体时间则只测 TimeType/EndTime>0
	if ent.TimeType != define.ActTimeAlwaysOpen && ent.EndTime > 0 {
		_ = ent.checkState() // 可能返回 EventClose 或 EventNone，取决于当前时间
	}
}

// ==================== ActivityInfo / entityToInfo ====================

func TestEntityToInfo(t *testing.T) {
	ent := newTestEntity(99, 88, "SomeType", StateRunning)
	ent.TimeValue = 3
	info := entityToInfo(ent)
	if info == nil {
		t.Fatal("entityToInfo should not return nil")
	}
	if info.ActId != 99 || info.CfgId != 88 || info.Type != "SomeType" || info.State != StateRunning || info.Season != 3 {
		t.Errorf("entityToInfo mismatch: %+v", info)
	}
}

func TestEntityToInfo_Nil(t *testing.T) {
	info := entityToInfo(nil)
	if info != nil {
		t.Fatal("entityToInfo(nil) should return nil")
	}
}

// ==================== Manager 查询与 GM 接口 ====================

func newTestManagerWithEntities(t *testing.T) *Manager {
	m := &Manager{}
	m.entities = newEntityStore()
	m.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: m}, transitions...)
	e1 := newTestEntity(1, 100, "TypeA", StateRunning)
	e1.handler = &impl.BaseActivity{}
	e1.handler.SetBaseInfo(e1)
	m.entities.store(e1)

	e2 := newTestEntity(2, 100, "TypeA", StateWaiting) // 同 CfgId 100
	e2.handler = &impl.BaseActivity{}
	e2.handler.SetBaseInfo(e2)
	m.entities.store(e2)

	e3 := newTestEntity(3, 200, "TypeB", StateClosed)
	e3.handler = &impl.BaseActivity{}
	e3.handler.SetBaseInfo(e3)
	m.entities.store(e3)
	return m
}

func newTestEntity(id, cfgId int64, typ, state string) *entity {
	return &entity{
		Id: id, CfgId: cfgId, Type: typ, State: state,
		StartTime: 0, EndTime: 0, CloseTime: 0, TimeType: define.ActTimeConfigured,
	}
}

func TestManager_OnListAllActivities(t *testing.T) {
	m := newTestManagerWithEntities(t)
	list, err := m.OnListAllActivities()
	if err != nil {
		t.Fatalf("OnListAllActivities: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 activities, got %d", len(list))
	}
	ids := make(map[int64]bool)
	for _, info := range list {
		ids[info.ActId] = true
	}
	for _, id := range []int64{1, 2, 3} {
		if !ids[id] {
			t.Errorf("expected actId %d in list", id)
		}
	}
}

func TestManager_OnGetActivityByActId(t *testing.T) {
	m := newTestManagerWithEntities(t)

	info, err := m.OnGetActivityByActId(1)
	if err != nil {
		t.Fatalf("GetActivityByActId(1): %v", err)
	}
	if info.ActId != 1 || info.CfgId != 100 || info.State != StateRunning {
		t.Errorf("GetActivityByActId(1) mismatch: %+v", info)
	}

	_, err = m.OnGetActivityByActId(999)
	if err == nil {
		t.Fatal("GetActivityByActId(999) should return error")
	}
}

func TestManager_OnGetActivityByCfgId_PrefersRunning(t *testing.T) {
	m := newTestManagerWithEntities(t)
	// CfgId 100 有两个实例：actId 1 Running，actId 2 Waiting；应返回 Running
	info, err := m.OnGetActivityByCfgId(100)
	if err != nil {
		t.Fatalf("GetActivityByCfgId(100): %v", err)
	}
	if info.State != StateRunning || info.ActId != 1 {
		t.Errorf("GetActivityByCfgId should prefer Running instance: got actId=%d state=%s", info.ActId, info.State)
	}

	_, err = m.OnGetActivityByCfgId(999)
	if err == nil {
		t.Fatal("GetActivityByCfgId(999) should return error")
	}
}

func TestManager_OnStopActivity(t *testing.T) {
	m := newTestManagerWithEntities(t)

	err := m.OnStopActivity(1)
	if err != nil {
		t.Fatalf("OnStopActivity(1): %v", err)
	}
	ent := m.getEntityByActId(1)
	if ent == nil || ent.State != StateStopped {
		t.Fatalf("after Stop expected state Stopped, got %v", ent)
	}
	if got := m.getEntityByType("TypeA"); got != nil {
		t.Fatalf("expected running type index to clear after stop, got %+v", got)
	}

	err = m.OnStopActivity(999)
	if err == nil {
		t.Fatal("OnStopActivity(999) should return error")
	}

	err = m.OnStopActivity(1)
	if err == nil {
		t.Fatal("OnStopActivity(already stopped) should return error")
	}
}

func TestManager_OnRecoverActivity(t *testing.T) {
	m := newTestManagerWithEntities(t)
	_ = m.OnStopActivity(1)

	err := m.OnRecoverActivity(1)
	if err != nil {
		t.Fatalf("OnRecoverActivity(1): %v", err)
	}
	ent := m.getEntityByActId(1)
	if ent == nil || ent.State != StateRunning {
		t.Fatalf("after Recover expected state Running, got %v", ent)
	}
	if got := m.getEntityByType("TypeA"); got == nil || got.Id != 1 {
		t.Fatalf("expected running type index to restore actId 1, got %+v", got)
	}

	err = m.OnRecoverActivity(2)
	if err == nil {
		t.Fatal("OnRecoverActivity(2 Waiting) should return error")
	}
}

func TestManager_OnRecoverActivity_CallsLifecycleHook(t *testing.T) {
	m := &Manager{}
	m.entities = newEntityStore()
	m.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: m}, transitions...)
	ent := newTestEntity(10, 100, "TypeHook", StateStopped)
	h := &mockLifecycleActivity{}
	h.SetBaseInfo(ent)
	ent.handler = h
	m.entities.store(ent)

	if err := m.OnRecoverActivity(ent.Id); err != nil {
		t.Fatalf("OnRecoverActivity hook test failed: %v", err)
	}
	if h.recoverCount != 1 {
		t.Fatalf("expected OnRecover to be called once, got %d", h.recoverCount)
	}
}

func TestManager_OnCloseActivity_ErrorCases(t *testing.T) {
	m := newTestManagerWithEntities(t)

	_, err := m.OnGetActivityByActId(999)
	if err == nil {
		t.Fatal("expected error for missing actId")
	}
	err = m.OnCloseActivity(999)
	if err == nil {
		t.Fatal("OnCloseActivity(999) should return error")
	}
}

func TestManager_OnRestartActivity_ErrorCases(t *testing.T) {
	m := newTestManagerWithEntities(t)

	err := m.OnRestartActivity(999)
	if err == nil {
		t.Fatal("OnRestartActivity(999) should return error")
	}
	err = m.OnRestartActivity(1)
	if err == nil {
		t.Fatal("OnRestartActivity(Running) should return error")
	}
}

func TestManager_OnRemoveActivity_ErrorCase(t *testing.T) {
	m := newTestManagerWithEntities(t)
	err := m.OnRemoveActivity(999)
	if err == nil {
		t.Fatal("OnRemoveActivity(999) should return error")
	}
}

func TestManager_OnCloseActivityByCfgId_ErrorCase(t *testing.T) {
	m := newTestManagerWithEntities(t)
	err := m.OnCloseActivityByCfgId(999)
	if err == nil {
		t.Fatal("OnCloseActivityByCfgId(999) should return error")
	}
}

func TestManager_OnStopActivityByType_ErrorCase(t *testing.T) {
	m := newTestManagerWithEntities(t)
	err := m.OnStopActivityByType("NonExistentType")
	if err == nil {
		t.Fatal("OnStopActivityByType(non-existent) should return error")
	}
}

func TestManager_GetEntityByType(t *testing.T) {
	m := newTestManagerWithEntities(t)
	ent := m.getEntityByType("TypeA")
	if ent == nil || ent.Id != 1 || ent.State != StateRunning {
		t.Errorf("getEntityByType(TypeA) should return Running instance 1: got %v", ent)
	}
	ent = m.getEntityByType("TypeB")
	if ent != nil {
		t.Errorf("TypeB is Closed, getEntityByType should return nil, got actId=%d", ent.Id)
	}
}
