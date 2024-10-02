//go:build tinygo || !cgo

package glgl

import (
	"errors"
	"log/slog"
)

type Window struct{}

var errNoCgo = errors.New("glgl needs cgo")

func InitWithCurrentWindow33(cfg WindowConfig) (*Window, func(), error) {
	return nil, nil, errNoCgo
}

// MaxComputeInvoc returns maximum number of invocations/warps per workgroup on the local GPU. The GL context must be actual.
func MaxComputeInvocations() int {
	return 32
}

func Version() string { return errNoCgo.Error() }

func EnableDebugOutput(log *slog.Logger) {}

func compileSources(ss ShaderSource) (program Program, err error) {
	return Program{}, errNoCgo
}

func Err() error { return errNoCgo }

func (p Program) Bind()   {}
func (p Program) Unbind() {}

const (
	ProfileAny int = iota
	ProfileCore
	ProfileCompat
)
