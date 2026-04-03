import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Money Coach Privacy Policy",
  description:
    "Privacy Policy for Money Coach, including collection, use, retention, and deletion of account and portfolio data."
};

export default function PrivacyPage() {
  return (
    <main className="legal-page">
      <div className="legal-shell">
        <a className="legal-back" href="/">
          ← Back to Money Coach
        </a>
        <h1>Money Coach Privacy Policy</h1>
        <p className="legal-meta">Effective date: February 20, 2026</p>

        <section>
          <h2>1. Scope</h2>
          <p>
            This Privacy Policy explains how Money Coach collects, uses, shares,
            and retains information when you use our apps and related services.
          </p>
        </section>

        <section>
          <h2>2. Information We Collect</h2>
          <p>We may collect the following categories of information:</p>
          <ul>
            <li>Account data (email, authentication identifiers).</li>
            <li>
              Portfolio input data you submit (for example screenshot uploads,
              symbols, amounts, and preferences).
            </li>
            <li>
              Usage and device data (app version, timestamps, coarse diagnostics
              logs).
            </li>
            <li>
              Billing metadata from app stores or payment providers (without
              storing full payment card numbers).
            </li>
          </ul>
        </section>

        <section>
          <h2>3. How We Use Information</h2>
          <ul>
            <li>To provide portfolio analysis, insights, and alerts.</li>
            <li>To operate authentication, subscriptions, and account security.</li>
            <li>To monitor reliability, prevent abuse, and improve performance.</li>
            <li>To comply with legal obligations and enforce our Terms.</li>
          </ul>
        </section>

        <section>
          <h2>4. Sharing</h2>
          <p>
            We may share limited data with service providers that process data
            on our behalf (for example infrastructure, authentication, email,
            analytics, and payment operations), subject to contractual
            safeguards.
          </p>
          <p>
            We may also disclose information when required by law or to protect
            rights, safety, and security.
          </p>
        </section>

        <section>
          <h2>5. Retention</h2>
          <p>
            We retain data for as long as needed to provide the Service and
            meet legal, accounting, fraud-prevention, and enforcement
            obligations.
          </p>
        </section>

        <section>
          <h2>6. Your Choices</h2>
          <ul>
            <li>You can update profile preferences in-app.</li>
            <li>You can manage subscription renewal in your app store account.</li>
            <li>
              You can delete your account in-app via
              Settings→Account &amp; Data→Delete Account.
            </li>
          </ul>
        </section>

        <section>
          <h2>7. Security</h2>
          <p>
            We use reasonable technical and organizational controls to protect
            data. No system can guarantee absolute security, and you are
            responsible for protecting your own account credentials.
          </p>
        </section>

        <section>
          <h2>8. Children</h2>
          <p>
            The Service is not intended for children under 13 (or the minimum
            age required in your jurisdiction), and we do not knowingly collect
            personal information from children.
          </p>
        </section>

        <section>
          <h2>9. Policy Updates</h2>
          <p>
            We may update this Privacy Policy periodically. The latest version
            is always posted on this page with the effective date.
          </p>
        </section>

        <section>
          <h2>10. Contact</h2>
          <p>
            For privacy requests or questions, contact{" "}
            <a href="mailto:support@moneycoach.cc">support@moneycoach.cc</a>.
          </p>
        </section>
      </div>
    </main>
  );
}
