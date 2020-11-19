# Introduce
GoMark is a tool for system status monitoring and it implements some variables exactly like bvar in brpc, but using go instead of c++.

For more detailed information, see:
https://github.com/apache/incubator-brpc/blob/master/docs/en/bvar.md



# Usage
## Init
gomark monitor variable statistics through HTTP web server, so you need call
```go
StartHTTPServer(port int)
```
to start HTTP service before using.

## Variable Create
A variable is an entity that maintains all information of statistics. There are several variables, and can be created by:
```
NewLatencyRecorder
NewAdder
NewCounter
NewQPS
NewMaxer
NewWindowMaxer
```

Each of above will create a variable and returns as `gmi.Marker`, which is an interface:
```go
// Marker is an interface to provide variable marking.
type Marker interface {
	// Mark a number, the number definition is bound to marker itself.
	Mark(n int32)
	// Stop this marking.
	Cancel()
}
```

Call `Mark` to send a marking point to variable and `Cancel` to stop using(and never use it).

# Monitor

Visit http://ip:port/vars to monitor system statistics.

# Performance
Test method: in one goroutine, continuouesly marking for 10 million marks, get the time elasped and
calculate how many marks can be done in one second(QPS).


updated 2020.11.19

| Variable        | QPS     |
| --              | --      |
| MaxWindow       | 6214338 |
| Maxer           | 6497730 |
| Adder           | 6445707 |
| LatencyRecorder | 5411037 |
| QPS             | 6837910 |
| Counter         | 6163937 |


# Test
`cmd/main/gomark.go` is used run tests. It use glog for logging, so add ` -stderrthreshold=INFO` in command line:
```
go run gomark.go -stderrthreshold=INFO 
```
Read the usage and run test.
