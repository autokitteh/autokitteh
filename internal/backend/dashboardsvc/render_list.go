package dashboardsvc

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

type listItem[M proto.Message] interface {
	ToProto() M
	FieldsOrder() []string
	HideFields() []string
	ExtraFields() map[string]any
}

type list struct {
	Scope            any
	Headers          []string
	UnformattedItems [][]string
	Items            [][]template.HTML
	N                int
}

func genListData[T listItem[M], M proto.Message](scope any, xs []T, drops ...string) list {
	var (
		m M
		x T
	)

	allowed := func(string) bool { return true }
	drops = append(drops, x.HideFields()...)
	if len(drops) > 0 {
		hasDrop := kittehs.ContainedIn(drops...)
		allowed = func(s string) bool { return !hasDrop(s) }
	}

	fs := m.ProtoReflect().Descriptor().Fields()

	hdrs := make([]string, fs.Len())

	i := 0
	for ; i < fs.Len(); i++ {
		fd := fs.Get(i)

		if fd.Kind() == protoreflect.MessageKind {
			switch fd.Message().FullName() {
			case "google.protobuf.Timestamp":
			case "google.protobuf.Duration":
				// let it go
			default:
				continue
			}
		}

		if name := fmt.Sprint(fd.Name()); allowed(name) {
			if fd.Cardinality() == protoreflect.Repeated {
				name = "#" + name
			}

			hdrs[i] = name
		}
	}

	for n := range x.ExtraFields() {
		hdrs = append(hdrs, n)
	}

	fo := x.FieldsOrder()
	sort.Slice(hdrs, func(i, j int) bool {
		ki, _ := kittehs.FindFirst(fo, func(n string) bool { return n == hdrs[i] })
		kj, _ := kittehs.FindFirst(fo, func(n string) bool { return n == hdrs[j] })

		if ki < 0 && kj < 0 {
			return hdrs[i] < hdrs[j]
		}

		if ki < 0 {
			return false
		}

		if kj < 0 {
			return true
		}

		return ki < kj
	})

	items := make([][]template.HTML, len(xs))
	uitems := make([][]string, len(xs))
	for i, x := range xs {
		xr := x.ToProto().ProtoReflect()
		item := make([]template.HTML, 0, fs.Len())
		uitem := make([]string, 0, fs.Len())
		for i, n := range hdrs {
			var v string

			if n != "" {
				if i < fs.Len() {
					fd := fs.ByName(protoreflect.Name(n))
					fv := xr.Get(fd)
					v = fv.String()

					if fd.Kind() == protoreflect.MessageKind {
						switch fd.Message().FullName() {
						case "google.protobuf.Timestamp":
							ts := fv.Message().Interface().(*timestamppb.Timestamp)
							v = ts.AsTime().Format("2006-01-02 15:04:05")
						case "google.protobuf.Duration":
							d := fv.Message().Interface().(*durationpb.Duration)
							v = d.String()
						default:
							continue
						}
					} else if fd.Kind() == protoreflect.EnumKind {
						v = fmt.Sprint(fd.Enum().Values().ByNumber(fv.Enum()).Name())
					} else if en := fd.Enum(); en != nil {
						v = fmt.Sprint(en.Values().ByNumber(fv.Enum()).Name())
					} else if fd.Cardinality() == protoreflect.Repeated {
						v = strconv.Itoa(fv.List().Len())
					}
				} else {
					v = fmt.Sprint(x.ExtraFields()[n])
				}

				item = append(item, formatField(n, v))
				uitem = append(uitem, v)
			}
		}
		items[i] = item
		uitems[i] = uitem
	}

	hdrs = kittehs.FilterZeroes(hdrs)

	return list{scope, hdrs, uitems, items, len(items)}
}

func renderList(w http.ResponseWriter, r *http.Request, title string, l list) {
	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "list.html", struct {
		Message string
		Title   string
		List    any
	}{
		Title: title,
		List:  l,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
