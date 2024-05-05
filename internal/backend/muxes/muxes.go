package muxes

import "net/http"

type Muxes struct {
	API, Root *http.ServeMux
	WrapAuth  func(http.Handler) http.Handler
}

func (m *Muxes) Handle(pattern string, handler http.Handler) {
	m.Root.Handle(pattern, handler)
}

func (m *Muxes) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Root.HandleFunc(pattern, handler)
}

func (m *Muxes) AuthHandle(pattern string, handler http.Handler) {
	m.Root.Handle(pattern, m.WrapAuth(handler))
}

func (m *Muxes) AuthHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Root.Handle(pattern, m.WrapAuth(http.HandlerFunc(handler)))
}
