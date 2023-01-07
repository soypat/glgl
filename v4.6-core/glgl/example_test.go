package glgl_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"golang.org/x/exp/slog"
)

const v = `#shader vertex
#version 330

out float v_myOutput;
in float myInput;
uniform float addThis;

void main() {
	v_myOutput = myInput + addThis;
}
#shader fragment
#version 330
in float v_myOutput;
out float myOutput;
void main() {
	myOutput = v_myOutput;
}`

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func Example_helloComputeWorld() {
	window, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:         "Hello Compute World",
		Width:         80,
		Height:        80,
		NotResizable:  true,
		Version:       [2]int{4, 6},
		OpenGLProfile: glfw.OpenGLCoreProfile,
		ForwardCompat: true,
	})
	if err != nil {
		slog.Error("failed to initialize", err)
		os.Exit(1)
	}
	defer glfw.Terminate()
	defer window.Destroy()
	source, err := glgl.ParseCombined(strings.NewReader(`
#shader vertex
#version 410

in float inValue;

out float outValue;

void main() {
	float modifiedValue = inValue + 10.0;
	gl_Position = vec4(modifiedValue, 0.0, 0.0, 1.0);
	outValue = modifiedValue;
}
#shader fragment
#version 410

in float outValue;
out vec4 fragval;
void main() {
	fragval = vec4(outValue,0.0,0.0,1.0);
}
`))
	prog, err := glgl.NewProgram(source)
	if err != nil {
		slog.Error("failed to initialize glfw", err)
		os.Exit(1)
	}
	defer prog.Delete()
	prog.Bind()
	vao := glgl.NewVAO() // Configure the Vertex Array Object.
	// Create the Position Buffer Object.
	input := []float32{1, 2, 3}
	output := make([]float32, len(input))
	inputBO, err := glgl.NewVertexBuffer(glgl.StaticDraw, input)
	if err != nil {
		slog.Error("creating positions vertex buffer", err)
		return
	}
	err = vao.AddAttribute(inputBO, glgl.AttribLayout{
		Program: prog,
		Type:    glgl.Float32,
		Name:    "inValue\x00",
		Packing: 1,
		Stride:  1,
	})
	if err != nil {
		slog.Error("adding input attribute", err)
		return
	}
	outputBO, err := glgl.NewVertexBuffer(glgl.StaticRead, output)
	if err != nil {
		slog.Error("creating positions vertex buffer", err)
		return
	}
	err = prog.BindFrag("finalValue\x00")
	if err != nil {
		slog.Error("binding frag", err)
		return
	}
	err = vao.AddAttribute(outputBO, glgl.AttribLayout{
		Program: prog,
		Type:    glgl.Float32,
		Name:    "outValue\x00",
		Packing: 1,
		Stride:  1,
	})
	if err != nil {
		slog.Error("adding output attribute", err)
		return
	}
	const uniform = 1
	// Set uniform variable `u_color` in source code.
	err = prog.SetUniform1f("addThis\x00", uniform)
	if err != nil {
		slog.Error("setting `addThis` uniform", err)
	}
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(input)))
	err = glgl.GetBufferData(output, outputBO)
	if err != nil {
		slog.Error("getting buffer data", err)
	}
	fmt.Println(output)
	// Output:
	// [2, 3, 4]
}

func Example_coloredSquare() {
	// Very basic index buffer example.
	const shader = `
#shader vertex
#version 330

in vec3 vert;

void main() {
	gl_Position = vec4(vert.xyz, 1.0);
}

#shader fragment
#version 330

out vec4 outputColor;

uniform vec4 u_color;

void main() {
	outputColor = u_color;
}`

	// Square with indices:
	// 3----2
	// |    |
	// 0----1
	var positions = []float32{
		-0.5, -0.5, // 0
		0.5, -0.5, // 1
		0.5, 0.5, // 2
		-0.5, 0.5, //3
	}
	var indices = []uint32{
		0, 1, 2, // Lower right triangle.
		0, 2, 3, // Upper left triangle.
	}
	window, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:         "Index Buffers",
		Width:         800,
		Height:        800,
		NotResizable:  true,
		Version:       [2]int{4, 6},
		OpenGLProfile: glfw.OpenGLCoreProfile,
		ForwardCompat: true,
	})
	defer glfw.Terminate()
	fmt.Println("OpenGL version", glgl.Version())

	// Separate vertex and fragment shaders from source code.
	source, err := glgl.ParseCombined(strings.NewReader(shader))
	if err != nil {
		slog.Error("parse combined source fail", err)
		return
	}

	// Configure the vertex and fragment shaders
	program, err := glgl.NewProgram(source)
	if err != nil {
		slog.Error("compile fail", err)
		return
	}
	defer program.Delete()
	program.Bind()

	err = program.BindFrag("outputColor\x00")
	if err != nil {
		slog.Error("program bind frag fail", err)
		return
	}
	// Configure the Vertex Array Object.
	vao := glgl.NewVAO()

	// Create the Position Buffer Object.
	vbo, err := glgl.NewVertexBuffer(glgl.StaticDraw, positions)
	if err != nil {
		slog.Error("creating positions vertex buffer", err)
		return
	}
	err = vao.AddAttribute(vbo, glgl.AttribLayout{
		Program: program,
		Type:    gl.FLOAT,
		Name:    "vert\x00",
		Packing: 2,
		Stride:  2 * 4, // 2 floats, each 4 bytes wide.
	})
	if err != nil {
		slog.Error("adding attribute vert", err)
		return
	}

	// Create Index Buffer Object.
	_, err = glgl.NewIndexBuffer(indices)
	if err != nil {
		slog.Error("creating index buffer", err)
		return
	}

	// Set uniform variable `u_color` in source code.
	err = program.SetUniformName4f("u_color\x00", 0.2, 0.3, 0.8, 1)
	if err != nil {
		slog.Error("creating index buffer", err)
		return
	}
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

		program.SetUniformName4f("u_color\x00", float32(time.Now().UnixMilli()%1000)/1000, .5, .3, 1)
		// Maintenance
		glfw.SwapInterval(1) // Can prevent epilepsy for high frequency
		window.SwapBuffers()
		glfw.PollEvents()
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
		}
	}
	// Output:
	// None.
}
