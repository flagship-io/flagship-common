package decision

import (
	"errors"
	"reflect"
	"strings"

	"github.com/flagship-io/flagship-proto/targeting"

	protoStruct "github.com/golang/protobuf/ptypes/struct"
)

// TargetingMatch returns true if a visitor ID and context match the variationGroup targeting
func TargetingMatch(variationGroup *VariationsGroup, visitorID string, context map[string]*protoStruct.Value) (bool, error) {
	globalMatch := false
	for _, targetingGroup := range variationGroup.Targetings.GetTargetingGroups() {
		matchGroup := len(targetingGroup.GetTargetings()) > 0
		for _, targeting := range targetingGroup.GetTargetings() {
			v, ok := context[targeting.GetKey().GetValue()]
			switch targeting.GetKey().GetValue() {
			case "fs_all_users":
				return true, nil
			case "fs_users":
				v = &protoStruct.Value{
					Kind: &protoStruct.Value_StringValue{
						StringValue: visitorID,
					},
				}
				ok = true
			}

			if ok {
				matchTargeting, err := targetingMatchOperator(targeting.GetOperator(), targeting.GetValue(), v)
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

func isANDListOperator(operator targeting.Targeting_TargetingOperator) bool {
	return operator == targeting.Targeting_NOT_CONTAINS || operator == targeting.Targeting_NOT_EQUALS
}

func isORListOperator(operator targeting.Targeting_TargetingOperator) bool {
	return operator == targeting.Targeting_CONTAINS || operator == targeting.Targeting_EQUALS
}

func targetingMatchOperator(operator targeting.Targeting_TargetingOperator, targetingValue *protoStruct.Value, contextValue *protoStruct.Value) (bool, error) {
	match := false
	var err error

	listValues := contextValue.GetListValue()
	if listValues != nil && len(listValues.GetValues()) > 0 && listValues.GetValues()[0].GetKind() != targetingValue.GetKind() {
		return false, errors.New("Targeting and Context list value kinds mismatch")
	}

	if listValues != nil {
		for _, v := range listValues.GetValues() {
			subValueMatch, err := targetingMatchOperator(operator, targetingValue, v)
			if err != nil {
				return false, nil
			}
			if operator == targeting.Targeting_EQUALS {
				match = match || (err == nil && subValueMatch)
			}
			if operator == targeting.Targeting_NOT_EQUALS {
				match = match && (err == nil && subValueMatch)
			}
		}
		return match, nil
	}

	// Except for targeting value of type list, check that context and targeting types are equals
	if targetingValue.GetListValue() == nil && reflect.TypeOf(targetingValue.GetKind()) != reflect.TypeOf(contextValue.GetKind()) {
		return false, errors.New("Targeting and Context value kinds mismatch")
	}

	switch targetingValue.Kind.(type) {
	case (*protoStruct.Value_StringValue):
		targetingValueCasted := targetingValue.GetStringValue()
		contextValueCasted := contextValue.GetStringValue()
		match, err = targetingMatchOperatorString(operator, targetingValueCasted, contextValueCasted)
	case (*protoStruct.Value_BoolValue):
		targetingValueCasted := targetingValue.GetBoolValue()
		contextValueCasted := contextValue.GetBoolValue()
		match, err = targetingMatchOperatorBool(operator, targetingValueCasted, contextValueCasted)
	case (*protoStruct.Value_NumberValue):
		targetingValueCasted := targetingValue.GetNumberValue()
		contextValueCasted := contextValue.GetNumberValue()
		match, err = targetingMatchOperatorNumber(operator, targetingValueCasted, contextValueCasted)
	case (*protoStruct.Value_ListValue):
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

func targetingMatchOperatorString(operator targeting.Targeting_TargetingOperator, targetingValue string, contextValue string) (bool, error) {
	switch operator {
	case targeting.Targeting_LOWER_THAN:
		return strings.ToLower(contextValue) < strings.ToLower(targetingValue), nil
	case targeting.Targeting_GREATER_THAN:
		return strings.ToLower(contextValue) > strings.ToLower(targetingValue), nil
	case targeting.Targeting_LOWER_THAN_OR_EQUALS:
		return strings.ToLower(contextValue) <= strings.ToLower(targetingValue), nil
	case targeting.Targeting_GREATER_THAN_OR_EQUALS:
		return strings.ToLower(contextValue) >= strings.ToLower(targetingValue), nil
	case targeting.Targeting_EQUALS:
		return strings.EqualFold(contextValue, targetingValue), nil
	case targeting.Targeting_NOT_EQUALS:
		return !strings.EqualFold(contextValue, targetingValue), nil
	case targeting.Targeting_STARTS_WITH:
		return strings.HasPrefix(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case targeting.Targeting_ENDS_WITH:
		return strings.HasSuffix(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case targeting.Targeting_CONTAINS:
		return strings.Contains(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	case targeting.Targeting_NOT_CONTAINS:
		return !strings.Contains(strings.ToLower(contextValue), strings.ToLower(targetingValue)), nil
	// case "regex":
	// 	match, err := regexp.MatchString(targetingValue, contextValue)
	// 	return match, err
	default:
		return false, errors.New("Operator not handled")
	}
}

func targetingMatchOperatorNumber(operator targeting.Targeting_TargetingOperator, targetingValue float64, contextValue float64) (bool, error) {
	switch operator {
	case targeting.Targeting_LOWER_THAN:
		return contextValue < targetingValue, nil
	case targeting.Targeting_GREATER_THAN:
		return contextValue > targetingValue, nil
	case targeting.Targeting_LOWER_THAN_OR_EQUALS:
		return contextValue <= targetingValue, nil
	case targeting.Targeting_GREATER_THAN_OR_EQUALS:
		return contextValue >= targetingValue, nil
	case targeting.Targeting_EQUALS:
		return contextValue == targetingValue, nil
	case targeting.Targeting_NOT_EQUALS:
		return contextValue != targetingValue, nil
	default:
		return false, errors.New("Operator not handled")
	}
}

func targetingMatchOperatorBool(operator targeting.Targeting_TargetingOperator, targetingValue bool, contextValue bool) (bool, error) {
	switch operator {
	case targeting.Targeting_EQUALS:
		return contextValue == targetingValue, nil
	case targeting.Targeting_NOT_EQUALS:
		return contextValue != targetingValue, nil
	default:
		return false, errors.New("Operator not handled")
	}
}
