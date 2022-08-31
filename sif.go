package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Info struct {
	Key   string
	Value string
}

func fetchOS(c chan<- Info) {
	text, _ := os.ReadFile("/etc/os-release")

	re := regexp.MustCompile(`PRETTY_NAME="+(.*)"+\n`)
	osName := string(re.FindSubmatch(text)[1])

	c <- Info{"os", osName}
}

func fetchKernel(c chan<- Info) {
	text, _ := os.ReadFile("/proc/version")
	kernel := strings.Split(string(text), " ")[2]

	c <- Info{"kernel", kernel}
}

func fetchHost(c chan<- Info) {
	text, _ := os.ReadFile("/sys/devices/virtual/dmi/id/product_version")
	host := strings.TrimSpace(string(text))

	c <- Info{"host", host}
}

func fetchCPU(c chan<- Info) {
	text, _ := os.ReadFile("/proc/cpuinfo")

	re := regexp.MustCompile(`model name\s*: (.*)\n`)
	cpu := string(re.FindSubmatch(text)[1])

	cpu = regexp.MustCompile(`\(.*\)|@.*`).ReplaceAllString(cpu, "")
	cpu = strings.TrimSuffix(cpu, " CPU ")

	c <- Info{"cpu", cpu}
}

func fetchGPU(c chan<- Info) {
	out, _ := exec.Command("lspci").CombinedOutput()

	re := regexp.MustCompile(`VGA.*: (.*) \(`)
	gpu := string(re.FindSubmatch(out)[1])

	gpu = strings.ReplaceAll(gpu, "Corporation ", "")

	c <- Info{"gpu", gpu}
}

func fetchMem(c chan<- Info) {
	text, _ := os.ReadFile("/proc/meminfo")

	re := regexp.MustCompile(`MemTotal:\s*(\d*) kB`)
	x, _ := strconv.Atoi(string(re.FindSubmatch(text)[1]))

	mem := fmt.Sprintf("%v MiB", x/1024)

	c <- Info{"mem", mem}
}

func main() {
	fetches := []func(chan<- Info){fetchOS, fetchKernel, fetchHost, fetchCPU, fetchGPU, fetchMem}

	c := make(chan Info)
	for _, f := range fetches {
		go f(c)
	}

	sys := make(map[string]string)
	for i := 0; i < len(fetches); i++ {
		x := <-c
		sys[x.Key] = x.Value
	}

	fmt.Println()
	fmt.Printf("%15v │ %v\n", "os", sys["os"])
	fmt.Printf("%15v │ %v\n", "kernel", sys["kernel"])
	fmt.Printf("%15v │\n", "")
	fmt.Printf("%15v │ %v\n", "host", sys["host"])
	fmt.Printf("%15v │ %v\n", "cpu", sys["cpu"])
	fmt.Printf("%15v │ %v\n", "gpu", sys["gpu"])
	fmt.Printf("%15v │ %v\n", "mem", sys["mem"])
}
