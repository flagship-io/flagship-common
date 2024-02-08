package decision

import (
	"fmt"
	"maps"
	"sync"
	"testing"

	"github.com/flagship-io/flagship-common/targeting"
	"github.com/flagship-io/flagship-proto/decision_response"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

var mu = sync.Mutex{}

var campaigns = []*Campaign{
	{
		ID:           "a1",
		BucketRanges: [][]float64{{0., 100.}},
		VariationGroups: []*VariationGroup{
			{
				ID:         "vga",
				Targetings: createBoolTargeting(),
				Variations: []*Variation{
					{
						ID:         "vgav1",
						Allocation: 50,
						Modifications: &decision_response.Modifications{
							Type:  decision_response.ModificationsType_FLAG,
							Value: structpb.NewStringValue("toto1").GetStructValue(),
						},
					}, {
						ID:         "vgav2",
						Allocation: 50,
						Modifications: &decision_response.Modifications{
							Type:  decision_response.ModificationsType_FLAG,
							Value: structpb.NewStringValue("toto2").GetStructValue(),
						},
					},
				},
			},
		},
	},
	{
		ID:           "a2",
		BucketRanges: [][]float64{{20., 30.}},
		VariationGroups: []*VariationGroup{
			{
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
	mu.Lock()
	defer func() {
		mu.Unlock()
	}()
	return cache[environmentID+id], nil
}

func localSetCache(environmentID string, id string, assignment *VisitorAssignments) error {
	mu.Lock()
	cache[environmentID+id] = assignment
	mu.Unlock()
	return nil
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

	newAssignments := map[string]*VisitorCache{
		"vg_id": {
			VariationID: "v_id",
			Activated:   true,
		},
		"vg2_id": {
			VariationID: "v2_id",
			Activated:   true,
		},
	}
	newAssignmentsDG := map[string]*VisitorCache{
		"vg_id": {
			VariationID: "vdg_id",
			Activated:   true,
		},
		"vg3_id": {
			VariationID: "v3_id",
			Activated:   true,
		},
	}

	cache[envID+decisionGroup] = &VisitorAssignments{
		Assignments: newAssignmentsDG,
	}

	assignments, err = getCache(envID, visitorID, anonymousID, decisionGroup, false, localGetCache)
	assert.Nil(t, err)
	assert.EqualValues(t, newAssignmentsDG, assignments.DecisionGroup.getAssignments())
	assert.Nil(t, assignments.Standard)
	assert.Nil(t, assignments.Anonymous)

	cache[envID+visitorID] = &VisitorAssignments{
		Assignments: maps.Clone(newAssignments),
	}

	assignments, err = getCache(envID, visitorID, anonymousID, decisionGroup, false, localGetCache)
	assert.Nil(t, err)
	assert.EqualValues(t, newAssignmentsDG, assignments.DecisionGroup.getAssignments())
	assert.EqualValues(t, newAssignments, assignments.Standard.getAssignments())
	assert.Nil(t, assignments.Anonymous)

	cache[envID+anonymousID] = &VisitorAssignments{
		Assignments: maps.Clone(newAssignments),
	}

	assignments, err = getCache(envID, visitorID, anonymousID, decisionGroup, true, localGetCache)
	assert.Nil(t, err)
	assert.EqualValues(t, newAssignmentsDG, assignments.DecisionGroup.getAssignments())
	assert.EqualValues(t, newAssignments, assignments.Standard.getAssignments())
	assert.EqualValues(t, newAssignments, assignments.Anonymous.getAssignments())
}

func TestDecisionCache(t *testing.T) {
	vi := Visitor{}
	vi.ID = "v1"
	vi.Context = &targeting.Context{
		Standard: targeting.ContextMap{
			"isVIP": structpb.NewBoolValue(true),
		},
	}

	ei := Environment{}
	ei.ID = "e123"
	ei.Campaigns = campaigns
	for _, vg := range ei.Campaigns[0].VariationGroups {
		vg.Campaign = ei.Campaigns[0]
	}

	options := DecisionOptions{
		TriggerHit: true,
	}
	// no options

	handlers := DecisionHandlers{}
	handlers.GetCache = mockGetCache
	handlers.SaveCache = mockSaveCache
	handlers.ActivateCampaigns = mockActivateCampaigns

	decision, err := GetDecision(vi, ei, options, handlers)

	// check that campaign matching visitor is returned. Also check that the second variation is set
	assert.Nil(t, err)
	assert.Len(t, decision.Campaigns, 1)
	assert.Equal(t, "vgav2", decision.Campaigns[0].Variation.Id.Value)

	vi.DecisionGroup = "dg"
	decision, err = GetDecision(vi, ei, options, handlers)

	// check that campaign matching visitor is returned. Also check that the second variation is set
	assert.Nil(t, err)
	assert.Len(t, decision.Campaigns, 1)
	assert.Equal(t, "vgav1", decision.Campaigns[0].Variation.Id.Value)

	// change the allocation so that visitor should change variation if the cache is disabled
	ei.Campaigns[0].VariationGroups[0].Variations[0].Allocation = 10
	ei.Campaigns[0].VariationGroups[0].Variations[1].Allocation = 90

	decision, err = GetDecision(vi, ei, options, handlers)

	// check that campaign matching visitor is returned. Also check that the first variation is not chosen
	assert.Nil(t, err)
	assert.Len(t, decision.Campaigns, 1)
	assert.Equal(t, "vgav2", decision.Campaigns[0].Variation.Id.Value)

	// Reset the allocations
	ei.Campaigns[0].VariationGroups[0].Variations[0].Allocation = 50
	ei.Campaigns[0].VariationGroups[0].Variations[1].Allocation = 50

	// Set "real" local cache to persist visitor allocation
	handlers.GetCache = localGetCache
	handlers.SaveCache = localSetCache
	ei.CacheEnabled = true
	decision, _ = GetDecision(vi, ei, options, handlers)

	// check that campaign matching visitor is returned. Also check that the first variation is not chosen
	assert.Equal(t, decision.Campaigns[0].Variation.Id.Value, "vgav1")

	// change the allocation so that visitor should change variation if the cache is disabled
	ei.Campaigns[0].VariationGroups[0].Variations[0].Allocation = 10
	ei.Campaigns[0].VariationGroups[0].Variations[1].Allocation = 90

	decision, _ = GetDecision(vi, ei, options, handlers)

	// check that campaign matching visitor is returned. Also check that the first variation is not chosen
	assert.Equal(t, decision.Campaigns[0].Variation.Id.Value, "vgav1")

	// Reset the allocations
	ei.Campaigns[0].VariationGroups[0].Variations[0].Allocation = 50
	ei.Campaigns[0].VariationGroups[0].Variations[1].Allocation = 50
}

func TestDecisionBucketInNoCache(t *testing.T) {
	vi := Visitor{}
	vi.ID = "v1"
	vi.Context = &targeting.Context{
		Standard: targeting.ContextMap{
			"isVIP": structpb.NewBoolValue(true),
		},
	}
	vi.DecisionGroup = "decision"

	ei := Environment{}
	ei.ID = "e123"
	ei.Campaigns = campaigns
	for _, vg := range ei.Campaigns[0].VariationGroups {
		vg.Campaign = ei.Campaigns[0]
	}
	for _, vg := range ei.Campaigns[1].VariationGroups {
		vg.Campaign = ei.Campaigns[1]
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

func TestVisitorShouldNotBeAssignedWhenVariationDeleted(t *testing.T) {
	vi := Visitor{}
	vi.ID = "v1"
	vi.Context = &targeting.Context{
		Standard: targeting.ContextMap{
			"isVIP": structpb.NewBoolValue(true),
		},
	}

	ei := Environment{}
	ei.ID = "e123"
	ei.CacheEnabled = true
	ei.Campaigns = campaigns
	for _, vg := range ei.Campaigns[0].VariationGroups {
		vg.Campaign = ei.Campaigns[0]
	}

	options := DecisionOptions{
		TriggerHit: true,
	}
	// no options

	handlers := DecisionHandlers{}
	handlers.GetCache = mockGetCache
	handlers.SaveCache = mockSaveCache
	handlers.ActivateCampaigns = mockActivateCampaigns

	// delete variation and check that visitor is not returned
	campaignVars := ei.Campaigns[0].VariationGroups[0].Variations
	ei.Campaigns[0].VariationGroups[0].Variations = campaignVars[1:]

	decision, _ := GetDecision(vi, ei, options, handlers)

	assert.Len(t, decision.Campaigns, 0)
}
