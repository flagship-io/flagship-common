package targeting

import (
	"strings"

	protoTargeting "github.com/flagship-io/flagship-proto/targeting"
	"golang.org/x/mod/semver"
	"google.golang.org/protobuf/types/known/structpb"
)

type SemverStrategy struct {
	Targeting    *protoTargeting.Targeting_InnerTargeting
	ContextValue *structpb.Value
}

func (ss SemverStrategy) ShouldIgnoreTargeting() bool {
	return ss.ContextValue == nil
}

func (ss SemverStrategy) Match() (bool, error) {
	contextValue := prefixDefaultVString(ss.ContextValue.GetStringValue())
	targetingValue := prefixDefaultVString(ss.Targeting.Value.GetStringValue())
	if semver.IsValid(contextValue) && semver.IsValid(targetingValue) {
		switch ss.Targeting.Operator {
		case protoTargeting.Targeting_LOWER_THAN:
			return semver.Compare(contextValue, targetingValue) == -1, nil
		case protoTargeting.Targeting_GREATER_THAN:
			return semver.Compare(contextValue, targetingValue) == 1, nil
		case protoTargeting.Targeting_LOWER_THAN_OR_EQUALS:
			return semver.Compare(contextValue, targetingValue) <= 0, nil
		case protoTargeting.Targeting_GREATER_THAN_OR_EQUALS:
			return semver.Compare(contextValue, targetingValue) >= 0, nil
		case protoTargeting.Targeting_EQUALS:
			return semver.Compare(contextValue, targetingValue) == 0, nil
		case protoTargeting.Targeting_NOT_EQUALS:
			return semver.Compare(contextValue, targetingValue) != 0, nil
		}
	}
	// when value is not semver valid or operator is wrong, return the default targeting strategy
	return DefaultStrategy(ss).Match()
}

func prefixDefaultVString(input string) string {
	if !strings.HasPrefix(input, "v") {
		return "v" + input
	}
	return input
}
