package sort

import (
	"io/fs"
	"strings"
)

// FoldersFirst is a custom sorter for directories to appear first in a list of
// files.
type FoldersFirst []fs.DirEntry

// Len implement the sort.Interface for FoldersFirst
func (f FoldersFirst) Len() int { return len(f) }

// Swap implement the sort.Interface for FoldersFirst
func (f FoldersFirst) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// Less implement the sort.Interface for FoldersFirst
func (f FoldersFirst) Less(i, j int) bool {
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
