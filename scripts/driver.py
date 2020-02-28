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
    pass

class TimingTestP1(TestGfeP1, BaseTimingTest):
    pass
class TimingTestP2(TestGfeP2, BaseTimingTest):
    pass
class TimingTestP3(TestGfeP3, BaseTimingTest):
    pass


# test = TimingTest()

instructions = []
def intOperands(xlen):
    return [ (1 << (8*idx)) - 1 for idx in range(0,8) ]

def floatOperands(mantissabits, expbits):
    def ones(n):
        return (1 << n) - 1

    mantissaMask = ones(mantissabits)
    expMask      = ones(expbits)
    def oneRound():
        mantissa = random.randint(1,mantissaMask) # 23 bits
        exponent = random.randint(1,expMask) << mantissabits
        subnormal   = random.randint(1,mantissaMask)
        subnormal2  = random.randint(1,mantissaMask)
        return [ 0, # Zero
                 mantissa | exponent, # Normal float
                 subnormal, # Subnormal
                 expMask << mantissabits, # Positive Inf
                 (expMask << mantissabits) | (1 << 31), # Negative Inf
                 (expMask << mantissabits) | subnormal2, # NAN
        ]
    ops = oneRound()
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

def sweep(sweepConfig):
    workDir = tempfile.TemporaryDirectory(prefix="timing-work")
    srcLoc = os.path.join(os.path.dirname(sys.argv[0]), "../src")
    buildLoc = os.path.join(workDir.name, "work")
    shutil.copytree(srcLoc, buildLoc)
    os.chdir(buildLoc)
    os.mkdir("build")
    # Get the operands we are going to test
    operands = sweepOperands(sweepConfig)
    i = sweepConfig['instr']
    t = sweepConfig['instr_type']
    n = len(operands)*len(operands)
    c = 0

    print(f"Sweeping {t} instruction: {i}")
    for o1 in operands:
        for o2 in operands:
            c = c + 1
            try:
                subprocess.check_output(f"make INST={i} OP1={o1} OP2={o2} TYPE={t}".split(" "), stderr=subprocess.STDOUT)
            except subprocess.CalledProcessError as e:
                print(e.output)
                sys.exit(1)
            pct = math.floor((c / n) * 40)
            print("Compiling Tests: [" + "#"*pct + "-"*(40-pct) + "]", end="\r")

    c = 0
    print("\n")
    breakpoint()
    for f in glob.glob(os.path.join('build', '*')):
        c  = c + 1
        pct = math.floor((c / n) * 40)
        print("Running Tests: [" + "#"*pct + "-"*(40-pct) + "]", end="\r")
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

    args = parser.parse_args(arglist)

    proc = parseProc(args.proc)

    return { 'xlen'       : 32 if 'p1' == proc[1] else 64,
             'ptype'      : proc[1],
             'instr'      : args.instr,
             'instr_type' : instrType(args.instr),
             'mode'       : args.cmd,
             'dry_run'    : args.dry_run,
             'no_cleanup' : args.keep_temp
           }

def main(args):
    cfg = parseConfig(args)
    if cfg['mode'] == "sweep":
        sweep(cfg)

if __name__ == "__main__":
    main(sys.argv[1:])
