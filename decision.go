package decision

import (
	"encoding/base64"
	"sync"

	"github.com/flagship-io/flagship-proto/decision_response"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GetDecision return a decision response from visitor & environment infos
func GetDecision(
	visitorInfos Visitor,
	environmentInfos Environment,
	options DecisionOptions,
	handlers DecisionHandlers,
) (*decision_response.DecisionResponse, error) {

	envID := environmentInfos.ID
	visitorID := visitorInfos.ID
	anonymousID := visitorInfos.AnonymousID
	visitorContext := visitorInfos.Context
	decisionGroup := visitorInfos.DecisionGroup
	tracker := options.Tracker

	if decisionGroup != "" {
		// encode decision group if set
		decisionGroup = base64.StdEncoding.EncodeToString([]byte(decisionGroup))
	}

	// Initialize campaign response to be returned
	decisionResponse := &decision_response.DecisionResponse{}
	decisionResponse.VisitorId = wrapperspb.String(visitorID)
	decisionResponse.Campaigns = []*decision_response.Campaign{}

	// Initialize future variation groups variation assignments
	newVGAssignments := make(map[string]*VisitorCache)
	newVGAssignmentsAnonymous := make(map[string]*VisitorCache)

	// Initialize future campaign activations
	campaignActivations := []*VisitorActivation{}

	// Initialize has AB Test assigned
	hasABCampaign := false

	tracker.TimeTrack("start compute targetings")

	// 0. Deduplicate campaigns with the same ID
	logger.Logf(InfoLevel, "deduplicating campaigns by ID")
	campaignsArray := deduplicateCampaigns(environmentInfos.Campaigns)

	// 1. Get variation group for each campaign that matches visitor context
	logger.Logf(InfoLevel, "getting variation groups that match visitor ID and context")
	variationGroups := getCampaignsVG(campaignsArray, visitorID, visitorContext)
	tracker.TimeTrack("end compute targetings")

	// 2.a Check if anonymous / visitor reconciliation is enabled and relevant here
	enableReconciliation := environmentInfos.UseReconciliation && anonymousID != ""

	// 2.b Check if cache is enabled
	enableCache := isCacheEnabled(environmentInfos, variationGroups)

	// 2.c Load all cache in parallel
	var err error
	allCacheAssignments := &allVisitorAssignments{}
	if enableCache {
		tracker.TimeTrack("start find existing vID in Cache DB")
		logger.Logf(InfoLevel, "loading assignments cache from DB")
		allCacheAssignments, err = getCache(environmentInfos.ID, visitorID, anonymousID, decisionGroup, enableReconciliation, handlers.GetCache)
		tracker.TimeTrack("end find existing vID in Cache DB")

		if err != nil {
			logger.Logf(ErrorLevel, "error occured when getting cached assignments: %v", err)
			return decisionResponse, nil
		}
	}

	// 2.d Load previously assigned AB Tests to handle single assignment option
	previousVisVGsAB := []string{}
	if environmentInfos.SingleAssignment {
		previousVisVGsAB = getActivatedABVGIds(variationGroups, allCacheAssignments.Standard.getAssignments())
	}

	// 3. Compute or get from cache each variation group variation assignment
	for _, vg := range variationGroups {

		// 3.1 Skip according to single assignment rule
		if shouldSkipVG(environmentInfos, vg, previousVisVGsAB, hasABCampaign) {
			logger.Logf(DebugLevel, "Campaign %s has been skipped because of single assignment rule", vg.Campaign.ID)
			continue
		}

		// 3.2 Skip according to bucket allocation rule
		if shouldSkipBucketVG(options.EnableBucketAllocation == nil || *options.EnableBucketAllocation, visitorID, vg.Campaign) {
			logger.Logf(DebugLevel, "visitor ID %s does not fall into the campaign's buckets. Skipping campaign", visitorID)
			continue
		}

		// 3.3 Choose the variation group assigned variation
		// according to cache assignments, visitor ID and decision group and options
		chosenVariationResult, err := chooseVariation(
			visitorID,
			decisionGroup,
			vg,
			*allCacheAssignments,
			options)

		// If variation assignment failed, return the response for single campaign, other move to the next variation group
		if err != nil {
			if options.CampaignID != "" {
				return decisionResponse, err
			}
			continue
		}

		// 3.4 Add the new cache assignment for visitor and anonymous
		if chosenVariationResult.newAssignment != nil {
			newVGAssignments[vg.ID] = chosenVariationResult.newAssignment
		}
		if chosenVariationResult.newAssignmentAnonymous != nil {
			newVGAssignmentsAnonymous[vg.ID] = chosenVariationResult.newAssignmentAnonymous
		}

		// 3.5 If decision should trigger activation hit, add it to list of activations
		if options.TriggerHit {
			anonymousIDActivate := visitorID
			if enableReconciliation {
				anonymousIDActivate = anonymousID
			}
			campaignActivations = append(campaignActivations, &VisitorActivation{
				EnvironmentID:    envID,
				VisitorID:        visitorID,
				AnonymousID:      anonymousIDActivate,
				VariationGroupID: vg.ID,
				VariationID:      chosenVariationResult.chosenVariation.ID,
			})
		}

		// 3.6 Serialize campaign response and add it to the to global response campaign list
		decisionResponse.Campaigns = append(
			decisionResponse.Campaigns,
			buildCampaignResponse(vg, chosenVariationResult.chosenVariation, options.ExposeAllKeys))

		// 3.7 Remember if AB campaign for single assignment
		if vg.Campaign.Type == "ab" {
			hasABCampaign = true
		}
	}

	// 4. Handle all side effects in parallel
	var wg sync.WaitGroup

	// 4.1 Saves all assignments
	if enableCache && handlers.SaveCache != nil {
		saveCacheAssignments(&wg, handlers, envID, visitorID, "visitor ID", newVGAssignments)
		saveCacheAssignments(&wg, handlers, envID, anonymousID, "anonymous ID", newVGAssignmentsAnonymous)
		saveCacheAssignments(&wg, handlers, envID, decisionGroup, "decision group", newVGAssignments)
	}

	// 4.2 Sends all activation events
	if len(campaignActivations) > 0 && handlers.ActivateCampaigns != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.TimeTrack("start activating campaigns hit")
			logger.Logf(InfoLevel, "activating %d campaigns and variations", len(campaignActivations))
			err := handlers.ActivateCampaigns(campaignActivations)
			if err != nil {
				logger.Logf(ErrorLevel, "error occured on campaign activation: %v", err)
			}
			tracker.TimeTrack("end activating campaigns hit")
		}()
	}
	wg.Wait()

	return decisionResponse, nil
}
