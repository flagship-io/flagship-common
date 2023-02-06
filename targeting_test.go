package decision

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flagship-io/flagship-common/targeting"
	targetingProto "github.com/flagship-io/flagship-proto/targeting"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestAllUsersTargeting(t *testing.T) {
	targetingsTest := &targetingProto.Targeting{
		TargetingGroups: []*targetingProto.Targeting_TargetingGroup{
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					// AND
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "fs_all_users"},
						Value:    structpb.NewStringValue(""),
					},
				},
			},
		},
	}

	match, err := targetingMatch(targetingsTest, "visitor_id", &targeting.Context{})
	assert.Nil(t, err)
	assert.True(t, match)

	targetingsTest.TargetingGroups[0].Targetings = append(targetingsTest.TargetingGroups[0].Targetings, &targetingProto.Targeting_InnerTargeting{
		Operator: targetingProto.Targeting_EQUALS,
		Key:      &wrapperspb.StringValue{Value: "key"},
		Value:    structpb.NewStringValue("value"),
	})

	match, err = targetingMatch(targetingsTest, "visitor_id", &targeting.Context{})
	assert.Nil(t, err)
	assert.False(t, match)

	match, err = targetingMatch(targetingsTest, "visitor_id", &targeting.Context{
		Standard: targeting.ContextMap{
			"key": structpb.NewStringValue("value"),
		},
	})
	assert.Nil(t, err)
	assert.True(t, match)
}

func TestTargetingMatch(t *testing.T) {
	targetingConf := &targetingProto.Targeting{
		TargetingGroups: []*targetingProto.Targeting_TargetingGroup{
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					{
						Operator: targetingProto.Targeting_NOT_EXISTS,
						Key:      wrapperspb.String("test"),
						Value:    structpb.NewBoolValue(true),
					},
				},
			},
		},
	}

	// test context without provider (default behavior)
	context := &targeting.Context{
		Standard: targeting.ContextMap{},
	}
	match, err := targetingMatch(targetingConf, "visitor_id", context)
	assert.Nil(t, err)
	assert.True(t, match)

	context.Standard["test"] = structpb.NewBoolValue(true)
	match, err = targetingMatch(targetingConf, "visitor_id", context)
	assert.Nil(t, err)
	assert.False(t, match)

	targetingConf.TargetingGroups[0].Targetings[0].Operator = targetingProto.Targeting_EXISTS
	match, err = targetingMatch(targetingConf, "visitor_id", context)
	assert.Nil(t, err)
	assert.True(t, match)
}

func TestSemverTargetingMatch(t *testing.T) {
	targetingConf := &targetingProto.Targeting{
		TargetingGroups: []*targetingProto.Targeting_TargetingGroup{
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					{
						Operator: targetingProto.Targeting_GREATER_THAN,
						Key:      wrapperspb.String("semverAppVersion"),
						Value:    structpb.NewStringValue("2.1.0"),
					},
				},
			},
		},
	}

	// test context without provider (default behavior)
	context := &targeting.Context{
		Standard: targeting.ContextMap{
			"semverAppVersion": structpb.NewStringValue("10.1.0"),
		},
	}
	match, err := targetingMatch(targetingConf, "visitor_id", context)
	assert.Nil(t, err)
	assert.True(t, match)
}

func TestComplexTargeting(t *testing.T) {
	targetingsTest := &targetingProto.Targeting{
		TargetingGroups: []*targetingProto.Targeting_TargetingGroup{
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "featureType"},
						Value:    structpb.NewStringValue("deployment"),
					},
					// AND
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "accountName"},
						Value:    structpb.NewStringValue("Flagship Demo"),
					},
					// AND
					{
						Operator: targetingProto.Targeting_CONTAINS,
						Key:      &wrapperspb.StringValue{Value: "fs_users"},
						Value:    structpb.NewStringValue("@abtasty.com"),
					},
				},
			},
			// OR
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "isVIP"},
						Value:    structpb.NewBoolValue(true),
					},
				},
			},
			// OR
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "gender"},
						Value:    structpb.NewStringValue("female"),
						Provider: &wrapperspb.StringValue{Value: "myprovider"},
					},
				},
			},
			// OR
			{
				Targetings: []*targetingProto.Targeting_InnerTargeting{
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "gender"},
						Value:    structpb.NewStringValue("male"),
						Provider: &wrapperspb.StringValue{Value: "myprovider"},
					},
					{
						Operator: targetingProto.Targeting_EQUALS,
						Key:      &wrapperspb.StringValue{Value: "country"},
						Value:    structpb.NewStringValue("FR"),
						Provider: &wrapperspb.StringValue{Value: "myprovider"},
					},
				},
			},
		},
	}

	// test context without provider (default behavior)
	context := &targeting.Context{
		Standard: targeting.ContextMap{},
	}
	context.Standard["accountName"] = structpb.NewStringValue("Flagship Demo")
	context.Standard["featureType"] = structpb.NewStringValue("deployment")
	test, err := targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.True(t, test)

	context.Standard["featureType"] = structpb.NewStringValue("ab")
	test, err = targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.False(t, test)

	context.Standard["gender"] = structpb.NewStringValue("female")
	test, err = targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.False(t, test)

	context.Standard["isVIP"] = structpb.NewBoolValue(true)
	test, err = targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.True(t, test)

	// test context with provider (mixpanel, segment)
	context = &targeting.Context{
		Standard: targeting.ContextMap{},
		IntegrationProviders: map[string]targeting.ContextMap{
			"myprovider": {},
		},
	}
	context.IntegrationProviders["myprovider"] = targeting.ContextMap{
		"gender": structpb.NewStringValue("female"),
	}
	test, err = targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.True(t, test)

	context.IntegrationProviders["myprovider"]["gender"] = structpb.NewStringValue("male")
	context.IntegrationProviders["myprovider"]["country"] = structpb.NewStringValue("EN")
	test, err = targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.False(t, test)

	context.IntegrationProviders["myprovider"]["country"] = structpb.NewStringValue("FR")
	test, err = targetingMatch(targetingsTest, "test@abtasty.com", context)
	assert.Nil(t, err)
	assert.True(t, test)
}
