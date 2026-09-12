import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Go Image Optimizer",
  description: "Upload an image and send it through a Go-powered optimization flow.",
  icons: {
    icon: "/images/branding/favicon.ico",
    apple: "/images/branding/logo.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
