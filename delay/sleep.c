#include <stdint.h>
#include <stdbool.h>

// Loop the given times, where one loop takes four CPU cycles.
bool tinygo_drivers_sleep(uint32_t cycles) {
    // In this function, a [n] comment indicates the number of cycles an
    // instruction or a set of instructions take. This is typically 1 for most
    // arithmetic instructions, and a bit more for branches.
#if __ARM_ARCH_6M__ || __ARM_ARCH_7M__ || __ARM_ARCH_7EM__
    // Inline assembly for Cortex-M0/M0+/M3/M4/M7.
    // The Cortex-M0 (but not M0+) takes one more cycle, so is off by 12.5%.
    // Others should be basically cycle-accurate (with a slight overhead to
    // calculate the number of cycles). Unfortunately, there doesn't appear to
    // be a preprocessor macro to detect the Cortex-M0 specifically (although we
    // could rely on macros like NRF51).

    // Each loop takes 8 cycles (5 nops, 1 sub, and 2 for the branch).
    uint32_t loops = (cycles + 7) / 8;
    __asm__ __volatile__(
        "1:\n\t"
        "nop\n\t"               // [5] nops
        "nop\n\t"
        "nop\n\t"
        "nop\n\t"
        "nop\n\t"
        "subs %[loops], #1\n\t" // [1]
        "bne 1b"                // [1-4], at least 2 cycles if taken
        : [loops]"+r"(loops)
    );
    return true;
#elif __XTENSA__
    // Inline assembly for Xtensa.
    // I don't know exactly how many cycles a branch takes, so I've taken a
    // conservative guess and assume it takes only one cycle. In practice, it's
    // probably more than that.
    uint32_t loops = (cycles + 7) / 8;
    __asm__ __volatile__(
        "1:\n\t"
        "nop\n\t"                         // [6] nops
        "nop\n\t"
        "nop\n\t"
        "nop\n\t"
        "nop\n\t"
        "nop\n\t"
        "addi %[loops], %[loops], -1\n\t" // [1]
        "bnez %[loops], 1b"               // [1?]
        : [loops]"+r"(loops)
    );
    return true;
#else
    // Unknown architecture, so fall back to time.Sleep.
    return false;
#endif
}
