const codeBlockStyle: React.CSSProperties = {
  background: '#020617',
  borderRadius: '8px',
  padding: '16px',
  overflowX: 'auto',
  fontSize: '13px',
  lineHeight: '1.6',
  color: '#e2e8f0',
  border: '1px solid rgba(51, 65, 85, 0.5)',
};

const keywordStyle: React.CSSProperties = { color: '#06b6d4' };
const stringStyle: React.CSSProperties = { color: '#34d399' };
const commentStyle: React.CSSProperties = { color: '#64748b' };

export default function Docs() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-12 sm:px-6">
      <h1 className="mb-2 text-3xl font-bold text-white">API Documentation</h1>
      <p className="mb-12 text-navy-400">
        Integrate Pixshift image conversion into your application.
      </p>

      {/* Authentication */}
      <section className="mb-12">
        <h2 className="mb-4 text-xl font-semibold text-white">Authentication</h2>
        <p className="mb-4 text-sm text-navy-300">
          All API requests require an API key passed via the <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">X-API-Key</code> header.
          Generate keys from your <a href="/dashboard" className="text-accent hover:text-accent-light">Dashboard</a>.
        </p>
        <pre style={codeBlockStyle}>
          <code>
            <span style={keywordStyle}>curl</span> -H <span style={stringStyle}>"X-API-Key: pxs_your_api_key"</span> \{'\n'}
            {'  '}https://api.pixshift.dev/api/formats
          </code>
        </pre>
      </section>

      {/* Convert Endpoint */}
      <section className="mb-12">
        <h2 className="mb-4 text-xl font-semibold text-white">Convert Endpoint</h2>
        <div className="mb-3 flex items-center gap-2">
          <span className="rounded bg-green-500/15 px-2 py-0.5 text-xs font-semibold text-green-400">POST</span>
          <code className="text-sm text-navy-200">/api/convert</code>
        </div>
        <p className="mb-4 text-sm text-navy-300">
          Convert an image to a different format with optional transforms.
          Send as <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">multipart/form-data</code>.
        </p>

        <h3 className="mb-2 text-sm font-semibold text-navy-200">Parameters</h3>
        <div className="mb-4 overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-navy-700/50 text-left text-xs text-navy-500">
                <th className="pb-2 pr-4 font-medium">Field</th>
                <th className="pb-2 pr-4 font-medium">Type</th>
                <th className="pb-2 pr-4 font-medium">Required</th>
                <th className="pb-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody className="text-navy-300">
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">file</code></td><td className="py-2 pr-4">file</td><td className="py-2 pr-4">Yes</td><td className="py-2">Input image file</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">format</code></td><td className="py-2 pr-4">string</td><td className="py-2 pr-4">Yes</td><td className="py-2">Output format (jpeg, png, webp, avif, heic, tiff, gif, jxl)</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">quality</code></td><td className="py-2 pr-4">int</td><td className="py-2 pr-4">No</td><td className="py-2">Output quality 1-100 (default 92)</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">width</code></td><td className="py-2 pr-4">int</td><td className="py-2 pr-4">No</td><td className="py-2">Resize width in pixels</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">height</code></td><td className="py-2 pr-4">int</td><td className="py-2 pr-4">No</td><td className="py-2">Resize height in pixels</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">grayscale</code></td><td className="py-2 pr-4">bool</td><td className="py-2 pr-4">No</td><td className="py-2">Convert to grayscale</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">sharpen</code></td><td className="py-2 pr-4">bool</td><td className="py-2 pr-4">No</td><td className="py-2">Apply sharpening</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">blur</code></td><td className="py-2 pr-4">float</td><td className="py-2 pr-4">No</td><td className="py-2">Blur radius (0-20)</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">brightness</code></td><td className="py-2 pr-4">int</td><td className="py-2 pr-4">No</td><td className="py-2">Brightness adjustment (-100 to 100)</td></tr>
              <tr className="border-b border-navy-700/30"><td className="py-2 pr-4"><code className="text-accent">contrast</code></td><td className="py-2 pr-4">int</td><td className="py-2 pr-4">No</td><td className="py-2">Contrast adjustment (-100 to 100)</td></tr>
              <tr><td className="py-2 pr-4"><code className="text-accent">watermark_text</code></td><td className="py-2 pr-4">string</td><td className="py-2 pr-4">No</td><td className="py-2">Watermark text overlay</td></tr>
            </tbody>
          </table>
        </div>

        <h3 className="mb-2 text-sm font-semibold text-navy-200">Example</h3>
        <pre style={codeBlockStyle}>
          <code>
            <span style={keywordStyle}>curl</span> -X POST https://api.pixshift.dev/api/convert \{'\n'}
            {'  '}-H <span style={stringStyle}>"X-API-Key: pxs_your_api_key"</span> \{'\n'}
            {'  '}-F <span style={stringStyle}>"file=@photo.jpg"</span> \{'\n'}
            {'  '}-F <span style={stringStyle}>"format=webp"</span> \{'\n'}
            {'  '}-F <span style={stringStyle}>"quality=85"</span> \{'\n'}
            {'  '}-F <span style={stringStyle}>"width=800"</span> \{'\n'}
            {'  '}-o <span style={stringStyle}>output.webp</span>
          </code>
        </pre>

        <h3 className="mt-4 mb-2 text-sm font-semibold text-navy-200">Response</h3>
        <p className="text-sm text-navy-300">
          Returns the converted image as a binary stream with the appropriate <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">Content-Type</code> header.
          On error, returns JSON with an <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">error</code> field.
        </p>
      </section>

      {/* Formats Endpoint */}
      <section className="mb-12">
        <h2 className="mb-4 text-xl font-semibold text-white">Formats Endpoint</h2>
        <div className="mb-3 flex items-center gap-2">
          <span className="rounded bg-blue-500/15 px-2 py-0.5 text-xs font-semibold text-blue-400">GET</span>
          <code className="text-sm text-navy-200">/api/formats</code>
        </div>
        <p className="mb-4 text-sm text-navy-300">
          List all supported input and output formats.
        </p>
        <pre style={codeBlockStyle}>
          <code>
            <span style={keywordStyle}>curl</span> -H <span style={stringStyle}>"X-API-Key: pxs_your_api_key"</span> \{'\n'}
            {'  '}https://api.pixshift.dev/api/formats{'\n'}
            {'\n'}
            <span style={commentStyle}>{'// Response:'}</span>{'\n'}
            {'{'}{'\n'}
            {'  '}<span style={stringStyle}>"formats"</span>: [<span style={stringStyle}>"jpeg"</span>, <span style={stringStyle}>"png"</span>, <span style={stringStyle}>"webp"</span>, <span style={stringStyle}>"avif"</span>, <span style={stringStyle}>"heic"</span>, <span style={stringStyle}>"tiff"</span>, <span style={stringStyle}>"gif"</span>, <span style={stringStyle}>"jxl"</span>]{'\n'}
            {'}'}
          </code>
        </pre>
      </section>

      {/* MCP Integration */}
      <section className="mb-12">
        <h2 className="mb-4 text-xl font-semibold text-white">MCP Integration</h2>
        <p className="mb-4 text-sm text-navy-300">
          Pixshift can run as a <strong className="text-navy-200">Model Context Protocol (MCP)</strong> server,
          enabling AI assistants like Claude to convert images directly.
        </p>
        <h3 className="mb-2 text-sm font-semibold text-navy-200">Claude Desktop Configuration</h3>
        <pre style={codeBlockStyle}>
          <code>
            {'{'}{'\n'}
            {'  '}<span style={stringStyle}>"mcpServers"</span>: {'{'}{'\n'}
            {'    '}<span style={stringStyle}>"pixshift"</span>: {'{'}{'\n'}
            {'      '}<span style={stringStyle}>"command"</span>: <span style={stringStyle}>"pixshift"</span>,{'\n'}
            {'      '}<span style={stringStyle}>"args"</span>: [<span style={stringStyle}>"--mcp"</span>]{'\n'}
            {'    '}{'}'}{'\n'}
            {'  '}{'}'}{'\n'}
            {'}'}
          </code>
        </pre>
        <p className="mt-4 text-sm text-navy-300">
          Available MCP tools: <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">convert_image</code>,{' '}
          <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">list_formats</code>,{' '}
          <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">get_image_info</code>.
        </p>
      </section>

      {/* Rate Limits */}
      <section>
        <h2 className="mb-4 text-xl font-semibold text-white">Rate Limits</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-navy-700/50 text-left text-xs text-navy-500">
                <th className="pb-2 pr-4 font-medium">Tier</th>
                <th className="pb-2 pr-4 font-medium">Requests / day</th>
                <th className="pb-2 font-medium">Max file size</th>
              </tr>
            </thead>
            <tbody className="text-navy-300">
              <tr className="border-b border-navy-700/30">
                <td className="py-2 pr-4">Anonymous</td>
                <td className="py-2 pr-4">5</td>
                <td className="py-2">10 MB</td>
              </tr>
              <tr className="border-b border-navy-700/30">
                <td className="py-2 pr-4">Free</td>
                <td className="py-2 pr-4">100</td>
                <td className="py-2">10 MB</td>
              </tr>
              <tr>
                <td className="py-2 pr-4">Pro</td>
                <td className="py-2 pr-4">10,000</td>
                <td className="py-2">50 MB</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p className="mt-4 text-sm text-navy-400">
          Rate limit headers: <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">X-RateLimit-Remaining</code>,{' '}
          <code className="rounded bg-navy-800 px-1.5 py-0.5 text-accent">X-RateLimit-Reset</code>.
        </p>
      </section>
    </div>
  );
}
