package main

import (
	"fmt"
	"gitlab.com/ashay/bagpipe"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type dtype_t int

const (
	dtype_int dtype_t = 0
	dtype_sp  dtype_t = 1
	dtype_dp  dtype_t = 2
	dtype_mem dtype_t = 3
)

var CC = "riscv64-unknown-elf-gcc"
var CFLAGS = "-I include -mcmodel=medany -std=gnu99 -O2"
var LDFLAGS = "-static -nostartfiles -T test.ld"

func compile_driver(inst string, op1 string, op2 string,
	driver_file string) string {

	if bagpipe.FileExists(driver_file) == false {
		log.Fatal(driver_file + " not found")
	}

	obj_driver := bagpipe.CreateTempFile("driver.o.")

	cmd := CC + " " + CFLAGS + " -c " + driver_file + " -o " + obj_driver +
		" -DINST=" + inst + " -DOP1=0x" + op1 + " -DOP2=0x" + op2

	if len(inst) == 2 && (inst[:1] == "l" || inst[:1] == "s") {
		bagpipe.ExecCommand(cmd+" -DPREFIX=x", bagpipe.WorkingDirectory())
	} else if inst[:2] == "fl" || inst[:2] == "fs" {
		bagpipe.ExecCommand(cmd+" -DPREFIX=f", bagpipe.WorkingDirectory())
	} else {
		bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
	}

	return obj_driver
}

func assemble_crt() string {
	if bagpipe.FileExists("crt.S") == false {
		log.Fatal("crt.S not found")
	}

	obj_file := bagpipe.CreateTempFile("crt.o.")
	cmd := CC + " " + CFLAGS + " -c crt.S -o " + obj_file

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
	return obj_file
}

func compile_syscalls() string {
	if bagpipe.FileExists("syscalls.c") == false {
		log.Fatal("syscalls.c not found")
	}

	obj_file := bagpipe.CreateTempFile("syscalls.o.")
	cmd := CC + " " + CFLAGS + " -c syscalls.c -o " + obj_file

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
	return obj_file
}

func link(bin string, obj_crt string, obj_syscalls string, i1 string,
	op1 string, i2 string, op2 string, dtype dtype_t) {

	out_file := bin + "." + i1 + "." + i2
	bagpipe.UpdateStatus("building " + out_file + " ...")

	var obj_driver string

	if dtype == dtype_int {
		obj_driver = compile_driver(bin, op1, op2, "int-driver.c")
	} else if dtype == dtype_sp {
		obj_driver = compile_driver(bin, op1, op2, "sp-driver.c")
	} else if dtype == dtype_dp {
		obj_driver = compile_driver(bin, op1, op2, "dp-driver.c")
	} else if dtype == dtype_mem {
		obj_driver = compile_driver(bin, op1, op2, "mem-driver.c")
	} else {
		log.Fatal("invalid data type")
	}

	cmd := CC + " " + obj_driver + " " + obj_crt + " " + obj_syscalls +
		" -O2 " + LDFLAGS + " -o " + out_file

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
	bagpipe.DeleteFile(obj_driver)
}

func build(objects []string, operands []string, dtype dtype_t) {
	obj_crt := assemble_crt()
	obj_syscalls := compile_syscalls()

	for _, object := range objects {
		for i1, op1 := range operands {
			if dtype == dtype_mem {
				link(object, obj_crt, obj_syscalls, strconv.Itoa(i1), op1, "0",
					"0", dtype_mem)
			} else {
				for i2, op2 := range operands {
					link(object, obj_crt, obj_syscalls, strconv.Itoa(i1), op1,
						strconv.Itoa(i2), op2, dtype)
				}
			}
		}
	}

	bagpipe.DeleteFile(obj_crt)
	bagpipe.DeleteFile(obj_syscalls)

	bagpipe.ClearStatus()
}

func clean(objects []string, operands []string, dtype dtype_t) {
	bagpipe.UpdateStatus("cleaning ...")

	for _, object := range objects {
		for i1, _ := range operands {
			if dtype == dtype_mem {
				out_file := object + "." + strconv.Itoa(i1) + ".0"

				if bagpipe.FileExists(out_file) {
					bagpipe.DeleteFile(out_file)
				}
			} else {
				for i2, _ := range operands {
					out_file := object + "." + strconv.Itoa(i1) + "." +
						strconv.Itoa(i2)

					if bagpipe.FileExists(out_file) {
						bagpipe.DeleteFile(out_file)
					}
				}
			}
		}
	}

	aux_objects := []string{
		"fdiv.s.n.n", "fdiv.s.n.s", "fdiv.s.s.n", "fdiv.s.s.s", "fdiv.d.n.n",
		"fdiv.d.n.s", "fdiv.d.s.n", "fdiv.d.s.s", "div.i.i", "divu.i.i",
		"rem.i.i", "remu.i.i",
	}

	for _, object := range aux_objects {
		if bagpipe.FileExists(object) {
			bagpipe.DeleteFile(object)
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

func exec(instr string, str_i1 string, str_i2 string, emulator_dir string,
	emulator_bin string) string {

	working_dir := bagpipe.WorkingDirectory()
	exec_file := instr + "." + str_i1 + "." + str_i2

	if bagpipe.FileExists(working_dir+"/"+exec_file) == false {
		log.Fatal("could not find " + exec_file)
	}

	cmd := emulator_bin + " -s 0 -c " + working_dir + "/" + exec_file
	return bagpipe.ExecCommand(cmd, emulator_dir)
}

func parse(line string) (string, string) {
	pattern := "instrs[\\s]*(\\d*)[\\s]*cycles[\\s]*(\\d*)"
	regex := regexp.MustCompile(pattern)

	if regex.MatchString(line) == false {
		log.Fatal("could not parse output: \"" + line + "\"")
	}

	match := regex.FindStringSubmatch(line)
	instr_count := match[1]
	cycle_count := match[2]

	return instr_count, cycle_count
}

func run(instr string, str_i1 string, str_i2 string, log_filename string,
	emulator_dir string, emulator_bin string, data_dir string) {

	exec_file := instr + "." + str_i1 + "." + str_i2
	bagpipe.UpdateStatus("running " + exec_file + " ... ")

	output := exec(instr, str_i1, str_i2, emulator_dir, emulator_bin)
	instrs, cycles := parse(output)

	log_line := str_i1 + " " + str_i2 + " " + instrs + " " + cycles
	bagpipe.AppendFile(data_dir+"/"+log_filename, log_line+"\n")
}

func run_benchmark(arch string, instr string, operands []string,
	dtype dtype_t) {
	var emulator_dir string
	var emulator_bin string

	if arch == "rock" {
		emulator_bin = "./emulator-freechips.rocketchip.system-DefaultConfig"
		emulator_dir = bagpipe.HomeDirectory() + "/src/rocket-chip/emulator"
	} else if arch == "boom" {
		emulator_bin = "./simulator-boom.system-BoomConfig"
		emulator_dir = bagpipe.HomeDirectory() + "/src/boom-template/verisim"
	}

	data_dir := bagpipe.WorkingDirectory() + "/../results/" + arch + "/data"

	log_filename := "out." + instr
	if bagpipe.FileExists(data_dir + "/" + log_filename) {
		bagpipe.DeleteFile(data_dir + "/" + log_filename)
	}

	for i1, _ := range operands {
		str_i1 := strconv.Itoa(i1)

		if dtype == dtype_mem {
			run(instr, str_i1, "0", log_filename, emulator_dir, emulator_bin,
				data_dir)
		} else {
			for i2, _ := range operands {
				str_i2 := strconv.Itoa(i2)
				run(instr, str_i1, str_i2, log_filename, emulator_dir,
					emulator_bin, data_dir)
			}
		}
	}

	bagpipe.UpdateStatus("test complete, results in results/" + arch +
		"/data/" + log_filename + ".\n")
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

func generate_int_operand() string {
	var word uint64
	rand.Seed(time.Now().UTC().UnixNano())

	for hamm_weight := 0; hamm_weight < rand.Intn(8); hamm_weight += 1 {
		word |= (1 << uint64(rand.Intn(63))) // Don't set the sign bit.
	}

	return fmt.Sprintf("%016x", word)
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

	if operand_type == "i" {
		return generate_int_operand()
	}

	log.Fatal("failed to recognize operand type!")
	return fmt.Sprintf("%016x", 0)
}

func rand_benchmark(arch string, opr1 string, opr2 string, instr string,
	dtype dtype_t) {
	var emulator_dir string
	var emulator_bin string

	if arch == "rock" {
		emulator_bin = "./emulator-freechips.rocketchip.system-DefaultConfig"
		emulator_dir = bagpipe.HomeDirectory() + "/src/rocket-chip/emulator"
	} else if arch == "boom" {
		emulator_bin = "./simulator-boom.system-BoomConfig"
		emulator_dir = bagpipe.HomeDirectory() + "/src/boom-template/verisim"
	}

	data_dir := bagpipe.WorkingDirectory() + "/../results/" + arch + "/data"

	log_filename := "out." + instr + "." + opr1 + "." + opr2
	if bagpipe.FileExists(data_dir + "/" + log_filename) {
		bagpipe.DeleteFile(data_dir + "/" + log_filename)
	}

	k_repeat_ctr := 1000

	obj_crt := assemble_crt()
	obj_syscalls := compile_syscalls()

	for ctr := 0; ctr < k_repeat_ctr; ctr += 1 {
		left_opr := generate_operand(opr1, dtype)
		right_opr := generate_operand(opr2, dtype)

		link(instr, obj_crt, obj_syscalls, opr1, left_opr, opr2, right_opr,
			dtype)

		status := strconv.Itoa(ctr+1) + " of " + strconv.Itoa(k_repeat_ctr)
		bagpipe.UpdateStatus("randomizing " + instr + " [" + status + "] ... ")

		output := exec(instr, opr1, opr2, emulator_dir, emulator_bin)
		instrs, cycles := parse(output)

		log_line := left_opr + " " + right_opr + " " + instrs + " " + cycles
		bagpipe.AppendFile(data_dir+"/"+log_filename, log_line+"\n")
	}

	bagpipe.DeleteFile(obj_crt)
	bagpipe.DeleteFile(obj_syscalls)

	bagpipe.UpdateStatus("test complete, results in results/" + arch +
		"/data/" + log_filename + ".\n")
}

func main() {
	int_objects := []string{
		"sll", "srl", "sra", "add", "sub", "xor", "and", "or", "slt", "sltu",
		"mul", "mulh", "mulhsu", "mulhu", "div", "divu", "rem", "remu",
	}

	int_operands := []string{
		"0", "f", "ff", "fff", "ffff", "fffff", "ffffff", "fffffff",
		"ffffffff", "fffffffff", "ffffffffff", "fffffffffff", "ffffffffffff",
		"fffffffffffff", "ffffffffffffff", "fffffffffffffff",
		"7fffffffffffffff",
	}

	sp_objects := []string{
		"fadd.s", "fsub.s", "fmul.s", "fdiv.s", "fsgnj.s", "fsgnjn.s",
		"fsgnjx.s", "fmin.s", "fmax.s",
	}

	sp_operands := []string{
		"00000000", "40600000", "00084000", "7f800000", "ff800000", "7f800200",
	}

	dp_objects := []string{
		"fadd.d", "fsub.d", "fmul.d", "fdiv.d", "fsgnj.d", "fsgnjn.d",
		"fsgnjx.d", "fmin.d", "fmax.d",
	}

	dp_operands := []string{
		"0000000000000000", "4025000000000000", "0000000000000400",
		"7ff0000000000000", "fff0000000000000", "7ff0000000000001",
	}

	mem_objects := []string{
		"lb", "sb", "lh", "sh", "lw", "sw", "flw", "fsw", "fld", "fsd",
	}

	mem_operands := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}

	if len(os.Args[1:]) == 0 {
		build(int_objects, int_operands, dtype_int)
	} else {
		for _, cmd := range os.Args[1:] {
			if cmd == "clean" {
				clean(dp_objects, dp_operands, dtype_int)
				clean(sp_objects, sp_operands, dtype_sp)
				clean(int_objects, int_operands, dtype_dp)
				clean(mem_objects, mem_operands, dtype_mem)
			} else if cmd == "build-int" {
				build(int_objects, int_operands, dtype_int)
			} else if cmd == "build-sp" {
				build(sp_objects, sp_operands, dtype_sp)
			} else if cmd == "build-dp" {
				build(dp_objects, dp_operands, dtype_dp)
			} else if cmd == "build-mem" {
				build(mem_objects, mem_operands, dtype_mem)
			} else if strings.HasPrefix(cmd, "run-") {
				arch := cmd[4:8]
				instr := cmd[9:]

				if arch != "rock" && arch != "boom" {
					log.Fatal("did not recognize architecture.")
				}

				if contains(int_objects, instr) {
					run_benchmark(arch, instr, int_operands, dtype_int)
				} else if contains(sp_objects, instr) {
					run_benchmark(arch, instr, sp_operands, dtype_sp)
				} else if contains(dp_objects, instr) {
					run_benchmark(arch, instr, dp_operands, dtype_dp)
				} else if contains(mem_objects, instr) {
					run_benchmark(arch, instr, mem_operands, dtype_mem)
				} else {
					log.Fatal(instr + " not found among instructions")
				}
			} else if strings.HasPrefix(cmd, "rand-") {
				rand.Seed(time.Now().UTC().UnixNano())

				fields := strings.Split(cmd, "-")
				arch := fields[1]

				if arch != "rock" && arch != "boom" {
					log.Fatal("did not recognize architecture.")
				}

				instr := fields[2]
				opr1 := fields[3]
				opr2 := fields[4]

				if opr1 != "i" && opr1 != "n" && opr1 != "s" {
					log.Fatal("did not recognize type of operand #1.")
				}

				if opr2 != "i" && opr2 != "n" && opr2 != "s" {
					log.Fatal("did not recognize type of operand #2.")
				}

				if contains(sp_objects, instr) {
					rand_benchmark(arch, opr1, opr2, instr, dtype_sp)
				} else if contains(dp_objects, instr) {
					rand_benchmark(arch, opr1, opr2, instr, dtype_dp)
				} else if contains(int_objects, instr) {
					rand_benchmark(arch, opr1, opr2, instr, dtype_int)
				}
			}
		}
	}
}
