package task

import "errors"

var ErrTimeout = errors.New("receive timeout")

var ErrInterrupt = errors.New("receive interrupt")
