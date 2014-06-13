package main

import (
	"flag"
	"fmt"
	"github.com/billhathaway/volleyLT/common"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"net/rpc"
	"os"
	"strings"
	"sync"
)

func main() {
	var url string
	var requests int
	var concurrency int
	var keepAlive bool
	var agentCount int
	var agents string
	var etcdServers string
	var useEtcd bool
	var etcdPath string
	var fullReport bool
	var summaryReport bool
	var verbose bool

	flag.IntVar(&requests, "n", 10, "number of requests")
	flag.IntVar(&concurrency, "c", 1, "amount of concurrency")
	flag.BoolVar(&keepAlive, "k", true, "use HTTP keepalive")
	flag.IntVar(&agentCount, "agentCount", 1, "number of agents to use (only needs to be set when using etcd)")
	flag.StringVar(&agents, "agents", "localhost:9876", "csv of server:port pairs")
	flag.StringVar(&etcdServers, "etcdServers", "http://localhost:4001", "csv of etcd URLs")
	flag.StringVar(&etcdPath, "etcdPath", "/volleyAgent", "etcd path for registered agents")
	flag.BoolVar(&useEtcd, "etcd", false, "use etcd for agents (otherwise specify with -agents)")
	flag.BoolVar(&fullReport, "full", false, "show full report")
	flag.BoolVar(&summaryReport, "sum", true, "show summary report")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("ERROR: url must be supplied as final argument")
		flag.PrintDefaults()
		os.Exit(1)

	}
	url = flag.Arg(0)

	var servers []string

	if useEtcd {
		client := etcd.NewClient(strings.Split(etcdServers, ","))
		response, err := client.Get(etcdPath, false, true)
		if err != nil {
			log.Fatal("Error talking to etcd servers=%s path=%d err=%s\n", etcdServers, etcdPath, err.Error())
		}
		for _, node := range response.Node.Nodes {
			servers = append(servers, strings.TrimPrefix(node.Key, etcdPath+"/"))
		}
	} else {
		servers = strings.Split(agents, ",")
		if agentCount == 1 && len(servers) > 1 {
			agentCount = len(servers)
		}
	}

	if agentCount > len(servers) {
		log.Fatalf("Error wanted %d agents but only able to find %d (%v)\n", agentCount, len(servers), servers)
	}

	loadTestRequest := volley.Request{
		Url:               url,
		Requests:          requests,
		DisableKeepAlives: !keepAlive,
		Concurrency:       concurrency,
	}

	responses := make([]*volley.SessionResponse, agentCount)
	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(agentCount)
	for i := 0; i < agentCount; i++ {
		go func(i int) {
			loadTestResponse := &volley.SessionResponse{}
			if verbose {
				log.Printf("connecting to %s\n", servers[i])
			}
			client, err := rpc.DialHTTP("tcp", servers[i])
			if err != nil {
				log.Fatal("dialing:", err)
			}
			err = client.Call("Controller.Execute", loadTestRequest, loadTestResponse)
			if err != nil {
				log.Fatal("calling %s returned :", servers[i], err)
			}
			if verbose {
				log.Printf("received response from %s\n", servers[i])
			}
			mutex.Lock()
			responses[i] = loadTestResponse
			mutex.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()

	if fullReport {
		raw := &rawReporter{}
		fmt.Println(raw.report(responses))
	}

	if summaryReport {
		stats := &statReporter{}
		fmt.Println(stats.report(responses))
	}

}
