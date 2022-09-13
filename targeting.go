package decision

import (
	"errors"
	"reflect"
	"strings"

	"github.com/flagship-io/flagship-common/targeting"
	protoTargeting "github.com/flagship-io/flagship-proto/targeting"
	"google.golang.org/protobuf/types/known/structpb"
)

// targetingMatch returns true if a visitor ID and context match the variationGroup targeting
func targetingMatch(targetings *protoTargeting.Targeting, visitorID string, context *targeting.Context) (bool, error) {
	globalMatch := false
	for _, targetingGroup := range targetings.GetTargetingGroups() {
		matchGroup := len(targetingGroup.GetTargetings()) > 0
		for _, t := range targetingGroup.GetTargetings() {
			v, ok := context.GetValueByProvider(t.GetKey().GetValue(), t.GetProvider().GetValue())
			switch t.GetKey().GetValue() {
			case "fs_all_users":
				// All users targeting will
				continue
			case "fs_users":
				v = structpb.NewStringValue(visitorID)
				ok = true
			}

			if ok || isEmptyContextOperator(t.GetOperator()) {
				matchTargeting, err := targetingMatchOperator(t.GetOperator(), t.GetValue(), v)
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

func isANDListOperator(operator protoTargeting.Targeting_TargetingOperator) bool {
	return operator == protoTargeting.Targeting_NOT_CONTAINS || operator == protoTargeting.Targeting_NOT_EQUALS
}

func isORListOperator(operator protoTargeting.Targeting_TargetingOperator) bool {
	return operator == protoTargeting.Targeting_CONTAINS || operator == protoTargeting.Targeting_EQUALS
}

func isEmptyContextOperator(operator protoTargeting.Targeting_TargetingOperator) bool {
	return operator == protoTargeting.Targeting_EXISTS || operator == protoTargeting.Targeting_NOT_EXISTS
}

func targetingMatchOperator(operator protoTargeting.Targeting_TargetingOperator, targetingValue *structpb.Value, contextValue *structpb.Value) (bool, error) {
	match := false
	var err error

	listValues := contextValue.GetListValue()
	if listValues != nil && len(listValues.GetValues()) > 0 && reflect.TypeOf(listValues.GetValues()[0].GetKind()) != reflect.TypeOf(targetingValue.GetKind()) {
		return false, errors.New("Targeting and Context list value kinds mismatch")
	}

	if listValues != nil {
		match = isANDListOperator(operator)
		for _, v := range listValues.GetValues() {
			subValueMatch, err := targetingMatchOperator(operator, targetingValue, v)
			if err != nil {
				return false, nil
			}
			if isANDListOperator(operator) {
				match = match && err == nil && subValueMatch
			}
			if isORListOperator(operator) {
				match = match || (err == nil && subValueMatch)
			}
		}
		return match, nil
	}

	if isEmptyContextOperator(operator) {
		targetingValueCasted := targetingValue.GetBoolValue()
		return targetingMatchOperatorEmptyContext(operator, targetingValueCasted, contextValue)
	}

	// Except for targeting value of type list, check that context and targeting types are equals
	if targetingValue.GetListValue() == nil && reflect.TypeOf(targetingValue.GetKind()) != reflect.TypeOf(contextValue.GetKind()) {
		return false, errors.New("Targeting and Context value kinds mismatch")
	}

	switch targetingValue.Kind.(type) {
	case (*structpb.Value_StringValue):
		targetingValueCasted := targetingValue.GetStringValue()
		contextValueCasted := contextValue.GetStringValue()
		match, err = targetingMatchOperatorString(operator, targetingValueCasted, contextValueCasted)
	case (*structpb.Value_BoolValue):
		targetingValueCasted := targetingValue.GetBoolValue()
		contextValueCasted := contextValue.GetBoolValue()
		match, err = targetingMatchOperatorBool(operator, targetingValueCasted, contextValueCasted)
	case (*structpb.Value_NumberValue):
		targetingValueCasted := targetingValue.GetNumberValue()
		contextValueCasted := contextValue.GetNumberValue()
		match, err = targetingMatchOperatorNumber(operator, targetingValueCasted, contextValueCasted)
	case (*structpb.Value_ListValue):
		targetingList := targetingValue.GetListValue()
		match = isANDListOperator(operator)
		for _, v := range targetingList.GetValues() {
			subValueMatch, err := targetingMatchOperator(operator, v, contextValue)
			if isANDListOperator(operator) {
				match = match && err == nil && subValueMatch
			}
			if isORListOperator(operator) {
				match = match || (err == nil && subValueMatch)
			}
		}
	}
	return match, err
}

func targetingMatchOperatorString(operator protoTargeting.Targeting_TargetingOperator, targetingValue string, contextValue string) (bool, error) {
	switch operator {
	case protoTargeting.Targeting_LOWER_THAN:
		return strings.ToLower(contextValue) < strings.ToLower(targetingValue), nil
	case protoTargeting.Targeting_GREATER_THAN:
		return strings.ToLower(contextValue) > strings.ToLower(targetingValue), nil
	case protoTargeting.Targeting_LOWER_THAN_OR_EQUALS:
		return strings.ToLower(contextValue) <= strings.ToLower(targetingValue), nil
	case protoTargeting.Targeting_GREATER_THAN_OR_EQUALS:
		return strings.ToLower(contextValue) >= strings.ToLower(targetingValue), nil
	case protoTargeting.Targeting_EQUALS:
		return strings.EqualFold(contextValue, targetingValue), nil
	case protoTargeting.Targeting_NOT_EQUALS:
		return !strings.EqualFold(contextValue, targetingValue), nil
	case protoTargeting.Targeting_STARTS_WITH:
		return strings.HasPrefix(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case protoTargeting.Targeting_ENDS_WITH:
		return strings.HasSuffix(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case protoTargeting.Targeting_CONTAINS:
		return strings.Contains(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case protoTargeting.Targeting_NOT_CONTAINS:
		return !strings.Contains(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	// case "regex":
	// 	match, err := regexp.MatchString(targetingValue, contextValue)
	// 	return match, err
	default:
		return false, errors.New("Operator not handled")
	}
}

func targetingMatchOperatorNumber(operator protoTargeting.Targeting_TargetingOperator, targetingValue float64, contextValue float64) (bool, error) {
	switch operator {
	case protoTargeting.Targeting_LOWER_THAN:
		return contextValue < targetingValue, nil
	case protoTargeting.Targeting_GREATER_THAN:
		return contextValue > targetingValue, nil
	case protoTargeting.Targeting_LOWER_THAN_OR_EQUALS:
		return contextValue <= targetingValue, nil
	case protoTargeting.Targeting_GREATER_THAN_OR_EQUALS:
		return contextValue >= targetingValue, nil
	case protoTargeting.Targeting_EQUALS:
		return contextValue == targetingValue, nil
	case protoTargeting.Targeting_NOT_EQUALS:
		return contextValue != targetingValue, nil
	default:
		return false, errors.New("operator not handled")
	}
}

func targetingMatchOperatorBool(operator protoTargeting.Targeting_TargetingOperator, targetingValue bool, contextValue bool) (bool, error) {
	switch operator {
	case protoTargeting.Targeting_EQUALS:
		return contextValue == targetingValue, nil
	case protoTargeting.Targeting_NOT_EQUALS:
		return contextValue != targetingValue, nil
	default:
		return false, errors.New("operator not handled")
	}
}

func targetingMatchOperatorEmptyContext(operator protoTargeting.Targeting_TargetingOperator, targetingValue bool, contextValue *structpb.Value) (bool, error) {
	switch operator {
	case protoTargeting.Targeting_EXISTS:
		return contextValue != nil && targetingValue || contextValue == nil && !targetingValue, nil
	case protoTargeting.Targeting_NOT_EXISTS:
		return contextValue == nil && targetingValue || contextValue != nil && !targetingValue, nil
	default:
		return false, errors.New("operator not handled")
	}
}
