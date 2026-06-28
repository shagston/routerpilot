package planner

import "github.com/shagston/routerpilot/sdk/types"

func NextSegment(steps []types.Task) (segment []types.Task, remaining []types.Task, isContextSegment bool) {
	if len(steps) == 0 {
		return nil, nil, false
	}

	purpose := effectivePurpose(steps[0])

	segmentIDs := make(map[types.TaskID]bool)
	segment = nil

	// Collect initial contiguous same-purpose block
	for i, t := range steps {
		if effectivePurpose(t) != purpose {
			remaining = steps[i:]
			break
		}
		segmentIDs[t.ID] = true
		segment = append(segment, t)
	}

	if remaining == nil {
		remaining = []types.Task{}
	}

	// Pull in any missing dependencies from remaining
	for {
		added := false
		for _, t := range segment {
			for _, depID := range t.Dependencies {
				if segmentIDs[depID] {
					continue
				}
				for j, rt := range remaining {
					if rt.ID == depID {
						segmentIDs[depID] = true
						segment = append(segment, rt)
						remaining = append(remaining[:j], remaining[j+1:]...)
						added = true
						break
					}
				}
			}
		}
		if !added {
			break
		}
	}

	// Defer tasks whose dependencies still aren't in the segment
	var deferred []types.Task
	var safe []types.Task
	for _, t := range segment {
		depsMet := true
		for _, depID := range t.Dependencies {
			if !segmentIDs[depID] {
				depsMet = false
				break
			}
		}
		if depsMet {
			safe = append(safe, t)
		} else {
			deferred = append(deferred, t)
		}
	}
	segment = safe
	remaining = append(deferred, remaining...)

	if len(segment) == 0 {
		return nil, steps, false
	}

	return segment, remaining, effectivePurpose(segment[0]) == types.TaskPurposeContext
}

func SplitLeadingActions(steps []types.Task) (actions []types.Task, remaining []types.Task) {
	index := 0
	for index < len(steps) && effectivePurpose(steps[index]) != types.TaskPurposeContext {
		index++
	}
	return steps[:index], steps[index:]
}

func effectivePurpose(t types.Task) types.TaskPurpose {
	if t.Purpose == "" {
		return types.TaskPurposeAction
	}
	return t.Purpose
}
