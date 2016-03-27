package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpError(t *testing.T) {
	t.Parallel()

	err := new(HttpError)
	err.Message = "Bad Request"
	err.Status = 400

	assert.Equal(t, "Bad Request", err.Error())
}

func TestBadRedisValues(t *testing.T) {
	t.Parallel()

	err := new(BadRedisValues)
	err.CacheKey = "test:foof"
	assert.Equal(t, "Redis query for key test:foof returned a malformatted value", err.Error())
}
