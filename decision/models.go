package decision

import (
	"github.com/flagship-io/flagship-proto/targeting"
)

// VariationInfo stores the variation information for decision making
type Variation struct {
	ID      string
	Traffic float32
}

// VariationsGroupInfo stores the variation group information for decision making
type VariationsGroup struct {
	ID         string
	Targetings *targeting.Targeting
	Variations []*Variation
}
