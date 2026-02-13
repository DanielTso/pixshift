import { useState, useRef, useCallback } from 'react';
import { useConverterStore } from '../../stores/converter';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function PreviewPane() {
  const preview = useConverterStore((s) => s.preview);
  const resultUrl = useConverterStore((s) => s.resultUrl);
  const originalSize = useConverterStore((s) => s.originalSize);
  const resultSize = useConverterStore((s) => s.resultSize);
  const [sliderPos, setSliderPos] = useState(50);
  const containerRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);

  const handleMove = useCallback((clientX: number) => {
    if (!containerRef.current || !dragging.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    setSliderPos((x / rect.width) * 100);
  }, []);

  const handleMouseDown = useCallback(() => {
    dragging.current = true;
    const handleMouseMove = (e: MouseEvent) => handleMove(e.clientX);
    const handleMouseUp = () => {
      dragging.current = false;
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }, [handleMove]);

  if (!preview) return null;

  const hasResult = !!resultUrl;
  const ratio = resultSize > 0 ? ((1 - resultSize / originalSize) * 100).toFixed(1) : null;

  return (
    <div className="flex flex-col gap-3">
      <div
        ref={containerRef}
        className="relative overflow-hidden rounded-xl border border-navy-700/50 bg-navy-950"
        style={{ minHeight: '300px' }}
      >
        {hasResult ? (
          <>
            {/* After image (full) */}
            <img src={resultUrl} alt="Converted" className="block w-full" draggable={false} />

            {/* Before image (clipped) */}
            <div
              className="absolute inset-0 overflow-hidden"
              style={{ width: `${sliderPos}%` }}
            >
              <img
                src={preview}
                alt="Original"
                className="block w-full"
                style={{ width: containerRef.current ? `${containerRef.current.offsetWidth}px` : '100%' }}
                draggable={false}
              />
            </div>

            {/* Slider handle */}
            <div
              className="absolute top-0 bottom-0 cursor-col-resize"
              style={{ left: `${sliderPos}%`, transform: 'translateX(-50%)' }}
              onMouseDown={handleMouseDown}
            >
              <div className="h-full w-0.5 bg-white/80" />
              <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 flex h-8 w-8 items-center justify-center rounded-full bg-white shadow-lg">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#0f172a" strokeWidth="2">
                  <polyline points="15 18 9 12 15 6" />
                  <polyline points="9 18 15 12 9 6" transform="translate(6 0)" />
                </svg>
              </div>
            </div>

            {/* Labels */}
            <div className="absolute top-3 left-3 rounded-md bg-navy-900/80 px-2 py-1 text-xs text-navy-300 backdrop-blur">
              Original
            </div>
            <div className="absolute top-3 right-3 rounded-md bg-navy-900/80 px-2 py-1 text-xs text-navy-300 backdrop-blur">
              Converted
            </div>
          </>
        ) : (
          <img src={preview} alt="Preview" className="block w-full" />
        )}
      </div>

      {/* Size info */}
      <div className="flex items-center justify-center gap-6 text-sm">
        <span className="text-navy-400">
          Original: <span className="text-navy-200">{formatBytes(originalSize)}</span>
        </span>
        {resultSize > 0 && (
          <>
            <span className="text-navy-400">
              Converted: <span className="text-navy-200">{formatBytes(resultSize)}</span>
            </span>
            {ratio && (
              <span className={`font-medium ${Number(ratio) > 0 ? 'text-green-400' : 'text-red-400'}`}>
                {Number(ratio) > 0 ? `-${ratio}%` : `+${Math.abs(Number(ratio))}%`}
              </span>
            )}
          </>
        )}
      </div>
    </div>
  );
}
