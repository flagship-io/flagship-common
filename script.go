package decision

import (
	"github.com/dop251/goja"
	"github.com/flagship-io/flagship-common/targeting"
	"google.golang.org/protobuf/types/known/structpb"
)

var vm = goja.New()

type scriptingContext struct {
	VisitorID      string
	VisitorContext *targeting.Context
}

func computeValue(value *structpb.Value, context *scriptingContext) (*structpb.Value, error) {
	switch value.Kind.(type) {
	case *structpb.Value_StructValue:
		structValue := value.GetStructValue()
		if t, ok := structValue.Fields["type"]; ok && t.GetStringValue() == "script" {
			contextMap := map[string]interface{}{}
			for k, v := range context.VisitorContext.Standard {
				contextMap[k] = v.AsInterface()
			}
			for _, integrationContext := range context.VisitorContext.IntegrationProviders {
				for k, v := range integrationContext {
					contextMap[k] = v.AsInterface()
				}
			}
			vm.Set("$visitor", map[string]interface{}{
				"id":      context.VisitorID,
				"context": contextMap,
			})
			script, ok := structValue.Fields["script"]
			if ok {
				v, err := vm.RunString(script.GetStringValue())
				if err != nil {
					logger.Logf(InfoLevel, "")
					return value, nil
				}
				return structpb.NewValue(v.Export())
			}
		}
		return value, nil
	default:
		return value, nil
	}
}
