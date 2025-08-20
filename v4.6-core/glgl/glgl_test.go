package glgl_test

import (
	"testing"

	"github.com/soypat/glgl/v4.6-core/glgl"
)

func TestWindow(t *testing.T) {
	window, term, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:         "My great window",
		NotResizable:  false,
		Version:       [2]int{3, 3},
		OpenGLProfile: glgl.ProfileCore,
		ForwardCompat: true,
		Width:         1,
		Height:        1,
	})
	if err != nil {
		t.Log(err)
		t.Skip()
	}
	term()
	_ = window
}
