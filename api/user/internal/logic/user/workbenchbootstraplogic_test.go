package user

import (
	"testing"

	"GoZero-AI/api/user/internal/types"
)

func TestBuildWorkbenchActionsUsesWorkbenchRoutes(t *testing.T) {
	actions := buildWorkbenchActions(1, types.WorkbenchResumeSummary{}, types.WorkbenchKnowledgeSummary{})
	routes := make(map[string]string, len(actions))
	for _, action := range actions {
		routes[action.Key] = action.Route
	}

	want := map[string]string{
		"new_interview":    "/workbench/new",
		"upload_resume":    "/workbench/resume",
		"import_knowledge": "/workbench/knowledge",
		"review_report":    "/workbench",
	}
	for key, route := range want {
		if routes[key] != route {
			t.Fatalf("route[%s] = %q, want %q", key, routes[key], route)
		}
	}
}
