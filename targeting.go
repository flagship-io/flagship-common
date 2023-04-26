package decision

import (
	"github.com/flagship-io/flagship-common/targeting"
	protoTargeting "github.com/flagship-io/flagship-proto/targeting"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	AllUsersTargetingKey = "fs_all_users"
	UserTargetingKey     = "fs_users"
	SemverTargetingKey   = "semverAppVersion"
)

// targetingMatch returns true if a visitor ID and context match the variationGroup targeting
func targetingMatch(targetings *protoTargeting.Targeting, visitorID string, context *targeting.Context) (bool, error) {
	globalMatch := false
	for _, targetingGroup := range targetings.GetTargetingGroups() {
		matchGroup := len(targetingGroup.GetTargetings()) > 0
		for _, t := range targetingGroup.GetTargetings() {
			v, _ := context.GetValueByProvider(t.GetKey().GetValue(), t.GetProvider().GetValue())

			var targetingStrategy targeting.Strategy
			switch t.GetKey().GetValue() {
			case AllUsersTargetingKey:
				continue
			case UserTargetingKey:
				targetingStrategy = targeting.DefaultStrategy{
					Targeting:    t,
					ContextValue: structpb.NewStringValue(visitorID),
				}
			case SemverTargetingKey:
				targetingStrategy = targeting.SemverStrategy{
					Targeting:    t,
					ContextValue: structpb.NewStringValue(visitorID),
				}
			default:
				targetingStrategy = targeting.DefaultStrategy{
					Targeting:    t,
					ContextValue: v,
				}
			}

			if !targetingStrategy.ShouldIgnoreTargeting() {
				matchTargeting, err := targetingStrategy.Match()
				if err != nil {
					return false, err
				}

				matchGroup = matchGroup && matchTargeting
			} else {
				matchGroup = false
			}
		}
		globalMatch = globalMatch || matchGroup
	}

	return globalMatch, nil
}
