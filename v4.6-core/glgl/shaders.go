package glgl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
)

// ParseCombinedBasic parses a file with vertex and fragment #shader pragmas inspired
// by [The Cherno]'s take on shader file segmenting. This method of writing
// shaders lets one keep vertex and fragment shader source code in the same file:
//
//	// Anything above first #shader pragma is ignored.
//	#shader vertex
//	void main() {
//	    gl_Position = vec4(1.0,0.0,0.0, 1.0);
//	}
//
//	#shader fragment
//	void main() {
//	    gl_Frag = gl_Position/2;
//	}
//
// [The Cherno]: https://www.youtube.com/watch?v=2pv0Fbo-7ms&list=PLlrATfBNZ98foTJPJ_Ev03o2oq3-GGOS2&index=9&t=724s&ab_channel=TheCherno
func ParseCombined(r io.Reader) (ss ShaderSource, err error) {
	const (
		shaderNone = iota
		shaderVertex
		shaderFragment
		shaderNum
	)
	nothing := bytes.NewBuffer(nil)
	vertexBuf := bytes.NewBuffer(nil)
	fragBuf := bytes.NewBuffer(nil)
	buffers := [shaderNum]*bytes.Buffer{
		shaderNone:     nothing,
		shaderVertex:   vertexBuf,
		shaderFragment: fragBuf,
	}
	scanner := bufio.NewScanner(r)
	currentShader := shaderNone
	for scanner.Scan() {
		line := scanner.Bytes()
		if currentShader != shaderNone && !bytes.HasPrefix(bytes.TrimSpace(line), []byte("#shader ")) {
			buffers[currentShader].Write(line)
			buffers[currentShader].WriteByte('\n')
			continue
		}
		if bytes.Index(line, []byte("vertex")) > 0 {
			currentShader = shaderVertex
		} else if bytes.Index(line, []byte("fragment")) > 0 {
			currentShader = shaderFragment
		}
	}

	// Null terminated strings.
	if vertexBuf.Len() > 0 {
		vertexBuf.WriteByte(0)
	}
	if fragBuf.Len() > 0 {
		fragBuf.WriteByte(0)
	}
	return ShaderSource{Vertex: vertexBuf.String(), Fragment: fragBuf.String()}, scanner.Err()
}

// CompileBasic compiles two OpenGL vertex and fragment shaders
// and returns a program with the current OpenGL context.
// It returns an error if compilation, linking or validation fails.
func compileBasic(vertexSrcCode, fragmentSrcCode string) (program uint32, err error) {
	if !strings.HasSuffix(vertexSrcCode, "\x00") {
		return 0, errors.New("vertex shader source has no null terminator")
	}
	if !strings.HasSuffix(fragmentSrcCode, "\x00") {
		return 0, errors.New("fragment shader source has no null terminator")
	}
	program = gl.CreateProgram()
	vid, err := compile(gl.VERTEX_SHADER, vertexSrcCode)
	if err != nil {
		return 0, fmt.Errorf("vertex shader compile: %w", err)
	}
	fid, err := compile(gl.FRAGMENT_SHADER, fragmentSrcCode)
	if err != nil {
		return 0, fmt.Errorf("fragment shader compile: %w", err)
	}
	gl.AttachShader(program, vid)
	gl.AttachShader(program, fid)
	gl.LinkProgram(program)
	log := ivLog(program, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog)
	if len(log) > 0 {
		return 0, fmt.Errorf("link failed: %v", log)
	}
	// We should technically call DetachShader after linking... https://www.youtube.com/watch?v=71BLZwRGUJE&list=PLlrATfBNZ98foTJPJ_Ev03o2oq3-GGOS2&index=7&ab_channel=TheCherno
	gl.ValidateProgram(program)
	log = ivLog(program, gl.VALIDATE_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog)
	if len(log) > 0 {
		return 0, fmt.Errorf("validation failed: %v", log)
	}

	// We can clean up.
	gl.DeleteShader(vid)
	gl.DeleteShader(fid)
	return program, nil
}

func compile(shaderType uint32, sourceCode string) (uint32, error) {
	id := gl.CreateShader(shaderType)
	csources, free := gl.Strs(sourceCode)
	gl.ShaderSource(id, 1, csources, nil)
	free()
	gl.CompileShader(id)

	// We now check the errors during compile, if there were any.
	log := ivLog(id, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog)
	if len(log) > 0 {
		return 0, errors.New(log)
	}
	return id, nil
}

// ivLog is a helper function for extracting log data
// from a Shader compilation step or program linking.
//
//	log := ivLog(id, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog)
//	if len(log) > 0 {
//		return 0, errors.New(log)
//	}
func ivLog(id, plName uint32, getIV func(program uint32, pname uint32, params *int32), getInfo func(program uint32, bufSize int32, length *int32, infoLog *uint8)) string {
	var iv int32
	getIV(id, plName, &iv)
	if iv == gl.FALSE {
		var logLength int32
		getIV(id, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		getInfo(id, logLength, &logLength, &log[0])
		return string(log[:len(log)-1]) // we exclude the last null character.
	}
	return ""
}
