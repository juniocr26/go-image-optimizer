import Image from "next/image";

export function ImageComparisonPreview() {
  return (
    <div className="relative mx-auto h-[510px] w-full max-w-[650px] sm:h-[610px] lg:h-[700px] lg:max-w-none">
      <div
        aria-hidden="true"
        className="absolute left-[3%] top-[16%] h-[330px] w-[210px] rotate-[-10deg] overflow-hidden rounded-[28px] border-[6px] border-white bg-white opacity-45 shadow-[0_24px_70px_rgba(15,42,80,0.25)] blur-[3px] sm:h-[410px] sm:w-[260px] lg:left-[11%] lg:top-[20%]"
      >
        <Image
          alt=""
          className="h-full w-full object-cover saturate-50"
          fill
          priority
          sizes="(min-width: 1024px) 260px, 210px"
          src="/images/hero/mountain.png"
        />
      </div>

      <div
        aria-hidden="true"
        className="absolute right-[-4%] top-[23%] h-[360px] w-[230px] rotate-[10deg] overflow-hidden rounded-[28px] border-[6px] border-white bg-white opacity-45 shadow-[0_24px_70px_rgba(15,42,80,0.25)] blur-[4px] sm:h-[450px] sm:w-[280px] lg:right-[1%] lg:top-[27%]"
      >
        <Image
          alt=""
          className="h-full w-full object-cover saturate-50"
          fill
          priority
          sizes="(min-width: 1024px) 280px, 230px"
          src="/images/hero/mountain.png"
        />
      </div>

      <div className="absolute left-1/2 top-[8%] h-[410px] w-[300px] -translate-x-1/2 rotate-[12deg] rounded-[30px] bg-white p-[7px] shadow-[0_28px_78px_rgba(15,42,80,0.3)] sm:h-[535px] sm:w-[392px] lg:left-[53%] lg:top-[12%]">
        <div className="relative h-full overflow-hidden rounded-[24px]">
          <Image
            alt="Mountain lake image preview"
            className="h-full w-full object-cover"
            fill
            priority
            sizes="(min-width: 1024px) 392px, (min-width: 640px) 392px, 300px"
            src="/images/hero/mountain.png"
          />

          <div className="absolute inset-y-0 left-0 w-1/2 overflow-hidden">
            <div className="relative h-full w-[200%]">
              <Image
                alt=""
                className="h-full w-full object-cover brightness-[0.78] contrast-[0.86] saturate-[0.42]"
                fill
                sizes="(min-width: 1024px) 392px, (min-width: 640px) 392px, 300px"
                src="/images/hero/mountain.png"
              />
            </div>
            <div className="absolute inset-0 bg-[#243850]/20" />
          </div>

          <div className="absolute inset-y-0 left-1/2 w-[2px] -translate-x-1/2 bg-white/95 shadow-[0_0_16px_rgba(255,255,255,0.85)]" />
          <div className="absolute left-8 top-5 rounded-xl bg-[#203047]/80 px-5 py-3 text-sm font-bold text-white shadow-lg sm:left-9 sm:text-base">
            Before
          </div>
          <div className="absolute right-5 top-12 rotate-[-2deg] rounded-xl bg-white px-5 py-3 text-sm font-black text-[#00a87d] shadow-lg sm:right-6 sm:text-base">
            Optimized
          </div>
          <div className="absolute left-1/2 top-[56%] grid h-12 w-12 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full bg-white text-[#14234a] shadow-[0_12px_28px_rgba(8,18,54,0.22)]">
            <CompareIcon />
          </div>
        </div>
      </div>

      <div className="handwritten pointer-events-none absolute right-[3%] top-0 hidden rotate-[-6deg] text-[25px] font-bold leading-[1.05] text-[#6f7f9d] sm:block lg:right-[7%] lg:top-[2%] lg:text-[29px]">
        <p>Compression keeps dimensions</p>
        <p>Smaller when possible</p>
        <svg
          aria-hidden="true"
          className="absolute -right-12 top-11 h-24 w-20 rotate-[16deg] overflow-visible"
          fill="none"
          viewBox="0 0 76 96"
        >
          <path
            d="M7 5c35 1 54 18 52 49-.6 12-6 24-17 36"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="3"
          />
          <path
            d="M32 76l9 15 17-8"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="3"
          />
        </svg>
      </div>

      <div className="handwritten pointer-events-none absolute bottom-[5%] left-[10%] hidden rotate-[-9deg] text-[24px] font-bold leading-[1.05] text-[#6f7f9d] sm:block lg:bottom-[10%] lg:left-[18%] lg:text-[27px]">
        <svg
          aria-hidden="true"
          className="absolute -left-12 -top-10 h-20 w-20 -rotate-[14deg] overflow-visible"
          fill="none"
          viewBox="0 0 76 80"
        >
          <path
            d="M61 70C31 64 13 47 17 24c2-11 9-18 19-22"
            stroke="currentColor"
            strokeLinecap="round"
            strokeWidth="3"
          />
          <path
            d="M32 2l7-1 3 16"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="3"
          />
        </svg>
        <p>Measured sizes</p>
        <p>Clear result</p>
      </div>
    </div>
  );
}

function CompareIcon() {
  return (
    <svg aria-hidden="true" className="h-6 w-6" fill="none" viewBox="0 0 24 24">
      <path
        d="M8 7 4 11l4 4M4 11h7M16 17l4-4-4-4M20 13h-7"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.4"
      />
    </svg>
  );
}
