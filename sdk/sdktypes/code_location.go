package sdktypes

import (
	"fmt"
	"strconv"
	"strings"

	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
)

const (
	symbolSeparator = ","
	pathSeparator   = ":"
	rowColSeparator = "."
)

type CodeLocation struct {
	object[*CodeLocationPB, CodeLocationTraits]
}

func init() { registerObject[CodeLocation]() }

type CodeLocationPB = programv1.CodeLocation

type CodeLocationTraits struct{ immutableObjectTrait }

func (CodeLocationTraits) Validate(m *CodeLocationPB) error         { return nil }
func (t CodeLocationTraits) StrictValidate(m *CodeLocationPB) error { return nonzeroMessage(m) }

func (l CodeLocation) Path() string { return l.read().Path }
func (l CodeLocation) Col() uint32  { return l.read().Col }
func (l CodeLocation) Row() uint32  { return l.read().Row }
func (l CodeLocation) Name() string { return l.read().Name }

func canonicalString(m *CodeLocationPB) string {
	if m == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.Path)

	name := m.Name

	if m.Row != 0 || m.Col != 0 {
		b.WriteString(fmt.Sprintf("%s%d", pathSeparator, m.Row))

		if m.Col != 0 {
			b.WriteString(fmt.Sprintf("%s%d", rowColSeparator, m.Col))
		}

		if name != "" {
			b.WriteString(symbolSeparator)
		}
	} else if name != "" {
		b.WriteString(pathSeparator)
	}

	if name != "" {
		b.WriteString(name)
	}

	return b.String()
}

func (l CodeLocation) CanonicalString() string { return canonicalString(l.m) }

func CodeLocationFromProto(m *CodeLocationPB) (CodeLocation, error) {
	return FromProto[CodeLocation](m)
}

func StrictParseCodeLocation(s string) (CodeLocation, error) {
	return Strict(ParseCodeLocation(s))
}

func ParseCodeLocation(s string) (CodeLocation, error) {
	if s == "" {
		return CodeLocation{}, nil
	}

	path, after, ok := strings.Cut(s, pathSeparator)
	if !ok {
		return CodeLocationFromProto(&CodeLocationPB{Path: path})
	}

	rowCol := func(s string) (uint32, uint32, error) {
		r, c, ok := strings.Cut(s, rowColSeparator)

		rr, err := strconv.ParseUint(r, 0, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("row: %w", err)
		}

		var cc uint64

		if ok {
			var err error
			cc, err = strconv.ParseUint(c, 0, 32)
			if err != nil {
				return 0, 0, fmt.Errorf("col: %w", err)
			}
		}

		return uint32(rr), uint32(cc), nil
	}

	a, b, ok := strings.Cut(after, symbolSeparator)
	if ok {
		if _, err := strconv.Atoi(a); strings.Contains(a, rowColSeparator) || err == nil {
			r, c, err := rowCol(a)
			if err != nil {
				return CodeLocation{}, err
			}

			return CodeLocationFromProto(&CodeLocationPB{
				Path: path,
				Row:  r,
				Col:  c,
				Name: b,
			})
		}

		return CodeLocationFromProto(&CodeLocationPB{
			Path: path,
			Name: b,
		})
	}

	if _, err := strconv.Atoi(a); strings.Contains(a, rowColSeparator) || err == nil {
		r, c, err := rowCol(a)
		if err != nil {
			return CodeLocation{}, err
		}

		return CodeLocationFromProto(&CodeLocationPB{
			Path: path,
			Row:  r,
			Col:  c,
		})
	}

	return CodeLocationFromProto(&CodeLocationPB{
		Path: path,
		Name: a,
	})
}
