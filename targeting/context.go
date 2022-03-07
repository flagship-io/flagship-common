package targeting

import (
	"google.golang.org/protobuf/types/known/structpb"
)

type Context struct {
	Standard             ContextMap
	IntegrationProviders map[string]ContextMap
}

type ContextMap map[string]*structpb.Value

func (c *Context) GetValueByProvider(key string, provider string) (*structpb.Value, bool) {
	if provider == "" {
		value, exists := c.Standard[key]
		return value, exists
	} else {
		if ok := c.IntegrationProviders[provider]; ok == nil {
			return nil, false
		}
		value, exists := c.IntegrationProviders[provider][key]
		return value, exists
	}
}
