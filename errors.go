package shaker

import (
	"errors"
)

var (
	ErrRessourceNotFound       = errors.New("ressource not found")
	ErrInvalidHandlerSignature = errors.New("invalid handler signature")
)
