package l

func N(l L) Nullable {
	if l != nil {
		l = l.SkipCaller(1)
	}

	return Nullable{L: l}
}

type Nullable struct{ L }

var _ L = &Nullable{}

func nullableWrap(l L) L { return &Nullable{L: l} }

func Unwrap(l L) L {
	if l == nil {
		return nil
	}

	for {
		if unw, ok := l.(interface{ Unwrap() L }); ok {
			l = unw.Unwrap()
			continue
		}

		break
	}

	return l
}

func (n *Nullable) Unwrap() L { return n.L }

func (n *Nullable) Set(l L) { n.L = l }

func (n Nullable) SkipCaller(i int) L {
	if n.L == nil {
		return n
	}

	return nullableWrap(n.L.SkipCaller(i))
}

func (n Nullable) Named(name string) L {
	if n.L == nil {
		return n
	}

	return nullableWrap(n.L.Named(name))
}

func (n Nullable) With(vs ...interface{}) L {
	if n.L == nil {
		return n
	}

	return nullableWrap(n.L.With(vs...))
}

func (n Nullable) Debug(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Debug(s, args...)
}

func (n Nullable) Debugf(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Debugf(s, args...)
}

func (n Nullable) Info(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Info(s, args...)
}

func (n Nullable) Infof(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Infof(s, args...)
}

func (n Nullable) Warn(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Warn(s, args...)
}

func (n Nullable) Warnf(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Warnf(s, args...)
}

func (n Nullable) Error(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Error(s, args...)
}

func (n Nullable) Errorf(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Errorf(s, args...)
}

func (n Nullable) Fatal(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Fatal(s, args...)
}

func (n Nullable) Fatalf(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Fatalf(s, args...)
}

func (n Nullable) Panic(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Panic(s, args...)
}

func (n Nullable) Panicf(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.Panicf(s, args...)
}

func (n Nullable) DPanic(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.DPanic(s, args...)
}

func (n Nullable) DPanicf(s string, args ...interface{}) {
	if n.L == nil {
		return
	}

	n.L.DPanicf(s, args...)
}
