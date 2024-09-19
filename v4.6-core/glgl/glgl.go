//go:build !tinygo && cgo

package glgl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"golang.org/x/exp/constraints"
)

// Version returns the running OpenGL version as a string.
func Version() string { return gl.GoStr(gl.GetString(gl.VERSION)) }

type Type uint32

const (
	Int8    Type = gl.BYTE
	Uint8   Type = gl.UNSIGNED_BYTE
	Int16   Type = gl.SHORT
	Uint16  Type = gl.UNSIGNED_SHORT
	Int32   Type = gl.INT
	Uint32  Type = gl.UNSIGNED_INT
	Float32 Type = gl.FLOAT
)

var (
	ErrStringNotNullTerminated = errors.New("string not null terminated")
)

// func CheckMemory() {
// 	var buf [4]uint32

// 	gl.GetIntegerv(gl.GPU_, &buf[0])

// }

// EnableDebugOutput writes debug output to log via glDebugMessageCallback.
// If log is nil then the default slog package logger is used.
func EnableDebugOutput(log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		attrs := []slog.Attr{
			slog.Uint64("source", uint64(source)),
			slog.Uint64("gltype", uint64(gltype)),
			slog.Uint64("severity", uint64(severity)),
			slog.Uint64("length", uint64(length)),
		}
		var level slog.Level
		switch gltype {
		case gl.DEBUG_TYPE_ERROR:
			level = slog.LevelError
		case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:
			level = slog.LevelWarn
		// case gl.DEBUG_TYPE_OTHER:
		// 	level = slog.LevelDebug
		default:
			level = slog.LevelInfo
		}
		log.LogAttrs(context.Background(), level, message, attrs...)
	}, nil)
}

// func debug() {
// 	const bufsize = 32 * 1024
// 	var buf [bufsize]byte
// 	gl.SOURCE
// 	gl.GetDebugMessageLog(1024, bufsize)
// }

// VertexArray ties data layout with vertex buffer(s).
// Is aware of data layout via VertexAttribPointer* calls.
// Vertex array parameters are client state, that is to say the GPU is unaware of it.
// Loosely speaking, a vertex array
type VertexArray struct {
	rid uint32
}

// AttribLayout is a low level configuration struct
// for adding vertex buffers attribute layouts to a vertex array object.
type AttribLayout struct {
	// The OpenGL program identifier.
	Program Program
	// Type is a OpenGL enum representing the underlying type. Valid types include
	// gl.FLOAT, gl.UNSIGNED_INT, gl.UNSIGNED_BYTE, gl.BYTE etc.
	Type Type
	// Name is the identifier of the attribute in the
	// vertex shader source code finished with a null terminator.
	Name string
	// Packing is a value between 1 and 4 and represents how many
	// of the type are present at the attribute location.
	//
	// Example:
	// When w orking with a vec3 attribute in the shader source code
	// with a gl.Float type, then the Packing is 3 since there are
	// 3 floats packed at each attribute location.
	Packing int
	// Stride is the distance in bytes between attributes in the buffer.
	Stride int
	// Offset is the starting offset with which to start
	// traversing the vertex buffer.
	Offset int
	// specifies whether fixed-point data values should be normalized (when true)
	// or converted directly as fixed-point values (when false) when they are accessed.
	// Usually left as false?
	Normalize bool
}

// NewVAO creates a vertex array object and binds it to current context.
func NewVAO() VertexArray {
	// Configure the Vertex Array Object.
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	return VertexArray{rid: vao}
}

func (vao VertexArray) Bind()   { gl.BindVertexArray(vao.rid) }
func (vao VertexArray) Unbind() { gl.BindVertexArray(0) }

func (vao VertexArray) AddAttribute(vbo VertexBuffer, layout AttribLayout) error {
	if !strings.HasSuffix(layout.Name, "\x00") {
		return ErrStringNotNullTerminated
	}
	if layout.Type == 0 || layout.Packing < 1 || layout.Packing > 4 {
		return errors.New("invalid argument")
	}
	vbo.Bind()
	vertAttrib := gl.GetAttribLocation(layout.Program.rid, gl.Str(layout.Name))
	if vertAttrib < 0 {
		return errors.New("vertex attribute not found:" + layout.Name[:len(layout.Name)-1])
	}
	gl.EnableVertexAttribArray(uint32(vertAttrib))
	// VAO: Vertex Array Object is bound to the vertex buffer on this call.
	// What this line is saying is that `vertAttrib`` index is going to be bound
	// to the current gl.ARRAY_BUFFER (vbo).
	// It also stores size, type, normalized, stride and pointer as vertex array
	// state, in addition to the current vertex array buffer object binding. https://registry.khronos.org/OpenGL-Refpages/gl4/html/glVertexAttribPointer.xhtml
	gl.VertexAttribPointerWithOffset(uint32(vertAttrib), int32(layout.Packing), uint32(layout.Type),
		layout.Normalize, int32(layout.Stride), uintptr(layout.Offset))
	return Err()
}

// BufferUsage is a required hint given to the GPU that provide a general description of
// how exactly the user will be using the buffer object so as to better optimize performance.
//
// There are two independent parts to the usage pattern:
// how the user will be reading/writing from/to the buffer,
// and how often the user will be changing it relative to the use of the data.
//
//   - DRAW: The user will be writing data to the buffer, but the user will not read it.
//   - READ: The user will not be writing data, but the user will be reading it back.
//   - COPY: The user will be neither writing nor reading the data.
//
// There are three hints for how frequently the user will be changing the buffer's data.
//
//   - STATIC: The user will set the data once.
//   - DYNAMIC: The user will set the data occasionally.
//   - STREAM: The user will be changing the data after every use. Or almost every use.
//
// DRAW is useful for, as the name suggests, drawing. The user is uploading data, but only the GL is reading it.
//
// READ is used when a buffer object is used as the destination for OpenGL commands.
// This could be rendering to a Buffer Texture, using arbitrary writes to buffer textures,
// doing a pixel transfer into a buffer object, using Transform Feedback, or any other OpenGL operation that writes to buffer objects.
//
// COPY is used when a buffer object is used to pass data from one place in OpenGL to another.
type BufferUsage uint32

// Buffer Usages. See BufferUsage documentation for detailed information.
const (
	StaticDraw  BufferUsage = gl.STATIC_DRAW
	StaticRead  BufferUsage = gl.STATIC_READ
	StaticCopy  BufferUsage = gl.STATIC_COPY
	DynamicDraw BufferUsage = gl.DYNAMIC_DRAW
	DynamicRead BufferUsage = gl.DYNAMIC_READ
	DynamicCopy BufferUsage = gl.DYNAMIC_COPY
	StreamDraw  BufferUsage = gl.STREAM_DRAW
	StreamRead  BufferUsage = gl.STREAM_READ
	StreamCopy  BufferUsage = gl.STREAM_COPY
)

// VertexBuffer contains bytes, no information on the layout or type.
// Buffer objects are said to be "server state", compared to vertex array parameters as "client state".
type VertexBuffer struct {
	// Renderer ID. If using OpenGL is the id set on buffer creation.
	rid uint32
}

// NewVertexBuffer creates a new vertex buffer and binds it.
func NewVertexBuffer[T any](usage BufferUsage, data []T) (VertexBuffer, error) {
	var vbo VertexBuffer
	vertexSize := unsafe.Sizeof(data[0])
	vertPtr := unsafe.Pointer(&data[0])
	gl.GenBuffers(1, &vbo.rid)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.rid)
	gl.BufferData(gl.ARRAY_BUFFER, int(vertexSize)*len(data), vertPtr, uint32(usage))
	return vbo, Err()
}

func (vbo VertexBuffer) Bind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.rid)
}
func (vbo VertexBuffer) Unbind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}
func (vbo VertexBuffer) Delete() {
	gl.DeleteBuffers(1, &vbo.rid)
}

type AccessUsage uint32

const WriteOnly, ReadOnly, ReadOrWrite AccessUsage = gl.WRITE_ONLY, gl.READ_ONLY, gl.READ_WRITE

// MapBufferData maps vertex buffer memory on the GPU to client space in the form
// of a slice.
func MapBufferData[T any](vbo VertexBuffer, length int, access AccessUsage) ([]T, error) {
	vertexSize := unsafe.Sizeof(*new(T))
	ptr := gl.MapNamedBufferRange(vbo.rid, 0, int(vertexSize)*length, uint32(access))
	err := Err()
	if err != nil {
		return nil, err
	}
	if ptr == nil {
		panic("got nil pointer from MapNamedBufferRange")
	}

	return unsafe.Slice((*T)(ptr), length), nil
}

func GetBufferData[T any](dst []T, vbo VertexBuffer) error {
	vertexSize := unsafe.Sizeof(dst[0])
	vertPtr := unsafe.Pointer(&dst[0])
	// gl.GetBufferDat
	gl.GetBufferSubData(gl.ARRAY_BUFFER, 0, len(dst)*int(vertexSize), vertPtr)
	// gl.GetNamedBufferSubData(vbo.rid, 0, len(dst)*int(vertexSize), vertPtr)
	return Err()
}

type IndexBuffer struct {
	// Renderer ID. If using OpenGL is the id set on buffer creation.
	rid uint32
}

func NewIndexBuffer(data []uint32) (IndexBuffer, error) {
	return newIndexBuffer(gl.STATIC_DRAW, data)
}

func newIndexBuffer(usage uint32, data []uint32) (IndexBuffer, error) {
	var ibo IndexBuffer
	const IndexSize = unsafe.Sizeof(data[0])
	vertPtr := unsafe.Pointer(&data[0])
	gl.GenBuffers(1, &ibo.rid)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo.rid)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(IndexSize)*len(data), vertPtr, usage)
	return ibo, Err()
}

func (vbo IndexBuffer) Bind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, vbo.rid)
}

func (vbo IndexBuffer) Unbind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
}

func (vbo IndexBuffer) Delete() {
	gl.DeleteBuffers(1, &vbo.rid)
}

type Texture struct {
	rid uint32
	// Usually GL_TEXTURE_2D.
	target uint32
	// Usually TEXTURE0.
	unit uint32
}

func MaxTextureSlots() (textureUnits int) {
	var tu int32
	ptr := &tu
	var p runtime.Pinner
	p.Pin(ptr)
	defer p.Unpin()
	gl.GetIntegerv(gl.MAX_TEXTURE_IMAGE_UNITS, ptr)
	return int(*ptr)
}

func MaxTextureBinded() (textureBounds int) {
	var tu int32
	ptr := &tu
	var p runtime.Pinner
	p.Pin(ptr)
	defer p.Unpin()
	gl.GetIntegerv(gl.MAX_COMBINED_TEXTURE_IMAGE_UNITS, ptr)
	return int(*ptr)
}

// Bind receives a slot onto which to bind from 0 to 32.
func (t Texture) Bind(activeSlot int) {
	gl.ActiveTexture(gl.TEXTURE0 + uint32(activeSlot))
	gl.BindTexture(t.target, t.rid)
}

//	func (t Texture) Unbind() {
//		if err := Err(); err != nil {
//			panic(err)
//		}
//		gl.ActiveTexture(0)
//		gl.BindTexture(t.target, 0)
//		if err := Err(); err != nil {
//			panic(err)
//		}
//	}
func (t Texture) Delete() {
	// gl.BindTexture(t.target, 0)
	// if err := Err(); err != nil {
	// 	panic(err)
	// }
	gl.DeleteTextures(1, &t.rid)
	if err := Err(); err != nil {
		panic(err)
	}
}

type TextureType uint32

const Texture2D TextureType = gl.TEXTURE_2D

// TextureImgConfig builds an image based texture.
// Below are common formats:
// - Base internal. i.e: gl.RED, gl.RG, gl.RGBA, gl.DEPTH_COMPONENT
// - Sized internal: gl.R8, gl.R16, gl.RGB4, gl.R32F, gl.RGBA32F.
type TextureImgConfig struct {
	// Specifies the target texture. Must be one of:
	//  GL_TEXTURE_2D, GL_PROXY_TEXTURE_2D, GL_TEXTURE_1D_ARRAY, GL_PROXY_TEXTURE_1D_ARRAY, GL_TEXTURE_RECTANGLE, GL_PROXY_TEXTURE_RECTANGLE, GL_TEXTURE_CUBE_MAP_POSITIVE_X, GL_TEXTURE_CUBE_MAP_NEGATIVE_X, GL_TEXTURE_CUBE_MAP_POSITIVE_Y, GL_TEXTURE_CUBE_MAP_NEGATIVE_Y, GL_TEXTURE_CUBE_MAP_POSITIVE_Z, GL_TEXTURE_CUBE_MAP_NEGATIVE_Z, or GL_PROXY_TEXTURE_CUBE_MAP.
	Type   TextureType
	Width  int
	Height int
	Border int32
	// Specifies the number of color components in the texture.
	// Can use base, sized or compressed internal formats: See [TextureImgConfig] for more.
	// If not set uses Format.
	InternalFormat int32
	// Specifies format of the pixel data. Accepts:
	//  GL_RED, GL_RG, GL_RGB, GL_BGR, GL_RGBA, GL_BGRA, GL_RED_INTEGER, GL_RG_INTEGER, GL_RGB_INTEGER, GL_BGR_INTEGER, GL_RGBA_INTEGER, GL_BGRA_INTEGER, GL_STENCIL_INDEX, GL_DEPTH_COMPONENT, GL_DEPTH_STENCIL.
	Format uint32

	// Specifies the data type of the pixel data. Accepts
	//   GL_UNSIGNED_BYTE, GL_BYTE, GL_UNSIGNED_SHORT, GL_SHORT, GL_UNSIGNED_INT, GL_INT, GL_HALF_FLOAT, GL_FLOAT, GL_UNSIGNED_BYTE_3_3_2, GL_UNSIGNED_BYTE_2_3_3_REV, GL_UNSIGNED_SHORT_5_6_5, GL_UNSIGNED_SHORT_5_6_5_REV, GL_UNSIGNED_SHORT_4_4_4_4, GL_UNSIGNED_SHORT_4_4_4_4_REV, GL_UNSIGNED_SHORT_5_5_5_1, GL_UNSIGNED_SHORT_1_5_5_5_REV, GL_UNSIGNED_INT_8_8_8_8, GL_UNSIGNED_INT_8_8_8_8_REV, GL_UNSIGNED_INT_10_10_10_2, and GL_UNSIGNED_INT_2_10_10_10_REV.
	Xtype uint32
	// Magnification filtering. gl.NEAREST or gl.LINEAR.
	MagFilter int32
	// Minification filtering. gl.NEAREST or gl.LINEAR.
	MinFilter int32
	// Textures coordinates usually range from (0,0) to (1,1). Wrap indicates
	// how OpenGL is to repeat the texture outside this range.
	// gl.REPEAT, gl.MIRRORED_REPEAT, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_BORDER.
	Wrap int32

	// Specifies a token indicating the type of access that will be performed on the image.
	Access AccessUsage
	// Optional parameters below

	Layered bool
	Layer   int32
	// Specifies the level-of-detail number. Level 0 is the base image level. If target is GL_TEXTURE_RECTANGLE or GL_PROXY_TEXTURE_RECTANGLE, level must be 0.
	Level int32
	// Specifies the unit on which to bind the image onto the texture.
	// This is the binding point for image2D uniforms.
	ImageUnit uint32

	// TextureUnit is the texture unit onto which the texture is loaded (glActiveTexture).
	// TextureUnit starts at 0 and goes all the way up to MaxTextureSlots().
	TextureUnit int
}

func (cfg TextureImgConfig) PixelSize() int {
	var mul, sz int
	switch cfg.Format {
	case gl.RED, gl.RED_INTEGER:
		mul = 1
	case gl.RG, gl.RG_INTEGER:
		mul = 2
	case gl.RGB, gl.RGB_INTEGER:
		mul = 3
	case gl.RGBA, gl.RGBA_INTEGER:
		mul = 4
	default:
		panic("unsupported format. file an issue or PR with its addition!")
	}
	switch cfg.Xtype {
	case gl.FLOAT, gl.INT:
		sz = 4
	default:
		panic("unsupported xtype. file an issue or PR with its addition!")
	}
	return mul * sz
}
func assertImgSameSize[T any](cfg TextureImgConfig, data []T) error {
	sz := cfg.PixelSize() * cfg.Width * cfg.Height
	bufSize := len(data) * int(unsafe.Sizeof(data[0])) // If you are getting panic here please use nil as data.
	if sz != bufSize {
		return errors.New("data size not match to be allocated")
	}
	return nil
}

// NewTextureFromImage creates a new Texture from an image and binds it to the current context.
func NewTextureFromImage[T any](cfg TextureImgConfig, data []T) (Texture, error) {
	var outTexture uint32
	var ptr unsafe.Pointer = nil
	if data != nil {
		if err := assertImgSameSize(cfg, data); err != nil {
			return Texture{}, err
		}
		ptr = unsafe.Pointer(&data[0])
	}
	gl.GenTextures(1, &outTexture)
	tex := Texture{
		rid:    outTexture,
		target: uint32(cfg.Type),
		unit:   uint32(gl.TEXTURE0 + cfg.TextureUnit),
	}
	tex.Bind(cfg.TextureUnit)

	internalFormat := zdefault(cfg.InternalFormat, int32(cfg.Format))
	gl.TexImage2D(tex.target, cfg.Level, internalFormat, int32(cfg.Width), int32(cfg.Height),
		cfg.Border, cfg.Format, cfg.Xtype, ptr)
	// Use default values since OpenGL does not do sane defaults: https://medium.com/@daniel.coady/compute-shaders-in-opengl-4-3-d1c741998c03
	gl.TexParameteri(tex.target, gl.TEXTURE_MAG_FILTER, zdefault(cfg.MagFilter, gl.NEAREST))
	gl.TexParameteri(tex.target, gl.TEXTURE_MIN_FILTER, zdefault(cfg.MinFilter, gl.NEAREST))
	gl.TexParameteri(tex.target, gl.TEXTURE_WRAP_S, zdefault(cfg.Wrap, gl.REPEAT))
	gl.TexParameteri(tex.target, gl.TEXTURE_WRAP_T, zdefault(cfg.Wrap, gl.REPEAT))

	// For following call: format specifies the format that is to be used when performing
	// formatted stores into the image from shaders. format must be compatible with the
	// texture's internal format and must be one of the formats listed in the following table.
	gl.BindImageTexture(cfg.ImageUnit, outTexture, cfg.Level, cfg.Layered, cfg.Layer,
		uint32(cfg.Access), uint32(internalFormat))
	return tex, Err()
}

// SetImage2D sets an existing texture's values on the GPU.
func SetImage2D[T any](tex Texture, cfg TextureImgConfig, data []T) error {
	var ptr unsafe.Pointer = nil
	if data != nil {
		ptr = unsafe.Pointer(&data[0])
	}
	internalFormat := zdefault(cfg.InternalFormat, int32(cfg.Format))
	gl.TextureBarrier()
	gl.TexImage2D(tex.unit, cfg.Level, internalFormat,
		int32(cfg.Width), int32(cfg.Height), cfg.Border, cfg.Format, cfg.Xtype, ptr)
	return Err()
}

func GetImage[T any](dst []T, tex Texture, cfg TextureImgConfig) error {
	if len(dst) == 0 {
		return errors.New("dst cannot be nil or zero length")
	}
	if err := assertImgSameSize(cfg, dst); err != nil {
		return err
	}
	gl.TextureBarrier()
	gl.GetTexImage(tex.target, cfg.Level, cfg.Format, cfg.Xtype, unsafe.Pointer(&dst[0]))
	return Err()
}

// ClearErrors clears all of OpenGL's errors in it's log.
func ClearErrors() {
	i := 0
	for gl.GetError() != gl.NO_ERROR {
		i++
		if i > 2000 {
			panic("forever loop in clear errors. Has the context terminated?")
		}
	}
}

// Err returns a non-nil glErrors if errors are foudn in OpenGL's GetError buffer.
// After a call to Err no more errors should be returned until the next GL call.
func Err() error {
	code := gl.GetError()
	if code == gl.NO_ERROR {
		return nil
	}
	errs := glErrors{glError(code)}
	for {
		code = gl.GetError()
		if code == gl.NO_ERROR {
			return errs
		}
		errs = append(errs, glError(code))
		if len(errs) > 61 {
			lastIdx := len(errs) - 1
			return fmt.Errorf("possible forever loop in Err. Context may be terminated. errs[0]=%v, errs[%d]=%v(%d)", errs[0].String(), lastIdx, errs[lastIdx].String(), uint32(errs[lastIdx]))
		}
	}
}

// Refactor this so we can unwrap a error type local to glgl.
type glErrors []glError

func (ge glErrors) Error() (errstr string) {
	if len(ge) == 0 {
		return "no gl errors"
	}
	for i, e := range ge {
		errstr += e.String()
		if i != len(ge)-1 {
			errstr += "; "
		}
	}
	return errstr
}

type glError uint32

func (ge glError) String() (s string) {
	switch ge {
	case gl.INVALID_ENUM:
		s = "invalid enum"
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		s = "invalid framebuffer operation"
	case gl.INVALID_INDEX:
		s = "invalid index"
	case gl.INVALID_OPERATION:
		s = "invalid operation"
	case gl.INVALID_VALUE:
		s = "invalid value"
	default:
		s = "glError(" + strconv.Itoa(int(ge)) + ")"
	}
	return s
}

// zdefault is a helper function that returns the Default
// value if got is zero.
func zdefault[T constraints.Integer](got, Default T) T {
	if got == 0 {
		return Default
	}
	return got
}
