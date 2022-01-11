package decision

import (
	"errors"

	"github.com/spaolacci/murmur3"
)

var hash = murmur3.New32()

func genHashFloat(visitorID string, vgID string) (float32, error) {
	hash.Reset()
	_, err := hash.Write([]byte(vgID + visitorID))

	if err != nil {
		return 0, err
	}

	hashed := hash.Sum32()
	return float32(hashed % 100), nil
}

// getRandomAllocation returns a random allocation for a variationGroup
func getRandomAllocation(visitorID string, variationGroup *VariationGroup, isCumulativeAlloc bool) (*Variation, error) {
	z, err := genHashFloat(visitorID, variationGroup.ID)
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
	return nil, errors.New("Visitor untracked")
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
