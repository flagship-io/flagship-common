package decision

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/flagship-io/flagship-common/utils"
	"github.com/flagship-io/flagship-proto/decision_response"
	"github.com/flagship-io/flagship-proto/targeting"
	protoStruct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/types/known/structpb"
)

// VariationInfo stores the variation information for decision making
type Variation struct {
	ID            string
	Allocation    float32
	Modifications *decision_response.Modifications
	Reference     bool
}

// VariationsGroupInfo stores the variation group information for decision making
type VariationsGroup struct {
	ID           string
	CampaignID   string
	CampaignType string
	CreatedAt    time.Time
	Targetings   *targeting.Targeting
	Variations   []*Variation
}

// VisitorVGCacheItem represents a visitor variation group cache item for a variation group
type VisitorVGCacheItem struct {
	VariationID string
	Activated   bool
}

// VisitorAssignments represents a visitor assignment for a variation group
type VisitorAssignments struct {
	Timestamp   int64
	Assignments map[string]*VisitorVGCacheItem
}

type VisitorInfo struct {
	ID            string
	AnonymousID   string
	DecisionGroup string
	Context       map[string]*structpb.Value
}

type EnvironmentInfo struct {
	ID                string
	Campaigns         map[string]*CampaignInfo
	IsPanic           bool
	SingleAssignment  bool
	UseReconciliation bool
	CacheEnabled      bool
}

type DecisionOptions struct {
	TriggerHit    bool
	CampaignID    string
	Tracker       *utils.Tracker
	ExposeAllKeys bool
}

type VisitorActivation struct {
	EnvironmentID    string
	VisitorID        string
	AnonymousID      string
	VariationGroupID string
	VariationID      string
}

type DecisionHandlers struct {
	GetCache          func(environmentID string, id string) (*VisitorAssignments, error)
	SaveCache         func(environmentID string, id string, assignment *VisitorAssignments) error
	ActivateCampaigns func(activations []*VisitorActivation) error
}

type DecisionResponse struct {
	Campaigns []*decision_response.Campaign
}

// GetAssignments returns all the assigments
func (va *VisitorAssignments) GetAssignments() map[string]*VisitorVGCacheItem {
	if va == nil {
		return nil
	}

	return va.Assignments
}

// GetAssignment returns the existing assignment of the visitor for the vg ID
func (va *VisitorAssignments) GetAssignment(vgID string) (*VisitorVGCacheItem, bool) {
	if va == nil {
		return nil, false
	}

	existing, ok := va.Assignments[vgID]
	return existing, ok
}

// CampaignInfo stores the campaign information for decision making
type CampaignInfo struct {
	ID               string
	CustomID         *string
	VariationsGroups map[string]*VariationsGroup
	Type             string
	CreatedAt        time.Time
}

type byCreatedAtCampaigns []*CampaignInfo

func (s byCreatedAtCampaigns) Len() int {
	return len(s)
}
func (s byCreatedAtCampaigns) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byCreatedAtCampaigns) Less(i, j int) bool {
	return s[i].CreatedAt.After(s[j].CreatedAt)
}

type byCreatedAtVG []*VariationsGroup

func (s byCreatedAtVG) Len() int {
	return len(s)
}
func (s byCreatedAtVG) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byCreatedAtVG) Less(i, j int) bool {
	return s[i].CreatedAt.Before(s[j].CreatedAt)
}

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
		vg.CampaignType = campaign.Type
		existingCampaignVG[campaign.ID] = true
		campaignVG = append(campaignVG, vg)
	}
	return campaignVG
}
