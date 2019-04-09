# RISC-V Instruction Latency Tests #

This code measures the latency of various RISC-V instructions from the basic
ISA and from the M, F, and D extensions on the
[Rocket](https://github.com/freechipsproject/rocket-chip) and
[BOOM](https://github.com/ucb-bar/riscv-boom) using a Verilator-based
simulation.


## Prerequisites ##

This code uses a driver script written in Go and some post-processing scripts
written in R.

 - Recent versions of [golang](https://golang.org/) and
   [R](https://www.r-project.org/).

 - [`riscv-tools`](https://github.com/riscv/riscv-tools) in `PATH`.


## Setting Up ##

Fetch the relevant dependent packages for the Go and R scripts.

    # Fetches code for file system and multi-threading APIs.
    $ go get gitlab.com/ashay/bagpipe

    # Fetches code for interpolating and plotting results.
    $ R --no-save < scripts/init.R


### How to Gather Data and Plot Results

    cd src

    # Adjust the maximum number of concurrent measurements based on the number
    # of processor cores.
    $ grep "var MAX_THREAD_COUNT" ../scripts/driver.go
    var MAX_THREAD_COUNT = 4

    # Gather measurements for a specific instruction and processor.

    # Here, we are sweeping through interspersed operands of the 'mul'
    # instruction on the 'rocket' chip.  This will take a while to complete.

    $ go run ../scripts/driver.go sweep --instr mul --arch rocket
    test complete, results in ../results/rocket/data/out.mul.integer.integer

    # For giggles, plot the measurements (of the *integer* instruction) using a heat map.
    $ Rscript ../scripts/plot-int.R ../results/rocket/data/out.mul.integer.integer

    # Generate interpolated results in /tmp/pred.out.
    $ Rscript ../scripts/interpolate.R ../results/rocket/data/out.mul.integer.integer > /tmp/pred.out

    # Validate the results of the interpolation.
    $ go run ../scripts/driver.go validate --instr mul --arch rocket --prediction-file /tmp/pred.out
    95th percentile error: 0.84 cycle(s), 99th percentile error: 0.91 cycle(s), maximum error: 0.91 cycle(s)

    # The results above indicate that the interpolation script was able to
    # predict the cycle count for previously-unobserved operands for the 'mul'
    # instruction on the 'rocket' chip within at most 0.91 cycles, with a
    # 95th-percentile error of 0.84 cycles.


### Note about Debug Interrupts

If you plan to tweak the loop count in `src/*-driver.c`, note that running the
test for too long may cause the (occassional) debug interrupts from the
simulation to perturb the results.  In particular, if you see a wide variation
in the instruction count (despite only changing the operand values), then the
debug interrupts are suspect.  See
https://github.com/freechipsproject/rocket-chip/issues/1495 for details and how
to know whether the debug interrupt occured.


## Results

The following tables show the slowdown in instruction execution based on the
choice of operand values.  For integer instructions, the operands range from
0x0 to 0x7fff\_ffff\_ffff\_ffff, where each increment represents the lowermost
bits (in multiples of 4) being set to 1s.  For floating-point instructions, the
operands are from the set { zero, normal, subnormal, +inf, -inf, and
not-a-number }.

### Rocket

Plots for rocket chip are located [here](rocket-results.md).

|  **Instruction(s)** | **Slowdown** |
| :------------------ | -----------: |
| [`fdiv.d`](results/rocket/plots/plot-fdiv.d.png) | 10.8x |
| [`div`](results/rocket/plots/plot-div.png), [`divu`](results/rocket/plots/plot-divu.png), [`rem`](results/rocket/plots/plot-rem.png), [`remu`](results/rocket/plots/plot-remu.png) | 10.3x |
| [`fdiv.s`](results/rocket/plots/plot-fdiv.s.png) | 5.3x |
| [`mul`](results/rocket/plots/plot-mul.png) | 1.7x |


### BOOM

Plots for boom chip are located [here](boom-results.md).

| **Instruction(s)** | **Slowdown** |
| :----------------- | -----------: |
| [`div`](results/boom/plots/plot-div.png), [`divu`](results/boom/plots/plot-divu.png), [`rem`](results/boom/plots/plot-rem.png), [`remu`](results/boom/plots/plot-remu.png) | 6.1x |
| [`fdiv.s`](results/boom/plots/plot-fdiv.s.png), [`fdiv.d`](results/boom/plots/plot-fdiv.d.png) | 3.3x |


## Implementation of Interpolation Script ##

The interpolation code uses Delaunay Triangulation to form an irregular grid of
measurements for different operand values, before computing the Barycentric
coordinates for the requested (i.e. previously-unobserved) point within one of
the triangles computed in the previous step. As the third and final step, the
value at the requested point in the 2D space is interpolated based on the
values of the three points of the containing triangle.


## Debugging ##

Here are the steps to reproduce some of the output produced by
`scripts/driver.go`.  Let's assume that we are interested in benchmarking the
`mul` instruction on the 'Rocket' core.

    # First, pick two operands (say 0xa0 and 0x1f) to pass to the `mul`
    # instruction and bake them into the object file.  The build command is
    # located in src/build-cmd.txt.

    $ riscv64-unknown-elf-gcc -Ienv -Icommon -mcmodel=medany -static -std=gnu99 -O2 -ffast-math -fno-common -fno-builtin-printf -DINST=mul -DOP1=0xa0 -DOP2=0x1f -o test int-driver.c common/syscalls.c common/crt.S -static -nostdlib -nostartfiles -lm -lgcc -T common/test.ld

    # Run the resulting binary (test) on the Rocket core.

    $ ${HOME}/src/rocket-chip/emulator/emulator-freechips.rocketchip.system-DefaultConfig -s 0 ./test
    This emulator compiled with JTAG Remote Bitbang client. To enable, use +jtag_rbb_enable=1.
    Listening on port 43751
    instrs  000b    cycles  0007

    # The above output indicates that a total of 0xb (i.e. 11) instructions
    # were executed in a total of 0x07 cycles for a mean latency of 0.6 cycles.


## Known Issues ##

R, and hence the interpolation scripts, do not support arbitrary-precision
integer arithmetic.  This places restrictions on the the interpolation script's
ability to create a Delaunay mesh, effectively producing garbage results.
Consequently, the `driver.go` script that collects measurements limits the
maximum operand value to 2<sup>56</sup> - 1.

A possible workaround is to convert the (integer) operand values to double-precision
floating-point numbers, but such numbers lose precision in the low bits for
large values, resulting in aliasing, and thus incorrect results.  Furthermore,
converting from double-precision numbers to integer hexadecimal numbers is a
risky and error-prone step.
