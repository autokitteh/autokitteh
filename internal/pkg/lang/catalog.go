package lang

import (
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

type NewLangFunc func(l L.L, name string) (Lang, error)

type CatalogLang struct {
	New  NewLangFunc
	Exts []string // extensions without the prefix dot.
}

type Catalog interface {
	Register(name string, l CatalogLang)

	// Returns lang name -> extensions.
	List() map[string][]string

	// Return a lang for the given name and scope.
	// If scope is empty, return a new lang.
	Acquire(name, scope string) (Lang, error)
}
