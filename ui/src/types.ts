// TypeScript interfaces matching pkg/types/schema.go

export type DiffAction = 'add' | 'modify' | 'delete' | 'rename' | 'copy';

export type SecretType =
  | 'aws_key'
  | 'aws_secret'
  | 'github_token'
  | 'github_pat'
  | 'slack_token'
  | 'slack_webhook'
  | 'private_key'
  | 'gcp_key'
  | 'azure_key'
  | 'jwt'
  | 'generic_api'
  | 'password'
  | 'high_entropy'
  | 'database_url'
  | 'ssh_key'
  | 'telegram_token'
  | 'discord_token';

export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface Signature {
  name: string;
  email: string;
  when: string; // ISO timestamp
}

export interface FileDiff {
  path: string;
  oldPath?: string;
  action: DiffAction;
  isBinary: boolean;
  additions: number;
  deletions: number;
  patch?: string;
}

export interface CommitNode {
  hash: string;
  shortHash: string;
  author: Signature;
  committer: Signature;
  message: string;
  messageBody?: string;
  parents: string[];
  treeHash: string;
  files?: FileDiff[];
  secretHits: number;
  isMerge: boolean;
  isOrphan: boolean;
  reachableBy?: string[];
}

export interface BranchRef {
  name: string;
  hash: string;
  isRemote: boolean;
  upstream?: string;
}

export interface TagRef {
  name: string;
  hash: string;
  targetHash: string;
  isAnnotated: boolean;
  tagger?: string;
  message?: string;
}

export interface Refs {
  branches: BranchRef[];
  tags: TagRef[];
}

export interface RepoMeta {
  path: string;
  remotes: Record<string, string>;
  headRef: string;
  isBare: boolean;
  description?: string;
}

export interface StashEntry {
  index: number;
  hash: string;
  message: string;
  author: Signature;
  files?: FileDiff[];
}

export interface SecretMatch {
  type: SecretType;
  pattern: string;
  value: string;
  file: string;
  line: number;
  commitHash: string;
  entropy?: number;
  severity: Severity;
}

export interface ReconStats {
  totalCommits: number;
  orphanCommits: number;
  totalBranches: number;
  totalTags: number;
  totalStashes: number;
  totalSecrets: number;
  criticalSecrets: number;
  filesScanned: number;
  linesScanned: number;
}

export interface ReconData {
  meta: RepoMeta;
  refs: Refs;
  commits: CommitNode[];
  orphans: CommitNode[];
  stashes: StashEntry[];
  secrets: SecretMatch[];
  stats: ReconStats;
  scanTime: string; // ISO timestamp
}

// Global injection point for Go backend
declare global {
  interface Window {
    GIT_RECON_DATA?: ReconData;
  }
}
