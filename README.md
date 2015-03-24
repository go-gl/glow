Glow
====

Glow is an OpenGL binding generator for Go. Glow parses the [OpenGL XML API registry](https://cvs.khronos.org/svn/repos/ogl/trunk/doc/registry/public/api/) to produce a machine-generated cgo bridge between Go functions and native OpenGL functions. Glow is a fork of [GoGL2](https://github.com/chsc/gogl2).

Features:
- Go functions that mirror the C specification using Go types.
- Support for multiple OpenGL APIs (GL/GLES/EGL/WGL/GLX/EGL), versions, and profiles.
- Support for extensions (including debug callbacks).

See the [open issues](https://github.com/go-gl/glow/issues) for caveats about the current state of the implementation.

Generated Packages
------------------

Generated OpenGL binding packages are available in the [go-gl/gl](https://github.com/go-gl/gl) repository.

Custom Packages
---------------

If the prebuilt, go-gettable packages are not suitable for your needs you can build your own. For example,

    go get github.com/go-gl/glow
    cd $GOPATH/src/github.com/go-gl/glow
    go build
    ./glow download
    ./glow generate -api=gl -version=3.3 -profile=core -remext=GL_ARB_cl_event
    go install ./gl-core/3.3/gl

A few notes about the flags to `generate`:
- `api`: One of `gl`, `egl`, `wgl`, or `glx`.
- `version`: The API version to generate. The `all` pseudo-version includes all functions and enumerations for the specified API.
- `profile`: For `gl` packages with version 3.2 or higher, `core` or `compatibility` ([explanation](http://www.opengl.org/wiki/Core_And_Compatibility_in_Contexts)).
- `addext`: A regular expression describing which extensions to include. `.*` by default, including everything.
- `restrict`: A JSON file that explicitly lists what enumerations / functions that Glow should generate (see example.json).
- `remext`: A regular expression describing which extensions to exclude. Empty by default, excluding nothing. Takes precedence over explicitly added regular expressions.
- `lenientInit`: Flag to disable strict function availability checks at `Init` time. By default if any non-extension function pointer cannot be loaded then initialization fails; when this flag is set initialization will succeed with missing functions. Note that on some platforms unavailable functions will load successfully even but fail upon invocation so check against the OpenGL context what is supported.
