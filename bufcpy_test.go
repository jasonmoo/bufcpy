package bufcpy

import (
	"bytes"
	"log"
	"testing"
)

func must_parse(s string) int {
	size, err := goutils.ParseHumanReadableSize(s)
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
	if ct := goutils.DiffBytes(to, from); ct > 0 {
		t.Errorf("[]byte arrays are different.  Found %d non-matching bytes.", ct)
	}
}
func T_WrapperN(f func(to, from []byte, n int), bufsize string, n int, t *testing.T) {
	// safe to run in parallel
	t.Parallel()
	bs := must_parse(bufsize)
	to, from := make([]byte, bs), bytes.Repeat([]byte{1}, bs)
	f(to, from, n)
	if ct := goutils.DiffBytes(to, from); ct > 0 {
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
func TestCgoMemcpy(t *testing.T)                 { T_Wrapper(CgoMemcpy, "1mb", t) }
func TestRecursiveDacCopy(t *testing.T)          { T_WrapperN(RecursiveDacCopy, "1mb", 1, t) }
func TestRecursiveDacCgoMemcpy(t *testing.T)     { T_WrapperN(RecursiveDacCgoMemcpy, "1mb", 1, t) }
func TestPartitionedCopy(t *testing.T)           { T_WrapperN(PartitionedCopy, "1mb", 2, t) }
func TestPartitionedCgoMemcpy(t *testing.T)      { T_WrapperN(PartitionedCgoMemcpy, "1mb", 2, t) }
func Test2BytesParallelAssignment(t *testing.T)  { T_Wrapper(TwoBytesParallelAssignment, "1mb", t) }
func Test4BytesParallelAssignment(t *testing.T)  { T_Wrapper(FourBytesParallelAssignment, "1mb", t) }
func Test8BytesParallelAssignment(t *testing.T)  { T_Wrapper(EightBytesParallelAssignment, "1mb", t) }
func Test16BytesParallelAssignment(t *testing.T) { T_Wrapper(SixteenBytesParallelAssignment, "1mb", t) }

// benchmarks
func BenchmarkNativeCopy1mb(b *testing.B)  { B_Wrapper(NativeCopy, "1mb", b) }
func BenchmarkNativeCopy2mb(b *testing.B)  { B_Wrapper(NativeCopy, "2mb", b) }
func BenchmarkNativeCopy4mb(b *testing.B)  { B_Wrapper(NativeCopy, "4mb", b) }
func BenchmarkNativeCopy8mb(b *testing.B)  { B_Wrapper(NativeCopy, "8mb", b) }

func BenchmarkCgoMemcpy1mb(b *testing.B)  { B_Wrapper(CgoMemcpy, "1mb", b) }
func BenchmarkCgoMemcpy2mb(b *testing.B)  { B_Wrapper(CgoMemcpy, "2mb", b) }
func BenchmarkCgoMemcpy4mb(b *testing.B)  { B_Wrapper(CgoMemcpy, "4mb", b) }
func BenchmarkCgoMemcpy8mb(b *testing.B)  { B_Wrapper(CgoMemcpy, "8mb", b) }

func BenchmarkRecursiveDacCopyDepth1Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "1mb", 1, b) }
func BenchmarkRecursiveDacCopyDepth1Size2mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "2mb", 1, b) }
func BenchmarkRecursiveDacCopyDepth1Size4mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "4mb", 1, b) }
func BenchmarkRecursiveDacCopyDepth1Size8mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "8mb", 1, b) }

func BenchmarkRecursiveDacCopyDepth2Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "1mb", 2, b) }
func BenchmarkRecursiveDacCopyDepth2Size2mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "2mb", 2, b) }
func BenchmarkRecursiveDacCopyDepth2Size4mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "4mb", 2, b) }
func BenchmarkRecursiveDacCopyDepth2Size8mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "8mb", 2, b) }

func BenchmarkRecursiveDacCopyDepth3Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "1mb", 3, b) }
func BenchmarkRecursiveDacCopyDepth3Size2mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "2mb", 3, b) }
func BenchmarkRecursiveDacCopyDepth3Size4mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "4mb", 3, b) }
func BenchmarkRecursiveDacCopyDepth3Size8mb(b *testing.B) { B_WrapperN(RecursiveDacCopy, "8mb", 3, b) }

func BenchmarkRecursiveDacCgoMemcpyDepth1Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "1mb", 1, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth1Size2mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "2mb", 1, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth1Size4mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "4mb", 1, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth1Size8mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "8mb", 1, b) }

func BenchmarkRecursiveDacCgoMemcpyDepth2Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "1mb", 2, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth2Size2mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "2mb", 2, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth2Size4mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "4mb", 2, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth2Size8mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "8mb", 2, b) }

func BenchmarkRecursiveDacCgoMemcpyDepth3Size1mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "1mb", 3, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth3Size2mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "2mb", 3, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth3Size4mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "4mb", 3, b) }
func BenchmarkRecursiveDacCgoMemcpyDepth3Size8mb(b *testing.B) { B_WrapperN(RecursiveDacCgoMemcpy, "8mb", 3, b) }

func BenchmarkPartitionedCopyParts2Size1mb(b *testing.B) { B_WrapperN(PartitionedCopy, "1mb", 2, b) }
func BenchmarkPartitionedCopyParts2Size2mb(b *testing.B) { B_WrapperN(PartitionedCopy, "2mb", 2, b) }
func BenchmarkPartitionedCopyParts2Size4mb(b *testing.B) { B_WrapperN(PartitionedCopy, "4mb", 2, b) }
func BenchmarkPartitionedCopyParts2Size8mb(b *testing.B) { B_WrapperN(PartitionedCopy, "8mb", 2, b) }

func BenchmarkPartitionedCopyParts4Size1mb(b *testing.B) { B_WrapperN(PartitionedCopy, "1mb", 4, b) }
func BenchmarkPartitionedCopyParts4Size2mb(b *testing.B) { B_WrapperN(PartitionedCopy, "2mb", 4, b) }
func BenchmarkPartitionedCopyParts4Size4mb(b *testing.B) { B_WrapperN(PartitionedCopy, "4mb", 4, b) }
func BenchmarkPartitionedCopyParts4Size8mb(b *testing.B) { B_WrapperN(PartitionedCopy, "8mb", 4, b) }

func BenchmarkPartitionedCopyParts8Size1mb(b *testing.B) { B_WrapperN(PartitionedCopy, "1mb", 8, b) }
func BenchmarkPartitionedCopyParts8Size2mb(b *testing.B) { B_WrapperN(PartitionedCopy, "2mb", 8, b) }
func BenchmarkPartitionedCopyParts8Size4mb(b *testing.B) { B_WrapperN(PartitionedCopy, "4mb", 8, b) }
func BenchmarkPartitionedCopyParts8Size8mb(b *testing.B) { B_WrapperN(PartitionedCopy, "8mb", 2, b) }

func BenchmarkPartitionedCgoMemcpyParts2Size1mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "1mb", 2, b) }
func BenchmarkPartitionedCgoMemcpyParts2Size2mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "2mb", 2, b) }
func BenchmarkPartitionedCgoMemcpyParts2Size4mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "4mb", 2, b) }
func BenchmarkPartitionedCgoMemcpyParts2Size8mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "8mb", 2, b) }

func BenchmarkPartitionedCgoMemcpyParts4Size1mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "1mb", 4, b) }
func BenchmarkPartitionedCgoMemcpyParts4Size2mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "2mb", 4, b) }
func BenchmarkPartitionedCgoMemcpyParts4Size4mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "4mb", 4, b) }
func BenchmarkPartitionedCgoMemcpyParts4Size8mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "8mb", 4, b) }

func BenchmarkPartitionedCgoMemcpyParts8Size1mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "1mb", 8, b) }
func BenchmarkPartitionedCgoMemcpyParts8Size2mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "2mb", 8, b) }
func BenchmarkPartitionedCgoMemcpyParts8Size4mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "4mb", 8, b) }
func BenchmarkPartitionedCgoMemcpyParts8Size8mb(b *testing.B) { B_WrapperN(PartitionedCgoMemcpy, "8mb", 2, b) }
