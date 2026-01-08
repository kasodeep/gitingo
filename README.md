# Git (in Go)

## Implementation

### Commands

- The package contains detail about how each command is executed and the respective methods.

#### Init

1. repoRoot represented the current directory where we create the .git equivalent.
2. It checks for the repository to be initialized already and prints the error.
3. Then it created the .git folder with all the necessary folders and files.

#### Add

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

#### Status

#### Commit

Parse index
    ↓
Build in-memory tree hierarchy
    ↓
Serialize trees (binary)
    ↓
Write tree objects
    ↓
Create commit object