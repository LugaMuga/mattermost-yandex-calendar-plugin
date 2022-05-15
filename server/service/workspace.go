package service

import (
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"sync"
)

type Workspace struct {
	repo *repository.WorkspaceRepo
	sync.Mutex
}

func NewWorkspaceService(workspaceRepo *repository.WorkspaceRepo) *Workspace {
	return &Workspace{
		repo: workspaceRepo,
	}
}

func (w *Workspace) AddUser(userId string) {
	w.Lock()
	defer w.Unlock()
	userIds := w.GetUserIds()
	if userIds == nil {
		userIds = make(map[string]bool)
	}
	if !userIds[userId] {
		userIds[userId] = true
		w.repo.SaveUserIds(userIds)
	}
}

func (w *Workspace) DeleteUser(userId string) {
	w.Lock()
	defer w.Unlock()
	userIds := w.GetUserIds()
	if userIds[userId] {
		delete(userIds, userId)
		w.repo.SaveUserIds(userIds)
	}
	w.repo.DeleteUser(userId)
}

func (w *Workspace) GetUserIds() map[string]bool {
	userIds := w.repo.GetUserIds()
	if userIds == nil {
		return make(map[string]bool)
	}
	return *userIds
}
