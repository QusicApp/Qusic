package util

// StringError is a string that implements the error interface and can be made a constant
type StringError string

func (err StringError) Error() string {
	return string(err)
}
