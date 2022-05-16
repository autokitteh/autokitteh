package svc

import (
	"net"
	"reflect"
)

type Providers struct{ Vs []interface{} }

func (p *Providers) Add(v ...interface{}) { p.Vs = append(p.Vs, v...) }

func (p *Providers) Get(dst interface{}) bool {
	dt := reflect.TypeOf(dst)
	if dt.Kind() != reflect.Ptr {
		panic("dst must be a ptr")
	}

	det := dt.Elem()

	for _, v := range p.Vs {
		vt := reflect.TypeOf(v)

		if vt == det || (det.Kind() == reflect.Interface && vt.Implements(det)) {
			reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(v))
			return true
		}
	}

	return false
}

type GRPCAddr net.Addr
