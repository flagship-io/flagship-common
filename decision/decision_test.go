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
					CampaignID: "a",
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
					CampaignID: "b",
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
