package manifest

import "fmt"

type Log func(string)

func (l Log) Printf(f string, xs ...any) { l(fmt.Sprintf(f, xs...)) }

type keyer interface {
	GetKey() string
}

func (l Log) For(kind string, keyer keyer) Log {
	if kind != "" {
		kind += " "
	}
	return Log(func(f string) { l.Printf("%s%v: %v", kind, keyer.GetKey(), f) })
}
