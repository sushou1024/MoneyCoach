package app

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
	"sync"

	xhkgdata "github.com/jackcpku/moneycoach/app-backend/data"
)

type hkChineseName struct {
	Traditional string
	Simplified  string
}

var (
	hkNamesOnce   sync.Once
	hkNamesByCode map[string]hkChineseName
)

func hkNameForLanguage(symbol, language string) (string, bool) {
	if !isChineseLanguage(language) {
		return "", false
	}
	code := hkCodeFromSymbol(symbol)
	if code == "" {
		return "", false
	}
	names, ok := hkNames()[code]
	if !ok {
		return "", false
	}
	if isTraditionalChinese(language) && names.Traditional != "" {
		return names.Traditional, true
	}
	if isSimplifiedChinese(language) && names.Simplified != "" {
		return names.Simplified, true
	}
	return "", false
}

func hkNames() map[string]hkChineseName {
	hkNamesOnce.Do(func() {
		hkNamesByCode = parseXHKGCSV(xhkgdata.XHKGCSV)
	})
	if hkNamesByCode == nil {
		return map[string]hkChineseName{}
	}
	return hkNamesByCode
}

func parseXHKGCSV(raw []byte) map[string]hkChineseName {
	out := make(map[string]hkChineseName)
	if len(raw) == 0 {
		return out
	}
	reader := csv.NewReader(bytes.NewReader(raw))
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 3 {
			continue
		}
		code := normalizeHKCode(record[0])
		if code == "" {
			continue
		}
		traditional := strings.TrimSpace(record[1])
		simplified := strings.TrimSpace(record[2])
		if traditional == "" && simplified == "" {
			continue
		}
		out[code] = hkChineseName{
			Traditional: traditional,
			Simplified:  simplified,
		}
	}
	return out
}

func normalizeHKCode(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return ""
		}
	}
	trimmed = strings.TrimLeft(trimmed, "0")
	if trimmed == "" {
		return ""
	}
	return trimmed
}

func hkCodeFromSymbol(symbol string) string {
	normalized, ok := normalizeHKSymbol(symbol)
	if !ok {
		return ""
	}
	return strings.TrimSuffix(normalized, ".HK")
}

func isSimplifiedChinese(language string) bool {
	tag := normalizeLanguageTag(language)
	if strings.Contains(tag, "simplified") || strings.Contains(tag, "zh-cn") || strings.Contains(tag, "zh-hans") {
		return true
	}
	return strings.Contains(language, "简体") || strings.Contains(language, "簡體")
}

func isTraditionalChinese(language string) bool {
	tag := normalizeLanguageTag(language)
	if strings.Contains(tag, "traditional") || strings.Contains(tag, "zh-tw") || strings.Contains(tag, "zh-hk") || strings.Contains(tag, "zh-hant") {
		return true
	}
	return strings.Contains(language, "繁體") || strings.Contains(language, "繁体")
}

func isChineseLanguage(language string) bool {
	return isSimplifiedChinese(language) || isTraditionalChinese(language)
}

func normalizeLanguageTag(language string) string {
	normalized := strings.ToLower(strings.TrimSpace(language))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}
