package main

import "strings"

const outputLanguageToken = "{{OUTPUT_LANGUAGE}}"

func applyOutputLanguage(prompt, outputLanguage string) string {
	if strings.TrimSpace(outputLanguage) == "" {
		outputLanguage = "English"
	}
	if strings.Contains(prompt, outputLanguageToken) {
		return strings.ReplaceAll(prompt, outputLanguageToken, outputLanguage)
	}
	return strings.TrimSpace(prompt) + "\n\nOUTPUT_LANGUAGE = \"" + outputLanguage + "\""
}
