package task

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"xfx/core/config"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/agent"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_task"
)

var initTaskConfigOnce sync.Once

type fakePlayer struct {
	messages []any
}

func (f *fakePlayer) Self() agent.PID                          { return nil }
func (f *fakePlayer) Cast(pid agent.PID, msg any)              {}
func (f *fakePlayer) Call(pid agent.PID, msg any) (any, error) { return nil, nil }
func (f *fakePlayer) Invoke(mod, fn string, args ...any) (any, error) {
	return nil, nil
}
func (f *fakePlayer) InvokeP(pid agent.PID, fn string, args ...any) (any, error) {
	return nil, nil
}
func (f *fakePlayer) Send(msg any)          { f.messages = append(f.messages, msg) }
func (f *fakePlayer) Watch(pid agent.PID)   {}
func (f *fakePlayer) Unwatch(pid agent.PID) {}
func (f *fakePlayer) OnSave(isSync bool)    {}
func (f *fakePlayer) Stop()                 {}

func initTaskTestConfig(t *testing.T) {
	t.Helper()
	initTaskConfigOnce.Do(func() {
		log.DefaultInit()
		_, filename, _, _ := runtime.Caller(0)
		configDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "core", "config", "json")
		config.InitConfig(configDir)
	})
}

func newTestPlayer() *model.Player {
	pl := &model.Player{
		Hero: &model.Hero{Hero: map[int32]*model.HeroOption{1001: {Id: 1001, Level: 12, Stage: 0}}},
	}
	pl.SetProp(define.PlayerPropHeroId, 1001, false)
	Init(pl)
	return pl
}

func setStableResetTimes(pl *model.Player) {
	now := utils.Now().Unix()
	pl.Task.ResetAt[define.TaskTypeDaily] = now
	pl.Task.ResetAt[define.TaskTypeWeek] = now
	pl.Task.ResetAt[define.TaskTypeMonth] = now
	pl.Task.ResetAt[define.TaskTypeGuild] = now
	pl.Task.ResetAt[define.TaskTypePassportDaily] = now
	pl.Task.ResetAt[define.TaskTypePassportWeek] = now
}

func findGroup(groups []*proto_task.TaskGroup, bucketType int32) *proto_task.TaskGroup {
	for _, group := range groups {
		if group.BucketType == bucketType {
			return group
		}
	}
	return nil
}

func TestBuildTaskSnapshotIncludesGroups(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()

	setBucket(pl, define.TaskTypeDaily, map[int32]*model.Task{
		101: {Id: 101, TaskType: define.TaskLoginXTimes, Condition: 1, ExtraCondition: 0, InitialProcess: 0, ReceiveAward: true},
	})
	setTaskInfo(pl, define.TaskLoginXTimes, 0, 3, false)
	setPoint(pl, define.TaskActivityTypeDaily, 25)
	setClaimMap(pl, claimTypeDaily, map[int32]bool{1: true})

	snapshot := buildTaskSnapshot(pl)
	if len(snapshot.Groups) != len(bucketPolicies) {
		t.Fatalf("group count = %d, want %d", len(snapshot.Groups), len(bucketPolicies))
	}

	daily := findGroup(snapshot.Groups, define.TaskTypeDaily)
	if daily == nil {
		t.Fatal("daily group missing")
	}
	if daily.Point != 25 {
		t.Fatalf("daily point = %d, want 25", daily.Point)
	}
	if !daily.ClaimRecord[1] {
		t.Fatal("daily claim record missing reward 1")
	}
	if daily.Tasks[101] == nil || daily.Tasks[101].Progress != 3 || !daily.Tasks[101].Rewarded {
		t.Fatalf("daily task state = %+v, want progress=3 rewarded=true", daily.Tasks[101])
	}
}

func TestBuildIncrementalPushOnlyReturnsChangedGroups(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()
	setBucket(pl, define.TaskTypeDaily, map[int32]*model.Task{
		101: {Id: 101, TaskType: define.TaskLoginXTimes, Condition: 1, ExtraCondition: 0},
	})
	setBucket(pl, define.TaskTypeWeek, map[int32]*model.Task{
		201: {Id: 201, TaskType: define.TaskDrawCard, Condition: 1, ExtraCondition: 0},
	})
	setPoint(pl, define.TaskActivityTypeDaily, 10)

	push := buildIncrementalPush(pl, define.TaskLoginXTimes, 0, 2)
	if len(push.Groups) != 1 {
		t.Fatalf("push groups = %d, want 1", len(push.Groups))
	}
	group := push.Groups[0]
	if group.BucketType != define.TaskTypeDaily {
		t.Fatalf("bucketType = %d, want %d", group.BucketType, define.TaskTypeDaily)
	}
	if group.Point != 10 {
		t.Fatalf("point = %d, want 10", group.Point)
	}
	if group.Tasks[101] == nil || group.Tasks[101].Progress != 2 || group.Tasks[101].Rewarded {
		t.Fatalf("group task = %+v, want progress=2 rewarded=false", group.Tasks[101])
	}
}

func TestReqReceiveActivePointRewardUpdatesClaimGroup(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()
	pl.Bag = &model.Bag{Items: map[int32]int32{}}
	ctx := &fakePlayer{}
	now := utils.Now().Unix()
	pl.Task.ResetAt[define.TaskTypeDaily] = now
	pl.Task.ResetAt[define.TaskTypeWeek] = now
	pl.Task.ResetAt[define.TaskTypeMonth] = now
	pl.Task.ResetAt[define.TaskTypeGuild] = now
	pl.Task.ResetAt[define.TaskTypePassportDaily] = now
	pl.Task.ResetAt[define.TaskTypePassportWeek] = now
	setPoint(pl, define.TaskActivityTypeDaily, 20)
	setClaimMap(pl, claimTypeDaily, map[int32]bool{})
	setBucket(pl, define.TaskTypeDaily, map[int32]*model.Task{})

	ReqReceiveActivePointReward(ctx, pl, &proto_task.C2SGetActivePointReward{RewardId: 1})

	var push *proto_task.PushTask
	var resp *proto_task.S2CGetActivePointReward
	for _, msg := range ctx.messages {
		switch v := msg.(type) {
		case *proto_task.PushTask:
			push = v
		case *proto_task.S2CGetActivePointReward:
			resp = v
		}
	}
	if push == nil {
		t.Fatalf("messages = %#v, want one *proto_task.PushTask", ctx.messages)
	}
	if len(push.Groups) != 1 {
		t.Fatalf("push groups = %d, want 1", len(push.Groups))
	}
	daily := push.Groups[0]
	if daily.BucketType != define.TaskTypeDaily {
		t.Fatalf("bucketType = %d, want %d", daily.BucketType, define.TaskTypeDaily)
	}
	if !daily.ClaimRecord[1] {
		t.Fatal("expected claim record for reward 1")
	}
	if resp == nil || !resp.Succ {
		t.Fatalf("messages = %#v, want successful S2CGetActivePointReward", ctx.messages)
	}
	if pl.Bag.Items[102] != 10 {
		t.Fatalf("bag item 102 = %d, want 10", pl.Bag.Items[102])
	}
}

func TestLoadTaskFromConfigAchieveLoadsNextTask(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()

	var firstID int32
	var nextID int32
	for _, cfg := range config.Task.All() {
		if cfg.Type == define.TaskTypeAchieve && cfg.FrontTask == 0 && cfg.BackTask != 0 {
			firstID = cfg.Id
			nextID = cfg.BackTask
			break
		}
	}
	if firstID == 0 || nextID == 0 {
		t.Skip("no achieve task chain found in config")
	}

	setBucket(pl, define.TaskTypeAchieve, map[int32]*model.Task{
		firstID: {Id: firstID, ReceiveAward: true},
	})

	next := loadTaskFromConfig(pl, define.TaskTypeAchieve)
	if len(next) != 1 {
		t.Fatalf("next achieve count = %d, want 1", len(next))
	}
	if next[nextID] == nil {
		t.Fatalf("expected next achieve task %d, got %v", nextID, next)
	}
}

func TestReqReceiveRewardRejectsMissingTask(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()
	ctx := &fakePlayer{}
	setStableResetTimes(pl)
	setBucket(pl, define.TaskTypeDaily, map[int32]*model.Task{})

	ReqReceiveReward(ctx, pl, &proto_task.C2SGetReward{BucketType: define.TaskTypeDaily, TaskId: 1})

	if len(ctx.messages) != 1 {
		t.Fatalf("message count = %d, want 1", len(ctx.messages))
	}
	resp, ok := ctx.messages[0].(*proto_task.S2CGetReward)
	if !ok {
		t.Fatalf("message type = %T, want *proto_task.S2CGetReward", ctx.messages[0])
	}
	if resp.Succ {
		t.Fatal("expected reward response to fail for missing task")
	}
}

func TestResetTaskDailyClearsPointClaimAndLimit(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()
	ctx := &fakePlayer{}
	setBucket(pl, define.TaskTypeDaily, map[int32]*model.Task{
		1: {Id: 1, TaskType: define.TaskLoginXTimes, Condition: 1},
	})
	setPoint(pl, define.TaskActivityTypeDaily, 88)
	setClaimMap(pl, claimTypeDaily, map[int32]bool{1: true})
	pl.Task.TaskLimit[define.TaskLoginXTimes] = 3
	pl.Task.ResetAt[define.TaskTypeDaily] = 1
	pl.Task.ResetAt[define.TaskTypePassportDaily] = 1

	resetTask(ctx, pl, define.TaskTypeDaily)

	if getPoint(pl, define.TaskActivityTypeDaily) != 0 {
		t.Fatalf("daily point = %d, want 0", getPoint(pl, define.TaskActivityTypeDaily))
	}
	if len(getClaimMap(pl, claimTypeDaily)) != 0 {
		t.Fatalf("daily claims = %v, want empty", getClaimMap(pl, claimTypeDaily))
	}
	if len(pl.Task.TaskLimit) != 0 {
		t.Fatalf("task limit = %v, want empty", pl.Task.TaskLimit)
	}
	if len(getBucket(pl, define.TaskTypeDaily)) == 0 {
		t.Fatal("daily bucket not reloaded")
	}
	if pl.Task.ResetAt[define.TaskTypeDaily] == 1 {
		t.Fatal("daily reset time not refreshed")
	}
}

func TestDispatchRespectsTaskCompleteLimit(t *testing.T) {
	initTaskTestConfig(t)
	pl := newTestPlayer()
	ctx := &fakePlayer{}
	setStableResetTimes(pl)
	setBucket(pl, define.TaskTypeDaily, map[int32]*model.Task{
		1: {Id: 1, TaskType: define.TaskLoginXTimes, Condition: 1, ExtraCondition: 0},
	})

	Dispatch(ctx, pl, define.TaskLoginXTimes, 1, 0, true)
	if pl.Task.TaskLimit[define.TaskLoginXTimes] != 1 {
		t.Fatalf("first dispatch limit = %d, want 1", pl.Task.TaskLimit[define.TaskLoginXTimes])
	}
	if len(ctx.messages) != 1 {
		t.Fatalf("first dispatch messages = %d, want 1", len(ctx.messages))
	}

	Dispatch(ctx, pl, define.TaskLoginXTimes, 1, 0, true)
	if pl.Task.TaskLimit[define.TaskLoginXTimes] != 1 {
		t.Fatalf("second dispatch limit = %d, want stay 1", pl.Task.TaskLimit[define.TaskLoginXTimes])
	}
	if len(ctx.messages) != 1 {
		t.Fatalf("second dispatch messages = %d, want still 1", len(ctx.messages))
	}
	if getTaskProgress(pl, define.TaskLoginXTimes, 0) != 1 {
		t.Fatalf("task progress = %d, want 1", getTaskProgress(pl, define.TaskLoginXTimes, 0))
	}
}

var _ global.IPlayer = (*fakePlayer)(nil)
