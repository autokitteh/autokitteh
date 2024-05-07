package dashboardsvc

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

type listItem[M proto.Message] interface {
	ToProto() M
	FieldsOrder() []string
	HideFields() []string
}

type list struct {
	Headers          []string
	UnformattedItems [][]string
	Items            [][]template.HTML
	N                int
}

func genListData[T listItem[M], M proto.Message](xs []T, drops ...string) list {
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

	for i := 0; i < fs.Len(); i++ {
		fd := fs.Get(i)

		if fd.Kind() == protoreflect.MessageKind || fd.Cardinality() == protoreflect.Repeated {
			continue
		}

		if name := fmt.Sprint(fd.Name()); allowed(name) {
			hdrs[i] = name
		}
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
		for _, n := range hdrs {
			if n != "" {
				fd := fs.ByName(protoreflect.Name(n))
				fv := xr.Get(fd)
				v := fv.String()

				if en := fd.Enum(); en != nil {
					v = fmt.Sprint(en.Values().ByNumber(fv.Enum()).Name())
				}

				item = append(item, formatField(fmt.Sprint(n), v))
				uitem = append(uitem, v)
			}
		}
		items[i] = item
		uitems[i] = uitem
	}

	hdrs = kittehs.FilterZeroes(hdrs)

	return list{hdrs, uitems, items, len(items)}
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
