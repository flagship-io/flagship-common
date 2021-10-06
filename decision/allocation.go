package decision

import (
	"errors"

	"github.com/spaolacci/murmur3"
)

var hash = murmur3.New32()

// GetRandomAllocation returns a random allocation for a variationGroup
func GetRandomAllocation(visitorID string, variationGroup *VariationsGroup) (*Variation, error) {
	hash.Reset()
	_, err := hash.Write([]byte(variationGroup.ID + visitorID))

	if err != nil {
		return nil, err
	}

	hashed := hash.Sum32()
	z := hashed % 100

	sumAlloc := 0
	for _, v := range variationGroup.Variations {
		sumAlloc += int(v.Allocation)
		if int(z) < sumAlloc {
			return v, nil
		}
	}

	// If no variation alloc, returns empty
	return nil, errors.New("Visitor untracked")
}
