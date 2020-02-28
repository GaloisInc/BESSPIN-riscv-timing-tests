// See LICENSE for license details.

#include <string.h>
#include "uart16550.h"
#include "fdt.h"

volatile uint32_t* uart16550 = 0;

#define UART_REG_QUEUE     0
#define UART_REG_LINESTAT  5
#define UART_REG_STATUS_RX 0x01
#define UART_REG_STATUS_TX 0x20

static inline uint32_t bswap(uint32_t x)
{
  uint32_t y = (x & 0x00FF00FF) <<  8 | (x & 0xFF00FF00) >>  8;
  uint32_t z = (y & 0x0000FFFF) << 16 | (y & 0xFFFF0000) >> 16;
  return z;
}

int uart16550_putchar(uint8_t ch)
{
  while ((uart16550[UART_REG_LINESTAT] & UART_REG_STATUS_TX) == 0);
  uart16550[UART_REG_QUEUE] = ch;
  return 0;
}

int uart16550_getchar()
{
  while ((uart16550[UART_REG_LINESTAT] & UART_REG_STATUS_RX) == 0);
  return uart16550[UART_REG_QUEUE];
}

struct uart16550_scan
{
  int compat;
  uint64_t reg;
  uint32_t *speed;
  uint32_t *freq;
};

static void uart16550_open(const struct fdt_scan_node *node, void *extra)
{
  struct uart16550_scan *scan = (struct uart16550_scan *)extra;
  memset(scan, 0, sizeof(*scan));
  scan->speed = 0;
  scan->freq  = 0;
}

static void uart16550_prop(const struct fdt_scan_prop *prop, void *extra)
{
  struct uart16550_scan *scan = (struct uart16550_scan *)extra;
  if (!strcmp(prop->name, "compatible") && !strcmp((const char*)prop->value, "ns16550a")) {
    scan->compat = 1;
  } else if (!strcmp(prop->name, "reg")) {
    fdt_get_address(prop->node->parent, prop->value, &scan->reg);
  } else if (!strcmp(prop->name, "current-speed")) {
    scan->speed = prop->value;
  } else if (!strcmp(prop->name, "clock-frequency")) {
    scan->freq = prop->value;
  }
}

static void uart16550_done(const struct fdt_scan_node *node, void *extra)
{
  struct uart16550_scan *scan = (struct uart16550_scan *)extra;
  if (!scan->compat || !scan->reg || uart16550) return;

  uart16550 = (void*)(uintptr_t)scan->reg;
  // http://wiki.osdev.org/Serial_Ports
  uart16550[1] = 0x00;    // Disable all interrupts
  uart16550[3] = 0x80;    // Enable DLAB (set baud rate divisor)

  if (scan->speed && scan->freq) {
    uart16550[0] = bswap(scan->freq[0]) / (16u * bswap(scan->speed[0])) % 0x100u; // Divisor lo byte
    uart16550[1] = bswap(scan->freq[0]) / (16u * bswap(scan->speed[0])) >> 8;     // Divisor hi byte
  } else {
    uart16550[0] = 0x2d;    // Set divisor to 3 (lo byte) 38400 baud
    uart16550[1] = 0x00;    //                  (hi byte)
  }
  uart16550[3] = 0x03;    // 8 bits, no parity, one stop bit
  uart16550[2] = 0xC7;    // Enable FIFO, clear them, with 14-byte threshold
}

void query_uart16550(uintptr_t fdt)
{
  struct fdt_cb cb;
  struct uart16550_scan scan;

  memset(&cb, 0, sizeof(cb));
  cb.open = uart16550_open;
  cb.prop = uart16550_prop;
  cb.done = uart16550_done;
  cb.extra = &scan;

  fdt_scan(fdt, &cb);
  // Temporarily hardcode the UART address if the device tree does not have one
  if (!uart16550) {
    uart16550 = (void*)(uintptr_t)0x62300000;
  }
//  if (scan.speed && scan.freq) {
//    printf("\r\n");
//    printf("UART 16550 configured with options: baud = %d | freq = %d\r\n",
//      bswap(scan.speed[0]), bswap(scan.freq[0]));
//  }
}
