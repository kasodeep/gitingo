// Package helper provides low-level object store operations used by every
// other package. Nothing in here should know about branches, commits, or
// the index — it only knows about raw bytes and file paths.
package helper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Object store
// ─────────────────────────────────────────────────────────────────────────────

// PrepareObject wraps content in a git-style header and returns the
// full byte slice and its SHA-256 hex hash.
//
// Format: "<type> <size>\x00<content>"
func PrepareObject(objType string, content []byte) (full []byte, hash string) {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	full = append([]byte(header), content...)
	sum := sha256.Sum256(full)
	return full, hex.EncodeToString(sum[:])
}

// WriteObject writes content to objects/<hash[:2]>/<hash[2:]> and returns
// the hash. Silently deduplicates: if the object already exists it is
// not rewritten.
func WriteObject(gitDir, objType string, content []byte) string {
	full, hash := PrepareObject(objType, content)

	objPath := objectPath(gitDir, hash)
	if _, err := os.Stat(objPath); err == nil {
		return hash // already stored
	}

	_ = os.MkdirAll(filepath.Dir(objPath), 0755)
	_ = os.WriteFile(objPath, full, 0644)
	return hash
}

// ReadObject reads an object by hash and returns the content after the
// null-byte header separator. Returns (nil, false) if the object is missing.
func ReadObject(gitDir, hash string) ([]byte, bool) {
	data, err := os.ReadFile(objectPath(gitDir, hash))
	if err != nil {
		return nil, false
	}
	_, content, ok := bytes.Cut(data, []byte{0})
	return content, ok
}

// VerifyObject checks that an object exists and starts with the expected
// type prefix. Returns an error if not found or type mismatches.
func VerifyObject(gitDir, hash, objType string) error {
	if len(hash) < 6 {
		return fmt.Errorf("hash too short")
	}
	data, err := os.ReadFile(objectPath(gitDir, hash))
	if err != nil {
		return fmt.Errorf("object not found: %s", hash)
	}
	if !strings.HasPrefix(string(data), objType) {
		return fmt.Errorf("object %s is not of type %s", hash, objType)
	}
	return nil
}

// objectPath returns the filesystem path for an object hash.
func objectPath(gitDir, hash string) string {
	return filepath.Join(gitDir, "objects", hash[:2], hash[2:])
}

// ─────────────────────────────────────────────────────────────────────────────
// File helpers
// ─────────────────────────────────────────────────────────────────────────────

// ReadFileContent reads a file and returns its git mode, raw content, and
// whether the read succeeded. Symlinks are read as their target path.
func ReadFileContent(path string) (mode string, content []byte, ok bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", nil, false
	}

	mode = gitMode(info)

	if mode == "120000" {
		target, err := os.Readlink(path)
		if err != nil {
			return "", nil, false
		}
		return mode, []byte(target), true
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, false
	}
	return mode, data, true
}

// IsDirectory reports whether path exists and is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// gitMode maps an os.FileInfo to a git file mode string.
func gitMode(info os.FileInfo) string {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return "120000"
	case info.Mode()&0111 != 0:
		return "100755"
	default:
		return "100644"
	}
}
