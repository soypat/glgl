package glgl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// Vertex and Fragment are null terminated strings with source code.
type ShaderSource struct {
	// Vertex and Fragment are null terminated strings with source code.
	Vertex   string
	Fragment string
	Compute  string
	Include  string
}

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
// `compute` and `includeashead` are also valid #shader pragmas.
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
	}, scanner.Err()
}
