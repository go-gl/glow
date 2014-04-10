Glow
======

Glow is an OpenGL binding generator for Go. Glow parses the [OpenGL XML API registry](https://cvs.khronos.org/svn/repos/ogl/trunk/doc/registry/public/api/) to produce a machine-generated cgo bridge between Go functions and native OpenGL functions. Glow is a fork of [GoGL2](https://github.com/chsc/gogl2).

Usage
-----

Until the API is stable Glow requires you to generate your own bindings.

    go get github.com/errcw/glow
    go build
    ... TBD
    
Once the bindings are installed you can use them with the appropriate import statements.

```Go
import (
  "github.com/errcw/glow/gl/4.4/gl"
  "github.com/errcw/glow/glt"
  _ "github.com/errcw/glow/procaddr/auto"
)
```

A few notes about the packages:
- *gl*: Contains the OpenGL functions and enumeration values for the imported version.
- *glt*: Contains helper functions for OpenGL type conversions. Of particular note is `glt.Ptr` which takes a Go array or slice or pointer and returns a corresponding `uintptr` to use with functions expecting data pointers.
- *procaddr*: Contains platform-specific functions for [loading OpenGL functions](https://www.opengl.org/wiki/Load_OpenGL_Functions). The `auto` subpackage will automatically select an appropriate implementation based on the build environment. Importing a `procaddr` subpackage is necessary for the `gl` package to work.

Once you have imported the necessary packages you must initialize the bindings.

```Go
func main() {
  if err := gl.Init(); err != nil {
    panic(err)
  }
}
```

A note about threading and goroutines. The bindings do not expose a mechanism to make an OpenGL context current on a different thread so you must restrict your usage to the thread on which you called `gl.Init()`. To do so you should use [LockOSThread](https://code.google.com/p/go-wiki/wiki/LockOSThread).

Examples
--------

A simple example illustrating how to use the bindings is available the [examples](https://github.com/errcw/glow/tree/master/examples) directory.
