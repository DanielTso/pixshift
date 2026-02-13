import { useState, useRef, useCallback } from 'react';
import { useConverterStore } from '../../stores/converter';

function collectImageFiles(fileList: FileList): File[] {
  const files: File[] = [];
  for (let i = 0; i < fileList.length; i++) {
    const f = fileList.item(i);
    if (f && f.type.startsWith('image/')) files.push(f);
  }
  return files;
}

export default function DropZone({ compact = false }: { compact?: boolean }) {
  const addFiles = useConverterStore((s) => s.addFiles);
  const [dragOver, setDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const images = collectImageFiles(e.dataTransfer.files);
      if (images.length > 0) addFiles(images);
    },
    [addFiles],
  );

  const handleSelect = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (!e.target.files) return;
      const images = collectImageFiles(e.target.files);
      if (images.length > 0) addFiles(images);
      e.target.value = '';
    },
    [addFiles],
  );

  if (compact) {
    return (
      <div
        onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
        className={`flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 border-dashed px-4 py-3 transition-all ${
          dragOver
            ? 'border-accent bg-accent/5'
            : 'border-navy-700 hover:border-navy-500 hover:bg-navy-800/30'
        }`}
      >
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          multiple
          onChange={handleSelect}
          className="hidden"
        />
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className={dragOver ? 'text-accent' : 'text-navy-400'}
        >
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
        <span className={`text-sm ${dragOver ? 'text-accent' : 'text-navy-400'}`}>
          Drop more images or click to add
        </span>
      </div>
    );
  }

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
        multiple
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
        {dragOver ? 'Drop your images here' : 'Drop your images here or click to select'}
      </p>
      <p className="mt-2 text-sm text-navy-500">
        Supports JPEG, PNG, WebP, AVIF, HEIC, TIFF, GIF, JXL
      </p>
    </div>
  );
}
