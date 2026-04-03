package app

import (
	"fmt"
	"strings"
)

const (
	copyStopLossReason       = "portfolio_watch_stop_loss_reason"
	copyStopLossAction       = "portfolio_watch_stop_loss_action"
	copyTakeProfitReason     = "portfolio_watch_take_profit_reason"
	copyTakeProfitAction     = "portfolio_watch_take_profit_action"
	copySafetyOrderReason    = "action_alert_safety_reason"
	copySafetyOrderAction    = "action_alert_safety_action"
	copyTrailingReason       = "action_alert_trailing_reason"
	copyTrailingAction       = "action_alert_trailing_action"
	copyDCAReason            = "action_alert_dca_reason"
	copyDCAAction            = "action_alert_dca_action"
	copyAdditionReason       = "action_alert_addition_reason"
	copyAdditionAction       = "action_alert_addition_action"
	copyFundingReason        = "action_alert_funding_reason"
	copyFundingAction        = "action_alert_funding_action"
	copyTrendReason          = "action_alert_trend_reason"
	copyTrendActionHold      = "action_alert_trend_action_hold"
	copyTrendActionHoldOrAdd = "action_alert_trend_action_hold_or_add"
	copyTrendActionReduce    = "action_alert_trend_action_reduce"
	copyRebalanceReason      = "action_alert_rebalance_reason"
	copyRebalanceAction      = "action_alert_rebalance_action"
	copyMarketAlphaReason    = "market_alpha_reason"
	copyMarketAlphaAction    = "market_alpha_action"
)

var insightCopyTemplates = map[string]map[string]string{
	"English": {
		copyStopLossReason:       "Price touched stop-loss %s.",
		copyStopLossAction:       "Consider executing the stop-loss to limit drawdown.",
		copyTakeProfitReason:     "Price reached take-profit %s.",
		copyTakeProfitAction:     "Consider selling %s%% for %s.",
		copySafetyOrderReason:    "Price reached safety order %s.",
		copySafetyOrderAction:    "Suggested buy: %s.",
		copyTrailingReason:       "Drawdown %s%% from peak %s reached.",
		copyTrailingAction:       "Consider trimming to lock in gains.",
		copyDCAReason:            "DCA scheduled at %s.",
		copyDCAAction:            "Suggested buy: %s.",
		copyAdditionReason:       "PnL hit %s%% target.",
		copyAdditionAction:       "Suggested add: %s.",
		copyFundingReason:        "Funding rate %s%% with basis %s%%.",
		copyFundingAction:        "Consider hedging to capture funding.",
		copyTrendReason:          "Trend shifted to %s.",
		copyTrendActionHold:      "Trend stable: hold and monitor.",
		copyTrendActionHoldOrAdd: "Uptrend: consider adding on strength.",
		copyTrendActionReduce:    "Downtrend: consider reducing exposure.",
		copyRebalanceReason:      "Portfolio mix drifted %s%% from target.",
		copyRebalanceAction:      "Rebalance to return to target weights.",
		copyMarketAlphaReason:    "RSI %s on %s with close below lower band.",
		copyMarketAlphaAction:    "Watch for a rebound confirmation.",
	},
	"Simplified Chinese": {
		copyStopLossReason:       "价格触及止损价 %s。",
		copyStopLossAction:       "考虑执行止损以控制回撤。",
		copyTakeProfitReason:     "价格触及止盈价 %s。",
		copyTakeProfitAction:     "建议卖出约%s%%（%s）。",
		copySafetyOrderReason:    "价格触及补仓价 %s。",
		copySafetyOrderAction:    "建议买入：%s。",
		copyTrailingReason:       "回撤%s%%，高点%s。",
		copyTrailingAction:       "考虑减仓锁定利润。",
		copyDCAReason:            "定投时间：%s。",
		copyDCAAction:            "建议买入：%s。",
		copyAdditionReason:       "收益率达到%s%%触发加仓。",
		copyAdditionAction:       "建议加仓：%s。",
		copyFundingReason:        "资金费率%s%%，基差%s%%。",
		copyFundingAction:        "考虑对冲获取资金费率收益。",
		copyTrendReason:          "趋势切换为%s。",
		copyTrendActionHold:      "趋势稳定，继续持有观望。",
		copyTrendActionHoldOrAdd: "趋势向上，可考虑加仓。",
		copyTrendActionReduce:    "趋势转弱，考虑降低仓位。",
		copyRebalanceReason:      "组合配置偏离目标约%s%%。",
		copyRebalanceAction:      "建议再平衡以回到目标权重。",
		copyMarketAlphaReason:    "RSI %s（%s周期），且收盘价跌破布林下轨。",
		copyMarketAlphaAction:    "关注反弹确认信号。",
	},
	"Traditional Chinese": {
		copyStopLossReason:       "價格觸及止損價 %s。",
		copyStopLossAction:       "考慮執行止損以控制回撤。",
		copyTakeProfitReason:     "價格觸及止盈價 %s。",
		copyTakeProfitAction:     "建議賣出約%s%%（%s）。",
		copySafetyOrderReason:    "價格觸及補倉價 %s。",
		copySafetyOrderAction:    "建議買入：%s。",
		copyTrailingReason:       "回撤%s%%，高點%s。",
		copyTrailingAction:       "考慮減倉鎖定利潤。",
		copyDCAReason:            "定投時間：%s。",
		copyDCAAction:            "建議買入：%s。",
		copyAdditionReason:       "收益率達到%s%%觸發加倉。",
		copyAdditionAction:       "建議加倉：%s。",
		copyFundingReason:        "資金費率%s%%，基差%s%%。",
		copyFundingAction:        "考慮對沖獲取資金費率收益。",
		copyTrendReason:          "趨勢切換為%s。",
		copyTrendActionHold:      "趨勢穩定，持有觀望。",
		copyTrendActionHoldOrAdd: "趨勢向上，可考慮加倉。",
		copyTrendActionReduce:    "趨勢轉弱，考慮降低倉位。",
		copyRebalanceReason:      "組合配置偏離目標約%s%%。",
		copyRebalanceAction:      "建議再平衡以回到目標權重。",
		copyMarketAlphaReason:    "RSI %s（%s週期），且收盤價跌破布林下軌。",
		copyMarketAlphaAction:    "關注反彈確認訊號。",
	},
	"Japanese": {
		copyStopLossReason:       "価格が損切りライン %s に到達しました。",
		copyStopLossAction:       "損失を抑えるため損切りを検討してください。",
		copyTakeProfitReason:     "価格が利確目標 %s に到達しました。",
		copyTakeProfitAction:     "約%s%%を売却（%s）。",
		copySafetyOrderReason:    "価格がナンピン価格 %s に到達しました。",
		copySafetyOrderAction:    "推奨買い：%s。",
		copyTrailingReason:       "下落%s%%（高値%s）。",
		copyTrailingAction:       "利益確定の検討を。",
		copyDCAReason:            "積立予定時刻：%s。",
		copyDCAAction:            "推奨買い：%s。",
		copyAdditionReason:       "損益が%s%%に到達。",
		copyAdditionAction:       "推奨追加：%s。",
		copyFundingReason:        "資金調達率 %s%%、ベーシス %s%%。",
		copyFundingAction:        "資金調達収益のヘッジを検討。",
		copyTrendReason:          "トレンドが%sに転換。",
		copyTrendActionHold:      "トレンド維持：保有継続。",
		copyTrendActionHoldOrAdd: "上昇トレンド：追加を検討。",
		copyTrendActionReduce:    "下落トレンド：縮小を検討。",
		copyRebalanceReason:      "ポートフォリオ配分が目標から約%s%%ずれています。",
		copyRebalanceAction:      "目標比率に戻すためリバランスを検討してください。",
		copyMarketAlphaReason:    "RSI %s（%s足）、下限バンド割れ。",
		copyMarketAlphaAction:    "反発確認を待ちましょう。",
	},
	"Korean": {
		copyStopLossReason:       "가격이 손절가 %s에 도달했습니다.",
		copyStopLossAction:       "손실을 줄이기 위해 손절을 고려하세요.",
		copyTakeProfitReason:     "가격이 목표가 %s에 도달했습니다.",
		copyTakeProfitAction:     "약 %s%% 매도 (%s).",
		copySafetyOrderReason:    "가격이 추가매수 가격 %s에 도달했습니다.",
		copySafetyOrderAction:    "추천 매수: %s.",
		copyTrailingReason:       "하락%s%% (고점 %s).",
		copyTrailingAction:       "이익 확정을 고려하세요.",
		copyDCAReason:            "정기매수 예정: %s.",
		copyDCAAction:            "추천 매수: %s.",
		copyAdditionReason:       "수익률이 %s%%에 도달했습니다.",
		copyAdditionAction:       "추천 추가매수: %s.",
		copyFundingReason:        "펀딩비 %s%%, 베이시스 %s%%.",
		copyFundingAction:        "펀딩 수익 헤지를 고려하세요.",
		copyTrendReason:          "추세가 %s로 전환되었습니다.",
		copyTrendActionHold:      "추세 유지: 보유 유지.",
		copyTrendActionHoldOrAdd: "상승 추세: 추가 매수를 고려하세요.",
		copyTrendActionReduce:    "하락 추세: 비중 축소를 고려하세요.",
		copyRebalanceReason:      "포트폴리오 비중이 목표 대비 약 %s%% 벗어났습니다.",
		copyRebalanceAction:      "목표 비중으로 되돌리기 위해 리밸런싱을 고려하세요.",
		copyMarketAlphaReason:    "RSI %s(%s 기준), 하단 밴드 이탈.",
		copyMarketAlphaAction:    "반등 확인을 기다리세요.",
	},
}

var trendStateLabels = map[string]map[string]string{
	"English": {
		"strong_up":   "Strong Uptrend",
		"up":          "Uptrend",
		"down":        "Downtrend",
		"strong_down": "Strong Downtrend",
		"neutral":     "Neutral",
	},
	"Simplified Chinese": {
		"strong_up":   "强势上升",
		"up":          "上升",
		"down":        "下降",
		"strong_down": "强势下降",
		"neutral":     "中性",
	},
	"Traditional Chinese": {
		"strong_up":   "強勢上升",
		"up":          "上升",
		"down":        "下降",
		"strong_down": "強勢下降",
		"neutral":     "中性",
	},
	"Japanese": {
		"strong_up":   "強い上昇トレンド",
		"up":          "上昇トレンド",
		"down":        "下降トレンド",
		"strong_down": "強い下降トレンド",
		"neutral":     "中立",
	},
	"Korean": {
		"strong_up":   "강한 상승 추세",
		"up":          "상승 추세",
		"down":        "하락 추세",
		"strong_down": "강한 하락 추세",
		"neutral":     "중립",
	},
}

func insightCopy(language, key string, args ...any) string {
	language = strings.TrimSpace(language)
	if language == "" {
		language = "English"
	}
	templates := insightCopyTemplates[language]
	if templates == nil {
		templates = insightCopyTemplates["English"]
	}
	template := templates[key]
	if template == "" {
		template = insightCopyTemplates["English"][key]
	}
	if template == "" {
		return ""
	}
	return fmt.Sprintf(template, args...)
}

func localizedTrendState(language, state string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		language = "English"
	}
	labels := trendStateLabels[language]
	if labels == nil {
		labels = trendStateLabels["English"]
	}
	label := labels[state]
	if label == "" {
		label = trendStateLabels["English"][state]
	}
	if label == "" {
		return state
	}
	return label
}

func trendActionCopyKey(params map[string]any) string {
	action := strings.ToLower(strings.TrimSpace(getStringParam(params, "trend_action")))
	switch action {
	case "hold_or_add":
		return copyTrendActionHoldOrAdd
	case "reduce_exposure":
		return copyTrendActionReduce
	default:
		return copyTrendActionHold
	}
}
