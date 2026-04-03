import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Money Coach Support",
  description:
    "Support center for Money Coach app users, including subscription, account, and troubleshooting help."
};

export default function SupportPage() {
  return (
    <main className="legal-page">
      <div className="legal-shell">
        <a className="legal-back" href="/">
          ← Back to Money Coach
        </a>
        <h1>Money Coach Support</h1>
        <p className="legal-meta">Updated: February 20, 2026</p>

        <section>
          <h2>1. Contact Support</h2>
          <p>
            For help with your account, billing, or app usage, contact{" "}
            <a href="mailto:support@moneycoach.cc">support@moneycoach.cc</a>.
          </p>
          <p>
            Please include your account email and a short problem description so
            we can help faster.
          </p>
        </section>

        <section>
          <h2>2. Typical Response Window</h2>
          <ul>
            <li>General support: within 1-2 business days.</li>
            <li>Billing or access issues: prioritized same day when possible.</li>
          </ul>
        </section>

        <section>
          <h2>3. Subscription Help</h2>
          <ul>
            <li>
              iOS subscriptions: open App Store → Account → Subscriptions.
            </li>
            <li>
              Manage renewals and cancellations directly in your App Store
              settings.
            </li>
            <li>
              If you changed devices/accounts, use in-app <strong>Restore Purchases</strong>.
            </li>
          </ul>
        </section>

        <section>
          <h2>4. Account Access</h2>
          <ul>
            <li>
              Existing accounts sign in with email + password.
            </li>
            <li>
              New account creation requires a one-time verification code sent to
              your email.
            </li>
            <li>
              You can also sign in with Apple or Google where available.
            </li>
          </ul>
        </section>

        <section>
          <h2>5. Delete Account</h2>
          <p>
            In the app, go to Settings → Account &amp; Data → Delete Account.
            Deletion is permanent.
          </p>
        </section>

        <section>
          <h2>6. Legal</h2>
          <p>
            Terms of Use: <a href="/terms">moneycoach.cc/terms</a>
          </p>
          <p>
            Privacy Policy: <a href="/privacy">moneycoach.cc/privacy</a>
          </p>
        </section>
      </div>
    </main>
  );
}
