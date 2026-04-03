package app

import "encoding/json"

func planLayers(params map[string]any) []map[string]any {
	layers, ok := params["layers"].([]map[string]any)
	if ok {
		return layers
	}
	var result []map[string]any
	if raw, ok := params["layers"].([]any); ok {
		for _, entry := range raw {
			if layer, ok := entry.(map[string]any); ok {
				result = append(result, layer)
			}
		}
	}
	return result
}

func planSafetyOrders(params map[string]any) []map[string]any {
	orders, ok := params["safety_orders"].([]map[string]any)
	if ok {
		return orders
	}
	var result []map[string]any
	if raw, ok := params["safety_orders"].([]any); ok {
		for _, entry := range raw {
			if order, ok := entry.(map[string]any); ok {
				result = append(result, order)
			}
		}
	}
	return result
}

func planAdditions(params map[string]any) []map[string]any {
	additions, ok := params["additions"].([]map[string]any)
	if ok {
		return additions
	}
	var result []map[string]any
	if raw, ok := params["additions"].([]any); ok {
		for _, entry := range raw {
			if addition, ok := entry.(map[string]any); ok {
				result = append(result, addition)
			}
		}
	}
	return result
}

func getFloatParam(params map[string]any, key string) (float64, bool) {
	value, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		parsed, err := v.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func getStringParam(params map[string]any, key string) string {
	value, ok := params[key]
	if !ok {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
