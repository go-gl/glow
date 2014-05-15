package glt

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Enum uint32
type Bitfield uint32
type Sync unsafe.Pointer
type DebugProc unsafe.Pointer

// Ptr takes a pointer, slice, or array and returns its GL-compatible address.
func Ptr(data interface{}) uintptr {
	if data == nil {
		return uintptr(0)
	}
	v := reflect.ValueOf(data)
	switch v.Type().Kind() {
	case reflect.Ptr:
		e := v.Elem()
		switch e.Kind() {
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return e.UnsafeAddr()
		}
	case reflect.Uintptr:
		return v.Pointer()
	case reflect.Slice:
		return v.Index(0).UnsafeAddr()
	case reflect.Array:
		return v.UnsafeAddr()
	}
	panic(fmt.Sprintf("Unsupproted type %s; must be a pointer, slice, or array", v.Type()))
}
