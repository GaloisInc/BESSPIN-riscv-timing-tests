#!/usr/bin/env python
import sys, os, glob
import argparse
import random
import math
import tempfile
import shutil
import subprocess

sys.path.append(os.environ['BESSPIN_GFE_SCRIPT_DIR'])
from test_gfe_unittest import TestGfeP1,TestGfeP2,TestGfeP3

class BaseTimingTest:
    def run_timing_test(self, elf):
        self.gfe.gdb_session.interrupt()
        self.gfe.launchElf(elf,verify=False)

    def collect_lines(self, n):
        collected = []
        nlines = 0
        while nlines < n:
            pending = self.gfe.uart_session.in_waiting
            if pending:
                fetched = self.gfe.uart_session.read(pending)
                fetchedLines = fetched.decode('utf-8').rstrip().split('\n')
                for l in fetchedLines:
                    if 'instr' in l:
                        collected.append(l)
                nlines = len(collected)
        return collected


class TimingTestP1(TestGfeP1, BaseTimingTest):
    pass
class TimingTestP2(TestGfeP2, BaseTimingTest):
    pass
class TimingTestP3(TestGfeP3, BaseTimingTest):
    pass

class TimingMockTester:
    def __init__(self):
        self.lines = []
    def setUp(self):
        pass
    def setupUart(self):
        pass
    def run_timing_test(self, f):
        words = os.path.basename(f).split(".")
        if words[1] == "int":
            # test.int.int_instr.op1.op2
            op1,op2 = words[3:]
        else:
            # test.fop.finstr.[ds].op1.op2
            op1,op2 = words[4:]
        op1 = hex(int(op1))[2:]
        op2 = hex(int(op2))[2:]
        self.lines.append(f"op1\t{op1}\top2\t{op2}\tinstrs\tb\tcycles\t35")

    def collect_lines(self, n):
        ret = self.lines[:n]
        self.lines = self.lines[n:]
        return ret

instructions = []
def intOperands(xlen):
    return [ (1 << (8*idx)) - 1 for idx in range(0,8) ]

def floatOperands(mantissabits, expbits):
    def ones(n):
        return (1 << n) - 1

    mantissaMask = ones(mantissabits)
    expMask      = ones(expbits)
    fixed        = [ 0,
                     expMask << mantissabits, # Positive Inf
                     (expMask << mantissabits) | (1 << 31), # Negative Inf
                   ]
    def oneRound():
        mantissa = random.randint(1,mantissaMask) # 23 bits
        exponent = random.randint(1,expMask) << mantissabits
        subnormal   = random.randint(1,mantissaMask)
        subnormal2  = random.randint(1,mantissaMask)
        return [ mantissa | exponent, # Normal float
                 subnormal, # Subnormal
                 (expMask << mantissabits) | subnormal2, # NAN
               ]
    ops = fixed
    ops = ops + oneRound()
    ops = ops + oneRound()
    ops = ops + oneRound()
    return ops

def singlePrecOperands(xlen):
    return floatOperands(23,8)

def doublePrecOperands(xlen):
    return floatOperands(52,11)

OPERANDS = {
    'int'    : intOperands,
    'single' : singlePrecOperands,
    'double' : doublePrecOperands
}

INT_OPS = [ "sll",
            "srl",
            "sra",
            "add",
            "sub",
            "xor",
            "and",
            "or",
            "slt",
            "sltu",
	        "mul",
            "mulh",
            "mulhsu",
            "mulhu",
            "div",
            "divu",
            "rem",
            "remu" ]

FL_OPS = [ "fadd",
           "fsub",
           "fmul",
           "fdiv",
           "fsgnj",
           "fsgnjn",
           "fsgnjx",
           "fmin",
           "fmax" ]
SP_FL_OPS = [ op + ".s" for op in FL_OPS ]
DP_FL_OPS = [ op + ".d" for op in FL_OPS ]

def sweepOperands(sweepConfig):
    return OPERANDS[sweepConfig['instr_type']](sweepConfig['xlen'])

def buildTests(i, t, operands):
    f = open("build.sh", "w")
    cmd = []
    for o1 in operands:
        for o2 in operands:
            f.write(f"make INST={i} OP1={o1} OP2={o2} TYPE={t}\n")
    f.close()
    try:
        print("Compiling tests...",end="",flush=True)
        subprocess.check_output(["bash", "build.sh"], stderr=subprocess.STDOUT)
        print("Done")
    except subprocess.CalledProcessError as e:
        print(e.output)
        sys.exit(1)

def format_line(line):
    _, op1, _, op2, _, istrs, _, cycles = line.split('\t')
    op1    = int(op1,16)
    op2    = int(op2,16)
    istrs  = int(istrs,16)
    cycles = int(cycles,16)
    cpi    = cycles/istrs
    return f"{op1} {op2} {hex(istrs)[2:]} {cpi}\n"

def format_lines(lines):
    return [ format_line(line) for line in lines ]

def sweep(sweepConfig):
    # Pick the tester
    hFpga = None
    if sweepConfig['dry_run']:
        hFpga = TimingMockTester()
    elif sweepConfig['ptype'] == 'p1':
        hFpga = TimingTestP1()
    elif sweepConfig['ptype'] == 'p2':
        hFpga = TimingTestP2()
    elif sweepConfig['ptype'] == 'p3':
        hFpga = TimingTestP3()
    else:
        print("I don't know what sort of fpga to use!")
        print(f"Unhandled ptype: {sweepConfig['ptype']}")
        sys.exit(1)

    hFpga.setUp()
    hFpga.setupUart()

    # Set up build dir
    workDirName = sweepConfig['work_dir']
    if workDirName is None:
        workDir = tempfile.TemporaryDirectory(prefix="timing-work")
        workDirName = workDir.name

    srcLoc = os.path.join(os.path.dirname(sys.argv[0]), "../src")
    buildLoc = os.path.join(workDirName, "work")
    shutil.copytree(srcLoc, buildLoc)
    os.chdir(buildLoc)
    os.mkdir("build")

    # Get the operands we are going to test
    operands = sweepOperands(sweepConfig)
    i = sweepConfig['instr']
    t = sweepConfig['instr_type']
    n = len(operands)*len(operands)

    print(f"Sweeping {t} instruction: {i}")

    # Generate a script containing all the make invocations
    # so that we don't open so many subprocesses
    buildTests(i, t, operands)

    # Execute each test
    files = glob.glob(os.path.join('build', '*'))
    c = 0 # number of tests for pretty-printing a progress bar
    lines = []
    for f in files:
        c  = c + 1
        pct = math.floor((c / n) * 40)
        hFpga.run_timing_test(f)
        l = hFpga.collect_lines(1)
        lines.append(l[0])
        print(f"Running:\t{c}/{n}[" + "#"*pct + "-"*(40-pct) + "]", end="\r")
    print("")
    for l in format_lines(lines):
        sweepConfig['output'].write(l)

DESCR = 'Collect instruction timing statistics'

def parseProc(procstr):
    try:
        name, p = procstr.split('_')
    except:
        print("Expected <proc>_<p> for processor")
        print("  e.g. chisel_p1, bluespec_p2")
        sys.exit(1)
    return (name, p)

def instrType(instr):
    if instr in INT_OPS:
        ty   = 'int'
    elif instr in SP_FL_OPS:
        ty   = 'single'
    elif instr in DP_FL_OPS:
        ty   = 'double'
    else:
        print(f"Unrecognized instruction: {instr}")
        sys.exit(1)

    return ty

def parseConfig(arglist):
    parser     = argparse.ArgumentParser(description=DESCR)
    subparsers = parser.add_subparsers(dest='cmd')

    sweep      = subparsers.add_parser('sweep')
    sweep.add_argument('--instr', type=str)
    sweep.add_argument('--proc',  type=str)
    sweep.add_argument('--dry_run', action='store_true', default=False)
    sweep.add_argument('--keep_temp', action='store_true', default=False)
    sweep.add_argument('--output', type=str)
    sweep.add_argument('--work_dir', type=str, default=None)

    args = parser.parse_args(arglist)

    proc = parseProc(args.proc)

    return { 'xlen'       : 32 if 'p1' == proc[1] else 64,
             'ptype'      : proc[1],
             'instr'      : args.instr,
             'instr_type' : instrType(args.instr),
             'mode'       : args.cmd,
             'dry_run'    : args.dry_run,
             'no_cleanup' : args.keep_temp,
             'output'     : args.output,
             'work_dir'   : args.work_dir
           }

def main(args):
    cfg = parseConfig(args)
    if cfg['mode'] == "sweep":
        try:
            output = cfg['output']
            out = open(cfg['output'], 'w')
            cfg['output'] = out
        except:
            print(f"Error opening {output} for writing")
            sys.exit(1)
        sweep(cfg)


if __name__ == "__main__":
    main(sys.argv[1:])
