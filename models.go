package decision

import (
	"time"

	"github.com/flagship-io/flagship-proto/decision_response"
	"github.com/flagship-io/flagship-proto/targeting"
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
type VariationGroup struct {
	ID         string
	Campaign   *Campaign
	CreatedAt  time.Time
	Targetings *targeting.Targeting
	Variations []*Variation
}

// VisitorCache represents a visitor variation group cache item for a variation group
type VisitorCache struct {
	VariationID string
	Activated   bool
}

// VisitorAssignments represents a visitor assignment for a variation group
type VisitorAssignments struct {
	Timestamp   int64
	Assignments map[string]*VisitorCache
}

type Visitor struct {
	ID            string
	AnonymousID   string
	DecisionGroup string
	Context       map[string]*structpb.Value
}

type Environment struct {
	ID                string
	Campaigns         map[string]*Campaign
	IsPanic           bool
	SingleAssignment  bool
	UseReconciliation bool
	CacheEnabled      bool
}

type DecisionOptions struct {
	TriggerHit             bool
	CampaignID             string
	Tracker                *Tracker
	ExposeAllKeys          bool
	IsCumulativeAlloc      bool
	EnableBucketAllocation *bool
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

// Campaign stores the campaign information for decision making
type Campaign struct {
	ID              string
	Slug            *string
	VariationGroups map[string]*VariationGroup
	Type            string
	CreatedAt       time.Time
	BucketRanges    [][]float64
}

type byCreatedAtCampaigns []*Campaign
type byCreatedAtVG []*VariationGroup

// GetAssignments returns all the assigments
func (va *VisitorAssignments) getAssignments() map[string]*VisitorCache {
	if va == nil {
		return nil
	}

	return va.Assignments
}

// GetAssignment returns the existing assignment of the visitor for the vg ID
func (va *VisitorAssignments) getAssignment(vgID string) (*VisitorCache, bool) {
	if va == nil {
		return nil, false
	}

	existing, ok := va.Assignments[vgID]
	return existing, ok
}

func (s byCreatedAtCampaigns) Len() int {
	return len(s)
}
func (s byCreatedAtCampaigns) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byCreatedAtCampaigns) Less(i, j int) bool {
	return s[i].CreatedAt.After(s[j].CreatedAt)
}

func (s byCreatedAtVG) Len() int {
	return len(s)
}
func (s byCreatedAtVG) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byCreatedAtVG) Less(i, j int) bool {
	return s[i].CreatedAt.Before(s[j].CreatedAt)
}
