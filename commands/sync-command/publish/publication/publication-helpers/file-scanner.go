package publication_helpers

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileInfo holds the necessary data for sorting and processing.
type FileInfo struct {
	Path  string
	Mtime time.Time // Modification Time (Used for sorting)
	Size  int64     // Size of the file in bytes
}

/*
FindNextFilesInFolder find the next file that satisfies all criteria;
- it's a file (not a folder)
- correct file extension (e.g. XML)
- not empty
- not too large
- not written in the last X seconds (we think that the write is complete)
*/
func FindNextFilesInFolder(dirPath string, fileExtension string, maxSizeBytes int64, minAgeMillis time.Duration) (fileInfo []FileInfo, error error) {
	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	candidateFiles := make([]FileInfo, 0)
	now := time.Now()

	for _, entry := range entries {

		filePath := filepath.Join(dirPath, entry.Name())
		info, err := getEntryInfo(entry, filePath)
		if err != nil {
			log.Printf("Warning: Skipping file %s due to stat error: %v", entry.Name(), err)
			continue
		}

		// Skip folders
		if info.IsDir() {
			continue
		}

		// Skip hidden files like ".notYetComplete.xml.part"
		if isIgnoredFile(entry.Name()) {
			continue
		}

		// Skip empty files
		if info.Size() == 0 {
			log.Printf("Skipping file %s: File is empty.", entry.Name())
			continue
		}

		// Filter out files that have been recently written - and possibly are still being written.
		age := now.Sub(info.ModTime())
		if age < minAgeMillis {
			continue
		}

		// Filter out files that are too large
		if info.Size() > maxSizeBytes {
			log.Printf("Skipping file %s: Size (%d bytes) exceeds maximum allowed size (%d bytes).",
				entry.Name(), info.Size(), maxSizeBytes)
			continue
		}

		// Filter out any file extensions that don't match the published record
		if !strings.HasSuffix(strings.ToLower(entry.Name()), fileExtension) {
			log.Printf("Skipping file %s: Does not match required extension %s.", entry.Name(), fileExtension)
			continue
		}

		// Filter out any files that have companion lock files
		if hasLockFile(entry.Name(), entries) {
			continue
		}

		// File passes all checks - add it to the candidates
		candidateFiles = append(candidateFiles, FileInfo{
			Path:  filePath,
			Mtime: info.ModTime(),
			Size:  info.Size(),
		})
	}

	// order by oldest first
	sort.Slice(candidateFiles, func(i, j int) bool {
		return candidateFiles[i].Mtime.Before(candidateFiles[j].Mtime)
	})

	return candidateFiles, nil
}

// isIgnoredFile checks for common files that should not be processed (e.g., system files).
func isIgnoredFile(name string) bool {
	// Skip files starting with a dot (hidden files, .DS_Store, etc.)
	if len(name) > 0 && name[0] == '.' {
		return true
	}

	// Skip common lock and backup file patterns (~, #)
	if strings.HasPrefix(name, "~") || strings.HasPrefix(name, "#") {
		return true
	}

	// Skip Microsoft Office temporary/lock files
	if strings.HasPrefix(name, "~$") {
		return true
	}

	return false
}

// hasLockFile checks for any files with common companion lock files
func hasLockFile(name string, entries []os.DirEntry) bool {

	// Suffixes: Lock file is named {original_filename}{suffix}
	lockSuffixes := []string{".lock", ".tmp", ".lck"}

	// Prefixes: Lock file is named {prefix}{original_filename}
	lockPrefixes := []string{"~", "#", "~$"}

	//search the dir listing again, looking for lock files
	for _, entry := range entries {
		lockName := entry.Name()

		// A lock file must be a file, not a directory.
		if entry.IsDir() {
			continue
		}

		// Check for Suffix Companion Lock: {target_name}{suffix}
		for _, suffix := range lockSuffixes {
			expectedLockName := name + suffix
			if lockName == expectedLockName {
				return true
			}
		}

		// Check for Prefix Companion Lock: {prefix}{target_name}
		for _, prefix := range lockPrefixes {
			expectedLockName := prefix + name
			if lockName == expectedLockName {
				return true
			}
		}
	}

	return false
}

// getEntryInfo tries to get fs.FileInfo from a directory entry.
// This handles cases where the entry's info might not be readily available
// (permissions and locking can sometimes cause this).
func getEntryInfo(entry fs.DirEntry, filePath string) (fs.FileInfo, error) {
	info, err := entry.Info()
	if err == nil {
		return info, nil
	}
	// Fallback to os.Stat if DirEntry.Info() failed
	return os.Stat(filePath)
}
