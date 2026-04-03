package data

import _ "embed"

// XHKGCSV provides the embedded Hong Kong stock name mapping data.
//
//go:embed XHKG.csv
var XHKGCSV []byte
