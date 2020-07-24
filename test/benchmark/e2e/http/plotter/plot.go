package main

import (
	"bufio"
	"flag"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {

	flag.Parse()

	filename := flag.Arg(0)

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	plots := make(map[string]plotter.XYs, 0)

	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		}

		cvs := strings.Split(string(line), ",")

		parallelism, _ := strconv.Atoi(cvs[0])
		payloadSize, _ := strconv.Atoi(cvs[1])
		outputSenders, _ := strconv.Atoi(cvs[2])
		nsPerOp, _ := strconv.Atoi(cvs[3])
		allocedBytesPerOp, _ := strconv.Atoi(cvs[4])

		//fmt.Printf("parallelism %d \n", parallelism)
		//fmt.Printf("payloadSize %d \n", payloadSize)
		//fmt.Printf("outputSenders %d \n", outputSenders)
		//fmt.Printf("nsPerOp %d \n", nsPerOp)
		//fmt.Printf("allocedBytesPerOp %d \n\n", allocedBytesPerOp)
		{
			key := fmt.Sprintf("ns/op %d [%d]", payloadSize, outputSenders)
			if _, found := plots[key]; !found {
				plots[key] = make(plotter.XYs, 0)
			}
			plots[key] = append(plots[key], plotter.XY{
				X: float64(parallelism),
				Y: float64(nsPerOp),
			})
		}
		{
			key := fmt.Sprintf("allocs/op %d [%d]", payloadSize, outputSenders)
			if _, found := plots[key]; !found {
				plots[key] = make(plotter.XYs, 0)
			}
			plots[key] = append(plots[key], plotter.XY{
				X: float64(parallelism),
				Y: float64(allocedBytesPerOp),
			})
		}

	}

	{
		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = "Nanoseconds Per Op"
		p.X.Label.Text = "Parallelism"
		p.Y.Label.Text = "Nanoseconds"

		lines := make([]interface{}, 0)
		for k, v := range plots {
			if strings.HasPrefix(k, "ns/op") {
				lines = append(lines, k, v)
			}
		}

		if err = plotutil.AddLinePoints(p, lines...); err != nil {
			panic(err)
		}

		// Save the plot to a PNG file.
		if err := p.Save(4*vg.Inch, 4*vg.Inch, "ns-op.png"); err != nil {
			panic(err)
		}
	}
	{
		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = "Allocations in Bytes Per Op"
		p.X.Label.Text = "Parallelism"
		p.Y.Label.Text = "Count"

		lines := make([]interface{}, 0)
		for k, v := range plots {
			if strings.HasPrefix(k, "allocs/op") {
				lines = append(lines, k, v)
			}
		}

		if err = plotutil.AddLinePoints(p, lines...); err != nil {
			panic(err)
		}

		// Save the plot to a PNG file.
		if err := p.Save(4*vg.Inch, 4*vg.Inch, "allocs-op.png"); err != nil {
			panic(err)
		}
	}
}
