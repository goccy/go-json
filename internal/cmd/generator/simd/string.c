#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <immintrin.h>

uint64_t findEscapeIndex64(char *buf, int len) {
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
      return __builtin_ctz(masked);
    }
    sp += 8;
  }
  return 8 * chunkLen;
}

uint64_t findEscapeIndex128(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const __m64 space  = (__m64)(lsb * 0x20);
  static const __m64 quote  = (__m64)(lsb * '"');
  static const __m64 escape = (__m64)(lsb * '\\');
  static const __m64 lt     = (__m64)(lsb * '<');
  static const __m64 gt     = (__m64)(lsb * '>');
  static const __m64 amp    = (__m64)(lsb * '&');

  __m128i zeroV   = _mm_setzero_si128();
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
    return findEscapeIndex64(sp, len - idx);
  }
  return idx;
}

uint64_t findEscapeIndex256(char *buf, int len) {
  static const uint64_t lsb = 0x0101010101010101;
  static const uint64_t msb = 0x8080808080808080;

  static const __m64 space  = (__m64)(lsb * 0x20);
  static const __m64 quote  = (__m64)(lsb * '"');
  static const __m64 escape = (__m64)(lsb * '\\');
  static const __m64 lt     = (__m64)(lsb * '<');
  static const __m64 gt     = (__m64)(lsb * '>');
  static const __m64 amp    = (__m64)(lsb * '&');

  __m256i zeroV   = _mm256_setzero_si256();
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
    return findEscapeIndex128(sp, remainLen);
  } else if (remainLen >= 8) {
    return findEscapeIndex64(sp, remainLen);
  }
  return idx;
}
