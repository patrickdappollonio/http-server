package main

import (
	"os"
	"strings"
)

type foldersFirst []os.FileInfo

func (f foldersFirst) Len() int      { return len(f) }
func (f foldersFirst) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f foldersFirst) Less(i, j int) bool {
	if f[i].IsDir() && f[j].IsDir() {
		return strings.ToLower(f[i].Name()) < strings.ToLower(f[j].Name())
	}

	if !f[i].IsDir() && f[j].IsDir() {
		return false
	}

	if f[i].IsDir() && !f[j].IsDir() {
		return true
	}

	return strings.ToLower(f[i].Name()) < strings.ToLower(f[j].Name())
}
