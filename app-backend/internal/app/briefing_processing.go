package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type briefingPromptInput struct {
	UserID       string                `json:"user_id"`
	BaseCurrency string                `json:"base_currency"`
	NetWorthUSD  float64               `json:"net_worth_usd"`
	Holdings     []briefingHoldingItem `json:"holdings"`
	ValuationAt  string                `json:"valuation_at"`
}

type briefingHoldingItem struct {
	Symbol    string  `json:"symbol"`
	AssetType string  `json:"asset_type"`
	Amount    float64 `json:"amount"`
	ValueUSD  float64 `json:"value_usd"`
}

type briefingItem struct {
	Type     string `json:"type"`
	Priority int    `json:"priority"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	PushText string `json:"push_text"`
}

func (s *Server) processDailyBriefing(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	if !s.hasActiveEntitlement(ctx, userID) {
		return nil
	}

	var user User
	if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return err
	}
	if user.ActivePortfolioSnapshot == nil {
		return nil
	}

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return err
	}

	snapshot, holdings, err := s.loadSnapshotWithHoldings(ctx, *user.ActivePortfolioSnapshot)
	if err != nil {
		return err
	}
	if len(holdings) == 0 {
		return nil
	}

	// Build the input for Gemini.
	baseCurrency := normalizeCurrency(profile.BaseCurrency)
	if baseCurrency == "" {
		baseCurrency = "USD"
	}

	holdingItems := make([]briefingHoldingItem, 0, len(holdings))
	for _, h := range holdings {
		holdingItems = append(holdingItems, briefingHoldingItem{
			Symbol:    h.Symbol,
			AssetType: h.AssetType,
			Amount:    h.Amount,
			ValueUSD:  h.ValueUSD,
		})
	}

	input := briefingPromptInput{
		UserID:       userID,
		BaseCurrency: baseCurrency,
		NetWorthUSD:  snapshot.NetWorthUSD,
		Holdings:     holdingItems,
		ValuationAt:  snapshot.ValuationAsOf.Format(time.RFC3339),
	}

	outputLanguage := resolveOutputLanguage(profile.Language, "")
	prompt := applyOutputLanguage(s.prompts.DailyBriefing, outputLanguage)
	experienceLevel := profile.Experience
	if experienceLevel == "" {
		experienceLevel = "intermediate"
	}
	prompt = strings.ReplaceAll(prompt, "{{EXPERIENCE_LEVEL}}", experienceLevel)

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: prompt}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: mustJSON(input)}}}},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.6,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	var parsed []briefingItem
	_, err = s.gemini.callGeminiJSON(ctx, request, &parsed)
	if err != nil {
		return fmt.Errorf("daily briefing gemini error user=%s: %w", userID, err)
	}

	if len(parsed) == 0 {
		return nil
	}

	// Determine the briefing date in the user's local timezone.
	loc := parseTZ(profile.Timezone)
	briefingDate := time.Now().In(loc).Format("2006-01-02")

	now := time.Now().UTC()
	records := make([]Briefing, 0, len(parsed))
	for _, item := range parsed {
		records = append(records, Briefing{
			ID:           newID("brf"),
			UserID:       userID,
			Type:         item.Type,
			Priority:     item.Priority,
			Title:        item.Title,
			Body:         item.Body,
			PushText:     item.PushText,
			BriefingDate: briefingDate,
			CreatedAt:    now,
		})
	}

	// Save all briefings.
	if err := s.db.DB().WithContext(ctx).Create(&records).Error; err != nil {
		return fmt.Errorf("daily briefing save error user=%s: %w", userID, err)
	}

	// Pick the top 2 by priority for push notification (priority 1 = highest).
	sort.Slice(records, func(i, j int) bool {
		return records[i].Priority < records[j].Priority
	})
	pushCount := 2
	if len(records) < pushCount {
		pushCount = len(records)
	}
	topBriefings := records[:pushCount]

	if err := s.sendBriefingPushNotifications(ctx, userID, topBriefings); err != nil {
		s.logger.Printf("daily briefing push error user=%s err=%v", userID, err)
	}

	return nil
}
