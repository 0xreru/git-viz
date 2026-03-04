import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Ghost, ShieldAlert, GitMerge, KeyRound } from 'lucide-react';
import type { CommitNode as CommitNodeType } from '../types';

export type CommitNodeData = CommitNodeType & Record<string, unknown>;

interface CommitNodeProps {
  data: CommitNodeData;
}

const CommitNode = memo(({ data }: CommitNodeProps) => {
  const isOrphan = data.isOrphan;
  const hasSecrets = data.secretHits > 0;
  const isMerge = data.isMerge;

  // Truncate message to 25 chars
  const shortMessage = data.message.length > 25 
    ? data.message.slice(0, 22) + '...' 
    : data.message;

  return (
    <>
      <Handle
        type="target"
        position={Position.Top}
        style={{ background: '#30363d', border: '1px solid #484f58', width: 8, height: 8 }}
      />
      
      <div
        style={{
          background: isOrphan ? '#1a0a0a' : '#161b22',
          border: `2px solid ${
            isOrphan ? '#f85149' : 
            hasSecrets ? '#d29922' : 
            isMerge ? '#a371f7' : '#30363d'
          }`,
          borderRadius: '8px',
          padding: '8px 12px',
          minWidth: '160px',
          maxWidth: '200px',
          fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace',
          boxShadow: isOrphan 
            ? '0 0 12px rgba(248, 81, 73, 0.4)' 
            : hasSecrets 
              ? '0 0 8px rgba(210, 153, 34, 0.3)'
              : '0 2px 8px rgba(0,0,0,0.3)',
          cursor: 'pointer',
          transition: 'all 0.15s ease',
        }}
      >
        {/* Header: Hash + Badges */}
        <div style={{ 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          marginBottom: '4px',
        }}>
          <code style={{ 
            color: isOrphan ? '#f85149' : '#58a6ff', 
            fontSize: '13px',
            fontWeight: 600,
          }}>
            {data.shortHash}
          </code>
          
          <div style={{ display: 'flex', gap: '4px', alignItems: 'center' }}>
            {isOrphan && (
              <span title="Orphan/Dangling Commit" style={{ color: '#f85149' }}>
                <Ghost size={13} />
              </span>
            )}
            {hasSecrets && (
              <span 
                title={`${data.secretHits} secret(s) found`} 
                style={{ 
                  color: '#f85149',
                  animation: 'pulse 1.5s infinite',
                }}
              >
                <ShieldAlert size={13} />
              </span>
            )}
            {isMerge && (
              <span title="Merge Commit" style={{ color: '#a371f7' }}>
                <GitMerge size={13} />
              </span>
            )}
          </div>
        </div>

        {/* Commit Message */}
        <div style={{ 
          color: '#c9d1d9', 
          fontSize: '11px',
          lineHeight: '1.3',
          wordBreak: 'break-word',
        }}>
          {shortMessage}
        </div>

        {/* Author + Time */}
        <div style={{ 
          color: '#8b949e', 
          fontSize: '10px',
          marginTop: '4px',
          display: 'flex',
          justifyContent: 'space-between',
        }}>
          <span>{data.author.name.split(' ')[0]}</span>
          <span>{new Date(data.author.when).toLocaleDateString()}</span>
        </div>

        {/* Secret indicator bar */}
        {hasSecrets && (
          <div style={{
            marginTop: '6px',
            padding: '2px 6px',
            background: 'rgba(248, 81, 73, 0.15)',
            border: '1px solid rgba(248, 81, 73, 0.3)',
            borderRadius: '4px',
            fontSize: '10px',
            color: '#f85149',
            textAlign: 'center',
          }}>
            <KeyRound size={10} style={{ marginRight: '4px', verticalAlign: 'middle' }} />
            {data.secretHits} secret{data.secretHits > 1 ? 's' : ''}
          </div>
        )}
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        style={{ background: '#30363d', border: '1px solid #484f58', width: 8, height: 8 }}
      />

      {/* Inline keyframes for pulse animation */}
      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
      `}</style>
    </>
  );
});

CommitNode.displayName = 'CommitNode';

export default CommitNode;
