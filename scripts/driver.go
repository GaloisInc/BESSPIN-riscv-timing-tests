package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"gitlab.com/ashay/bagpipe"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type dtype_t int

const (
	dtype_xxx dtype_t = 0
	dtype_int dtype_t = 1
	dtype_sp  dtype_t = 2
	dtype_dp  dtype_t = 3
	dtype_mem dtype_t = 4
)

var MAX_THREAD_COUNT = 4

var CC = "riscv64-unknown-elf-gcc"
var CFLAGS = "-I include -mcmodel=medany -std=gnu99 -O2"
var LDFLAGS = "-static -nostartfiles -T test.ld"

var BOOM_DIR = bagpipe.HomeDirectory() + "/src/boom-template/verisim"
var BOOM_BIN = "./simulator-boom.system-BoomConfig"

var ROCKET_DIR = bagpipe.HomeDirectory() + "/src/rocket-chip/emulator"
var ROCKET_BIN = "./emulator-freechips.rocketchip.system-DefaultConfig"

var int_objects = []string{
	"sll", "srl", "sra", "add", "sub", "xor", "and", "or", "slt", "sltu",
	"mul", "mulh", "mulhsu", "mulhu", "div", "divu", "rem", "remu",
}

var int_operands = []string{
	"0", "f", "ff", "fff", "ffff", "fffff", "ffffff", "fffffff",
	"ffffffff", "fffffffff", "ffffffffff", "fffffffffff", "ffffffffffff",
	"fffffffffffff", "ffffffffffffff", "fffffffffffffff",
	"7fffffffffffffff",
}

var sp_objects = []string{
	"fadd.s", "fsub.s", "fmul.s", "fdiv.s", "fsgnj.s", "fsgnjn.s",
	"fsgnjx.s", "fmin.s", "fmax.s",
}

var sp_operands = []string{
	"00000000", "40600000", "00084000", "7f800000", "ff800000", "7f800200",
}

var dp_objects = []string{
	"fadd.d", "fsub.d", "fmul.d", "fdiv.d", "fsgnj.d", "fsgnjn.d",
	"fsgnjx.d", "fmin.d", "fmax.d",
}

var dp_operands = []string{
	"0000000000000000", "4025000000000000", "0000000000000400",
	"7ff0000000000000", "fff0000000000000", "7ff0000000000001",
}

var mem_objects = []string{
	"lb", "sb", "lh", "sh", "lw", "sw", "flw", "fsw", "fld", "fsd",
}

var mem_operands = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

func get_dtype(instr string) dtype_t {
	if contains(sp_objects, instr) {
		return dtype_sp
	}

	if contains(dp_objects, instr) {
		return dtype_dp
	}

	if contains(int_objects, instr) {
		return dtype_int
	}

	if contains(mem_objects, instr) {
		return dtype_mem
	}

	return dtype_xxx
}

func link(bin string, op1 string, op2 string) string {
	var driver_file string

	dtype := get_dtype(bin)
	exe_file := bagpipe.CreateTempFile(bin + ".")

	if dtype == dtype_int {
		driver_file = "int-driver.c"
	} else if dtype == dtype_sp {
		driver_file = "sp-driver.c"
	} else if dtype == dtype_dp {
		driver_file = "dp-driver.c"
	} else if dtype == dtype_mem {
		driver_file = "mem-driver.c"
	} else {
		log.Fatal("invalid data type")
	}

	cmd := CC + " " + CFLAGS + " -DINST=" + bin + " -DOP1=0x" + op1 +
		" -DOP2=0x" + op2 + " crt.S syscalls.c " + driver_file + " -O2 " +
		LDFLAGS + " -o " + exe_file

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
	return exe_file
}

func clean(objects []string, dtype dtype_t) {
	bagpipe.UpdateStatus("cleaning ...")

	for _, object := range objects {
		if dtype == dtype_int {
			if bagpipe.FileExists(object + ".i.i") {
				bagpipe.DeleteFile(object + ".i.i")
			}
		} else if dtype == dtype_sp || dtype == dtype_dp {
			if bagpipe.FileExists(object + ".n.n") {
				bagpipe.DeleteFile(object + ".n.n")
			}

			if bagpipe.FileExists(object + ".n.s") {
				bagpipe.DeleteFile(object + ".n.s")
			}

			if bagpipe.FileExists(object + ".s.n") {
				bagpipe.DeleteFile(object + ".s.n")
			}

			if bagpipe.FileExists(object + ".s.s") {
				bagpipe.DeleteFile(object + ".s.s")
			}
		}
	}

	bagpipe.ClearStatus()
}

func contains(objects []string, object string) bool {
	for _, element := range objects {
		if element == object {
			return true
		}
	}

	return false
}

func exec(emulator_dir string, emulator_bin string, exe_file string) string {

	if bagpipe.FileExists(exe_file) == false {
		log.Fatal("could not find " + exe_file)
	}

	cmd := emulator_bin + " -s 0 -c " + exe_file
	return bagpipe.ExecCommand(cmd, emulator_dir)
}

func parse(line string) (string, string) {
	pattern := "instrs[\\s]*([0-9a-f]*)[\\s]*cycles[\\s]*([0-9a-f]*)"
	regex := regexp.MustCompile(pattern)

	if regex.MatchString(line) == false {
		log.Fatal("could not parse output: \"" + line + "\"")
	}

	match := regex.FindStringSubmatch(line)
	i_instr_count, err := strconv.ParseUint(match[1], 16, 64)
	bagpipe.CheckError(err)

	i_cycle_count, err := strconv.ParseUint(match[2], 16, 64)
	bagpipe.CheckError(err)

	f_cycle_count := float64(i_cycle_count) / float64(i_instr_count)

	instr_count := strconv.FormatUint(i_instr_count, 16)
	cycle_count := strconv.FormatFloat(f_cycle_count, 'f', 2, 64)

	return instr_count, cycle_count
}

func all_ones(bits int64) int64 {
	return (1 << uint64(bits)) - 1
}

func rand_int(min int64, max int64) int64 {
	return rand.Int63n(max-min) + min
}

func generate_subnormal_sp_operand() string {
	random_subnormal := rand_int(1, all_ones(23))
	return fmt.Sprintf("%016x", random_subnormal)
}

func generate_subnormal_dp_operand() string {
	random_subnormal := rand_int(1, all_ones(53))
	return fmt.Sprintf("%016x", random_subnormal)
}

func generate_normal_sp_operand() string {
	random_subnormal := rand_int(1, all_ones(23))

	exponent_mask := rand_int(1, all_ones(8)) << 23
	random_normal := random_subnormal | exponent_mask

	return fmt.Sprintf("%016x", random_normal)
}

func generate_normal_dp_operand() string {
	random_subnormal := rand_int(1, all_ones(53))

	exponent_mask := rand_int(1, all_ones(11)) << 52
	random_normal := random_subnormal | exponent_mask

	return fmt.Sprintf("%016x", random_normal)
}

func generate_operand(operand_type string, dtype dtype_t) string {
	if operand_type == "n" {
		if dtype == dtype_sp {
			return generate_normal_sp_operand()
		}

		if dtype == dtype_dp {
			return generate_normal_dp_operand()
		}

		log.Fatal("failed to recognize data type!")
	}

	if operand_type == "s" {
		if dtype == dtype_sp {
			return generate_subnormal_sp_operand()
		}

		if dtype == dtype_dp {
			return generate_subnormal_dp_operand()
		}

		log.Fatal("failed to recognize data type!")
	}

	log.Fatal("failed to recognize operand type!")
	return fmt.Sprintf("%016x", 0)
}

type input_t struct {
	Instr        string
	Left_op      string
	Right_op     string
	Emulator_dir string
	Emulator_bin string
	Aux_field    string
}

type output_t struct {
	Instr_count string
	Cycle_count string
}

func __exec_one(input input_t) output_t {
	exe_file := link(input.Instr, input.Left_op, input.Right_op)
	exec_output := exec(input.Emulator_dir, input.Emulator_bin, exe_file)

	bagpipe.DeleteFile(exe_file)
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

func get_emulator_dir(arch string) string {
	switch arch {
	case "rocket":
		return ROCKET_DIR

	case "boom":
		return BOOM_DIR
	}

	return "--unknown--"
}

func get_emulator_bin(arch string) string {
	switch arch {
	case "rocket":
		return ROCKET_BIN

	case "boom":
		return BOOM_BIN
	}

	return "--unknown--"
}

func sweep_instr_operands(arch string, op1 string, op2 string, instr string) {
	emulator_dir := get_emulator_dir(arch)
	emulator_bin := get_emulator_bin(arch)
	data_dir := bagpipe.WorkingDirectory() + "/../results/" + arch + "/data"

	log_filename := "out." + instr + "." + op1 + "." + op2
	if bagpipe.FileExists(data_dir + "/" + log_filename) {
		bagpipe.DeleteFile(data_dir + "/" + log_filename)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	operands := generate_int_operands()

	test_count := len(operands) * len(operands)
	sprinter := bagpipe.NewSprinter(exec_one, MAX_THREAD_COUNT, test_count)

	for idx1, operand1 := range operands {
		s_op1 := fmt.Sprintf("%016x", operand1)

		for idx2, operand2 := range operands {
			s_op2 := fmt.Sprintf("%016x", operand2)

			idx := idx1*len(operands) + idx2
			status := fmt.Sprintf("%4d of %4d", idx, test_count)

			bagpipe.UpdateStatus("testing " + instr + " [" + status +
				" ] ... ")

			input := input_t{Instr: instr, Left_op: s_op1, Right_op: s_op2,
				Emulator_dir: emulator_dir, Emulator_bin: emulator_bin}

			sprinter.FeedWorker(input)
		}
	}

	for idx := 0; idx < sprinter.ResultCount(); idx += 1 {
		__input, __output := sprinter.ReadResult()

		input := decode_input(__input)
		output := decode_output(__output)

		l_op, err := strconv.ParseUint(input.Left_op, 16, 64)
		bagpipe.CheckError(err)

		r_op, err := strconv.ParseUint(input.Right_op, 16, 64)
		bagpipe.CheckError(err)

		s_l_op := fmt.Sprintf("%d", l_op)
		s_r_op := fmt.Sprintf("%d", r_op)

		log_line := s_l_op + " " + s_r_op + " " + output.Instr_count + " " + output.Cycle_count

		bagpipe.AppendFile(data_dir+"/"+log_filename, log_line+"\n")
	}

	bagpipe.UpdateStatus("test complete, results in results/" + arch +
		"/data/" + log_filename + ".\n")
}

func generate_int_operands() []uint64 {
	var operand_list []uint64
	operand_list = append(operand_list, 0)

	for idx := uint64(0); idx < 60; idx += 1 {
		operand := uint64(1) << idx
		operand_list = append(operand_list, operand)
	}

	return operand_list
}

func test_prediction(arch string, instr string, pred_file string) {
	file_contents := bagpipe.ReadFile(pred_file)
	lines := strings.Split(file_contents, "\n")

	type exp_t struct {
		x_value         string
		y_value         string
		predicted_value float64
		actual_value    float64
	}

	var exps []exp_t

	for _, line := range lines {
		if len(line) != 0 {
			fields := strings.Split(line, " ")
			x_value := fields[0]
			y_value := fields[1]

			prediction, err := strconv.ParseFloat(fields[2], 64)
			bagpipe.CheckError(err)

			exp := exp_t{x_value: x_value, y_value: y_value,
				predicted_value: prediction, actual_value: 0}
			exps = append(exps, exp)
		}
	}

	emulator_dir := get_emulator_dir(arch)
	emulator_bin := get_emulator_bin(arch)

	sprinter := bagpipe.NewSprinter(exec_one, MAX_THREAD_COUNT, len(lines)-1)

	for idx, exp := range exps {
		status := fmt.Sprintf("%4d of %4d", idx, len(exps))
		bagpipe.UpdateStatus("testing pred for " + instr + " [ " + status +
			" ] ... ")

		s_predicted := strconv.FormatFloat(exp.predicted_value, 'f', 2, 64)

		input := input_t{Instr: instr, Left_op: exp.x_value,
			Right_op: exp.y_value, Emulator_dir: emulator_dir,
			Emulator_bin: emulator_bin, Aux_field: s_predicted}

		sprinter.FeedWorker(input)
	}

	var diffs []float64

	for idx := 0; idx < sprinter.ResultCount(); idx += 1 {
		__input, __output := sprinter.ReadResult()

		input := decode_input(__input)
		output := decode_output(__output)

		predicted, err := strconv.ParseFloat(input.Aux_field, 64)
		bagpipe.CheckError(err)

		actual, err := strconv.ParseFloat(output.Cycle_count, 64)
		bagpipe.CheckError(err)

		diffs = append(diffs, math.Abs(actual-predicted))
	}

	print_prediction_stats(diffs)
}

func print_prediction_stats(diffs []float64) {
	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i] < diffs[j] {
			return true
		}

		return false
	})

	idx_95 := int(0.95 * float64(len(diffs)))
	idx_99 := int(0.99 * float64(len(diffs)))
	idx_max := len(diffs) - 1

	status := fmt.Sprintf("95th percentile error: %.2f cycle(s), 99th "+
		"percentile error: %.2f cycle(s), maximum error: %.2f cycle(s)\n",
		diffs[idx_95], diffs[idx_99], diffs[idx_max])

	bagpipe.UpdateStatus(status)
}

func is_valid_instr(instr string) bool {
	return contains(sp_objects, instr) || contains(dp_objects, instr) ||
		contains(int_objects, instr)
}

func is_valid_arch(arch string) bool {
	return arch == "rocket" || arch == "boom"
}

func is_valid_operand_type(operand_type string) bool {
	return operand_type == "integer" || operand_type == "normal" ||
		operand_type == "subnormal"
}

func main() {
	flag.Parse()

	clean_cmd := flag.NewFlagSet("clean", flag.ExitOnError)
	sweep_cmd := flag.NewFlagSet("sweep", flag.ExitOnError)
	validate_cmd := flag.NewFlagSet("validate", flag.ExitOnError)

	sweep_instr := sweep_cmd.String("instr", "",
		"`instruction` to test (required)")

	sweep_arch := sweep_cmd.String("arch", "rocket",
		"`architecture` (rocket or boom)")

	sweep_op1 := sweep_cmd.String("operand1-type", "integer",
		"`type` of the first operand to the instruction "+
			"(integer / normal / subnormal)")

	sweep_op2 := sweep_cmd.String("operand2-type", "integer",
		"`type` of the second operand to the instruction "+
			"(integer / normal / subnormal)")

	validate_instr := validate_cmd.String("instr", "",
		"`instruction` to validate (required)")

	validate_arch := validate_cmd.String("arch", "rocket",
		"`architecture` (rocket or boom)")

	validate_results := validate_cmd.String("prediction-file", "",
		"`file` containing predictions produced by the R script (required)")

	args := os.Args[1:]

	if len(args) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if args[0] == "clean" {
	} else if args[0] == "sweep" {
		sweep_cmd.Parse(args[1:])
	} else if args[0] == "validate" {
		validate_cmd.Parse(args[1:])
	} else {
		fmt.Println("command not in { clean | sweep | validate }, exiting ...")
		os.Exit(1)
	}

	if clean_cmd.Parsed() {
		clean(sp_objects, dtype_sp)
		clean(dp_objects, dtype_dp)
		clean(int_objects, dtype_int)
		clean(mem_objects, dtype_mem)
	} else if sweep_cmd.Parsed() {
		if is_valid_arch(*sweep_arch) == false {
			fmt.Println("Invalid 'arch' for the 'sweep' command.\n")
			sweep_cmd.PrintDefaults()
			os.Exit(1)
		}

		if is_valid_operand_type(*sweep_op1) == false {
			fmt.Println("Invalid 'operand1-type' for the 'sweep' command.\n")
			sweep_cmd.PrintDefaults()
			os.Exit(1)
		}

		if is_valid_operand_type(*sweep_op2) == false {
			fmt.Println("Invalid 'operand2-type' for the 'sweep' command.\n")
			sweep_cmd.PrintDefaults()
			os.Exit(1)
		}

		if is_valid_instr(*sweep_instr) == false {
			fmt.Println("Invalid 'instr' for the 'sweep' command.\n")
			sweep_cmd.PrintDefaults()
			os.Exit(1)
		}

		sweep_instr_operands(*sweep_arch, *sweep_op1, *sweep_op2, *sweep_instr)
	} else if validate_cmd.Parsed() {
		if is_valid_arch(*validate_arch) == false {
			fmt.Println("Invalid 'arch' for the 'validate' command.\n")
			validate_cmd.PrintDefaults()
			os.Exit(1)
		}

		if is_valid_instr(*validate_instr) == false {
			fmt.Println("Invalid 'instr' for the 'validate' command.\n")
			validate_cmd.PrintDefaults()
			os.Exit(1)
		}

		test_prediction(*validate_arch, *validate_instr, *validate_results)
	}
}
