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
