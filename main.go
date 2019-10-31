package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"

	lxr "github.com/pegnet/LXRHash"
	"github.com/shirou/gopsutil/mem"
)

var lx lxr.LXRHash
var opr []byte

func runtest(miners int) {

	running := true
	done := make(chan int, miners)

	var hashes uint64

	start := time.Now()
	for i := 0; i < miners; i++ {
		go func(id int) {
			n := newninc(id)
			for running {
				lx.Hash(append(opr, n.Nonce...))
				atomic.AddUint64(&hashes, 1)
				n.next()
			}
			done <- 1
		}(i)
	}

	for i := 0; i < 60; i++ {
		fmt.Printf("Running 60 second test with %d miners: %ds", miners, i)
		time.Sleep(time.Second)
		fmt.Print("\r")
	}
	fmt.Println()

	percent, _ := cpu.Percent(0, true)
	running = false
	dur := time.Since(start)
	total := hashes

	for i := 0; i < miners; i++ {
		<-done
	}

	rate := float64(total) / dur.Seconds()
	fmt.Println("Finished test in", dur)
	fmt.Printf("%15s:", "CPU Usage")
	for idx, prct := range percent {
		fmt.Printf("[%d: %.2f] ", idx, prct)
	}
	fmt.Println()
	fmt.Printf("%15s: %d\n", "Total hashes", total)
	fmt.Printf("%15s: %d\n", "Total hashrate", int(rate))
	fmt.Printf("%15s: %d\n", "Per miner", int(rate/float64(miners)))
	fmt.Println("=====================================")

}

func main() {
	fmt.Printf("Benchmarking LXR Hash\n")
	fmt.Println("=====================================")
	lx.Verbose(true)
	lx.Init(lxr.Seed, lxr.MapSizeBits, lxr.HashSize, lxr.Passes)
	opr = lx.Hash([]byte("foo"))

	fmt.Printf("%10s = %x, %d, %d, %d\n", "Hash Init", lxr.Seed, lxr.MapSizeBits, lxr.HashSize, lxr.Passes)
	ctx := context.Background()
	to, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	c, err := cpu.InfoWithContext(to)
	if err != nil {
		fmt.Println("There was an error querying the CPU info. Please try again.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%10s = %s\n", "CPU", c[0].ModelName)
	cores := runtime.NumCPU()
	fmt.Printf("%10s = %d\n", "Cores", cores)
	v, _ := mem.VirtualMemory()
	fmt.Printf("%10s = %d MB\n", "Total RAM", v.Total/1024/1024)

	h, _ := host.Info()

	fmt.Printf("%10s = %s\n", "OS", h.OS)
	fmt.Printf("%10s = %s\n", "Platform", h.Platform)
	fmt.Println("=====================================")

	runtest(1)
	if cores > 1 {
		runtest(cores)
	}
}
