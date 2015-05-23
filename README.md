# Go I/O scalability for many cores

## Environment

* 2* AWS EC2 c4.8xlarge (36 virtual cores)
* Use placement group for low latency and 10Gbit Ehternet
* Amazon Linux AMI (SR-IOV enabled)

```console
$ echo "options ixgbevf InterruptThrottleRate=10000" > /etc/modprobe.d/ixgbevf.conf
$ shutdown -r now  # reboot!

$ sudo -s
 echo 32768 > /proc/sys/net/core/rps_sock_flow_entries
 echo 32768 > /sys/class/net/eth0/queues/rx-0/rps_flow_cnt
 echo 32768 > /sys/class/net/eth0/queues/rx-1/rps_flow_cnt
 echo 'ffff' > /sys/class/net/eth0/queues/rx-0/rps_cpus
 echo 'ffff' > /sys/class/net/eth0/queues/rx-1/rps_cpus
```

```attack.sh
#!/bin/sh
URL=$1
set -x
set -e
curl -v ${URL}
wrk -t8  -c8   ${URL}
wrk -t16 -c16  ${URL}
wrk -t32 -c32  ${URL}
wrk -t36 -c128 ${URL}
wrk -t36 -c512 ${URL}
```

## nginx

config:

```nginx.conf
worker_processes 36;
daemon off;
events {
    worker_connections  1000000;
}

http {
    include     mime.types;
    access_log  off;
    sendfile    on;
    tcp_nopush  on;
    tcp_nodelay on;
    etag        off;

    server {
        listen 8080;
        location = /hello {
            echo "Hello, World\n";
        }
    }
}
```

result:

```console
$ ./attack.sh http://10.0.1.70:8080/hello
+ set -e
+ curl -v http://10.0.1.70:8080/hello
*   Trying 10.0.1.70...
* Connected to 10.0.1.70 (10.0.1.70) port 8080 (#0)
> GET /hello HTTP/1.1
> User-Agent: curl/7.40.0
> Host: 10.0.1.70:8080
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: openresty/1.7.4.1
< Date: Tue, 07 Apr 2015 08:34:28 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
<
Hello, World
* Connection #0 to host 10.0.1.70 left intact
+ wrk -t8 -c8 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   193.32us   51.20us   1.57ms   83.29%
    Req/Sec     5.11k   299.35     6.84k    78.34%
  410759 requests in 10.10s, 72.06MB read
Requests/sec:  40669.44
Transfer/sec:      7.13MB
+ wrk -t16 -c16 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   193.83us   49.02us   2.58ms   81.92%
    Req/Sec     5.08k   323.33     6.82k    79.95%
  817566 requests in 10.10s, 143.42MB read
Requests/sec:  80949.48
Transfer/sec:     14.20MB
+ wrk -t32 -c32 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   207.68us   43.24us   1.47ms   84.49%
    Req/Sec     4.74k    95.83     5.16k    78.37%
  1523527 requests in 10.10s, 267.27MB read
Requests/sec: 150847.59
Transfer/sec:     26.46MB
+ wrk -t36 -c128 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   275.15us  107.67us   9.26ms   90.40%
    Req/Sec    10.58k   386.17    11.36k    88.01%
  3828830 requests in 10.10s, 671.69MB read
Requests/sec: 379107.47
Transfer/sec:     66.51MB
+ wrk -t36 -c512 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.32ms    9.24ms 204.54ms   99.31%
    Req/Sec    19.80k     4.82k   38.40k    86.31%
  7054604 requests in 10.10s, 1.21GB read
Requests/sec: 698483.94
Transfer/sec:    122.53MB
```

## Go 1.4.2 Basic

GOMAXPROCS=36:

```console
$ ./attack.sh http://10.0.1.70:8080/json
+ set -e
+ curl -v http://10.0.1.70:8080/json
*   Trying 10.0.1.70...
* Connected to 10.0.1.70 (10.0.1.70) port 8080 (#0)
> GET /json HTTP/1.1
> User-Agent: curl/7.40.0
> Host: 10.0.1.70:8080
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Tue, 07 Apr 2015 08:37:14 GMT
< Content-Length: 28
<
{"message":"Hello, World!"}
* Connection #0 to host 10.0.1.70 left intact
+ wrk -t8 -c8 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   242.89us  112.94us   1.22ms   89.80%
    Req/Sec     4.28k    75.15     4.46k    72.52%
  343969 requests in 10.10s, 44.61MB read
Requests/sec:  34056.43
Transfer/sec:      4.42MB
+ wrk -t16 -c16 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   277.25us  174.65us   5.98ms   87.73%
    Req/Sec     3.87k    76.26     4.09k    68.38%
  622075 requests in 10.10s, 80.68MB read
Requests/sec:  61594.63
Transfer/sec:      7.99MB
+ wrk -t32 -c32 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   346.39us  226.63us   1.64ms   84.55%
    Req/Sec     3.15k    82.64     3.50k    69.39%
  1010541 requests in 10.10s, 131.07MB read
Requests/sec: 100059.58
Transfer/sec:     12.98MB
+ wrk -t36 -c128 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   529.05us  328.25us   6.17ms   83.28%
    Req/Sec     5.84k   237.27     7.72k    81.71%
  2105500 requests in 10.10s, 273.08MB read
Requests/sec: 208455.58
Transfer/sec:     27.04MB
+ wrk -t36 -c512 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.71ms    9.32ms 207.11ms   99.39%
    Req/Sec    13.21k     2.35k   31.67k    89.80%
  4703510 requests in 10.10s, 610.04MB read
Requests/sec: 465698.02
Transfer/sec:     60.40MB
```

## Go 1.4.2 w/o GC

When `GOGC=off`, I run test one by one, to avoid eat all RAM on machine.

server:

```console
$ GOGC=off GOMAXPROCS=36 ./hello
```

attacker:

```
$ wrk -t8 -c8 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   205.89us   30.86us   2.42ms   89.75%
    Req/Sec     4.84k   144.84     6.07k    86.39%
  388736 requests in 10.10s, 50.42MB read
Requests/sec:  38488.59
Transfer/sec:      4.99MB

$ wrk -t16 -c16 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   210.93us   46.41us   4.23ms   90.24%
    Req/Sec     4.72k   232.51     4.96k    85.46%
  759754 requests in 10.10s, 98.54MB read
Requests/sec:  75222.67
Transfer/sec:      9.76MB

$ wrk -t32 -c32 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   240.22us   80.49us   5.33ms   89.63%
    Req/Sec     4.17k   349.29     4.80k    65.38%
  1339810 requests in 10.10s, 173.77MB read
Requests/sec: 132655.10
Transfer/sec:     17.21MB

$ wrk -t36 -c128 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   417.66us  387.97us  43.04ms   98.66%
    Req/Sec     7.39k   307.37     8.22k    81.78%
  2670129 requests in 10.10s, 346.31MB read
Requests/sec: 264377.30
Transfer/sec:     34.29MB

$ wrk -t36 -c512 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.50ms    9.14ms 206.18ms   99.18%
    Req/Sec    16.29k     3.17k   32.78k    89.57%
  5780745 requests in 10.10s, 749.76MB read
Requests/sec: 572360.91
Transfer/sec:     74.24MB
```

## Go 1.4.2 Prefork

```
$ GOMAXPROCS=2 ./hello -prefork=18
```

```console
$ ./attack.sh http://10.0.1.70:8080/json
+ set -e
+ curl -v http://10.0.1.70:8080/json
*   Trying 10.0.1.70...
* Connected to 10.0.1.70 (10.0.1.70) port 8080 (#0)
> GET /json HTTP/1.1
> User-Agent: curl/7.40.0
> Host: 10.0.1.70:8080
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Tue, 07 Apr 2015 08:47:37 GMT
< Content-Length: 28
<
{"message":"Hello, World!"}
* Connection #0 to host 10.0.1.70 left intact
+ wrk -t8 -c8 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   216.42us   61.30us   0.96ms   85.89%
    Req/Sec     4.64k   142.94     5.61k    81.81%
  373272 requests in 10.10s, 48.41MB read
Requests/sec:  36957.93
Transfer/sec:      4.79MB
+ wrk -t16 -c16 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   227.62us   65.86us   1.56ms   82.51%
    Req/Sec     4.41k   277.19     4.83k    69.86%
  709051 requests in 10.10s, 91.96MB read
Requests/sec:  70205.38
Transfer/sec:      9.11MB
+ wrk -t32 -c32 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   318.50us  139.26us   8.83ms   88.18%
    Req/Sec     3.18k   249.68     4.34k    65.90%
  1021312 requests in 10.10s, 132.46MB read
Requests/sec: 101120.30
Transfer/sec:     13.12MB
+ wrk -t36 -c128 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   698.10us    4.66ms 198.31ms   99.64%
    Req/Sec     5.89k   457.05     7.62k    80.03%
  2122893 requests in 10.10s, 275.34MB read
Requests/sec: 210191.17
Transfer/sec:     27.26MB
+ wrk -t36 -c512 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.87ms   11.77ms 204.88ms   98.98%
    Req/Sec    17.24k     3.26k   34.95k    89.48%
  5948459 requests in 10.10s, 771.51MB read
Requests/sec: 588988.82
Transfer/sec:     76.39MB
```

## Go 1.5 

Git commit: b40421f32c37064f5eb9b00f4f5aebe7243be6cd

I found `rngd` eats some CPU.

`$ GOMAXPROCS=36 ./hello1.5`:

```console
./attack.sh http://10.0.1.70:8080/json
+ set -e
+ curl -v http://10.0.1.70:8080/json
*   Trying 10.0.1.70...
* Connected to 10.0.1.70 (10.0.1.70) port 8080 (#0)
> GET /json HTTP/1.1
> User-Agent: curl/7.40.0
> Host: 10.0.1.70:8080
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Tue, 07 Apr 2015 08:50:12 GMT
< Content-Length: 28
<
{"message":"Hello, World!"}
* Connection #0 to host 10.0.1.70 left intact
+ wrk -t8 -c8 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   216.23us   88.12us   1.62ms   95.46%
    Req/Sec     4.74k    83.58     5.31k    82.53%
  380754 requests in 10.10s, 49.38MB read
Requests/sec:  37698.36
Transfer/sec:      4.89MB
+ wrk -t16 -c16 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   231.00us  128.55us   4.05ms   96.52%
    Req/Sec     4.57k   112.93     4.77k    72.14%
  733789 requests in 10.10s, 95.17MB read
Requests/sec:  72653.09
Transfer/sec:      9.42MB
+ wrk -t32 -c32 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   287.34us  209.98us   5.40ms   94.15%
    Req/Sec     3.85k   131.83     4.30k    67.50%
  1236730 requests in 10.10s, 160.40MB read
Requests/sec: 122445.54
Transfer/sec:     15.88MB
+ wrk -t36 -c128 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.33ms    5.95ms 110.46ms   98.59%
    Req/Sec     4.90k     2.29k    7.92k    55.66%
  1765885 requests in 10.10s, 229.03MB read
Requests/sec: 174838.71
Transfer/sec:     22.68MB
+ wrk -t36 -c512 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.61ms   16.18ms 373.71ms   98.81%
    Req/Sec     4.80k     4.29k   21.85k    74.34%
  1712196 requests in 10.10s, 222.07MB read
Requests/sec: 169531.78
Transfer/sec:     21.99MB
```

## Go 1.5 w/o GC

```
GOGC=off GOMAXPROCS=36 ./hello1.5
```

```
$ wrk -t36 -c256  http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 256 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   549.48us  735.69us  32.42ms   98.42%
    Req/Sec    13.64k   806.55    17.31k    88.97%
  4922988 requests in 10.10s, 638.51MB read
Requests/sec: 487451.24
Transfer/sec:     63.22MB
```

### trace

```
$ GODEBUG=gctrace=1 GOMAXPROCS=36 ./hello1.5
gc #1 @18.212s 1%: 0+2+0+11371+358 ms clock, 1+2+1+11371+717 ms cpu, 4->7736->7632 MB, 36 P
```


##  Go 1.4 scheduler trace

```
SCHED 2746ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=9 runqueue=53 [0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 2796ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=9 runqueue=24 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0]
SCHED 2847ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=9 runqueue=8 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 2898ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=1 idlethreads=8 runqueue=0 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 2948ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=9 runqueue=0 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 2999ms: gomaxprocs=32 idleprocs=3 threads=42 spinningthreads=1 idlethreads=12 runqueue=9 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3050ms: gomaxprocs=32 idleprocs=5 threads=42 spinningthreads=1 idlethreads=13 runqueue=2 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3100ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=40 runqueue=64 [0 0 0 0 1 0 0 0 0 0 1 0 0 0 0 0 0 1 1 1 0 1 0 0 0 0 0 1 1 0 0 0]
SCHED 3151ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=31 runqueue=51 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3202ms: gomaxprocs=32 idleprocs=0 threads=42 spinningthreads=0 idlethreads=18 runqueue=33 [0 0 0 0 0 1 0 0 0 0 1 1 0 0 0 1 1 0 0 0 1 0 1 0 0 1 0 0 0 0 0 0]
SCHED 3253ms: gomaxprocs=32 idleprocs=0 threads=48 spinningthreads=0 idlethreads=15 runqueue=5 [0 5 1 4 0 0 0 1 0 4 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3306ms: gomaxprocs=32 idleprocs=0 threads=48 spinningthreads=0 idlethreads=15 runqueue=160 [1 1 0 2 2 0 2 1 0 1 2 0 0 1 0 2 2 0 0 1 2 1 1 1 0 0 0 1 1 1 1 0]
SCHED 3357ms: gomaxprocs=32 idleprocs=8 threads=48 spinningthreads=2 idlethreads=22 runqueue=0 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3408ms: gomaxprocs=32 idleprocs=15 threads=48 spinningthreads=4 idlethreads=24 runqueue=2 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3459ms: gomaxprocs=32 idleprocs=19 threads=48 spinningthreads=1 idlethreads=33 runqueue=2 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3509ms: gomaxprocs=32 idleprocs=0 threads=48 spinningthreads=2 idlethreads=14 runqueue=20 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3560ms: gomaxprocs=32 idleprocs=0 threads=48 spinningthreads=0 idlethreads=15 runqueue=8 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
SCHED 3611ms: gomaxprocs=32 idleprocs=0 threads=48 spinningthreads=0 idlethreads=44 runqueue=16 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
```


## Low Latency Environment

```console
$ echo "options ixgbevf InterruptThrottleRate=3" > /etc/modprobe.d/ixgbevf.conf
$ shutdown -r now  # reboot!

$ sudo -s
 echo 32768 > /proc/sys/net/core/rps_sock_flow_entries
 echo 32768 > /sys/class/net/eth0/queues/rx-0/rps_flow_cnt
 echo 32768 > /sys/class/net/eth0/queues/rx-1/rps_flow_cnt
 echo 'ffff' > /sys/class/net/eth0/queues/rx-0/rps_cpus
 echo 'ffff' > /sys/class/net/eth0/queues/rx-1/rps_cpus
```

### nginx

```
$ ./attack.sh http://10.0.1.70:8080/hello
+ set -e
+ curl -v http://10.0.1.70:8080/hello
*   Trying 10.0.1.70...
* Connected to 10.0.1.70 (10.0.1.70) port 8080 (#0)
> GET /hello HTTP/1.1
> User-Agent: curl/7.40.0
> Host: 10.0.1.70:8080
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: openresty/1.7.4.1
< Date: Tue, 07 Apr 2015 13:09:30 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
<
Hello, World
* Connection #0 to host 10.0.1.70 left intact
+ wrk -t8 -c8 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   645.69us  180.83us   1.29ms   89.90%
    Req/Sec     1.53k   226.11     1.70k    88.54%
  122546 requests in 10.10s, 21.50MB read
Requests/sec:  12132.85
Transfer/sec:      2.13MB
+ wrk -t16 -c16 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   682.41us  222.11us   1.35ms   83.71%
    Req/Sec     1.45k   190.45     1.70k    62.57%
  231697 requests in 10.10s, 40.65MB read
Requests/sec:  22940.22
Transfer/sec:      4.02MB
+ wrk -t32 -c32 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.92ms  291.01us   3.60ms   53.47%
    Req/Sec     1.07k   231.42     1.70k    73.31%
  342501 requests in 10.01s, 60.08MB read
Requests/sec:  34208.07
Transfer/sec:      6.00MB
+ wrk -t36 -c128 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.02ms  263.01us   4.05ms   71.10%
    Req/Sec     2.90k   314.53     4.30k    66.10%
  1044146 requests in 10.10s, 183.17MB read
Requests/sec: 103379.84
Transfer/sec:     18.14MB
+ wrk -t36 -c512 http://10.0.1.70:8080/hello
Running 10s test @ http://10.0.1.70:8080/hello
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.38ms    5.09ms 204.67ms   99.61%
    Req/Sec    11.91k     1.07k   12.92k    96.15%
  4308189 requests in 10.10s, 755.78MB read
Requests/sec: 426547.50
Transfer/sec:     74.83MB
```

### Go 1.4.2

```
$ ./attack.sh http://10.0.1.70:8080/json
+ set -e
+ curl -v http://10.0.1.70:8080/json
*   Trying 10.0.1.70...
* Connected to 10.0.1.70 (10.0.1.70) port 8080 (#0)
> GET /json HTTP/1.1
> User-Agent: curl/7.40.0
> Host: 10.0.1.70:8080
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Tue, 07 Apr 2015 13:11:52 GMT
< Content-Length: 28
<
{"message":"Hello, World!"}
* Connection #0 to host 10.0.1.70 left intact
+ wrk -t8 -c8 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  8 threads and 8 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.85ms  298.02us   1.96ms   57.86%
    Req/Sec     1.17k   325.56     1.71k    69.00%
  93432 requests in 10.02s, 12.12MB read
Requests/sec:   9328.06
Transfer/sec:      1.21MB
+ wrk -t16 -c16 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.86ms  302.92us   2.07ms   62.93%
    Req/Sec     1.17k   310.30     1.71k    48.31%
  185691 requests in 10.01s, 24.08MB read
Requests/sec:  18545.90
Transfer/sec:      2.41MB
+ wrk -t32 -c32 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  32 threads and 32 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   848.65us  355.67us   9.95ms   89.05%
    Req/Sec     1.19k   294.65     1.67k    49.69%
  378340 requests in 10.01s, 49.07MB read
Requests/sec:  37799.68
Transfer/sec:      4.90MB
+ wrk -t36 -c128 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 128 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.08ms  343.85us   3.11ms   71.14%
    Req/Sec     2.78k   375.02     4.07k    79.16%
  995751 requests in 10.10s, 129.15MB read
Requests/sec:  98590.56
Transfer/sec:     12.79MB
+ wrk -t36 -c512 http://10.0.1.70:8080/json
Running 10s test @ http://10.0.1.70:8080/json
  36 threads and 512 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.55ms    4.81ms 205.74ms   99.80%
    Req/Sec    10.01k   614.86    11.38k    91.56%
  3609680 requests in 10.10s, 468.17MB read
Requests/sec: 357388.26
Transfer/sec:     46.35MB
```

### Go 1.4.2 with patch

patch:

```
```

result:


