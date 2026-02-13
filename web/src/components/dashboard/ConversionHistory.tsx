import { useState, useEffect } from 'react';
import { apiFetch } from '../../api/client';

interface Conversion {
  id: string;
  created_at: string;
  input_format: string;
  output_format: string;
  input_size: number;
  output_size: number;
  source: string;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function ConversionHistory() {
  const [conversions, setConversions] = useState<Conversion[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch<{ conversions: Conversion[] }>('/internal/conversions?limit=20')
      .then((res) => setConversions(res.conversions))
      .catch(() => setConversions([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="rounded-xl border border-navy-700/50 bg-navy-800/50 p-6">
      <h3 className="mb-4 text-sm font-semibold text-white">Recent Conversions</h3>

      {loading ? (
        <div className="py-8 text-center text-sm text-navy-500">Loading...</div>
      ) : conversions.length === 0 ? (
        <div className="py-8 text-center text-sm text-navy-500">No conversions yet</div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-navy-700/50 text-left text-xs text-navy-500">
                <th className="pb-2 pr-4 font-medium">Date</th>
                <th className="pb-2 pr-4 font-medium">Input</th>
                <th className="pb-2 pr-4 font-medium">Output</th>
                <th className="pb-2 pr-4 font-medium">Reduction</th>
                <th className="pb-2 font-medium">Source</th>
              </tr>
            </thead>
            <tbody>
              {conversions.map((c) => {
                const reduction = c.input_size > 0
                  ? ((1 - c.output_size / c.input_size) * 100).toFixed(1)
                  : '0';
                return (
                  <tr key={c.id} className="border-b border-navy-700/30 text-navy-300">
                    <td className="py-2.5 pr-4 text-navy-400">
                      {new Date(c.created_at).toLocaleDateString()}
                    </td>
                    <td className="py-2.5 pr-4">
                      <span className="rounded bg-navy-700 px-1.5 py-0.5 text-xs uppercase">
                        {c.input_format}
                      </span>
                    </td>
                    <td className="py-2.5 pr-4">
                      <span className="rounded bg-accent/15 px-1.5 py-0.5 text-xs uppercase text-accent">
                        {c.output_format}
                      </span>
                    </td>
                    <td className="py-2.5 pr-4">
                      <span className={Number(reduction) > 0 ? 'text-green-400' : 'text-navy-400'}>
                        {Number(reduction) > 0 ? `-${reduction}%` : `${reduction}%`}
                      </span>
                      <span className="ml-1 text-xs text-navy-500">
                        ({formatBytes(c.input_size)} → {formatBytes(c.output_size)})
                      </span>
                    </td>
                    <td className="py-2.5 capitalize text-navy-400">{c.source}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
