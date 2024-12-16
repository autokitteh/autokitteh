package parsers

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"maps"

	"golang.org/x/tools/txtar"
	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	Parsers = map[string]func(io.Reader) (sdktypes.Value, error){
		"txt":         ParseText,
		"json":        ParseJSON,
		"yaml":        ParseYAML,
		"kitteh.yaml": ParseRawYAML,
		"kitteh.json": ParseRawJSON,
		"csv":         ParseCSV,
		"b64":         ParseBase64,
		"hex":         ParseHex,
		"xml":         ParseXML,
		"txtar":       ParseTxTar,
	}

	Extensions              = kittehs.IterToSlice(maps.Keys(Parsers))
	ExtensionsWithDotPrefix = kittehs.Transform(Extensions, func(s string) string { return "." + s })
)

func ParseJSON(r io.Reader) (sdktypes.Value, error) {
	// TODO: allow to read lists and sets?

	var m map[string]any

	if err := json.NewDecoder(r).Decode(&m); err != nil {
		// TODO: ProgramError.
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(m)
}

func ParseRawJSON(r io.Reader) (sdktypes.Value, error) {
	var v sdktypes.Value

	// TODO: there are some stuff that must not be allowed to be parseled, such as Functions.
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		// TODO: ProgramError.
		return sdktypes.InvalidValue, err
	}

	return v, nil
}

func ParseYAML(r io.Reader) (sdktypes.Value, error) {
	// TODO: allow to read lists and sets?

	var m map[string]any

	if err := yaml.NewDecoder(r).Decode(&m); err != nil {
		// TODO: ProgramError.
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(m)
}

func ParseRawYAML(r io.Reader) (sdktypes.Value, error) {
	var v sdktypes.Value

	// TODO: there are some stuff that must not be allowed to be parseled, such as Functions.
	if err := yaml.NewDecoder(r).Decode(&v); err != nil {
		// TODO: ProgramError.
		return sdktypes.InvalidValue, err
	}

	return v, nil
}

func ParseXML(r io.Reader) (sdktypes.Value, error) {
	// TODO: allow to read lists and sets?

	var m map[string]any

	if err := xml.NewDecoder(r).Decode(&m); err != nil {
		// TODO: ProgramError.
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(m)
}

func ParseText(r io.Reader) (sdktypes.Value, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"text": sdktypes.NewStringValue(string(data)),
	}), nil
}

func ParseCSV(r io.Reader) (sdktypes.Value, error) {
	rs, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	v := kittehs.Must1(sdktypes.WrapValue(rs))

	return sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"data": v,
	}), nil
}

func ParseHex(r io.Reader) (sdktypes.Value, error) {
	encoded, err := io.ReadAll(r)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	decoded, err := hex.DecodeString(string(encoded))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"data": sdktypes.NewBytesValue(decoded),
	}), nil
}

func ParseBase64(r io.Reader) (sdktypes.Value, error) {
	encoded, err := io.ReadAll(r)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"data": sdktypes.NewBytesValue(decoded),
	}), nil
}

func ParseTxTar(r io.Reader) (sdktypes.Value, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	archive := txtar.Parse(data)

	files := kittehs.ListToMap(archive.Files, func(f txtar.File) (string, sdktypes.Value) {
		return f.Name, sdktypes.NewStringValue(string(f.Data))
	})

	return sdktypes.NewDictValueFromStringMap(map[string]sdktypes.Value{
		"comment": sdktypes.NewStringValue(string(archive.Comment)),
		"files":   sdktypes.NewDictValueFromStringMap(files),
	}), nil
}
