export type EntitlementStatus = 'active' | 'grace' | 'expired'

export interface Entitlement {
  status: EntitlementStatus
  provider?: string
  plan_id?: string
  current_period_end?: string
}

export interface UserProfile {
  user_id: string
  email?: string | null
  total_paid_amount: number
  markets: string[]
  experience: string
  style: string
  pain_points: string[]
  risk_preference: string
  risk_level: string
  language: string
  timezone: string
  base_currency: string
  notification_prefs: Record<string, boolean>
  active_portfolio_snapshot_id?: string | null
  entitlement: Entitlement
}

export interface UploadBatchCreateResponse {
  upload_batch_id: string
  status: string
  image_uploads: {
    image_id: string
    upload_url: string
    headers: Record<string, string>
  }[]
  expires_at: string
}

export interface UploadBatchNeedsReview {
  status: 'needs_review'
  upload_batch_id: string
  base_currency?: string
  base_fx_rate_to_usd?: number
  images: {
    image_id: string
    status: string
    error_reason?: string | null
    platform_guess?: string | null
    warnings?: string[]
    is_duplicate?: boolean
    duplicate_of_image_id?: string | null
  }[]
  ocr_assets: OCRAsset[]
  ambiguities: OCRAmbiguity[]
  summary?: { success_images: number; ignored_images: number; unsupported_images: number }
}

export interface UploadBatchComplete {
  status: 'completed'
  portfolio_snapshot_id: string
  calculation_id?: string
  transaction_ids?: string[]
  warnings?: string[]
}

export interface OCRAsset {
  asset_id: string
  image_id: string
  symbol_raw: string
  symbol?: string | null
  exchange_mic?: string | null
  name?: string | null
  logo_url?: string | null
  asset_type?: string | null
  amount: number
  value_from_screenshot?: number | null
  display_currency?: string | null
  value_usd_priced_draft?: number | null
  value_display_draft?: number | null
  manual_value_usd?: number | null
  manual_value_display?: number | null
  confidence?: number | null
  price_as_of?: string | null
  avg_price?: number | null
  avg_price_display?: number | null
  pnl_percent?: number | null
}

export interface OCRAmbiguity {
  symbol_raw: string
  candidates: {
    asset_type: string
    symbol: string
    asset_key: string
    exchange_mic?: string | null
    name?: string | null
  }[]
}

export interface PortfolioSnapshot {
  portfolio_snapshot_id: string
  market_data_snapshot_id: string
  valuation_as_of: string
  snapshot_type: string
  net_worth_usd: number
  holdings: PortfolioHolding[]
  unpriced_holdings: PortfolioHolding[]
  dashboard_metrics?: DashboardMetrics
}

export interface PortfolioHolding {
  asset_type: string
  symbol: string
  asset_key: string
  exchange_mic?: string | null
  name?: string | null
  logo_url?: string | null
  amount: number
  value_usd_priced: number
  current_price: number
  value_quote: number
  quote_currency: string
  valuation_status: string
  pricing_source: string
  balance_type: string
  avg_price?: number | null
  avg_price_quote?: number | null
  avg_price_source?: string | null
  pnl_percent?: number | null
  action_bias?: 'accumulate' | 'wait' | 'hold' | 'reduce' | null
}

export interface DashboardMetrics {
  net_worth_usd: number
  net_worth_display: number
  base_currency: string
  base_fx_rate_to_usd?: number
  health_score: number
  health_status: string
  volatility_score: number
  valuation_as_of: string
  metrics_incomplete: boolean
  score_mode: string
}

export interface MarketRegime {
  as_of: string
  scope: string
  regime: 'risk_on' | 'neutral' | 'risk_off'
  trend_strength: 'strong' | 'medium' | 'weak'
  metrics: {
    alpha_30d: number
    volatility_30d_annualized: number
    max_drawdown_90d: number
    avg_pairwise_corr: number
    cash_pct: number
    top_asset_pct: number
    priced_coverage_pct: number
  }
  trend_breadth: {
    up_count: number
    down_count: number
    neutral_count: number
    weighted_score: number
  }
  drivers: {
    id: string
    kind: string
    tone: 'positive' | 'neutral' | 'caution'
    value_text: string
    value?: number
    up_count?: number
    down_count?: number
    total_count?: number
  }[]
  portfolio_impact: {
    id: string
    kind: string
  }[]
  actions: {
    id: string
    kind: string
  }[]
  leaders: AssetRegimeLeader[]
  laggards: AssetRegimeLeader[]
  featured_assets: FeaturedAssetBrief[]
}

export interface AssetRegimeLeader {
  asset_key: string
  symbol: string
  asset_type: string
  name?: string | null
  logo_url?: string | null
  change_30d: number
  weight_pct: number
  trend_state: string
}

export interface FeaturedAssetBrief {
  asset_key: string
  symbol: string
  asset_type: string
  name?: string | null
  logo_url?: string | null
  action_bias: 'accumulate' | 'wait' | 'hold' | 'reduce'
  summary_signal: string
  weight_pct: number
  beta_to_portfolio: number
  signal_count: number
  latest_signal_severity?: string | null
}

export interface AssetBrief {
  as_of: string
  asset_key: string
  symbol: string
  asset_type: string
  name?: string | null
  logo_url?: string | null
  exchange_mic?: string | null
  quote_currency: string
  current_price: number
  price_change_24h: number
  price_change_7d: number
  price_change_30d: number
  action_bias: 'accumulate' | 'wait' | 'hold' | 'reduce'
  summary_signal: string
  entry_zone: {
    low: number
    high: number
    basis: string
  }
  invalidation: {
    price: number
    reason: string
  }
  technicals: {
    rsi_14: number
    bollinger_upper: number
    bollinger_lower: number
    ma_20: number
    ma_50: number
    ma_200: number
    trend_state: string
    trend_strength: string
  }
  portfolio_fit: {
    is_held: boolean
    weight_pct: number
    beta_to_portfolio: number
    role: string
    concentration_impact: string
    risk_flag: string
  }
  why_now: {
    id: string
    kind: string
  }[]
  related_insights: {
    id: string
    type: string
    severity: string
    trigger_reason: string
    created_at: string
    plan_id?: string | null
    strategy_id?: string | null
  }[]
  related_plans: {
    calculation_id: string
    plan_id: string
    strategy_id: string
    priority: string
    rationale: string
    expected_outcome: string
  }[]
}

export interface PreviewReport {
  meta_data: { calculation_id: string }
  valuation_as_of: string
  market_data_snapshot_id: string
  report_header?: {
    health_score?: { value: number; status: string }
    volatility_dashboard?: { value: number; status: string }
  }
  asset_allocation?: {
    label: string
    weight_pct: number
    value_usd: number
    value_display?: number
    display_currency?: string
  }[]
  identified_risks?: { risk_id: string; type: string; severity: string; teaser_text: string }[]
  locked_projection?: { potential_upside: string; cta: string }
  fixed_metrics?: { net_worth_usd: number; health_score: number; health_status: string; volatility_score: number }
  net_worth_display?: number
  base_currency?: string
  base_fx_rate_to_usd?: number
}

export interface PaidReport extends PreviewReport {
  charts?: { radar_chart?: Record<string, number> }
  risk_insights?: { risk_id: string; type: string; severity: string; message: string }[]
  optimization_plan?: {
    plan_id: string
    strategy_id: string
    asset_type: string
    symbol: string
    asset_key: string
    quote_currency?: string
    linked_risk_id: string
    priority?: string
    parameters: Record<string, any>
    execution_summary?: string
    rationale: string
    expected_outcome: string
  }[]
  daily_alpha_signal?: InsightItem | null
  risk_summary?: string
  exposure_analysis?: { risk_id: string; type: string; severity: string; message: string }[]
  actionable_advice?: {
    plan_id: string
    strategy_id: string
    asset_type: string
    symbol: string
    asset_key: string
    quote_currency?: string
    linked_risk_id: string
    priority?: string
    parameters: Record<string, any>
    execution_summary?: string
    rationale: string
    expected_outcome: string
  }[]
  the_verdict?: { constructive_comment?: string }
}

export interface ReportPlan {
  plan_id: string
  strategy_id: string
  asset_type: string
  symbol: string
  asset_key: string
  quote_currency: string
  linked_risk_id: string
  priority?: string
  parameters: Record<string, any>
  rationale: string
  expected_outcome: string
  chart_series: any[]
}

export interface InsightItem {
  id: string
  type: string
  asset: string
  asset_key?: string
  timeframe?: string
  severity: string
  trigger_reason: string
  trigger_key: string
  strategy_id?: string
  plan_id?: string
  suggested_action?: string
  suggested_quantity?: Record<string, any>
  cta_payload?: Record<string, any>
  created_at: string
  expires_at: string
}

export interface BillingPlan {
  plan_id: string
  name: string
  interval: string
  price: number
  currency: string
  product_ids?: Record<string, string>
}
