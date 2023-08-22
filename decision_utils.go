package decision

import (
	"github.com/flagship-io/flagship-common/targeting"
	"github.com/flagship-io/flagship-proto/decision_response"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// deduplicateCampaigns returns the first campaign that matches the campaign ID
func deduplicateCampaigns(campaigns []*Campaign) []*Campaign {
	// Use ID array to sort variation groups by ID to force order of iteration
	cArray := []*Campaign{}
	cIDs := map[string]bool{}
	for _, c := range campaigns {
		if _, ok := cIDs[c.ID]; !ok {
			cArray = append(cArray, c)
		}
		cIDs[c.ID] = true

	}
	return cArray
}

// getVariationGroup returns the first variationGroup that matches the visitorId and context
func getVariationGroup(variationGroups []*VariationGroup, visitorID string, context *targeting.Context) *VariationGroup {
	for _, variationGroup := range variationGroups {
		match, err := targetingMatch(variationGroup.Targetings, visitorID, context)
		if err != nil {
			logger.Logf(WarnLevel, "targeting match error variationGroupId %s, user %s: %s", variationGroup.ID, visitorID, err)
		}
		if match {
			return variationGroup
		}
	}
	return nil
}

// getCampaignsVG returns the variation groups that target visitor
func getCampaignsVG(campaigns []*Campaign, visitorID string, context *targeting.Context) []*VariationGroup {
	campaignVG := []*VariationGroup{}
	existingCampaignVG := make(map[string]bool)
	for _, campaign := range campaigns {
		_, ok := existingCampaignVG[campaign.ID]
		if ok {
			// Campaign already handled (maybe because of custom ID)
			continue
		}

		vg := getVariationGroup(campaign.VariationGroups, visitorID, context)

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
func getPreviousABVGIds(variationGroups []*VariationGroup, existingVar map[string]*VisitorCache) []string {
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
func buildCampaignResponse(vg *VariationGroup, variation *Variation, exposeAllKeys bool) *decision_response.Campaign {
	campaignResponse := decision_response.Campaign{
		Id:                 wrapperspb.String(vg.Campaign.ID),
		Name:               wrapperspb.String(vg.Campaign.Name),
		VariationGroupId:   wrapperspb.String(vg.ID),
		VariationGroupName: wrapperspb.String(vg.Name),
	}
	if vg.Campaign.Slug != nil {
		campaignResponse.Slug = wrapperspb.String(*vg.Campaign.Slug)
	}

	if exposeAllKeys {
		logger.Logf(DebugLevel, "filling non existant keys in variation with null value")
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
	} else {
		if variation.Modifications != nil && variation.Modifications.Value != nil {
			for key, val := range variation.Modifications.Value.Fields {
				_, okCast := val.GetKind().(*structpb.Value_NullValue)
				if okCast {
					// Remove nil value keys if shouldFillKeys is false
					delete(variation.Modifications.Value.Fields, key)
				}
			}
		}
	}

	protoModif := &decision_response.Variation{
		Id:            wrapperspb.String(variation.ID),
		Name:          wrapperspb.String(variation.Name),
		Modifications: variation.Modifications,
		Reference:     variation.Reference,
	}

	campaignResponse.Variation = protoModif
	campaignResponse.Type = wrapperspb.String(vg.Campaign.Type)
	return &campaignResponse
}
