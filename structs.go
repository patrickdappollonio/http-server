package main

import (
	"os"
)

type foldersFirst []os.FileInfo

func (f foldersFirst) Len() int      { return len(f) }
func (f foldersFirst) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f foldersFirst) Less(i, j int) bool {
	if f[i].IsDir() && f[j].IsDir() {
		return f[i].Name() < f[j].Name()
	}

	if !f[i].IsDir() && f[j].IsDir() {
		return false
	}

	if f[i].IsDir() && !f[j].IsDir() {
		return true
	}

	return f[i].Name() < f[j].Name()
}
