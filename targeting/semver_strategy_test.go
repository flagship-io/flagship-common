package targeting

import (
	"testing"

	targetingProto "github.com/flagship-io/flagship-proto/targeting"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSemverTargetingAll(t *testing.T) {
	testSemverTargeting(targetingProto.Targeting_LOWER_THAN, "v2.0.1", "v2.0.2", t, true)
	testSemverTargeting(targetingProto.Targeting_LOWER_THAN, "v2.0.2", "v2.0.1", t, false)
	testSemverTargeting(targetingProto.Targeting_LOWER_THAN, "v2.0.2", "v2.0.11", t, true)
	testSemverTargeting(targetingProto.Targeting_GREATER_THAN, "v2.0.2", "v2.0.1", t, true)
	testSemverTargeting(targetingProto.Targeting_GREATER_THAN, "v2.0.1", "v2.0.2", t, false)
	testSemverTargeting(targetingProto.Targeting_GREATER_THAN, "v2.0.11", "v2.0.2", t, true)
	testSemverTargeting(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, "v2.0.1", "v2.0.1", t, true)
	testSemverTargeting(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, "v2.0.1", "v2.0.2", t, true)
	testSemverTargeting(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, "v2.0.1", "v2.0.1", t, true)
	testSemverTargeting(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, "v2.0.2", "v2.0.1", t, true)
	testSemverTargeting(targetingProto.Targeting_EQUALS, "v1.0.0", "v1.0.0", t, true)
	testSemverTargeting(targetingProto.Targeting_NOT_EQUALS, "v1.0.0", "v2.0.0", t, true)
	testSemverTargeting(targetingProto.Targeting_GREATER_THAN, "2.0.11", "2.0.2", t, true) // add default "v" to version
	testSemverTargeting(targetingProto.Targeting_EQUALS, "toto", "toto", t, true)          // default strategy applied
}

func testSemverTargeting(operator targetingProto.Targeting_TargetingOperator, value string, targetingValue string, t *testing.T, shouldMatch bool) {
	targeting := SemverStrategy{
		Targeting: &targetingProto.Targeting_InnerTargeting{
			Operator: operator,
			Value:    structpb.NewStringValue(targetingValue),
		},
		ContextValue: structpb.NewStringValue(value),
	}
	assert.False(t, targeting.ShouldIgnoreTargeting())
	match, err := targeting.Match()
	assert.Nil(t, err)
	assert.Equal(t, match, shouldMatch)
}
