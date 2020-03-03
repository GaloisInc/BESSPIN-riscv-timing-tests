#include <stdio.h>
#include "util.h"

#define xstr(s) str(s)
#define str(s)  #s

__attribute__((aligned(256)))
void empty_loop(unsigned long* instr_count, unsigned long* cycle_count) {
    *instr_count = 0;
    *cycle_count = 0;

    asm volatile (
            "jal            x0, _init;"
            "xor            x24, zero, zero;"

        "_loop:"
            "fence;"
            "addi           x20,  x20,  1;"
            "bleu           x20,  x21,  _loop;"

            "jal            x0, _term;"

        "_init:"
            "mv             x20,    zero;"
            "li             x21,    10;"
            "li             x22, "  xstr(OP1) ";"
            "li             x23, "  xstr(OP2) ";"

            "csrr           x25,  mcycle;"
            "csrr           x26,  minstret;"

            "jal            x0, _loop;"

        "_term:"
            "csrr           x27,  mcycle;"
            "csrr           x28,  minstret;"
            "fence;"

            "subw           %[c], x27, x25;"
            "subw           %[i], x28, x26;"

        : [c] "=r" (*cycle_count), [i] "=r" (*instr_count)
        :
        : "x0", "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28",
            "cc"
    );
}

__attribute__((aligned(256)))
void busy_loop(unsigned long* instr_count, unsigned long* cycle_count) {
    *instr_count = 0;
    *cycle_count = 0;

    asm volatile (
            "jal            x0, init;"

        "loop:"
            xstr(INST)  "   x24,  x22,  x23;"
            "fence;"
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

        : [c] "=r" (*cycle_count), [i] "=r" (*instr_count)
        :
        : "x0", "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28",
            "cc"
    );
}

int main(int argc, char* argv[]) {
    unsigned long empty_cycles = 0, empty_instrs = 0;
    empty_loop(&empty_instrs, &empty_cycles);

    unsigned long busy_cycles = 0, busy_instrs = 0;
    busy_loop(&busy_instrs, &busy_cycles);

    unsigned long instrs = busy_instrs - empty_instrs;
    unsigned long cycles = busy_cycles - empty_cycles;

    printf("op1\t%llx\top2\t%llx\tinstrs\t%x\tcycles\t%x\n", (unsigned long long)OP1, (unsigned long long)OP2, instrs, cycles);

    return 0;
}
