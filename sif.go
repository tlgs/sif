package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func fetchOS(c chan string) {
	text, _ := os.ReadFile("/etc/os-release")

	re := regexp.MustCompile(`PRETTY_NAME="+(.*)"+\n`)
	osName := string(re.FindSubmatch(text)[1])

	c <- osName
}

func fetchKernel(c chan string) {
	out, _ := exec.Command("uname", "-r").CombinedOutput()
	kernel := strings.TrimSpace(string(out))

	c <- kernel
}

func fetchHost(c chan string) {
	text, _ := os.ReadFile("/sys/devices/virtual/dmi/id/product_version")
	host := strings.TrimSpace(string(text))

	c <- host
}

func fetchCPU(c chan string) {
	text, _ := os.ReadFile("/proc/cpuinfo")

	re := regexp.MustCompile(`model name\s*: (.*)\n`)
	cpu := string(re.FindSubmatch(text)[1])

	cpu = regexp.MustCompile(`\(.*\)|@.*`).ReplaceAllString(cpu, "")
	cpu = strings.TrimSuffix(cpu, " CPU ")

	c <- cpu
}

func fetchGPU(c chan string) {
	out, _ := exec.Command("lspci").CombinedOutput()

	re := regexp.MustCompile(`VGA.*: (.*) \(`)
	gpu := string(re.FindSubmatch(out)[1])

	gpu = strings.ReplaceAll(gpu, "Corporation ", "")

	c <- gpu
}

func fetchWM(c chan string) {
	cmd := exec.Command("xprop", "-root", "_NET_SUPPORTING_WM_CHECK")
	out, _ := cmd.CombinedOutput()

	rootID := regexp.MustCompile(`.* `).ReplaceAllString(string(out), "")

	cmd = exec.Command("xprop", "-id", string(rootID), "_NET_WM_NAME")
	out, _ = cmd.CombinedOutput()

	re := regexp.MustCompile(`.*= "+(.*)"+`)
	wm := string(re.FindSubmatch(out)[1])

	c <- wm
}

func ParseDisplayAttrs(s string) string {
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
	s := flag.String("s", "italic", "comma-separated list of display attributes")
	flag.Parse()

	var c []chan string
	for i := 0; i < 6; i++ {
		c = append(c, make(chan string))
	}

	go fetchOS(c[0])
	go fetchKernel(c[1])
	go fetchHost(c[2])
	go fetchCPU(c[3])
	go fetchGPU(c[4])
	go fetchWM(c[5])

	style := ParseDisplayAttrs(*s)
	f := func(v string) string { return style + v + "\x1b[0m" }

	fmt.Println()
	fmt.Printf("%15s │ %s\n", "os", f(<-c[0]))
	fmt.Printf("%15s │ %s\n", "kernel", f(<-c[1]))
	fmt.Printf("%15s │\n", "")
	fmt.Printf("%15s │ %s\n", "host", f(<-c[2]))
	fmt.Printf("%15s │ %s\n", "cpu", f(<-c[3]))
	fmt.Printf("%15s │ %s\n", "gpu", f(<-c[4]))
	fmt.Printf("%15s │\n", "")
	fmt.Printf("%15s │ %s\n", "wm", f(<-c[5]))
}
