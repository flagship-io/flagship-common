package decision

import (
	"errors"
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testVariationGroupAlloc(vg VariationGroup, t *testing.T, isCumulativeAlloc bool) {
	counts := []int{}
	for i := 1; i < 100000; i++ {
		vAlloc, err := getRandomAllocation(strconv.Itoa(rand.Int()), &vg, isCumulativeAlloc)

		if err != nil {
			log.Println(err.Error())
			return
		}

		for i, v := range vg.Variations {
			if v.ID == vAlloc.ID {
				for len(counts) <= i {
					counts = append(counts, 0)
				}
				counts[i]++
			}
		}
	}

	countTotal := 0
	for i, v := range counts {
		t.Logf("Count v%d : %d", i+1, v)
		countTotal += v
	}

	nbVarWithTraffic := 0
	hasVariationFullTraffic := false
	for i, v := range vg.Variations {
		if i == 0 || i > 0 && v.Allocation != 0 && !hasVariationFullTraffic {
			nbVarWithTraffic++
		}
		if v.Allocation == 100 {
			hasVariationFullTraffic = true
		}
	}

	countWithTraffic := 0
	for _, v := range counts {
		if v > 0 {
			countWithTraffic++
		}
	}
	if countWithTraffic != nbVarWithTraffic {
		t.Errorf("Problem with stats: assigned vars : %d, nb var total : %d", countWithTraffic, nbVarWithTraffic)
	}

	previousRatio := float32(0)
	for i, v := range counts {
		correctRatio := vg.Variations[i].Allocation / 100
		if isCumulativeAlloc {
			if i >= 1 {
				previousRatio += vg.Variations[i-1].Allocation
			}
			correctRatio = (vg.Variations[i].Allocation - previousRatio) / 100
		}
		ratio := float32(v) / float32(countTotal)
		if correctRatio-ratio > 0.05 {
			t.Errorf("Problem with stats: ratio %f, correctRatio : %f", ratio, correctRatio)
		}
	}
}

func TestVariationAllocation(t *testing.T) {
	variationArray := []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 50})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 50})

	variationsGroupInfo := VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t, false)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 34})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t, false)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 25})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 25})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 40})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t, false)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 90})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 0})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 0})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t, false)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 90})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 100})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 100})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 100})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t, true)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 90})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 0})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 0})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t, false)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 50})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 0})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	allocErrors := []error{}
	nbTrials := 100000
	for i := 1; i < nbTrials; i++ {
		_, err := getRandomAllocation(strconv.Itoa(rand.Int()), &variationsGroupInfo, false)

		if err != nil {
			allocErrors = append(allocErrors, err)
			continue
		}
	}
	errRatio := float64(len(allocErrors)) / float64(nbTrials)
	log.Printf("errRatio: %f", errRatio)
	isRatioCorrect := 0.5-errRatio < 0.05
	assert.EqualValues(t, errors.New("Visitor untracked"), allocErrors[0])
	assert.True(t, isRatioCorrect)
}
func TestisVisitorInBucket(t *testing.T) {
	// visitor in bucket 100% should be allowed
	is, err := isVisitorInBucket("123", &Campaign{
		ID:           "123",
		BucketRanges: [][]float64{{0., 100.}},
	})
	assert.Nil(t, err)
	assert.True(t, is)

	// visitor in bucket in its appropriate bucket should be allowed
	is, err = isVisitorInBucket("123", &Campaign{
		ID:           "123",
		BucketRanges: [][]float64{{71., 71.5}},
	})
	assert.Nil(t, err)
	assert.True(t, is)

	// visitor in bucket in one of its appropriate bucket should be allowed
	is, err = isVisitorInBucket("123", &Campaign{
		ID:           "123",
		BucketRanges: [][]float64{{71., 71.5}, {40., 50.}},
	})
	assert.Nil(t, err)
	assert.True(t, is)

	// same condition but on an other campaign, the visitor have the same hash and should be
	is, err = isVisitorInBucket("123", &Campaign{
		ID:           "456",
		BucketRanges: [][]float64{{71., 71.5}},
	})
	assert.Nil(t, err)
	assert.True(t, is)

	// no buckets, visitor not allocated
	is, err = isVisitorInBucket("123", &Campaign{
		ID:           "999",
		BucketRanges: [][]float64{},
	})
	assert.Nil(t, err)
	assert.False(t, is)
}
