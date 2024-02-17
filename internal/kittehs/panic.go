package kittehs

var panicFunc = func(msg any) { panic(msg) }

func SetPanicFunc(f func(any)) { panicFunc = f }

func Panic(msg any) { panicFunc(msg) }
