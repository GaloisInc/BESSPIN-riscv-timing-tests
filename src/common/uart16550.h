// See LICENSE for license details.

#ifndef _RISCV_16550_H
#define _RISCV_16550_H

#include <stdint.h>

extern volatile uint32_t* uart16550;

int uart16550_putchar(uint8_t ch);
int uart16550_getchar();
void query_uart16550(uintptr_t dtb);

#endif
