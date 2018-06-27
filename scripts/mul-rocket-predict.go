package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gitlab.com/ashay/bagpipe"
	"math"
	"math/bits"
)

const k_bit_range = 64

func compute_latency(multiplicand uint64) float64 {
	msb := k_bit_range - bits.LeadingZeros64(multiplicand)

	if msb <= 16 {
		return 2.27
	}

	if msb <= 24 {
		return 2.90
	}

	if msb <= 32 {
		return 3.81
	}

	if msb <= 40 {
		return 4.72
	}

	if msb <= 48 {
		return 5.63
	}

	return 6.54
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

		multiplicand, err := strconv.ParseUint(fields[0], 16, 64)
		if err != nil {
			log.Fatal(err)
		}

		__actual_time, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			log.Fatal(err)
		}

		actual_time := (__actual_time - 39) / 11

		predicted_time := compute_latency(multiplicand)
		difference := math.Abs(actual_time - predicted_time)
		differences = append(differences, difference)
	}

	avg := sum(differences) / float64(len(differences))
	stdev := stdev(differences, avg)

	fmt.Printf("error = %8.2f cycles   +/- %-8.2f   sample size = %v\n",
		avg, stdev, len(differences))
}
