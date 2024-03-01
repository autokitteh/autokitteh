package sdktypes

import (
	"fmt"
	"strconv"
	"strings"

	programv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/program/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

const (
	symbolSeparator = ","
	pathSeparator   = ":"
	rowColSeparator = "."
)

type CodeLocationPB = programv1.CodeLocation

type CodeLocation = *object[*CodeLocationPB]

var (
	CodeLocationFromProto       = makeFromProto(validateCodeLocation)
	StrictCodeLocationFromProto = makeFromProto(strictValidateCodeLocation)
	ToStrictCodeLocation        = makeWithValidator(strictValidateCodeLocation)
)

func strictValidateCodeLocation(pb *programv1.CodeLocation) error {
	return validateCodeLocation(pb)
}

func validateCodeLocation(pb *programv1.CodeLocation) error {
	// NOTE: ensure to return correct sdk errors if any
	return nil
}

func GetCodeLocationPath(l CodeLocation) string {
	if l == nil {
		return ""
	}

	return l.pb.Path
}

func GetCodeLocationRowCol(l CodeLocation) (uint32, uint32) {
	if l == nil {
		return 0, 0
	}

	return l.pb.Row, l.pb.Col
}

func GetCodeLocationName(l CodeLocation) string {
	if l == nil {
		return ""
	}

	return l.pb.Name
}

func GetCodeLocationCanonicalString(l CodeLocation) string {
	if l == nil {
		return ""
	}

	return getCodeLocationPBCanonicalString(l.pb)
}

func getCodeLocationPBCanonicalString(pb *CodeLocationPB) string {
	if pb == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(pb.Path)

	name := pb.Name

	if pb.Row != 0 || pb.Col != 0 {
		b.WriteString(fmt.Sprintf("%s%d", pathSeparator, pb.Row))

		if pb.Col != 0 {
			b.WriteString(fmt.Sprintf("%s%d", rowColSeparator, pb.Col))
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

func StrictParseCodeLocation(s string) (CodeLocation, error) {
	if s == "" {
		return nil, sdkerrors.ErrInvalidArgument
	}

	return ParseCodeLocation(s)
}

func ParseCodeLocation(s string) (CodeLocation, error) {
	if s == "" {
		return nil, nil
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
				return nil, err
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
			return nil, err
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
