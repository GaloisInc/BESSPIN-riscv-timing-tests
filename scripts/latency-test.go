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

var BOOM_DIR = bagpipe.HomeDirectory() + "/src/boom-template/verisim"
var BOOM_BIN = "./simulator-boom.system-BoomConfig"

var ROCKET_DIR = bagpipe.HomeDirectory() + "/src/rocket-chip/emulator"
var ROCKET_BIN = "./emulator-freechips.rocketchip.system-DefaultConfig"

type input_t struct {
	Binary_path  string
	Emulator_dir string
	Emulator_bin string
}

type output_t struct {
	Instr_count string
	Cycle_count string
}

func exec(emulator_dir string, emulator_bin string, exe_file string) string {
	if bagpipe.FileExists(exe_file) == false {
		panic("could not find " + exe_file)
	}

	cmd := emulator_bin + " -s 0 " + exe_file
	return bagpipe.ExecCommand(cmd, emulator_dir)
}

func __exec_one(input input_t) output_t {
	exec_output := exec(input.Emulator_dir, input.Emulator_bin,
		input.Binary_path)

	instrs, cycles := parse(exec_output)
	return output_t{Instr_count: instrs, Cycle_count: cycles}
}

func exec_one(input bytes.Buffer) bytes.Buffer {
	__input := decode_input(input)
	__output := __exec_one(__input)
	return encode_output(__output)
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

func kernel(noop_count int, repetitions int) (uint64, uint64, uint64, uint64) {
	exe_file := bagpipe.CreateTempFile("latency-test.")

	cmd := fmt.Sprintf("riscv64-unknown-elf-gcc -Ienv -Icommon "+
		"-mcmodel=medany -static -std=gnu99 -O2 -ffast-math -fno-common "+
		"-fno-builtin-printf -DNOOP_COUNT=%d -o %s latency-driver.c "+
		"common/syscalls.c common/crt.S -static -nostdlib -nostartfiles -lm "+
		"-lgcc -T common/test.ld", noop_count, exe_file)

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())

	max_instrs := uint64(0)
	max_cycles := uint64(0)

	min_instrs := uint64(math.MaxUint64)
	min_cycles := uint64(math.MaxUint64)

	sprinter := bagpipe.NewSprinter(exec_one, MAX_THREADS, repetitions)

	for idx := 0; idx < repetitions; idx += 1 {
		input := input_t{Binary_path: exe_file, Emulator_dir: ROCKET_DIR,
			Emulator_bin: ROCKET_BIN}

		sprinter.FeedWorker(input)
	}

	for idx := 0; idx < sprinter.ResultCount(); idx += 1 {
		_, __output := sprinter.ReadResult()
		output := decode_output(__output)

		instrs, err := strconv.ParseUint(output.Instr_count, 16, 64)
		bagpipe.CheckError(err)

		cycles, err := strconv.ParseUint(output.Cycle_count, 16, 64)
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

	bagpipe.DeleteFile(exe_file)
	return min_instrs, max_instrs, min_cycles, max_cycles
}

func main() {
	for noop_count := 0; noop_count < 1024; noop_count += 1 {
		min_instrs, max_instrs, min_cycles, max_cycles := kernel(noop_count, 4)

		fmt.Printf("noops: %4d\tinstrs: [ %d - %d ]\tcycles: [ %d - %d ]\n",
			noop_count, min_instrs, max_instrs, min_cycles, max_cycles)
	}
}
