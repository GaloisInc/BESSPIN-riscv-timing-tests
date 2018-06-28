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

type record_t struct {
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

type point_t struct {
	x float64
	y float64
}

type line_t struct {
	start point_t
	end   point_t
}

func parse_lines(lines []string) []record_t {
	var records []record_t

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

		records = append(records, record_t{field1, field2, field3, field4})
	}

	return records
}

func remove_duplicates(records []record_t) []record_t {
	var new_records []record_t
	new_records = append(new_records, records[0])

	for idx := 1; idx < len(records); idx += 1 {
		prev_record := records[idx-1]
		curr_record := records[idx]

		if curr_record.left_operand != prev_record.left_operand ||
			curr_record.right_operand != prev_record.right_operand {
			new_records = append(new_records, curr_record)
		}
	}

	return new_records
}

func point_distance(start point_t, end point_t) float64 {
	diff := point_t{start.x - end.x, start.y - end.y}
	return math.Sqrt(diff.x*diff.x + diff.y*diff.y)
}

func line_distance(line line_t, point point_t) float64 {
	end := line.end
	start := line.start

	line_angle := math.Atan2(start.y-end.y, start.x-end.x)
	if line_angle > math.Pi/2 {
		line_angle = math.Pi - line_angle
	}

	point_angle := math.Atan2(start.y-point.y, start.x-point.x)
	if point_angle > math.Pi/2 {
		point_angle = math.Pi - point_angle
	}

	return math.Abs(line_angle-point_angle) * point_distance(start, point)
}

func find_inflection_records(records []record_t) []record_t {
	last_idx := len(records) - 1

	if last_idx == 0 {
		return records
	}

	start_x := float64(records[0].left_operand)
	start_y := float64(records[0].cycle_count)
	start_point := point_t{start_x, start_y}

	end_x := float64(records[last_idx].left_operand)
	end_y := float64(records[last_idx].cycle_count)
	end_point := point_t{end_x, end_y}

	line := line_t{start_point, end_point}

	max_distance := 0.0
	farthest_record := 0

	for idx := 1; idx < len(records)-1; idx += 1 {
		point_x := float64(records[idx].left_operand)
		point_y := float64(records[idx].cycle_count)
		point_distance := line_distance(line, point_t{point_x, point_y})

		if point_distance > max_distance {
			farthest_record = idx
			max_distance = point_distance
		}
	}

	const k_threshold_distance = 5

	if max_distance <= k_threshold_distance {
		return []record_t{records[0], records[last_idx]}
	}

	// refine everything from 0 to idx (inclusive).
	left_records := find_inflection_records(records[:farthest_record+1])

	// refine everything from idx to len-1 (inclusive).
	right_records := find_inflection_records(records[farthest_record:])

	// don't include the farthest record twice.
	return append(left_records, right_records[1:]...)
}

func read_records(filename string) []record_t {
	contents := bagpipe.ReadFile(filename)
	lines := strings.Split(contents, "\n")
	records := parse_lines(lines)

	sort.Slice(records, func(i, j int) bool {
		if records[i].cycle_count < records[j].cycle_count {
			return true
		}

		if records[i].cycle_count == records[j].cycle_count &&
			records[i].left_operand < records[j].left_operand {
			return true
		}

		if records[i].cycle_count == records[j].cycle_count &&
			records[i].left_operand == records[j].left_operand &&
			records[i].right_operand < records[j].right_operand {
			return true
		}

		return false
	})

	return records
}

func infer_single_dimension_ranges(filename string) []value_range_t {
	records := read_records(filename)
	unique_records := remove_duplicates(records)
	inflection_records := find_inflection_records(unique_records)

	var value_ranges []value_range_t
	var value_range value_range_t

	value_range.start_value = 0
	value_range.latency = inflection_records[0].cycle_count

	for idx := 0; idx < len(inflection_records); idx += 1 {
		if inflection_records[idx].cycle_count != value_range.latency {
			value_range.end_value = inflection_records[idx].left_operand - 1
			value_ranges = append(value_ranges, value_range)

			value_range.latency = inflection_records[idx].cycle_count
			value_range.start_value = inflection_records[idx].left_operand
		}
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
