package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"

	lxr "github.com/pegnet/LXRHash"
	"github.com/shirou/gopsutil/mem"
)

var lx lxr.LXRHash
var opr []byte

func runtest(miners int) {

	stoppers := make([]chan bool, miners)
	for i := range stoppers {
		stoppers[i] = make(chan bool, 1)
	}

	results := make(chan int, miners)

	start := time.Now()
	for i := 0; i < miners; i++ {
		go func(id int) {
			hashes := 0
			n := newninc(id)
			for {
				select {
				case <-stoppers[id]:
					results <- hashes
					return
				default:
				}

				lx.Hash(append(opr, n.Nonce...))
				hashes++
				n.next()
			}
		}(i)
	}
	for i := 0; i < 60; i++ {
		fmt.Printf("Running 60 second test with %d miners: %ds", miners, i)
		time.Sleep(time.Second)
		fmt.Print("\r")
	}
	fmt.Println()

	percent, _ := cpu.Percent(0, true)
	for _, s := range stoppers {
		s <- true
	}

	total := 0
	for i := 0; i < miners; i++ {
		total += <-results
	}
	dur := time.Since(start)
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
	c, _ := cpu.Info()
	fmt.Printf("%10s = %s\n", "CPU", c[0].ModelName)
	cores := runtime.NumCPU()
	fmt.Printf("%10s = %d\n", "Cores", cores)
	v, _ := mem.VirtualMemory()
	fmt.Printf("%10s = %d MB\n", "Total RAM", v.Total/1024/1024)

	h, _ := host.Info()

	fmt.Printf("%10s = %s\n", "OS", h.OS)
	fmt.Printf("%10s = %s\n", "Platform", h.Platform)
	fmt.Println("=====================================")

	for i := 1; i <= cores+2; i++ {
		runtest(i)
	}
}
