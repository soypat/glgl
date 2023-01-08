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
// ParseCombined performs no calls to the GL.
//
// [The Cherno]: https://www.youtube.com/watch?v=2pv0Fbo-7ms&list=PLlrATfBNZ98foTJPJ_Ev03o2oq3-GGOS2&index=9&t=724s&ab_channel=TheCherno
func ParseCombined(r io.Reader) (ss ShaderSource, err error) {
	const (
		shaderNone = iota
		shaderVertex
		shaderFragment
		shaderCompute
		shaderHeader
		shaderNum
	)
	nothing := bytes.NewBuffer(nil)
	vertexBuf := bytes.NewBuffer(nil)
	fragBuf := bytes.NewBuffer(nil)
	computeBuf := bytes.NewBuffer(nil)
	includeBuf := bytes.NewBuffer(nil)
	buffers := [shaderNum]*bytes.Buffer{
		shaderNone:     nothing,
		shaderVertex:   vertexBuf,
		shaderFragment: fragBuf,
		shaderCompute:  computeBuf,
		shaderHeader:   includeBuf,
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
		got := bytes.Fields(line)
		if len(got) != 2 {
			continue
		}
		switch string(got[1]) {
		case "includeashead":
			currentShader = shaderHeader
		case "vertex":
			currentShader = shaderVertex
		case "fragment", "pixel":
			currentShader = shaderFragment
		case "compute":
			currentShader = shaderCompute
		default:
			return ShaderSource{}, errors.New("unexpected #shader pragma value:" + string(got[1]))
		}
	}
	isrc := includeBuf.Bytes()
	var vsrc, fsrc, csrc []byte
	if vertexBuf.Len() > 0 {
		vsrc = append(vsrc, isrc...)
		vertexBuf.WriteByte(0)
		vsrc = append(vsrc, vertexBuf.Bytes()...)
	}
	if computeBuf.Len() > 0 {
		csrc = append(csrc, isrc...)
		computeBuf.WriteByte(0)
		csrc = append(csrc, computeBuf.Bytes()...)
	}
	if fragBuf.Len() > 0 {
		fsrc = append(fsrc, isrc...)
		fragBuf.WriteByte(0)
		fsrc = append(fsrc, fragBuf.Bytes()...)
	}
	return ShaderSource{
			Vertex:   string(vsrc),
			Fragment: string(fsrc),
			Compute:  string(csrc),
			Include:  string(isrc),
		},
		scanner.Err()
}

// CompileBasic compiles two OpenGL vertex and fragment shaders
// and returns a program with the current OpenGL context.
// It returns an error if compilation, linking or validation fails.
func compileSources(ss ShaderSource) (program uint32, err error) {
	// Note: glDeleteShader only flags a shader for deletion.
	// They are not deleted until they are detached from the program.
	// Beware: multiple calls to glDeleteShader on the same shader will cause an error on GL's side.
	program = gl.CreateProgram()
	if program == 0 {
		if err := Err(); err != nil {
			return 0, fmt.Errorf("got invalid program id: %w", err)
		}
		return 0, fmt.Errorf("silently got invalid program ID")
	}
	if len(ss.Vertex) > 0 {
		vid, err := compile(gl.VERTEX_SHADER, ss.Vertex)
		if err != nil {
			return 0, fmt.Errorf("vertex shader compile: %w", err)
		}
		gl.AttachShader(program, vid)
		// We can clean up.
		defer gl.DeleteShader(vid)
	}
	if len(ss.Fragment) > 0 {
		fid, err := compile(gl.FRAGMENT_SHADER, ss.Fragment)
		if err != nil {
			return 0, fmt.Errorf("fragment shader compile: %w", err)
		}
		gl.AttachShader(program, fid)
		defer gl.DeleteShader(fid)
	}
	if len(ss.Compute) > 0 {
		cid, err := compile(gl.COMPUTE_SHADER, ss.Compute)
		if err != nil {
			return 0, fmt.Errorf("compute shader compile: %w", err)
		}
		gl.AttachShader(program, cid)
		defer gl.DeleteShader(cid)
	}

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

	return program, nil
}

func compile(shaderType uint32, sourceCodes ...string) (uint32, error) {
	if err := Err(); err != nil {
		return 0, fmt.Errorf("unhandled error before compiling: %w", err)
	}
	var sourceLengths []int32
	for i := range sourceCodes {
		if !strings.HasSuffix(sourceCodes[i], "\x00") {
			return 0, errors.New("source missing null terminator")
		}
		sourceLengths = append(sourceLengths, int32(len(sourceCodes[i])))
	}

	id := gl.CreateShader(shaderType)
	if id == 0 {
		if err := Err(); err != nil {
			return 0, fmt.Errorf("got invalid shader ID: %w", err)
		}
		return 0, fmt.Errorf("silently got invalid shader id 0")
	}
	fmt.Println("id start", id)
	csources, free := gl.Strs(sourceCodes...)
	gl.ShaderSource(id, int32(len(sourceCodes)), csources, &sourceLengths[0])
	free()

	gl.CompileShader(id)
	if err := Err(); err != nil {
		return 0, fmt.Errorf("error after compiling shader: %w", err)
	}
	// We now check the errors during compile, if there were any.
	fmt.Println("id before log", id)
	log := ivLog(id, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog)
	fmt.Println("id after log", id)
	if len(log) > 0 {
		return 0, errors.New(log)
	}
	// if !gl.IsShader(id) {
	// 	return 0, errors.New("shader ID unexpectedly does not correspond to shader")
	// }
	return id, Err()
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
		if logLength == 0 {
			// panic(fmt.Sprintf("unexpected false iv for plName=0x%#X with no log", plName))
			fmt.Print("TODO: fix ivLog in shaders.go")
			return ""
		}
		log := make([]byte, logLength)
		getInfo(id, logLength, &logLength, &log[0])
		return string(log[:len(log)-1]) // we exclude the last null character.
	}
	return ""
}
