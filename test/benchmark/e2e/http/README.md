# E2E Benchmark to compare HTTP binding & HTTP transport

This benchmark aims to provide a comparison between package `github.com/cloudevents/sdk-go/pkg/bindings/http` and `github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http`

## Metrics

Test keys are:

* Parallelism (Configuration of `GOMAXPROCS` value, from 1 to `runtime.NumCPU()`)
* Payload size in kb

## Run and visualize results

### Build

```shell script
go build -o main go
```

### Run all tests

```shell script
./main --bench=baseline-binary
./main --bench=binding-binary-to-binary 
./main --bench=client-binary
```

### Plot results

An example plot script is provided to plot parallelism - nanoseconds/ops, given the payload size:

```shell script
gnuplot -c plot_parallelism_ns.gnuplot <payload_size_kb>
```

Example:

```shell script
gnuplot -c plot_parallelism_ns.gnuplot 16
```

#### Plotter

To view some results without gnuplot (from the plotter directory),

```shell script
../main --bench=baseline-binary && go run plot.go baseline-binary.cvs
```

