import { useState, useRef, useCallback } from 'react';
import { useConverterStore } from '../../stores/converter';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

const mimeToLabel: Record<string, string> = {
  'image/jpeg': 'JPEG',
  'image/png': 'PNG',
  'image/webp': 'WebP',
  'image/avif': 'AVIF',
  'image/heic': 'HEIC',
  'image/heif': 'HEIF',
  'image/tiff': 'TIFF',
  'image/gif': 'GIF',
  'image/jxl': 'JXL',
  'image/bmp': 'BMP',
  'image/svg+xml': 'SVG',
};

export default function PreviewPane() {
  const files = useConverterStore((s) => s.files);
  const activeFileId = useConverterStore((s) => s.activeFileId);
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

  const entry = files.find((f) => f.id === activeFileId);
  if (!entry) return null;

  const { file, preview, resultUrl, originalSize, resultSize, imageDimensions, resultDimensions } = entry;
  const hasResult = !!resultUrl;
  const ratio = resultSize > 0 ? ((1 - resultSize / originalSize) * 100).toFixed(1) : null;
  const formatType = file?.type ? (mimeToLabel[file.type] || file.type.replace('image/', '').toUpperCase()) : null;

  return (
    <div className="flex flex-col gap-3">
      <div
        ref={containerRef}
        className="relative overflow-hidden rounded-xl border border-navy-700/50 bg-navy-950"
        style={{ minHeight: '300px', maxHeight: '500px' }}
      >
        {hasResult && preview ? (
          <>
            {/* After image (full) */}
            <img src={resultUrl} alt="Converted" className="block max-h-[500px] w-full object-contain" draggable={false} />

            {/* Before image (clipped) */}
            <div
              className="absolute inset-0 overflow-hidden"
              style={{ width: `${sliderPos}%` }}
            >
              <img
                src={preview}
                alt="Original"
                className="block max-h-[500px] w-full object-contain"
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
        ) : hasResult ? (
          <>
            {/* Result only — original format not previewable by browser */}
            <img src={resultUrl} alt="Converted" className="block max-h-[500px] w-full object-contain" draggable={false} />
            <div className="absolute top-3 right-3 rounded-md bg-navy-900/80 px-2 py-1 text-xs text-navy-300 backdrop-blur">
              Converted
            </div>
          </>
        ) : preview ? (
          <img src={preview} alt="Preview" className="block max-h-[500px] w-full object-contain" />
        ) : (
          <div className="flex min-h-[300px] flex-col items-center justify-center gap-3 text-navy-400">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-navy-600">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="M21 15l-5-5L5 21" />
            </svg>
            <p className="text-sm">Preview not available for {formatType || 'this format'}</p>
            <p className="text-xs text-navy-500">Convert to see the result</p>
          </div>
        )}
      </div>

      {/* Metadata bar */}
      <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-1 text-xs text-navy-400">
        {imageDimensions && <span>{imageDimensions.width} x {imageDimensions.height}</span>}
        {formatType && (
          <>
            <span className="text-navy-600">&bull;</span>
            <span>{formatType}</span>
          </>
        )}
        <span className="text-navy-600">&bull;</span>
        <span>{formatBytes(originalSize)}</span>
        {resultSize > 0 && resultDimensions && (
          <>
            <span className="text-navy-600">&rarr;</span>
            <span className="text-accent">{resultDimensions.width} x {resultDimensions.height}</span>
            <span className="text-navy-600">&bull;</span>
            <span className="text-accent">{formatBytes(resultSize)}</span>
            {ratio && (
              <span className={`font-medium ${Number(ratio) > 0 ? 'text-green-400' : 'text-red-400'}`}>
                ({Number(ratio) > 0 ? `-${ratio}%` : `+${Math.abs(Number(ratio))}%`})
              </span>
            )}
          </>
        )}
      </div>
    </div>
  );
}
