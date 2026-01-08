package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func WriteObject(repoRoot, objType string, content []byte) string {
	// <type> <size>\0<content>
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	full := append([]byte(header), content...)

	sum := sha256.Sum256(full)
	hash := hex.EncodeToString(sum[:])

	objDir := filepath.Join(repoRoot, git_folder, "objects", hash[:2])
	objPath := filepath.Join(objDir, hash[2:])

	// Deduplication
	if _, err := os.Stat(objPath); err == nil {
		return hash
	}

	if err := os.MkdirAll(objDir, 0755); err != nil {
		p.Error(err.Error())
	}

	if err := os.WriteFile(objPath, full, 0644); err != nil {
		p.Error(err.Error())
	}

	return hash
}

func WriteCommitObject(
	repoRoot string,
	treeHash string,
	parentHash string,
	message string,
) string {
	var buf bytes.Buffer

	// Required
	buf.WriteString("tree ")
	buf.WriteString(treeHash)
	buf.WriteByte('\n')

	// Optional parent
	if parentHash != "" {
		buf.WriteString("parent ")
		buf.WriteString(parentHash)
		buf.WriteByte('\n')
	}

	// Minimal author/committer (can improve later)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	buf.WriteString("author gitingo <gitingo@local> ")
	buf.WriteString(timestamp)
	buf.WriteByte('\n')

	buf.WriteString("committer gitingo <gitingo@local> ")
	buf.WriteString(timestamp)
	buf.WriteByte('\n')

	// Blank line before message
	buf.WriteByte('\n')
	buf.WriteString(message)
	buf.WriteByte('\n')

	return WriteObject(repoRoot, "commit", buf.Bytes())
}
