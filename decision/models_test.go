package decision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAssignments(t *testing.T) {
	var va *VisitorAssignments
	assert.Nil(t, va.GetAssignments())

	assignments := map[string]*VisitorVGCacheItem{}
	va = &VisitorAssignments{
		Assignments: assignments,
	}
	assert.Equal(t, assignments, va.GetAssignments())
}

func TestGetAssignment(t *testing.T) {
	var va *VisitorAssignments
	r, ok := va.GetAssignment("test_vg_id")
	assert.False(t, ok)
	assert.Nil(t, r)

	assignment := &VisitorVGCacheItem{
		VariationID: "vid",
		Activated:   true,
	}
	assignments := map[string]*VisitorVGCacheItem{
		"test_vg_id": assignment,
	}
	va = &VisitorAssignments{
		Assignments: assignments,
	}
	r, ok = va.GetAssignment("test_vg_id")
	assert.True(t, ok)
	assert.Equal(t, assignment, r)
}
