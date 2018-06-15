package main

import (
	"gitlab.com/ashay/bagpipe"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type dtype_t int

const (
	dtype_int dtype_t = 0
	dtype_sp  dtype_t = 1
	dtype_dp  dtype_t = 2
)

var CC = "riscv64-unknown-elf-gcc"
var CFLAGS = "-I include -mcmodel=medany -std=gnu99 -O2"
var LDFLAGS = "-static -nostartfiles -T test.ld"

func compile_driver(inst string, op1 string, op2 string, driver_file string) {
	if bagpipe.FileExists(driver_file) == false {
		log.Fatal(driver_file + " not found")
	}

	cmd := CC + " " + CFLAGS + " -c " + driver_file + " -o driver.o -DINST=" +
		inst + " -DOP1=0x" + op1 + " -DOP2=0x" + op2

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
}

func assemble_crt() {
	if bagpipe.FileExists("crt.S") == false {
		log.Fatal("crt.S not found")
	}

	cmd := CC + " " + CFLAGS + " -c crt.S -o crt.o"
	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
}

func compile_syscalls() {
	if bagpipe.FileExists("syscalls.c") == false {
		log.Fatal("syscalls.c not found")
	}

	cmd := CC + " " + CFLAGS + " -c syscalls.c -o syscalls.o"
	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
}

func link(bin string, i1 int, op1 string, i2 int, op2 string, dtype dtype_t) {
	out_file := bin + "." + strconv.Itoa(i1) + "." + strconv.Itoa(i2)
	bagpipe.UpdateStatus("building " + out_file + " ...")

	if dtype == dtype_int {
		compile_driver(bin, op1, op2, "int-driver.c")
	} else if dtype == dtype_sp {
		compile_driver(bin, op1, op2, "sp-driver.c")
	} else if dtype == dtype_dp {
		compile_driver(bin, op1, op2, "dp-driver.c")
	} else {
		log.Fatal("invalid data type")
	}

	cmd := CC + " driver.o crt.o syscalls.o -O2 " + LDFLAGS + " -o " + out_file

	bagpipe.ExecCommand(cmd, bagpipe.WorkingDirectory())
}

func build(objects []string, operands []string, dtype dtype_t) {
	assemble_crt()
	compile_syscalls()

	for _, object := range objects {
		for i1, op1 := range operands {
			for i2, op2 := range operands {
				link(object, i1, op1, i2, op2, dtype)
			}
		}
	}

	bagpipe.UpdateStatus("build complete.\n")
}

func clean(objects []string, operands []string) {
	bagpipe.UpdateStatus("cleaning ...")

	for _, object := range objects {
		for i1, _ := range operands {
			for i2, _ := range operands {
				out_file := object + "." + strconv.Itoa(i1) + "." +
					strconv.Itoa(i2)

				if bagpipe.FileExists(out_file) {
					bagpipe.DeleteFile(out_file)
				}
			}
		}
	}

	if bagpipe.FileExists("crt.o") {
		bagpipe.DeleteFile("crt.o")
	}

	if bagpipe.FileExists("driver.o") {
		bagpipe.DeleteFile("driver.o")
	}

	if bagpipe.FileExists("syscalls.o") {
		bagpipe.DeleteFile("syscalls.o")
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

func run(instr string, str_i1 string, str_i2 string, log_filename string,
	emulator_dir string, emulator_bin string, plots_dir string) {
	working_dir := bagpipe.WorkingDirectory()

	exec_file := instr + "." + str_i1 + "." + str_i2

	if bagpipe.FileExists(working_dir+"/"+exec_file) == false {
		log.Fatal("could not find " + exec_file)
	}

	cmd := emulator_bin + " -s 0 -c " + working_dir + "/" + exec_file

	bagpipe.UpdateStatus("running " + exec_file + " ... ")
	output := bagpipe.ExecCommand(cmd, emulator_dir)

	pattern := "instrs[\\s]*(\\d*)[\\s]*cycles[\\s]*(\\d*)"
	regex := regexp.MustCompile(pattern)

	if regex.MatchString(output) == false {
		log.Fatal("could not parse output: \"" + output + "\"")
	}

	match := regex.FindStringSubmatch(output)
	instr_count := match[1]
	cycle_count := match[2]

	log_line := str_i1 + " " + str_i2 + " " + instr_count + " " + cycle_count
	bagpipe.AppendFile(plots_dir+"/"+log_filename, log_line+"\n")
}

func run_benchmark(arch string, instr string, operands []string) {
	var emulator_dir string
	var emulator_bin string

	if arch == "rock" {
		emulator_bin = "./emulator-freechips.rocketchip.system-DefaultConfig"
		emulator_dir = bagpipe.HomeDirectory() + "/src/rocket-chip/emulator"
	} else if arch == "boom" {
		emulator_bin = "./simulator-boom.system-BoomConfig"
		emulator_dir = bagpipe.HomeDirectory() + "/src/boom-template/verisim"
	}

	plots_dir := bagpipe.WorkingDirectory() + "/../results/" + arch + "/data"

	log_filename := "out." + instr
	if bagpipe.FileExists(plots_dir + "/" + log_filename) {
		bagpipe.DeleteFile(plots_dir + "/" + log_filename)
	}

	for i1, _ := range operands {
		str_i1 := strconv.Itoa(i1)

		for i2, _ := range operands {
			str_i2 := strconv.Itoa(i2)
			run(instr, str_i1, str_i2, log_filename, emulator_dir, emulator_bin,
				plots_dir)
		}
	}

	bagpipe.UpdateStatus("test complete, results in " + arch + "/" +
		log_filename + ".\n")
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

	if len(os.Args[1:]) == 0 {
		build(int_objects, int_operands, dtype_int)
	} else {
		for _, cmd := range os.Args[1:] {
			if cmd == "clean" {
				clean(dp_objects, dp_operands)
				clean(sp_objects, sp_operands)
				clean(int_objects, int_operands)
			} else if cmd == "build-int" {
				build(int_objects, int_operands, dtype_int)
			} else if cmd == "build-sp" {
				build(sp_objects, sp_operands, dtype_sp)
			} else if cmd == "build-dp" {
				build(dp_objects, dp_operands, dtype_dp)
			} else if strings.HasPrefix(cmd, "run-") {
				arch := cmd[4:8]

				if arch != "rock" && arch != "boom" {
					log.Fatal("did not recognize architecture.")
				}

				instr := cmd[9:]
				if contains(int_objects, instr) {
					run_benchmark(arch, instr, int_operands)
				} else if contains(sp_objects, instr) {
					run_benchmark(arch, instr, sp_operands)
				} else if contains(dp_objects, instr) {
					run_benchmark(arch, instr, dp_operands)
				} else {
					log.Fatal(instr + " not found among instructions")
				}
			}
		}
	}
}
