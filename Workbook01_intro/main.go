package main

import (
	"fmt"
	"time"
)

func count(id int) {
	for i := 0; i < 3; i++ {
		fmt.Printf("Process %d loop % d \n", id, i)
		time.Sleep(1 * time.Second)
	}
}
func main() {
	for i := 0; i < 4; i++ {
		go count(i)
	}
	time.Sleep(4 * time.Second)
}
