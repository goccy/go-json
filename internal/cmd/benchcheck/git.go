package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// defaultBaseRefs are tried in order when the base ref is not specified.
// The remote-tracking branch is for environments without the local branch, such as CI.
var defaultBaseRefs = []string{"master", "origin/master"}

type repository struct {
	root string
}

func (r *repository) git(ctx context.Context, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.root
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func openRepository(ctx context.Context) (*repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	root, err := (&repository{root: wd}).git(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	return &repository{root: root}, nil
}

func (r *repository) commit(ctx context.Context, ref string) (string, error) {
	return r.git(ctx, "rev-parse", "--verify", ref+"^{commit}")
}

// baseCommit resolves the commit to compare with.
func (r *repository) baseCommit(ctx context.Context, ref string) (string, error) {
	if ref != "" {
		return r.commit(ctx, ref)
	}
	for _, candidate := range defaultBaseRefs {
		if commit, err := r.commit(ctx, candidate); err == nil {
			return commit, nil
		}
	}
	return "", fmt.Errorf("failed to find the base branch: tried %s", strings.Join(defaultBaseRefs, ", "))
}

// dirty reports whether the working tree has changes which are not committed yet.
func (r *repository) dirty(ctx context.Context) (bool, error) {
	out, err := r.git(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// defaultCacheDir is located under the common git directory
// so that it is shared by all worktrees and never gets committed.
func (r *repository) defaultCacheDir(ctx context.Context) (string, error) {
	commonDir, err := r.git(ctx, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(r.root, commonDir)
	}
	return filepath.Join(commonDir, "benchcheck"), nil
}

// checkout creates a temporary worktree of the commit at dir and returns the function to remove it.
func (r *repository) checkout(ctx context.Context, dir, commit string) (func(), error) {
	if _, err := r.git(ctx, "worktree", "add", "--detach", dir, commit); err != nil {
		return nil, err
	}
	cleanup := func() {
		// ctx may already be canceled by a signal, but the worktree must be removed anyway.
		if _, err := r.git(context.WithoutCancel(ctx), "worktree", "remove", "--force", dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return cleanup, nil
}
