package main

import (
	"fmt"
	"gitlab.com/ashay/bagpipe"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type point_t struct {
	left_operand  uint64
	right_operand uint64
	instr_count   uint64
	cycle_count   uint64
}

type value_range_t struct {
	start_value uint64
	end_value   uint64
	latency     uint64
}

func parse_lines(lines []string) []point_t {
	var points []point_t

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		fields := strings.Split(line, " ")

		field1, err := strconv.ParseUint(fields[0], 16, 64)
		bagpipe.CheckError(err)

		field2, err := strconv.ParseUint(fields[1], 16, 64)
		bagpipe.CheckError(err)

		field3, err := strconv.ParseUint(fields[2], 10, 64)
		bagpipe.CheckError(err)

		field4, err := strconv.ParseUint(fields[3], 10, 64)
		bagpipe.CheckError(err)

		points = append(points, point_t{field1, field2, field3, field4})
	}

	return points
}

func trim_points(points []point_t) []point_t {
	var new_points []point_t
	new_points = append(new_points, points[0])

	for idx := 1; idx < len(points); idx += 1 {
		prev_point := points[idx-1]
		curr_point := points[idx]

		if curr_point.cycle_count != prev_point.cycle_count {
			new_points = append(new_points, curr_point)
		}
	}

	return new_points
}

func perpendicular_distance(start_x uint64, start_y uint64, end_x uint64,
	end_y uint64, point_x uint64, point_y uint64) float64 {

	diff_y := end_y - start_y
	diff_x := end_x - start_x
	diff := diff_y*point_x - diff_x*point_y
	numerator := math.Abs(float64(diff + end_x*start_y - start_x*end_y))

	sq_diff_y := (end_y - start_y) * (end_y - start_y)
	sq_diff_x := (end_x - start_x) * (end_x - start_x)
	denominator := math.Sqrt(float64(sq_diff_y + sq_diff_x))

	return numerator / denominator
}

func refine_points(points []point_t) []point_t {
	start_point := points[0]
	end_point := points[len(points)-1]

	start_x := start_point.left_operand
	start_y := start_point.cycle_count

	end_x := end_point.left_operand
	end_y := end_point.cycle_count

	farthest_point := 0
	max_distance := 0.0

	const k_threshold_distance = 0.01

	for idx := 1; idx < len(points)-1; idx += 1 {
		point := points[idx]
		point_x := point.left_operand
		point_y := point.cycle_count

		point_distance := perpendicular_distance(start_x, start_y, end_x,
			end_y, point_x, point_y)

		if point_distance > max_distance {
			farthest_point = idx
			max_distance = point_distance
		}
	}

	if max_distance <= k_threshold_distance {
		return []point_t{start_point, end_point}
	}

	// refine everything from 0 to idx (inclusive).
	left_points := refine_points(points[:farthest_point+1])

	// refine everything from idx to len-1 (inclusive).
	right_points := refine_points(points[farthest_point:])

	// don't include the farthest point twice.
	return append(left_points, right_points[1:]...)
}

func infer_single_dimension_ranges(filename string) []value_range_t {
	contents := bagpipe.ReadFile(filename)
	lines := strings.Split(contents, "\n")
	points := parse_lines(lines)

	sort.Slice(points, func(i, j int) bool {
		if points[i].cycle_count < points[j].cycle_count {
			return true
		}

		if points[i].cycle_count == points[j].cycle_count &&
			points[i].left_operand < points[j].left_operand {
			return true
		}

		if points[i].cycle_count == points[j].cycle_count &&
			points[i].left_operand == points[j].left_operand &&
			points[i].right_operand < points[j].right_operand {
			return true
		}

		return false
	})

	trimmed_points := trim_points(points)
	refined_points := refine_points(trimmed_points)

	var value_range value_range_t
	var value_ranges []value_range_t

	value_range.latency = refined_points[0].cycle_count
	value_range.start_value = refined_points[0].left_operand

	for idx := 1; idx < len(refined_points); idx += 1 {
		value_range.end_value = refined_points[idx].left_operand - 1
		value_ranges = append(value_ranges, value_range)

		value_range.latency = refined_points[idx].cycle_count
		value_range.start_value = refined_points[idx].left_operand
	}

	value_range.end_value = 0xffffffffffffffff
	value_ranges = append(value_ranges, value_range)

	return value_ranges
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatal("need file containing results from randomization.")
	}

	fmt.Printf("%-16s - %-16s  ->  %s\n", "[start value]", "[end value]",
		"[latency]")

	for _, value_range := range infer_single_dimension_ranges(args[0]) {
		fmt.Printf("%016x - %016x  -> %5d cycle(s)\n", value_range.start_value,
			value_range.end_value, value_range.latency)
	}
}
