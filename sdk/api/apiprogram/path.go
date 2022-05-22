package apiprogram

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"

	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"

	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
)

var pathRe = regexp.MustCompile(`` +
	/* scheme:  */ `^(?:(\$?[[:word:]]+):)?` +
	/* path     */ `((?:[[:word:]]|[\/~\.\-\+])+)` +
	/* #version */ `(?:#([[:word:]]+))?$`,
)

type PathPB = pbprogram.Path

type Path struct{ pb *pbprogram.Path }

func (p *Path) PB() *pbprogram.Path {
	if p == nil || p.pb == nil {
		return nil
	}

	return proto.Clone(p.pb).(*pbprogram.Path)
}

func (p *Path) Clone() *Path { return &Path{pb: p.PB()} }

func (p *Path) Path() string    { return p.pb.Path }
func (p *Path) Scheme() string  { return p.pb.Scheme }
func (p *Path) Version() string { return p.pb.Version }

func (p *Path) WithVersion(v string) *Path {
	p = p.Clone()
	p.pb.Version = v
	return p
}

func (p *Path) PathAndVersion() string {
	if p.pb.Version == "" {
		return p.pb.Path
	}

	return fmt.Sprintf("%s#%s", p.pb.Path, p.pb.Version)
}

func (p *Path) Ext() string { return filepath.Ext(p.Path()) }

func (p *Path) String() string {
	if p == nil || p.pb == nil {
		return ""
	}

	s := p.Scheme()
	if s != "" {
		s += ":"
	}

	s += p.Path()

	if v := p.Version(); v != "" {
		s += fmt.Sprintf("#%s", v)
	}

	return s
}

func (p *Path) Equal(o *Path) bool { return p.String() == o.String() }

func (p *Path) IsInternal() bool { return strings.HasPrefix(p.Scheme(), "$") }

func (p *Path) PluginID() (apiplugin.PluginID, bool) {
	if p == nil || p.pb == nil {
		return "", false
	}

	if (p.Scheme() == "" && !strings.Contains(p.Path(), "/")) || p.Scheme() == "plugin" {
		pv := p.PathAndVersion()

		if !strings.Contains(pv, ".") {
			pv = fmt.Sprintf("internal.%s", pv)
		}

		return apiplugin.PluginID(pv), true
	}

	return "", false
}

func MustPathFromProto(pb *pbprogram.Path) *Path {
	p, err := PathFromProto(pb)
	if err != nil {
		panic(err)
	}
	return p
}

func PathFromProto(pb *pbprogram.Path) (*Path, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return (&Path{pb: pb}).Clone(), nil
}

func NewPath(scheme, path, version string) (*Path, error) {
	return PathFromProto(&pbprogram.Path{Scheme: scheme, Path: path, Version: version})
}

func ParsePathString(s string) (*Path, error) {
	ms := pathRe.FindAllStringSubmatch(s, -1)
	if ms == nil {
		return nil, fmt.Errorf("invalid path")
	}

	return PathFromProto(&pbprogram.Path{
		Scheme:  ms[0][1],
		Path:    ms[0][2],
		Version: ms[0][3],
	})
}

func ParsePathStringOr(s string, or *Path) *Path {
	p, err := ParsePathString(s)
	if err != nil {
		return or
	}
	return p
}

func MustParsePathString(s string) *Path {
	p, err := ParsePathString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func JoinPaths(root, path *Path) (*Path, error) {
	if path.IsInternal() {
		return path, nil
	}

	if path.Scheme() != "" && root.Scheme() != path.Scheme() {
		return nil, fmt.Errorf("root %q and path scheme %q differ ", root.Scheme(), path.Scheme())
	}

	newPath := filepath.Join(root.Path(), path.Path())
	// [# path-protect #]
	if root.Path() != "." && !strings.HasPrefix(newPath, root.Path()) {
		return nil, fmt.Errorf("path %q must be under root %q", newPath, root.Path())
	}

	ver := root.Version()
	if ver == "" {
		ver = path.Version()
	} else if pver := path.Version(); pver != "" && pver != ver {
		return nil, fmt.Errorf("path version %q must be same as root version %q", pver, ver)
	}

	return NewPath(root.Scheme(), newPath, ver)
}
