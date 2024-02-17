package apply

import (
	"fmt"
)

type Log struct {
	Msg  string         `json:"msg"`
	Data map[string]any `json:"data"`
}

func (l Log) String() string { return fmt.Sprintf("%s %v", l.Msg, l.Data) }

func (l *Log) set(k string, v any) *Log {
	l.Data[k] = v
	return l
}
