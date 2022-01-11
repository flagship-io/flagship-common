package decision

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/flagship-io/flagship-common/utils"
	"github.com/flagship-io/flagship-proto/decision_response"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type AssignmentResult struct {
	result      *VisitorAssignments
	visitorType string
	err         error
}

type AllVisitorAssignments struct {
	Standard  *VisitorAssignments
	Anonymous *VisitorAssignments
}

func getCache(
	environmentID string,
	visitorID string,
	anonymousID string,
	decisionGroup string,
	enableReconciliation bool,
	getCacheHandler func(environmentID string, id string) (*VisitorAssignments, error)) (*AllVisitorAssignments, error) {

	cacheChan := make(chan (*AssignmentResult))
	allAssignments := &AllVisitorAssignments{
		Standard: &VisitorAssignments{
			Assignments: map[string]*VisitorVGCacheItem{},
		},
	}

	var err error

	go func(c chan (*AssignmentResult)) {
		newAssignments, err := getCacheHandler(environmentID, visitorID)
		c <- &AssignmentResult{
			result:      newAssignments,
			visitorType: "standard",
			err:         err,
		}
	}(cacheChan)

	go func(c chan (*AssignmentResult)) {
		var newAssignmentsAnonymous *VisitorAssignments
		var err error
		if enableReconciliation {
			newAssignmentsAnonymous, err = getCacheHandler(environmentID, anonymousID)
		}
		c <- &AssignmentResult{
			result:      newAssignmentsAnonymous,
			visitorType: "anonymous",
			err:         err,
		}
	}(cacheChan)

	go func(c chan (*AssignmentResult)) {
		var newAssignmentsDG *VisitorAssignments
		var err error
		if decisionGroup != "" {
			newAssignmentsDG, err = getCacheHandler(environmentID, decisionGroup)
		}
		c <- &AssignmentResult{
			result:      newAssignmentsDG,
			visitorType: "decisionGroup",
			err:         err,
		}
	}(cacheChan)

	assignmentsDG := &VisitorAssignments{}
	for i := 0; i < 3; i++ {
		r := <-cacheChan
		switch r.visitorType {
		case "standard":
			allAssignments.Standard = r.result
		case "anonymous":
			allAssignments.Anonymous = r.result
		case "decisionGroup":
			assignmentsDG = r.result
		}
		err = r.err
	}

	// Assign decision group assignments to visitor
	if allAssignments.Standard.GetAssignments() != nil {
		for k, v := range assignmentsDG.GetAssignments() {
			allAssignments.Standard.Assignments[k] = v
		}
	}

	return allAssignments, err
}

// GetDecision return a decision response from visitor & environment infos
func GetDecision(
	visitorInfos VisitorInfo,
	environmentInfos EnvironmentInfo,
	options DecisionOptions,
	handlers DecisionHandlers,
) (*decision_response.DecisionResponse, error) {

	envID := environmentInfos.ID
	visitorID := visitorInfos.ID
	anonymousID := visitorInfos.AnonymousID
	visitorContext := visitorInfos.Context
	decisionGroup := visitorInfos.DecisionGroup
	tracker := options.Tracker
	triggerHit := options.TriggerHit

	if decisionGroup != "" {
		// 1bis. Compute decision group if set
		decisionGroup = envID + ":" + base64.StdEncoding.EncodeToString([]byte(decisionGroup))
	}

	decisionResponse := &decision_response.DecisionResponse{}
	decisionResponse.VisitorId = wrapperspb.String(visitorID)
	decisionResponse.Campaigns = []*decision_response.Campaign{}

	newVGAssignments := make(map[string]*VisitorVGCacheItem)
	newVGAssignmentsAnonymous := make(map[string]*VisitorVGCacheItem)

	tracker.TimeTrack("Start compute targetings")

	// 0. Order campaigns by created at
	campaignsArray := GetCampaignsArray(environmentInfos.Campaigns)

	// 1. Get variation group for each campaign that matches visitor context
	variationGroups := GetCampaignsVG(campaignsArray, visitorID, visitorContext)
	tracker.TimeTrack("End compute targetings")

	// 2. Get cache variation assignments from cache DB
	tracker.TimeTrack("Start find existing vID in Cache DB")

	enableReconciliation := environmentInfos.UseReconciliation && anonymousID != ""
	hasMultipleVariations := false
	for _, vg := range variationGroups {
		if len(vg.Variations) > 1 {
			hasMultipleVariations = true
			break
		}
	}
	// Enable caching if customer package allows it,
	// and if 1vis1test or XP-C is enabled or at least one campaign as multiple variations,
	enableCache := environmentInfos.CacheEnabled && (hasMultipleVariations || environmentInfos.SingleAssignment || environmentInfos.UseReconciliation)

	// 2.a Get visitor assignments asynchronously

	// 2.c Run parallel execution
	allCacheAssignments := &AllVisitorAssignments{}
	var err error
	if enableCache {
		allCacheAssignments, err = getCache(environmentInfos.ID, visitorID, anonymousID, decisionGroup, enableReconciliation, handlers.GetCache)
	}

	tracker.TimeTrack("End find existing vID in Cache DB")

	if err != nil {
		log.Println("CacheDB error occured : " + err.Error())
		return decisionResponse, nil
	}

	// Handle single assignment clients
	previousVisVGsAB := []string{}
	if environmentInfos.SingleAssignment {
		previousVisVGsAB = getPreviousABVGIds(variationGroups, allCacheAssignments.Standard.GetAssignments())
	}

	// Initialize future campaign activations
	cActivations := []*VisitorActivation{}

	// Initialize has AB Test deployed
	hasABCampaign := false

	// 3. Compute or get from cache each variation group  variation affectation
	for _, vg := range variationGroups {

		if vg.Campaign == nil {
			return nil, errors.New("variation group should have a campaign")
		}

		// Handle single assignment clients
		if environmentInfos.SingleAssignment && vg.Campaign.Type == "ab" {
			if len(previousVisVGsAB) > 0 && !utils.IsInStringArray(vg.ID, previousVisVGsAB) {
				// Visitor has already been assigned to a variation
				continue
			}

			if hasABCampaign && len(previousVisVGsAB) == 0 {
				// AB campaign has already been added to the response
				continue
			}
		}

		var vid string
		isNew := false
		isNewAnonymous := false
		existingAssignment, ok := allCacheAssignments.Standard.GetAssignment(vg.ID)
		existingAssignmentAnonymous, okAnonymous := allCacheAssignments.Anonymous.GetAssignment(vg.ID)

		var existingVariation *Variation
		var existingAnonymousVariation *Variation
		if ok || okAnonymous {
			for _, v := range vg.Variations {
				if ok && v.ID == existingAssignment.VariationID {
					existingVariation = v
				}
				if okAnonymous && v.ID == existingAssignmentAnonymous.VariationID {
					existingAnonymousVariation = v
				}
			}

			if existingVariation == nil && existingAnonymousVariation == nil {
				// Variation has been deleted
				continue
			}
		}
		var chosenVariation *Variation

		// manage the bucket allocation of the visitor
		// if the visitor already have been allocated to a variation, we want to bypass the bucket allocation
		enableBucketAllocation := options.EnableBucketAllocation == nil || *options.EnableBucketAllocation

		// If already has variation && assigned variation ID  exist, visitor should not be re-assigned
		if ok && existingVariation != nil {
			vid = existingAssignment.VariationID
			chosenVariation = existingVariation
			enableBucketAllocation = false
		} else if enableReconciliation && okAnonymous && existingAnonymousVariation != nil {
			// If reconciliation is on, find anonymous variation as set vid to that variation ID
			vid = existingAssignmentAnonymous.VariationID
			chosenVariation = existingAnonymousVariation
			enableBucketAllocation = false
			isNew = true
		} else {
			// Else compute new allocation
			chosenVariation, err = GetRandomAllocation(visitorID, vg, options.IsCumulativeAlloc)
			if err != nil {
				log.Println(fmt.Sprintf("Error on new allocation : %v", err))
				if options.CampaignID != "" {
					return decisionResponse, err
				}
				continue
			}
			vid = chosenVariation.ID
			isNew = true
			isNewAnonymous = true
		}

		if enableBucketAllocation {
			isInBucket, err := IsVisitorInBucket(visitorID, vg.Campaign)
			if err != nil {
				log.Println(fmt.Sprintf("Error on bucket allocation for campaign %v : %v", vg.Campaign.ID, err))
			}

			if !isInBucket {
				continue
			}
		}

		// 3.1 If allocation is newly computed and not only 1 variation,
		// or if campaign activation not saved and should be
		// tag this vg alloc to be saved
		alreadyActivated := ok && existingAssignment.Activated
		if triggerHit && !alreadyActivated || isNew {
			newVGAssignments[vg.ID] = &VisitorVGCacheItem{
				VariationID: vid,
				Activated:   triggerHit,
			}
		}

		// 3.1 If anonymous allocation is newly computed and not only 1 variation,
		// or if campaign activation not saved and should be
		// tag this vg alloc to be saved
		alreadyActivatedAnonymous := okAnonymous && existingAssignmentAnonymous.Activated
		if triggerHit && !alreadyActivatedAnonymous || isNewAnonymous {
			newVGAssignmentsAnonymous[vg.ID] = &VisitorVGCacheItem{
				VariationID: vid,
				Activated:   triggerHit,
			}
		}

		if triggerHit {
			anonymousIDActivate := visitorID
			if enableReconciliation {
				anonymousIDActivate = anonymousID
			}
			cActivations = append(cActivations, &VisitorActivation{
				EnvironmentID:    envID,
				VisitorID:        visitorID,
				AnonymousID:      anonymousIDActivate,
				VariationGroupID: vg.ID,
				VariationID:      vid,
			})
		}

		// 3.3 Build single campaign response from variation
		campaignResponse := buildCampaignResponse(vg, chosenVariation, options.ExposeAllKeys)

		// 3.4 Add campaign response to global response
		decisionResponse.Campaigns = append(decisionResponse.Campaigns, campaignResponse)

		// 3.5 Remember if AB campaign
		if vg.Campaign.Type == "ab" {
			hasABCampaign = true
		}
	}

	now := time.Now()
	_, _, _, _ = <-utils.RunTaskAsync(func() {
		if enableCache && len(newVGAssignments) > 0 && handlers.SaveCache != nil {
			// 4 Persist visitor ID new vg assignments to cache db
			err := handlers.SaveCache(envID, visitorID, &VisitorAssignments{
				Timestamp:   now.Unix(),
				Assignments: newVGAssignments,
			})
			if err != nil {
				log.Printf("Error occured on cache saving: %v", err)
			}
		}
	}), utils.RunTaskAsync(func() {
		if enableCache && len(newVGAssignmentsAnonymous) > 0 && handlers.SaveCache != nil {
			// 4 Persist anonymous ID new vg assignments to cache db
			err := handlers.SaveCache(envID, anonymousID, &VisitorAssignments{
				Timestamp:   now.Unix(),
				Assignments: newVGAssignmentsAnonymous,
			})
			if err != nil {
				log.Printf("Error occured on cache saving: %v", err)
			}
		}
	}), utils.RunTaskAsync(func() {
		if enableCache && decisionGroup != "" && handlers.SaveCache != nil {
			// 5 Persist decision group new vg assignments to cache db
			err := handlers.SaveCache(envID, decisionGroup, &VisitorAssignments{
				Timestamp:   now.Unix(),
				Assignments: newVGAssignments,
			})
			if err != nil {
				log.Printf("Error occured on cache saving: %v", err)
			}
		}
	}), <-utils.RunTaskAsync(func() {
		if len(cActivations) > 0 && handlers.ActivateCampaigns != nil {
			tracker.TimeTrack("Start activating campaigns hit")
			err := handlers.ActivateCampaigns(cActivations)
			if err != nil {
				log.Printf("Error occured on campaign activation: %v", err)
			}
			tracker.TimeTrack("End activating campaigns hit")
		}
	})

	return decisionResponse, nil
}
