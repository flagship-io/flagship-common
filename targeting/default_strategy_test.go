package targeting

import (
	"testing"

	targetingProto "github.com/flagship-io/flagship-proto/targeting"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func testTargetingNumber(operator targetingProto.Targeting_TargetingOperator, targetingValue float64, value float64, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorNumber(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %f, v: %f, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingBoolean(operator targetingProto.Targeting_TargetingOperator, targetingValue bool, value bool, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorBool(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingString(operator targetingProto.Targeting_TargetingOperator, targetingValue string, value string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	match, err := targetingMatchOperatorString(operator, targetingValue, value)

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting number %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, value, match, err)
	}
}

func testTargetingListString(operator targetingProto.Targeting_TargetingOperator, targetingValues []string, value string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	stringValues := structpb.ListValue{}
	for _, str := range targetingValues {
		stringValues.Values = append(stringValues.Values, structpb.NewStringValue(str))
	}

	match, err := targetingMatchOperator(operator, structpb.NewListValue(&stringValues), structpb.NewStringValue(value))

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting list %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValues, value, match, err)
	}
}

func testTargetingContextListString(operator targetingProto.Targeting_TargetingOperator, targetingValue string, contextValue []string, t *testing.T, shouldMatch bool, shouldRaiseError bool) {
	stringValues := structpb.ListValue{}
	for _, str := range contextValue {
		stringValues.Values = append(stringValues.Values, structpb.NewStringValue(str))
	}

	match, err := targetingMatchOperator(operator, structpb.NewStringValue(targetingValue), structpb.NewListValue(&stringValues))

	if ((err != nil && !shouldRaiseError) || (shouldRaiseError && err == nil)) || (match != shouldMatch) {
		t.Errorf("Targeting list %v not working - tv : %v, v: %v, match : %v, err: %v", operator, targetingValue, contextValue, match, err)
	}
}

// TestNumberTargeting checks all possible number targeting
func TestNumberTargeting(t *testing.T) {
	testTargetingNumber(targetingProto.Targeting_LOWER_THAN, 11, 10, t, true, false)
	testTargetingNumber(targetingProto.Targeting_LOWER_THAN, 10, 10, t, false, false)
	testTargetingNumber(targetingProto.Targeting_LOWER_THAN, 9, 10, t, false, false)

	testTargetingNumber(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, 11, 10, t, true, false)
	testTargetingNumber(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, 9, 10, t, false, false)

	testTargetingNumber(targetingProto.Targeting_GREATER_THAN, 11, 10, t, false, false)
	testTargetingNumber(targetingProto.Targeting_GREATER_THAN, 10, 10, t, false, false)
	testTargetingNumber(targetingProto.Targeting_GREATER_THAN, 9, 10, t, true, false)

	testTargetingNumber(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, 11, 10, t, false, false)
	testTargetingNumber(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, 9, 10, t, true, false)

	testTargetingNumber(targetingProto.Targeting_NOT_EQUALS, 11, 10, t, true, false)
	testTargetingNumber(targetingProto.Targeting_NOT_EQUALS, 10, 10, t, false, false)
	testTargetingNumber(targetingProto.Targeting_NOT_EQUALS, 9, 10, t, true, false)

	testTargetingNumber(targetingProto.Targeting_EQUALS, 11, 10, t, false, false)
	testTargetingNumber(targetingProto.Targeting_EQUALS, 10, 10, t, true, false)
	testTargetingNumber(targetingProto.Targeting_EQUALS, 9, 10, t, false, false)

	testTargetingNumber(targetingProto.Targeting_CONTAINS, 11, 10, t, false, true)
	testTargetingNumber(targetingProto.Targeting_ENDS_WITH, 10, 10, t, false, true)
	testTargetingNumber(targetingProto.Targeting_STARTS_WITH, 9, 10, t, false, true)
}

// TestBooleanTargeting checks all possible boolean targeting
func TestBooleanTargeting(t *testing.T) {
	testTargetingBoolean(targetingProto.Targeting_NOT_EQUALS, true, false, t, true, false)
	testTargetingBoolean(targetingProto.Targeting_NOT_EQUALS, true, true, t, false, false)
	testTargetingBoolean(targetingProto.Targeting_NOT_EQUALS, false, true, t, true, false)

	testTargetingBoolean(targetingProto.Targeting_EQUALS, true, false, t, false, false)
	testTargetingBoolean(targetingProto.Targeting_EQUALS, true, true, t, true, false)
	testTargetingBoolean(targetingProto.Targeting_EQUALS, false, true, t, false, false)

	testTargetingBoolean(targetingProto.Targeting_CONTAINS, true, false, t, false, true)
	testTargetingBoolean(targetingProto.Targeting_ENDS_WITH, true, false, t, false, true)
	testTargetingBoolean(targetingProto.Targeting_STARTS_WITH, true, false, t, false, true)
	testTargetingBoolean(targetingProto.Targeting_GREATER_THAN, true, false, t, false, true)
	testTargetingBoolean(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, true, false, t, false, true)
	testTargetingBoolean(targetingProto.Targeting_LOWER_THAN, true, false, t, false, true)
	testTargetingBoolean(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, true, false, t, false, true)
}

// TestStringTargeting checks all possible string targeting
func TestStringTargeting(t *testing.T) {
	testTargetingString(targetingProto.Targeting_LOWER_THAN, "abc", "abd", t, false, false)
	testTargetingString(targetingProto.Targeting_LOWER_THAN, "abc", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_LOWER_THAN, "abd", "abc", t, true, false)

	testTargetingString(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, "abc", "abd", t, false, false)
	testTargetingString(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, "abc", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_LOWER_THAN_OR_EQUALS, "abd", "abc", t, true, false)

	testTargetingString(targetingProto.Targeting_GREATER_THAN, "abc", "abd", t, true, false)
	testTargetingString(targetingProto.Targeting_GREATER_THAN, "abc", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_GREATER_THAN, "abd", "abc", t, false, false)

	testTargetingString(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(targetingProto.Targeting_GREATER_THAN_OR_EQUALS, "abd", "abc", t, false, false)

	testTargetingString(targetingProto.Targeting_NOT_EQUALS, "abc", "abd", t, true, false)
	testTargetingString(targetingProto.Targeting_NOT_EQUALS, "abc", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_NOT_EQUALS, "", "", t, false, false)
	testTargetingString(targetingProto.Targeting_NOT_EQUALS, "", " ", t, true, false)

	testTargetingString(targetingProto.Targeting_EQUALS, "abc", "abd", t, false, false)
	testTargetingString(targetingProto.Targeting_EQUALS, "abc", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_EQUALS, "ABC", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_EQUALS, "", "", t, true, false)
	testTargetingString(targetingProto.Targeting_EQUALS, "", " ", t, false, false)

	testTargetingString(targetingProto.Targeting_CONTAINS, "b", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_CONTAINS, "B", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_CONTAINS, "d", "abc", t, false, false)

	testTargetingString(targetingProto.Targeting_ENDS_WITH, "c", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_ENDS_WITH, "C", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_ENDS_WITH, "d", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_ENDS_WITH, "a", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_ENDS_WITH, "", "abc", t, true, false)

	testTargetingString(targetingProto.Targeting_STARTS_WITH, "a", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_STARTS_WITH, "A", "abc", t, true, false)
	testTargetingString(targetingProto.Targeting_STARTS_WITH, "d", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_STARTS_WITH, "c", "abc", t, false, false)
	testTargetingString(targetingProto.Targeting_STARTS_WITH, "", "abc", t, true, false)
}

// TestListStringTargeting checks all possible string list targeting
func TestListStringTargeting(t *testing.T) {
	testTargetingListString(targetingProto.Targeting_EQUALS, []string{"abc"}, "abd", t, false, false)
	testTargetingListString(targetingProto.Targeting_EQUALS, []string{"abc"}, "abc", t, true, false)
	testTargetingListString(targetingProto.Targeting_NOT_EQUALS, []string{"abc"}, "abd", t, true, false)
	testTargetingListString(targetingProto.Targeting_NOT_EQUALS, []string{"abc"}, "abc", t, false, false)

	testTargetingListString(targetingProto.Targeting_EQUALS, []string{"abc", "bcd"}, "abd", t, false, false)
	testTargetingListString(targetingProto.Targeting_EQUALS, []string{"abc", "bcd"}, "abc", t, true, false)
	testTargetingListString(targetingProto.Targeting_NOT_EQUALS, []string{"abc", "bcd"}, "abd", t, true, false)
	testTargetingListString(targetingProto.Targeting_NOT_EQUALS, []string{"abc", "bcd"}, "abc", t, false, false)

	testTargetingListString(targetingProto.Targeting_CONTAINS, []string{"abc", "bcd"}, "abcd", t, true, false)
	testTargetingListString(targetingProto.Targeting_CONTAINS, []string{"abc", "bcd"}, "xyz", t, false, false)
	testTargetingListString(targetingProto.Targeting_NOT_CONTAINS, []string{"abc", "bcd"}, "xyz", t, true, false)
	testTargetingListString(targetingProto.Targeting_NOT_CONTAINS, []string{"abc", "bcd"}, "abcd", t, false, false)
}

// TestListStringTargeting checks all possible string list targeting
func TestContextListStringTargeting(t *testing.T) {
	testTargetingContextListString(targetingProto.Targeting_EQUALS, "abc", []string{"abd"}, t, false, false)
	testTargetingContextListString(targetingProto.Targeting_EQUALS, "abc", []string{"abc"}, t, true, false)
	testTargetingContextListString(targetingProto.Targeting_NOT_EQUALS, "abc", []string{"abd"}, t, true, false)
	testTargetingContextListString(targetingProto.Targeting_NOT_EQUALS, "abc", []string{"abc"}, t, false, false)

	testTargetingContextListString(targetingProto.Targeting_EQUALS, "abd", []string{"abc", "bcd"}, t, false, false)
	testTargetingContextListString(targetingProto.Targeting_EQUALS, "abc", []string{"abc", "bcd"}, t, true, false)
	testTargetingContextListString(targetingProto.Targeting_NOT_EQUALS, "abd", []string{"abc", "bcd"}, t, true, false)
	testTargetingContextListString(targetingProto.Targeting_NOT_EQUALS, "abc", []string{"abc", "bcd"}, t, false, false)

	testTargetingContextListString(targetingProto.Targeting_CONTAINS, "abc", []string{"abcd", "bcd"}, t, true, false)
	testTargetingContextListString(targetingProto.Targeting_CONTAINS, "xyz", []string{"abc", "bcd"}, t, false, false)
	testTargetingContextListString(targetingProto.Targeting_NOT_CONTAINS, "xyz", []string{"abc", "bcd"}, t, true, false)
	testTargetingContextListString(targetingProto.Targeting_NOT_CONTAINS, "abc", []string{"abcd", "bcd"}, t, false, false)
}

// TestEmptyContextTargeting checks all possible empty context targeting
func TestEmptyContextTargeting(t *testing.T) {
	_, err := targetingMatchOperatorEmptyContext(targetingProto.Targeting_CONTAINS, true, nil)
	assert.NotNil(t, err)

	match, err := targetingMatchOperatorEmptyContext(targetingProto.Targeting_EXISTS, true, nil)
	assert.Nil(t, err)
	assert.False(t, match)

	match, err = targetingMatchOperatorEmptyContext(targetingProto.Targeting_EXISTS, true, &structpb.Value{})
	assert.Nil(t, err)
	assert.True(t, match)

	match, err = targetingMatchOperatorEmptyContext(targetingProto.Targeting_EXISTS, false, &structpb.Value{})
	assert.Nil(t, err)
	assert.False(t, match)

	match, err = targetingMatchOperatorEmptyContext(targetingProto.Targeting_NOT_EXISTS, true, nil)
	assert.Nil(t, err)
	assert.True(t, match)

	match, err = targetingMatchOperatorEmptyContext(targetingProto.Targeting_NOT_EXISTS, true, &structpb.Value{})
	assert.Nil(t, err)
	assert.False(t, match)

	match, err = targetingMatchOperatorEmptyContext(targetingProto.Targeting_NOT_EXISTS, false, &structpb.Value{})
	assert.Nil(t, err)
	assert.True(t, match)
}
