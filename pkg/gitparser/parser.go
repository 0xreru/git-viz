package gitparser

import (
	"fmt"
	"sort"
	"strings"

	"git-recon-viz/pkg/types"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type ParseOptions struct {
	IncludeDiffs bool
	MaxCommits   int
	Verbose      bool
}

type ParseResult struct {
	Data         *types.ReconData
	ReachableSet map[plumbing.Hash]struct{}
}

func Parse(gitDir string, opts ParseOptions) (*ParseResult, error) {
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		repo, err = git.PlainOpen(strings.TrimSuffix(gitDir, "/.git"))
		if err != nil {
			return nil, fmt.Errorf("failed to open repository: %w", err)
		}
	}

	result := &ParseResult{
		Data: &types.ReconData{
			Refs:    types.Refs{},
			Commits: make([]types.CommitNode, 0),
			Stashes: make([]types.StashEntry, 0),
		},
		ReachableSet: make(map[plumbing.Hash]struct{}),
	}

	if err := extractMeta(repo, gitDir, result.Data); err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	refHeads, err := extractRefs(repo, result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract refs: %w", err)
	}

	if err := walkCommits(repo, refHeads, result, opts); err != nil {
		return nil, fmt.Errorf("failed to walk commits: %w", err)
	}

	if err := extractStashes(repo, result.Data, opts); err != nil {
		if opts.Verbose {
			fmt.Printf("[-] Stash extraction failed: %v\n", err)
		}
	}

	result.Data.Stats.TotalCommits = len(result.Data.Commits)
	result.Data.Stats.TotalBranches = len(result.Data.Refs.Branches)
	result.Data.Stats.TotalTags = len(result.Data.Refs.Tags)
	result.Data.Stats.TotalStashes = len(result.Data.Stashes)

	return result, nil
}

func extractMeta(repo *git.Repository, gitDir string, data *types.ReconData) error {
	data.Meta.Path = gitDir

	cfg, err := repo.Config()
	if err == nil {
		data.Meta.IsBare = cfg.Core.IsBare
	}

	head, err := repo.Head()
	if err == nil {
		data.Meta.HeadRef = head.Name().String()
	}

	data.Meta.Remotes = make(map[string]string)
	remotes, err := repo.Remotes()
	if err == nil {
		for _, r := range remotes {
			cfg := r.Config()
			if len(cfg.URLs) > 0 {
				data.Meta.Remotes[cfg.Name] = cfg.URLs[0]
			}
		}
	}

	return nil
}

func extractRefs(repo *git.Repository, data *types.ReconData) ([]plumbing.Hash, error) {
	var commitHeads []plumbing.Hash

	refs, err := repo.References()
	if err != nil {
		return nil, err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().String()
		hash := ref.Hash()

	if ref.Type() == plumbing.SymbolicReference {
			return nil
		}

		switch {
		case ref.Name().IsBranch():
			branch := types.BranchRef{
				Name:     ref.Name().Short(),
				Hash:     hash.String(),
				IsRemote: false,
			}
			data.Refs.Branches = append(data.Refs.Branches, branch)
			commitHeads = append(commitHeads, hash)

		case ref.Name().IsRemote():
			branch := types.BranchRef{
				Name:     ref.Name().Short(),
				Hash:     hash.String(),
				IsRemote: true,
			}
			data.Refs.Branches = append(data.Refs.Branches, branch)
			commitHeads = append(commitHeads, hash)

		case ref.Name().IsTag():
			tag := types.TagRef{
				Name: ref.Name().Short(),
				Hash: hash.String(),
			}

			// Check if annotated tag
			tagObj, err := repo.TagObject(hash)
			if err == nil {
				tag.IsAnnotated = true
				tag.TargetHash = tagObj.Target.String()
				tag.Message = tagObj.Message
				tag.Tagger = tagObj.Tagger.Name
				commitHeads = append(commitHeads, tagObj.Target)
			} else {
				tag.TargetHash = hash.String()
				commitHeads = append(commitHeads, hash)
			}
			data.Refs.Tags = append(data.Refs.Tags, tag)

		case strings.HasPrefix(name, "refs/stash"):
			commitHeads = append(commitHeads, hash)

		default:
			commitHeads = append(commitHeads, hash)
		}

		return nil
	})

	return commitHeads, err
}

func walkCommits(repo *git.Repository, heads []plumbing.Hash, result *ParseResult, opts ParseOptions) error {
	headSet := make(map[plumbing.Hash]struct{})
	for _, h := range heads {
		headSet[h] = struct{}{}
	}

	commitIter, err := repo.Log(&git.LogOptions{
		All:   true,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return walkCommitsManual(repo, heads, result, opts)
	}
	defer commitIter.Close()

	commitCount := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		if opts.MaxCommits > 0 && commitCount >= opts.MaxCommits {
			return storer.ErrStop
		}

		result.ReachableSet[c.Hash] = struct{}{}

		node := commitToNode(c)
		node.ReachableBy = findReachableBy(c.Hash, result.Data)

		if opts.IncludeDiffs {
			files, err := extractDiff(c)
			if err == nil {
				node.Files = files
			}
		}

		result.Data.Commits = append(result.Data.Commits, node)
		commitCount++
		return nil
	})

	return err
}

func walkCommitsManual(repo *git.Repository, heads []plumbing.Hash, result *ParseResult, opts ParseOptions) error {
	visited := make(map[plumbing.Hash]struct{})
	queue := make([]plumbing.Hash, 0, len(heads))
	queue = append(queue, heads...)

	commitCount := 0

	for len(queue) > 0 {
		if opts.MaxCommits > 0 && commitCount >= opts.MaxCommits {
			break
		}

		hash := queue[0]
		queue = queue[1:]

		if _, seen := visited[hash]; seen {
			continue
		}
		visited[hash] = struct{}{}
		result.ReachableSet[hash] = struct{}{}

		commit, err := repo.CommitObject(hash)
		if err != nil {
			continue
		}

		node := commitToNode(commit)
		node.ReachableBy = findReachableBy(hash, result.Data)

		if opts.IncludeDiffs {
			files, err := extractDiff(commit)
			if err == nil {
				node.Files = files
			}
		}

		result.Data.Commits = append(result.Data.Commits, node)
		commitCount++

		for _, parent := range commit.ParentHashes {
			if _, seen := visited[parent]; !seen {
				queue = append(queue, parent)
			}
		}
	}
	return nil
}

func commitToNode(c *object.Commit) types.CommitNode {
	node := types.CommitNode{
		Hash:      c.Hash.String(),
		ShortHash: c.Hash.String()[:7],
		Author: types.Signature{
			Name:  c.Author.Name,
			Email: c.Author.Email,
			When:  c.Author.When,
		},
		Committer: types.Signature{
			Name:  c.Committer.Name,
			Email: c.Committer.Email,
			When:  c.Committer.When,
		},
		TreeHash: c.TreeHash.String(),
		IsMerge:  len(c.ParentHashes) > 1,
		IsOrphan: false,
		Parents:  make([]string, 0, len(c.ParentHashes)),
	}

	// Split message into subject and body
	lines := strings.SplitN(c.Message, "\n", 2)
	node.Message = strings.TrimSpace(lines[0])
	if len(lines) > 1 {
		node.MessageBody = strings.TrimSpace(lines[1])
	}

	for _, p := range c.ParentHashes {
		node.Parents = append(node.Parents, p.String())
	}

	return node
}

func findReachableBy(hash plumbing.Hash, data *types.ReconData) []string {
	return nil
}

func extractDiff(c *object.Commit) ([]types.FileDiff, error) {
	var files []types.FileDiff

	var parentTree *object.Tree
	if c.NumParents() > 0 {
		parent, err := c.Parent(0)
		if err == nil {
			parentTree, _ = parent.Tree()
		}
	}

	currentTree, err := c.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := object.DiffTree(parentTree, currentTree)
	if err != nil {
		return nil, err
	}

	for _, change := range changes {
		fd := types.FileDiff{}

		from, to, err := change.Files()
		if err != nil {
			continue
		}

		switch {
		case from == nil && to != nil:
			fd.Action = types.ActionAdd
			fd.Path = to.Name
		case from != nil && to == nil:
			fd.Action = types.ActionDelete
			fd.Path = from.Name
		case from != nil && to != nil:
			if from.Name != to.Name {
				fd.Action = types.ActionRename
				fd.OldPath = from.Name
				fd.Path = to.Name
			} else {
				fd.Action = types.ActionModify
				fd.Path = to.Name
			}
		}

		if from != nil && isBinaryMode(from.Mode) || to != nil && isBinaryMode(to.Mode) {
			fd.IsBinary = true
		}

		if !fd.IsBinary {
			patch, err := change.Patch()
			if err == nil {
				fd.Patch = patch.String()
				// Count additions/deletions
				for _, fp := range patch.FilePatches() {
					for _, chunk := range fp.Chunks() {
						lines := strings.Split(chunk.Content(), "\n")
						switch chunk.Type() {
						case 1: // Add
							fd.Additions += len(lines)
						case 2: // Delete
							fd.Deletions += len(lines)
						}
					}
				}
			}
		}

		files = append(files, fd)
	}

	return files, nil
}

func isBinaryMode(mode filemode.FileMode) bool {
	return !mode.IsRegular()
}

func extractStashes(repo *git.Repository, data *types.ReconData, opts ParseOptions) error {
	stashRef, err := repo.Reference(plumbing.ReferenceName("refs/stash"), true)
	if err != nil {
		return nil
	}

	stashCommit, err := repo.CommitObject(stashRef.Hash())
	if err != nil {
		return err
	}

	entry := types.StashEntry{
		Index:   0,
		Hash:    stashCommit.Hash.String(),
		Message: stashCommit.Message,
		Author: types.Signature{
			Name:  stashCommit.Author.Name,
			Email: stashCommit.Author.Email,
			When:  stashCommit.Author.When,
		},
	}

	if opts.IncludeDiffs {
		files, err := extractDiff(stashCommit)
		if err == nil {
			entry.Files = files
		}
	}

	data.Stashes = append(data.Stashes, entry)
	return nil
}

func SortCommitsByTime(commits []types.CommitNode) {
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Committer.When.After(commits[j].Committer.When)
	})
}
