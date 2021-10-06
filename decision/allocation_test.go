package decision

import (
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func testVariationGroupAlloc(vg VariationsGroup, t *testing.T) {
	counts := []int{}
	for i := 1; i < 100000; i++ {
		vAlloc, err := GetRandomAllocation(strconv.Itoa(rand.Int()), &vg)

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
	for i, v := range vg.Variations {
		if i == 0 || i > 0 && v.Allocation != 0 {
			nbVarWithTraffic++
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
		if i >= 1 {
			previousRatio += vg.Variations[i-1].Allocation
		}
		correctRatio := (vg.Variations[i].Allocation - previousRatio) / 100
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

	variationsGroupInfo := VariationsGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 34})

	variationsGroupInfo = VariationsGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 25})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 35})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 40})

	variationsGroupInfo = VariationsGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 90})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 0})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 0})

	variationsGroupInfo = VariationsGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 90})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 0})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 0})

	variationsGroupInfo = VariationsGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)
}
