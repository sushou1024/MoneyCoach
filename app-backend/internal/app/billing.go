package app

import "strings"

type billingPlan struct {
	PlanID          string
	Name            string
	Interval        string
	Price           float64
	Currency        string
	GoogleProductID string
	GoogleVerifyID  string
	StripePriceID   string
	AppleProductID  string
}

func (s *Server) billingPlans() []billingPlan {
	weeklyProduct := strings.TrimSpace(s.cfg.GooglePlayWeeklyProductID)
	annualProduct := strings.TrimSpace(s.cfg.GooglePlayAnnualProductID)
	weeklyVerify := weeklyProduct
	annualVerify := annualProduct
	return []billingPlan{
		{
			PlanID:          "weekly",
			Name:            "Weekly",
			Interval:        "week",
			Price:           9.99,
			Currency:        "USD",
			GoogleProductID: weeklyProduct,
			GoogleVerifyID:  weeklyVerify,
			StripePriceID:   strings.TrimSpace(s.cfg.StripePriceIDWeekly),
			AppleProductID:  strings.TrimSpace(s.cfg.AppleIapProductIDWeekly),
		},
		{
			PlanID:          "annual",
			Name:            "Annual",
			Interval:        "year",
			Price:           99.9,
			Currency:        "USD",
			GoogleProductID: annualProduct,
			GoogleVerifyID:  annualVerify,
			StripePriceID:   strings.TrimSpace(s.cfg.StripePriceIDYearly),
			AppleProductID:  strings.TrimSpace(s.cfg.AppleIapProductIDYearly),
		},
	}
}

func (s *Server) billingPlanByProductID(productID string) (billingPlan, bool) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return billingPlan{}, false
	}
	for _, plan := range s.billingPlans() {
		if plan.GoogleProductID == productID || plan.GoogleVerifyID == productID || plan.StripePriceID == productID || plan.AppleProductID == productID || plan.PlanID == productID {
			return plan, true
		}
	}
	return billingPlan{}, false
}
