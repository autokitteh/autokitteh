package z

import (
	"go.uber.org/zap"

	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

type ZL struct{ Z *zap.SugaredLogger }

var _ L.L = &ZL{}

func NewL(cfg Config, f func(*zap.Config), zopts []zap.Option) (*ZL, error) {
	if zopts == nil {
		zopts = DefaultOpts
	}

	z, err := New(cfg, f, append(zopts, zap.AddCallerSkip(1)))
	if err != nil {
		return nil, err
	}

	return &ZL{Z: z}, nil
}

func FromL(l L.L) *zap.SugaredLogger {
	if l == nil {
		return nil
	}

	return L.Unwrap(l).(*ZL).Z
}

func wrap(z *zap.SugaredLogger) L.L { return &ZL{Z: z} }

func (zl *ZL) Named(n string) L.L         { return wrap(zl.Z.Named(n)) }
func (zl *ZL) With(vs ...interface{}) L.L { return wrap(zl.Z.With(vs...)) }

func (zl *ZL) SkipCaller(i int) L.L {
	return wrap(zl.Z.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar())
}

func (zl *ZL) Debug(s string, args ...interface{})   { zl.Z.Debugw(s, args...) }
func (zl *ZL) Debugf(s string, args ...interface{})  { zl.Z.Debugf(s, args...) }
func (zl *ZL) Info(s string, args ...interface{})    { zl.Z.Infow(s, args...) }
func (zl *ZL) Infof(s string, args ...interface{})   { zl.Z.Infof(s, args...) }
func (zl *ZL) Warn(s string, args ...interface{})    { zl.Z.Warnw(s, args...) }
func (zl *ZL) Warnf(s string, args ...interface{})   { zl.Z.Warnf(s, args...) }
func (zl *ZL) Error(s string, args ...interface{})   { zl.Z.Errorw(s, args...) }
func (zl *ZL) Errorf(s string, args ...interface{})  { zl.Z.Errorf(s, args...) }
func (zl *ZL) Fatal(s string, args ...interface{})   { zl.Z.Fatalw(s, args...) }
func (zl *ZL) Fatalf(s string, args ...interface{})  { zl.Z.Fatalf(s, args...) }
func (zl *ZL) Panic(s string, args ...interface{})   { zl.Z.Panicw(s, args...) }
func (zl *ZL) Panicf(s string, args ...interface{})  { zl.Z.Panicf(s, args...) }
func (zl *ZL) DPanic(s string, args ...interface{})  { zl.Z.DPanicw(s, args...) }
func (zl *ZL) DPanicf(s string, args ...interface{}) { zl.Z.DPanicf(s, args...) }
