import { useState } from 'react'

const downloads = [
  { os: 'Windows', arch: 'x64', file: 'zap_1.0.0_windows_amd64.zip', icon: '[]' },
  { os: 'macOS', arch: 'Apple Silicon', file: 'zap_1.0.0_darwin_arm64.tar.gz', icon: '()' },
  { os: 'macOS', arch: 'Intel', file: 'zap_1.0.0_darwin_amd64.tar.gz', icon: '()' },
  { os: 'Linux', arch: 'x64', file: 'zap_1.0.0_linux_amd64.tar.gz', icon: '><' },
  { os: 'Linux', arch: 'ARM64', file: 'zap_1.0.0_linux_arm64.tar.gz', icon: '><' },
]

const baseUrl = 'https://github.com/blackcoderx/zap/releases/download/v1.0.0'

function Install() {
  const [copied, setCopied] = useState(false)
  const installCmd = 'git clone https://github.com/blackcoderx/zap.git && cd zap && go build -o zap ./cmd/zap'

  const copyToClipboard = () => {
    navigator.clipboard.writeText(installCmd)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <section id="install" className="bg-smoke py-16 border-b-4 border-ash">
      <div className="max-w-4xl mx-auto px-6">
        <h2 className="text-2xl font-bold mb-8 text-center">
          <span className="text-mustard">&gt;</span> Download
        </h2>

        {/* Download buttons */}
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-12">
          {downloads.map((d) => (
            <a
              key={d.file}
              href={`${baseUrl}/${d.file}`}
              className="border-4 border-ash hover:border-mustard p-4 text-center transition-colors group"
            >
              <div className="text-2xl font-bold text-mustard mb-1 font-mono">{d.icon}</div>
              <div className="font-bold text-sm">{d.os}</div>
              <div className="text-silver text-xs">{d.arch}</div>
            </a>
          ))}
        </div>

        <div className="text-center mb-12">
          <a
            href="https://github.com/blackcoderx/zap/releases/latest"
            className="text-mustard hover:underline text-sm"
          >
            View all releases &rarr;
          </a>
        </div>

        <h3 className="text-xl font-bold mb-6 text-center">
          <span className="text-silver">&gt;</span> Or build from source
        </h3>

        <div className="space-y-8">
          <div>
            <p className="text-silver mb-3 text-sm">Prerequisites: Go 1.25+ and Ollama (or Gemini API key)</p>
            <div className="bg-charcoal border-4 border-ash p-4 flex justify-between items-center gap-4">
              <code className="text-mustard text-sm overflow-x-auto flex-1">
                {installCmd}
              </code>
              <button
                onClick={copyToClipboard}
                className="border-2 border-silver text-silver px-3 py-1 text-sm hover:border-mustard hover:text-mustard transition-colors shrink-0"
              >
                {copied ? 'COPIED!' : 'COPY'}
              </button>
            </div>
          </div>

          <div className="border-4 border-ash p-6">
            <h3 className="font-bold mb-4">Quick Setup</h3>
            <ol className="space-y-3 text-silver text-sm">
              <li className="flex gap-3">
                <span className="text-mustard font-bold">1.</span>
                <span>Run <code className="text-bone">./zap</code> to start the setup wizard</span>
              </li>
              <li className="flex gap-3">
                <span className="text-mustard font-bold">2.</span>
                <span>Select your LLM provider (Ollama local, Ollama cloud, or Gemini)</span>
              </li>
              <li className="flex gap-3">
                <span className="text-mustard font-bold">3.</span>
                <span>Choose your API framework (gin, fastapi, express, etc.)</span>
              </li>
              <li className="flex gap-3">
                <span className="text-mustard font-bold">4.</span>
                <span>Start debugging: <code className="text-bone">&gt; GET http://localhost:8000/api/users</code></span>
              </li>
            </ol>
          </div>

          <div className="grid md:grid-cols-2 gap-6">
            <div className="border-4 border-ash p-6">
              <h3 className="font-bold mb-3">Keyboard Shortcuts</h3>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-silver">Send message</span>
                  <code className="text-mustard">Enter</code>
                </div>
                <div className="flex justify-between">
                  <span className="text-silver">Input history</span>
                  <code className="text-mustard">Shift+Up/Down</code>
                </div>
                <div className="flex justify-between">
                  <span className="text-silver">Copy response</span>
                  <code className="text-mustard">Ctrl+Y</code>
                </div>
                <div className="flex justify-between">
                  <span className="text-silver">Stop/Quit</span>
                  <code className="text-mustard">Esc</code>
                </div>
              </div>
            </div>

            <div className="border-4 border-ash p-6">
              <h3 className="font-bold mb-3">Supported Frameworks</h3>
              <div className="flex flex-wrap gap-2 text-sm">
                {['gin', 'echo', 'chi', 'fiber', 'fastapi', 'flask', 'django', 'express', 'nestjs', 'hono', 'spring', 'laravel', 'rails', 'actix', 'axum'].map(fw => (
                  <span key={fw} className="border-2 border-ash px-2 py-1 text-silver">
                    {fw}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

export default Install
