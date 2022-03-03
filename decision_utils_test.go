package decision

import (
	"testing"
	"time"

	"github.com/flagship-io/flagship-common/internal/targeting"
	"github.com/flagship-io/flagship-proto/decision_response"
	protoTargeting "github.com/flagship-io/flagship-proto/targeting"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func createNumberTargeting() *protoTargeting.Targeting {
	targetingGroups := []*protoTargeting.Targeting_TargetingGroup{}
	targetingGroups = append(targetingGroups, &protoTargeting.Targeting_TargetingGroup{
		Targetings: []*protoTargeting.Targeting_InnerTargeting{{
			Operator: protoTargeting.Targeting_EQUALS,
			Key:      &wrappers.StringValue{Value: "age"},
			Value:    structpb.NewNumberValue(30),
		}},
	})
	return &protoTargeting.Targeting{TargetingGroups: targetingGroups}
}

func createBoolTargeting() *protoTargeting.Targeting {
	targetingGroups := []*protoTargeting.Targeting_TargetingGroup{}
	targetingGroups = append(targetingGroups, &protoTargeting.Targeting_TargetingGroup{
		Targetings: []*protoTargeting.Targeting_InnerTargeting{{
			Operator: protoTargeting.Targeting_EQUALS,
			Key:      &wrappers.StringValue{Value: "isVIP"},
			Value:    structpb.NewBoolValue(true),
		}},
	})
	return &protoTargeting.Targeting{TargetingGroups: targetingGroups}
}

func TestDeduplicateCampaigns(t *testing.T) {
	campaignsArray := []*Campaign{&Campaign{
		ID:        "testIDNEW",
		CreatedAt: time.Now(),
	}, &Campaign{
		ID:        "testIDOLD",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}, &Campaign{
		ID:        "testIDMIDDLE",
		CreatedAt: time.Now().Add(-30 * time.Minute),
	}, &Campaign{
		ID:        "testIDNEW",
		CreatedAt: time.Now(),
	}}

	campaigns := deduplicateCampaigns(campaignsArray)

	if len(campaigns) != 3 {
		t.Errorf("Expected 3 campaigns, got %v", len(campaigns))
	}
}

func TestGetPreviousABVGIds(t *testing.T) {
	vgs := []*VariationGroup{{
		Campaign: &Campaign{
			ID:   "testCampaignId1",
			Type: "ab",
		},
		ID:        "testId1",
		CreatedAt: time.Now(),
	}, {
		Campaign: &Campaign{
			ID:   "testCampaignId2",
			Type: "flag",
		},
		ID:        "testId2",
		CreatedAt: time.Now(),
	}, {
		Campaign: &Campaign{
			ID:   "testCampaignId3",
			Type: "ab",
		},
		ID:        "testId1",
		CreatedAt: time.Now(),
	}, {
		Campaign: &Campaign{
			ID:   "testCampaignId4",
			Type: "ab",
		},
		ID:        "testId3",
		CreatedAt: time.Now(),
	}, {
		Campaign: &Campaign{
			ID:   "testCampaignId5",
			Type: "ab",
		},
		ID:        "testId4",
		CreatedAt: time.Now(),
	}}

	existingAssignments := map[string]*VisitorCache{
		"testId1": {
			VariationID: "vid1",
			Activated:   true,
		},
		"testId2": {
			VariationID: "vid1",
			Activated:   true,
		},
		"testId3": {
			VariationID: "vid2",
			Activated:   false,
		},
		"testId4": {
			VariationID: "vid2",
			Activated:   true,
		},
	}

	previousVGIds := getPreviousABVGIds(vgs, existingAssignments)
	assert.EqualValues(t, []string{"testId1", "testId4"}, previousVGIds)
}

func TestGetCampaignsVG(t *testing.T) {
	vg1 := &VariationGroup{
		Campaign: &Campaign{
			ID: "testCampaignId",
		},
		ID:         "testId",
		Targetings: createNumberTargeting(),
		CreatedAt:  time.Now(),
	}
	vgs := []*VariationGroup{vg1}

	vgNotTargeted := &VariationGroup{
		Campaign: &Campaign{
			ID: "testCampaignIdNotTargeted",
		},
		ID:         "testIdNotTargeted",
		Targetings: createBoolTargeting(),
		CreatedAt:  time.Now(),
	}
	vgsNotTargeted := []*VariationGroup{vgNotTargeted}

	context := targeting.Context{
		Rows: targeting.ContextRows{
			"age": structpb.NewNumberValue(30),
		},
	}
	campaignInfos := []*Campaign{
		{
			ID:              "testCampaignId",
			VariationGroups: vgs,
		},
		{
			ID:              "testCampaignId",
			VariationGroups: vgs,
		},
		{
			ID:              "testCampaignIdNotTargeted",
			VariationGroups: vgsNotTargeted,
		},
	}
	vgsResp := getCampaignsVG(campaignInfos, "testVID", context)
	assert.Equal(t, vg1, vgsResp[0])
	assert.Equal(t, 1, len(vgsResp))
}

func TestBuildCampaignResponse(t *testing.T) {
	value1, _ := structpb.NewStruct(map[string]interface{}{
		"bool":   true,
		"string": "string",
		"number": 1,
	})

	value2, _ := structpb.NewStruct(map[string]interface{}{
		"bool2": true,
	})

	var1 := &Variation{
		ID: "vaid1",
		Modifications: &decision_response.Modifications{
			Value: value1,
		},
	}

	var2 := &Variation{
		ID: "vaid2",
		Modifications: &decision_response.Modifications{
			Value: value2,
		},
	}

	var3 := &Variation{
		ID:            "vaid3",
		Modifications: nil,
	}

	vg := &VariationGroup{
		Campaign: &Campaign{
			ID:   "cid",
			Type: "ab",
		},
		ID:         "vgid",
		Variations: []*Variation{var1, var2},
	}

	resp := buildCampaignResponse(vg, var1, false)
	assert.NotNil(t, resp)
	assert.Equal(t, "cid", resp.Id.Value)
	assert.Equal(t, "ab", resp.Type.Value)
	assert.Equal(t, "vgid", resp.VariationGroupId.Value)
	assert.Equal(t, value1, resp.Variation.Modifications.Value)

	resp = buildCampaignResponse(vg, var1, true)
	assert.NotNil(t, resp)
	assert.Equal(t, "cid", resp.Id.Value)
	assert.Equal(t, "ab", resp.Type.Value)
	assert.Equal(t, "vgid", resp.VariationGroupId.Value)
	assert.Equal(t, 4, len(resp.Variation.Modifications.Value.Fields))
	assert.EqualValues(t, map[string]interface{}{
		"bool":   true,
		"string": "string",
		"number": 1.,
		"bool2":  nil,
	}, resp.Variation.Modifications.Value.AsMap())

	resp = buildCampaignResponse(vg, var3, true)
	assert.NotNil(t, resp)
	assert.Equal(t, "cid", resp.Id.Value)
	assert.Equal(t, "ab", resp.Type.Value)
	assert.Equal(t, "vgid", resp.VariationGroupId.Value)
	assert.Equal(t, 4, len(resp.Variation.Modifications.Value.Fields))
	assert.EqualValues(t, map[string]interface{}{
		"bool":   nil,
		"string": nil,
		"number": nil,
		"bool2":  nil,
	}, resp.Variation.Modifications.Value.AsMap())
}
