package vercel

import "testing"

func TestDeploymentStateLabel(t *testing.T) {
	tests := []struct {
		name       string
		deployment Deployment
		want       string
	}{
		{name: "state wins", deployment: Deployment{State: "READY", ReadyState: "BUILDING"}, want: "READY"},
		{name: "ready fallback", deployment: Deployment{ReadyState: "ERROR"}, want: "ERROR"},
		{name: "unknown", deployment: Deployment{}, want: "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.deployment.StateLabel(); got != tt.want {
				t.Fatalf("StateLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeploymentEventBuildLogLineSupportsTopLevelShape(t *testing.T) {
	event := deploymentEvent{
		Type:    "stdout",
		Created: flexibleInt(1783609780974),
		Text:    "Build cache uploaded: 3.851s",
		Info: eventInfo{
			Type:       "build",
			Name:       "bld_123",
			Entrypoint: ".",
		},
	}

	line := event.BuildLogLine()
	if line.CreatedAt != 1783609780974 {
		t.Fatalf("unexpected CreatedAt: %d", line.CreatedAt)
	}
	if line.Type != "stdout" {
		t.Fatalf("unexpected Type: %q", line.Type)
	}
	if line.Step != "bld_123" {
		t.Fatalf("unexpected Step: %q", line.Step)
	}
	if line.Entrypoint != "." {
		t.Fatalf("unexpected Entrypoint: %q", line.Entrypoint)
	}
	if line.Text != "Build cache uploaded: 3.851s" {
		t.Fatalf("unexpected Text: %q", line.Text)
	}
}
