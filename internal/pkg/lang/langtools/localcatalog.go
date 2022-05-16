package langtools

import (
	"fmt"
	"sync"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type langent struct {
	lang.CatalogLang

	cache map[string]lang.Lang // scope -> lang. TODO: eviction.
	lock  sync.Mutex
}

type LocalCatalog struct {
	L     L.Nullable
	Langs map[string]*langent
}

var (
	PermissiveCatalog    = &LocalCatalog{}
	DeterministicCatalog = &LocalCatalog{}
)

func (c *LocalCatalog) Register(name string, l lang.CatalogLang) {
	if l.New == nil {
		panic("new is nil")
	}

	if c.Langs == nil {
		c.Langs = make(map[string]*langent)
	}

	if _, found := c.Langs[name]; found {
		panic(fmt.Errorf("lang %q is already registered", name))
	}

	c.Langs[name] = &langent{CatalogLang: l}
}

func (c *LocalCatalog) List() map[string][]string {
	m := make(map[string][]string, len(c.Langs))

	for n, l := range c.Langs {
		m[n] = l.CatalogLang.Exts
	}

	return m
}

func (c *LocalCatalog) Acquire(name, scope string) (lang.Lang, error) {
	l := c.Langs[name]
	if l == nil {
		return nil, fmt.Errorf("%w: %q", lang.ErrLangNotRegistered, name)
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	if scope == "" {
		ll, err := l.New(c.L.Named(name), name)
		if err != nil {
			return nil, fmt.Errorf("new %q: %w", name, err)
		}

		return ll, nil
	}

	if ll := l.cache[scope]; ll != nil {
		return ll, nil
	}

	ll, err := l.New(c.L.Named(name).With("scope", scope), name)
	if err != nil {
		return nil, fmt.Errorf("new %q: %w", name, err)
	}

	if l.cache == nil {
		l.cache = make(map[string]lang.Lang)
	}

	l.cache[scope] = ll

	return ll, nil
}
