package bufcpy

// #include <string.h>
import "C"

import (
	"unsafe"
	"bytes"
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

/*


package bufcpy

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"strconv"
	"time"
	"unsafe"
)
// #include <string.h>
import "C"

const (
	DefaultRuns     int = 100
	DefaultMinBuf   string = "1kb"
	DefaultMaxBuf   string = "4mb"
	DefaultMaxRoutines int = 8
)

type Result struct {
	score   time.Duration
	message string
}
type Results []Result

func (r Results) Len() int           { return len(r) }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool { return r[i].score < r[j].score }




func PartitionedEqual(ct int, to, from []byte) bool {
	done, chunk := make(chan bool, ct), len(to)/ct

	// first and last goroutines are run special because of slice syntax
	c := (ct - 2) * chunk
	go func(c int) { done <- bytes.Equal(to[:c], from[:c]) }(chunk)
	go func(c int) { done <- bytes.Equal(to[c:], from[c:]) }(c)

	for ; c >= chunk; c -= chunk {
		go func(c, cc int) { done <- bytes.Equal(to[c:cc], from[c:cc]) }(c, c+chunk)
	}

	for i := 0; i < ct; i++ {
		if n := <-done; !n {
			return false
		}
	}

	return true
}
func PartitionedCgoMemcmp(ct int, to, from []byte) bool {
	done, chunk := make(chan int, ct), len(to)/ct

	// first and last goroutines are run special because of slice syntax
	c := (ct - 2) * chunk
	go func(c int) { done <- int(C.memcmp(unsafe.Pointer(&(to[0])), unsafe.Pointer(&(from[0])), C.size_t(c))) }(chunk)
	go func(c int) { done <- int(C.memcmp(unsafe.Pointer(&(to[c])), unsafe.Pointer(&(from[c])), C.size_t(len(from)-c))) }(c)

	C_chunk := C.size_t(chunk)
	for ; c >= chunk; c -= chunk {
		go func(c int) { done <- int(C.memcmp(unsafe.Pointer(&(to[c])), unsafe.Pointer(&(from[c])), C_chunk)) }(c)
	}

	for i := 0; i < ct; i++ {
		if n := <-done; n != 0 {
			return false
		}
	}
	return true
}
func ensure_power_of_2(i int) {
	if !is_power_of_2(i) {
		log.Fatalf("%d is not a power of 2", i)
	}
}
func ensure_equals(to, from []byte) {
	if !bytes.Equal(to, from) {
		log.Fatal("buffers do not match!")
	}
}

func RunGoroutineCopy(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {
		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			zero(to)
			start := time.Now()

			done, chunk := make(chan int), len(to)/v

			// first and last goroutines are run special because of slice syntax
			c := (v - 2) * chunk
			go func(c int) { copy(to[:c], from[:c]); done <- 1 }(chunk)
			go func(c int) { copy(to[c:], from[c:]); done <- 1 }(c)

			for ; c >= chunk; c -= chunk {
				go func(c, cc int) { copy(to[c:cc], from[c:cc]); done <- 1 }(c, c+chunk)
			}

			for i := 0; i < v; i++ {
				<-done
			}

			t += time.Since(start)
		}
		ensure_equals(to, from)
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%22d goroutines copy(): %s", v, t/time.Duration(Runs))})
	}

	return resp
}
func RunGoroutineCgoMemcpy(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {
		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			zero(to)
			start := time.Now()

			done, chunk := make(chan int), len(to)/v

			// first and last goroutines are run special because of slice syntax
			c := (v - 2) * chunk

			go func(c int) { C.memcpy(unsafe.Pointer(&(to[0])), unsafe.Pointer(&(from[0])), C.size_t(c)); done <- 1	}(chunk)
			go func(c int) { C.memcpy(unsafe.Pointer(&(to[c])), unsafe.Pointer(&(from[c])), C.size_t(len(from)-c)); done <- 1 }(c)

			C_chunk := C.size_t(chunk)
			for ; c >= chunk; c -= chunk {
				go func(c int) { C.memcpy(unsafe.Pointer(&(to[c])), unsafe.Pointer(&(from[c])), C_chunk); done <- 1	}(c)
			}

			for i := 0; i < v; i++ {
				<-done
			}

			t += time.Since(start)
		}
		ensure_equals(to, from)
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%18d goroutines C.memcpy(): %s", v, t/time.Duration(Runs))})
	}

	return resp
}


// @todo fix
func RunGoroutineRangeCopy(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {

		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			zero(to)
			start := time.Now()

			done, chunk := make(chan int), len(to)/v

			// first and last goroutines are run special because of slice syntax
			c := (v - 2) * chunk
			go func() {
				for i, b := range from[:chunk] {
					to[i] = b
				}
				done <- 1
			}()
			go func(c int) {
				for i, b := range from[c:] {
					to[i] = b
				}
				done <- 1
			}(c)
			for c > chunk {
				c -= chunk
				go func(c, cc int) {
					for i, b := range from[c:cc] {
						to[i] = b
					}
					done <- 1
				}(c-chunk, c)
			}

			for i := 0; i < v; i++ {
				<-done
			}

			t += time.Since(start)
		}
		ensure_equals(to, from)
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%18d goroutines range copy: %s", v, t/time.Duration(Runs))})
	}

	return resp
}

func RunBytesEqualSingle(to, from []byte) Result {

	t := time.Duration(0)
	for i := 0; i < Runs; i++ {
		runtime.GC()
		start := time.Now()

		if !bytes.Equal(to, from) {
			fmt.Println("buffers do not match!")
			os.Exit(1)
		}

		t += time.Since(start)
	}
	return Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%40s: %s", "bytes.Equal()", t/time.Duration(Runs))}
}
func RunGoroutineBytesEqual(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {
		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			start := time.Now()

			done, chunk := make(chan bool), len(to)/v

			// first and last goroutines are run special because of slice syntax
			c := (v - 2) * chunk
			go func(c int) { done <- bytes.Equal(to[:c], from[:c]) }(chunk)
			go func(c int) { done <- bytes.Equal(to[c:], from[c:]) }(c)

			for ; c >= chunk; c -= chunk {
				go func(c, cc int) { done <- bytes.Equal(to[c:cc], from[c:cc]) }(c, c+chunk)
			}

			for i := 0; i < v; i++ {
				if n := <-done; !n {
					fmt.Printf("bytes did not match!\n")
					os.Exit(1)
				}
			}

			t += time.Since(start)
		}
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%15d goroutines bytes.Equal(): %s", v, t/time.Duration(Runs))})
	}

	return resp
}
func RunConcurrentBytesEqual(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {
		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			start := time.Now()

			if !concurrent_equal(v, to, from) {
				fmt.Printf("bytes did not match!\n")
				os.Exit(1)
			}

			t += time.Since(start)
		}
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%21d concurrent_equal(): %s", v, t/time.Duration(Runs))})
	}

	return resp
}
func RunCgoMemcmpSingle(to, from []byte) Result {

	t := time.Duration(0)
	for i := 0; i < Runs; i++ {
		runtime.GC()
		start := time.Now()

		if C.memcmp(unsafe.Pointer(&(to[0])), unsafe.Pointer(&(from[0])), C.size_t(len(from))) != 0 {
			fmt.Println("buffers do not match!")
			os.Exit(1)
		}

		t += time.Since(start)
	}
	return Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%40s: %s", "C.memcmp()", t/time.Duration(Runs))}
}
func RunGoroutineCgoMemcmp(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {
		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			zero(to)
			start := time.Now()

			done, chunk := make(chan int), len(to)/v

			// first and last goroutines are run special because of slice syntax
			c := (v - 2) * chunk

			go func(c int) { done <- int(C.memcmp(unsafe.Pointer(&(to[0])), unsafe.Pointer(&(from[0])), C.size_t(c))) }(chunk)
			go func(c int) { done <- int(C.memcmp(unsafe.Pointer(&(to[c])), unsafe.Pointer(&(from[c])), C.size_t(len(from)-c))) }(c)

			C_chunk := C.size_t(chunk)
			for ; c >= chunk; c -= chunk {
				go func(c int) { done <- int(C.memcmp(unsafe.Pointer(&(to[c])), unsafe.Pointer(&(from[c])), C_chunk)) }(c)
			}

			for i := 0; i < v; i++ {
				if n := <-done; n != 0 {
					fmt.Printf("bytes did not match!\n")
					os.Exit(1)
				}
			}

			t += time.Since(start)
		}
		ensure_equals(to, from)
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%18d goroutines C.memcmp(): %s", v, t/time.Duration(Runs))})
	}

	return resp
}
func RunConcurrentCgoMemcmp(to, from []byte, parts int) []Result {

	resp := make(Results, 0, parts+1)

	for v := 2; v <= parts; v <<= 1 {
		t := time.Duration(0)
		for i := 0; i < Runs; i++ {
			runtime.GC()
			zero(to)
			start := time.Now()

			if !concurrent_cgo_memcmp(v, to, from) {
				fmt.Printf("bytes did not match!\n")
				os.Exit(1)
			}

			t += time.Since(start)
		}
		ensure_equals(to, from)
		resp = append(resp, Result{score: t / time.Duration(Runs), message: fmt.Sprintf("%5d goroutines concurrent_cgo_memcmp(): %s", v, t/time.Duration(Runs))})
	}

	return resp
}




var (
	BufMaxString, BufMinString, BufSizeString string
	BufMax, BufMin, BufSize, Runs, MaxRoutines, MaxProcs int
	Copy, Compare, AllTests bool
	Usage bool
)

func init() {
	flag.IntVar(&Runs, "runs", DefaultRuns, "How many times to run each test")
	flag.StringVar(&BufMinString, "bufmin", DefaultMinBuf, "Run tests on a range of buffer sizes")
	flag.StringVar(&BufMaxString, "bufmax", DefaultMaxBuf, "Run tests on a range of buffer sizes")
	flag.StringVar(&BufSizeString, "bufsize", "0b", "Run tests on a single buffer size")
	flag.IntVar(&MaxRoutines, "maxroutines", DefaultMaxRoutines, "Maximum number of parts for tests that run concurrently on partitions (min: 2)")
	flag.IntVar(&MaxProcs, "maxprocs", 0, "Number of threads in scheduled thread-pool (GOMAXPROCS) default: runtime.NumCPU()")
	flag.BoolVar(&Copy, "copy", false, "Run the copy tests")
	flag.BoolVar(&Compare, "compare", false, "Run the compare tests")
	flag.BoolVar(&AllTests, "all", false, "Run all the tests including slower ones")
}

func print_usage() {
	fmt.Println("Boof usage:")
	fmt.Println("go run bc.go")
	flag.PrintDefaults()
	fmt.Println()
	os.Exit(0)
}

func version_info() {

}

func main() {

	flag.Parse()
	if flag.NFlag() == 0 {
		print_usage()
	}

	t1, _ := human_size_to_bytes(BufMinString)
	t2, _ := human_size_to_bytes(BufMaxString)
	t3, _ := human_size_to_bytes(BufSizeString)
	BufMin, BufMax, BufSize = int(t1), int(t2), int(t3)

	if BufSize > 0 {
		BufMin, BufMax = BufSize, BufSize
	}
	if MaxProcs == 0 {
		MaxProcs = runtime.NumCPU()
	}

	runtime.GOMAXPROCS(MaxProcs)

	// 1kb -> MaxMB
	for s := BufMin; s <= BufMax; s <<= 1 {

		to, from := make([]byte, s), make([]byte, s)

		if Copy {
			fmt.Printf("\nTesting %s buffer copies\n", human_size(s))
			fill(from)

			runs, baseline := Results{}, RunNativeCopySingle(to, from)
			runs = append(runs, baseline)
			runs = concat(runs, RunRecursiveDacCopy(to, from, MaxRoutines))
			runs = concat(runs, RunGoroutineCopy(to, from, MaxRoutines))
			runs = concat(runs, RunConcurrentCopy(to, from, MaxRoutines))
			runs = append(runs, RunCgoMemcpySingle(to, from))
			runs = concat(runs, RunGoroutineCgoMemcpy(to, from, MaxRoutines))
			runs = concat(runs, RunConcurrentCgoMemcpy(to, from, MaxRoutines))
			runs = concat(runs, RunRecursiveDacCgoMemcpy(to, from, MaxRoutines))

			if AllTests {
				runs = append(runs, RunOneByteCopySingle(to, from))
				runs = append(runs, RunTwoBytesParallelCopySingle(to, from))
				runs = append(runs, RunFourBytesParallelCopySingle(to, from))
				runs = append(runs, RunEightBytesParallelCopySingle(to, from))
				runs = append(runs, RunSixteenBytesParallelCopySingle(to, from))
			}

			sort.Sort(runs)
			if runs[0].score < baseline.score {
				timediff, deltax := baseline.score - runs[0].score, float64(baseline.score)/float64(runs[0].score)-1
				fmt.Printf("Winner is %s\nDelta: %s/%.2f%%/%.2fx faster than native copy()\n\n", strings.TrimSpace(runs[0].message), timediff, (float64(timediff)/float64(runs[0].score))*float64(100), deltax )
			} else {
				fmt.Printf("Winner is %s\n\n", strings.TrimSpace(runs[0].message))
			}

			for _, msg := range runs {
				fmt.Println(msg.message)
			}
		}

		if Compare {
			fmt.Printf("\nTesting %s buffer compare\n", human_size(s))
			runs, baseline := Results{}, RunBytesEqualSingle(to, from)
			runs = append(runs, baseline)
			runs = concat(runs, RunGoroutineBytesEqual(to, from, MaxRoutines))
			runs = concat(runs, RunConcurrentBytesEqual(to, from, MaxRoutines))
			runs = append(runs, RunCgoMemcmpSingle(to, from))
			runs = concat(runs, RunGoroutineCgoMemcmp(to, from, MaxRoutines))
			runs = concat(runs, RunConcurrentCgoMemcmp(to, from, MaxRoutines))

			sort.Sort(runs)
			if runs[0].score < baseline.score {
				timediff, deltax := baseline.score - runs[0].score, float64(baseline.score)/float64(runs[0].score)
				fmt.Printf("Winner is %s\nDelta: %s/%.2f%%/%.2fx faster than native bytes.Equal()\n\n", strings.TrimSpace(runs[0].message), timediff, float64(deltax)*float64(100), deltax )
			} else {
				fmt.Printf("Winner is %s\n\n", strings.TrimSpace(runs[0].message))
			}

			for _, msg := range runs {
				fmt.Println(msg.message)
			}
		}

		fmt.Println("\n")

	}

}

*/
