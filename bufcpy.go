package bufcpy

// #include <string.h>
import "C"

import (
	"bytes"
	"unsafe"
)

// Copy functions
//
func NativeCopy(to, from []byte) {
	copy(to, from)
}
func CgoMemcpy(to, from []byte) {
	// seriously this can't be a good idea
	C.memcpy(unsafe.Pointer(&(to[0])), unsafe.Pointer(&(from[0])), C.size_t(len(to)))
}

// recursive divide and conquer copy
// 2**(depth+1) == number of partitions
func RecursiveDacCopy(to, from []byte, depth int) {
	done, h, depth := make(chan int, 2), len(to)/2, depth-1
	if depth > 0 {
		go func() { RecursiveDacCopy(to[:h], from[:h], depth); done <- 1 }()
		go func() { RecursiveDacCopy(to[h:], from[h:], depth); done <- 1 }()
	} else {
		go func() { copy(to[:h], from[:h]); done <- 1 }()
		go func() { copy(to[h:], from[h:]); done <- 1 }()
	}
	<-done
	<-done
}

func RecursiveDacCgoMemcpy(to, from []byte, depth int) {
	done, h, depth := make(chan int, 2), len(to)/2, depth-1
	if depth > 0 {
		go func() { RecursiveDacCgoMemcpy(to[:h], from[:h], depth); done <- 1 }()
		go func() { RecursiveDacCgoMemcpy(to[h:], from[h:], depth); done <- 1 }()
	} else {
		go func() { CgoMemcpy(to[:h], from[:h]); done <- 1 }()
		go func() { CgoMemcpy(to[h:], from[h:]); done <- 1 }()
	}
	<-done
	<-done
}

func PartitionedCopy(to, from []byte, parts int) {
	done, lento, chunk := make(chan int, parts), len(to), len(to)/parts

	// using from-beginning slicing to ensure the center portion
	// of the slice range aligns with our chunk size
	offset := len(to) % chunk
	if offset > 0 {
		go func() { copy(to[:offset], from[:offset]); done <- 1 }()
	}

	for i, end := offset, lento; i < end; i += chunk {
		go func(c int) { copy(to[c:c+chunk], from[c:c+chunk]); done <- 1 }(i)
	}

	for i := 0; i < parts; i++ {
		<-done
	}
}
func PartitionedCgoMemcpy(to, from []byte, parts int) {
	done, lento, chunk := make(chan int, parts), len(to), len(to)/parts

	// using from-beginning slicing to ensure the center portion
	// of the slice range aligns with our chunk size
	offset := len(to) % chunk
	if offset > 0 {
		go func() { CgoMemcpy(to[:offset], from[:offset]); done <- 1 }()
	}

	for i, end := offset, lento; i < end; i += chunk {
		go func(c int) { CgoMemcpy(to[c:c+chunk], from[c:c+chunk]); done <- 1 }(i)
	}

	for i := 0; i < parts; i++ {
		<-done
	}
}

// Compare functions
//
func CgoMemcmp(to, from []byte) bool {
	// seriously this can't be a good idea
	return C.memcmp(unsafe.Pointer(&(to[0])), unsafe.Pointer(&(from[0])), C.size_t(len(to))) == 0
}

func PartitionedEqual(to, from []byte, parts int) bool {
	if len(to) != len(from) {
		return false
	}

	match, lento, chunk := make(chan bool, parts), len(to), len(to)/parts

	// using from-beginning slicing to ensure the center portion
	// of the slice range aligns with our chunk size
	offset := len(to) % chunk
	if offset > 0 {
		go func() { match <- bytes.Equal(to[:offset], from[:offset]) }()
	}

	for i, end := offset, lento; i < end; i += chunk {
		go func(c int) { match <- bytes.Equal(to[c:c+chunk], from[c:c+chunk]) }(i)
	}

	for i := 0; i < parts; i++ {
		if !<-match {
			return false
		}
	}

	return true
}
func PartitionedCgoMemcmp(to, from []byte, parts int) bool {
	if len(to) != len(from) {
		return false
	}

	match, lento, chunk := make(chan bool, parts), len(to), len(to)/parts

	// using from-beginning slicing to ensure the center portion
	// of the slice range aligns with our chunk size
	offset := len(to) % chunk
	if offset > 0 {
		go func() { match <- CgoMemcmp(to[:offset], from[:offset]) }()
	}

	for i, end := offset, lento; i < end; i += chunk {
		go func(c int) { match <- CgoMemcmp(to[c:c+chunk], from[c:c+chunk]) }(i)
	}

	for i := 0; i < parts; i++ {
		if !<-match {
			return false
		}
	}

	return true
}
func RecursiveDacEqual(to, from []byte, depth int) bool {
	match, h, depth := make(chan bool, 2), len(to)/2, depth-1
	if depth > 0 {
		go func() { match <- RecursiveDacEqual(to[:h], from[:h], depth) }()
		go func() { match <- RecursiveDacEqual(to[h:], from[h:], depth) }()
	} else {
		go func() { match <- bytes.Equal(to[:h], from[:h]) }()
		go func() { match <- bytes.Equal(to[h:], from[h:]) }()
	}
	if !<-match || !<-match {
		return false
	}
	return true
}
func RecursiveDacCgoMemcmp(to, from []byte, depth int) bool {
	match, h, depth := make(chan bool, 2), len(to)/2, depth-1
	if depth > 0 {
		go func() { match <- RecursiveDacCgoMemcmp(to[:h], from[:h], depth) }()
		go func() { match <- RecursiveDacCgoMemcmp(to[h:], from[h:], depth) }()
	} else {
		go func() { match <- CgoMemcmp(to[:h], from[:h]) }()
		go func() { match <- CgoMemcmp(to[h:], from[h:]) }()
	}
	if !<-match || !<-match {
		return false
	}
	return true
}
