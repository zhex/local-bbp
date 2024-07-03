package runner

import (
	"github.com/google/uuid"
	"github.com/zhex/local-bbp/internal/common"
)

type info struct {
	ProjectID   string
	ProjectName string
	Owner       string
	OwnerID     string
	RepoID      string
	BranchName  string
	CommitID    string
}

func newInfo(hostPath string) *info {
	branch, _ := common.GetGitBranch(hostPath)
	commit, _ := common.GetGitCommit(hostPath)
	owner, _ := common.GetGitOwner(hostPath)

	return &info{
		ProjectID:   uuid.New().String(),
		ProjectName: "local-bbp",
		Owner:       owner,
		OwnerID:     uuid.New().String(),
		BranchName:  branch,
		CommitID:    commit,
		RepoID:      uuid.New().String(),
	}
}
