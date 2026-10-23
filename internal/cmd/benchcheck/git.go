package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
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
	out, err := r.gitOutput(ctx, args...)
	return strings.TrimSpace(out), err
}

// gitOutput returns the output of the git command as it is.
func (r *repository) gitOutput(ctx context.Context, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.root
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
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

// addWorktree creates a temporary worktree of the commit at dir and returns the function to remove it.
func (r *repository) addWorktree(ctx context.Context, dir, commit string) (func(), error) {
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

// copyWorkingTree copies the files of the working tree to dir, with the changes which are not committed yet: the
// tracked files, as a worktree of a commit has them, and the untracked ones which are not ignored, except the ones
// which the go command doesn't read ( see ignoredByGo ). The ignored files are not copied, as a worktree of the base
// doesn't have them either.
func (r *repository) copyWorkingTree(ctx context.Context, dir string) error {
	tracked, err := r.gitOutput(ctx, "ls-files", "-z", "--cached")
	if err != nil {
		return err
	}
	untracked, err := r.gitOutput(ctx, "ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return err
	}
	names := strings.Split(tracked, "\x00")
	for _, name := range strings.Split(untracked, "\x00") {
		if !ignoredByGo(name) {
			names = append(names, name)
		}
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		src := filepath.Join(r.root, filepath.FromSlash(name))
		dst := filepath.Join(dir, filepath.FromSlash(name))
		info, err := os.Lstat(src)
		if errors.Is(err, fs.ErrNotExist) {
			// a tracked file which is removed and not committed yet.
			continue
		}
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		switch {
		case info.Mode().IsRegular():
			err = copyFile(dst, src)
		case info.Mode()&fs.ModeSymlink != 0:
			var target string
			target, err = os.Readlink(src)
			if err == nil {
				err = os.Symlink(target, dst)
			}
		default:
			// a directory: a repository inside the working tree, untracked or a submodule, whose files a worktree
			// of a commit doesn't have either.
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ignoredByGo reports whether the go command ignores the file at the slash-separated path name when it finds
// packages: "Directory and file names that begin with "." or "_" are ignored by the go tool" ( go help packages ).
// A working tree may have such directories of untracked scratch files, as large as the rest of it many times over,
// which are not copied. An untracked file there which a benchmark reads, or which //go:embed names, is not copied
// either: it has to be tracked.
func ignoredByGo(name string) bool {
	for _, elem := range strings.Split(name, "/") {
		if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") {
			return true
		}
	}
	return false
}
