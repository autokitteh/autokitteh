package webhooks

import (
	"archive/zip"
	"context"
	_ "embed"
	"fmt"
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

	l := h.logger.With(zap.String("connection_id", cidStr))

	internalError := func(desc string, err error) {
		l.Error(desc, zap.Error(err))
		http.Error(w, desc, http.StatusInternalServerError)
	}

	cid, err := sdktypes.StrictParseConnectionID(cidStr)
	if err != nil {
		l.Info("failed to parse connection id", zap.Error(err))
		http.Error(w, fmt.Errorf("failed to parse connection ID: %w", err).Error(), http.StatusBadRequest)
		return
	}

	vars, err := h.getVars(r.Context(), cid)
	if err != nil {
		internalError("failed to get connection vars", err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"manifest.zip\"")

	zw := zip.NewWriter(w)

	fw, err := zw.Create("manifest.json")
	if err != nil {
		internalError("failed to create manifest.json", err)
		return
	}

	if err := manifestTemplate.Execute(fw, vars); err != nil {
		internalError("failed to execute manifest template", err)
		return
	}

	fw, err = zw.Create("color.png")
	if err != nil {
		internalError("failed to create color.png", err)
		return
	}

	if _, err := fw.Write(manifestColorImage); err != nil {
		internalError("failed to write color.png", err)
		return
	}

	fw, err = zw.Create("outline.png")
	if err != nil {
		internalError("failed to create outline.png", err)
		return
	}

	if _, err := fw.Write(manifestOutlineImage); err != nil {
		internalError("failed to write outline.png", err)
		return
	}

	if err := zw.Close(); err != nil {
		internalError("failed to close zip writer", err)
		return
	}
}
