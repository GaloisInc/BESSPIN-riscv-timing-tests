# RISC-V Instruction Latency Tests #

This code measures the latency of various RISC-V instructions from the basic
ISA and from the M, F, and D extensions on SSITH processors


## Prerequisites ##

This code uses a driver script written in Python and some post-processing scripts
written in R.

 - Python 3.7 or higher

 - [`riscv-tools`](https://github.com/riscv/riscv-tools) in `PATH`.


## Setting Up ##

Fetch the relevant dependent packages for the R scripts.

    # Fetches code for interpolating and plotting results.
    $ R --no-save < scripts/init.R

### How to Gather Data and Plot Results

    # Gather measurements for a specific instruction and processor.

    # Here, we are sweeping through interspersed operands of the 'mul'
    # instruction on the 'rocket' chip.  This will take a while to complete.

    $ python scripts/driver.go sweep mul chisel_p2 results/out.mul.integer.integer
    test complete, results in results/out.mul.integer.integer

    # For giggles, plot the measurements (of the *integer* instruction) using a heat map.
    $ Rscript scripts/plot-int.R results/out.mul.integer.integer

    # Generate interpolated results in /tmp/pred.out.
    $ Rscript scripts/interpolate.R results/data/out.mul.integer.integer > /tmp/pred.out

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
