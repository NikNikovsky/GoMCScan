package main

import (
	"GoMCScan/mcping"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aherve/gopool"
)
var pinged int
var completed int
var found int
var startTime time.Time

const usage = `Usage of MCScan:
    MCScan [-T Threads] [-t Timeout] [-p PortRange] [-o output]
Options:
    -T, --threads number of threads to use
    -t, --timeout timeout in seconds
    -h, --help prints help information 
`

func main() {
	var threads int
	var timeout int
	flag.IntVar(&threads, "T", 1000, "number of threads to use")
	flag.IntVar(&threads, "threads", 1000, "number of threads to use")
	flag.IntVar(&timeout, "t", 1, "timeout in seconds")
	flag.IntVar(&timeout, "timeout", 1, "timeout in seconds")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	pool := gopool.NewPool(threads)
	ports := []uint16{25565}
	loopBlock(timeout, 176, 9, ports, pool)
	pool.Wait()
}

func loopBlock(timeout int, a uint8, b uint8, ports []uint16, pool *gopool.GoPool) {
	startTime = time.Now()
	for _, port := range ports {
		for j := 0; j < 255; j++ {
			for k := 0; k < 255; k++ {
				var ip = fmt.Sprintf("%v.%v.%v.%v", a, b, j, k)
				pool.Add(1)
				go pingIt(ip, port, timeout, pool)
				pinged++
			}
		}
	}
}

func pingIt(ip string, port uint16, timeout int, pool *gopool.GoPool) {
	defer pool.Done()
	data, _, err := mcping.PingWithTimeout(ip, port, time.Duration(timeout)*time.Second)
	completed++
	if err == nil {
		fmt.Println("Found")
		sampleBytes, _ := json.Marshal(data.Sample)
		sample := string(sampleBytes)
		if sample == "null" {
			sample = "[]"
		}
		formatted := fmt.Sprintf("{\"Ip\":\"%v:%v\", \"Version\":%q, \"Motd\":%q, \"Players:%v/%v\", \"Sample\":%v}", ip, port, data.Version, data.Motd, data.PlayerCount.Online, data.PlayerCount.Max, sample)
		fmt.Println(formatted)
		found++
		fmt.Printf("%v/%v, %v percent complete\n", completed, pinged, uint8(100*float64(completed)/float64(pinged)))
		fmt.Printf("Time Elapsed: %v min, finding rate: %v servers per second", time.Since(startTime).Minutes(), int(float64(found)/float64(time.Since(startTime).Seconds())))
		record(formatted)
	} else {
		//fmt.Println(err)
	}
}

func record(data string) {
	f, err := os.OpenFile("out/scan.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(data + "\n"); err != nil {
		log.Println(err)
	}
}
