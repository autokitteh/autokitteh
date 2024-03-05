package tar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type tarFile struct {
	hdr  *tar.Header
	data []byte
}

type TarArchive struct {
	data []tarFile
}

func NewTarFile() *TarArchive {
	return &TarArchive{}
}

func FromBytes(b []byte, gzipped bool) (*TarArchive, error) {
	var reader io.Reader
	reader = bytes.NewReader(b)

	if gzipped {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}

		defer gzReader.Close()
		reader = gzReader
	}
	r := tar.NewReader(reader)

	tf := NewTarFile()
	for {
		hdr, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		buf := bytes.NewBuffer(make([]byte, 0, hdr.Size))
		if _, err := io.Copy(buf, r); err != nil {
			return nil, fmt.Errorf("read %q: %w", hdr.Name, err)
		}

		tf.data = append(tf.data, tarFile{hdr: hdr, data: buf.Bytes()})
	}

	return tf, nil
}

func (ta *TarArchive) AddDir(dir fs.ReadDirFS, root string) error {
	return fs.WalkDir(dir, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			ta.Add(path, data)
		}
		return nil
	})
}

func (ta *TarArchive) AddFile(dir fs.ReadDirFS, file string) error {
	return ta.AddDir(dir, file)
}

func (ta *TarArchive) Add(name string, data []byte) {
	ta.data = append(ta.data, tarFile{hdr: &tar.Header{Name: name, Size: int64(len(data)), Mode: 0644}, data: data})
}

func (ta *TarArchive) Bytes(gzipped bool) ([]byte, error) {
	if len(ta.data) == 0 {
		return []byte{}, nil
	}

	var buf bytes.Buffer
	if err := ta.readInto(&buf, gzipped); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ta *TarArchive) Content() (map[string][]byte, error) {
	if len(ta.data) == 0 {
		return map[string][]byte{}, nil
	}

	return kittehs.ListToMap(ta.data, func(c tarFile) (string, []byte) {
		return c.hdr.Name, c.data
	}), nil
}

func (ta *TarArchive) readInto(writer io.Writer, gzipped bool) error {
	if gzipped {
		gwWriter := gzip.NewWriter(writer)
		defer gwWriter.Close()
		writer = gwWriter
	}

	t := tar.NewWriter(writer)
	defer t.Close()
	for _, c := range ta.data {
		if err := t.WriteHeader(c.hdr); err != nil {
			return err
		}

		if _, err := t.Write(c.data); err != nil {
			return err
		}
	}

	return nil
}
