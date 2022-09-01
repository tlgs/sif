package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type result struct {
	key   string
	value string
	err   error
}

func fetchOS(c chan<- result) {
	text, err := os.ReadFile("/etc/os-release")
	if err != nil {
		c <- result{err: err}
		return
	}

	re := regexp.MustCompile(`PRETTY_NAME="+(.*)"+\n`)
	osName := string(re.FindSubmatch(text)[1])

	c <- result{"os", osName, nil}
}

func fetchKernel(c chan<- result) {
	text, err := os.ReadFile("/proc/version")
	if err != nil {
		c <- result{err: err}
		return
	}

	kernel := strings.Split(string(text), " ")[2]

	c <- result{"kernel", kernel, nil}
}

func fetchHost(c chan<- result) {
	text, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_version")
	if err != nil {
		c <- result{err: err}
		return
	}

	host := strings.TrimSpace(string(text))

	c <- result{"host", host, nil}
}

func fetchCPU(c chan<- result) {
	text, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		c <- result{err: err}
		return
	}

	re := regexp.MustCompile(`model name\s*: (.*)\n`)
	cpu := string(re.FindSubmatch(text)[1])

	cpu = regexp.MustCompile(`\(.*\)|@.*`).ReplaceAllString(cpu, "")
	cpu = strings.TrimSuffix(cpu, " CPU ")

	c <- result{"cpu", cpu, nil}
}

func fetchGPU(c chan<- result) {
	out, err := exec.Command("lspci").CombinedOutput()
	if err != nil {
		c <- result{err: err}
		return
	}

	re := regexp.MustCompile(`VGA.*: (.*) \(`)
	gpu := string(re.FindSubmatch(out)[1])

	gpu = strings.ReplaceAll(gpu, "Corporation ", "")

	c <- result{"gpu", gpu, nil}
}

func fetchMem(c chan<- result) {
	text, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		c <- result{err: err}
		return
	}

	re := regexp.MustCompile(`MemTotal:\s*(\d*) kB`)
	n, err := strconv.Atoi(string(re.FindSubmatch(text)[1]))
	if err != nil {
		c <- result{err: err}
		return
	}

	mem := fmt.Sprintf("%v MiB", n/1024)

	c <- result{"mem", mem, nil}
}

func main() {
	fetchers := []func(chan<- result){fetchOS, fetchKernel, fetchHost, fetchCPU, fetchGPU, fetchMem}

	c := make(chan result)
	for _, f := range fetchers {
		go f(c)
	}

	sys := make(map[string]string)
	for range fetchers {
		x := <-c
		if x.err != nil {
			panic(x.err)
		}

		sys[x.key] = x.value
	}

	fmt.Println()
	fmt.Printf("%15v │ %v\n", "os", sys["os"])
	fmt.Printf("%15v │ %v\n", "kernel", sys["kernel"])
	fmt.Printf("%15v │\n", "")
	fmt.Printf("%15v │ %v\n", "host", sys["host"])
	fmt.Printf("%15v │ %v\n", "cpu", sys["cpu"])
	fmt.Printf("%15v │ %v\n", "gpu", sys["gpu"])
	fmt.Printf("%15v │ %v\n", "mem", sys["mem"])
	fmt.Println()
}
