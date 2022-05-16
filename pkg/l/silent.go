package l

type Silent struct{ L }

var _ L = Silent{}

func silentWrap(l L) L { return &Silent{L: l} }

func (sl Silent) SkipCaller(i int) L {
	return silentWrap(sl.L.SkipCaller(i))
}

func (sl Silent) Named(name string) L {
	return silentWrap(sl.L.Named(name))
}

func (sl Silent) With(vs ...interface{}) L {
	return silentWrap(sl.L.With(vs...))
}

func (sl Silent) Debug(s string, args ...interface{}) {
	sl.L.Debug(s, args...)
}

func (sl Silent) Debugf(s string, args ...interface{}) {
	sl.L.Debugf(s, args...)
}

func (sl Silent) Info(s string, args ...interface{}) {
	// Info -> Debug
	sl.L.Debug(s, args...)
}

func (sl Silent) Infof(s string, args ...interface{}) {
	// Info -> Debug
	sl.L.Debugf(s, args...)
}

func (sl Silent) Warn(s string, args ...interface{}) {
	sl.L.Warn(s, args...)
}

func (sl Silent) Warnf(s string, args ...interface{}) {
	sl.L.Warnf(s, args...)
}

func (sl Silent) Error(s string, args ...interface{}) {
	sl.L.Error(s, args...)
}

func (sl Silent) Errorf(s string, args ...interface{}) {
	sl.L.Errorf(s, args...)
}

func (sl Silent) Fatal(s string, args ...interface{}) {
	sl.L.Fatal(s, args...)
}

func (sl Silent) Fatalf(s string, args ...interface{}) {
	sl.L.Fatalf(s, args...)
}

func (sl Silent) Panic(s string, args ...interface{}) {
	sl.L.Panic(s, args...)
}

func (sl Silent) Panicf(s string, args ...interface{}) {
	sl.L.Panicf(s, args...)
}

func (sl Silent) DPanic(s string, args ...interface{}) {
	sl.L.DPanic(s, args...)
}

func (sl Silent) DPanicf(s string, args ...interface{}) {
	sl.L.DPanicf(s, args...)
}
