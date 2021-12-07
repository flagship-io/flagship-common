package decision

import (
	"fmt"
	"log"
	"sort"

	"github.com/flagship-io/flagship-proto/decision_response"
	protoStruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GetCampaignsArray returns the first campaign that matches the campaign ID
func GetCampaignsArray(campaigns map[string]*CampaignInfo) []*CampaignInfo {
	// Use ID array to sort variation groups by ID to force order of iteration
	cArray := []*CampaignInfo{}
	cIDs := map[string]bool{}
	for _, c := range campaigns {
		if _, ok := cIDs[c.ID]; !ok {
			cArray = append(cArray, c)
		}
		cIDs[c.ID] = true

	}
	sort.Sort(byCreatedAtCampaigns(cArray))
	return cArray
}

// GetVariationGroup returns the first variationGroup that matches the visitorId and context
func GetVariationGroup(variationGroups map[string]*VariationsGroup, visitorID string, context map[string]*protoStruct.Value) *VariationsGroup {
	// Use ID array to sort variation groups by ID to force order of iteration
	vgArray := []*VariationsGroup{}
	for _, vg := range variationGroups {
		vgArray = append(vgArray, vg)
	}
	sort.Sort(byCreatedAtVG(vgArray))

	for _, variationGroup := range vgArray {
		match, err := TargetingMatch(variationGroup, visitorID, context)
		if err != nil {
			log.Println(fmt.Sprintf("Targeting match error variationGroupId %s, user %s: %s", variationGroup.ID, visitorID, err))
		}
		if match {
			return variationGroup
		}
	}
	return nil
}

// GetCampaignsVG returns the variation groups that target visitor
func GetCampaignsVG(campaigns []*CampaignInfo, visitorID string, context map[string]*structpb.Value) []*VariationsGroup {
	campaignVG := []*VariationsGroup{}
	existingCampaignVG := make(map[string]bool)
	for _, campaign := range campaigns {
		_, ok := existingCampaignVG[campaign.ID]
		if ok {
			// Campaign already handled (maybe because of custom ID)
			continue
		}

		vg := GetVariationGroup(campaign.VariationsGroups, visitorID, context)

		if vg == nil {
			continue
		}
		vg.Campaign = campaign
		existingCampaignVG[campaign.ID] = true
		campaignVG = append(campaignVG, vg)
	}
	return campaignVG
}

// getPreviousABVGIds returns previously assigned AB test campaigns for visitor
func getPreviousABVGIds(variationGroups []*VariationsGroup, existingVar map[string]*VisitorVGCacheItem) []string {
	previousVisVGsAB := []string{}
	alreadyAdded := map[string]bool{}
	for _, vg := range variationGroups {
		if vg.Campaign.Type != "ab" {
			continue
		}
		existingVariations, ok := existingVar[vg.ID]
		_, added := alreadyAdded[vg.ID]
		if ok && existingVariations.Activated && !added {
			previousVisVGsAB = append(previousVisVGsAB, vg.ID)
			alreadyAdded[vg.ID] = true
		}
	}
	return previousVisVGsAB
}

// buildCampaignResponse creates a decision campaign response, filling out empty flag keys for each variation if needed
func buildCampaignResponse(vg *VariationsGroup, variation *Variation, shouldFillKeys bool) *decision_response.Campaign {
	campaignResponse := decision_response.Campaign{
		Id: &wrappers.StringValue{
			Value: vg.Campaign.ID,
		},
		VariationGroupId: &wrappers.StringValue{
			Value: vg.ID,
		},
	}

	if shouldFillKeys {
		if variation.Modifications == nil {
			variation.Modifications = &decision_response.Modifications{}
		}
		if variation.Modifications.Value == nil {
			variation.Modifications.Value = &structpb.Struct{}
		}
		if variation.Modifications.Value.Fields == nil {
			variation.Modifications.Value.Fields = map[string]*structpb.Value{}
		}
		for _, v := range vg.Variations {
			if v.Modifications != nil && v.Modifications.Value != nil && v.Modifications.Value.Fields != nil {
				for key := range v.Modifications.Value.Fields {
					if _, ok := variation.Modifications.Value.Fields[key]; !ok {
						variation.Modifications.Value.Fields[key] = &structpb.Value{Kind: &structpb.Value_NullValue{}}
					}
				}
			}
		}
	}

	protoModif := &decision_response.Variation{
		Id: &wrappers.StringValue{
			Value: variation.ID,
		},
		Modifications: variation.Modifications,
		Reference:     variation.Reference,
	}

	campaignResponse.Variation = protoModif
	campaignResponse.Type = wrapperspb.String(vg.Campaign.Type)
	return &campaignResponse
}
