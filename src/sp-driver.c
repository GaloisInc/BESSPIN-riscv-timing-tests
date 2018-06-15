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
            "mv             x20,    zero;"
            "li             x21,    10;"

            "li             x22, "  xstr(OP1) ";"
            "fmv.w.x        f22,    x22;"

            "li             x23, "  xstr(OP2) ";"
            "fmv.w.x        f23,    x23;"

            "csrr           x25,  mcycle;"
            "csrr           x26,  minstret;"

        "loop:"
            xstr(INST)  "   f24,    f22,  f23;"
            "addi           x20,    x20,  1;"
            "bleu           x20,    x21,  loop;"

            "csrr           x27,  mcycle;"
            "csrr           x28,  minstret;"
            "fence;"

            "subw           %[c], x27, x25;"
            "subw           %[i], x28, x26;"

        : [c] "=r" (cycles), [i] "=r" (instrs)
        :
        : "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28", "f22",
            "f23", "f24", "cc"
    );

    printf("instrs\t%8d\tcycles\t%8d\n", (int) instrs, (int) cycles);
    barrier(nc);

    exit(0);
}
