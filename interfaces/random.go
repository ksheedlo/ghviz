package interfaces

import (
	"crypto/rand"
	"math/big"
)

type RandomTagger interface {
	RandomTag() (string, error)
}

type RandomTaggerFunc func() (string, error)

func (f RandomTaggerFunc) RandomTag() (string, error) {
	return f()
}

var RandomTag = RandomTaggerFunc(func() (string, error) {
	tagInt, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return "", err
	}
	return tagInt.Text(36), nil
})
