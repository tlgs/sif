package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type FetchFunc func() string

type Info struct {
	Value string
	fetch FetchFunc
}

func (i *Info) Fetch(wg *sync.WaitGroup) {
	i.Value = i.fetch()
	wg.Done()
}

func fetchOS() string {
	text, _ := os.ReadFile("/etc/os-release")

	re := regexp.MustCompile(`PRETTY_NAME="+(.*)"+\n`)
	osName := string(re.FindSubmatch(text)[1])

	return osName
}

func fetchKernel() string {
	out, _ := exec.Command("uname", "-r").CombinedOutput()
	kernel := strings.TrimSpace(string(out))

	return kernel
}

func fetchHost() string {
	text, _ := os.ReadFile("/sys/devices/virtual/dmi/id/product_version")
	host := strings.TrimSpace(string(text))

	return host
}

func fetchCPU() string {
	text, _ := os.ReadFile("/proc/cpuinfo")

	re := regexp.MustCompile(`model name\s*: (.*)\n`)
	cpu := string(re.FindSubmatch(text)[1])

	cpu = regexp.MustCompile(`\(.*\)`).ReplaceAllString(cpu, "")
	cpu = regexp.MustCompile(`@.*`).ReplaceAllString(cpu, "")
	cpu = strings.TrimSuffix(cpu, " CPU ")

	return cpu
}

func fetchGPU() string {
	out, _ := exec.Command("lspci").CombinedOutput()

	re := regexp.MustCompile(`VGA.*: (.*) \(`)
	gpu := string(re.FindSubmatch(out)[1])

	gpu = strings.ReplaceAll(gpu, "Corporation ", "")

	return gpu
}

func fetchWM() string {
	cmd := exec.Command("xprop", "-root", "_NET_SUPPORTING_WM_CHECK")
	out, _ := cmd.CombinedOutput()

	rootID := regexp.MustCompile(`.* `).ReplaceAllString(string(out), "")

	cmd = exec.Command("xprop", "-id", string(rootID), "_NET_WM_NAME")
	out, _ = cmd.CombinedOutput()

	re := regexp.MustCompile(`.*= "+(.*)"+`)
	wm := string(re.FindSubmatch(out)[1])

	return wm
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

	sys := map[string]*Info{
		"os":     {"", fetchOS},
		"kernel": {"", fetchKernel},
		"host":   {"", fetchHost},
		"cpu":    {"", fetchCPU},
		"gpu":    {"", fetchGPU},
		"wm":     {"", fetchWM},
	}

	var wg sync.WaitGroup
	for _, info := range sys {
		wg.Add(1)
		go info.Fetch(&wg)
	}
	wg.Wait()

	style := ParseDisplayAttrs(*s)
	f := func(v string) string { return style + v + "\x1b[0m" }

	fmt.Println()
	fmt.Printf("%15s │ %s\n", "os", f(sys["os"].Value))
	fmt.Printf("%15s │ %s\n", "kernel", f(sys["kernel"].Value))
	fmt.Printf("%15s │\n", "")
	fmt.Printf("%15s │ %s\n", "host", f(sys["host"].Value))
	fmt.Printf("%15s │ %s\n", "cpu", f(sys["cpu"].Value))
	fmt.Printf("%15s │ %s\n", "gpu", f(sys["gpu"].Value))
	fmt.Printf("%15s │\n", "")
	fmt.Printf("%15s │ %s\n", "wm", f(sys["wm"].Value))
}
