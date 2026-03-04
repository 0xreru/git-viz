package gitparser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"git-recon-viz/pkg/types"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type OrphanOptions struct {
	IncludeDiffs bool
	Verbose      bool
}

func FindOrphans(gitDir string, reachable map[plumbing.Hash]struct{}, opts OrphanOptions) ([]types.CommitNode, error) {
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		repo, err = git.PlainOpen(strings.TrimSuffix(gitDir, "/.git"))
		if err != nil {
			return nil, fmt.Errorf("failed to open repository: %w", err)
		}
	}

	orphans := make([]types.CommitNode, 0)

	storer := repo.Storer
	iter, err := storer.IterEncodedObjects(plumbing.CommitObject)
	if err != nil {
		if opts.Verbose {
			fmt.Println("[*] Storer iteration failed, using filesystem scan")
		}
		return findOrphansFilesystem(gitDir, repo, reachable, opts)
	}

	err = iter.ForEach(func(obj plumbing.EncodedObject) error {
		hash := obj.Hash()

		if _, found := reachable[hash]; found {
			return nil
		}

		commit, err := decodeCommit(repo, hash)
		if err != nil {
			if opts.Verbose {
				fmt.Printf("[-] Failed to decode orphan %s: %v\n", hash.String()[:7], err)
			}
			return nil
		}

		node := orphanToNode(repo, commit, opts)

		if opts.IncludeDiffs {
			files, err := extractOrphanDiff(repo, commit)
			if err == nil {
				node.Files = files
			}
		}

		orphans = append(orphans, node)

		if opts.Verbose {
			fmt.Printf("[+] Orphan found: %s | %s | %s\n",
				node.ShortHash,
				node.Author.When.Format("2006-01-02"),
				truncate(node.Message, 50))
		}

		return nil
	})

	if err != nil && err != io.EOF {
		return orphans, err
	}

	packedOrphans, err := findOrphansInPackfiles(gitDir, repo, reachable, opts)
	if err == nil {
		orphans = append(orphans, packedOrphans...)
	}

	orphans = deduplicateOrphans(orphans)
	return orphans, nil
}

func decodeCommit(repo *git.Repository, hash plumbing.Hash) (*object.Commit, error) {
	return repo.CommitObject(hash)
}

func orphanToNode(repo *git.Repository, c *object.Commit, opts OrphanOptions) types.CommitNode {
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
		IsOrphan: true,
		Parents:  make([]string, 0, len(c.ParentHashes)),
	}

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

func extractOrphanDiff(repo *git.Repository, c *object.Commit) ([]types.FileDiff, error) {
	var files []types.FileDiff

	var parentTree *object.Tree
	if c.NumParents() > 0 {
		parentHash := c.ParentHashes[0]
		parent, err := repo.CommitObject(parentHash)
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

		patch, err := change.Patch()
		if err == nil {
			fd.Patch = patch.String()
			for _, fp := range patch.FilePatches() {
				for _, chunk := range fp.Chunks() {
					lines := strings.Split(chunk.Content(), "\n")
					switch chunk.Type() {
					case 1:
						fd.Additions += len(lines)
					case 2:
						fd.Deletions += len(lines)
					}
				}
			}
		}

		files = append(files, fd)
	}

	return files, nil
}

func findOrphansFilesystem(gitDir string, repo *git.Repository, reachable map[plumbing.Hash]struct{}, opts OrphanOptions) ([]types.CommitNode, error) {
	orphans := make([]types.CommitNode, 0)

	objectsDir := filepath.Join(gitDir, "objects")
	if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
		objectsDir = filepath.Join(strings.TrimSuffix(gitDir, ".git"), "objects")
	}

	for i := 0; i < 256; i++ {
		prefix := fmt.Sprintf("%02x", i)
		subdir := filepath.Join(objectsDir, prefix)

		entries, err := os.ReadDir(subdir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// Reconstruct full hash
			hashStr := prefix + entry.Name()
			if len(hashStr) != 40 {
				continue
			}

			hash := plumbing.NewHash(hashStr)

			if _, found := reachable[hash]; found {
				continue
			}

			commit, err := repo.CommitObject(hash)
			if err != nil {
				continue
			}

			node := orphanToNode(repo, commit, opts)

			if opts.IncludeDiffs {
				files, err := extractOrphanDiff(repo, commit)
				if err == nil {
					node.Files = files
				}
			}

			orphans = append(orphans, node)

			if opts.Verbose {
				fmt.Printf("[+] Orphan (loose): %s | %s\n", node.ShortHash, truncate(node.Message, 50))
			}
		}
	}

	return orphans, nil
}

func findOrphansInPackfiles(gitDir string, repo *git.Repository, reachable map[plumbing.Hash]struct{}, opts OrphanOptions) ([]types.CommitNode, error) {
	packDir := filepath.Join(gitDir, "objects", "pack")
	if _, err := os.Stat(packDir); os.IsNotExist(err) {
		return nil, nil
	}
	return nil, nil
}

func deduplicateOrphans(orphans []types.CommitNode) []types.CommitNode {
	seen := make(map[string]struct{})
	result := make([]types.CommitNode, 0, len(orphans))

	for _, o := range orphans {
		if _, exists := seen[o.Hash]; !exists {
			seen[o.Hash] = struct{}{}
			result = append(result, o)
		}
	}

	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func FindDanglingBlobs(gitDir string, opts OrphanOptions) ([]DanglingBlob, error) {
	repo, err := git.PlainOpen(gitDir)
	if err != nil {
		repo, err = git.PlainOpen(strings.TrimSuffix(gitDir, "/.git"))
		if err != nil {
			return nil, err
		}
	}

	referencedBlobs := make(map[plumbing.Hash]struct{})

	treeIter, err := repo.Storer.IterEncodedObjects(plumbing.TreeObject)
	if err != nil {
		return nil, err
	}

	treeIter.ForEach(func(obj plumbing.EncodedObject) error {
		tree, err := object.DecodeTree(repo.Storer, obj)
		if err != nil {
			return nil
		}
		for _, entry := range tree.Entries {
			if entry.Mode.IsFile() {
				referencedBlobs[entry.Hash] = struct{}{}
			}
		}
		return nil
	})

	var dangling []DanglingBlob

	blobIter, err := repo.Storer.IterEncodedObjects(plumbing.BlobObject)
	if err != nil {
		return nil, err
	}

	blobIter.ForEach(func(obj plumbing.EncodedObject) error {
		if _, referenced := referencedBlobs[obj.Hash()]; !referenced {
			blob, err := object.DecodeBlob(obj)
			if err != nil {
				return nil
			}

			reader, err := blob.Reader()
			if err != nil {
				return nil
			}
			defer reader.Close()

			buf := make([]byte, 4096)
			n, _ := reader.Read(buf)

			dangling = append(dangling, DanglingBlob{
				Hash:    obj.Hash().String(),
				Size:    obj.Size(),
				Preview: string(buf[:n]),
			})
		}
		return nil
	})

	return dangling, nil
}

type DanglingBlob struct {
	Hash    string `json:"hash"`
	Size    int64  `json:"size"`
	Preview string `json:"preview"`
}
