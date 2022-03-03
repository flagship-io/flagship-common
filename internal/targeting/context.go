package targeting

import (
	"google.golang.org/protobuf/types/known/structpb"
)

type Context struct {
	Rows      ContextRows
	Providers map[string]ContextRows
}

type ContextRows map[string]*structpb.Value

func (c *Context) GetValueByProvider(key string, provider string) (*structpb.Value, bool) {
	if provider == "" {
		value, exists := c.Rows[key]
		return value, exists
	} else {
		if ok := c.Providers[provider]; ok == nil {
			return nil, false
		}
		value, exists := c.Providers[provider][key]
		return value, exists
	}
}
