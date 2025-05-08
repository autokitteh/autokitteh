package pythonrt

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func getRequirementsHash(artifact []byte) (string, error) {
	tf := tar.NewReader(bytes.NewReader(artifact))

	for {
		hdr, err := tf.Next()
		if errors.Is(err, io.EOF) {
			return "", nil
		}

		if err != nil {
			return "", err
		}

		if hdr.Name != "requirements.txt" {
			continue
		}

		var b bytes.Buffer

		if _, err := io.Copy(&b, tf); err != nil {
			return "", err
		}

		return kittehs.SHA256Hash(b.Bytes())
	}
}
