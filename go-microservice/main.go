package main

import (
	"fmt"
	"github.com/shyhirt/load-balancer-nginx-go-php-fpm/go-microservice/store"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	portNums := 1
	fromAddr := "127.0.0.1:8888"
	var (
		toAddr []string
		toUrls []*url.URL
	)
	for i := 0; i <= portNums; i++ {
		toAddr = append(toAddr, fmt.Sprintf("127.0.0.1:801%d", i))
		toUrls = append(toUrls, parseToUrl(toAddr[i]))
	}

	proxy := loadBalancingReverseProxy(toUrls...)
	log.Println("Starting proxy server on", fromAddr)
	go func() {
		handler := http.HandlerFunc(handleRequest)
		http.Handle("/", handler)
		http.ListenAndServe(":9090", nil)
	}()
	if err := http.ListenAndServe(fromAddr, proxy); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func parseToUrl(addr string) *url.URL {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	toUrl, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}
	return toUrl
}

func loadBalancingReverseProxy(targets ...*url.URL) *httputil.ReverseProxy {
	hostErr := parseToUrl("http://localhost:9090")
	var targetNum int
	ports := len(targets)
	rpc := 1
	if tmp, err := strconv.Atoi(os.Getenv("REQ_PER_SEC")); err == nil {
		rpc = tmp
	}
	log.Println("RATE LIMIT", rpc, "req per seconds")
	store := store.New(rpc)
	director := func(req *http.Request) {
		if targetNum >= ports-1 {
			targetNum = 0
		}
		target := targets[targetNum]
		if store.Allow(req.UserAgent()) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path, req.URL.RawPath = target.Path, target.RawPath
		} else {
			req.URL = hostErr
			req.Method = "GET"
		}
		targetNum++
	}
	return &httputil.ReverseProxy{Director: director}
}
func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte{})
	return
}
