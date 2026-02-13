import { useConverterStore } from '../stores/converter';
import DropZone from '../components/converter/DropZone';
import PreviewPane from '../components/converter/PreviewPane';
import FormatPicker from '../components/converter/FormatPicker';
import TransformPanel from '../components/converter/TransformPanel';
import DownloadButton from '../components/converter/DownloadButton';

export default function Home() {
  const file = useConverterStore((s) => s.file);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      {!file ? (
        <div className="flex flex-col items-center gap-8">
          <div className="text-center">
            <h1 className="mb-3 text-4xl font-bold tracking-tight text-white sm:text-5xl">
              Convert images <span className="text-accent">instantly</span>
            </h1>
            <p className="mx-auto max-w-xl text-lg text-navy-400">
              Drop any image and convert to JPEG, PNG, WebP, AVIF, HEIC, TIFF, GIF, or JXL.
              Free, fast, and private.
            </p>
          </div>
          <div className="w-full max-w-2xl">
            <DropZone />
          </div>
          <div className="flex items-center gap-6 text-sm text-navy-500">
            <span className="flex items-center gap-1.5">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              </svg>
              Private &mdash; files never stored
            </span>
            <span className="flex items-center gap-1.5">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
              20 free conversions / day
            </span>
          </div>
        </div>
      ) : (
        <div className="flex flex-col gap-6 lg:flex-row">
          <div className="flex flex-1 flex-col gap-6">
            <PreviewPane />
            <FormatPicker />
            <DownloadButton />
          </div>
          <div className="w-full lg:w-80">
            <TransformPanel />
          </div>
        </div>
      )}
    </div>
  );
}
