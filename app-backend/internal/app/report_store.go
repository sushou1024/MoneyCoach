package app

import (
	"context"
	"encoding/json"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (s *Server) loadLockedPlansForSnapshot(ctx context.Context, snapshotID string) (string, []lockedPlan, map[string]string, error) {
	if snapshotID == "" {
		return "", nil, nil, nil
	}
	var calc Calculation
	if err := s.db.DB().WithContext(ctx).
		Where("portfolio_snapshot_id = ? AND status_paid = ?", snapshotID, "ready").
		Order("created_at desc").
		First(&calc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, nil, nil
		}
		return "", nil, nil, err
	}

	strategyRows, err := s.loadReportStrategies(ctx, calc.ID)
	if err != nil {
		return "", nil, nil, err
	}
	plans := buildLockedPlansFromStrategyRows(strategyRows)

	riskByID := make(map[string]string)
	var riskRows []ReportRisk
	if err := s.db.DB().WithContext(ctx).Where("calculation_id = ?", calc.ID).Find(&riskRows).Error; err != nil {
		return "", plans, nil, err
	}
	for _, row := range riskRows {
		riskByID[row.RiskID] = row.Severity
	}
	return calc.ID, plans, riskByID, nil
}

func (s *Server) loadLatestPaidPlansForUser(ctx context.Context, userID string) ([]lockedPlan, error) {
	if userID == "" {
		return nil, nil
	}
	var calc Calculation
	if err := s.db.DB().WithContext(ctx).
		Joins("JOIN portfolio_snapshots ON portfolio_snapshots.id = calculations.portfolio_snapshot_id").
		Where("portfolio_snapshots.user_id = ? AND calculations.status_paid = ?", userID, "ready").
		Order("calculations.paid_at desc nulls last, calculations.created_at desc").
		First(&calc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	strategyRows, err := s.loadReportStrategies(ctx, calc.ID)
	if err != nil {
		return nil, err
	}
	return buildLockedPlansFromStrategyRows(strategyRows), nil
}

func (s *Server) loadLatestPaidCalculationID(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", nil
	}
	var calc Calculation
	if err := s.db.DB().WithContext(ctx).
		Joins("JOIN portfolio_snapshots ON portfolio_snapshots.id = calculations.portfolio_snapshot_id").
		Where("portfolio_snapshots.user_id = ? AND calculations.status_paid = ?", userID, "ready").
		Order("calculations.paid_at desc nulls last, calculations.created_at desc").
		First(&calc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return calc.ID, nil
}

func (s *Server) loadReportStrategies(ctx context.Context, calculationID string) ([]ReportStrategy, error) {
	var strategyRows []ReportStrategy
	if err := s.db.DB().WithContext(ctx).Where("calculation_id = ?", calculationID).Find(&strategyRows).Error; err != nil {
		return nil, err
	}
	return strategyRows, nil
}

func buildLockedPlansFromStrategyRows(strategyRows []ReportStrategy) []lockedPlan {
	plans := make([]lockedPlan, 0, len(strategyRows))
	for _, row := range strategyRows {
		plans = append(plans, lockedPlan{
			PlanID:       row.PlanID,
			StrategyID:   row.StrategyID,
			AssetType:    row.AssetType,
			Symbol:       row.Symbol,
			AssetKey:     row.AssetKey,
			LinkedRiskID: row.LinkedRiskID,
			Parameters:   decodeStrategyParameters(row.Parameters),
		})
	}
	return plans
}

func decodeStrategyParameters(raw datatypes.JSON) map[string]any {
	params := make(map[string]any)
	if len(raw) == 0 {
		return params
	}
	_ = json.Unmarshal(raw, &params)
	return params
}
