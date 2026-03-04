import { useMemo, useState } from 'react';
import { Search, BarChart2, Ghost, GitBranch, KeyRound, FolderOpen, ArrowUp } from 'lucide-react';
import type { ReconData, CommitNode } from './types';
import GraphView from './components/GraphView';

// Mock data for local development when Go backend isn't injecting data
const MOCK_DATA: ReconData = {
  meta: {
    path: '/mock/repo/.git',
    remotes: {
      origin: 'git@github.com:target/secret-project.git',
    },
    headRef: 'refs/heads/main',
    isBare: false,
  },
  refs: {
    branches: [
      { name: 'main', hash: 'a1b2c3d4e5f6789012345678901234567890abcd', isRemote: false },
      { name: 'origin/main', hash: 'a1b2c3d4e5f6789012345678901234567890abcd', isRemote: true },
      { name: 'feature/auth', hash: 'b2c3d4e5f6789012345678901234567890abcde1', isRemote: false },
    ],
    tags: [
      { name: 'v1.0.0', hash: 'c3d4e5f6789012345678901234567890abcdef12', targetHash: 'c3d4e5f6789012345678901234567890abcdef12', isAnnotated: false },
    ],
  },
  commits: [
    {
      hash: 'a1b2c3d4e5f6789012345678901234567890abcd',
      shortHash: 'a1b2c3d',
      author: { name: 'Alice Dev', email: 'alice@example.com', when: '2026-03-04T10:30:00Z' },
      committer: { name: 'Alice Dev', email: 'alice@example.com', when: '2026-03-04T10:30:00Z' },
      message: 'feat: add user authentication',
      messageBody: 'Implements JWT-based auth flow with refresh tokens.',
      parents: ['b2c3d4e5f6789012345678901234567890abcde1'],
      treeHash: 'tree123',
      files: [
        { path: 'src/auth.ts', action: 'add', isBinary: false, additions: 150, deletions: 0, patch: '+const SECRET_KEY = "sk_live_abc123def456";\n+export function authenticate() {}' },
      ],
      secretHits: 1,
      isMerge: false,
      isOrphan: false,
      reachableBy: ['refs/heads/main'],
    },
    {
      hash: 'b2c3d4e5f6789012345678901234567890abcde1',
      shortHash: 'b2c3d4e',
      author: { name: 'Bob Hacker', email: 'bob@example.com', when: '2026-03-03T15:00:00Z' },
      committer: { name: 'Bob Hacker', email: 'bob@example.com', when: '2026-03-03T15:00:00Z' },
      message: 'chore: initial commit',
      parents: [],
      treeHash: 'tree456',
      files: [
        { path: 'README.md', action: 'add', isBinary: false, additions: 10, deletions: 0 },
      ],
      secretHits: 0,
      isMerge: false,
      isOrphan: false,
    },
    {
      hash: 'c3d4e5f6789012345678901234567890abcdef12',
      shortHash: 'c3d4e5f',
      author: { name: 'Alice Dev', email: 'alice@example.com', when: '2026-03-02T09:00:00Z' },
      committer: { name: 'Alice Dev', email: 'alice@example.com', when: '2026-03-02T09:00:00Z' },
      message: 'Merge branch feature/payments',
      parents: ['b2c3d4e5f6789012345678901234567890abcde1', 'd4e5f6789012345678901234567890abcdef123'],
      treeHash: 'tree789',
      secretHits: 0,
      isMerge: true,
      isOrphan: false,
    },
  ],
  orphans: [
    {
      hash: 'deadbeef1234567890123456789012345678dead',
      shortHash: 'deadbee',
      author: { name: 'Eve Intern', email: 'eve@example.com', when: '2026-02-28T14:00:00Z' },
      committer: { name: 'Eve Intern', email: 'eve@example.com', when: '2026-02-28T14:00:00Z' },
      message: 'WIP: add AWS credentials for testing',
      messageBody: 'DO NOT MERGE - contains real keys',
      parents: ['b2c3d4e5f6789012345678901234567890abcde1'],
      treeHash: 'tree_orphan',
      files: [
        { path: '.env', action: 'add', isBinary: false, additions: 5, deletions: 0, patch: '+AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\n+AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' },
      ],
      secretHits: 2,
      isMerge: false,
      isOrphan: true,
    },
  ],
  stashes: [],
  secrets: [
    {
      type: 'generic_api',
      pattern: 'Stripe Live Secret Key',
      value: 'sk_live_abc123def456...',
      file: 'src/auth.ts',
      line: 1,
      commitHash: 'a1b2c3d4e5f6789012345678901234567890abcd',
      severity: 'critical',
    },
    {
      type: 'aws_key',
      pattern: 'AWS Access Key ID',
      value: 'AKIAIOSFODNN7EXAMPLE',
      file: '.env',
      line: 1,
      commitHash: 'deadbeef1234567890123456789012345678dead',
      severity: 'critical',
    },
    {
      type: 'aws_secret',
      pattern: 'AWS Secret Access Key',
      value: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
      file: '.env',
      line: 2,
      commitHash: 'deadbeef1234567890123456789012345678dead',
      severity: 'critical',
    },
  ],
  stats: {
    totalCommits: 3,
    orphanCommits: 1,
    totalBranches: 3,
    totalTags: 1,
    totalStashes: 0,
    totalSecrets: 3,
    criticalSecrets: 3,
    filesScanned: 4,
    linesScanned: 165,
  },
  scanTime: '2026-03-04T12:00:00Z',
};

function useReconData(): ReconData {
  return useMemo(() => {
    // Check for Go-injected data first
    if (window.GIT_RECON_DATA) {
      return window.GIT_RECON_DATA;
    }
    // Fallback to mock data for local dev
    console.warn('[git-recon] No injected data found, using mock data');
    return MOCK_DATA;
  }, []);
}

export default function App() {
  const data = useReconData();
  const [selectedCommit, setSelectedCommit] = useState<CommitNode | null>(null);

  const severityColor = (sev: string) => {
    switch (sev) {
      case 'critical': return '#f85149';
      case 'high': return '#db6d28';
      case 'medium': return '#d29922';
      case 'low': return '#3fb950';
      default: return '#8b949e';
    }
  };

  // Get secrets related to selected commit
  const commitSecrets = selectedCommit 
    ? data.secrets.filter(s => s.commitHash === selectedCommit.hash)
    : [];

  return (
    <div style={{ 
      display: 'flex', 
      flexDirection: 'column', 
      height: '100vh', 
      overflow: 'hidden',
      background: '#0d1117',
    }}>
      {/* Header */}
      <header style={{ 
        borderBottom: '1px solid #30363d', 
        padding: '12px 20px',
        flexShrink: 0,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h1 style={{ fontSize: '20px', fontWeight: 600, margin: 0, display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Search size={20} />
            git-recon
            <span style={{ fontSize: '12px', color: '#8b949e', marginLeft: '12px', fontWeight: 400 }}>
              {data.meta.path}
            </span>
          </h1>
          <div style={{ display: 'flex', gap: '20px', fontSize: '13px', alignItems: 'center' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              <BarChart2 size={14} /> {data.stats.totalCommits}
            </span>
            <span style={{ color: '#f85149', display: 'flex', alignItems: 'center', gap: '4px' }}>
              <Ghost size={14} /> {data.stats.orphanCommits}
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              <GitBranch size={14} /> {data.stats.totalBranches}
            </span>
            <span style={{ color: data.stats.criticalSecrets > 0 ? '#f85149' : '#3fb950', display: 'flex', alignItems: 'center', gap: '4px' }}>
              <KeyRound size={14} /> {data.stats.totalSecrets}
            </span>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Graph Panel */}
        <div style={{ flex: 1, position: 'relative' }}>
          <GraphView data={data} onNodeClick={setSelectedCommit} />
        </div>

        {/* Side Panel */}
        <aside style={{ 
          width: selectedCommit ? '400px' : '0px',
          borderLeft: selectedCommit ? '1px solid #30363d' : 'none',
          background: '#161b22',
          overflow: 'auto',
          transition: 'width 0.2s ease',
          flexShrink: 0,
        }}>
          {selectedCommit && (
            <div style={{ padding: '16px' }}>
              {/* Close button */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                <h2 style={{ fontSize: '14px', margin: 0, color: '#8b949e' }}>COMMIT DETAILS</h2>
                <button
                  onClick={() => setSelectedCommit(null)}
                  style={{
                    background: 'none',
                    border: 'none',
                    color: '#8b949e',
                    cursor: 'pointer',
                    fontSize: '18px',
                    padding: '4px',
                  }}
                >
                  ✕
                </button>
              </div>

              {/* Commit Hash */}
              <div style={{ marginBottom: '16px' }}>
                <code style={{ 
                  color: selectedCommit.isOrphan ? '#f85149' : '#58a6ff',
                  fontSize: '16px',
                  fontWeight: 600,
                }}>
                  {selectedCommit.shortHash}
                  {selectedCommit.isOrphan && (
                    <span style={{ marginLeft: '8px', display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                      <Ghost size={14} /> ORPHAN
                    </span>
                  )}
                </code>
                <div style={{ color: '#8b949e', fontSize: '11px', marginTop: '4px', wordBreak: 'break-all' }}>
                  {selectedCommit.hash}
                </div>
              </div>

              {/* Message */}
              <div style={{ marginBottom: '16px' }}>
                <div style={{ color: '#c9d1d9', fontWeight: 500 }}>{selectedCommit.message}</div>
                {selectedCommit.messageBody && (
                  <div style={{ color: '#8b949e', fontSize: '13px', marginTop: '8px', whiteSpace: 'pre-wrap' }}>
                    {selectedCommit.messageBody}
                  </div>
                )}
              </div>

              {/* Author */}
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                fontSize: '13px',
                color: '#8b949e',
                marginBottom: '16px',
                paddingBottom: '16px',
                borderBottom: '1px solid #30363d',
              }}>
                <span>{selectedCommit.author.name}</span>
                <span>{new Date(selectedCommit.author.when).toLocaleString()}</span>
              </div>

              {commitSecrets.length > 0 && (
                <div style={{ marginBottom: '16px' }}>
                  <h3 style={{ fontSize: '13px', color: '#f85149', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <KeyRound size={14} /> Secrets Found ({commitSecrets.length})
                  </h3>
                  {commitSecrets.map((s, i) => (
                    <div
                      key={i}
                      style={{
                        background: 'rgba(248, 81, 73, 0.1)',
                        border: '1px solid rgba(248, 81, 73, 0.3)',
                        borderRadius: '6px',
                        padding: '10px',
                        marginBottom: '8px',
                        fontSize: '12px',
                        fontFamily: 'monospace',
                      }}
                    >
                      <div style={{ color: severityColor(s.severity), fontWeight: 600, marginBottom: '4px' }}>
                        {s.pattern}
                      </div>
                      <div style={{ color: '#7ee787' }}>{s.file}:{s.line}</div>
                      <div style={{ 
                        color: '#f0883e', 
                        marginTop: '4px', 
                        wordBreak: 'break-all',
                        maxHeight: '60px',
                        overflow: 'auto',
                      }}>
                        {s.value}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {selectedCommit.files && selectedCommit.files.length > 0 && (
                <div>
                  <h3 style={{ fontSize: '13px', color: '#8b949e', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <FolderOpen size={14} /> Files Changed ({selectedCommit.files.length})
                  </h3>
                  {selectedCommit.files.map((f, i) => (
                    <div
                      key={i}
                      style={{
                        background: '#0d1117',
                        borderRadius: '6px',
                        padding: '10px',
                        marginBottom: '8px',
                        fontSize: '12px',
                        fontFamily: 'monospace',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                        <span style={{ 
                          color: f.action === 'add' ? '#3fb950' : 
                                 f.action === 'delete' ? '#f85149' : '#d29922' 
                        }}>
                          {f.action === 'add' ? '+' : f.action === 'delete' ? '-' : '~'} {f.path}
                        </span>
                        <span style={{ color: '#8b949e' }}>
                          <span style={{ color: '#3fb950' }}>+{f.additions}</span>
                          {' / '}
                          <span style={{ color: '#f85149' }}>-{f.deletions}</span>
                        </span>
                      </div>
                      {f.patch && (
                        <pre style={{
                          background: '#161b22',
                          padding: '8px',
                          borderRadius: '4px',
                          overflow: 'auto',
                          maxHeight: '150px',
                          fontSize: '11px',
                          margin: '8px 0 0 0',
                          color: '#8b949e',
                        }}>
                          {f.patch}
                        </pre>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {selectedCommit.parents.length > 0 && (
                <div style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid #30363d' }}>
                  <h3 style={{ fontSize: '13px', color: '#8b949e', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <ArrowUp size={14} /> Parents
                  </h3>
                  {selectedCommit.parents.map((p, i) => (
                    <code key={i} style={{ 
                      display: 'block',
                      color: '#58a6ff', 
                      fontSize: '12px',
                      marginBottom: '4px',
                    }}>
                      {p.slice(0, 7)}
                    </code>
                  ))}
                </div>
              )}
            </div>
          )}
        </aside>
      </div>

      {/* Footer Stats */}
      <footer style={{
        borderTop: '1px solid #30363d',
        padding: '8px 20px',
        fontSize: '11px',
        color: '#8b949e',
        display: 'flex',
        justifyContent: 'space-between',
        flexShrink: 0,
      }}>
        <span>Scanned: {new Date(data.scanTime).toLocaleString()}</span>
        <span>
          {data.stats.filesScanned} files · {data.stats.linesScanned} lines
        </span>
      </footer>
    </div>
  );
}
