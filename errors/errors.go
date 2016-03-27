package errors

import "fmt"

type HttpError struct {
	Message string
	Status  int
}

func (e *HttpError) Error() string {
	return e.Message
}

type BadRedisValues struct {
	CacheKey string
}

func (e *BadRedisValues) Error() string {
	return fmt.Sprintf("Redis query for key %s returned a malformatted value", e.CacheKey)
}
