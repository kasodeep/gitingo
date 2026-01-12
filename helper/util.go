package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// <type> <size>\0<content>
func WriteObject(gitDir, objType string, content []byte) string {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	full := append([]byte(header), content...)

	sum := sha256.Sum256(full)
	hash := hex.EncodeToString(sum[:])

	objDir := filepath.Join(gitDir, "objects", hash[:2])
	objPath := filepath.Join(objDir, hash[2:])

	// Deduplication
	if _, err := os.Stat(objPath); err == nil {
		return hash
	}

	_ = os.MkdirAll(objDir, 0755)
	_ = os.WriteFile(objPath, full, 0644)

	return hash
}

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

func gitMode(info os.FileInfo) string {
	if info.Mode()&os.ModeSymlink != 0 {
		return "120000"
	}
	if info.Mode()&0111 != 0 {
		return "100755"
	}
	return "100644"
}

func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
