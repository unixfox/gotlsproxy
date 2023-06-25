package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	http "github.com/saucesteals/fhttp"
	"github.com/saucesteals/mimic"
)

var listenAddress string
var timeout int
var printErrors bool

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	_, errWrite := w.Write([]byte(err.Error()))
	if errWrite != nil {
		log.Printf("ERROR Proxy2Client: %v", errWrite)
	}
}

var (
	latestVersion = mimic.MustGetLatestVersion(mimic.PlatformWindows)
)

func hello(w http.ResponseWriter, req *http.Request) {

	req.URL.Scheme = "https"

	m, _ := mimic.Chromium(mimic.BrandChrome, latestVersion)
	ua := fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", m.Version())

	client := &http.Client{
		Transport: m.ConfigureTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}),
	}

	do_req, _ := http.NewRequest(req.Method, fmt.Sprintf("%s", req.URL), req.Body)

	do_req.Header = req.Header
	do_req.Header.Del("Accept-Encoding")
	do_req.Header.Del("connection")
	do_req.Header.Del("user-agent")
	do_req.Header.Set("user-agent", ua)

	response, err := client.Do(do_req)

	if err != nil {
		log.Printf("%v", err)
		writeError(w, err)
		return
	}

	defer response.Body.Close()

	if printErrors && (response.StatusCode >= 400) {
		log.Printf("Response status %s", response.Status)
		log.Printf("== request ==")
		log.Printf("%v", req)
		log.Printf("== response ==")
		log.Printf("%v", response)
	}

	for name, h := range response.Header {
		w.Header().Add(name, strings.Join(h, " "))
	}

	w.Header().Del("Content-Encoding")

	w.WriteHeader(response.StatusCode)

	response_body, err := io.ReadAll(response.Body)

	_, err = w.Write(response_body)
	if err != nil {
		log.Printf("ERROR Proxy2Client: %v", err)
	}

}

func main() {
	flag.StringVar(&listenAddress, "bind", "127.0.0.1:8080", "Listening address to bind to")
	flag.IntVar(&timeout, "timeout", 60, "Request timeout")
	flag.BoolVar(&printErrors, "print-errors", false, "Print request and response when an error (4xx and 5xx) is returned from upstream server")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s [flags]

Flags:
`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	http.HandleFunc("/", hello)
	log.Printf("Listening on %s", listenAddress)
	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
