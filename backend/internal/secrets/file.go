package secrets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/config"
)

type fileSecrets struct {
	secrets  map[string]map[string]string
	filePath string
	logger   *zap.Logger
}

// NewFileSecrets initializes a (fake but simple and persistent) secrets manager
// for local non-production usage, in the form of a JSON file. The file is read
// only once, when a new client is initialized, and overwritten whenever Set()
// is called. DO NOT STORE REAL SECRETS IN THIS WAY FOR LONG PERIODS OF TIME!
func NewFileSecrets(l *zap.Logger) (Secrets, error) {
	s := &fileSecrets{
		secrets:  map[string]map[string]string{},
		filePath: persistentFilePath(l),
		logger:   l,
	}
	// TODO(ENG-146,ENG-160): This log line screws up CLI tests because it contains a
	// timestamp, use sed/grep to overcome this, and uncomment this log line.
	// l.Warn("Using an insecure local file to manage user secrets",
	// 	zap.String("path", s.filePath),
	// )
	s.loadFromFile()
	return s, nil
}

func persistentFilePath(l *zap.Logger) string {
	return filepath.Join(config.DataHomeDir(), "fake_secret_store.json")
}

func (s *fileSecrets) Set(scope, name string, data map[string]string) error {
	s.secrets[secretPath(scope, name)] = data
	return s.updateFile()
}

func (s *fileSecrets) Get(scope, name string) (map[string]string, error) {
	data, ok := s.secrets[secretPath(scope, name)]
	if !ok {
		return nil, nil
	}
	return data, nil
}

func (s *fileSecrets) Append(scope, name, token string) error {
	data, err := s.Get(scope, name)
	if err != nil {
		return err
	}

	if data == nil {
		data = map[string]string{}
	}
	data[token] = time.Now().UTC().Format(time.RFC3339)
	err = s.Set(scope, name, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *fileSecrets) Delete(scope, name string) error {
	delete(s.secrets, secretPath(scope, name))
	return s.updateFile()
}

func (s *fileSecrets) loadFromFile() {
	if s.filePath == "" {
		return
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// TODO(ENG-146): This log line screws up CLI tests because it contains a
		// timestamp, use sed/grep to overcome this, and uncomment this log line.
		// s.logger.Warn("Secrets file not found",
		// 	zap.String("filePath", s.filePath),
		// )
		return
	}
	if err := json.Unmarshal(data, &s.secrets); err != nil {
		s.logger.Error("JSON unmarshalling error when loading secrets",
			zap.String("filePath", s.filePath),
			zap.Error(err),
		)
		return
	}
}

func (s *fileSecrets) updateFile() error {
	if s.filePath == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.secrets, "", "  ")
	if err != nil {
		s.logger.Error("JSON marshalling error when updating secrets",
			zap.Error(err),
		)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o700); err != nil {
		s.logger.Error("Directory creation error when writing secrets",
			zap.String("dir", filepath.Dir(s.filePath)),
			zap.Error(err),
		)
		return err
	}
	if err := os.WriteFile(s.filePath, data, 0o600); err != nil {
		s.logger.Error("File writing error when updating secrets",
			zap.String("filePath", s.filePath),
			zap.Error(err),
		)
		return err
	}
	return nil
}
