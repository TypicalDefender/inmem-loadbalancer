In-Memory Experimental Load Balancer
===

⚡️ Blazing Fast, Super Simple, Single Node load balancer written in Go


```
go run .
```

### Tasklist

 [x] Basic working LB with pre-configured backend
 [ ] Benchmark performance against NGINX and other LBs
 [x] Ability to add register a new backend on the fly
 [ ] `stats` command to output stats per backend server. Stats could include number of requests, average time taken, etc.
 [x] Pluggable algorithm for load balancing: Round Robin, Consistent Hasshing, etc.
 [ ] Export LB metrics and visualize it through Grafana.
 [ ] Connection Pooling
 [ ] If protocol/scheme specific then identify response status/codes and dump it into metrics
 [ ] Ability to healthcheck and stop/start routing traffic to heaalthy ones
 [x] Consistent Hashing for load balancing
 [x] Unique ID to request
 [x] Simulate Balancing Strategy by giving request id
 [x] Topology per strategy
 [ ] Handle backend becoming unhealthy for all balancing strategy
