package pythonrt

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteUserCodeToFS(t *testing.T) {
	t.Run("empty files map", func(t *testing.T) {
		files := map[string][]byte{}

		result, err := writeUserCodeToFS(files)

		require.NoError(t, err)
		require.NotEmpty(t, result.codePath)
		require.False(t, result.hasCustomRequirements)
		require.Equal(t, path.Join(result.codePath, "user_requirements.txt"), result.requirementsFilePath)

		// Verify directory exists and is writable
		info, err := os.Stat(result.codePath)
		require.NoError(t, err)
		require.True(t, info.IsDir())

		// Clean up
		defer os.RemoveAll(result.codePath)
	})

	t.Run("single Python file", func(t *testing.T) {
		files := map[string][]byte{
			"main.py": []byte("print('Hello, World!')"),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		require.NotEmpty(t, result.codePath)
		require.False(t, result.hasCustomRequirements)

		// Verify file was created with correct content
		content, err := os.ReadFile(path.Join(result.codePath, "main.py"))
		require.NoError(t, err)
		require.Equal(t, []byte("print('Hello, World!')"), content)
	})

	t.Run("multiple files with subdirectories", func(t *testing.T) {
		files := map[string][]byte{
			"main.py":           []byte("import utils.helper\nprint('main')"),
			"utils/helper.py":   []byte("def help(): pass"),
			"utils/__init__.py": []byte(""),
			"config.json":       []byte(`{"key": "value"}`),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		require.False(t, result.hasCustomRequirements)

		// Verify all files were created
		mainContent, err := os.ReadFile(path.Join(result.codePath, "main.py"))
		require.NoError(t, err)
		require.Equal(t, []byte("import utils.helper\nprint('main')"), mainContent)

		helperContent, err := os.ReadFile(path.Join(result.codePath, "utils/helper.py"))
		require.NoError(t, err)
		require.Equal(t, []byte("def help(): pass"), helperContent)

		initContent, err := os.ReadFile(path.Join(result.codePath, "utils/__init__.py"))
		require.NoError(t, err)
		require.Equal(t, []byte(""), initContent)

		configContent, err := os.ReadFile(path.Join(result.codePath, "config.json"))
		require.NoError(t, err)
		require.Equal(t, []byte(`{"key": "value"}`), configContent)
	})

	t.Run("requirements.txt file", func(t *testing.T) {
		files := map[string][]byte{
			"main.py":          []byte("import requests"),
			"requirements.txt": []byte("requests==2.28.1\nnumpy>=1.20.0"),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		require.True(t, result.hasCustomRequirements)
		require.Equal(t, path.Join(result.codePath, "user_requirements.txt"), result.requirementsFilePath)

		// Verify requirements.txt was renamed to user_requirements.txt
		requirementsContent, err := os.ReadFile(path.Join(result.codePath, "user_requirements.txt"))
		require.NoError(t, err)
		require.Equal(t, []byte("requests==2.28.1\nnumpy>=1.20.0"), requirementsContent)

		// Verify original requirements.txt doesn't exist
		_, err = os.Stat(path.Join(result.codePath, "requirements.txt"))
		require.True(t, os.IsNotExist(err))

		// Verify main.py still exists
		mainContent, err := os.ReadFile(path.Join(result.codePath, "main.py"))
		require.NoError(t, err)
		require.Equal(t, []byte("import requests"), mainContent)
	})

	t.Run("files with dots in name are skipped", func(t *testing.T) {
		files := map[string][]byte{
			"main.py":    []byte("print('main')"),
			".gitignore": []byte("*.pyc"),
			".env":       []byte("SECRET=value"),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		// Verify main.py was created
		_, err = os.Stat(path.Join(result.codePath, "main.py"))
		require.NoError(t, err)

		// Verify dotfiles were skipped
		_, err = os.Stat(path.Join(result.codePath, ".gitignore"))
		require.True(t, os.IsNotExist(err))

		_, err = os.Stat(path.Join(result.codePath, ".env"))
		require.True(t, os.IsNotExist(err))
	})

	t.Run("directory paths are skipped", func(t *testing.T) {
		files := map[string][]byte{
			"main.py": []byte("print('main')"),
			"utils/":  []byte("should be ignored"),
			"config/": []byte("also ignored"),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		// Verify main.py was created
		_, err = os.Stat(path.Join(result.codePath, "main.py"))
		require.NoError(t, err)

		// Verify directory entries were skipped (directories shouldn't exist as files)
		_, err = os.Stat(path.Join(result.codePath, "utils"))
		require.True(t, os.IsNotExist(err))

		_, err = os.Stat(path.Join(result.codePath, "config"))
		require.True(t, os.IsNotExist(err))
	})

	t.Run("nested directories are created", func(t *testing.T) {
		files := map[string][]byte{
			"deep/nested/path/file.py": []byte("# deeply nested"),
			"another/path/test.py":     []byte("# another path"),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		// Verify nested file was created
		content, err := os.ReadFile(path.Join(result.codePath, "deep/nested/path/file.py"))
		require.NoError(t, err)
		require.Equal(t, []byte("# deeply nested"), content)

		content, err = os.ReadFile(path.Join(result.codePath, "another/path/test.py"))
		require.NoError(t, err)
		require.Equal(t, []byte("# another path"), content)
	})

	t.Run("empty requirements.txt", func(t *testing.T) {
		files := map[string][]byte{
			"main.py":          []byte("print('no deps')"),
			"requirements.txt": []byte(""),
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		require.True(t, result.hasCustomRequirements)

		// Verify empty requirements file was created
		content, err := os.ReadFile(path.Join(result.codePath, "user_requirements.txt"))
		require.NoError(t, err)
		require.Equal(t, []byte(""), content)
	})

	t.Run("binary file content", func(t *testing.T) {
		binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
		files := map[string][]byte{
			"main.py":   []byte("print('with binary')"),
			"image.png": binaryData,
		}

		result, err := writeUserCodeToFS(files)
		require.NoError(t, err)
		defer os.RemoveAll(result.codePath)

		// Verify binary file was written correctly
		content, err := os.ReadFile(path.Join(result.codePath, "image.png"))
		require.NoError(t, err)
		require.Equal(t, binaryData, content)
	})
}
