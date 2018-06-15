## RISCV Instruction Latency Tests

This code measures the latency of various RISV instructions from the Base ISA and from the M, F, and D extensions on the Rocket chip (https://github.com/freechipsproject/rocket-chip) using a Verilator-based simulation.


### Prerequisites

  - Go (with the `gitlab.com/ashay/bagpipe` package)
  - R (with `ggplot` and `reshape2` packages)
  - Scripts assume `riscv64-unknown-elf-gcc` is in `PATH`, and that the Rocket chip is built in `${HOME}/src/rocket-chip/emulator`.


### How to Gather Data and Plot Results

    cd src
    
    # Specify targets to build
    go run ../scripts/driver.go build-int build-sp build-dp
    
    # Run tests for a specific instructions
    go run ../scripts/driver.go run-fdiv.s
    go run ../scripts/driver.go run-add
    go run ../scripts/driver.go run-div
    
    # Plot data (for all instructions)
    cd ../scripts
    R --no-save < plot.R


### Note about Debug Interrupts

If you plan to tweak the loop count in `src/*-driver.c`, note that running the test for too long may cause the (occassional) debug interrupts from the simulation to perturb the results.  In particular, if you see a wide variation in the instruction count (despite only changing the operand values), then the debug interrupts are suspect.  See https://github.com/freechipsproject/rocket-chip/issues/1495 for details and how to know whether the debug interrupt occured.


## Results

### Interesting Cases
<img src = "results/plots/plot-div.png" width = "400px" />
<img src = "results/plots/plot-mul.png" width = "400px" />
<img src = "results/plots/plot-fdiv.s.png" width = "400px" />

### Not-So-Interesting Cases
<img src = "results/plots/plot-add.png" width = "400px" />
<img src = "results/plots/plot-and.png" width = "400px" />
<img src = "results/plots/plot-divu.png" width = "400px" />
<img src = "results/plots/plot-mulh.png" width = "400px" />
<img src = "results/plots/plot-mulhsu.png" width = "400px" />
<img src = "results/plots/plot-mulhu.png" width = "400px" />
<img src = "results/plots/plot-or.png" width = "400px" />
<img src = "results/plots/plot-rem.png" width = "400px" />
<img src = "results/plots/plot-remu.png" width = "400px" />
<img src = "results/plots/plot-sll.png" width = "400px" />
<img src = "results/plots/plot-slt.png" width = "400px" />
<img src = "results/plots/plot-sltu.png" width = "400px" />
<img src = "results/plots/plot-sra.png" width = "400px" />
<img src = "results/plots/plot-srl.png" width = "400px" />
<img src = "results/plots/plot-sub.png" width = "400px" />                                                                                                                        <img src = "results/plots/plot-xor.png" width = "400px" />
<img src = "results/plots/plot-fmax.s.png" width = "400px" />
<img src = "results/plots/plot-fmax.d.png" width = "400px" />
<img src = "results/plots/plot-fmin.s.png" width = "400px" />
<img src = "results/plots/plot-fmin.d.png" width = "400px" />
<img src = "results/plots/plot-fsgnjx.s.png" width = "400px" />
<img src = "results/plots/plot-fsgnjx.d.png" width = "400px" />
<img src = "results/plots/plot-fsgnjn.s.png" width = "400px" />
<img src = "results/plots/plot-fsgnjn.d.png" width = "400px" />
<img src = "results/plots/plot-fsgnj.s.png" width = "400px" />
<img src = "results/plots/plot-fsgnj.d.png" width = "400px" />
<img src = "results/plots/plot-fdiv.d.png" width = "400px" />
<img src = "results/plots/plot-fmul.s.png" width = "400px" />
<img src = "results/plots/plot-fmul.d.png" width = "400px" />
<img src = "results/plots/plot-fsub.s.png" width = "400px" />
<img src = "results/plots/plot-fsub.d.png" width = "400px" />
<img src = "results/plots/plot-fadd.s.png" width = "400px" />
<img src = "results/plots/plot-fadd.d.png" width = "400px" />
