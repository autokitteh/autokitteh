package webhooks

import (
	"archive/zip"
	"context"
	_ "embed"
	"html/template"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

//go:embed manifest/manifest.json
var manifestTemplateText string

//go:embed manifest/color.png
var manifestColorImage []byte

//go:embed manifest/outline.png
var manifestOutlineImage []byte

var manifestTemplate = template.Must(template.New("manifest").Parse(manifestTemplateText))

const (
	ManifestPath = "/azurebot/{cid}/manifest.zip"
)

func (h handler) getVars(ctx context.Context, cid sdktypes.ConnectionID) (vars Vars, err error) {
	var vs sdktypes.Vars

	vs, err = h.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
		return
	}

	vs.Decode(&vars)

	return
}

// HandleManifest generates a bundle zip file to be used by a Teams Administrator
// to create an app that corresponds for this connection.
func (h handler) HandleManifest(w http.ResponseWriter, r *http.Request) {
	cidStr := r.PathValue("cid")

	cid, err := sdktypes.StrictParseConnectionID(cidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l := h.logger.With(zap.String("connection_id", cid.String()))

	vars, err := h.getVars(r.Context(), cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"manifest.zip\"")

	zw := zip.NewWriter(w)

	fw, err := zw.Create("manifest.json")
	if err != nil {
		l.Error("failed to create manifest.json", zap.Error(err))
		return
	}

	if err := manifestTemplate.Execute(fw, vars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fw, err = zw.Create("color.png")
	if err != nil {
		l.Error("failed to create color.png", zap.Error(err))
		return
	}

	if _, err := fw.Write(manifestColorImage); err != nil {
		l.Error("failed to write color.png", zap.Error(err))
		return
	}

	fw, err = zw.Create("outline.png")
	if err != nil {
		l.Error("failed to create outline.png", zap.Error(err))
		return
	}

	if _, err := fw.Write(manifestOutlineImage); err != nil {
		l.Error("failed to write outline.png", zap.Error(err))
		return
	}

	zw.Close()
}
