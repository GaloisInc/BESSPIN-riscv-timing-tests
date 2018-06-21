#include <assert.h>
#include <stdlib.h>
#include <stdio.h>

#include "util.h"

#define xstr(s) str(s)
#define str(s)  #s

void thread_entry(int cid, int nc)
{
    size_t cycles = 0;
    size_t instrs = 0;

    asm volatile (
            "jal            x0, init;"

        "loop:"
            xstr(INST)  "   x24,  x22,  x23;"
            "addi           x20,  x20,  1;"
            "bleu           x20,  x21,  loop;"

            "jal            x0, term;"

        "init:"
            "mv             x20,    zero;"
            "li             x21,    10;"
            "li             x22, "  xstr(OP1) ";"
            "li             x23, "  xstr(OP2) ";"

            "csrr           x25,  mcycle;"
            "csrr           x26,  minstret;"

            "jal            x0, loop;"

        "term:"
            "csrr           x27,  mcycle;"
            "csrr           x28,  minstret;"
            "fence;"

            "subw           %[c], x27, x25;"
            "subw           %[i], x28, x26;"

        : [c] "=r" (cycles), [i] "=r" (instrs)
        :
        : "x0", "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28",
            "cc"
    );

    printf("instrs\t%8d\tcycles\t%8d\n", (int) instrs, (int) cycles);
    barrier(nc);

    exit(0);
}
