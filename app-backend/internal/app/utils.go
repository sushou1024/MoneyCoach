package app

import "sort"

func mapKeys[T any](input map[string]T) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func futuresSymbols(input map[string]futuresPremiumIndex) []string {
	values := make([]string, 0, len(input))
	for _, item := range input {
		if item.Symbol == "" {
			continue
		}
		values = append(values, item.Symbol)
	}
	sort.Strings(values)
	return values
}
