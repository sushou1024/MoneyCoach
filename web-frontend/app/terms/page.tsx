import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Money Coach Terms of Use",
  description:
    "Terms of Use for Money Coach, including subscription billing, account deletion, and acceptable use."
};

export default function TermsPage() {
  return (
    <main className="legal-page">
      <div className="legal-shell">
        <a className="legal-back" href="/">
          ← Back to Money Coach
        </a>
        <h1>Money Coach Terms of Use</h1>
        <p className="legal-meta">Effective date: February 20, 2026</p>

        <section>
          <h2>1. Agreement</h2>
          <p>
            These Terms of Use govern your access to and use of Money Coach
            applications, websites, and related services (the &quot;Service&quot;). By
            creating an account or using the Service, you agree to these Terms.
          </p>
        </section>

        <section>
          <h2>2. Eligibility and Account</h2>
          <p>
            You are responsible for maintaining the confidentiality of your
            account credentials and for all activity under your account. You
            must provide accurate information and keep it updated.
          </p>
        </section>

        <section>
          <h2>3. Subscription and Billing</h2>
          <p>
            Money Coach offers optional auto-renewable subscriptions. Payment
            is charged to your Apple ID or Google Play account at purchase
            confirmation.
          </p>
          <p>
            Subscriptions renew automatically unless canceled at least 24 hours
            before the end of the current billing period. Your account is
            charged for renewal within 24 hours before period end.
          </p>
          <p>
            You can manage or cancel your subscription in your App Store or
            Google Play account settings after purchase.
          </p>
        </section>

        <section>
          <h2>4. Financial Disclaimer</h2>
          <p>
            The Service provides educational analytics and information only. It
            does not provide investment, legal, tax, or accounting advice and
            does not recommend buying or selling any financial instrument.
          </p>
        </section>

        <section>
          <h2>5. Acceptable Use</h2>
          <p>
            You agree not to misuse the Service, attempt unauthorized access,
            reverse engineer protected components, or use the Service in a way
            that could harm users, systems, or legal rights.
          </p>
        </section>

        <section>
          <h2>6. Account Deletion</h2>
          <p>
            You may delete your account in-app at any time via
            Settings→Account &amp; Data→Delete Account. Deletion is permanent and
            removes your associated data from active systems, subject to lawful
            retention obligations.
          </p>
        </section>

        <section>
          <h2>7. Limitation of Liability</h2>
          <p>
            To the maximum extent permitted by law, Money Coach is not liable
            for indirect, incidental, special, consequential, or punitive
            damages, or any loss of profits, revenues, data, or goodwill.
          </p>
        </section>

        <section>
          <h2>8. Changes to Terms</h2>
          <p>
            We may update these Terms from time to time. The updated version is
            effective when posted on this page with a new effective date.
          </p>
        </section>

        <section>
          <h2>9. Contact</h2>
          <p>
            For questions about these Terms, contact{" "}
            <a href="mailto:support@moneycoach.cc">support@moneycoach.cc</a>.
          </p>
        </section>
      </div>
    </main>
  );
}
