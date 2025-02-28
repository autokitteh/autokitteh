package file

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"go.autokitteh.dev/autokitteh/internal/kvstore"
	"go.autokitteh.dev/autokitteh/internal/kvstore/checks"
	"go.autokitteh.dev/autokitteh/internal/kvstore/encoding"
)

var defaultFilenameExtension = "json"

// store is a gokv.store implementation for storing key-value pairs as files.
type store struct {
	// For locking the locks map
	// (no two goroutines may create a lock for a filename that doesn't have a lock yet).
	locksLock *sync.Mutex
	// For locking file access.
	fileLocks         map[string]*sync.RWMutex
	filenameExtension string
	directory         string
	codec             encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s *store) Set(_ context.Context, k string, v any) error {
	if err := checks.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	escapedKey := url.PathEscape(k)

	// Prepare file lock.
	lock := s.prepFileLock(escapedKey)

	filename := escapedKey
	if s.filenameExtension != "" {
		filename += "." + s.filenameExtension
	}
	filePath := filepath.Clean(s.directory + "/" + filename)

	// File lock and file handling.
	lock.Lock()
	defer lock.Unlock()
	return os.WriteFile(filePath, data, 0o600)
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s *store) Get(_ context.Context, k string, v any) (found bool, err error) {
	if err := checks.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	escapedKey := url.PathEscape(k)

	// Prepare file lock.
	lock := s.prepFileLock(escapedKey)

	filename := escapedKey
	if s.filenameExtension != "" {
		filename += "." + s.filenameExtension
	}
	filePath := filepath.Clean(s.directory + "/" + filename)

	// File lock and file handling.
	lock.RLock()
	// Deferring the unlocking would lead to the unmarshalling being done during the lock, which is bad for performance.
	data, err := os.ReadFile(filePath)
	lock.RUnlock()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, s.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (s *store) Delete(_ context.Context, k string) error {
	if err := checks.CheckKey(k); err != nil {
		return err
	}

	escapedKey := url.PathEscape(k)

	// Prepare file lock.
	lock := s.prepFileLock(escapedKey)

	filename := escapedKey
	if s.filenameExtension != "" {
		filename += "." + s.filenameExtension
	}
	filePath := filepath.Clean(s.directory + "/" + filename)

	// File lock and file handling.
	lock.Lock()
	defer lock.Unlock()
	err := os.Remove(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Close closes the store.
// When called, some resources of the store are left for garbage collection.
func (s *store) Close() error {
	s.fileLocks = nil
	return nil
}

// prepFileLock returns an existing file lock or creates a new one
func (s *store) prepFileLock(escapedKey string) *sync.RWMutex {
	s.locksLock.Lock()
	defer s.locksLock.Unlock()

	lock, found := s.fileLocks[escapedKey]
	if !found {
		lock = new(sync.RWMutex)
		s.fileLocks[escapedKey] = lock
	}
	return lock
}

// Options are the options for the Go map store.
type Options struct {
	// The directory in which to store files.
	// Can be absolute or relative.
	// Optional ("gokv" by default).
	Directory string
	// Extension of the filename, e.g. "json".
	// This makes it easier for example to open a file with a text editor that supports syntax highlighting.
	// But it can lead to redundant and/or stale data if you switch the Codec back and forth!
	// You also should make sure to change this when changing the Codec,
	// although it doesn't matter for gokv, but it might be confusing when there's a gob file with a ".json" filename extension.
	// Set to "" to disable.
	// Optional ("json" by default).
	FilenameExtension *string
	// Encoding format.
	// Note: When you change this, you should also change the FilenameExtension if it's not empty ("").
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Directory: "gokv", Codec: encoding.JSON
var DefaultOptions = Options{
	Directory:         "gokv",
	FilenameExtension: &defaultFilenameExtension,
	Codec:             encoding.JSON,
}

// Newstore creates a new Go map store.
//
// You should call the Close() method on the store when you're done working with it.
func NewStore(options Options) (kvstore.Store, error) {
	result := store{}

	// Set default options
	if options.Directory == "" {
		options.Directory = DefaultOptions.Directory
	}
	if options.FilenameExtension == nil {
		options.FilenameExtension = DefaultOptions.FilenameExtension
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	err := os.MkdirAll(options.Directory, 0o700)
	if err != nil {
		return &result, err
	}

	result.directory = options.Directory
	result.locksLock = new(sync.Mutex)
	result.fileLocks = make(map[string]*sync.RWMutex)
	result.filenameExtension = *options.FilenameExtension
	result.codec = options.Codec

	return &result, nil
}
