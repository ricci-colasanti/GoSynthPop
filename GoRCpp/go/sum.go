package main

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

//export SumVec
func SumVec(vec *C.double, length C.int) C.double {
	if vec == nil || length <= 0 {
		return C.double(0)
	}
	slice := unsafe.Slice(vec, length)
	sum := C.double(0)
	for _, v := range slice {
		sum += v
	}
	return sum
}

func main() {}
