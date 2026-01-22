# Git (in Go)

## Repository

- Provides an abstract interface to deal with repository from a single source of truth.

## Index

- We represent the index file as a IndexEntry with mode, hash and the path leaving the base.
- It performs the function of parsing the idx file, and writing or updating it.

## Commands

- The package contains detail about how each command is executed and the respective methods.

### Init

1. Calls the NewRepository function, to get a new `Repository` struct.
2. Then, initiates the `Create` call, to load the folders, files, refs, and HEAD.

### Add
### Commit

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