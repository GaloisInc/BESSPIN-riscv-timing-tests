#!/usr/bin/env python
import sys, os, glob, stat
import argparse
import random
import math
import tempfile
import shutil
import subprocess

sys.path.append(os.environ['BESSPIN_GFE_SCRIPT_DIR'])
from test_gfe_unittest import TestGfeP1,TestGfeP2,TestGfeP3
read = 0
class BaseTimingTest:
    def run_timing_test(self, elf):
        self.gfe.gdb_session.interrupt()
        self.gfe.launchElf(elf,gdb_log=True,openocd_log=True,verify=True)

    def collect_lines(self, n):
        global read
        collected = []
        nlines = 0
        while nlines < n:
            pending = self.gfe.uart_session.in_waiting
            if pending:
                fetched = self.gfe.uart_session.read(pending)
                read = read + pending
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

def ones(n):
    return (1 << n) - 1

def floatOperands(mantissabits, expbits):

    mantissaMask = ones(mantissabits)
    expMask      = ones(expbits)
    fixed        = [ 0,
                     expMask << mantissabits, # Positive Inf
                     (expMask << mantissabits) | (1 << (mantissabits+expbits)), # Negative Inf
                   ]
    def oneRound():
        mantissa = random.randint(1,mantissaMask) # 23 bits
        exponent = random.randint(1,expMask) << mantissabits
        subnormal   = random.randint(1,mantissaMask)
        subnormal2  = random.randint(1,mantissaMask)
        return [ mantissa | exponent, # Normal float
                 subnormal, # Subnormal
                 (expMask << mantissabits) | subnormal2, # NAN
               ] + fixed
    ops = oneRound()
    ops = ops + oneRound()
    ops = ops + oneRound()
    return ops

def classify(v, mantissabits, expbits):
    widthmask = ones(mantissabits + expbits + 1)
    exp = (v >> mantissabits) & ones(expbits)
    if exp == 0:
        if (v << 1) & widthmask != 0:
            return "subnormal"
        return "zero"
    elif exp == ones(expbits):
        if (v << (expbits + 1)) & widthmask != 0:
            return "nan"
        if v >> (mantissabits+expbits) == 0:
            return "+inf"
        return "-inf"
    return "normal"


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

def buildTests(i, t, operands, xlen):
    f = open("build.sh", "w")
    cmd = []
    ts = []
    n = 0
    for o1 in operands:
        for o2 in operands:
            ts.append(f"build/test.{t}.{i}.{o1}.{o2}")
            f.write(f"make INST={i} OP1={o1} OP2={o2} TYPE={t} N={100} XLEN={xlen}\n")
            n = n+1
    print(f"{n} tests to build.")
    f.close()
    try:
        print("Compiling tests...",end="",flush=True)
        subprocess.check_output(["bash", "build.sh"], stderr=subprocess.STDOUT)
        print("Done")
    except subprocess.CalledProcessError as e:
        print(e.output)
        sys.exit(1)
    return ts

def format_line(ty, line):
    _, op1, _, op2, _, istrs, _, cycles = line.split('\t')
    op1    = int(op1,16)
    op2    = int(op2,16)
    if ty == "single":
        op1    = classify(op1, 23, 8)
        op2    = classify(op2, 23, 8)
    elif ty == "double":
        op1    = classify(op1, 52, 11)
        op2    = classify(op2, 52, 11)
    istrs  = int(istrs,16)
    cycles = int(cycles,16)
    cpi    = cycles/istrs
    return f"{op1} {op2} {hex(istrs)[2:]} {cpi}\n"

def format_lines(t,lines):
    return [ format_line(t,line) for line in lines ]

def setupBuildDir(sweepConfig):
    # Set up build dir -- either user provided
    # or a new temporary directory
    workDirName = sweepConfig['work_dir']
    if workDirName is None:
        workDir = tempfile.TemporaryDirectory(prefix="timing-work")
        workDirName = workDir.name
        # The directory will get cleaned up and hence deleted if we don't
        # keep around a live reference to it, so that's why we're saving it
        # in the config (and hey, it might be useful later)
        sweepConfig['work_dir'] = workDir

    # Copy the C sources to "<work-directory>/work"
    srcLoc = os.path.join(os.path.dirname(sys.argv[0]), "../src")
    buildLoc = os.path.join(workDirName, "work")
    shutil.copytree(srcLoc, buildLoc)
    dperm = stat.S_IRUSR | stat.S_IWUSR | stat.S_IXUSR
    fperm = stat.S_IRUSR | stat.S_IWUSR
    os.chmod(buildLoc, dperm)
    os.chdir(buildLoc)

    for root,dirs,fs in os.walk("."):
        for d in dirs:
            os.chmod(os.path.join(root,d), dperm)
        for f in fs:
            os.chmod(os.path.join(root,f), fperm)

    # Sometimes build is left over from development, so nuke it in
    # the work-dir copy
    try:
        os.mkdir("build")
    except FileExistsError as e:
        shutil.rmtree("build")
        os.mkdir("build")

    # Clean up any objects that were left lying around during e.g. development
    for d,_,fs in os.walk("."):
        for f in fs:
            if os.path.splitext(f)[1] == ".o":
                os.rm(os.path.join(d,f))

def getFpgaHandle(sweepConfig):
    if not sweepConfig['no_program']:
        print("Programming fpga with " + sweepConfig['proc'])
        try:
            subprocess.check_output(["gfe-program-fpga", sweepConfig['proc']])
        except subprocess.CalledProcessError as e:
            print(f"Failed to program fpga {sweepConfig['proc']}:")
            print(e)
            sys.exit(1)

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

    return hFpga

def sweep(sweepConfig):
    # Pick the tester
    hFpga = getFpgaHandle(sweepConfig)

    setupBuildDir(sweepConfig)

    # Get the operands we are going to test
    operands = sweepOperands(sweepConfig)
    i = sweepConfig['instr']
    ty = sweepConfig['instr_type']
    x = sweepConfig['xlen']

    print(f"Sweeping instruction: {i} [{ty}]")

    # Generate a script containing all the make invocations
    # so that we don't open so many subprocesses
    ts = buildTests(i, ty, operands, x)

    c = 0 # number of tests for pretty-printing a progress bar
    lines = []
    n = len(ts)
    for t in ts:
        c  = c + 1
        pct = math.floor((c / n) * 40)
        hFpga.run_timing_test(t)
        l = hFpga.collect_lines(1)
        lines.append(l[0])
        print(f"Running:\t{c}/{n} [" + "#"*pct + "-"*(40-pct) + "]", end="\r")
    print("")
    for l in format_lines(ty, lines):
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
    sweep.add_argument('instr', type=str)
    sweep.add_argument('proc',  type=str)
    sweep.add_argument('output', type=str)
    sweep.add_argument('--dry_run', action='store_true', default=False)
    sweep.add_argument('--keep_temp', action='store_true', default=False)
    sweep.add_argument('--work_dir', type=str, default=None)
    sweep.add_argument('--no_program', action='store_true', default=False)

    args = parser.parse_args(arglist)

    proc = parseProc(args.proc)

    return { 'xlen'       : 32 if 'p1' == proc[1] else 64,
             'ptype'      : proc[1],
             'proc'       : args.proc,
             'instr'      : args.instr,
             'instr_type' : instrType(args.instr),
             'mode'       : args.cmd,
             'dry_run'    : args.dry_run,
             'no_cleanup' : args.keep_temp,
             'output'     : args.output,
             'work_dir'   : args.work_dir,
             'no_program' : args.no_program
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
    try:
        main(sys.argv[1:])
    except SystemExit as e:
        print("die")
