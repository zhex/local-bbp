package runner

import (
	"github.com/google/uuid"
	"github.com/zhex/local-bbp/internal/common"
)

type ProjectInfo struct {
	Path       string
	ID         string
	Name       string
	Owner      string
	OwnerID    string
	RepoID     string
	BranchName string
	CommitID   string
}

func NewProjInfo(hostPath string) *ProjectInfo {
	branch, _ := common.GetGitBranch(hostPath)
	commit, _ := common.GetGitCommit(hostPath)
	owner, _ := common.GetGitOwner(hostPath)

	return &ProjectInfo{
		Path:       hostPath,
		ID:         uuid.New().String(),
		Name:       "local-bbp",
		Owner:      owner,
		OwnerID:    uuid.New().String(),
		BranchName: branch,
		CommitID:   commit,
		RepoID:     uuid.New().String(),
	}
}
