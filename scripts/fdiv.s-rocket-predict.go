package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gitlab.com/ashay/bagpipe"
	"math"
)

func compute_latency(dividend uint64, divisor uint64) float64 {
	return 23
}

func sum(numbers []float64) (total float64) {
	for _, x := range numbers {
		total += x
	}
	return total
}

func stdev(numbers []float64, mean float64) float64 {
	total := 0.0
	for _, number := range numbers {
		total += math.Pow(number-mean, 2)
	}
	variance := total / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatal("incorrect number of arguments.")
	}

	contents := bagpipe.ReadFile(args[0])
	lines := strings.Split(contents, "\n")

	var differences []float64

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		fields := strings.Split(line, " ")
		if len(fields) != 4 {
			continue
		}

		dividend, err := strconv.ParseUint(fields[0], 16, 64)
		if err != nil {
			log.Fatal(err)
		}

		divisor, err := strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			log.Fatal(err)
		}

		__actual_time, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			log.Fatal(err)
		}

		actual_time := (__actual_time - 39) / 11

		predicted_time := compute_latency(dividend, divisor)
		difference := math.Abs(actual_time - predicted_time)
		differences = append(differences, difference)
	}

	avg := sum(differences) / float64(len(differences))
	stdev := stdev(differences, avg)

	fmt.Printf("error = %8.2f cycles   +/- %-8.2f   sample size = %v\n",
		avg, stdev, len(differences))
}
