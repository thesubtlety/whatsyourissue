package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var verbose bool
var sshTimeout int

func main() {

	help := flag.Bool("h", false, "Print help")
	target := flag.String("t", "", "target ip or cidr range")
	threads := flag.Int("n", 100, "number of threads")
	flag.BoolVar(&verbose, "v", false, "print hosts without an issue/motd")
	flag.IntVar(&sshTimeout, "timeout", 10, "ssh timeout in seconds")
	flag.Parse()

	info, _ := os.Stdin.Stat()
	if *help || (*target == "" && info.Mode()&os.ModeNamedPipe == 0) {
		flag.PrintDefaults()
		os.Exit(0)
	}

	iplist := []string{}
	targets := make(chan string, *threads)
	var wg sync.WaitGroup

	//read input from pipeline else read args
	if info.Mode()&os.ModeNamedPipe != 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			val := scanner.Text()
			iplist = append(iplist, val)
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("error parsing input value: %v\n", err)
		}
	} else {
		iplist = append(iplist, *target)
	}

	//start as many workers as the targets channel
	for i := 0; i < cap(targets); i++ {
		go worker(targets, &wg)
	}

	//parse input, send to targets channel
	for _, r := range iplist {
		ips, err := parseCIDR(r)
		if err != nil {
			fmt.Printf("error parsing input value: %v\n", err)
		}
		for _, ip := range ips {
			wg.Add(1)
			targets <- ip
		}
	}

	wg.Wait()
	close(targets)
}

func worker(targets chan string, wg *sync.WaitGroup) {
	for ip := range targets {
		b := getBanner(ip)
		printBanner(ip, b)
		wg.Done()
	}
}

func getBanner(ip string) string {
	dest := fmt.Sprintf("%s:22", ip)
	var banner string

	config := &ssh.ClientConfig{
		User: "whatsyourissue",
		Auth: []ssh.AuthMethod{
			ssh.Password("whatsyourissue"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback: func(message string) error {
			banner = message
			return nil
		},
		Timeout: time.Duration(sshTimeout) * time.Second,
	}

	conn, err := ssh.Dial("tcp", dest, config)
	if err == nil {
		fmt.Printf("%v\n", conn)
		conn.Close()
	}

	return banner
}

func printBanner(ip string, banner string) {
	if banner != "" || verbose {
		re := regexp.MustCompile(`\r?\n`)
		r := re.ReplaceAllString(banner, "\n\t\t\t")
		fmt.Printf("%s\t\t%s\n", ip, r)
	}
}

//https://gist.github.com/kotakanbe/d3059af990252ba89a82
func parseCIDR(target string) ([]string, error) {

	if !strings.Contains(target, "/") {
		target = fmt.Sprintf("%v/32", target)
	}

	ip, ipnet, err := net.ParseCIDR(target)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
