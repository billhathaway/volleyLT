volleyLT - a distributed load tester
==
### Still early in development, please use a tool like JMeter or Tsung if you want to get real load testing work done for now.


The plan is that the client will talk to etcd to get a list of available load testing agents and then send commands for the agents to execute.
The agents will return the results which the client will summarize.


Running with a single agent
--
1. go get github.com/billhathaway/volleyLT/volleyAgent
2. go get github.com/billhathaway/volleyLT/volleyCli
3. volleyAgent & # start agent on default port
4. volleyCli     # run a mini load test
5. volleyCli -h  # to see options

Running with multiple agents (but not etcd)
--
1. pkill -f volleyAgent # make sure none are running
2. volleyAgent -port 9875 &
3. volleyAgent -port 9876 &
4. volleyCli -agents localhost:9875,localhost:9876

Running with multiple agents registered in a local etcd
--
1. pkill -f volleyAgent  # make sure none are running
2. volleyAgent -port 9875 -etcd &
3. volleyAgent -port 9876 -etcd &
4. volleyAgent -port 9877 -etcd &
5. volleyCli -agentCount 2 -etcd  # picks 2 out of the 3 agents registered with etcd





