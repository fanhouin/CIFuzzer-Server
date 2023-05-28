CC ?= gcc
CFLAGS = -Wall -g
BUILD_DIR = build
SRC_DIR = src

SRCS = $(wildcard $(SRC_DIR)/*.c)
OBJS = $(patsubst $(SRC_DIR)/%.c, $(BUILD_DIR)/%.o, $(SRCS))

EXECUTABLE = $(BUILD_DIR)/readfile

all: create_build_dir $(EXECUTABLE)

$(EXECUTABLE): $(OBJS)
	$(CC) $(CFLAGS) $^ -o $@

$(BUILD_DIR)/%.o: $(SRC_DIR)/%.c
	$(CC) $(CFLAGS) -c $< -o $@

create_build_dir:
	@mkdir -p $(BUILD_DIR)

clean:
	rm -f $(BUILD_DIR)/* 

.PHONY: all clean
