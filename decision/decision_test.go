package decision

import (
	"fmt"
	"testing"

	"github.com/flagship-io/flagship-proto/decision_response"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func mockGetCache(environmentID string, id string) (*VisitorAssignments, error) {
	vi := VisitorAssignments{}
	return &vi, nil
}

func mockSaveCache(environmentID string, id string, assignment *VisitorAssignments) error {
	fmt.Println("Save cache environment", environmentID, "id", id, "assignments", assignment)
	return nil
}

func mockActivateCampaigns(activations []*VisitorActivation) error {
	fmt.Println("Activate campaigns", activations)
	return nil
}

var cache = map[string]*VisitorAssignments{}

func localGetCache(environmentID string, id string) (*VisitorAssignments, error) {
	return cache[environmentID+id], nil
}

func TestGetCache(t *testing.T) {
	envID := "env_id"
	visitorID := "visitor_id"
	anonymousID := "anonymous_id"
	decisionGroup := "decisionGroup"

	assignments, err := getCache(envID, visitorID, anonymousID, decisionGroup, false, localGetCache)
	assert.Nil(t, err)
	assert.Nil(t, assignments.Standard)
	assert.Nil(t, assignments.Anonymous)

	newAssignments := map[string]*VisitorVGCacheItem{
		"vg_id": {
			VariationID: "v_id",
			Activated:   true,
		},
	}
	newAssignmentsDG := map[string]*VisitorVGCacheItem{
		"vg2_id": {
			VariationID: "v2_id",
			Activated:   true,
		},
	}
	cache[envID+visitorID] = &VisitorAssignments{
		Assignments: newAssignments,
	}

	assignments, err = getCache(envID, visitorID, anonymousID, decisionGroup, false, localGetCache)
	assert.Nil(t, err)
	assert.EqualValues(t, newAssignments, assignments.Standard.GetAssignments())
	assert.Nil(t, assignments.Anonymous)

	cache[envID+anonymousID] = &VisitorAssignments{
		Assignments: newAssignments,
	}
	cache[envID+decisionGroup] = &VisitorAssignments{
		Assignments: newAssignmentsDG,
	}

	assignments, err = getCache(envID, visitorID, anonymousID, decisionGroup, true, localGetCache)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(assignments.Standard.GetAssignments()))
	assert.EqualValues(t, newAssignments, assignments.Anonymous.GetAssignments())
}

func TestDecisionBucketInNoCache(t *testing.T) {
	vi := VisitorInfo{}
	vi.ID = "v1"
	vi.Context = map[string]*structpb.Value{
		"isVIP": structpb.NewBoolValue(true),
	}
	vi.DecisionGroup = "decision"

	ei := EnvironmentInfo{}
	ei.ID = "e123"
	ei.Campaigns = map[string]*CampaignInfo{
		"a": {
			ID:           "a1",
			BucketRanges: [][]float64{{0., 100.}},
			VariationsGroups: map[string]*VariationsGroup{
				"vga": {
					ID:         "vga",
					Targetings: createBoolTargeting(),
					Variations: []*Variation{
						{
							ID:         "vgav1",
							Allocation: 100,
							Modifications: &decision_response.Modifications{
								Type:  decision_response.ModificationsType_FLAG,
								Value: structpb.NewStringValue("toto").GetStructValue(),
							},
						},
					},
				},
			},
		},
		"b": {
			ID:           "a2",
			BucketRanges: [][]float64{{20., 30.}},
			VariationsGroups: map[string]*VariationsGroup{
				"vgb": {
					ID:         "vgb",
					Targetings: createBoolTargeting(),
					Variations: []*Variation{
						{
							ID:         "vgbv1",
							Allocation: 100,
							Modifications: &decision_response.Modifications{
								Type:  decision_response.ModificationsType_FLAG,
								Value: structpb.NewStringValue("tata").GetStructValue(),
							},
						},
					},
				},
			},
		},
	}
	for _, vg := range ei.Campaigns["a"].VariationsGroups {
		vg.Campaign = ei.Campaigns["a"]
	}
	for _, vg := range ei.Campaigns["b"].VariationsGroups {
		vg.Campaign = ei.Campaigns["b"]
	}

	options := DecisionOptions{}
	// no options

	handlers := DecisionHandlers{}
	handlers.GetCache = mockGetCache
	handlers.SaveCache = mockSaveCache
	handlers.ActivateCampaigns = mockActivateCampaigns

	decision, err := GetDecision(vi, ei, options, handlers)

	// only one campaign should be returned, they have the same targeting and allocation but the second one has not the right bucket allocation
	assert.Nil(t, err)
	assert.Len(t, decision.Campaigns, 1)
}
