CFLAGS=$(shell pkg-config --cflags cairo)
LIBS=$(shell pkg-config --libs cairo)
SRCS=$(wildcard *.c)
OBJS=$(patsubst %.c,%.o,$(SRCS))

all: logo-500x500.svg logo-70x70.png

logo-500x500.svg: gen
	./gen svg 500x500 logo-500x500.svg

logo-70x70.png: gen
	./gen png 450x45 logo-450x45.png

gen: $(OBJS)
	$(CC) -o gen $< $(LIBS)

%.o: %.c
	$(CC) -c $(CFLAGS) $<

clean:
	rm -rf *.o gen logo.svg

