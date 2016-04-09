package errors

type HttpError struct {
	Message string
	Status  int
}

func (e *HttpError) Error() string {
	return e.Message
}
