package main

import (
	"bytes"
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

func (r *Result) String() string     { return fmt.Sprintf("%s: %s", r.name, r.score) }
func (r Results) Len() int           { return len(r) }
func (r Results) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Results) Less(i, j int) bool { return r[i].score < r[j].score }

var (
	BufMaxString, BufMinString, BufSizeString, Step, Cpu, Parts   string
	BufMax, BufMin, BufSize, Runs, Top int
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
	flag.StringVar(&Step, "step", "^", "The interval of the buffer range to test.  +Ns, *N, ^ (default: ^ or square)")
	flag.StringVar(&Cpu, "cpu", "2", "Comma separated list of cpus to enable for benchmarking (default: runtime.NumCPU())")
	flag.StringVar(&Parts, "parts", "2", "Maximum number of parts for tests that run concurrently on partitions (min: 2)")
	flag.IntVar(&Top, "top", 10, "Show top n results in sorted order")
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

func PrintTopN(report Results, n int) {
	fmt.Printf("\nTop %d:\n", Top)
	for i, r := range Report {
		fields := strings.Fields(r.String())
		fmt.Printf("%-26s %36s %s\n", fields[0], fields[1], fields[2])
		if i >= Top-1 {
			break
		}
	}
	fmt.Println()
}

func main() {
	var err error

	flag.Parse()
	if flag.NFlag() == 0 {
		print_usage()
		os.Exit(0)
	}

	fmt.Println("==============")
	fmt.Println("Bufcpy Auditor")
	fmt.Println("==============\n")

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
		switch StepAction {
		case '+':
			StepAmount, err = utils.ParseHumanReadableSize(Step[1:])
			if err != nil {
				panic("can't read your handwriting.  try +1k")
			}
			if StepAmount < MinStepAmount {
				fmt.Println("Min step amount set at 11kb")
				StepAmount = MinStepAmount
			}
			break
		case '*':
			StepAmount, err = strconv.ParseInt(Step[1:], 10, 64)
			if err != nil {
				panic("can't read your handwriting.  try *2")
			}
			break
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

	hrs := func(i, precision int) string {
		return utils.FormatHumanReadableSize(int64(i), precision)
	}
	fmt.Println("Running benchmarks with the following settings:")
	fmt.Printf("Benchmarks: copy: %t, compare %t\n", Copy, Compare)
	fmt.Printf("bufmin: %d (%s), bufmax: %d (%s), bufsize: %d (%s)\n", BufMin, hrs(BufMin,2), BufMax, hrs(BufMax,2), BufSize, hrs(BufSize,2))
	fmt.Printf("step action: %c, step amount: %d (%s)\n", StepAction, StepAmount, hrs(int(StepAmount), 5))
	fmt.Printf("cpus: %s parts: %s runs: %d\n", utils.SprintIntSlice(CpuList), utils.SprintIntSlice(PartsList), Runs)
	fmt.Printf("Debugging enabled: %t\n", Debug)
	if Debug {
		fmt.Println("Debugging info:  http://localhost:8080/debug/pprof")
	}
	fmt.Printf("Showing Top %d results in sorted order\n", Top)
	fmt.Println("\n")


	if Copy {
		native := []func([]byte, []byte){
			bufcpy.NativeCopy,
			bufcpy.CgoMemcpy,
		}
		concurrent := []func([]byte, []byte, int){
			bufcpy.RecursiveDacCopy,
			bufcpy.RecursiveDacCgoMemcpy,
			bufcpy.PartitionedCopy,
			bufcpy.PartitionedCgoMemcpy,
		}

		for bufsize := BufMin; bufsize <= BufMax; nextBufSize(&bufsize) {

			fmt.Println("\n=== New Copy Benchmark ===\n")

			// native/memcpy aren't affected by cpus/parts
			// we can run them from this level of the loop
			to, from, time_to_beat := make([]byte, bufsize), make([]byte, bufsize), time.Duration(0)
			utils.FillBytes(from)

			for i, copyfunc := range native {

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
					name:  fmt.Sprintf("%-17s (bufsize=%s)", runtime.FuncForPC(reflect.ValueOf(copyfunc).Pointer()).Name(), utils.FormatHumanReadableSize(int64(bufsize), 1)),
					score: sum / time.Duration(Runs),
				}
				fmt.Println(result)
				Report = append(Report, result)
				if i == 0 {
					time_to_beat = Report[len(Report)-1].score
				}
			}
			fmt.Println()

			for _, copyfunc := range concurrent {

				for _, parts := range PartsList {

					for _, cpus := range CpuList {
						runtime.GOMAXPROCS(cpus)
						runtime.GC()

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
							name:  fmt.Sprintf("%-30s (cpus=%d,bufsize=%s,parts=%d)", runtime.FuncForPC(reflect.ValueOf(copyfunc).Pointer()).Name(), cpus, utils.FormatHumanReadableSize(int64(bufsize), 2), parts),
							score: sum / time.Duration(Runs),
						}
						fmt.Println(result)
						Report = append(Report, result)

					}

				}

			}

			sort.Sort(Report)
			top := Report[0]
			PrintTopN(Report, Top)

			name := strings.Fields(top.name)
			fmt.Printf("Copy Winner:  %s %s %s\n", name[0], name[1], top.score)
			if top.score < time_to_beat {
				fmt.Printf("Delta:  %10s %0.2f%% faster than copy() (%s)\n", time_to_beat-top.score, (1-(float64(top.score)/float64(time_to_beat)))*100, time_to_beat)
			}

			// on to the next one
			Report = nil
		}
	}

	if Compare {
		native := []func([]byte, []byte) bool {
			bytes.Equal,
			bufcpy.CgoMemcmp,
		}
		concurrent := []func([]byte, []byte, int) bool {
			bufcpy.PartitionedEqual,
			bufcpy.PartitionedCgoMemcmp,
			bufcpy.RecursiveDacEqual,
			bufcpy.RecursiveDacCgoMemcmp,
		}

		for bufsize := BufMin; bufsize <= BufMax; nextBufSize(&bufsize) {

			fmt.Println("\n=== New Compare Benchmark ===\n")

			// native/memcpy aren't affected by cpus/parts
			// we can run them from this level of the loop
			to, from, time_to_beat := make([]byte, bufsize), make([]byte, bufsize), time.Duration(0)

			fmt.Println("Native:")

			for i, comparefunc := range native {

				sum := time.Duration(0)
				for i := 0; i < Runs; i++ {
					runtime.GC()
					start := time.Now()
					comparefunc(to, from)
					end := time.Now()
					sum += end.Sub(start)
				}
				result := &Result{
					name:  fmt.Sprintf("%-17s (bufsize=%s)", runtime.FuncForPC(reflect.ValueOf(comparefunc).Pointer()).Name(), utils.FormatHumanReadableSize(int64(bufsize), 1)),
					score: sum / time.Duration(Runs),
				}
				fmt.Println(result)
				Report = append(Report, result)
				if i == 0 {
					time_to_beat = Report[len(Report)-1].score
				}
			}
			fmt.Println("\n\nConcurrent:")

			for _, comparefunc := range concurrent {

				for _, parts := range PartsList {

					for _, cpus := range CpuList {
						runtime.GOMAXPROCS(cpus)
						runtime.GC()

						sum := time.Duration(0)
						for i := 0; i < Runs; i++ {
							runtime.GC()
							start := time.Now()
							comparefunc(to, from, parts)
							end := time.Now()
							sum += end.Sub(start)
						}
						result := &Result{
							name:  fmt.Sprintf("%-30s (cpus=%d,bufsize=%s,parts=%d)", runtime.FuncForPC(reflect.ValueOf(comparefunc).Pointer()).Name(), cpus, utils.FormatHumanReadableSize(int64(bufsize), 2), parts),
							score: sum / time.Duration(Runs),
						}
						fmt.Println(result)
						Report = append(Report, result)

					}

				}

			}

			sort.Sort(Report)
			top := Report[0]
			PrintTopN(Report, Top)
			name := strings.Fields(top.name)
			fmt.Printf("\nEqual Winner:  %s %s %s\n", name[0], name[1], top.score)
			if top.score < time_to_beat {
				fmt.Printf("Delta:  %10s %0.2f%% faster than bytes.Equal (%s)\n", time_to_beat-top.score, (1-(float64(top.score)/float64(time_to_beat)))*100, time_to_beat)
			}

			// on to the next one
			Report = nil
		}
	}


}
