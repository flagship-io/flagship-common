package decision

import (
	"errors"

	"github.com/spaolacci/murmur3"
)

var VisitorNotTrackedError = errors.New("Visitor untracked")

func genHashFloat(visitorID string, vgID string) (float32, error) {
	hash := murmur3.New32()
	_, err := hash.Write([]byte(vgID + visitorID))

	if err != nil {
		return 0, err
	}

	hashed := hash.Sum32()
	return float32(hashed % 100), nil
}

// getRandomAllocation returns a random allocation for a variationGroup
func getRandomAllocation(visitorID string, decisionGroup string, variationGroup *VariationGroup, isCumulativeAlloc bool) (*Variation, error) {
	// performance shortcut to prevent hash generation
	if len(variationGroup.Variations) == 1 && variationGroup.Variations[0].Allocation == 100 {
		return variationGroup.Variations[0], nil
	}

	// Use decision group by default for decision hash, otherwise use visitor ID
	decisionID := visitorID
	if decisionGroup != "" {
		decisionID = decisionGroup
	}

	z, err := genHashFloat(decisionID, variationGroup.ID)
	if err != nil {
		return nil, err
	}

	sumAlloc := float32(0)
	for _, v := range variationGroup.Variations {
		sumAlloc += v.Allocation
		if isCumulativeAlloc {
			sumAlloc = v.Allocation
		}
		if z < sumAlloc {
			return v, nil
		}
	}

	// If no variation alloc, returns empty
	return nil, VisitorNotTrackedError
}

func isVisitorInBucket(visitorID string, campaign *Campaign) (bool, error) {
	if campaign == nil {
		return false, errors.New("campaign is null")
	}

	z, err := genHashFloat(visitorID, "")
	if err != nil {
		return false, err
	}

	for _, br := range campaign.BucketRanges {
		if z >= float32(br[0]) && z < float32(br[1]) {
			return true, nil
		}
	}
	return false, nil
}
