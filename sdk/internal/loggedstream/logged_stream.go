package loggedstream

import (
	"errors"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Stream[Rx, Tx any] interface {
	Send(*Tx) error
	Receive() (*Rx, error)
}

type LoggedStream[Rx, Tx any] struct {
	SL     *zap.SugaredLogger
	S      Stream[Rx, Tx]
	DescRx func(*Rx) string
	DescTx func(*Tx) string
	Level  zapcore.Level
}

func (ls *LoggedStream[Rx, Tx]) Send(req *Tx) error {
	sl := ls.SL.With("msg", req)

	sl.Logf(ls.Level, "TX: %s", ls.DescTx(req))

	err := ls.S.Send(req)
	if err != nil {
		sl.Errorw("send failed", "error", err)
		return err
	}

	return nil
}

func (ls *LoggedStream[Rx, Tx]) Receive() (*Rx, error) {
	res, err := ls.S.Receive()
	if err != nil {
		if errors.Is(err, io.EOF) {
			ls.SL.Log(ls.Level, "EOF")
			return nil, err
		}

		ls.SL.Errorw("receive failed", "error", err)
		return nil, err
	}

	ls.SL.With("msg", res).Logf(ls.Level, "RX: %s", ls.DescRx(res))

	return res, nil
}
