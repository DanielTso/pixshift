import { useConverterStore } from '../../stores/converter';

const formats = [
  { id: 'jpeg', label: 'JPEG', hint: 'Universal' },
  { id: 'png', label: 'PNG', hint: 'Lossless' },
  { id: 'webp', label: 'WebP', hint: 'Smallest' },
  { id: 'avif', label: 'AVIF', hint: 'Modern' },
  { id: 'heic', label: 'HEIC', hint: 'Apple' },
  { id: 'tiff', label: 'TIFF', hint: 'Print' },
  { id: 'gif', label: 'GIF', hint: 'Animated' },
  { id: 'jxl', label: 'JXL', hint: 'Next-gen' },
];

export default function FormatPicker() {
  const format = useConverterStore((s) => s.format);
  const setFormat = useConverterStore((s) => s.setFormat);

  return (
    <div className="flex flex-wrap justify-center gap-2">
      {formats.map((f) => (
        <button
          key={f.id}
          onClick={() => setFormat(f.id)}
          className={`flex flex-col items-center rounded-xl px-4 py-2.5 text-sm font-medium transition ${
            format === f.id
              ? 'bg-accent/15 text-accent ring-1 ring-accent/40'
              : 'bg-navy-800 text-navy-300 hover:bg-navy-700 hover:text-white'
          }`}
        >
          <span>{f.label}</span>
          <span className={`mt-0.5 text-[10px] ${format === f.id ? 'text-accent/70' : 'text-navy-500'}`}>
            {f.hint}
          </span>
        </button>
      ))}
    </div>
  );
}
