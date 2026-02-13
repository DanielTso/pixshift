import { useConverterStore } from '../../stores/converter';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function DownloadButton() {
  const files = useConverterStore((s) => s.files);
  const format = useConverterStore((s) => s.format);
  const converting = useConverterStore((s) => s.converting);
  const convertProgress = useConverterStore((s) => s.convertProgress);
  const convert = useConverterStore((s) => s.convert);
  const reset = useConverterStore((s) => s.reset);

  if (files.length === 0) return null;

  const doneFiles = files.filter((f) => f.status === 'done');
  const errorFiles = files.filter((f) => f.status === 'error');
  const pendingFiles = files.filter((f) => f.status === 'pending' || f.status === 'error');
  const allDone = doneFiles.length === files.length;
  const hasDone = doneFiles.length > 0;
  const totalResultSize = doneFiles.reduce((sum, f) => sum + f.resultSize, 0);

  function handleDownloadAll() {
    for (const entry of doneFiles) {
      if (!entry.resultUrl) continue;
      const baseName = entry.file.name.replace(/\.[^.]+$/, '');
      const a = document.createElement('a');
      a.href = entry.resultUrl;
      a.download = `${baseName}.${format}`;
      a.click();
    }
  }

  return (
    <div className="flex flex-col items-center gap-3">
      {/* Error summary */}
      {errorFiles.length > 0 && !converting && (
        <p className="text-sm text-red-400">
          {errorFiles.length} file{errorFiles.length > 1 ? 's' : ''} failed to convert
        </p>
      )}

      <div className="flex flex-wrap items-center justify-center gap-3">
        {converting ? (
          /* During conversion: progress indicator */
          <button
            disabled
            className="rounded-xl bg-accent px-8 py-3 text-sm font-semibold text-navy-900 opacity-80 cursor-not-allowed"
          >
            <span className="flex items-center gap-2">
              <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" className="opacity-25" />
                <path d="M4 12a8 8 0 018-8" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              </svg>
              Converting {convertProgress ? `${convertProgress.current} of ${convertProgress.total}` : ''}...
            </span>
          </button>
        ) : allDone ? (
          /* All done: download + clear */
          <>
            <button
              onClick={handleDownloadAll}
              className="rounded-xl bg-accent px-8 py-3 text-sm font-semibold text-navy-900 transition hover:bg-accent-light hover:shadow-lg hover:shadow-accent/20"
            >
              {files.length === 1
                ? `Download (${formatBytes(totalResultSize)})`
                : `Download All (${formatBytes(totalResultSize)})`}
            </button>
            <button
              onClick={reset}
              className="rounded-xl border border-navy-600 px-6 py-3 text-sm font-medium text-navy-300 transition hover:border-navy-500 hover:text-white"
            >
              Clear All
            </button>
          </>
        ) : hasDone && pendingFiles.length > 0 ? (
          /* Partial: some done, some pending/error — retry + download available */
          <>
            <button
              onClick={convert}
              className="rounded-xl bg-accent px-8 py-3 text-sm font-semibold text-navy-900 transition hover:bg-accent-light hover:shadow-lg hover:shadow-accent/20"
            >
              Retry Failed ({pendingFiles.length})
            </button>
            <button
              onClick={handleDownloadAll}
              className="rounded-xl border border-accent px-6 py-3 text-sm font-medium text-accent transition hover:bg-accent/10"
            >
              Download {doneFiles.length} Completed ({formatBytes(totalResultSize)})
            </button>
            <button
              onClick={reset}
              className="rounded-xl border border-navy-600 px-6 py-3 text-sm font-medium text-navy-300 transition hover:border-navy-500 hover:text-white"
            >
              Clear All
            </button>
          </>
        ) : (
          /* Nothing converted yet: convert all */
          <button
            onClick={convert}
            className="rounded-xl bg-accent px-8 py-3 text-sm font-semibold text-navy-900 transition hover:bg-accent-light hover:shadow-lg hover:shadow-accent/20"
          >
            {files.length === 1 ? 'Convert' : `Convert All (${files.length} files)`}
          </button>
        )}
      </div>
    </div>
  );
}
