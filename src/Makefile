TYPE ?= int
OP1    ?= 0
OP2    ?= 0
N      ?= 500
INST  ?= mul
XLEN ?= 64
DRIVER = $(TYPE)-driver.c

ifeq ($(XLEN),64)
else
	ARCH 		= -march=rv32im
	ABI             = -mabi=ilp32
endif

INCLUDE=-Ienv -Icommon
CFLAGS=-mcmodel=medany -static -std=gnu99 -O2 -ffast-math -fno-common -fno-builtin-printf -g
CFLAGS+=-DINST=$(INST) -DOP1=$(OP1) -DOP2=$(OP2) -DN=$(N) -DXLEN=$(XLEN)
CFLAGS+=$(ABI) $(ARCH) -Wa,-Ilegacy
LFLAGS=-static -nostdlib -nostartfiles -lm -lc -lgcc -T common/test.ld
COMMON=$(wildcard common/*.c) $(wildcard common/*.S)
COMMON_OBJ := $(COMMON:.c=.o)

# TODO: Change the gcc-dependent parts

.PHONY: test
%.o: %.c
	riscv64-unknown-elf-gcc $(INCLUDE) $(CFLAGS) -c $< -o $@

test: $(COMMON_OBJ)
	riscv64-unknown-elf-gcc $(INCLUDE) $(CFLAGS) -o build/test.$(TYPE).$(INST).$(OP1).$(OP2) \
                          $(DRIVER) $(COMMON_OBJ) $(LFLAGS)

debug: $(COMMON_OBJ)
	riscv64-unknown-elf-gcc $(INCLUDE) $(CFLAGS) -o build/test \
                          test.c $(COMMON_OBJ) $(LFLAGS)
