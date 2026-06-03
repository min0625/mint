// Copyright 2026 The Mint Authors.
package main

import "fmt"

var (
	version = "dev"
	commit  = ""
)

func main() {
	fmt.Printf("Hello, Mint! Version: %s, Commit: %s\n", version, commit)
}
