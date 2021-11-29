package decision

import (
	"fmt"
	"testing"

	"github.com/flagship-io/flagship-proto/decision_response"
	"github.com/flagship-io/flagship-proto/targeting"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

func targetingContext() *targeting.Targeting {
	return &targeting.Targeting{
		TargetingGroups: []*targeting.Targeting_TargetingGroup{
			{
				Targetings: []*targeting.Targeting_InnerTargeting{
					{
						Operator: targeting.Targeting_EQUALS,
						Key:      wrapperspb.String("test"),
						Value:    structpb.NewStringValue("decision"),
					},
				},
			},
		},
	}
}

func TestDecisionBucketInNoCache(t *testing.T) {
	vi := VisitorInfo{}
	vi.ID = "v1"
	vi.Context = map[string]*structpb.Value{
		"test": structpb.NewStringValue("decision"),
	}
	vi.DecisionGroup = "decision"

	ei := EnvironmentInfo{}
	ei.ID = "e123"
	ei.Campaigns = map[string]*CampaignInfo{
		"a": {
			ID:           "c1",
			BucketRanges: [][]float64{{0., 100.}},
			VariationsGroups: map[string]*VariationsGroup{
				"vga": {
					ID:         "vga",
					CampaignID: "a",
					Targetings: targetingContext(),
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
			ID:           "c1",
			BucketRanges: [][]float64{{40., 60.}},
			VariationsGroups: map[string]*VariationsGroup{
				"vgb": {
					ID:         "vgb",
					CampaignID: "b",
					Targetings: targetingContext(),
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

	fmt.Println(decision)
	fmt.Println(err)

	assert.Nil(t, err)
	assert.NotEmpty(t, decision.Campaigns)
}
