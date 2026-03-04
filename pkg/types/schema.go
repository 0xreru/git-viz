package types

import "time"

type ReconData struct {
	Meta     RepoMeta      `json:"meta"`
	Refs     Refs          `json:"refs"`
	Commits  []CommitNode  `json:"commits"`
	Orphans  []CommitNode  `json:"orphans"`
	Stashes  []StashEntry  `json:"stashes"`
	Secrets  []SecretMatch `json:"secrets"`
	Stats    ReconStats    `json:"stats"`
	ScanTime time.Time     `json:"scanTime"`
}

type RepoMeta struct {
	Path        string            `json:"path"`
	Remotes     map[string]string `json:"remotes"`
	HeadRef     string            `json:"headRef"`
	IsBare      bool              `json:"isBare"`
	Description string            `json:"description,omitempty"`
}

type Refs struct {
	Branches []BranchRef `json:"branches"`
	Tags     []TagRef    `json:"tags"`
}

type BranchRef struct {
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	IsRemote bool   `json:"isRemote"`
	Upstream string `json:"upstream,omitempty"`
}

type TagRef struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	TargetHash  string `json:"targetHash"`
	IsAnnotated bool   `json:"isAnnotated"`
	Tagger     string `json:"tagger,omitempty"`
	Message    string `json:"message,omitempty"`
}

type CommitNode struct {
	Hash        string       `json:"hash"`
	ShortHash   string       `json:"shortHash"`
	Author      Signature    `json:"author"`
	Committer   Signature    `json:"committer"`
	Message     string       `json:"message"`
	MessageBody string       `json:"messageBody,omitempty"`
	Parents     []string     `json:"parents"`
	TreeHash    string       `json:"treeHash"`
	Files       []FileDiff   `json:"files,omitempty"`
	SecretHits  int          `json:"secretHits"`
	IsMerge     bool         `json:"isMerge"`
	IsOrphan    bool         `json:"isOrphan"`
	ReachableBy []string     `json:"reachableBy,omitempty"`
}

type Signature struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	When  time.Time `json:"when"`
}

type FileDiff struct {
	Path      string     `json:"path"`
	OldPath   string     `json:"oldPath,omitempty"`
	Action    DiffAction `json:"action"`
	IsBinary  bool       `json:"isBinary"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	Patch     string     `json:"patch,omitempty"` // Unified diff content
}

type DiffAction string

const (
	ActionAdd    DiffAction = "add"
	ActionModify DiffAction = "modify"
	ActionDelete DiffAction = "delete"
	ActionRename DiffAction = "rename"
	ActionCopy   DiffAction = "copy"
)

type StashEntry struct {
	Index   int        `json:"index"`
	Hash    string     `json:"hash"`
	Message string     `json:"message"`
	Author  Signature  `json:"author"`
	Files   []FileDiff `json:"files,omitempty"`
}

type SecretMatch struct {
	Type       SecretType `json:"type"`
	Pattern    string     `json:"pattern"`
	Value      string     `json:"value"`
	File       string     `json:"file"`
	Line       int        `json:"line"`
	CommitHash string     `json:"commitHash"`
	Entropy    float64    `json:"entropy,omitempty"`
	Severity   Severity   `json:"severity"`
}

type SecretType string

const (
	SecretAWSKey        SecretType = "aws_key"
	SecretAWSSecret     SecretType = "aws_secret"
	SecretGitHubToken   SecretType = "github_token"
	SecretGitHubPAT     SecretType = "github_pat"
	SecretSlackToken    SecretType = "slack_token"
	SecretSlackWebhook  SecretType = "slack_webhook"
	SecretPrivateKey    SecretType = "private_key"
	SecretGCPKey        SecretType = "gcp_key"
	SecretAzureKey      SecretType = "azure_key"
	SecretJWT           SecretType = "jwt"
	SecretGenericAPI    SecretType = "generic_api"
	SecretPassword      SecretType = "password"
	SecretHighEntropy   SecretType = "high_entropy"
	SecretDatabaseURL   SecretType = "database_url"
	SecretSSHKey        SecretType = "ssh_key"
	SecretTelegramToken SecretType = "telegram_token"
	SecretDiscordToken  SecretType = "discord_token"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type ReconStats struct {
	TotalCommits    int `json:"totalCommits"`
	OrphanCommits  int `json:"orphanCommits"`
	TotalBranches  int `json:"totalBranches"`
	TotalTags      int `json:"totalTags"`
	TotalStashes   int `json:"totalStashes"`
	TotalSecrets   int `json:"totalSecrets"`
	CriticalSecrets int `json:"criticalSecrets"`
	FilesScanned   int `json:"filesScanned"`
	LinesScanned   int `json:"linesScanned"`
}
