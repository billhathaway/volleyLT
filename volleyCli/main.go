package main

import (
	"flag"
	"fmt"
	"github.com/billhathaway/volleyLT/common"
	"log"
	"net/rpc"
)

func main() {
	url := flag.String("url", "http://www.google.com/", "url to hit")
	requests := flag.Int("n", 10, "number of requests")
	concurrency := flag.Int("c", 1, "amount of concurrency")
	keepAlive := flag.Bool("k", false, "use HTTP keepalive")
	servers := flag.String("servers", "localhost:9876", "csv of server:port")
	flag.Parse()
	loadTestRequest := volley.Request{
		Url:               *url,
		Requests:          *requests,
		DisableKeepAlives: !*keepAlive,
		Concurrency:       *concurrency,
	}
	loadTestResponse := &volley.SessionResponse{}
	client, err := rpc.DialHTTP("tcp", *servers)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	client.Call("Controller.Execute", loadTestRequest, loadTestResponse)
	raw := &rawReporter{}
	stats := &statReporter{}

	fmt.Println(raw.report(loadTestResponse))
	fmt.Println(stats.report(loadTestResponse))
}
