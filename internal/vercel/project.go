package vercel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type LinkedProject struct {
	Dir         string
	ProjectID   string
	OrgID       string
	ProjectName string
}

type Account struct {
	ID   string
	Slug string
	Name string
}

type Project struct {
	ID          string
	Name        string
	AccountID   string
	AccountSlug string
	AccountName string
	Framework   string
	UpdatedAt   int64
	LinkedDir   string
	LinkedCWD   bool
	Pinned      bool
	Link        ProjectLink
}

type ProjectLink struct {
	Type             string `json:"type"`
	Repo             string `json:"repo"`
	Org              string `json:"org"`
	ProductionBranch string `json:"productionBranch"`
}

type projectFile struct {
	ProjectID   string `json:"projectId"`
	OrgID       string `json:"orgId"`
	ProjectName string `json:"projectName"`
}

func DiscoverProjects(dirs []string) ([]LinkedProject, error) {
	projects := make([]LinkedProject, 0, len(dirs))
	for _, dir := range dirs {
		project, err := discoverProject(dir)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func DiscoverProjectIfLinked(dir string) (LinkedProject, bool, error) {
	project, err := discoverProject(dir)
	if err != nil {
		if os.IsNotExist(rootCause(err)) {
			return LinkedProject{}, false, nil
		}
		return LinkedProject{}, false, err
	}
	return project, true, nil
}

func discoverProject(dir string) (LinkedProject, error) {
	expanded, err := expandPath(dir)
	if err != nil {
		return LinkedProject{}, err
	}

	absolute, err := filepath.Abs(expanded)
	if err != nil {
		return LinkedProject{}, fmt.Errorf("resolve %q: %w", dir, err)
	}

	data, err := os.ReadFile(filepath.Join(absolute, ".vercel", "project.json"))
	if err != nil {
		return LinkedProject{}, fmt.Errorf("%s is not linked to Vercel; run `vercel link` there first: %w", absolute, err)
	}

	var parsed projectFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		return LinkedProject{}, fmt.Errorf("parse %s: %w", filepath.Join(absolute, ".vercel", "project.json"), err)
	}
	if parsed.ProjectID == "" || parsed.OrgID == "" {
		return LinkedProject{}, fmt.Errorf("%s is missing projectId or orgId", filepath.Join(absolute, ".vercel", "project.json"))
	}
	if parsed.ProjectName == "" {
		parsed.ProjectName = filepath.Base(absolute)
	}

	return LinkedProject{
		Dir:         absolute,
		ProjectID:   parsed.ProjectID,
		OrgID:       parsed.OrgID,
		ProjectName: parsed.ProjectName,
	}, nil
}

func ProjectFromLinked(linked LinkedProject) Project {
	return Project{
		ID:        linked.ProjectID,
		Name:      linked.ProjectName,
		AccountID: linked.OrgID,
		LinkedDir: linked.Dir,
	}
}

func rootCause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

func expandPath(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	if len(path) > 1 && os.IsPathSeparator(path[1]) {
		return filepath.Join(home, path[2:]), nil
	}
	return "", fmt.Errorf("cannot expand user-qualified path %q", path)
}
