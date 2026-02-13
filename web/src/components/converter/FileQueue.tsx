import { useConverterStore } from '../../stores/converter';
import type { FileEntry } from '../../stores/converter';

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

function StatusOverlay({ status }: { status: FileEntry['status'] }) {
  if (status === 'converting') {
    return (
      <div className="absolute inset-0 flex items-center justify-center rounded-lg bg-navy-900/60">
        <svg className="h-5 w-5 animate-spin text-accent" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" className="opacity-25" />
          <path d="M4 12a8 8 0 018-8" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
        </svg>
      </div>
    );
  }
  if (status === 'done') {
    return (
      <div className="absolute right-0.5 bottom-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-green-500">
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      </div>
    );
  }
  if (status === 'error') {
    return (
      <div className="absolute right-0.5 bottom-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-red-500">
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round">
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </div>
    );
  }
  return null;
}

function Thumbnail({ entry, isActive }: { entry: FileEntry; isActive: boolean }) {
  const setActiveFile = useConverterStore((s) => s.setActiveFile);
  const removeFile = useConverterStore((s) => s.removeFile);
  const converting = useConverterStore((s) => s.converting);
  const formatLabel = entry.file.type ? (mimeToLabel[entry.file.type] || entry.file.type.replace('image/', '').toUpperCase()) : '?';

  return (
    <div
      onClick={() => setActiveFile(entry.id)}
      className={`group/thumb relative flex-shrink-0 cursor-pointer rounded-lg border-2 transition-all ${
        isActive
          ? 'border-accent shadow-[0_0_12px_rgba(6,182,212,0.2)]'
          : 'border-navy-700 hover:border-navy-500'
      }`}
      style={{ width: '80px' }}
    >
      {/* Remove button */}
      {!converting && (
        <button
          onClick={(e) => { e.stopPropagation(); removeFile(entry.id); }}
          className="absolute -top-1.5 -right-1.5 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-navy-700 text-navy-300 opacity-0 transition-opacity hover:bg-red-500 hover:text-white group-hover/thumb:opacity-100"
        >
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      )}

      {/* Thumbnail image or placeholder */}
      <div className="relative h-16 w-full overflow-hidden rounded-t-md bg-navy-800">
        {entry.preview ? (
          <img src={entry.preview} alt="" className="h-full w-full object-cover" draggable={false} />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <span className="text-[10px] font-bold text-navy-500">{formatLabel}</span>
          </div>
        )}
        <StatusOverlay status={entry.status} />
      </div>

      {/* File info */}
      <div className="px-1 py-0.5">
        <p className="truncate text-[10px] text-navy-400" title={entry.file.name}>
          {entry.file.name}
        </p>
        <p className="text-[9px] text-navy-600">{formatBytes(entry.originalSize)}</p>
      </div>
    </div>
  );
}

export default function FileQueue() {
  const files = useConverterStore((s) => s.files);
  const activeFileId = useConverterStore((s) => s.activeFileId);

  if (files.length <= 1) return null;

  return (
    <div className="flex gap-2 overflow-x-auto pb-1" style={{ scrollbarWidth: 'thin' }}>
      {files.map((entry) => (
        <Thumbnail key={entry.id} entry={entry} isActive={entry.id === activeFileId} />
      ))}
    </div>
  );
}
