import { useState, useRef, useCallback } from 'react';
import { useConverterStore } from '../../stores/converter';

export default function DropZone() {
  const setFile = useConverterStore((s) => s.setFile);
  const [dragOver, setDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const file = e.dataTransfer.files[0];
      if (file?.type.startsWith('image/')) {
        setFile(file);
      }
    },
    [setFile],
  );

  const handleSelect = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) setFile(file);
    },
    [setFile],
  );

  return (
    <div
      onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
      onDragLeave={() => setDragOver(false)}
      onDrop={handleDrop}
      onClick={() => inputRef.current?.click()}
      className={`group flex min-h-[400px] cursor-pointer flex-col items-center justify-center rounded-2xl border-2 border-dashed transition-all duration-300 ${
        dragOver
          ? 'border-accent bg-accent/5 shadow-[0_0_40px_rgba(6,182,212,0.1)]'
          : 'border-navy-600 hover:border-navy-500 hover:bg-navy-800/30'
      }`}
    >
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        onChange={handleSelect}
        className="hidden"
      />

      <div className={`mb-6 rounded-2xl p-5 transition-colors ${dragOver ? 'bg-accent/10' : 'bg-navy-800 group-hover:bg-navy-700'}`}>
        <svg
          width="48"
          height="48"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          className={`transition-colors ${dragOver ? 'text-accent' : 'text-navy-400 group-hover:text-navy-300'}`}
        >
          <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
          <polyline points="17 8 12 3 7 8" />
          <line x1="12" y1="3" x2="12" y2="15" />
        </svg>
      </div>

      <p className={`text-lg font-medium transition-colors ${dragOver ? 'text-accent' : 'text-navy-300'}`}>
        {dragOver ? 'Drop your image here' : 'Drop your image here or click to select'}
      </p>
      <p className="mt-2 text-sm text-navy-500">
        Supports JPEG, PNG, WebP, AVIF, HEIC, TIFF, GIF, JXL
      </p>
    </div>
  );
}
