CC = riscv64-unknown-elf-gcc

CFLAGS = -I include -mcmodel=medany -std=gnu99 -O2
LDFLAGS = -static -nostartfiles -T test.ld

OBJECTS = syscalls.o crt.o

all:    sll  srl  sra  add  sub  xor  and  or  slt  sltu  mul  mulh  mulhsu  \
        mulhu  div  divu  rem  remu

crt.o:  crt.S
	$(CC) $(CFLAGS) -c $< -o $@

sll:    $(OBJECTS)  driver.c
	$(CC) -DINST=sll $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

srl:    $(OBJECTS)  driver.c
	$(CC) -DINST=srl $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

sra:    $(OBJECTS)  driver.c
	$(CC) -DINST=sra $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

add:    $(OBJECTS)  driver.c
	$(CC) -DINST=add $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

sub:    $(OBJECTS)  driver.c
	$(CC) -DINST=sub $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

xor:    $(OBJECTS)  driver.c
	$(CC) -DINST=xor $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

and:    $(OBJECTS)  driver.c
	$(CC) -DINST=and $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

or:     $(OBJECTS)  driver.c
	$(CC) -DINST=or $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

slt:    $(OBJECTS)  driver.c
	$(CC) -DINST=slt $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

sltu:   $(OBJECTS)  driver.c
	$(CC) -DINST=sltu $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

mul:    $(OBJECTS)  driver.c
	$(CC) -DINST=mul $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

mulh:   $(OBJECTS)  driver.c
	$(CC) -DINST=mulh $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

mulhsu: $(OBJECTS)  driver.c
	$(CC) -DINST=mulhsu $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

mulhu:  $(OBJECTS)  driver.c
	$(CC) -DINST=mulhu $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

div:   $(OBJECTS)  driver.c
	$(CC) -DINST=div $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

divu:  $(OBJECTS)  driver.c
	$(CC) -DINST=divu $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

rem:   $(OBJECTS)  driver.c
	$(CC) -DINST=rem $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

remu:  $(OBJECTS)  driver.c
	$(CC) -DINST=remu $(CFLAGS) -c driver.c -o driver.o
	$(CC) driver.o $(OBJECTS) -O2 $(LDFLAGS) -o $@

clean:
	$(RM)   $(OBJECTS)  sll  srl  sra  add  sub  xor  and  or  slt  sltu  mul \
            mulh  mulhsu  mulhu  div  divu  rem  remu  driver.o
