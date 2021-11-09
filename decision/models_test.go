package decision

import (
	"testing"
	"time"

	"github.com/flagship-io/flagship-proto/targeting"
	protoStruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func TestGetCampaignsArray(t *testing.T) {
	campaignsMap := map[string]*CampaignInfo{}
	campaignsMap["testNEW"] = &CampaignInfo{
		ID:        "testIDNEW",
		CreatedAt: time.Now(),
	}
	campaignsMap["testOLD"] = &CampaignInfo{
		ID:        "testIDOLD",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	campaignsMap["testMIDDLE"] = &CampaignInfo{
		ID:        "testIDMIDDLE",
		CreatedAt: time.Now().Add(-30 * time.Minute),
	}
	campaignsMap["testDUPLICATE"] = &CampaignInfo{
		ID:        "testIDNEW",
		CreatedAt: time.Now(),
	}

	campaigns := GetCampaignsArray(campaignsMap)

	if campaigns[0].ID != campaignsMap["testNEW"].ID {
		t.Errorf("Expected newest campaign first, got %v", campaigns[0].ID)
	}

	if campaigns[1] != campaignsMap["testMIDDLE"] {
		t.Errorf("Expected middle campaign in middle, got %v", campaigns[1].ID)
	}

	if campaigns[2].ID != campaignsMap["testOLD"].ID {
		t.Errorf("Expected oldest campaign last, got %v", campaigns[2].ID)
	}

	if len(campaigns) != 3 {
		t.Errorf("Expected 3 campaigns, got %v", len(campaigns))
	}
}

func TestGetVariationGroup(t *testing.T) {
	vgs := map[string]*VariationsGroup{}

	targetingGroups := []*targeting.Targeting_TargetingGroup{}
	targetingGroups = append(targetingGroups, &targeting.Targeting_TargetingGroup{
		Targetings: []*targeting.Targeting_InnerTargeting{&targeting.Targeting_InnerTargeting{
			Operator: targeting.Targeting_EQUALS,
			Key:      &wrappers.StringValue{Value: "age"},
			Value: &protoStruct.Value{
				Kind: &protoStruct.Value_NumberValue{
					NumberValue: 30,
				},
			},
		}},
	})
	targeting := &targeting.Targeting{TargetingGroups: targetingGroups}

	vgs["testVGIDNEW"] = &VariationsGroup{
		CampaignID: "testCampaignIdNEW",
		ID:         "testId",
		Targetings: targeting,
		CreatedAt:  time.Now(),
	}

	vgs["testVGIDOLD"] = &VariationsGroup{
		CampaignID: "testCampaignIdOLD",
		ID:         "testId",
		Targetings: targeting,
		CreatedAt:  time.Now().Add(-30 * time.Minute),
	}

	context := map[string]*protoStruct.Value{
		"age": &protoStruct.Value{
			Kind: &protoStruct.Value_NumberValue{
				NumberValue: 30,
			},
		},
	}

	vg := GetVariationGroup(vgs, "testVID", context)

	if vg != vgs["testVGIDOLD"] {
		t.Errorf("Expected vg olg to match, got %v", vg)
	}
}
