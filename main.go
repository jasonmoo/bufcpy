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
	"bufcpy"
	"utils"
	"sort"
	"strings"
)

type (
	Result struct {
		name string
		score   time.Duration
	}
	Results []*Result
)
func (r *Result) String() string { return fmt.Sprintf("%s: %s", r.name, r.score) }
func (r Results) Len() int           { return len(r) }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool { return r[i].score < r[j].score }
func (r *Results) String() string { return fmt.Sprintf(strings.Repeat("%s\n",len(r)), r...) }

var (
	BufMaxString, BufMinString, BufSizeString, Step string
	BufMax, BufMin, BufSize, Runs, MaxParts, MaxProcs int
	Copy, Compare, AllTests bool
	StepAction byte
	StepAmount int64
	MinStepAmount int64
	Debug bool
)

func init() {
	flag.IntVar(&Runs, "Runs", 100, "How many times to run each test")
	flag.StringVar(&BufMinString, "bufmin", "1kb", "Run tests on a range of buffer sizes.")
	flag.StringVar(&BufMaxString, "bufmax", "4mb", "Run tests on a range of buffer sizes.")
	flag.StringVar(&BufSizeString, "bufsize", "", "Run benchmarks on a single buffer size, instead of using bufmin/bufmax.")
	flag.StringVar(&Step, "step", "^", "The interval of the buffer range to test.  +Ns, *Ns, ^ (default: ^ or square)")
	flag.IntVar(&MaxParts, "maxparts", 8, "Maximum number of parts for tests that run concurrently on partitions (min: 2)")
	flag.IntVar(&MaxProcs, "maxprocs", 0, "Number of threads (GOMAXPROCS) default: runtime.NumCPU()")
	flag.BoolVar(&Copy, "copy", false, "Run the copy tests")
	flag.BoolVar(&Compare, "compare", false, "Run the compare tests")
	flag.BoolVar(&Debug, "debug", false, "Expose debugging information at http://localhost:8080/debug/pprof during benchmarking")
}

func print_usage() {
	fmt.Println("bufcpy usage:")
	fmt.Println("go build main.go && ./main")
	flag.PrintDefaults()
	fmt.Println()
	os.Exit(0)
}

// modifies i by ref
func nextBufSize(i *int) {
	switch StepAction {
	case '+': *i += int(StepAmount)
	case '*': *i *= int(StepAmount)
	case '^': *i *= *i
	}
}


func main() {
	var err error

	flag.Parse()
	if flag.NFlag() == 0 {
		print_usage()
	}

	fmt.Println("Running bufcpy auditor")

	if Debug {
		// fire up the debugging server
		go func() {	log.Fatal(http.ListenAndServe(":8080", nil)) }()
	}

	t1, _ := utils.ParseHumanReadableSize(BufMinString)
	t2, _ := utils.ParseHumanReadableSize(BufMaxString)
	t3, _ := utils.ParseHumanReadableSize(BufSizeString)
	BufMin, BufMax, BufSize = int(t1), int(t2), int(t3)

	if BufSize > 0 {
		BufMin, BufMax = BufSize, BufSize
	}
	if MaxProcs == 0 {
		MaxProcs = runtime.NumCPU()
	}

	StepAction = Step[0]
	MinStepAmount, _ = utils.ParseHumanReadableSize("11kb")
	if len(Step) > 1 {
		StepAmount, err = utils.ParseHumanReadableSize(Step[1:])
		if err != nil {
			panic("can't read your handwriting.  try +1k or *2mb")
		}
		if StepAmount < MinStepAmount {
			fmt.Println("Min step amount set at 1kb")
			StepAmount = MinStepAmount
		}
	}

	for bufsize := BufMin; bufsize <= BufMax; nextBufSize(&bufsize) {

		// native/memcpy aren't affected by cpus/parts
		// we can run them from here

		to, from := make([]byte, bufsize), make([]byte, bufsize)
		utils.FillBytes(from)

		for _, copyfunc := range []func([]byte,[]byte){bufcpy.NativeCopy, bufcpy.CgoMemcpy} {

			sum := time.Duration(0)
			for i := 0; i < Runs; i++ {
				utils.ZeroBytes(to)
				start := time.Now()
				copyfunc(to,from)
				end := time.Now()
				sum += end.Sub(start)
			}
			result := &Result{
				name: fmt.Sprintf(reflect.TypeOf(copyfunc).Name()+"(bufsize=%d)", utils.FormatHumanReadableSize(int64(bufsize), 1)),
				score: sum/time.Duration(Runs),
			}
			fmt.Println(result)
			Results = append(Results, result)

		}


		for cpus := 1; cpus <= MaxProcs; cpus<<=1 {
			runtime.GOMAXPROCS(cpus)
			runtime.GC()

			for parts := 1; parts <= MaxParts; parts<<=1 {


				for _, copyfunc := range []func([]byte,[]byte,int){
					bufcpy.RecursiveDacCopy,
					bufcpy.RecursiveDacCgoMemcpy,
					bufcpy.PartitionedCopy,
					bufcpy.PartitionedCgoMemcpy,
				} {

					sum := time.Duration(0)
					for i := 0; i < Runs; i++ {
						utils.ZeroBytes(to)
						start := time.Now()
						copyfunc(to,from, parts)
						end := time.Now()
						sum += end.Sub(start)
					}
					result = &Result{
						name: fmt.Sprintf(reflect.TypeOf(copyfunc).Name()+"(cpus=%d,bufsize=%d,parts=%d)", cpus, utils.FormatHumanReadableSize(int64(bufsize), 1), parts),
						score: sum/time.Duration(Runs),
					}
					fmt.Println(result)
					Results = append(Results, result)

				}


			}

		}

		fmt.Println("Top 10:")
		sort.Sort(Results)
		fm.Println(Results[:10])

	}

}