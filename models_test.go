package decision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAssignments(t *testing.T) {
	var va *VisitorAssignments
	assert.Nil(t, va.getAssignments())

	assignments := map[string]*VisitorCache{}
	va = &VisitorAssignments{
		Assignments: assignments,
	}
	assert.Equal(t, assignments, va.getAssignments())
}

func TestGetAssignment(t *testing.T) {
	var va *VisitorAssignments
	r, ok := va.getAssignment("test_vg_id")
	assert.False(t, ok)
	assert.Nil(t, r)

	assignment := &VisitorCache{
		VariationID: "vid",
		Activated:   true,
	}
	assignments := map[string]*VisitorCache{
		"test_vg_id": assignment,
	}
	va = &VisitorAssignments{
		Assignments: assignments,
	}
	r, ok = va.getAssignment("test_vg_id")
	assert.True(t, ok)
	assert.Equal(t, assignment, r)
}
