import type { Metadata } from "next";
import { Libre_Baskerville, Spline_Sans_Mono } from "next/font/google";
import "./globals.css";

const display = Libre_Baskerville({
  subsets: ["latin"],
  weight: ["400", "700"],
  variable: "--font-display"
});

const body = Spline_Sans_Mono({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-body"
});

export const metadata: Metadata = {
  title: "Money Coach | AI Portfolio Discipline",
  description:
    "Money Coach is an AI portfolio discipline engine for crypto, stocks, and FX. Scan your holdings, see real risks, and get actionable strategies.",
  metadataBase: new URL("https://moneycoach.cc"),
  icons: {
    icon: "/favicon.svg",
    shortcut: "/favicon.svg"
  },
  openGraph: {
    title: "Money Coach | AI Portfolio Discipline",
    description:
      "Scan your holdings, see real risks, and get actionable strategies with institutional-grade clarity.",
    url: "https://moneycoach.cc",
    siteName: "Money Coach",
    type: "website"
  },
  twitter: {
    card: "summary_large_image",
    title: "Money Coach | AI Portfolio Discipline",
    description:
      "AI portfolio discipline for crypto, stocks, and FX. Real signals, no noise."
  }
};

export default function RootLayout({
  children
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${display.variable} ${body.variable}`}>
      <body>{children}</body>
    </html>
  );
}
