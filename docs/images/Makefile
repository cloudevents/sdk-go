.PHONY: all

all: $(patsubst %.dot,%.png,$(wildcard *.dot))

%.png: %.dot
	dot -Tpng $< -o $@