package main

import (
	"fmt"
	"runtime"
	"reflect"
	"time"
	"flag"
	"os"
	"log"
	"net/http"
)

type Increment struct {
	chunk int
	process func(int,int)
}

type (
	Result struct {
		name string
		score   time.Duration
	}
	Results []Result
)
func (r Results) Len() int           { return len(r) }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool { return r[i].score < r[j].score }

const (
	DefaultRuns      = 100
	DefaultMinBuf    = "1kb"
	DefaultMaxBuf    = "4mb"
	DefaultStep      = "^"
	DefaultMaxParts  = 8
)

var (
	BufMaxString, BufMinString, BufSizeString, Step string
	BufMax, BufMin, BufSize, Runs, MaxParts, MaxProcs int
	Copy, Compare, AllTests bool
	Debug bool

	Benchmarks Results
)

func init() {
	flag.IntVar(&Runs, "runs", DefaultRuns, "How many times to run each test")
	flag.StringVar(&BufMinString, "bufmin", DefaultMinBuf, "Run tests on a range of buffer sizes.")
	flag.StringVar(&BufMaxString, "bufmax", DefaultMaxBuf, "Run tests on a range of buffer sizes.")
	flag.StringVar(&BufSizeString, "bufsize", "", "Run benchmarks on a single buffer size, instead of using bufmin/bufmax.")
	flag.StringVar(&Step, "step", DefaultStep, "The interval of the buffer range to test.  +Ns, *Ns, ^ (default: ^ or square)")
	flag.IntVar(&MaxParts, "maxparts", DefaultMaxParts, "Maximum number of parts for tests that run concurrently on partitions (min: 2)")
	flag.IntVar(&MaxProcs, "maxprocs", 0, "Number of threads in scheduled thread-pool (GOMAXPROCS) default: runtime.NumCPU()")
	flag.BoolVar(&Copy, "copy", false, "Run the copy tests")
	flag.BoolVar(&Compare, "compare", false, "Run the compare tests")
	flag.BoolVar(&AllTests, "all", false, "Run all the benchmarks, including slower ones")
	flag.BoolVar(&Debug, "debug", false, "Expose debugging information at http://localhost:8080/debug/pprof during benchmarking")
}

func print_usage() {
	fmt.Println("bufcpy usage:")
	fmt.Println("go build main.go && ./main")
	flag.PrintDefaults()
	fmt.Println()
	os.Exit(0)
}

var next map[uint8]func(int,int) = map[uint8]func(int,int){
	uint8('+'):func(a,b int) { return a+b },
	uint8('*'):func(a,b int) { return a*b },
	uint8('^'):func(a,b int) { return a*a },
}

func nextBufSize(i *int) {
	chunk, err := ParseHumanReadableSize(BufMinString)
	if err != nil {
		panic("can't read your handwriting.  try 1k or 2mb")
	}
	if f, ok := next[Step[0]]; ok {
		*i = f(*i, int(chunk))
	}
}


func main() {

	flag.Parse()
	if flag.NFlag() == 0 {
		print_usage()
	}

	if Debug {
		// fire up the debugging server
		go func() {	log.Fatal(http.ListenAndServe(":8080", nil)) }()
	}

	t1, _ := ParseHumanReadableSize(BufMinString)
	t2, _ := ParseHumanReadableSize(BufMaxString)
	t3, _ := ParseHumanReadableSize(BufSizeString)
	BufMin, BufMax, BufSize = int(t1), int(t2), int(t3)

	if BufSize > 0 {
		BufMin, BufMax = BufSize, BufSize
	}
	if MaxProcs == 0 {
		MaxProcs = runtime.NumCPU()
	}



	for bufsize := BufMin; bufsize <= BufMax; nextBufSize(&i) {

		// native/memcpy aren't affected by cpus/parts
		// we can run them from here

		var sum time.Duration

		for copyfunc := range []func([]byte,[]byte){NativeCopy,CgoMemcpy} {

			sum = time.Duration(0)
			for i := 0; i < runs; i++ {
				start := time.Now()
				copyfunc(to,from)
				end := time.Now()
				sum += end-start
				ZeroBytes(to)
			}
			Benchmarks = append(Benchmarks, Result{
				name: fmt.Sprintf(reflect.TypeOf(copyfunc).Name()+"(cpus=%d,bufsize=%d)", cpus, FormatHumanReadableSize(bufsize)),
				score: sum/time.Duration(runs),
			})

		}


		for cpus := 1; cpus <= MaxProcs; cpus<<=1 {
			runtime.GOMAXPROCS(cpus)
			runtime.GC()
			to, from := make([]byte, bufsize), make([]byte, bufsize)
			FillBytes(from)

			for parts := 1; parts <= MaxParts; parts<<=1 {


				for copyfunc := range []func([]byte,[]byte){RecursiveDacCopy,RecursiveDacCgoMemcpy,PartitionedCopy,PartitionedCgoMemcpy} {

					sum = time.Duration(0)
					for i := 0; i < runs; i++ {
						start := time.Now()
						copyfunc(to,from)
						end := time.Now()
						sum += end-start
						ZeroBytes(to)
					}
					Benchmarks = append(Benchmarks, Result{
						name: fmt.Sprintf(reflect.TypeOf(copyfunc).Name()+"(cpus=%d,bufsize=%d,parts=%d)", cpus, FormatHumanReadableSize(bufsize), parts),
						score: sum/time.Duration(runs),
					})

				}


			}

		}


	}



}