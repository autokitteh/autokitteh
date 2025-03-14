package prepare

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func CopyRunnerCode(destDir string) error {
	// Get the absolute path of the runner code
	runnerDir := filepath.Join(filepath.Dir(os.Args[0]), "..", "..", "..", "runner")
	return filepath.Walk(runnerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip node_modules and dist directories
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "dist") {
			return filepath.SkipDir
		}

		// Get the relative path from runnerDir
		relPath, err := filepath.Rel(runnerDir, path)
		if err != nil {
			return err
		}

		// Create the destination path
		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Copy the file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}

func ExtractTar(destDir string, data []byte) error {
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		fp := filepath.Join(destDir, hdr.Name)
		if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
			return fmt.Errorf("create dir %q: %w", filepath.Dir(fp), err)
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return fmt.Errorf("create file %q: %w", fp, err)
		}

		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return fmt.Errorf("write to file %q: %w", fp, err)
		}

		f.Close()
	}

	return nil
}

func SetupTypeScript(dir string) error {
	cmd := exec.Command("npm", "install", "typescript", "@types/node", "ts-node")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install typescript: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return nil
}

func InstallDependencies(dir string) error {
	cmd := exec.Command("npm", "install")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return nil
}
