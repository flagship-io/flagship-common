package targeting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGetValueByProvider(t *testing.T) {
	context := Context{
		Rows: ContextRows{
			"feature": structpb.NewStringValue("test"),
		},
		Providers: map[string]ContextRows{
			"myprovider": {
				"gender": structpb.NewStringValue("male"),
			},
		},
	}

	// true

	value, ok := context.GetValueByProvider("feature", "")
	assert.True(t, ok)
	assert.Equal(t, structpb.NewStringValue("test"), value)

	value, ok = context.GetValueByProvider("gender", "myprovider")
	assert.True(t, ok)
	assert.Equal(t, structpb.NewStringValue("male"), value)

	// false

	value, ok = context.GetValueByProvider("gender", "")
	assert.False(t, ok)
	assert.Nil(t, value)

	value, ok = context.GetValueByProvider("feature", "myprovider")
	assert.False(t, ok)
	assert.Nil(t, value)

}
