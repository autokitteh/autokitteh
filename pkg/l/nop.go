package l

type nop struct{}

var Nop nop

func (nop) Named(string) L        { return Nop }
func (nop) With(...interface{}) L { return Nop }
func (nop) SkipCaller(int) L      { return Nop }

func (nop) Debug(string, ...interface{})   {}
func (nop) Debugf(string, ...interface{})  {}
func (nop) Info(string, ...interface{})    {}
func (nop) Infof(string, ...interface{})   {}
func (nop) Warn(string, ...interface{})    {}
func (nop) Warnf(string, ...interface{})   {}
func (nop) Error(string, ...interface{})   {}
func (nop) Errorf(string, ...interface{})  {}
func (nop) Fatal(string, ...interface{})   {}
func (nop) Fatalf(string, ...interface{})  {}
func (nop) Panic(string, ...interface{})   {}
func (nop) Panicf(string, ...interface{})  {}
func (nop) DPanic(string, ...interface{})  {}
func (nop) DPanicf(string, ...interface{}) {}
