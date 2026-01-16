package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

/*
It prepares the object by appending the <entry> -> <type> <size>\0<content>
In recursive format the obj will have <type> <size>\0<entry><entry>
*/
func PrepareObject(objType string, content []byte) ([]byte, string) {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	full := append([]byte(header), content...)

	sum := sha256.Sum256(full)
	return full, hex.EncodeToString(sum[:])
}

/*
It uses prepare object to get the required file content.
Then writes to objects folder with hash[:2]/hash[2:] format.
*/
func WriteObject(gitDir, objType string, content []byte) string {
	full, hash := PrepareObject(objType, content)

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

func VerifyObject(gitDir string, hash string, obj string) error {
	if len(hash) < 6 {
		return fmt.Errorf("hash too short")
	}

	objPath := filepath.Join(gitDir, "objects", hash[:2], hash[2:])
	data, err := os.ReadFile(objPath)
	if err != nil {
		return fmt.Errorf("object not found: %s", hash)
	}

	if !strings.HasPrefix(string(data), obj) {
		return fmt.Errorf("object is not a type of %s", obj)
	}

	return nil
}

/*
The methods opens the file stat, maps them to git modes for specificity.
Then, it returns the mode, content read, along with a bool indicator.
*/
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

/*
Checks if the given path is a directory or not.
*/
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
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
