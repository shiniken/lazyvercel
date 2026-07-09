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
