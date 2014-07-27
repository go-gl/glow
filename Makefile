# Makefile for generating the prebuilt (go-gettable) bindings.

bindings:
	./glow generate -api=gl -version=2.1
	./glow generate -api=gl -version=3.2 -profile=core
	./glow generate -api=gl -version=3.3 -profile=core
	./glow generate -api=gl -version=4.1 -profile=core
	./glow generate -api=gl -version=4.4 -profile=core
	./glow generate -api=gl -version=all -profile=core -lenientInit
	./glow generate -api=gl -version=3.2 -profile=compatibility
	./glow generate -api=gl -version=3.3 -profile=compatibility
	./glow generate -api=gl -version=4.1 -profile=compatibility
	./glow generate -api=gl -version=4.4 -profile=compatibility

install: bindings
	go install ./gl/2.1/gl
	go install ./gl-core/3.2/gl
	go install ./gl-core/3.3/gl
	go install ./gl-core/4.1/gl
	go install ./gl-core/4.4/gl
	go install ./gl-core/all/gl
	go install ./gl-compatibility/3.2/gl
	go install ./gl-compatibility/3.3/gl
	go install ./gl-compatibility/4.1/gl
	go install ./gl-compatibility/4.4/gl
