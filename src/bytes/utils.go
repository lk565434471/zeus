package bytes

import (
	"reflect"
	"unsafe"
)

func StringToBytes(str string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&str))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len

	return b
}
