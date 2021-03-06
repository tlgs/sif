package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

func fetchWM(c chan<- Info) {
	cmd := exec.Command("xprop", "-root", "_NET_SUPPORTING_WM_CHECK")
	out, _ := cmd.CombinedOutput()

	rootID := regexp.MustCompile(`.* `).ReplaceAllString(string(out), "")

	cmd = exec.Command("xprop", "-id", string(rootID), "_NET_WM_NAME")
	out, _ = cmd.CombinedOutput()

	re := regexp.MustCompile(`.*= "+(.*)"+`)
	wm := string(re.FindSubmatch(out)[1])

	switch wm {
	case "GNOME Shell":
		wm = "Mutter"
	}

	c <- Info{"wm", wm}
}

func parseDisplayAttrs(s string) string {
	const esc = "\x1b["
	ansi := map[string]string{
		"bold":          esc + "1m",
		"faint":         esc + "2m",
		"italic":        esc + "3m",
		"underline":     esc + "4m",
		"blink":         esc + "6m",
		"inverse":       esc + "7m",
		"invisible":     esc + "8m",
		"strikethrough": esc + "9m",

		"black":   esc + "30m",
		"red":     esc + "31m",
		"green":   esc + "32m",
		"yellow":  esc + "33m",
		"blue":    esc + "34m",
		"magenta": esc + "35m",
		"cyan":    esc + "36m",
		"white":   esc + "37m",
	}

	attrs := strings.Split(s, ",")

	var r string
	for _, v := range attrs {
		if seq, ok := ansi[v]; ok {
			r += seq
		}
	}

	return r
}

func main() {
	s := flag.String("s", "magenta", "comma-separated list of display attributes")
	flag.Parse()

	c := make(chan Info)
	go fetchOS(c)
	go fetchKernel(c)
	go fetchHost(c)
	go fetchCPU(c)
	go fetchGPU(c)
	go fetchWM(c)

	style := parseDisplayAttrs(*s)
	f := func(v string) string { return style + v + "\x1b[0m" }

	sys := make(map[string]string)
	for i := 0; i < 6; i++ {
		x := <-c
		sys[x.Key] = x.Value
	}

	fmt.Println()
	fmt.Printf("%15s ??? %s\n", "os", f(sys["os"]))
	fmt.Printf("%15s ??? %s\n", "kernel", f(sys["kernel"]))
	fmt.Printf("%15s ???\n", "")
	fmt.Printf("%15s ??? %s\n", "host", f(sys["host"]))
	fmt.Printf("%15s ??? %s\n", "cpu", f(sys["cpu"]))
	fmt.Printf("%15s ??? %s\n", "gpu", f(sys["gpu"]))
	fmt.Printf("%15s ???\n", "")
	fmt.Printf("%15s ??? %s\n", "wm", f(sys["wm"]))
}
