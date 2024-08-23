package kittehs

import (
	"errors"
	"fmt"
)

var ErrInterpolatorUnexpectedEndOfInput = errors.New("unexpected end of input")

type Interpolator struct {
	Left, Right  string
	EvaluateExpr func(string) (string, error)
	ValidateExpr func(string) error
}

func (i Interpolator) Validate(in string) error {
	v := i.ValidateExpr

	if v == nil {
		v = func(string) error { return nil }
	}

	_, err := i.execute(in, func(expr string) (string, error) { return "", v(expr) })
	return err
}

func (i Interpolator) Execute(in string) (string, error) {
	if i.EvaluateExpr == nil {
		return "", errors.New("missing Evaluate function")
	}

	return i.execute(in, i.EvaluateExpr)
}

func (i Interpolator) execute(in string, eval func(string) (string, error)) (out string, err error) {
	const (
		stateScan = iota
		stateEscape
		stateLeftBracket
		stateRightBracket
		stateExpr
	)

	var (
		expr    string
		bracket struct {
			start, end, index int
		}
	)

	state := stateScan

	var step func(int, rune) error
	step = func(pos int, ch rune) (err error) {
		switch state {
		case stateScan:
			if ch == 0 {
				return
			}

			if ch == '\\' {
				state = stateEscape
				return
			}

			if len(i.Left) != 0 && ch == rune(i.Left[0]) {
				bracket.index = 0
				bracket.start = pos
				expr = ""
				state = stateLeftBracket
				err = step(pos, ch)
				return
			}

			out += string(ch)
			return

		case stateEscape:
			if ch == 0 {
				err = ErrInterpolatorUnexpectedEndOfInput
				return
			}

			out += string(ch)
			state = stateScan
			return

		case stateLeftBracket:
			if len(i.Left) == bracket.index {
				state = stateExpr
				err = step(pos, ch)
				return
			}

			if ch != 0 {
				if i.Left[bracket.index] == byte(ch) {
					bracket.index++
					return
				}

				out += in[bracket.start:pos]
				state = stateScan
			}

			err = step(pos, ch)

			return

		case stateRightBracket:
			if len(i.Right) == bracket.index {
				// execute expression
				var result string
				if result, err = eval(expr); err != nil {
					err = fmt.Errorf("%d-%d: eval: %w", bracket.start, pos, err)
					return
				}

				out += result

				state = stateScan
				err = step(pos, ch)
				return
			}

			if ch == 0 {
				err = ErrInterpolatorUnexpectedEndOfInput
				return
			}

			if i.Right[bracket.index] == byte(ch) {
				bracket.index++
				return
			}

			out += in[bracket.start:pos]
			state = stateScan
			err = step(pos, ch)
			return

		case stateExpr:
			if ch == 0 {
				err = ErrInterpolatorUnexpectedEndOfInput
				return
			}

			if len(i.Right) != 0 && ch == rune(i.Right[0]) {
				bracket.index = 0
				state = stateRightBracket
				err = step(pos, ch)
				return
			}

			expr += string(ch)
			return

		default:
			panic("unreachable")
		}
	}

	for pos, ch := range in {
		if err = step(pos, ch); err != nil {
			return
		}
	}

	err = step(len(in), 0)

	return
}
