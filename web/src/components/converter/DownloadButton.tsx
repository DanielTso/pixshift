import { useConverterStore } from '../../stores/converter';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function DownloadButton() {
  const file = useConverterStore((s) => s.file);
  const result = useConverterStore((s) => s.result);
  const resultUrl = useConverterStore((s) => s.resultUrl);
  const resultSize = useConverterStore((s) => s.resultSize);
  const format = useConverterStore((s) => s.format);
  const converting = useConverterStore((s) => s.converting);
  const error = useConverterStore((s) => s.error);
  const convert = useConverterStore((s) => s.convert);
  const reset = useConverterStore((s) => s.reset);

  if (!file) return null;

  function handleDownload() {
    if (!result || !resultUrl) return;
    const baseName = file!.name.replace(/\.[^.]+$/, '');
    const a = document.createElement('a');
    a.href = resultUrl;
    a.download = `${baseName}.${format}`;
    a.click();
  }

  return (
    <div className="flex flex-col items-center gap-3">
      {error && (
        <p className="text-sm text-red-400">{error}</p>
      )}

      <div className="flex items-center gap-3">
        {result ? (
          <>
            <button
              onClick={handleDownload}
              className="rounded-xl bg-accent px-8 py-3 text-sm font-semibold text-navy-900 transition hover:bg-accent-light hover:shadow-lg hover:shadow-accent/20"
            >
              Download ({formatBytes(resultSize)})
            </button>
            <button
              onClick={reset}
              className="rounded-xl border border-navy-600 px-6 py-3 text-sm font-medium text-navy-300 transition hover:border-navy-500 hover:text-white"
            >
              Convert another
            </button>
          </>
        ) : (
          <button
            onClick={convert}
            disabled={converting}
            className="rounded-xl bg-accent px-8 py-3 text-sm font-semibold text-navy-900 transition hover:bg-accent-light hover:shadow-lg hover:shadow-accent/20 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {converting ? (
              <span className="flex items-center gap-2">
                <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" className="opacity-25" />
                  <path d="M4 12a8 8 0 018-8" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
                </svg>
                Converting...
              </span>
            ) : (
              'Convert'
            )}
          </button>
        )}
      </div>
    </div>
  );
}
