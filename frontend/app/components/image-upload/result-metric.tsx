type ResultMetricProps = {
  label: string;
  value: string;
};

export function ResultMetric({ label, value }: ResultMetricProps) {
  return (
    <div className="rounded-xl bg-white px-4 py-3">
      <dt className="text-xs font-black uppercase tracking-normal text-[#6f7f9d]">
        {label}
      </dt>
      <dd className="mt-1 text-lg font-black text-[#081236]">{value}</dd>
    </div>
  );
}
