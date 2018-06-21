## RISCV Instruction Latency Tests

This code measures the latency of various RISV instructions from the Base ISA and from the M, F, and D extensions on the Rocket chip (https://github.com/freechipsproject/rocket-chip) using a Verilator-based simulation.


### Prerequisites

  - Go (with the `gitlab.com/ashay/bagpipe` package)
  - R (with `ggplot2` and `reshape2` packages)
  - Scripts assume `riscv64-unknown-elf-gcc` is in `PATH`, and that the Rocket chip is built in `${HOME}/src/rocket-chip/emulator`.


### How to Gather Data and Plot Results

    cd src
    
    # Specify targets to build
    go run ../scripts/driver.go build-int build-sp build-dp
    
    # Run tests for a specific architecture and instruction
    go run ../scripts/driver.go run-rock-fdiv.s
    go run ../scripts/driver.go run-boom-add
    go run ../scripts/driver.go run-boom-div
    
    # Plot data (for all instructions)
    cd ../results/rock/plots
    R --no-save < ../../../scripts/plot.R

    cd ../results/boom/plots
    R --no-save < ../../../scripts/plot.R


### Note about Debug Interrupts

If you plan to tweak the loop count in `src/*-driver.c`, note that running the test for too long may cause the (occassional) debug interrupts from the simulation to perturb the results.  In particular, if you see a wide variation in the instruction count (despite only changing the operand values), then the debug interrupts are suspect.  See https://github.com/freechipsproject/rocket-chip/issues/1495 for details and how to know whether the debug interrupt occured.


## Results

The following tables show the slowdown in instruction execution based on the choice of operand values.  For integer instructions, the operands range from 0x0 to 0x7fff\_ffff\_ffff\_ffff, where each increment represents the lowermost bits (in multiples of 4) being set to 1s.  For floating-point instructions, the operands are from the set { zero, normal, subnormal, +inf, -inf, and not-a-number }.

### Rocket

Plots for rocket chip are located [here](rocket-results.md).

|  instruction(s) | slowdown |
| --------------- | -------- |
| [`fdiv.d`](results/rock/plots/plot-fdiv.d.png) | 10.8x |
| [`div`](results/rock/plots/plot-div.png), [`divu`](results/rock/plots/plot-divu.png), [`rem`](results/rock/plots/plot-rem.png), [`remu`](results/rock/plots/plot-remu.png) | 10.3x |
| [`fdiv.s`](results/rock/plots/plot-fdiv.s.png) | 5.3x |
| [`mul`](results/rock/plots/plot-mul.png) | 1.7x |


### BOOM

Plots for boom chip are located [here](boom-results.md).

|  instruction(s) | slowdown |
| --------------- | -------- |
| [`div`](results/boom/plots/plot-div.png), [`divu`](results/boom/plots/plot-divu.png), [`rem`](results/boom/plots/plot-rem.png), [`remu`](results/boom/plots/plot-remu.png) | 6.1x |
| [`fdiv.s`](results/boom/plots/plot-fdiv.s.png), [`fdiv.d`](results/boom/plots/plot-fdiv.d.png) | 3.3x |
