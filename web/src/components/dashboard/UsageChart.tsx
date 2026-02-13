import { useState, useEffect } from 'react';
import { apiFetch } from '../../api/client';

interface DailyUsage {
  date: string;
  count: number;
}

export default function UsageChart() {
  const [data, setData] = useState<DailyUsage[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch<{ days: DailyUsage[] }>('/internal/usage/daily')
      .then((res) => setData(res.days))
      .catch(() => setData([]))
      .finally(() => setLoading(false));
  }, []);

  const maxCount = Math.max(...data.map((d) => d.count), 1);

  if (loading) {
    return (
      <div className="rounded-xl border border-navy-700/50 bg-navy-800/50 p-6">
        <h3 className="mb-4 text-sm font-semibold text-white">Daily Conversions</h3>
        <div className="flex h-40 items-center justify-center text-sm text-navy-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-navy-700/50 bg-navy-800/50 p-6">
      <h3 className="mb-4 text-sm font-semibold text-white">Daily Conversions (7 days)</h3>
      <div className="flex h-40 items-end gap-2">
        {data.length === 0 ? (
          <div className="flex h-full w-full items-center justify-center text-sm text-navy-500">
            No conversion data yet
          </div>
        ) : (
          data.map((day) => (
            <div key={day.date} className="flex flex-1 flex-col items-center gap-1">
              <span className="text-xs text-navy-400">{day.count}</span>
              <div
                className="w-full rounded-t-md bg-accent/80 transition-all hover:bg-accent"
                style={{
                  height: `${Math.max((day.count / maxCount) * 120, 4)}px`,
                }}
              />
              <span className="text-[10px] text-navy-500">
                {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
