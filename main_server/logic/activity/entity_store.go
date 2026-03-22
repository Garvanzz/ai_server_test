package activity

import "sync"

type entityStore struct {
	mu            sync.RWMutex
	byID          map[int64]*entity
	byCfgID       map[int64]map[int64]*entity
	runningByType map[string]*entity
}

func newEntityStore() entityStore {
	return entityStore{
		byID:          make(map[int64]*entity),
		byCfgID:       make(map[int64]map[int64]*entity),
		runningByType: make(map[string]*entity),
	}
}

func (s *entityStore) load(id int64) (*entity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ent, ok := s.byID[id]
	return ent, ok
}

func (s *entityStore) store(ent *entity) {
	if ent == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if old, ok := s.byID[ent.Id]; ok {
		s.removeLocked(old)
	}

	s.byID[ent.Id] = ent
	if _, ok := s.byCfgID[ent.CfgId]; !ok {
		s.byCfgID[ent.CfgId] = make(map[int64]*entity)
	}
	s.byCfgID[ent.CfgId][ent.Id] = ent
	if ent.State == StateRunning {
		s.runningByType[ent.Type] = ent
	}
}

func (s *entityStore) delete(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ent, ok := s.byID[id]
	if !ok {
		return
	}
	s.removeLocked(ent)
}

func (s *entityStore) snapshot() []*entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*entity, 0, len(s.byID))
	for _, ent := range s.byID {
		result = append(result, ent)
	}
	return result
}

func (s *entityStore) resetChecked() {
	for _, ent := range s.snapshot() {
		ent.checked = false
	}
}

func (s *entityStore) updateState(ent *entity, fromState, toState string) {
	if ent == nil || fromState == toState {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if toState == StateRunning {
		s.runningByType[ent.Type] = ent
		return
	}
	if fromState != StateRunning {
		return
	}

	current, ok := s.runningByType[ent.Type]
	if !ok || current.Id != ent.Id {
		return
	}
	delete(s.runningByType, ent.Type)
	for _, candidate := range s.byID {
		if candidate.Id != ent.Id && candidate.Type == ent.Type && candidate.State == StateRunning {
			s.runningByType[ent.Type] = candidate
			break
		}
	}
}

func (s *entityStore) getByCfgID(cfgId int64) *entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entities, ok := s.byCfgID[cfgId]
	if !ok {
		return nil
	}

	var found *entity
	for _, ent := range entities {
		if ent.State == StateRunning {
			return ent
		}
		if found == nil {
			found = ent
		}
	}
	return found
}

func (s *entityStore) getRunningByType(typ string) *entity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runningByType[typ]
}

func (s *entityStore) removeLocked(ent *entity) {
	delete(s.byID, ent.Id)
	if entities, ok := s.byCfgID[ent.CfgId]; ok {
		delete(entities, ent.Id)
		if len(entities) == 0 {
			delete(s.byCfgID, ent.CfgId)
		}
	}
	if running, ok := s.runningByType[ent.Type]; ok && running.Id == ent.Id {
		delete(s.runningByType, ent.Type)
	}
}
