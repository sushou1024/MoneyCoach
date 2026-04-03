export default function Home() {
  return (
    <main className="page">
      <div className="container">
        <header className="nav">
          <div className="logo">Money Coach</div>
          <nav className="nav-links">
            <a href="#how">Method</a>
            <a href="#signals">Signals</a>
            <a href="#discipline">Discipline</a>
          </nav>
          <a className="button ghost" href="#waitlist">
            Get Early Access
          </a>
        </header>
      </div>

      <div className="ticker" aria-hidden="true">
        <div className="ticker-track">
          <span>BTC + Precision Discipline</span>
          <span>ETH + Risk Surface Mapping</span>
          <span>SOL + Drawdown Guardrails</span>
          <span>USD + Idle Cash Clarity</span>
          <span>SPY + Macro Shock Alerts</span>
          <span>FX + Cross-Market Exposure</span>
          <span>BTC + Precision Discipline</span>
          <span>ETH + Risk Surface Mapping</span>
          <span>SOL + Drawdown Guardrails</span>
          <span>USD + Idle Cash Clarity</span>
          <span>SPY + Macro Shock Alerts</span>
          <span>FX + Cross-Market Exposure</span>
        </div>
      </div>

      <section className="container hero">
        <div className="hero-copy">
          <div className="eyebrow">AI Portfolio Discipline</div>
          <h1 className="hero-title">
            Stop trading blind. <span>See the signal.</span>
          </h1>
          <p className="hero-subtitle">
            Money Coach unifies your crypto, stocks, and FX into one
            diagnostic. Scan holdings, surface the hidden risks, and execute
            strategies built for real portfolios.
          </p>
          <div className="hero-actions">
            <a className="button primary" href="#waitlist">
              Start my analysis
            </a>
            <a className="button ghost" href="#how">
              See the system
            </a>
          </div>
          <div className="hero-meta">
            Institutional-grade clarity for personal portfolios.
          </div>
        </div>

        <div className="hero-visual">
          <div className="diagnostic-stack">
            <div className="diagnostic-card">
              <div className="scan-line" aria-hidden="true" />
              <div className="diagnostic-header">
                <span>Portfolio Health</span>
                <span>Live Snapshot</span>
              </div>
              <div className="score">
                <strong>47</strong>
                <span>Critical Exposure</span>
              </div>
              <div className="risk-list">
                <div className="risk-item">
                  <span>Concentration Risk</span>
                  <span className="badge badge-high">High</span>
                </div>
                <div className="risk-item">
                  <span>Idle Cash Drag</span>
                  <span className="badge badge-medium">Medium</span>
                </div>
                <div className="risk-item">
                  <span>Volatility Shock</span>
                  <span className="badge badge-severe">Severe</span>
                </div>
              </div>
            </div>
            <div className="diagnostic-card">
              <div className="diagnostic-header">
                <span>Strategy Engine</span>
                <span>Adaptive</span>
              </div>
              <div className="risk-list">
                <div className="risk-item">
                  <span>S02 DCA Ladder</span>
                  <span className="badge badge-armed">Armed</span>
                </div>
                <div className="risk-item">
                  <span>S04 Profit Layers</span>
                  <span className="badge badge-ready">Ready</span>
                </div>
                <div className="risk-item">
                  <span>S22 Rebalance</span>
                  <span className="badge badge-queued">Queued</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="section" id="how">
        <div className="container">
          <div className="method-grid">
            <div className="method-copy">
              <div className="section-kicker">Method</div>
              <h2 className="section-title">
                A disciplined loop, not a one-off report.
              </h2>
              <p className="hero-subtitle">
                Money Coach runs as a continuous signal engine. Every scan
                recalibrates your risk, every alert ties back to a concrete
                portfolio action.
              </p>
              <div className="method-metrics">
                <div className="metric">
                  <span className="metric-label">Cadence</span>
                  <span className="metric-value">Adaptive</span>
                </div>
                <div className="metric">
                  <span className="metric-label">Coverage</span>
                  <span className="metric-value">Cross-market</span>
                </div>
                <div className="metric">
                  <span className="metric-label">Signal Latency</span>
                  <span className="metric-value">Minutes</span>
                </div>
              </div>
            </div>

            <div className="method-rail">
              <div className="rail-line" />
              <div className="rail-step">
                <div className="rail-index">01</div>
                <div>
                  <h3>Scan everything</h3>
                  <p>
                    Upload holdings screenshots and merge every venue instantly.
                  </p>
                </div>
              </div>
              <div className="rail-step">
                <div className="rail-index">02</div>
                <div>
                  <h3>Diagnose exposure</h3>
                  <p>
                    Quantify concentration, volatility, and idle capital drag.
                  </p>
                </div>
              </div>
              <div className="rail-step">
                <div className="rail-index">03</div>
                <div>
                  <h3>Execute with precision</h3>
                  <p>
                    Receive plan-level actions aligned with your risk posture.
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className="method-matrix">
            <div className="matrix-item">
              <span className="matrix-title">Unified Ledger</span>
              <span className="matrix-desc">
                Aggregates CEX, wallets, brokerage accounts, and FX cash.
              </span>
            </div>
            <div className="matrix-item">
              <span className="matrix-title">Risk Radar</span>
              <span className="matrix-desc">
                Maps real drawdown exposure with precision scoring.
              </span>
            </div>
            <div className="matrix-item">
              <span className="matrix-title">Strategy Plans</span>
              <span className="matrix-desc">
                Transforms risk into step-by-step action sequences.
              </span>
            </div>
            <div className="matrix-item">
              <span className="matrix-title">Signal Flow</span>
              <span className="matrix-desc">
                Portfolio watch, market alpha, and action alerts tuned to you.
              </span>
            </div>
            <div className="matrix-item">
              <span className="matrix-title">Capital Efficiency</span>
              <span className="matrix-desc">
                Shows how much idle cash is bleeding potential returns.
              </span>
            </div>
            <div className="matrix-item">
              <span className="matrix-title">Adaptive Cadence</span>
              <span className="matrix-desc">
                Rebalances when conditions change, not on a static timer.
              </span>
            </div>
          </div>
        </div>
      </section>

      <section className="section" id="signals">
        <div className="container">
          <div className="section-kicker">Signals</div>
          <h2 className="section-title">Every alert ties to a specific move.</h2>
          <p className="hero-subtitle">
            No generic news feed. Only actionable signals driven by your active
            portfolio and strategy set.
          </p>
          <div className="signal-board">
            <div className="signal-card">
              <h4>Portfolio Watch</h4>
              <p>Detects support breaks, profit targets, and risk escalation.</p>
            </div>
            <div className="signal-card">
              <h4>Market Alpha</h4>
              <p>Highlights oversold and momentum setups in your universe.</p>
            </div>
            <div className="signal-card">
              <h4>Action Alerts</h4>
              <p>Delivers timing cues for DCA, trailing stops, and rebalance.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="section" id="discipline">
        <div className="container split">
          <div>
            <div className="section-kicker">Discipline</div>
            <h2 className="section-title">From chaos to conviction.</h2>
            <p className="hero-subtitle">
              Money Coach is built for investors who want to operate with the
              same rigor as a professional desk, without the noise or the guesswork.
            </p>
          </div>
          <div className="diagnostic-card">
            <div className="diagnostic-header">
              <span>Diagnostic Summary</span>
              <span>Latest</span>
            </div>
            <div className="risk-list">
              <div className="risk-item">
                <span>Net Worth Clarity</span>
                <span className="badge badge-aligned">Aligned</span>
              </div>
              <div className="risk-item">
                <span>Risk/Reward Ratio</span>
                <span className="badge badge-rebalanced">Rebalanced</span>
              </div>
              <div className="risk-item">
                <span>Execution Cadence</span>
                <span className="badge badge-active">Active</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="section" id="waitlist">
        <div className="container">
          <div className="cta">
            <div className="section-kicker">Early Access</div>
            <h2 className="section-title">Build a portfolio you can defend.</h2>
            <p>
              Join Money Coach to receive launch updates and early access to
              portfolio diagnostics and action playbooks.
            </p>
            <div>
              <a className="button primary" href="#waitlist">
                Request access
              </a>
            </div>
          </div>
        </div>
      </section>

      <footer className="container footer">
        <div className="footer-brand">
          <strong>Money Coach</strong>
          <span>© 2026 Money Coach. All rights reserved.</span>
        </div>
        <div className="footer-links">
          <a href="/support">Support</a>
          <a href="/terms">Terms of Use</a>
          <a href="/privacy">Privacy Policy</a>
        </div>
      </footer>
    </main>
  );
}
