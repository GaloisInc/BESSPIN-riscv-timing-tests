#include <stdio.h>
#include "util.h"

#define xstr(s) str(s)
#define str(s)  #s

__attribute__((aligned(256)))
__attribute__((optimize("unroll-loops")))
void empty_loop(unsigned long* instr_count, unsigned long* cycle_count) {
    *instr_count = 0;
    *cycle_count = 0;

    asm volatile (
            "jal            x0, _init;"
    );

    for (int x = 0; x < NOOP_COUNT; x++) {
        asm volatile (
            "xor            x24, zero, zero;"
        );
    }

    asm volatile (
        "_init:"
            "mv             x20,    zero;"
            "li             x21,    10;"

            "csrr           x25,  mcycle;"
            "csrr           x26,  minstret;"

        "_loop:"
            "fence;"
            "addi           x20,  x20,  1;"
            "bleu           x20,  x21,  _loop;"

            "jal            x0, _term;"

        "_term:"
            "csrr           x27,  mcycle;"
            "csrr           x28,  minstret;"
            "fence;"

            "subw           %[c], x27, x25;"
            "subw           %[i], x28, x26;"

        : [c] "=r" (*cycle_count), [i] "=r" (*instr_count)
        :
        : "x0", "x20", "x21", "x24", "x25", "x26", "x27", "x28", "cc"
    );
}

int main(int argc, char* argv[]) {
    unsigned long cycles = 0, instrs = 0;
    empty_loop(&instrs, &cycles);

    printf("instrs\t%04x\tcycles\t%04x\n", instrs, cycles);

    return 0;
}
