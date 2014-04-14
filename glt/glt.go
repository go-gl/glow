package glt

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Enum uint32
type Bitfield uint32
type Pointer uintptr
type Sync unsafe.Pointer
type DebugProc unsafe.Pointer

func Ptr(data interface{}) Pointer {
	if data == nil {
		return Pointer(0)
	}
	v := reflect.ValueOf(data)
	switch v.Type().Kind() {
	case reflect.Ptr: // for pointers: *byte, *int, ...
		e := v.Elem()
		switch e.Kind() {
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return Pointer(e.UnsafeAddr())
		}
	case reflect.Uintptr:
		return Pointer(v.Pointer())
	case reflect.Slice: // for slices and arrays: []int, []float32, ...
		return Pointer(v.Index(0).UnsafeAddr())
	case reflect.Array:
		return Pointer(v.UnsafeAddr())
	}
	panic(fmt.Sprintf("unknown type: %s: must be a pointer or a slice.", v.Type()))
}

func (p Pointer) Offset(o uintptr) Pointer {
	return Pointer(uintptr(p) + uintptr(o))
}
