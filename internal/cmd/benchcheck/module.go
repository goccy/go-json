package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func goCommand(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run go %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// createModFile creates, in tmpDir, a copy of the go.mod of the benchmark module in benchDir
// whose replace directive for the library points to libDir.
// It makes it possible to run the benchmarks of the working tree against the library of another commit.
func createModFile(ctx context.Context, tmpDir, benchDir, libDir string) (string, error) {
	libPath, err := goCommand(ctx, libDir, "list", "-m")
	if err != nil {
		return "", err
	}
	goMod, err := goCommand(ctx, benchDir, "list", "-m", "-f", "{{.GoMod}}")
	if err != nil {
		return "", err
	}
	modFile := filepath.Join(tmpDir, "go.mod")
	if err := copyFile(modFile, goMod); err != nil {
		return "", err
	}
	// the go command looks for go.sum next to the file specified by -modfile.
	// go.sum doesn't exist if the benchmark module has no external dependency.
	goSum := filepath.Join(filepath.Dir(goMod), "go.sum")
	if err := copyFile(filepath.Join(tmpDir, "go.sum"), goSum); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if _, err := goCommand(ctx, benchDir, "mod", "edit", "-replace", libPath+"="+libDir, modFile); err != nil {
		return "", err
	}
	return modFile, nil
}

// hashDir returns the hash of the names and the contents of all files under dir.
func hashDir(dir string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), len(content))
		h.Write(content)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
