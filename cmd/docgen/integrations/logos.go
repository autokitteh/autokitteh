package integrations

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
)

type logoGenerator struct {
	outputDir string
}

func (g logoGenerator) Output() string {
	return fmt.Sprintf("Logo images (%s)", g.outputDir)
}

func NewLogoGenerator(rootDir string) Generator {
	outputDir := filepath.Join(rootDir, "int/img")
	if err := resetDir(outputDir); err != nil {
		log.Fatal(err)
	}

	return logoGenerator{outputDir: outputDir}
}

func (g logoGenerator) Generate(akURL string, n int, i *integrationsv1.Integration) {
	if i.LogoUrl == "" {
		g.copyDefaultLogo(i.UniqueName)
		return
	}

	u := i.LogoUrl
	if !strings.HasPrefix(u, "http") {
		u, _ = url.JoinPath(akURL, u)
	}
	g.downloadLogo(u, i.UniqueName)
}

func (g logoGenerator) copyDefaultLogo(filename string) {
	// TODO: Copy some TBD image (AK logo? Generic/random cat?) to
	filepath.Join(g.outputDir, filename+".png")
}

func (g logoGenerator) downloadLogo(imageURL, filename string) {
	resp, err := http.Get(imageURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	image, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	filename += filepath.Ext(imageURL)
	path := filepath.Join(g.outputDir, filename)
	if err := os.WriteFile(path, image, 0o644); err != nil {
		log.Fatal(err)
	}
}
