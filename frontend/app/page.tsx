import { DecorativeBackground } from "./components/home/decorative-background";
import { FeatureList } from "./components/home/feature-list";
import { Footer } from "./components/home/footer";
import { HeroCopy } from "./components/home/hero-copy";
import { ImageComparisonPreview } from "./components/home/image-comparison-preview";

export default function Home() {
  return (
    <main className="relative min-h-screen overflow-x-hidden bg-[#f6fbff] text-[#081236]">
      <DecorativeBackground />

      <section className="relative z-10 mx-auto flex min-h-screen w-full max-w-[1392px] flex-col px-5 py-8 sm:px-8 lg:px-12 lg:pb-0">
        <div className="grid flex-1 items-center gap-10 pt-8 md:pt-12 lg:grid-cols-[0.92fr_1.08fr] lg:gap-5 lg:pt-6">
          <HeroCopy />
          <ImageComparisonPreview />
        </div>

        <FeatureList />
        <Footer />
      </section>
    </main>
  );
}
