package main

import (
	"strconv"
	"bufcpy"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	"utils"
)

type (
	Result struct {
		name  string
		score time.Duration
	}
	Results []*Result
)

func (r *Result) String() string     { return fmt.Sprintf("%s:\t%s", r.name, r.score) }
func (r Results) Len() int           { return len(r) }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool { return r[i].score < r[j].score }

var (
	BufMaxString, BufMinString, BufSizeString, Step, Cpu, Parts   string
	BufMax, BufMin, BufSize, Runs  int
	CpuList, PartsList []int
	Copy, Compare, AllTests                           bool
	StepAction                                        byte
	StepAmount                                        int64
	MinStepAmount                                     int64
	Debug                                             bool

	Report Results
)

func init() {
	flag.StringVar(&BufMinString, "bufmin", "1kb", "Run tests on a range of buffer sizes.")
	flag.StringVar(&BufMaxString, "bufmax", "4mb", "Run tests on a range of buffer sizes.")
	flag.StringVar(&BufSizeString, "bufsize", "", "Run benchmarks on a single buffer size, instead of using bufmin/bufmax.")
	flag.StringVar(&Step, "step", "^", "The interval of the buffer range to test.  +Ns, *Ns, ^ (default: ^ or square)")
	flag.StringVar(&Cpu, "cpu", "2", "Comma separated list of cpus to enable for benchmarking (default: runtime.NumCPU())")
	flag.StringVar(&Parts, "parts", "2", "Maximum number of parts for tests that run concurrently on partitions (min: 2)")
	flag.IntVar(&Runs, "runs", 100, "How many times to run each test")
	flag.BoolVar(&Copy, "copy", false, "Run the copy tests")
	flag.BoolVar(&Compare, "compare", false, "Run the compare tests")
	flag.BoolVar(&Debug, "debug", false, "Expose debugging information at http://localhost:8080/debug/pprof during benchmarking")
}

func print_usage() {
	fmt.Println("bufcpy usage:")
	fmt.Println("go build main.go && ./main")
	flag.PrintDefaults()
	fmt.Println()
}

// modifies i by ref
func nextBufSize(i *int) {
	switch StepAction {
	case '+': *i += int(StepAmount); break
	case '*': *i *= int(StepAmount); break
	case '^': *i *= *i
	}
}

func main() {
	var err error

	flag.Parse()
	if flag.NFlag() == 0 {
		print_usage()
		os.Exit(0)
	}

	fmt.Println("Running bufcpy auditor\n")

	if Debug {
		// fire up the debugging server
		go func() { log.Fatal(http.ListenAndServe(":8080", nil)) }()
	}

	t1, _ := utils.ParseHumanReadableSize(BufMinString)
	t2, _ := utils.ParseHumanReadableSize(BufMaxString)
	t3, _ := utils.ParseHumanReadableSize(BufSizeString)
	BufMin, BufMax, BufSize = int(t1), int(t2), int(t3)

	if BufSize > 0 {
		BufMin, BufMax = BufSize, BufSize
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

	for _, ct := range strings.Split(Cpu, ",") {
		n, err := strconv.ParseInt(ct, 10, 0)
		if err != nil {
			fmt.Printf("Setting to default cpu: %d\n", runtime.NumCPU())
			n = int64(runtime.NumCPU())
		}
		CpuList = append(CpuList, int(n))
	}
	for _, ct := range strings.Split(Parts, ",") {
		n, err := strconv.ParseInt(ct, 10, 0)
		if err != nil {
			fmt.Println("Setting to default parts: 2")
			n = int64(2)
		}
		PartsList = append(PartsList, int(n))
	}

	for bufsize := BufMin; bufsize <= BufMax; nextBufSize(&bufsize) {

		// native/memcpy aren't affected by cpus/parts
		// we can run them from here
		time_to_beat := time.Duration(0)

		to, from := make([]byte, bufsize), make([]byte, bufsize)
		utils.FillBytes(from)

		for i, copyfunc := range []func([]byte, []byte){bufcpy.NativeCopy, bufcpy.CgoMemcpy} {

			sum := time.Duration(0)
			for i := 0; i < Runs; i++ {
				utils.ZeroBytes(to)
				runtime.GC()
				start := time.Now()
				copyfunc(to, from)
				end := time.Now()
				sum += end.Sub(start)
			}
			result := &Result{
				name:  fmt.Sprintf("%50s (bufsize=%s)", runtime.FuncForPC(reflect.ValueOf(copyfunc).Pointer()).Name(), utils.FormatHumanReadableSize(int64(bufsize), 1)),
				score: sum / time.Duration(Runs),
			}
			fmt.Println(result)
			Report = append(Report, result)
			if i == 0 {
				time_to_beat = Report[len(Report)-1].score
			}

		}

		for _, cpus := range CpuList {
			runtime.GOMAXPROCS(cpus)
			runtime.GC()

			for _, parts := range PartsList {

				for _, copyfunc := range []func([]byte, []byte, int){
					bufcpy.RecursiveDacCopy,
					bufcpy.RecursiveDacCgoMemcpy,
					bufcpy.PartitionedCopy,
					bufcpy.PartitionedCgoMemcpy,
				} {

					sum := time.Duration(0)
					for i := 0; i < Runs; i++ {
						utils.ZeroBytes(to)
						runtime.GC()
						start := time.Now()
						copyfunc(to, from, parts)
						end := time.Now()
						sum += end.Sub(start)
					}
					result := &Result{
						name:  fmt.Sprintf("%35s (cpus=%d,bufsize=%s,parts=%d)", runtime.FuncForPC(reflect.ValueOf(copyfunc).Pointer()).Name(), cpus, utils.FormatHumanReadableSize(int64(bufsize), 1), parts),
						score: sum / time.Duration(Runs),
					}
					fmt.Println(result)
					Report = append(Report, result)

				}

			}

		}

		fmt.Println("\nTop 10:")
		sort.Sort(Report)
		for i, r := range Report {
			fmt.Println(r)
			if i > 10 {
				break
			}
		}
		fmt.Println()
		top := Report[0]
		if top.score < time_to_beat {
			fmt.Printf("Delta: %s %s  - %0.2f%% faster than NativeCopy (%s)\n", time_to_beat-top.score, strings.TrimSpace(top.name), (1-(float64(top.score)/float64(time_to_beat)))*100, time_to_beat)
		}
	}

}
