import { useState } from 'react';
import { useConverterStore } from '../../stores/converter';

function Section({ title, children, defaultOpen = false }: { title: string; children: React.ReactNode; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className="border-b border-navy-700/50">
      <button
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-navy-300 hover:text-white"
      >
        {title}
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className={`transition-transform ${open ? 'rotate-180' : ''}`}
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      {open && <div className="px-4 pb-4 flex flex-col gap-3">{children}</div>}
    </div>
  );
}

function Slider({ label, value, onChange, min, max, step = 1 }: {
  label: string; value: number; onChange: (v: number) => void; min: number; max: number; step?: number;
}) {
  return (
    <label className="flex flex-col gap-1">
      <div className="flex items-center justify-between text-xs text-navy-400">
        <span>{label}</span>
        <span className="text-navy-300">{value}</span>
      </div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="h-1.5 w-full cursor-pointer appearance-none rounded-full bg-navy-700 accent-accent"
      />
    </label>
  );
}

function Toggle({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center justify-between">
      <span className="text-xs text-navy-400">{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={`relative h-5 w-9 rounded-full transition ${checked ? 'bg-accent' : 'bg-navy-600'}`}
      >
        <span className={`absolute left-0.5 top-0.5 h-4 w-4 rounded-full bg-white transition-transform ${checked ? 'translate-x-4' : ''}`} />
      </button>
    </label>
  );
}

export default function TransformPanel() {
  const options = useConverterStore((s) => s.options);
  const setOptions = useConverterStore((s) => s.setOptions);
  const [maintainRatio, setMaintainRatio] = useState(true);

  return (
    <div className="rounded-xl border border-navy-700/50 bg-navy-800/50">
      <div className="border-b border-navy-700/50 px-4 py-3">
        <h3 className="text-sm font-semibold text-white">Transform Options</h3>
      </div>

      <Section title="Quality" defaultOpen>
        <Slider
          label="Quality"
          value={options.quality ?? 92}
          onChange={(v) => setOptions({ quality: v })}
          min={1}
          max={100}
        />
      </Section>

      <Section title="Resize">
        <div className="flex gap-3">
          <label className="flex flex-1 flex-col gap-1">
            <span className="text-xs text-navy-400">Width</span>
            <input
              type="number"
              placeholder="Auto"
              value={options.width || ''}
              onChange={(e) => {
                const w = e.target.value ? Number(e.target.value) : undefined;
                setOptions({ width: w });
                if (maintainRatio && w) setOptions({ height: undefined });
              }}
              className="rounded-lg border border-navy-600 bg-navy-900 px-2.5 py-1.5 text-sm text-white placeholder-navy-500 outline-none focus:border-accent"
            />
          </label>
          <label className="flex flex-1 flex-col gap-1">
            <span className="text-xs text-navy-400">Height</span>
            <input
              type="number"
              placeholder="Auto"
              value={options.height || ''}
              onChange={(e) => {
                const h = e.target.value ? Number(e.target.value) : undefined;
                setOptions({ height: h });
                if (maintainRatio && h) setOptions({ width: undefined });
              }}
              className="rounded-lg border border-navy-600 bg-navy-900 px-2.5 py-1.5 text-sm text-white placeholder-navy-500 outline-none focus:border-accent"
            />
          </label>
        </div>
        <Toggle label="Maintain aspect ratio" checked={maintainRatio} onChange={setMaintainRatio} />
      </Section>

      <Section title="Filters">
        <Toggle label="Grayscale" checked={options.grayscale ?? false} onChange={(v) => setOptions({ grayscale: v })} />
        <Toggle label="Sharpen" checked={options.sharpen ?? false} onChange={(v) => setOptions({ sharpen: v })} />
        <Toggle label="Invert" checked={options.invert ?? false} onChange={(v) => setOptions({ invert: v })} />
        <Slider label="Blur" value={options.blur ?? 0} onChange={(v) => setOptions({ blur: v })} min={0} max={20} step={0.5} />
        <Slider label="Sepia" value={options.sepia ?? 0} onChange={(v) => setOptions({ sepia: v })} min={0} max={100} />
        <Slider label="Brightness" value={options.brightness ?? 0} onChange={(v) => setOptions({ brightness: v })} min={-100} max={100} />
        <Slider label="Contrast" value={options.contrast ?? 0} onChange={(v) => setOptions({ contrast: v })} min={-100} max={100} />
      </Section>

      <Section title="Watermark">
        <label className="flex flex-col gap-1">
          <span className="text-xs text-navy-400">Text</span>
          <input
            type="text"
            placeholder="Watermark text"
            value={options.watermarkText || ''}
            onChange={(e) => setOptions({ watermarkText: e.target.value || undefined })}
            className="rounded-lg border border-navy-600 bg-navy-900 px-2.5 py-1.5 text-sm text-white placeholder-navy-500 outline-none focus:border-accent"
          />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-xs text-navy-400">Position</span>
          <select
            value={options.watermarkPosition || 'bottom-right'}
            onChange={(e) => setOptions({ watermarkPosition: e.target.value })}
            className="rounded-lg border border-navy-600 bg-navy-900 px-2.5 py-1.5 text-sm text-white outline-none focus:border-accent"
          >
            <option value="top-left">Top Left</option>
            <option value="top-right">Top Right</option>
            <option value="bottom-left">Bottom Left</option>
            <option value="bottom-right">Bottom Right</option>
            <option value="center">Center</option>
          </select>
        </label>
        <Slider
          label="Opacity"
          value={options.watermarkOpacity ?? 50}
          onChange={(v) => setOptions({ watermarkOpacity: v })}
          min={0}
          max={100}
        />
      </Section>
    </div>
  );
}
