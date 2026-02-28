package model

import "xfx/proto/proto_stage"

type Stage struct {
	Stage      map[int32]map[int32]*ChapterOpt //存储周目章节为key
	CurStage   int32
	CurChapter int32
	CurCycle   int32 //当前周目
}

type ChapterOpt struct {
	Stages      map[int32]*StageOpt
	HiddenStory *HiddenStory //隐藏剧情
}

type HiddenStory struct {
	UnlockStory bool
	FinishStory bool
}

type StageOpt struct {
	Id        int32
	Exp       int32
	Pass      bool
	PassState int32
}

type BattleReportBack_StageBoss struct {
	Stage   int32
	Chapter int32
	Cycle   int32
	Data    interface{}
}

func ToStageProto(stage map[int32]map[int32]*ChapterOpt) map[int32]*proto_stage.CycleChapter {
	maps := make(map[int32]*proto_stage.CycleChapter)
	for k, v := range stage {
		stageMap := make(map[int32]*proto_stage.StageChapter)
		for kk, vv := range v {
			optMap := make(map[int32]*proto_stage.StateOption)
			if vv.Stages == nil {
				vv.Stages = make(map[int32]*StageOpt)
			}
			if vv.HiddenStory == nil {
				vv.HiddenStory = new(HiddenStory)
			}
			for _, va := range vv.Stages {
				optMap[va.Id] = &proto_stage.StateOption{
					Id:       va.Id,
					Exp:      va.Exp,
					Pass:     va.Pass,
					PassBoss: va.PassState,
				}
			}
			stageMap[kk] = &proto_stage.StageChapter{
				Id: optMap,
				Story: &proto_stage.HiddenStoryline{
					UnlockHidden: vv.HiddenStory.UnlockStory,
					FinishStory:  vv.HiddenStory.FinishStory,
				},
			}
		}
		maps[k] = &proto_stage.CycleChapter{
			Id: stageMap,
		}
	}
	return maps
}

func ToStageSingleProto(cycle, chapter int32, stage *StageOpt) map[int32]*proto_stage.CycleChapter {
	maps := make(map[int32]*proto_stage.CycleChapter)
	stageMap := make(map[int32]*proto_stage.StageChapter)
	optMap := make(map[int32]*proto_stage.StateOption)
	optMap[stage.Id] = &proto_stage.StateOption{
		Id:       stage.Id,
		Exp:      stage.Exp,
		Pass:     stage.Pass,
		PassBoss: stage.PassState,
	}
	stageMap[chapter] = &proto_stage.StageChapter{
		Id: optMap,
	}
	maps[cycle] = &proto_stage.CycleChapter{
		Id: stageMap,
	}
	return maps
}

// 获取最大的章节以及关卡
func (s *Stage) GetMaxChapterStage(cycle int32) (int32, int32) {
	if s.Stage == nil || len(s.Stage) == 0 {
		return 1, 10001
	}

	if _, ok := s.Stage[cycle]; !ok {
		return 1, 10001
	}

	cycleData := s.Stage[cycle]
	// 找到最大的章节ID
	var maxChapter int32
	for chapterID := range cycleData {
		if chapterID > maxChapter {
			maxChapter = chapterID
		}
	}

	// 在最大章节中找到最大的关卡ID
	chapterMap, exists := cycleData[maxChapter]
	if !exists || chapterMap == nil || chapterMap.Stages == nil || len(chapterMap.Stages) == 0 {
		return maxChapter, 0
	}

	var maxStage int32
	for stageID := range chapterMap.Stages {
		if stageID > maxStage {
			maxStage = stageID
		}
	}

	return maxChapter, maxStage
}

// 获取是否通关
func (s *Stage) GetIsPass(cycle, chapter, stage int32) bool {
	if s.Stage == nil || len(s.Stage) == 0 {
		return false
	}

	if _, ok := s.Stage[cycle]; !ok {
		return false
	}

	if _, ok := s.Stage[cycle][chapter]; !ok {
		return false
	}

	if _, ok := s.Stage[cycle][chapter].Stages[stage]; !ok {
		return false
	}

	return s.Stage[cycle][chapter].Stages[stage].Pass
}
