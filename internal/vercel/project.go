package vercel

import (
	"encoding/json"
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
		return LinkedProject{}, fmt.Errorf("%s is not linked to Vercel; run `vercel link` there first", absolute)
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
