package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

//export Square
func Square(x C.int) C.int {
	return x * x
}

//export Sum
func Sum(arr *C.int, length C.int) C.int {
	if length <= 0 {
		return 0
	}

	// Create a proper Go slice header
	slice := (*[1<<30 - 1]C.int)(unsafe.Pointer(arr))[:length:length]
	total := C.int(0)
	for i := C.int(0); i < length; i++ {
		total += slice[i]
	}
	return total
}

func main() {}
