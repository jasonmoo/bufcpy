package bufcpy

import (
	"bytes"
	"log"
	"testing"
	"utils"
)

func must_parse(s string) int {
	size, err := utils.ParseHumanReadableSize(s)
	if err != nil {
		log.Fatal("Unable to parse %s: %s", s, err)
	}
	return int(size)
}

func T_Wrapper(f func(to, from []byte), bufsize string, t *testing.T) {
	// safe to run in parallel
	t.Parallel()
	bs := must_parse(bufsize)
	to, from := make([]byte, bs), bytes.Repeat([]byte{1}, bs)
	f(to, from)
	if ct := utils.DiffBytes(to, from); ct > 0 {
		t.Errorf("[]byte arrays are different.  Found %d non-matching bytes.", ct)
	}
}
func T_WrapperN(f func(to, from []byte, n int), bufsize string, n int, t *testing.T) {
	// safe to run in parallel
	t.Parallel()
	bs := must_parse(bufsize)
	to, from := make([]byte, bs), bytes.Repeat([]byte{1}, bs)
	f(to, from, n)
	if ct := utils.DiffBytes(to, from); ct > 0 {
		t.Errorf("[]byte arrays are different.  Found %d non-matching bytes.", ct)
	}
}
func B_Wrapper(f func(to, from []byte), bufsize string, b *testing.B) {
	bs := must_parse(bufsize)
	to, from := make([]byte, bs), bytes.Repeat([]byte{1}, bs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(to, from)
	}
}
func B_WrapperN(f func(to, from []byte, n int), bufsize string, n int, b *testing.B) {
	bs := must_parse(bufsize)
	to, from := make([]byte, bs), bytes.Repeat([]byte{1}, bs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(to, from, n)
	}
}

// tests
func TestCgoMemcpy(t *testing.T)             {
	// programatic tests

	T_Wrapper(CgoMemcpy, "1mb", t) }
func TestRecursiveDacCopy(t *testing.T)      { T_WrapperN(RecursiveDacCopy, "1mb", 1, t) }
func TestRecursiveDacCgoMemcpy(t *testing.T) { T_WrapperN(RecursiveDacCgoMemcpy, "1mb", 1, t) }
func TestPartitionedCopy(t *testing.T)       { T_WrapperN(PartitionedCopy, "1mb", 2, t) }
func TestPartitionedCgoMemcpy(t *testing.T)  { T_WrapperN(PartitionedCgoMemcpy, "1mb", 2, t) }

// benchmarks
func BenchmarkNativeCopy1mb(b *testing.B)                      { B_Wrapper(NativeCopy, "1mb", b) }
func BenchmarkCgoMemcpy1mb(b *testing.B)                       { B_Wrapper(CgoMemcpy, "1mb", b) }
func BenchmarkRecursiveDacCopyDepth1Size1mb(b *testing.B)      { B_WrapperN(RecursiveDacCopy, "1mb", 1, b) }
func BenchmarkRecursiveDacCopyDepth2Size1mb(b *testing.B)      { B_WrapperN(RecursiveDacCopy, "1mb", 2, b) }
func BenchmarkRecursiveDacCopyDepth3Size1mb(b *testing.B)      { B_WrapperN(RecursiveDacCopy, "1mb", 3, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth1Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "1mb", 1, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth2Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "1mb", 2, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth3Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "1mb", 3, b) }
func BenchmarkPartitionedCopyParts2Size1mb(b *testing.B)       { B_WrapperN(PartitionedCopy, "1mb", 2, b) }
func BenchmarkPartitionedCopyParts4Size1mb(b *testing.B)       { B_WrapperN(PartitionedCopy, "1mb", 4, b) }
func BenchmarkPartitionedCopyParts8Size1mb(b *testing.B)       { B_WrapperN(PartitionedCopy, "1mb", 8, b) }
func BenchmarkPartitionedCgoMemcpyParts2Size1mb(b *testing.B)  { B_WrapperN(PartitionedCgoMemcpy, "1mb", 2, b) }
func BenchmarkPartitionedCgoMemcpyParts4Size1mb(b *testing.B)  { B_WrapperN(PartitionedCgoMemcpy, "1mb", 4, b) }
func BenchmarkPartitionedCgoMemcpyParts8Size1mb(b *testing.B)  { B_WrapperN(PartitionedCgoMemcpy, "1mb", 8, b) }
