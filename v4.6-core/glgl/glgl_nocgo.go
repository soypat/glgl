//go:build tinygo || !cgo

package glgl

import (
	"errors"

	"github.com/go-gl/glfw/v3.0/glfw"
)

var errNoCgo = errors.New("glgl needs Cgo")

func InitWithCurrentWindow33(cfg WindowConfig) (*glfw.Window, func(), error) {
	return nil, nil, errNoCgo
}
