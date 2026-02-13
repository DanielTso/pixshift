import { create } from 'zustand';
import { convertImage } from '../api/convert';
import type { ConvertOptions } from '../api/convert';

interface Dimensions {
  width: number;
  height: number;
}

export interface FileEntry {
  id: string;
  file: File;
  preview: string | null;
  result: Blob | null;
  resultUrl: string | null;
  status: 'pending' | 'converting' | 'done' | 'error';
  error: string | null;
  originalSize: number;
  resultSize: number;
  imageDimensions: Dimensions | null;
  resultDimensions: Dimensions | null;
}

interface ConverterState {
  files: FileEntry[];
  activeFileId: string | null;
  format: string;
  options: Omit<ConvertOptions, 'format'>;
  converting: boolean;
  convertProgress: { current: number; total: number } | null;
  addFiles: (files: File[]) => void;
  removeFile: (id: string) => void;
  setActiveFile: (id: string) => void;
  setFormat: (format: string) => void;
  setOptions: (options: Partial<Omit<ConvertOptions, 'format'>>) => void;
  convert: () => Promise<void>;
  reset: () => void;
}

const defaultOptions: Omit<ConvertOptions, 'format'> = {
  quality: 92,
};

function generatePreview(file: File, onResult: (url: string | null, dims: Dimensions | null) => void) {
  const blobUrl = URL.createObjectURL(file);
  const img = new Image();
  img.onload = () => {
    onResult(blobUrl, { width: img.naturalWidth, height: img.naturalHeight });
  };
  img.onerror = () => {
    URL.revokeObjectURL(blobUrl);
    onResult(null, null);
  };
  img.src = blobUrl;
}

function revokeEntry(entry: FileEntry) {
  if (entry.preview) URL.revokeObjectURL(entry.preview);
  if (entry.resultUrl) URL.revokeObjectURL(entry.resultUrl);
}

export const useConverterStore = create<ConverterState>((set, get) => ({
  files: [],
  activeFileId: null,
  format: 'webp',
  options: { ...defaultOptions },
  converting: false,
  convertProgress: null,

  addFiles: (newFiles) => {
    const entries: FileEntry[] = newFiles.map((file) => ({
      id: crypto.randomUUID(),
      file,
      preview: null,
      result: null,
      resultUrl: null,
      status: 'pending' as const,
      error: null,
      originalSize: file.size,
      resultSize: 0,
      imageDimensions: null,
      resultDimensions: null,
    }));

    const isFirst = get().files.length === 0;
    const first = entries[0] as FileEntry | undefined;
    const firstId = first?.id ?? null;
    set((state) => ({
      files: [...state.files, ...entries],
      activeFileId: isFirst && firstId ? firstId : state.activeFileId,
    }));

    // Generate previews asynchronously
    for (const entry of entries) {
      generatePreview(entry.file, (url, dims) => {
        set((state) => ({
          files: state.files.map((f) =>
            f.id === entry.id ? { ...f, preview: url, imageDimensions: dims } : f
          ),
        }));
      });
    }
  },

  removeFile: (id) => {
    const state = get();
    const entry = state.files.find((f) => f.id === id);
    if (entry) revokeEntry(entry);

    const remaining = state.files.filter((f) => f.id !== id);
    let nextActive = state.activeFileId;
    if (state.activeFileId === id) {
      const next = remaining[0] as FileEntry | undefined;
      nextActive = next?.id ?? null;
    }
    set({ files: remaining, activeFileId: nextActive });
  },

  setActiveFile: (id) => set({ activeFileId: id }),

  setFormat: (format) => set({ format }),

  setOptions: (opts) =>
    set((state) => ({ options: { ...state.options, ...opts } })),

  convert: async () => {
    const { files, format, options } = get();
    const pending = files.filter((f) => f.status === 'pending' || f.status === 'error');
    if (pending.length === 0) return;

    set({ converting: true, convertProgress: { current: 0, total: pending.length } });

    for (const [i, entry] of pending.entries()) {
      const entryId = entry.id;

      // Mark as converting
      set((state) => ({
        files: state.files.map((f) =>
          f.id === entryId ? { ...f, status: 'converting' as const, error: null } : f
        ),
        convertProgress: { current: i + 1, total: pending.length },
      }));

      try {
        const blob = await convertImage(entry.file, { format, ...options });
        const resultUrl = URL.createObjectURL(blob);

        // Get result dimensions asynchronously
        const img = new Image();
        img.onload = () => {
          set((state) => ({
            files: state.files.map((f) =>
              f.id === entryId ? { ...f, resultDimensions: { width: img.naturalWidth, height: img.naturalHeight } } : f
            ),
          }));
        };
        img.src = resultUrl;

        set((state) => ({
          files: state.files.map((f) =>
            f.id === entryId
              ? { ...f, status: 'done' as const, result: blob, resultUrl, resultSize: blob.size, resultDimensions: null }
              : f
          ),
        }));
      } catch (err) {
        set((state) => ({
          files: state.files.map((f) =>
            f.id === entryId
              ? { ...f, status: 'error' as const, error: (err as Error).message }
              : f
          ),
        }));
      }
    }

    set({ converting: false, convertProgress: null });
  },

  reset: () => {
    const { files } = get();
    for (const entry of files) revokeEntry(entry);
    set({
      files: [],
      activeFileId: null,
      format: 'webp',
      options: { ...defaultOptions },
      converting: false,
      convertProgress: null,
    });
  },
}));
