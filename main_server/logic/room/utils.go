package room

// 获取邀请id
func (mgr *Manager) getInviteId() int32 {
	mgr.inviteId++
	return mgr.inviteId
}
