import { useState, useEffect } from 'react';
import { apiFetch } from '../../api/client';

interface APIKey {
  id: string;
  prefix: string;
  name: string;
  created_at: string;
}

export default function APIKeyManager() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyFull, setNewKeyFull] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    loadKeys();
  }, []);

  async function loadKeys() {
    try {
      const res = await apiFetch<{ keys: APIKey[] }>('/internal/apikeys');
      setKeys(res.keys);
    } catch {
      setKeys([]);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!newKeyName.trim()) return;
    setCreating(true);
    try {
      const res = await apiFetch<{ key: string; api_key: APIKey }>('/internal/apikeys', {
        method: 'POST',
        body: JSON.stringify({ name: newKeyName }),
      });
      setNewKeyFull(res.key);
      setNewKeyName('');
      await loadKeys();
    } catch {
      // silently fail
    } finally {
      setCreating(false);
    }
  }

  async function handleRevoke(id: string) {
    try {
      await apiFetch(`/internal/apikeys/${id}`, { method: 'DELETE' });
      setKeys((prev) => prev.filter((k) => k.id !== id));
    } catch {
      // silently fail
    }
  }

  function handleCopy() {
    if (!newKeyFull) return;
    navigator.clipboard.writeText(newKeyFull);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="rounded-xl border border-navy-700/50 bg-navy-800/50 p-6">
      <h3 className="mb-4 text-sm font-semibold text-white">API Keys</h3>

      {newKeyFull && (
        <div className="mb-4 rounded-lg border border-accent/30 bg-accent/5 p-4">
          <p className="mb-2 text-xs text-accent">
            Copy this key now — it won't be shown again.
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 overflow-auto rounded-md bg-navy-900 px-3 py-2 text-sm text-white">
              {newKeyFull}
            </code>
            <button
              onClick={handleCopy}
              className="rounded-lg bg-accent/15 px-3 py-2 text-xs font-medium text-accent transition hover:bg-accent/25"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
          <button
            onClick={() => setNewKeyFull(null)}
            className="mt-2 text-xs text-navy-400 hover:text-navy-300"
          >
            Dismiss
          </button>
        </div>
      )}

      <form onSubmit={handleCreate} className="mb-4 flex gap-2">
        <input
          type="text"
          placeholder="Key name (e.g., Production)"
          value={newKeyName}
          onChange={(e) => setNewKeyName(e.target.value)}
          className="flex-1 rounded-lg border border-navy-600 bg-navy-900 px-3 py-2 text-sm text-white placeholder-navy-500 outline-none focus:border-accent"
        />
        <button
          type="submit"
          disabled={creating || !newKeyName.trim()}
          className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-navy-900 transition hover:bg-accent-light disabled:opacity-50"
        >
          {creating ? 'Creating...' : 'Create New Key'}
        </button>
      </form>

      {loading ? (
        <div className="py-4 text-center text-sm text-navy-500">Loading...</div>
      ) : keys.length === 0 ? (
        <div className="py-4 text-center text-sm text-navy-500">No API keys yet</div>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-navy-700/50 text-left text-xs text-navy-500">
              <th className="pb-2 pr-4 font-medium">Key</th>
              <th className="pb-2 pr-4 font-medium">Name</th>
              <th className="pb-2 pr-4 font-medium">Created</th>
              <th className="pb-2 font-medium"></th>
            </tr>
          </thead>
          <tbody>
            {keys.map((k) => (
              <tr key={k.id} className="border-b border-navy-700/30 text-navy-300">
                <td className="py-2.5 pr-4">
                  <code className="text-xs text-navy-400">{k.prefix}...</code>
                </td>
                <td className="py-2.5 pr-4">{k.name}</td>
                <td className="py-2.5 pr-4 text-navy-400">
                  {new Date(k.created_at).toLocaleDateString()}
                </td>
                <td className="py-2.5">
                  <button
                    onClick={() => handleRevoke(k.id)}
                    className="text-xs text-red-400 transition hover:text-red-300"
                  >
                    Revoke
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
