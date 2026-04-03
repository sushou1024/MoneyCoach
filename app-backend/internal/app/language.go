package app

import (
	"strings"
)

const outputLanguageToken = "{{OUTPUT_LANGUAGE}}"

func resolveOutputLanguage(profileLanguage, acceptLanguage string) string {
	if lang := strings.TrimSpace(profileLanguage); lang != "" {
		return normalizeOutputLanguage(lang)
	}
	if acceptLanguage != "" {
		parts := strings.Split(acceptLanguage, ",")
		if len(parts) > 0 {
			return normalizeOutputLanguage(strings.TrimSpace(parts[0]))
		}
	}
	return "English"
}

func normalizeOutputLanguage(tag string) string {
	lower := strings.ToLower(strings.TrimSpace(tag))
	switch {
	case strings.HasPrefix(lower, "en"):
		return "English"
	case lower == "zh-cn" || lower == "zh-hans":
		return "Simplified Chinese"
	case lower == "zh-tw" || lower == "zh-hant":
		return "Traditional Chinese"
	case strings.HasPrefix(lower, "ja"):
		return "Japanese"
	case strings.HasPrefix(lower, "ko"):
		return "Korean"
	default:
		return "English"
	}
}

func applyOutputLanguage(prompt, outputLanguage string) string {
	if strings.TrimSpace(outputLanguage) == "" {
		outputLanguage = "English"
	}
	if strings.Contains(prompt, outputLanguageToken) {
		return strings.ReplaceAll(prompt, outputLanguageToken, outputLanguage)
	}
	return strings.TrimSpace(prompt) + "\n\nOUTPUT_LANGUAGE = \"" + outputLanguage + "\""
}
