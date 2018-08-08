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

 - Rocket and/or BOOM emulators in `${HOME}/src/rocket-chip` and
   `${HOME}/src/boom-template` directories respectively.


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
| [`fdiv.d`](results/rock/plots/plot-fdiv.d.png) | 10.8x |
| [`div`](results/rock/plots/plot-div.png), [`divu`](results/rock/plots/plot-divu.png), [`rem`](results/rock/plots/plot-rem.png), [`remu`](results/rock/plots/plot-remu.png) | 10.3x |
| [`fdiv.s`](results/rock/plots/plot-fdiv.s.png) | 5.3x |
| [`mul`](results/rock/plots/plot-mul.png) | 1.7x |


#### Accuracy of Analytical Models of Timing

The latency from executing each of `div`, `divu`, `rem`, and `remu`
instructions varies between 2 to 64 cycles, whereas the latency of `fdiv.s`
instruction varies between 2 and 23 cycles.

The following table shows the mean and standard deviation of the observed error
in cycle counts.

| **Instruction**       | **Mean err** | **stdev** | **Prediction** |
| :-------------------- | -----------: | --------: | -------------: |
| `mul`                 | 0.02         | 0.03      | [see `mul` code](scripts/mul-rocket-predict.go)    |
| `div`                 | 1.40         | 1.82      | [see `div` code](scripts/divrem-rocket-predict.go) |
| `divu`                | 1.34         | 1.77      | [see `div` code](scripts/divrem-rocket-predict.go) |
| `rem`                 | 1.55         | 1.90      | [see `div` code](scripts/divrem-rocket-predict.go) |
| `remu`                | 1.44         | 1.55      | [see `div` code](scripts/divrem-rocket-predict.go) |
| `fdiv.s` (norm, norm) | 0.42         | 0.45      | 23 cycles      |
| `fdiv.s` (norm, subn) | 0.40         | 0.45      | 23 cycles      |
| `fdiv.s` (subn, norm) | 0.43         | 0.45      | 23 cycles      |
| `fdiv.s` (subn, subn) | 0.44         | 0.45      | 23 cycles      |
| `fdiv.s` (all others) |    -         |    -      |  2 cycles      |
| `fdiv.d` (norm, norm) | 0.48         | 0.32      | 50 cycles      |
| `fdiv.d` (norm, subn) | 0.62         | 0.69      | 50 cycles      |
| `fdiv.d` (subn, norm) | 0.62         | 0.67      | 50 cycles      |
| `fdiv.d` (subn, subn) | 0.42         | 0.45      | 53 cycles      |
| `fdiv.d` (all others) |    -         |    -      |  2 cycles      |


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
