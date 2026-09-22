import Image from "next/image";

export function DecorativeBackground() {
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 overflow-hidden"
    >
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_12%,rgba(255,255,255,0.98),rgba(255,255,255,0)_36%),radial-gradient(circle_at_75%_5%,rgba(227,242,255,0.92),rgba(255,255,255,0)_35%),linear-gradient(180deg,#f7fcff_0%,#f8fcff_58%,#eef8f7_100%)]" />
      <svg
        className="absolute left-[-6%] top-[35%] h-[260px] w-[112%] text-[#9fe4d9]/45 sm:top-[40%]"
        fill="none"
        preserveAspectRatio="none"
        viewBox="0 0 1500 260"
      >
        <path
          d="M0 88C153 20 220 39 336 135c107 88 218 67 340 3 136-71 248-47 383 38 151 95 290 83 441-20"
          stroke="currentColor"
          strokeLinecap="round"
          strokeWidth="34"
        />
      </svg>
      <div className="hero-grass absolute bottom-0 left-[-7%] h-[43%] min-h-[280px] w-[78%] max-w-[820px] opacity-75 sm:h-[48%] lg:h-[45%]">
        <Image
          alt=""
          className="h-full w-full object-cover object-left-bottom"
          fill
          priority
          sizes="(min-width: 1024px) 820px, 78vw"
          src="/images/hero/grassfield.png"
        />
      </div>
      <div className="hero-grass absolute bottom-[-9%] right-[-24%] h-[35%] min-h-[220px] w-[58%] rotate-[-6deg] opacity-35 blur-[2px] sm:right-[-18%] lg:right-[-15%]">
        <Image
          alt=""
          className="h-full w-full object-cover object-left-bottom"
          fill
          sizes="58vw"
          src="/images/hero/grassfield.png"
        />
      </div>
    </div>
  );
}
