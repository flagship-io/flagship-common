package decision

import (
	"errors"

	"github.com/flagship-io/flagship-common/internal/utils"
	"github.com/flagship-io/flagship-common/targeting"
	"github.com/flagship-io/flagship-proto/decision_response"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ChosenVariationResult struct {
	chosenVariation        *Variation
	newAssignment          *VisitorCache
	newAssignmentAnonymous *VisitorCache
}

// isCacheEnabled is true if environments config enables it,
// and if 1vis1test or XP-C is enabled or at least one campaign as multiple variations,
func isCacheEnabled(environmentInfos Environment, variationGroups []*VariationGroup) bool {
	hasMultipleVariations := false
	for _, vg := range variationGroups {
		if len(vg.Variations) > 1 {
			hasMultipleVariations = true
			break
		}
	}
	// Enable caching if customer package allows it,
	// and if 1vis1test or XP-C is enabled or at least one campaign as multiple variations,
	return environmentInfos.CacheEnabled && (hasMultipleVariations || environmentInfos.SingleAssignment || environmentInfos.UseReconciliation)
}

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

// getActivatedABVGIds returns previously assigned AB test campaigns for visitor
func getActivatedABVGIds(variationGroups []*VariationGroup, existingVar map[string]*VisitorCache) []string {
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

// shouldSkipVG returns true if the current variation should be skiped according to single assignment rule
func shouldSkipVG(environmentInfos Environment, vg *VariationGroup, previousVisVGsAB []string, hasABCampaign bool) bool {
	return environmentInfos.SingleAssignment && vg.Campaign.Type == "ab" &&
		(len(previousVisVGsAB) > 0 && !utils.IsInStringArray(vg.ID, previousVisVGsAB) || hasABCampaign)
}

// shouldSkipBucketVG returns true if the variation group should be skipped according to bucket allocation rule
func shouldSkipBucketVG(enableBucketAllocation bool, visitorID string, campaign *Campaign) bool {
	if enableBucketAllocation {
		isInBucket, err := isVisitorInBucket(visitorID, campaign)
		if err != nil {
			logger.Logf(WarnLevel, "error on bucket allocation for campaign %v: %v", campaign.ID, err)
		}

		if !isInBucket {
			logger.Logf(DebugLevel, "visitor ID %s does not fall into the campaign's buckets. Skipping campaign", visitorID)
			return true
		}
	}
	return false
}

// selectNewVariation selects a variation according the visitor ID or decision group, the variation group and decision options
func selectNewVariation(visitorID string, decisionGroup string, vg *VariationGroup, options DecisionOptions) (*Variation, error) {
	chosenVariation, err := getRandomAllocation(visitorID, decisionGroup, vg, options.IsCumulativeAlloc)
	if err != nil {
		if err == VisitorNotTrackedError {
			logger.Logf(InfoLevel, err.Error())
		} else {
			logger.Logf(WarnLevel, "error on new allocation : %v", err)
		}
		return nil, err
	}
	return chosenVariation, err
}

func chooseVariation(
	visitorID string,
	decisionGroup string,
	vg *VariationGroup,
	allCacheAssignments allVisitorAssignments,
	options DecisionOptions) (*ChosenVariationResult, error) {

	existingAssignment, ok := allCacheAssignments.Standard.getAssignment(vg.ID)
	existingAssignmentAnonymous, okAnonymous := allCacheAssignments.Anonymous.getAssignment(vg.ID)
	existingAssignmentDecisionGroup, okDecisionGroup := allCacheAssignments.DecisionGroup.getAssignment(vg.ID)

	var existingVariation *Variation
	var existingAnonymousVariation *Variation
	var existingDecisionGroupVariation *Variation

	var newAssignment *VisitorCache
	var newAssignmentAnonymous *VisitorCache

	if ok || okAnonymous || okDecisionGroup {
		for _, v := range vg.Variations {
			if ok && v.ID == existingAssignment.VariationID {
				existingVariation = v
			}
			if okAnonymous && v.ID == existingAssignmentAnonymous.VariationID {
				existingAnonymousVariation = v
			}
			if okDecisionGroup && v.ID == existingAssignmentDecisionGroup.VariationID {
				existingDecisionGroupVariation = v
			}
		}

		// If decision group variation is found, override the cached variation
		if existingDecisionGroupVariation != nil {
			existingVariation = existingDecisionGroupVariation
		}

		// Variation has been deleted
		if existingVariation == nil && existingAnonymousVariation == nil {
			logger.Logf(DebugLevel, "visitor ID %s was already assigned to deleted variation ID %s", visitorID, existingAssignment.VariationID)
			return nil, errors.New("visitor ID assigned to deleted variation")
		}
	}

	var isNew, isNewAnonymous bool
	var chosenVariation *Variation
	var err error

	// If already has variation && assigned variation ID  exist, visitor should not be re-assigned
	if existingVariation != nil {
		logger.Logf(DebugLevel, "visitor already assigned to variation ID %s", existingVariation.ID)
		chosenVariation = existingVariation
	} else if existingAnonymousVariation != nil {
		// If reconciliation is on, find anonymous variation as set vid to that variation ID
		logger.Logf(DebugLevel, "anonymous ID already assigned to variation ID %s", existingAnonymousVariation.ID)
		chosenVariation = existingAnonymousVariation
		isNew = true
	} else {
		// Else compute new allocation
		logger.Logf(DebugLevel, "assigning visitor ID to new variation")
		chosenVariation, err = selectNewVariation(visitorID, decisionGroup, vg, options)
		if err != nil {
			return nil, err
		}
		logger.Logf(DebugLevel, "visitor ID %s got assigned to variation ID %s", visitorID, chosenVariation.ID)
		isNew = true
		isNewAnonymous = true
	}

	// 3.1 If allocation is newly computed and not only 1 variation,
	// or if campaign activation not saved and should be
	// tag this vg alloc to be saved
	if options.TriggerHit && !(ok && existingAssignment.Activated) || isNew {
		newAssignment = &VisitorCache{
			VariationID: chosenVariation.ID,
			Activated:   options.TriggerHit,
		}
	}

	// 3.1bis If anonymous allocation is newly computed and not only 1 variation,
	// or if campaign activation not saved and should be
	// tag this vg alloc to be saved
	if options.TriggerHit && !(okAnonymous && existingAssignmentAnonymous.Activated) || isNewAnonymous {
		newAssignmentAnonymous = &VisitorCache{
			VariationID: chosenVariation.ID,
			Activated:   options.TriggerHit,
		}
	}

	return &ChosenVariationResult{
		chosenVariation:        chosenVariation,
		newAssignment:          newAssignment,
		newAssignmentAnonymous: newAssignmentAnonymous,
	}, nil
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
