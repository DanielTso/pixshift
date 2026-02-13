import { create } from 'zustand';
import { convertImage } from '../api/convert';
import type { ConvertOptions } from '../api/convert';

interface Dimensions {
  width: number;
  height: number;
}

interface ConverterState {
  file: File | null;
  preview: string | null;
  format: string;
  options: Omit<ConvertOptions, 'format'>;
  result: Blob | null;
  resultUrl: string | null;
  converting: boolean;
  error: string | null;
  originalSize: number;
  resultSize: number;
  imageDimensions: Dimensions | null;
  resultDimensions: Dimensions | null;
  setFile: (file: File | null) => void;
  setFormat: (format: string) => void;
  setOptions: (options: Partial<Omit<ConvertOptions, 'format'>>) => void;
  convert: () => Promise<void>;
  reset: () => void;
}

const defaultOptions: Omit<ConvertOptions, 'format'> = {
  quality: 92,
};

export const useConverterStore = create<ConverterState>((set, get) => ({
  file: null,
  preview: null,
  format: 'webp',
  options: { ...defaultOptions },
  result: null,
  resultUrl: null,
  converting: false,
  error: null,
  originalSize: 0,
  resultSize: 0,
  imageDimensions: null,
  resultDimensions: null,

  setFile: (file) => {
    const prev = get().preview;
    if (prev) URL.revokeObjectURL(prev);
    const prevResult = get().resultUrl;
    if (prevResult) URL.revokeObjectURL(prevResult);

    if (file) {
      const preview = URL.createObjectURL(file);
      const img = new Image();
      img.onload = () => {
        set({ imageDimensions: { width: img.naturalWidth, height: img.naturalHeight } });
      };
      img.src = preview;
      set({ file, preview, result: null, resultUrl: null, error: null, originalSize: file.size, resultSize: 0, imageDimensions: null, resultDimensions: null });
    } else {
      set({ file: null, preview: null, result: null, resultUrl: null, error: null, originalSize: 0, resultSize: 0, imageDimensions: null, resultDimensions: null });
    }
  },

  setFormat: (format) => set({ format }),

  setOptions: (opts) =>
    set((state) => ({ options: { ...state.options, ...opts } })),

  convert: async () => {
    const { file, format, options } = get();
    if (!file) return;

    set({ converting: true, error: null, result: null });
    try {
      const blob = await convertImage(file, { format, ...options });
      const resultUrl = URL.createObjectURL(blob);
      const img = new Image();
      img.onload = () => {
        set({ resultDimensions: { width: img.naturalWidth, height: img.naturalHeight } });
      };
      img.src = resultUrl;
      set({ result: blob, resultUrl, converting: false, resultSize: blob.size, resultDimensions: null });
    } catch (err) {
      set({ error: (err as Error).message, converting: false });
    }
  },

  reset: () => {
    const prev = get().preview;
    if (prev) URL.revokeObjectURL(prev);
    const prevResult = get().resultUrl;
    if (prevResult) URL.revokeObjectURL(prevResult);
    set({
      file: null,
      preview: null,
      format: 'webp',
      options: { ...defaultOptions },
      result: null,
      resultUrl: null,
      converting: false,
      error: null,
      originalSize: 0,
      resultSize: 0,
      imageDimensions: null,
      resultDimensions: null,
    });
  },
}));
