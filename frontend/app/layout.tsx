import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Go Image Optimizer",
  description: "Initial application shell for Go Image Optimizer.",
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
