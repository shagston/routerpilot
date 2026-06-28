package planner

import (
	"testing"

	"github.com/shagston/routerpilot/sdk/types"
)

func TestNextSegmentSplitsContextAndActionBlocks(t *testing.T) {
	steps := []types.Task{
		{ID: "c1", Purpose: types.TaskPurposeContext, Tool: "network.route_get"},
		{ID: "c2", Purpose: types.TaskPurposeContext, Tool: "network.interface_status"},
		{ID: "a1", Tool: "network.route_add"},
	}

	segment, remaining, isContext := NextSegment(steps)
	if !isContext || len(segment) != 2 || len(remaining) != 1 {
		t.Fatalf("unexpected first segment: segment=%+v remaining=%+v isContext=%v", segment, remaining, isContext)
	}

	segment, remaining, isContext = NextSegment(remaining)
	if isContext || len(segment) != 1 || len(remaining) != 0 {
		t.Fatalf("unexpected second segment: segment=%+v remaining=%+v isContext=%v", segment, remaining, isContext)
	}
}

func TestSplitLeadingActions(t *testing.T) {
	steps := []types.Task{
		{ID: "a1", Tool: "network.route_add"},
		{ID: "a2", Tool: "network.interface_set_state"},
		{ID: "c1", Purpose: types.TaskPurposeContext, Tool: "network.route_get"},
	}

	actions, remaining := SplitLeadingActions(steps)
	if len(actions) != 2 || len(remaining) != 1 {
		t.Fatalf("unexpected split: actions=%+v remaining=%+v", actions, remaining)
	}
}

func TestNextSegmentPullsDependenciesFromRemaining(t *testing.T) {
	steps := []types.Task{
		{ID: "a1", Tool: "network.route_add", Dependencies: []types.TaskID{"c1"}},
		{ID: "a2", Tool: "network.ping"},
		{ID: "c1", Purpose: types.TaskPurposeContext, Tool: "network.route_get"},
	}

	segment, remaining, isContext := NextSegment(steps)
	if isContext || len(segment) != 3 || len(remaining) != 0 {
		t.Fatalf("expected all tasks consolidated into one segment; got segment=%+v remaining=%+v isContext=%v", segment, remaining, isContext)
	}
	// a1 depends on c1, both must be in the same segment
	ids := map[types.TaskID]bool{}
	for _, t := range segment {
		ids[t.ID] = true
	}
	if !ids["a1"] || !ids["a2"] || !ids["c1"] {
		t.Fatalf("expected a1, a2, c1 all in segment, got %+v", segment)
	}
}

func TestNextSegmentPreservesPurposeBlocksWithoutCrossDeps(t *testing.T) {
	steps := []types.Task{
		{ID: "c1", Purpose: types.TaskPurposeContext, Tool: "network.route_get"},
		{ID: "a1", Tool: "network.ping"},
	}

	segment, remaining, isContext := NextSegment(steps)
	if !isContext || len(segment) != 1 || len(remaining) != 1 {
		t.Fatalf("expected single context task in first segment; got segment=%+v remaining=%+v", segment, remaining)
	}
	if remaining[0].ID != "a1" {
		t.Fatalf("expected a1 in remaining, got %+v", remaining)
	}
}
