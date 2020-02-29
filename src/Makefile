TYPE ?= int
OP1    ?= 0
OP2    ?= 0
INST  ?= mul
DRIVER = $(TYPE)-driver.c

INCLUDE=-Ienv -Icommon
CFLAGS=-mcmodel=medany -static -std=gnu99 -O2 -ffast-math -fno-common -fno-builtin-printf
CFLAGS+=-DINST=$(INST) -DOP1=$(OP1) -DOP2=$(OP2)
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
