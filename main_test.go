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

	hash := "606fd353bffdde26a0b42c7366165970661e123389fa9a3b6467fad0f5eebc79"

	_, err = tree.ParseTree(repo, hash)
	if err != nil {
		t.Fatal(err)
	}
}
