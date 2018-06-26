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

    # Randomize integer operands of div, divu, rem, and remu instructions.
    go run ../scripts/driver.go rand-rock-div-i-i
    go run ../scripts/driver.go rand-rock-divu-i-i
    go run ../scripts/driver.go rand-rock-rem-i-i
    go run ../scripts/driver.go rand-rock-remu-i-i

    # Validate prediction accuracy of div, divu, rem, and remu instructions.
    go run ../scripts/divrem-rocket-predict.go ../results/rock/data/out.div.i.i
    go run ../scripts/divrem-rocket-predict.go ../results/rock/data/out.divu.i.i
    go run ../scripts/divrem-rocket-predict.go ../results/rock/data/out.rem.i.i
    go run ../scripts/divrem-rocket-predict.go ../results/rock/data/out.remu.i.i

    # Randomize floating-point operands (pick from normal and subnormal values) of fdiv.s and fdiv.d instructions.
    go run ../scripts/driver.go rand-rock-fdiv.s-n-n
    go run ../scripts/driver.go rand-rock-fdiv.s-n-s
    go run ../scripts/driver.go rand-rock-fdiv.s-s-n
    go run ../scripts/driver.go rand-rock-fdiv.s-s-s

    go run ../scripts/driver.go rand-rock-fdiv.d-n-n
    go run ../scripts/driver.go rand-rock-fdiv.d-n-s
    go run ../scripts/driver.go rand-rock-fdiv.d-s-n
    go run ../scripts/driver.go rand-rock-fdiv.d-s-s
    
    # Plot variations in instruction latencies (for all instructions)
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

|  **Instruction(s)** | **Slowdown** |
| ------------------- | ------------ |
| [`fdiv.d`](results/rock/plots/plot-fdiv.d.png) | 10.8x |
| [`div`](results/rock/plots/plot-div.png), [`divu`](results/rock/plots/plot-divu.png), [`rem`](results/rock/plots/plot-rem.png), [`remu`](results/rock/plots/plot-remu.png) | 10.3x |
| [`fdiv.s`](results/rock/plots/plot-fdiv.s.png) | 5.3x |
| [`mul`](results/rock/plots/plot-mul.png) | 1.7x |


#### Accuracy of Analytical Models of Timing

The latency from executing each of `div`, `divu`, `rem`, and `remu` instructions varies between 2 to 64 cycles, whereas the latency of `fdiv.s` instruction varies between 2 and 23 cycles.

The following table shows the mean and standard deviation of the observed error in cycle counts.

| **Instruction**       | **Mean err** | **stdev** | **Prediction** |
| --------------------- | ------------ | --------- | -------------- |
| `div`                 | 1.40         | 1.82      | [see code](scripts/divrem-rocket-predict.go) |
| `divu`                | 1.34         | 1.77      | [see code](scripts/divrem-rocket-predict.go) |
| `rem`                 | 1.55         | 1.90      | [see code](scripts/divrem-rocket-predict.go) |
| `remu`                | 1.44         | 1.55      | [see code](scripts/divrem-rocket-predict.go) |
| --------------------- | ------------ | --------- | -------------- |
| `fdiv.s` (norm, norm) | 0.42         | 0.45      | 23 cycles      |
| `fdiv.s` (norm, subn) | 0.40         | 0.45      | 23 cycles      |
| `fdiv.s` (subn, norm) | 0.43         | 0.45      | 23 cycles      |
| `fdiv.s` (subn, subn) | 0.44         | 0.45      | 23 cycles      |
| `fdiv.s` (all others) |    -         |    -      |  2 cycles      |
| --------------------- | ------------ | --------- | -------------- |
| `fdiv.d` (norm, norm) | 0.48         | 0.32      | 50 cycles      |
| `fdiv.d` (norm, subn) | 0.62         | 0.69      | 50 cycles      |
| `fdiv.d` (subn, norm) | 0.62         | 0.67      | 50 cycles      |
| `fdiv.d` (subn, subn) | 0.42         | 0.45      | 53 cycles      |
| `fdiv.d` (all others) |    -         |    -      |  2 cycles      |
| --------------------- | ------------ | --------- | -------------- |


### BOOM

Plots for boom chip are located [here](boom-results.md).

| **Instruction(s)** | **Slowdown** |
| ------------------ | ------------ |
| [`div`](results/boom/plots/plot-div.png), [`divu`](results/boom/plots/plot-divu.png), [`rem`](results/boom/plots/plot-rem.png), [`remu`](results/boom/plots/plot-remu.png) | 6.1x |
| [`fdiv.s`](results/boom/plots/plot-fdiv.s.png), [`fdiv.d`](results/boom/plots/plot-fdiv.d.png) | 3.3x |
