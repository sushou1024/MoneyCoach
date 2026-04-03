package app

type ambiguityKey struct {
	SymbolRawNormalized string
	PlatformCategory    string
}

func newAmbiguityKey(symbolRaw, platformCategory string) ambiguityKey {
	return ambiguityKey{
		SymbolRawNormalized: normalizeSymbol(symbolRaw),
		PlatformCategory:    platformCategory,
	}
}
