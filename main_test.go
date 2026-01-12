package main

import (
	"os"
	"testing"

	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

func TestParseTree(t *testing.T) {
	base, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	repo, err := repository.GetRepository(base)
	if err != nil {
		t.Fatal(err)
	}

	hash := "235321f6b34a79fb8f5f9ae9ef06d893514f5b639b3b77ac66d89aa3af62da63"

	treeObj, err := tree.ParseTree(repo, hash)
	if err != nil {
		t.Fatal(err)
	}

	tree.PrintTree(treeObj, "")
}
