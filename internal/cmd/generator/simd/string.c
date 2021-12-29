#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <immintrin.h>

static const bool needEscape[256] = {
 // 0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0x00-0x0F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0x10-0x1F
	0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x20-0x2F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x30-0x3F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x40-0x4F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, // 0x50-0x5F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x60-0x6F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x70-0x7F
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0x80-0x8F
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0x90-0x9F
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0xA0-0xAF
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0xB0-0xBF
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0xC0-0xCF
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0xD0-0xDF
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0xE0-0xEF
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0xF0-0xFF
    };

uint64_t findHTMLEscapeIndex64(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const uint64_t space  = lsb * 0x20;
  static const uint64_t quote  = lsb * '"';
  static const uint64_t escape = lsb * '\\';
  static const uint64_t lt     = lsb * '<';
  static const uint64_t gt     = lsb * '>';
  static const uint64_t amp    = lsb * '&';

  char *sp = buf;
  size_t chunkLen = len / 8;
  int chunkIdx = 0;
  for (; chunkIdx < chunkLen; chunkIdx++) {
    uint64_t n    = *(uint64_t *)sp;
    uint64_t mask = n | (n - space) | ((n ^ quote) - lsb) | ((n ^ escape) - lsb) | ((n ^ lt) - lsb) | ((n ^ gt) - lsb) | ((n ^ amp) - lsb);
    uint64_t masked = mask & msb;
    if (masked != 0) {
      return __builtin_ctz(masked) / 8;
    }
    sp += 8;
  }
  return chunkIdx * 8;
}

uint64_t findHTMLEscapeIndex128(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const __m64 space  = (__m64)(lsb * 0x20);
  static const __m64 quote  = (__m64)(lsb * '"');
  static const __m64 escape = (__m64)(lsb * '\\');
  static const __m64 lt     = (__m64)(lsb * '<');
  static const __m64 gt     = (__m64)(lsb * '>');
  static const __m64 amp    = (__m64)(lsb * '&');

  __m128i msbV    = _mm_set_epi64((__m64)(msb), (__m64)(msb));
  __m128i lsbV    = _mm_set_epi64((__m64)(lsb), (__m64)(lsb));
  __m128i spaceV  = _mm_set_epi64(space, space);
  __m128i quoteV  = _mm_set_epi64(quote, quote);
  __m128i escapeV = _mm_set_epi64(escape, escape);
  __m128i ltV     = _mm_set_epi64(lt, lt);
  __m128i gtV     = _mm_set_epi64(gt, gt);
  __m128i ampV    = _mm_set_epi64(amp, amp);

  char *sp = buf;
  size_t chunkLen = len / 16;
  int chunkIdx = 0;
  for (; chunkIdx < chunkLen; chunkIdx++) {
    __m128i n       = _mm_loadu_si128((const void *)sp);
    __m128i spaceN  = _mm_sub_epi64(n, spaceV);
    __m128i quoteN  = _mm_sub_epi64(_mm_xor_si128(n, quoteV), lsbV);
    __m128i escapeN = _mm_sub_epi64(_mm_xor_si128(n, escapeV), lsbV);
    __m128i ltN     = _mm_sub_epi64(_mm_xor_si128(n, ltV), lsbV);
    __m128i gtN     = _mm_sub_epi64(_mm_xor_si128(n, gtV), lsbV);
    __m128i ampN    = _mm_sub_epi64(_mm_xor_si128(n, ampV), lsbV);

    __m128i mask = _mm_or_si128(_mm_or_si128(_mm_or_si128(_mm_or_si128(_mm_or_si128(_mm_or_si128(n, spaceN), quoteN), escapeN), ltN), gtN), ampN);
    int movemask = _mm_movemask_epi8(_mm_and_si128(mask, msbV));
    if (movemask != 0) {
      return __builtin_ctz(movemask);
    }
    sp += 16;
  }
  int idx = 16 * chunkLen;
  if (len - idx >= 8) {
    return idx + findHTMLEscapeIndex64(sp, len - idx);
  }
  return idx;
}

uint64_t findHTMLEscapeIndex256(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const __m64 space  = (__m64)(lsb * 0x20);
  static const __m64 quote  = (__m64)(lsb * '"');
  static const __m64 escape = (__m64)(lsb * '\\');
  static const __m64 lt     = (__m64)(lsb * '<');
  static const __m64 gt     = (__m64)(lsb * '>');
  static const __m64 amp    = (__m64)(lsb * '&');

  __m256i msbV    = _mm256_set1_epi64x(msb);
  __m256i lsbV    = _mm256_set1_epi64x(lsb);
  __m256i spaceV  = _mm256_set1_epi64x(space);
  __m256i quoteV  = _mm256_set1_epi64x(quote);
  __m256i escapeV = _mm256_set1_epi64x(escape);
  __m256i ltV     = _mm256_set1_epi64x(lt);
  __m256i gtV     = _mm256_set1_epi64x(gt);
  __m256i ampV    = _mm256_set1_epi64x(amp);

  char *sp = buf;
  size_t chunkLen = len / 32;
  int chunkIdx = 0;
  for (; chunkIdx < chunkLen; chunkIdx++) {
    __m256i n       = _mm256_loadu_si256((const void *)sp);
    __m256i spaceN  = _mm256_sub_epi64(n, spaceV);
    __m256i quoteN  = _mm256_sub_epi64(_mm256_xor_si256(n, quoteV), lsbV);
    __m256i escapeN = _mm256_sub_epi64(_mm256_xor_si256(n, escapeV), lsbV);
    __m256i ltN     = _mm256_sub_epi64(_mm256_xor_si256(n, ltV), lsbV);
    __m256i gtN     = _mm256_sub_epi64(_mm256_xor_si256(n, gtV), lsbV);
    __m256i ampN    = _mm256_sub_epi64(_mm256_xor_si256(n, ampV), lsbV);

    __m256i mask = _mm256_or_si256(_mm256_or_si256(_mm256_or_si256(_mm256_or_si256(_mm256_or_si256(_mm256_or_si256(n, spaceN), quoteN), escapeN), ltN), gtN), ampN);
    int movemask = _mm256_movemask_epi8(_mm256_and_si256(mask, msbV));
    if (movemask != 0) {
      return __builtin_ctz(movemask);
    }
    sp += 32;
  }
  int idx = 32 * chunkLen;
  int remainLen = len - idx;
  if (remainLen >= 16) {
    return idx + findHTMLEscapeIndex128(sp, remainLen);
  } else if (remainLen >= 8) {
    return idx + findHTMLEscapeIndex64(sp, remainLen);
  }
  return idx;
}

uint64_t findEscapeIndex64(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const uint64_t space  = lsb * 0x20;
  static const uint64_t quote  = lsb * '"';
  static const uint64_t escape = lsb * '\\';

  char *sp = buf;
  size_t chunkLen = len / 8;
  int chunkIdx = 0;
  for (; chunkIdx < chunkLen; chunkIdx++) {
    uint64_t n    = *(uint64_t *)sp;
    uint64_t mask = n | (n - space) | ((n ^ quote) - lsb) | ((n ^ escape) - lsb);
    uint64_t masked = mask & msb;
    if (masked != 0) {
      return __builtin_ctz(masked) / 8;
    }
    sp += 8;
  }
  int idx = 8 * chunkLen;
  bool *needEscape = needEscape;
  for ( ;idx < len; idx++) {
    if (needEscape[buf[idx]] != 0) {
      return idx;
    }
  }
  return len;
}

uint64_t findEscapeIndex128(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const __m64 space  = (__m64)(lsb * 0x20);
  static const __m64 quote  = (__m64)(lsb * '"');
  static const __m64 escape = (__m64)(lsb * '\\');

  __m128i msbV    = _mm_set_epi64((__m64)(msb), (__m64)(msb));
  __m128i lsbV    = _mm_set_epi64((__m64)(lsb), (__m64)(lsb));
  __m128i spaceV  = _mm_set_epi64(space, space);
  __m128i quoteV  = _mm_set_epi64(quote, quote);
  __m128i escapeV = _mm_set_epi64(escape, escape);

  char *sp = buf;
  size_t chunkLen = len / 16;
  int chunkIdx = 0;
  for (; chunkIdx < chunkLen; chunkIdx++) {
    __m128i n       = _mm_loadu_si128((const void *)sp);
    __m128i spaceN  = _mm_sub_epi64(n, spaceV);
    __m128i quoteN  = _mm_sub_epi64(_mm_xor_si128(n, quoteV), lsbV);
    __m128i escapeN = _mm_sub_epi64(_mm_xor_si128(n, escapeV), lsbV);

    __m128i mask = _mm_or_si128(_mm_or_si128(_mm_or_si128(n, spaceN), quoteN), escapeN);
    int movemask = _mm_movemask_epi8(_mm_and_si128(mask, msbV));
    if (movemask != 0) {
      return __builtin_ctz(movemask);
    }
    sp += 16;
  }
  int idx = 16 * chunkLen;
  int remainLen = len - idx;
  if (remainLen >= 8) {
    return idx + findEscapeIndex64(sp, remainLen);
  }
  bool *needEscape = needEscape;
  for (; idx < len; idx++) {
    if (needEscape[buf[idx]] != 0) {
      return idx;
    }
  }
  return len;
}

uint64_t findEscapeIndex256(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const __m64 space  = (__m64)(lsb * 0x20);
  static const __m64 quote  = (__m64)(lsb * '"');
  static const __m64 escape = (__m64)(lsb * '\\');

  __m256i msbV    = _mm256_set1_epi64x(msb);
  __m256i lsbV    = _mm256_set1_epi64x(lsb);
  __m256i spaceV  = _mm256_set1_epi64x(space);
  __m256i quoteV  = _mm256_set1_epi64x(quote);
  __m256i escapeV = _mm256_set1_epi64x(escape);

  char *sp = buf;
  size_t chunkLen = len / 32;
  int chunkIdx = 0;
  for (; chunkIdx < chunkLen; chunkIdx++) {
    __m256i n       = _mm256_loadu_si256((const void *)sp);
    __m256i spaceN  = _mm256_sub_epi64(n, spaceV);
    __m256i quoteN  = _mm256_sub_epi64(_mm256_xor_si256(n, quoteV), lsbV);
    __m256i escapeN = _mm256_sub_epi64(_mm256_xor_si256(n, escapeV), lsbV);

    __m256i mask = _mm256_or_si256(_mm256_or_si256(_mm256_or_si256(n, spaceN), quoteN), escapeN);
    int movemask = _mm256_movemask_epi8(_mm256_and_si256(mask, msbV));
    if (movemask != 0) {
      return __builtin_ctz(movemask) + chunkIdx * 32;
    }
    sp += 32;
  }
  int idx = 32 * chunkLen;
  int remainLen = len - idx;
  if (remainLen >= 16) {
    return idx + findEscapeIndex128(sp, remainLen);
  } else if (remainLen >= 8) {
    return idx + findEscapeIndex64(sp, remainLen);
  }
  bool *needEscape = needEscape;
  for (; idx < len; idx++) {
    if (needEscape[buf[idx]] != 0) {
      return idx;
    }
  }
  return len;
}
