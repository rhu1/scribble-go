// +build debug

package main

import "fmt"

func debugf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
