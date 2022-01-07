package fehler

// exitCodeError contains an error and an exit code. Such objects fulfill the
// interface of wrapped errors.
type exitCodeError struct {
	err  error
	code int
}

func (err exitCodeError) Error() string {
	return err.err.Error()
}

func (err exitCodeError) Unwrap() error {
	return err.err
}

func (err exitCodeError) ExitCode() int {
	return err.code
}

// ExitCode returns a new error with attached exit code.
func ExitCode(code int, err error) error {
	return exitCodeError{
		err:  err,
		code: code,
	}
}
