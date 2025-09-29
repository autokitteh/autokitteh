package common

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_walk(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directory structure:
	// tempDir/
	// ├── file1.txt
	// ├── file2.go
	// ├── .hidden_file
	// ├── _private_file
	// ├── subdir/
	// │   ├── file3.txt
	// │   └── file4.go
	// ├── .hidden_dir/
	// │   └── should_not_be_included.txt
	// └── _private_dir/
	//     └── should_not_be_included.txt

	// Create files and directories
	files := map[string]string{
		"file1.txt":                              "content1",
		"file2.go":                               "package main",
		".hidden_file":                           "hidden content",
		"_private_file":                          "private content",
		"subdir/file3.txt":                       "content3",
		"subdir/file4.go":                        "package sub",
		".hidden_dir/should_not_be_included.txt": "should not be included",
		"_private_dir/should_not_be_included.txt": "should not be included",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(tempDir, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	tests := []struct {
		name     string
		basePath string
		want     map[string][]byte
	}{
		{
			name:     "walk directory with filtering",
			basePath: tempDir,
			want: map[string][]byte{
				"file1.txt":        []byte("content1"),
				"file2.go":         []byte("package main"),
				"subdir/file3.txt": []byte("content3"),
				"subdir/file4.go":  []byte("package sub"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploads := make(map[string][]byte)
			walkFunc := walk(tt.basePath, uploads)

			err := filepath.WalkDir(tt.basePath, walkFunc)
			if err != nil {
				t.Fatalf("walk() error = %v", err)
			}

			if !reflect.DeepEqual(uploads, tt.want) {
				t.Errorf("walk() uploads = %v, want %v", uploads, tt.want)
			}
		})
	}
}

func Test_walk_error_handling(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file that we'll make unreadable
	unreadableFile := filepath.Join(tempDir, "unreadable.txt")
	if err := os.WriteFile(unreadableFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make the file unreadable
	if err := os.Chmod(unreadableFile, 0o000); err != nil {
		t.Fatalf("Failed to make file unreadable: %v", err)
	}

	// Restore permissions after test
	defer func() {
		os.Chmod(unreadableFile, 0o644)
	}()

	uploads := make(map[string][]byte)
	walkFunc := walk(tempDir, uploads)

	err := filepath.WalkDir(tempDir, walkFunc)
	if err == nil {
		t.Error("walk() expected error for unreadable file, got nil")
	}
}

func Test_walk_with_walk_error(t *testing.T) {
	uploads := make(map[string][]byte)
	walkFunc := walk("/nonexistent", uploads)

	// Simulate a walk error by calling the function directly with an error
	err := walkFunc("/nonexistent/path", nil, os.ErrNotExist)
	if err != os.ErrNotExist {
		t.Errorf("walk() should propagate walk errors, got %v, want %v", err, os.ErrNotExist)
	}
}

func Test_walk_hidden_and_private_filtering(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files and directories
	testCases := []struct {
		path      string
		content   string
		shouldAdd bool
	}{
		{"normal.txt", "normal", true},
		{".hidden.txt", "hidden", false},
		{"_private.txt", "private", false},
		{"subdir/normal.txt", "normal", true},
		{"subdir/.hidden.txt", "hidden", false},
		{"subdir/_private.txt", "private", false},
	}

	for _, tc := range testCases {
		fullPath := filepath.Join(tempDir, tc.path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(tc.content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Create hidden and private directories that should be skipped
	hiddenDir := filepath.Join(tempDir, ".hidden_dir")
	privateDir := filepath.Join(tempDir, "_private_dir")

	if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}
	if err := os.MkdirAll(privateDir, 0o755); err != nil {
		t.Fatalf("Failed to create private directory: %v", err)
	}

	// Add files in hidden/private dirs (should not be included)
	if err := os.WriteFile(filepath.Join(hiddenDir, "file.txt"), []byte("hidden dir content"), 0o644); err != nil {
		t.Fatalf("Failed to write file in hidden directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(privateDir, "file.txt"), []byte("private dir content"), 0o644); err != nil {
		t.Fatalf("Failed to write file in private directory: %v", err)
	}

	uploads := make(map[string][]byte)
	walkFunc := walk(tempDir, uploads)

	err := filepath.WalkDir(tempDir, walkFunc)
	if err != nil {
		t.Fatalf("walk() error = %v", err)
	}

	expected := map[string][]byte{
		"normal.txt":        []byte("normal"),
		"subdir/normal.txt": []byte("normal"),
	}

	if !reflect.DeepEqual(uploads, expected) {
		t.Errorf("walk() uploads = %v, want %v", uploads, expected)
	}
}
