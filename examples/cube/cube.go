// Command cube demonstrates simple Glow binding usage.
package main

import (
	"errors"
	"fmt"
	"github.com/errcw/glow/gl/3.3/gl"
	"github.com/errcw/glow/glt"
	"github.com/fzipp/geom"
	glfw "github.com/go-gl/glfw3"
	"runtime"
	"strings"
	"unsafe"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func glfwErrorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("GLFW error %v: %v\n", err, desc)
}

func checkerr() {
	e := gl.GetError()
	if e != gl.NO_ERROR {
		panic(e)
	}
}

func main() {
	// Initialize GLFW for window management
	glfw.SetErrorCallback(glfwErrorCallback)
	if !glfw.Init() {
		panic("failed to initialize glfw")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True)    // Necessary for OS X
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile) // Necessary for OS X
	window, err := glfw.CreateWindow(640, 480, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := glt.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Configure the scene
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	var projection geom.Mat4
	projection.Perspective(50.0, 640.0/480.0, 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, glt.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, true, toRowMajorFloats(&projection))
	checkerr()

	var camera geom.Mat4
	camera.LookAt(geom.V3(3, 3, 3), geom.V3(0, 0, 0), geom.V3(0, 1, 0))
	cameraUniform := gl.GetUniformLocation(program, glt.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, true, toRowMajorFloats(&camera))
	checkerr()

	var model geom.Mat4
	model.ID()
	modelUniform := gl.GetUniformLocation(program, glt.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, true, toRowMajorFloats(&model))
	checkerr()

	/*
		var texture uint32
		gl.GenTextures(1, &texture)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		textureUniform := gl.GetUniformLocation(program, glt.Str("tex\x00"))
		gl.Uniform1i(textureUniform, 0)
	*/

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	checkerr()
	gl.BindVertexArray(vao)
	checkerr()

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	checkerr()
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	checkerr()
	gl.BufferData(gl.ARRAY_BUFFER, len(cube)*4, glt.Ptr(cube), gl.STATIC_DRAW)
	checkerr()

	vertAttrib := uint32(gl.GetAttribLocation(program, glt.Str("vert\x00")))
	checkerr()
	gl.EnableVertexAttribArray(vertAttrib)
	checkerr()
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, 0)
	checkerr()

	texCoordAttrib := uint32(gl.GetAttribLocation(program, glt.Str("vertTexCoord\x00")))
	checkerr()
	gl.EnableVertexAttribArray(texCoordAttrib)
	checkerr()
	gl.VertexAttribPointer(vertAttrib, 2, gl.FLOAT, false, 5*4, 3*4)
	checkerr()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.Enable(gl.DEPTH_TEST)
	checkerr()
	gl.DepthFunc(gl.LESS)
	checkerr()

	gl.Viewport(0, 0, 640, 480)
	checkerr()
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
	checkerr()

	frames := 0
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		checkerr()

		gl.UseProgram(program)
		checkerr()
		gl.BindVertexArray(vao)
		checkerr()

		gl.ValidateProgram(program)
		var status int32
		gl.GetProgramiv(program, gl.VALIDATE_STATUS, &status)
		fmt.Println("Got validate status", status)

		gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3)
		checkerr()

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()

		frames++
	}
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	checkerr()

	gl.AttachShader(program, vertexShader)
	checkerr()
	gl.AttachShader(program, fragmentShader)
	checkerr()
	gl.LinkProgram(program)
	checkerr()

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	checkerr()
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, glt.Str(log))

		return 0, errors.New(fmt.Sprintf("failed to link program: %v", log))
	}

	//gl.DeleteShader(vertexShader)
	//gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	checkerr()

	csource := glt.Str(source)
	gl.ShaderSource(shader, 1, &csource, nil)
	checkerr()
	gl.CompileShader(shader)

	checkerr()
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	checkerr()
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, glt.Str(log))

		return 0, errors.New(fmt.Sprintf("failed to compile %v: %v", source, log))
	}

	return shader, nil
}

func toRowMajorFloats(m *geom.Mat4) *float32 {
	return (*float32)(unsafe.Pointer(m))
}

var vertexShader string = `
#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 vert;
in vec2 vertTexCoord;

out vec2 fragTexCoord;

void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330

//uniform sampler2D tex;

in vec2 fragTexCoord;

out vec4 finalColor;

void main() {
    //finalColor = texture(tex, fragTexCoord);
    finalColor = vec4(0, 0, 0, 0);
}
` + "\x00"

var cube = []float32{
	//  X, Y, Z, U, V
	// Bottom
	-1.0, -1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,

	// Top
	-1.0, 1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,

	// Front
	-1.0, -1.0, 1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,

	// Back
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 1.0,

	// Left
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,

	// Right
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
}
