package decision

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flagship-io/flagship-proto/targeting"
	protoStruct "github.com/golang/protobuf/ptypes/struct"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func testTargetingNumber(operator targeting.Targeting_TargetingOperator, targetingValue float64, value float64, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorNumber(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %f, v: %f, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingBoolean(operator targeting.Targeting_TargetingOperator, targetingValue bool, value bool, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorBool(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingString(operator targeting.Targeting_TargetingOperator, targetingValue string, value string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorString(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingListString(operator targeting.Targeting_TargetingOperator, targetingValues []string, value string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	stringValues := []*protoStruct.Value{}
	for _, str := range targetingValues {
		stringValue := &protoStruct.Value{
			Kind: &protoStruct.Value_StringValue{
				StringValue: str,
			},
		}
		stringValues = append(stringValues, stringValue)
	}

	match, err := targetingMatchOperator(operator, &protoStruct.Value{
		Kind: &protoStruct.Value_ListValue{
			ListValue: &protoStruct.ListValue{
				Values: stringValues,
			},
		},
	}, &protoStruct.Value{
		Kind: &protoStruct.Value_StringValue{
			StringValue: value,
		},
	})

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting list %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValues, value, match, err)
	}
}

func testTargetingContextListString(operator targeting.Targeting_TargetingOperator, targetingValue string, contextValue []string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	stringValues := []*protoStruct.Value{}
	for _, str := range contextValue {
		stringValue := &protoStruct.Value{
			Kind: &protoStruct.Value_StringValue{
				StringValue: str,
			},
		}
		stringValues = append(stringValues, stringValue)
	}

	match, err := targetingMatchOperator(operator, &protoStruct.Value{
		Kind: &protoStruct.Value_StringValue{
			StringValue: targetingValue,
		},
	}, &protoStruct.Value{
		Kind: &protoStruct.Value_ListValue{
			ListValue: &protoStruct.ListValue{
				Values: stringValues,
			},
		},
	})

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting list %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, contextValue, match, err)
	}
}

// TestNumberTargeting checks all possible number targeting
func TestNumberTargeting(t *testing.T) {
	testTargetingNumber(targeting.Targeting_LOWER_THAN, 11, 10, t, true, false)
	testTargetingNumber(targeting.Targeting_LOWER_THAN, 10, 10, t, false, false)
	testTargetingNumber(targeting.Targeting_LOWER_THAN, 9, 10, t, false, false)

	testTargetingNumber(targeting.Targeting_LOWER_THAN_OR_EQUALS, 11, 10, t, true, false)
	testTargetingNumber(targeting.Targeting_LOWER_THAN_OR_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(targeting.Targeting_LOWER_THAN_OR_EQUALS, 9, 10, t, false, false)

	testTargetingNumber(targeting.Targeting_GREATER_THAN, 11, 10, t, false, false)
	testTargetingNumber(targeting.Targeting_GREATER_THAN, 10, 10, t, false, false)
	testTargetingNumber(targeting.Targeting_GREATER_THAN, 9, 10, t, true, false)

	testTargetingNumber(targeting.Targeting_GREATER_THAN_OR_EQUALS, 11, 10, t, false, false)
	testTargetingNumber(targeting.Targeting_GREATER_THAN_OR_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(targeting.Targeting_GREATER_THAN_OR_EQUALS, 9, 10, t, true, false)

	testTargetingNumber(targeting.Targeting_NOT_EQUALS, 11, 10, t, true, false)
	testTargetingNumber(targeting.Targeting_NOT_EQUALS, 10, 10, t, false, false)
	testTargetingNumber(targeting.Targeting_NOT_EQUALS, 9, 10, t, true, false)

	testTargetingNumber(targeting.Targeting_EQUALS, 11, 10, t, false, false)
	testTargetingNumber(targeting.Targeting_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(targeting.Targeting_EQUALS, 9, 10, t, false, false)

	testTargetingNumber(targeting.Targeting_CONTAINS, 11, 10, t, false, true)
	testTargetingNumber(targeting.Targeting_ENDS_WITH, 10, 10, t, false, true)
	testTargetingNumber(targeting.Targeting_STARTS_WITH, 9, 10, t, false, true)
}

// TestBooleanTargeting checks all possible boolean targeting
func TestBooleanTargeting(t *testing.T) {
	testTargetingBoolean(targeting.Targeting_NOT_EQUALS, true, false, t, true, false)
	testTargetingBoolean(targeting.Targeting_NOT_EQUALS, true, true, t, false, false)
	testTargetingBoolean(targeting.Targeting_NOT_EQUALS, false, true, t, true, false)

	testTargetingBoolean(targeting.Targeting_EQUALS, true, false, t, false, false)
	testTargetingBoolean(targeting.Targeting_EQUALS, true, true, t, true, false)
	testTargetingBoolean(targeting.Targeting_EQUALS, false, true, t, false, false)

	testTargetingBoolean(targeting.Targeting_CONTAINS, true, false, t, false, true)
	testTargetingBoolean(targeting.Targeting_ENDS_WITH, true, false, t, false, true)
	testTargetingBoolean(targeting.Targeting_STARTS_WITH, true, false, t, false, true)
	testTargetingBoolean(targeting.Targeting_GREATER_THAN, true, false, t, false, true)
	testTargetingBoolean(targeting.Targeting_GREATER_THAN_OR_EQUALS, true, false, t, false, true)
	testTargetingBoolean(targeting.Targeting_LOWER_THAN, true, false, t, false, true)
	testTargetingBoolean(targeting.Targeting_LOWER_THAN_OR_EQUALS, true, false, t, false, true)
}

// TestStringTargeting checks all possible string targeting
func TestStringTargeting(t *testing.T) {
	testTargetingString(targeting.Targeting_LOWER_THAN, "abc", "abd", t, false, false)
	testTargetingString(targeting.Targeting_LOWER_THAN, "abc", "abc", t, false, false)
	testTargetingString(targeting.Targeting_LOWER_THAN, "abd", "abc", t, true, false)

	testTargetingString(targeting.Targeting_LOWER_THAN_OR_EQUALS, "abc", "abd", t, false, false)
	testTargetingString(targeting.Targeting_LOWER_THAN_OR_EQUALS, "abc", "abc", t, true, false)
	testTargetingString(targeting.Targeting_LOWER_THAN_OR_EQUALS, "abd", "abc", t, true, false)

	testTargetingString(targeting.Targeting_GREATER_THAN, "abc", "abd", t, true, false)
	testTargetingString(targeting.Targeting_GREATER_THAN, "abc", "abc", t, false, false)
	testTargetingString(targeting.Targeting_GREATER_THAN, "abd", "abc", t, false, false)

	testTargetingString(targeting.Targeting_GREATER_THAN_OR_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(targeting.Targeting_GREATER_THAN_OR_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(targeting.Targeting_GREATER_THAN_OR_EQUALS, "abd", "abc", t, false, false)

	testTargetingString(targeting.Targeting_NOT_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(targeting.Targeting_NOT_EQUALS, "abc", "abc", t, false, false)
	testTargetingString(targeting.Targeting_NOT_EQUALS, "", "", t, false, false)
	testTargetingString(targeting.Targeting_NOT_EQUALS, "", " ", t, true, false)

	testTargetingString(targeting.Targeting_EQUALS, "abc", "abd", t, false, false)
	testTargetingString(targeting.Targeting_EQUALS, "abc", "abc", t, true, false)
	testTargetingString(targeting.Targeting_EQUALS, "ABC", "abc", t, true, false)
	testTargetingString(targeting.Targeting_EQUALS, "", "", t, true, false)
	testTargetingString(targeting.Targeting_EQUALS, "", " ", t, false, false)

	testTargetingString(targeting.Targeting_CONTAINS, "b", "abc", t, true, false)
	testTargetingString(targeting.Targeting_CONTAINS, "B", "abc", t, true, false)
	testTargetingString(targeting.Targeting_CONTAINS, "d", "abc", t, false, false)

	testTargetingString(targeting.Targeting_ENDS_WITH, "c", "abc", t, true, false)
	testTargetingString(targeting.Targeting_ENDS_WITH, "C", "abc", t, true, false)
	testTargetingString(targeting.Targeting_ENDS_WITH, "d", "abc", t, false, false)
	testTargetingString(targeting.Targeting_ENDS_WITH, "a", "abc", t, false, false)
	testTargetingString(targeting.Targeting_ENDS_WITH, "", "abc", t, true, false)

	testTargetingString(targeting.Targeting_STARTS_WITH, "a", "abc", t, true, false)
	testTargetingString(targeting.Targeting_STARTS_WITH, "A", "abc", t, true, false)
	testTargetingString(targeting.Targeting_STARTS_WITH, "d", "abc", t, false, false)
	testTargetingString(targeting.Targeting_STARTS_WITH, "c", "abc", t, false, false)
	testTargetingString(targeting.Targeting_STARTS_WITH, "", "abc", t, true, false)
}

// TestListStringTargeting checks all possible string list targeting
func TestListStringTargeting(t *testing.T) {
	testTargetingListString(targeting.Targeting_EQUALS, []string{"abc"}, "abd", t, false, false)
	testTargetingListString(targeting.Targeting_EQUALS, []string{"abc"}, "abc", t, true, false)
	testTargetingListString(targeting.Targeting_NOT_EQUALS, []string{"abc"}, "abd", t, true, false)
	testTargetingListString(targeting.Targeting_NOT_EQUALS, []string{"abc"}, "abc", t, false, false)

	testTargetingListString(targeting.Targeting_EQUALS, []string{"abc", "bcd"}, "abd", t, false, false)
	testTargetingListString(targeting.Targeting_EQUALS, []string{"abc", "bcd"}, "abc", t, true, false)
	testTargetingListString(targeting.Targeting_NOT_EQUALS, []string{"abc", "bcd"}, "abd", t, true, false)
	testTargetingListString(targeting.Targeting_NOT_EQUALS, []string{"abc", "bcd"}, "abc", t, false, false)

	testTargetingListString(targeting.Targeting_CONTAINS, []string{"abc", "bcd"}, "abcd", t, true, false)
	testTargetingListString(targeting.Targeting_CONTAINS, []string{"abc", "bcd"}, "xyz", t, false, false)
	testTargetingListString(targeting.Targeting_NOT_CONTAINS, []string{"abc", "bcd"}, "xyz", t, true, false)
	testTargetingListString(targeting.Targeting_NOT_CONTAINS, []string{"abc", "bcd"}, "abcd", t, false, false)
}

// TestListStringTargeting checks all possible string list targeting
func TestContextListStringTargeting(t *testing.T) {
	testTargetingContextListString(targeting.Targeting_EQUALS, "abc", []string{"abd"}, t, false, false)
	testTargetingContextListString(targeting.Targeting_EQUALS, "abc", []string{"abc"}, t, true, false)
	testTargetingContextListString(targeting.Targeting_NOT_EQUALS, "abc", []string{"abd"}, t, true, false)
	testTargetingContextListString(targeting.Targeting_NOT_EQUALS, "abc", []string{"abc"}, t, false, false)

	testTargetingContextListString(targeting.Targeting_EQUALS, "abd", []string{"abc", "bcd"}, t, false, false)
	testTargetingContextListString(targeting.Targeting_EQUALS, "abc", []string{"abc", "bcd"}, t, true, false)
	testTargetingContextListString(targeting.Targeting_NOT_EQUALS, "abd", []string{"abc", "bcd"}, t, true, false)
	testTargetingContextListString(targeting.Targeting_NOT_EQUALS, "abc", []string{"abc", "bcd"}, t, false, false)

	testTargetingContextListString(targeting.Targeting_CONTAINS, "abc", []string{"abcd", "bcd"}, t, true, false)
	testTargetingContextListString(targeting.Targeting_CONTAINS, "xyz", []string{"abc", "bcd"}, t, false, false)
	testTargetingContextListString(targeting.Targeting_NOT_CONTAINS, "xyz", []string{"abc", "bcd"}, t, true, false)
	testTargetingContextListString(targeting.Targeting_NOT_CONTAINS, "abc", []string{"abcd", "bcd"}, t, false, false)
}

func TestComplexTargeting(t *testing.T) {
	vgTest := &VariationGroup{
		ID: "test-vg",
		Targetings: &targeting.Targeting{
			TargetingGroups: []*targeting.Targeting_TargetingGroup{
				&targeting.Targeting_TargetingGroup{
					Targetings: []*targeting.Targeting_InnerTargeting{
						&targeting.Targeting_InnerTargeting{
							Operator: targeting.Targeting_EQUALS,
							Key:      &wrapperspb.StringValue{Value: "featureType"},
							Value: &structpb.Value{
								Kind: &structpb.Value_StringValue{
									StringValue: "deployment",
								},
							},
						}, &targeting.Targeting_InnerTargeting{
							Operator: targeting.Targeting_EQUALS,
							Key:      &wrapperspb.StringValue{Value: "accountName"},
							Value: &structpb.Value{
								Kind: &structpb.Value_StringValue{
									StringValue: "Flagship Demo",
								},
							},
						}, &targeting.Targeting_InnerTargeting{
							Operator: targeting.Targeting_CONTAINS,
							Key:      &wrapperspb.StringValue{Value: "fs_users"},
							Value: &structpb.Value{
								Kind: &structpb.Value_StringValue{
									StringValue: "@abtasty.com",
								},
							},
						},
					},
				},

				&targeting.Targeting_TargetingGroup{
					Targetings: []*targeting.Targeting_InnerTargeting{
						&targeting.Targeting_InnerTargeting{
							Operator: targeting.Targeting_EQUALS,
							Key:      &wrapperspb.StringValue{Value: "isVIP"},
							Value: &structpb.Value{
								Kind: &structpb.Value_BoolValue{
									BoolValue: true,
								},
							},
						},
					},
				},
			},
		},
	}

	context := map[string]*structpb.Value{}
	context["accountName"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: "Flagship Demo",
		},
	}
	context["featureType"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: "deployment",
		},
	}
	test, err := targetingMatch(vgTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.True(t, test)

	context["featureType"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: "ab",
		},
	}
	test, err = targetingMatch(vgTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.False(t, test)

	context["isVIP"] = &structpb.Value{
		Kind: &structpb.Value_BoolValue{
			BoolValue: true,
		},
	}
	test, err = targetingMatch(vgTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.True(t, test)
}
