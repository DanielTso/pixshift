export interface ConvertOptions {
  format: string;
  quality?: number;
  width?: number;
  height?: number;
  grayscale?: boolean;
  sharpen?: boolean;
  blur?: number;
  invert?: boolean;
  sepia?: number;
  brightness?: number;
  contrast?: number;
  watermarkText?: string;
  watermarkPosition?: string;
  watermarkOpacity?: number;
}

export async function convertImage(file: File, options: ConvertOptions): Promise<Blob> {
  const form = new FormData();
  form.append('file', file);
  form.append('format', options.format);

  if (options.quality !== undefined) form.append('quality', String(options.quality));
  if (options.width) form.append('width', String(options.width));
  if (options.height) form.append('height', String(options.height));
  if (options.grayscale) form.append('grayscale', 'true');
  if (options.sharpen) form.append('sharpen', 'true');
  if (options.blur) form.append('blur', String(options.blur));
  if (options.invert) form.append('invert', 'true');
  if (options.sepia) form.append('sepia', String(options.sepia));
  if (options.brightness !== undefined) form.append('brightness', String(options.brightness));
  if (options.contrast !== undefined) form.append('contrast', String(options.contrast));
  if (options.watermarkText) form.append('watermark_text', options.watermarkText);
  if (options.watermarkPosition) form.append('watermark_position', options.watermarkPosition);
  if (options.watermarkOpacity !== undefined) form.append('watermark_opacity', String(options.watermarkOpacity));

  const res = await fetch('/internal/convert', {
    method: 'POST',
    body: form,
    credentials: 'include',
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }

  return res.blob();
}
