package app

const (
	fingerprintWarningText = "This may double-count holdings unless it is a different account."
	phashWarningText       = "This image looks similar to another upload; confirm it is a different account."
)

func buildImageWarnings(images []UploadImage, imageByID map[string]ocrImage, phashByID map[string]string) map[string][]string {
	warnings := make(map[string][]string)
	seenFingerprint := make(map[string]struct{})
	for _, image := range images {
		result, ok := imageByID[image.ID]
		if !ok {
			continue
		}
		if normalizeOCRStatus(result.Status) != "success" {
			continue
		}
		fingerprint := computeFingerprintV0(result.PlatformGuess, result.Assets)
		if fingerprint == "" {
			continue
		}
		if _, exists := seenFingerprint[fingerprint]; exists {
			warnings[image.ID] = appendWarning(warnings[image.ID], fingerprintWarningText)
			continue
		}
		seenFingerprint[fingerprint] = struct{}{}
	}

	seenHashes := make([]string, 0, len(images))
	for _, image := range images {
		hash := phashByID[image.ID]
		if hash == "" {
			continue
		}
		for _, prev := range seenHashes {
			distance, err := hammingDistanceHex(hash, prev)
			if err != nil {
				continue
			}
			if distance <= 8 {
				warnings[image.ID] = appendWarning(warnings[image.ID], phashWarningText)
				break
			}
		}
		seenHashes = append(seenHashes, hash)
	}

	return warnings
}

func appendWarning(values []string, warning string) []string {
	for _, existing := range values {
		if existing == warning {
			return values
		}
	}
	return append(values, warning)
}
