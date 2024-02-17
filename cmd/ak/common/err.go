package common

const (
	NotFoundExitCode   = 10
	NotAMemberExitCode = NotFoundExitCode
)

type ExitCodeError struct {
	Err  error
	Code int
}

func (e ExitCodeError) Error() string {
	return e.Err.Error()
}

var _ error = ExitCodeError{}

func NewExitCodeError(code int, err error) ExitCodeError {
	return ExitCodeError{Err: err, Code: code}
}
