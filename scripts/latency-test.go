package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"gitlab.com/ashay/bagpipe"
)

var MAX_THREADS = 4

var BOOM_DIR = "bin/"
var BOOM_BIN = "./simulator-boom.system-BoomConfig"

var ROCKET_DIR = "bin/"
var ROCKET_BIN = "./emulator-galois.system-P2Config"

type input_t struct {
	Noop_count   string
	Emulator_dir string
	Emulator_bin string
}

type output_t struct {
	Min_instr_count string
	Max_instr_count string
	Min_cycle_count string
	Max_cycle_count string
}

func encode_input(value input_t) bytes.Buffer {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(value)
	bagpipe.CheckError(err)

	return buffer
}

func encode_output(value output_t) bytes.Buffer {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(value)
	bagpipe.CheckError(err)

	return buffer
}

func decode_input(input bytes.Buffer) input_t {
	decoder := gob.NewDecoder(&input)

	var value input_t
	err := decoder.Decode(&value)
	bagpipe.CheckError(err)

	return value
}

func decode_output(output bytes.Buffer) output_t {
	decoder := gob.NewDecoder(&output)

	var value output_t
	err := decoder.Decode(&value)
	bagpipe.CheckError(err)

	return value
}

func parse(line string) (string, string) {
	pattern := "instrs[\\s]*([0-9a-f]*)[\\s]*cycles[\\s]*([0-9a-f]*)"
	regex := regexp.MustCompile(pattern)

	if regex.MatchString(line) == false {
		panic("could not parse output: \"" + line + "\"")
	}

	match := regex.FindStringSubmatch(line)
	i_instr_count, err := strconv.ParseUint(match[1], 16, 64)
	bagpipe.CheckError(err)

	i_cycle_count, err := strconv.ParseUint(match[2], 16, 64)
	bagpipe.CheckError(err)

	instr_count := strconv.FormatUint(i_instr_count, 16)
	cycle_count := strconv.FormatUint(i_cycle_count, 16)

	return instr_count, cycle_count
}

func exec(emulator_dir string, emulator_bin string, exe_file string) string {
	if bagpipe.FileExists(exe_file) == false {
		panic("could not find " + exe_file)
	}

	cmd := emulator_bin + " -s 0 " + exe_file
	return bagpipe.ExecCommand(cmd, emulator_dir)
}

func kernel(noop_count uint64, em_dir string, em_bin string) (uint64, uint64, uint64, uint64) {
	exe_file := bagpipe.CreateTempFile("latency-test.")

	cmd := fmt.Sprintf("riscv64-unknown-elf-gcc -Ienv -Icommon "+
		"-mcmodel=medany -static -std=gnu99 -O2 -ffast-math -fno-common "+
		"-fno-builtin-printf -DNOOP_COUNT=%d -o %s latency-driver.c "+
		"common/syscalls.c common/crt.S -static -nostdlib -nostartfiles -lm "+
		"-lgcc -T common/test.ld", noop_count, exe_file)

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory() + "/src")

	max_instrs := uint64(0)
	max_cycles := uint64(0)

	min_instrs := uint64(math.MaxUint64)
	min_cycles := uint64(math.MaxUint64)

	for idx := 0; idx < 3; idx += 1 {
		__instrs, __cycles := parse(exec(em_dir, em_bin, exe_file))

		instrs, err := strconv.ParseUint(__instrs, 16, 64)
		bagpipe.CheckError(err)

		cycles, err := strconv.ParseUint(__cycles, 16, 64)
		bagpipe.CheckError(err)

		if instrs < min_instrs {
			min_instrs = instrs
		}

		if instrs > max_instrs {
			max_instrs = instrs
		}

		if cycles < min_cycles {
			min_cycles = cycles
		}

		if cycles > max_cycles {
			max_cycles = cycles
		}
	}

	if bagpipe.FileExists(exe_file) {
		bagpipe.DeleteFile(exe_file)
	}

	return min_instrs, max_instrs, min_cycles, max_cycles
}

func __exec_one(input input_t) output_t {
	noop_count, err := strconv.ParseUint(input.Noop_count, 16, 64)
	bagpipe.CheckError(err)

	min_instrs, max_instrs, min_cycles, max_cycles := kernel(noop_count,
		input.Emulator_dir, input.Emulator_bin)

	s_min_instrs := strconv.FormatUint(min_instrs, 16)
	s_min_cycles := strconv.FormatUint(min_cycles, 16)
	s_max_instrs := strconv.FormatUint(max_instrs, 16)
	s_max_cycles := strconv.FormatUint(max_cycles, 16)

	return output_t{Min_instr_count: s_min_instrs,
		Max_instr_count: s_max_instrs, Min_cycle_count: s_min_cycles,
		Max_cycle_count: s_max_cycles}
}

func exec_one(input bytes.Buffer) bytes.Buffer {
	__input := decode_input(input)
	__output := __exec_one(__input)
	return encode_output(__output)
}

func main() {
	max_noops := 12
	const k_output_file = "latency-results.txt"

	if bagpipe.FileExists(k_output_file) {
		bagpipe.DeleteFile(k_output_file)
	}

	sprinter := bagpipe.NewSprinter(exec_one, MAX_THREADS, max_noops)

	for noop_count := uint64(0); noop_count < uint64(max_noops); noop_count += 1 {
		s_noop_count := strconv.FormatUint(noop_count, 16)

		input := input_t{Noop_count: s_noop_count, Emulator_dir: ROCKET_DIR,
			Emulator_bin: ROCKET_BIN}

		status := fmt.Sprintf("%4d of %4d", noop_count+1, max_noops)
		bagpipe.UpdateStatus("testing [" + status + " ] ... ")

		sprinter.FeedWorker(input)
	}

	for idx := 0; idx < sprinter.ResultCount(); idx += 1 {
		__input, __output := sprinter.ReadResult()

		input := decode_input(__input)
		output := decode_output(__output)

		noop_count, err := strconv.ParseUint(input.Noop_count, 16, 64)
		bagpipe.CheckError(err)

		min_instrs, err := strconv.ParseUint(output.Min_instr_count, 16, 64)
		bagpipe.CheckError(err)

		max_instrs, err := strconv.ParseUint(output.Max_instr_count, 16, 64)
		bagpipe.CheckError(err)

		min_cycles, err := strconv.ParseUint(output.Min_cycle_count, 16, 64)
		bagpipe.CheckError(err)

		max_cycles, err := strconv.ParseUint(output.Max_cycle_count, 16, 64)
		bagpipe.CheckError(err)

		log_line := fmt.Sprintf("noops: %4d\tinstrs: [ %d - %d ]\tcycles: "+
			"[ %d - %d ]\n", noop_count, min_instrs, max_instrs, min_cycles,
			max_cycles)
		bagpipe.AppendFile(k_output_file, log_line)
	}

	bagpipe.UpdateStatus("finished, results in latency-results.txt\n")
}
