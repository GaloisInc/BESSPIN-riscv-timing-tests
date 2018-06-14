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
            "li             x23, "  xstr(OP2) ";"

        "loop:"
            xstr(INST)  "   x24,  x22,  x23;"
            "addi           x20,  x20,  1;"
            "bleu           x20,  x21,  loop;"

        ::: "x20", "x21", "x22", "x23", "x24", "cc"
    );

    cycles += read_csr(mcycle);
    instrs += read_csr(minstret);

    asm volatile("fence");

    printf("instrs\t%8d\tcycles\t%8d\n", (int) instrs, (int) cycles);
    barrier(nc);

    exit(0);
}
