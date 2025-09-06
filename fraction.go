package drivers

import (
	"errors"
	"math"
)

const (
	signBitmask           = 0x80000000 // first bit
	expBitmask            = 0xFF       // next 8 bits, mask after shift is simpler to understand
	mantissaBits          = 23         // remaining 23 bits
	mantissaBitsMask      = 0x007FFFFF // mask sign and exponent bits
	bias                  = 127        // bias according IEEE754 specification for "binary32"
	maxAbsExponent        = 30         // to keep safe below MaxInt32 without using "2^31-1" for calculation
	maxAbsExponentEncoded = maxAbsExponent + bias
)

// Float32Fractions calculates the numerator and denominator from the encoded bits of the float32 number, which
// are placed as follows: s|hhh hhhh h|nnn nnnn nnnn nnnn nnnn nnnn
//
// sign S: 1 bit s, exponent-value E: 8 bits h (count bits r=8), mantis-value M: 23 bits n (count bits p=23)
//
// S = 1 for positive and -1 for negative values
//
// creating of real mantis "m" (1 =< m < 2) from M (0 =< (M/2^p) < 1; 0 =< M =< 2^p - 1):
// 24 bits mantis = 1nnn nnnn nnnn nnnn nnnn nnnn
// m = 1 + M/2^p; 2^p = 2^23 = 8388608
//
// creating of real exponent "e" (-126 =< e =< 127)
// e = E - B; B = 127
//
// f = m * 2^e
//
// Target values can be calculated as follows:
// num = (-1)^S * [2^k + M*2^(k-23)]; den = 2^(k-e)
//
// For fast calculation use the shift operation: y = a*2^x => y = (a<<x) ; "a" can be 1
//
// If internally int32 is a concern, k must be chosen dynamically as follows:
// k=30+e; if k>30; k=30 => means for f > 1 (e >= 0) => k=30, otherwise (e<0) => k=0..29
//
// example 18.4: exp=4, M=1258291
// k=34 => k=30
// den = 2^(30-4) = 67108864 = 0x04000000
// num = 2^30 + M*2^(30-23) = 2^30 + M*2^7 = 1073741824 + 161061248 = 1234803072 0x49999980
// f = num/den = 18.3999996185
//
// example 0.05: exp=-5, M=5033164
// k=25
// den = 2^(25--5) = 2^30 = 1073741824 = 0x40000000
// num = 2^25 + M*2^(25-23) = 2^25 + M*2^2 = 33554432 + 20132656 = 53687088 = 0x03333330
// f = num/den = 0.04999999702
//
// If a bigger range for the internal calculation is used and encoded values until just before return are used, some
// calculations and checks can be simplified/dropped. This means especially:
// * k can be constant 30
// * a constant can be used for "2^30" instead of calculation from k
// * a constant can be used for "2^(30-23)" instead of calculation from k
// * no differentiation is needed at beginning regarding the sign of e
//
// valid range for int32 for e: -30 <= e <= 30
// k=30 => means for f > 1 (e >= 0) => (k-e)=0..30, otherwise (e<0) => (k-e)=31..60
//
// if e < 0, "num" can and "den" will exceed the int32 limit and needs to be adjusted as follows:
// * let den = 2^30
// * adjust "num" according the missing accuracy of "den" now, which means dividing by 2^(-exp)
//
// example 18.4: exp=4, M=1258291
// k=34 => k=30
// den = 2^(30-4) = 67108864 = 0x04000000
// num = 2^30 + M*2^(30-23) = 2^30 + M*2^7 = 1073741824 + 161061248 = 1234803072 = 0x49999980
// f = num/den = 18.3999996185; num and den are guaranteed to be in the int32 range
//
// example 0.05: exp=-5, M=5033164
// k=25
// den64 = 2^(30--5) = 2^35 = 34359738368 = 0x0800000000
// num64 = 2^30 + M*2^(30-23) = 2^30 + M*2^7 = 1073741824 + 644244992 = 1717986816 = 0x66666600
// f64 = num64/den64 = 0.04999999702
// den = 2^30 = 1073741824
// num = num64/2^--5 = num64/2^5 = 1717986816/32 = 53687088 = 0x03333330
// f = num/den = 0.04999999702
//
// special cases:
// E=0, M=0: f=0 => num=0; den=1
// E=0, M>0 (very small numbers): |f| =< 2^-127; f=1/MaxInt32 (error=e^-95); f=0 (error=2^-127) => num=0; den=1
// please note: with this we loose the sign, but get higher accuracy
// E=255 (all bits set), M=0: +/-Inf; but for int32 numerator, +Inf is for e=31, means E=158, -Inf is for e=
// E=255 (all bits set), M>0: NaN
//
// See:
// https://de.wikipedia.org/wiki/IEEE_754
//
// Very good accuracy can be reached, similar to calculating with "math/big.Rat", but ~20 times faster:
// nrf52840: 1.526µs-4.577µs
//
// Considered other options: see function in test file
//
//nolint:cyclop // ok here
func Float32Fractions(f float32) (int32, int32, error) {
	const (
		k                   = maxAbsExponent
		twoPowK             = (1 << k)         // 2^30
		kMinusMantisssaBits = k - mantissaBits // 7
	)

	bits := math.Float32bits(f) // uint32

	encMantNumerator := bits & mantissaBitsMask                 // encrypted mantissa "M"
	encExpDenominator := int64(bits>>mantissaBits) & expBitmask // encrypted exponent "E"
	negative := bits&signBitmask > 0                            // use the sign bit as boolean

	/*
		fmt.Printf("%f > bits: %032b (%x) exp: %08b (%d) man: %023b (%d) neg: %t\n", f, bits, bits, encExpDenominator,
			encExpDenominator, encMantNumerator, encMantNumerator, negative)
	*/

	if encMantNumerator == 0 && encExpDenominator == 0 {
		// f was zero
		return 0, 1, nil
	}

	if encExpDenominator > maxAbsExponentEncoded {
		if negative {
			if encExpDenominator == maxAbsExponentEncoded+1 && encMantNumerator == 0 {
				// special case for "s"=-1; "m" exactly "1" and "e" exactly "31"
				return math.MinInt32, 1, nil
			}

			return math.MinInt32, 1, errors.New("input value exceeds -int32 range")
		}

		return math.MaxInt32, 1, errors.New("input value exceeds +int32 range")
	}

	if encExpDenominator < -maxAbsExponentEncoded {
		// treat this is zero, not 1/MaxInt32 (see comment)
		return 0, 1, nil
	}

	// num = (-1)^S * [2^k + M*2^(k-23)]
	encMantNumerator = twoPowK + encMantNumerator<<kMinusMantisssaBits // now decrypted mantissa without sign

	// start converting to 32 bit integer
	if encExpDenominator < bias {
		// small values f < 1, denominator is set to fixed 2^30 and numerator needs to be decreased
		encMantNumerator = encMantNumerator >> (bias - encExpDenominator)
		if negative {
			return -int32(encMantNumerator), twoPowK, nil //nolint:gosec // mantissa is max. 2^30 + 2^30 at this point
		}

		return int32(encMantNumerator), twoPowK, nil //nolint:gosec // mantissa is max. 2^30 + 2^30 at this point
	}

	// values f > 1, numerator and denominator can be used like expected
	// den = 2^(k-e)
	encExpDenominator = 1 << (k + bias - encExpDenominator) // now denominator

	if negative {
		//nolint:gosec // mantissa is max. 2^30 + 2^30 at this point, exponent is 2^(k-e), k=30
		return -int32(encMantNumerator), int32(encExpDenominator), nil
	}

	//nolint:gosec // mantissa is max. 2^30 + 2^30 at this point, exponent is 2^(k-e), k=30
	return int32(encMantNumerator), int32(encExpDenominator), nil
}
