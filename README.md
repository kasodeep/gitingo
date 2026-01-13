# Git (in Go)

## Commands

- The package contains detail about how each command is executed and the respective methods.

### Init

1. repoRoot represented the current directory where we create the .git equivalent.
2. It checks for the repository to be initialized already and prints the error.
3. Then it created the .git folder with all the necessary folders and files.

### Add

- It checks if the repo is initialized or not.
- Then, it parses the index file, which contains details about the files being tracked.
- It uses the `bufio.Scanner` and `fmt.SScanf` to parse the required simple pattern.
- Checking the isAll arg it modifies the index by walking the repoRoot and adding the file.

```go
package main

import (
    "bytes"
    "compress/zlib"
    "fmt"
)

func main() {
    content := []byte("Hello Git!\n")
    header := fmt.Sprintf("blob %d\x00", len(content))
    data := append([]byte(header), content...)

    var buf bytes.Buffer
    zw := zlib.NewWriter(&buf)
    zw.Write(data)
    zw.Close()

    compressed := buf.Bytes()
    fmt.Printf("%x\n", compressed) // This is your packed object
}
```

### Status

Thoughts behind the process

Stage 1:
  - Comparison of the WD and the current idx.
  - We need a way to create the idx of the working directory.
  - Our option is to call AddFromPath which itr and calls addFile.
  - If we change the structure to add a boolean param, one dependency is WriteObject returns the hash.
  - Altering that and see if we can create something reusable.

Stage 2:
  - Now we need to compare the index with the commit.
  - commit comes from branch and open the file.
  - we can get the tree hash from it.
  - our ParseTree func can parse the tree, we need a func to convert that to idx.
  - to convert a tree to idx, we need a func in tree.go since dependency and abstraction.
*/

### Commit

Parse index
    ↓
Build in-memory tree hierarchy
    ↓
Serialize trees (binary)
    ↓
Write tree objects
    ↓
Create commit object