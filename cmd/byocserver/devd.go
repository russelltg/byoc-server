package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/cortesi/termlog"
	"github.com/russelltg/byoc-server"
	"github.com/toqueteos/webbrowser"
	"gopkg.in/alecthomas/kingpin.v2"
)

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func main() {

	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Version(devd.Version)

	kingpin.Parse()

	hdrs := make(http.Header)

	// get executable path, we want to use that dir
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := path.Dir(ex)

	routes := []string{exPath}

	dd := devd.Devd{
		// Shaping
		Latency:  0,
		DownKbps: 0,
		UpKbps:   0,

		AddHeaders: &hdrs,

		// Livereload
		LivereloadRoutes: false,
		Livereload:       false,
		WatchPaths:       []string{},
		Excludes:         []string{},

		Credentials: nil,
	}

	if err = dd.AddRoutes(routes, []string{}); err != nil {
		kingpin.Fatalf("%s", err)
	}

	logger := termlog.NewLog()

	// Print out the goods
	ip, err := externalIP()
	if err != nil {
		panic(err)
	}
	fmt.Printf("***************** You're serving games for BYOC! Have people type %v:%v in their browser to download games!\n\n", ip, 8000)

	err = dd.Serve(
		ip,
		8000,
		"",
		logger,
		func(url string) {
			err = webbrowser.Open("http://" + ip + ":8000")
			if err != nil {
				kingpin.Errorf("Failed to open browser: %s", err)
			}
		},
	)
	if err != nil {
		kingpin.Fatalf("%s", err)
	}
}
