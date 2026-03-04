package scanner

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"git-recon-viz/pkg/types"
)

const maxMatchLength = 100

type secretRule struct {
	Type     types.SecretType
	Pattern  *regexp.Regexp
	Severity types.Severity
	Name     string
}

var (
	rules       []secretRule
	rulesMu     sync.RWMutex
	initialized bool
)

func Init(customFlagPattern string) error {
	rulesMu.Lock()
	defer rulesMu.Unlock()

	if !initialized {
		compileBasePatterns()
		initialized = true
	}

	if customFlagPattern != "" {
		return addCustomPatternLocked(customFlagPattern)
	}
	return nil
}

func compileBasePatterns() {
	patterns := []struct {
		regex    string
		typ      types.SecretType
		severity types.Severity
		name     string
	}{

		// === AWS ===
		{`AKIA[0-9A-Z]{16}`, types.SecretAWSKey, types.SeverityCritical, "AWS Access Key ID"},
		{`(?i)aws[_\-]?secret[_\-]?access[_\-]?key\s*[:=]\s*['"]?([A-Za-z0-9/+=]{40})['"]?`, types.SecretAWSSecret, types.SeverityCritical, "AWS Secret Access Key"},

		// === Private Keys ===
		{`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`, types.SecretPrivateKey, types.SeverityCritical, "Private Key Header"},
		{`-----BEGIN PGP PRIVATE KEY BLOCK-----`, types.SecretPrivateKey, types.SeverityCritical, "PGP Private Key"},
		{`-----BEGIN CERTIFICATE-----`, types.SecretSSHKey, types.SeverityMedium, "Certificate"},

		// === GitHub ===
		{`ghp_[0-9a-zA-Z]{36}`, types.SecretGitHubPAT, types.SeverityCritical, "GitHub PAT (Fine-grained)"},
		{`github_pat_[0-9a-zA-Z]{22}_[0-9a-zA-Z]{59}`, types.SecretGitHubPAT, types.SeverityCritical, "GitHub PAT (Classic)"},
		{`gho_[0-9a-zA-Z]{36}`, types.SecretGitHubToken, types.SeverityHigh, "GitHub OAuth Token"},
		{`ghu_[0-9a-zA-Z]{36}`, types.SecretGitHubToken, types.SeverityHigh, "GitHub User Token"},
		{`ghr_[0-9a-zA-Z]{36}`, types.SecretGitHubToken, types.SeverityHigh, "GitHub Refresh Token"},
		{`ghs_[0-9a-zA-Z]{36}`, types.SecretGitHubToken, types.SeverityHigh, "GitHub Server Token"},

		// === GitLab ===
		{`glpat-[0-9a-zA-Z\-_]{20,}`, types.SecretGenericAPI, types.SeverityCritical, "GitLab PAT"},
		{`glrt-[0-9a-zA-Z\-_]{20,}`, types.SecretGenericAPI, types.SeverityHigh, "GitLab Runner Token"},

		// === Slack ===
		{`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*`, types.SecretSlackToken, types.SeverityCritical, "Slack Token"},
		{`https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[a-zA-Z0-9]+`, types.SecretSlackWebhook, types.SeverityHigh, "Slack Webhook"},

		// === Discord ===
		{`(?i)discord[_\-]?(?:token|webhook)\s*[:=]\s*['"]?([A-Za-z0-9._\-]+)['"]?`, types.SecretDiscordToken, types.SeverityHigh, "Discord Token/Webhook"},
		{`https://discord(?:app)?\.com/api/webhooks/[0-9]+/[A-Za-z0-9_\-]+`, types.SecretDiscordToken, types.SeverityHigh, "Discord Webhook URL"},

		// === Telegram ===
		{`[0-9]+:AA[0-9A-Za-z\-_]{33}`, types.SecretTelegramToken, types.SeverityHigh, "Telegram Bot Token"},

		// === GCP ===
		{`AIza[0-9A-Za-z\-_]{35}`, types.SecretGCPKey, types.SeverityCritical, "Google API Key"},
		{`(?i)type["']?\s*:\s*["']?service_account`, types.SecretGCPKey, types.SeverityHigh, "GCP Service Account JSON"},

		// === Azure ===
		{`(?i)azure[_\-]?(?:client|tenant|subscription)[_\-]?(?:id|secret)\s*[:=]\s*['"]?([a-f0-9\-]{36})['"]?`, types.SecretAzureKey, types.SeverityCritical, "Azure Credential"},

		// === JWT ===
		{`eyJ[A-Za-z0-9\-_]+\.eyJ[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_.+/=]*`, types.SecretJWT, types.SeverityMedium, "JWT Token"},

		// === Database URLs ===
		{`(?i)(?:mysql|postgres|postgresql|mongodb|redis|memcached)://[^\s'"]+`, types.SecretDatabaseURL, types.SeverityCritical, "Database Connection URL"},
		{`(?i)(?:jdbc|odbc):[^\s'"]+`, types.SecretDatabaseURL, types.SeverityHigh, "JDBC/ODBC Connection"},

		// === Generic Credentials (catch-all) ===
		{`(?i)(?:password|passwd|pwd)\s*[:=]\s*['"]?([^\s'"]{8,64})['"]?`, types.SecretPassword, types.SeverityHigh, "Password Assignment"},
		{`(?i)(?:api[_\-]?key|apikey)\s*[:=]\s*['"]?([a-zA-Z0-9\-_]{16,64})['"]?`, types.SecretGenericAPI, types.SeverityHigh, "API Key"},
		{`(?i)(?:secret|token|auth)[_\-]?(?:key|token)?\s*[:=]\s*['"]?([a-zA-Z0-9\-_]{16,64})['"]?`, types.SecretGenericAPI, types.SeverityHigh, "Generic Secret/Token"},
		{`(?i)(?:access[_\-]?token)\s*[:=]\s*['"]?([a-zA-Z0-9\-_.]{16,128})['"]?`, types.SecretGenericAPI, types.SeverityHigh, "Access Token"},
		{`(?i)bearer\s+[a-zA-Z0-9\-_.~+/]+=*`, types.SecretGenericAPI, types.SeverityMedium, "Bearer Token"},

		// === Stripe ===
		{`sk_live_[0-9a-zA-Z]{24,}`, types.SecretGenericAPI, types.SeverityCritical, "Stripe Live Secret Key"},
		{`sk_test_[0-9a-zA-Z]{24,}`, types.SecretGenericAPI, types.SeverityMedium, "Stripe Test Secret Key"},
		{`pk_live_[0-9a-zA-Z]{24,}`, types.SecretGenericAPI, types.SeverityLow, "Stripe Live Publishable Key"},
		{`rk_live_[0-9a-zA-Z]{24,}`, types.SecretGenericAPI, types.SeverityCritical, "Stripe Live Restricted Key"},

		// === Twilio ===
		{`SK[0-9a-fA-F]{32}`, types.SecretGenericAPI, types.SeverityHigh, "Twilio API Key"},
		{`AC[a-z0-9]{32}`, types.SecretGenericAPI, types.SeverityMedium, "Twilio Account SID"},

		// === SendGrid ===
		{`SG\.[a-zA-Z0-9]{22}\.[a-zA-Z0-9\-_]{43}`, types.SecretGenericAPI, types.SeverityCritical, "SendGrid API Key"},

		// === Mailgun ===
		{`key-[0-9a-zA-Z]{32}`, types.SecretGenericAPI, types.SeverityHigh, "Mailgun API Key"},

		// === Heroku ===
		{`(?i)heroku[_\-]?api[_\-]?key\s*[:=]\s*['"]?([a-f0-9\-]{36})['"]?`, types.SecretGenericAPI, types.SeverityHigh, "Heroku API Key"},

		// === NPM ===
		{`npm_[A-Za-z0-9]{36}`, types.SecretGenericAPI, types.SeverityHigh, "NPM Token"},

		// === PyPI ===
		{`pypi-AgEIcHlwaS5vcmc[A-Za-z0-9\-_]{50,}`, types.SecretGenericAPI, types.SeverityHigh, "PyPI Token"},

		// === SSH Passphrase indicators ===
		{`(?i)(?:ssh[_\-]?pass(?:phrase)?|identity[_\-]?file)\s*[:=]\s*['"]?([^\s'"]+)['"]?`, types.SecretSSHKey, types.SeverityMedium, "SSH Config"},
	}

	rules = make([]secretRule, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p.regex)
		if err != nil {
			continue
		}
		rules = append(rules, secretRule{
			Type:     p.typ,
			Pattern:  re,
			Severity: p.severity,
			Name:     p.name,
		})
	}
}

func addCustomPatternLocked(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid custom flag pattern: %w", err)
	}

	rules = append(rules, secretRule{
		Type:     types.SecretGenericAPI,
		Pattern:  re,
		Severity: types.SeverityCritical,
		Name:     "Custom Flag Pattern",
	})

	return nil
}

func AddCustomPattern(pattern, name string, severity types.Severity) error {
	rulesMu.Lock()
	defer rulesMu.Unlock()

	if !initialized {
		compileBasePatterns()
		initialized = true
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	rules = append(rules, secretRule{
		Type:     types.SecretGenericAPI,
		Pattern:  re,
		Severity: severity,
		Name:     name,
	})

	return nil
}

type ScanOptions struct {
	Verbose bool
}

func Scan(data *types.ReconData, opts ScanOptions) {
	rulesMu.RLock()
	defer rulesMu.RUnlock()

	// Ensure initialized (lazy init if Init() wasn't called)
	if !initialized {
		rulesMu.RUnlock()
		Init("")
		rulesMu.RLock()
	}

	for i := range data.Commits {
		hits := scanCommit(&data.Commits[i])
		data.Secrets = append(data.Secrets, hits...)
	}

	for i := range data.Orphans {
		hits := scanCommit(&data.Orphans[i])
		data.Secrets = append(data.Secrets, hits...)
	}

	for i := range data.Stashes {
		hits := scanStash(&data.Stashes[i])
		data.Secrets = append(data.Secrets, hits...)
	}

	data.Stats.TotalSecrets = len(data.Secrets)
	data.Stats.CriticalSecrets = countBySeverity(data.Secrets, types.SeverityCritical)
}

func scanCommit(commit *types.CommitNode) []types.SecretMatch {
	var matches []types.SecretMatch

	for _, file := range commit.Files {
		if file.Patch == "" {
			continue
		}
		fileMatches := scanContent(file.Patch, file.Path, commit.Hash)
		matches = append(matches, fileMatches...)
	}

	msgMatches := scanContent(commit.Message+"\n"+commit.MessageBody, "[commit message]", commit.Hash)
	matches = append(matches, msgMatches...)

	// Update secret hit count on the commit
	commit.SecretHits = len(matches)

	return matches
}

func scanStash(stash *types.StashEntry) []types.SecretMatch {
	var matches []types.SecretMatch

	for _, file := range stash.Files {
		if file.Patch == "" {
			continue
		}
		fileMatches := scanContent(file.Patch, file.Path, stash.Hash)
		matches = append(matches, fileMatches...)
	}

	return matches
}

func scanContent(content, filePath, commitHash string) []types.SecretMatch {
	var matches []types.SecretMatch

	// Skip if content is too short to contain secrets
	if len(content) < 8 {
		return matches
	}

	// Track unique matches to avoid duplicates from overlapping patterns
	seen := make(map[string]struct{})

	lines := strings.Split(content, "\n")

	for _, rule := range rules {
		// Find all matches for this rule
		allMatches := rule.Pattern.FindAllStringIndex(content, -1)

		for _, loc := range allMatches {
			matchStr := content[loc[0]:loc[1]]

			// Dedup by value
			if _, exists := seen[matchStr]; exists {
				continue
			}
			seen[matchStr] = struct{}{}

			// Find line number
			lineNum := findLineNumber(content, loc[0], lines)

	match := types.SecretMatch{
				Type:       rule.Type,
				Pattern:    rule.Name,
				Value:      truncateMatch(matchStr),
				File:       filePath,
				Line:       lineNum,
				CommitHash: commitHash,
				Severity:   rule.Severity,
			}
			matches = append(matches, match)
		}
	}
	return matches
}

func findLineNumber(content string, offset int, lines []string) int {
	if offset >= len(content) {
		return len(lines)
	}

	lineNum := 1
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			lineNum++
		}
	}
	return lineNum
}

func truncateMatch(s string) string {
	// Clean up the match (remove newlines, collapse whitespace)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)

	if len(s) > maxMatchLength {
		return s[:maxMatchLength-3] + "..."
	}
	return s
}

func countBySeverity(secrets []types.SecretMatch, severity types.Severity) int {
	count := 0
	for _, s := range secrets {
		if s.Severity == severity {
			count++
		}
	}
	return count
}

func ScanBlob(content, blobHash string) []types.SecretMatch {
	return scanContent(content, "[dangling blob]", blobHash)
}

func GetRuleCount() int {
	return len(rules)
}
