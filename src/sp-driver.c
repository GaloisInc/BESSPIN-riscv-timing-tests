#include <assert.h>
#include <stdlib.h>
#include <stdio.h>

#include "util.h"

#define xstr(s) str(s)
#define str(s)  #s

void thread_entry(int cid, int nc)
{
    size_t cycles = 0 - read_csr(mcycle);
    size_t instrs = 0 - read_csr(minstret);

    asm volatile (
            "mv             x20,    zero;"
            "li             x21,    1024;"

            "li             x22, "  xstr(OP1) ";"
            "fmv.w.x        f22,    x22;"

            "li             x23, "  xstr(OP2) ";"
            "fmv.w.x        f23,    x23;"

        "loop:"
            xstr(INST)  "   f24,    f22,  f23;"
            "addi           x20,    x20,  1;"
            "bleu           x20,    x21,  loop;"

        ::: "x20", "x21", "x22", "x23", "x24", "f22", "f23", "f24", "cc"
    );

    cycles += read_csr(mcycle);
    instrs += read_csr(minstret);

    asm volatile("fence");

    printf("instrs\t%8d\tcycles\t%8d\n", (int) instrs, (int) cycles);
    barrier(nc);

    exit(0);
}
