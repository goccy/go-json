#include <stdio.h>
#include <stdint.h>
#include <string.h>

typedef struct GoString {
  char *buf;
  size_t len;
} GoString;
