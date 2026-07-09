package vercel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverProjectsReadsVercelProjectFile(t *testing.T) {
	dir := t.TempDir()
	vercelDir := filepath.Join(dir, ".vercel")
	if err := os.Mkdir(vercelDir, 0o755); err != nil {
		t.Fatal(err)
	}

	data := []byte(`{"projectId":"prj_123","orgId":"team_123","projectName":"example"}`)
	if err := os.WriteFile(filepath.Join(vercelDir, "project.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	projects, err := DiscoverProjects([]string{dir})
	if err != nil {
		t.Fatalf("DiscoverProjects returned error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected one project, got %d", len(projects))
	}
	if projects[0].ProjectID != "prj_123" || projects[0].OrgID != "team_123" || projects[0].ProjectName != "example" {
		t.Fatalf("unexpected project: %#v", projects[0])
	}
}

func TestDiscoverProjectsRejectsUnlinkedDirectory(t *testing.T) {
	_, err := DiscoverProjects([]string{t.TempDir()})
	if err == nil {
		t.Fatal("expected unlinked directory error")
	}
}

func TestDiscoverProjectIfLinkedIgnoresUnlinkedDirectory(t *testing.T) {
	_, ok, err := DiscoverProjectIfLinked(t.TempDir())
	if err != nil {
		t.Fatalf("DiscoverProjectIfLinked returned error: %v", err)
	}
	if ok {
		t.Fatal("expected unlinked directory to return ok=false")
	}
}

func TestMergeProjectsSelectsCWDLinkedProject(t *testing.T) {
	catalog := []Project{
		{ID: "prj_old", Name: "old", AccountID: "team_1", AccountSlug: "team", UpdatedAt: 20},
		{ID: "prj_cwd", Name: "cwd", AccountID: "team_1", AccountSlug: "team", UpdatedAt: 10},
	}
	cwd := LinkedProject{ProjectID: "prj_cwd", OrgID: "team_1", ProjectName: "cwd", Dir: "/tmp/cwd"}

	projects, initial := mergeProjects(catalog, nil, cwd, true)
	if projects[initial].ID != "prj_cwd" {
		t.Fatalf("expected cwd project selected, got %#v", projects[initial])
	}
	if !projects[initial].LinkedCWD {
		t.Fatal("expected cwd project to be marked LinkedCWD")
	}
}

func TestMergeProjectsFallsBackToMostRecentlyUpdated(t *testing.T) {
	catalog := []Project{
		{ID: "prj_old", Name: "old", AccountID: "team_1", AccountSlug: "team", UpdatedAt: 10},
		{ID: "prj_new", Name: "new", AccountID: "team_1", AccountSlug: "team", UpdatedAt: 20},
	}

	projects, initial := mergeProjects(catalog, nil, LinkedProject{}, false)
	if projects[initial].ID != "prj_new" {
		t.Fatalf("expected newest project selected, got %#v", projects[initial])
	}
}
