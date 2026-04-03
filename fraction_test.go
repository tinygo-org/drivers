//nolint:funlen // ok for tests
package drivers

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type result struct {
	num int32
	den int32
	e   float64
}

type test struct {
	startNum     int32
	endNum       int32
	startDen     int32
	endDen       int32
	stepCountNum int32
	stepCountDen int32
	wantErr      error
	epsilon      map[string]float64 // depends on test function
	run          map[string]bool    // depends on test function
	useGot64     map[string]bool    // depends on test function
}

type float32FractionsFunc func(f float32) (int32, int32, error)

const (
	bitsShift = "fract_bits_shift"
	ipB32     = "fract_int_part_base32"
	ipB10     = "fract_int_part_base10"
	floatMul  = "fract_float_multiple"
	bigRat    = "fract_bigrat"
	bigRatStr = "fract_bigrat_string_split"
	useStr    = "fract_string"
)

const (
	eSmall = 1e-37             // smallest
	e32    = 4.65661287525e-10 // 1/32 bit
	e24    = 5.96046483281e-08 // 1/24 bit
	e23    = 1.19209303762e-07 // 1/23 bit
)

func TestFloat32Fractions(t *testing.T) {
	// this tests only the values around maximum and minimum
	tests := map[string]struct {
		input   float32
		wantNum int32
		wantErr string
	}{
		"maximum_float": {
			input:   float32(2147483583.9999999), // float32 conversion error is -64
			wantNum: 2147483520,
		},
		"error_too_big": {
			input:   float32(2147483584.0), // "2147483584" is already interpreted as "2147483648"
			wantNum: math.MaxInt32,
			wantErr: "input value exceeds +int32 range",
		},
		"minimum_float": {
			input:   float32(-2147483776.0), // float32 conversion error is -128
			wantNum: -2147483648,
		},
		"error_too_small": {
			input:   float32(-2147483776.0000001),
			wantNum: math.MinInt32,
			wantErr: "input value exceeds -int32 range",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			// act
			// assert
			n, d, err := Float32Fractions(tc.input)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantNum, n)
			assert.Equal(t, int32(1), d)
		})
	}
}

func Test_float32Fractions(t *testing.T) {
	t.Skip() // due to very long run
	// this tests different approaches with many different inputs of int32/int32
	testFuncs := map[string]float32FractionsFunc{
		bitsShift: Float32Fractions,
		ipB32:     float32FractionsIntPartBase32,
		ipB10:     float32FractionsIntPartBase10,
		floatMul:  float32FractionsManyMult,
		bigRat:    float32FractionsBigRat,
		bigRatStr: float32FractionsBigRatStringSplit,
		useStr:    float32FractionsString,
	}

	tests := map[string]test{
		"single_just_for_debug": {
			startNum:     17459216,
			endNum:       17459216,
			startDen:     2147483584,
			endDen:       2147483584,
			stepCountNum: 1,
			stepCountDen: 1,
			epsilon: map[string]float64{
				useStr: e24,
			},
			run: map[string]bool{
				bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
		},
		"plus_inf": {
			startNum:     1,
			endNum:       1,
			startDen:     0,
			endDen:       0,
			stepCountNum: 1,
			stepCountDen: 1,
			wantErr:      errors.New("input value exceeds +int32 range"),
			run: map[string]bool{
				bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
		},
		"minus_inf": {
			startNum:     -1,
			endNum:       -1,
			startDen:     0,
			endDen:       0,
			stepCountNum: 1,
			stepCountDen: 1,
			wantErr:      errors.New("input value exceeds -int32 range"),
			run: map[string]bool{
				bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
		},
		"neg_fast": {
			startNum:     math.MinInt32,
			endNum:       0,
			startDen:     1,
			endDen:       math.MaxInt32,
			stepCountNum: 9,
			stepCountDen: 15,
			epsilon: map[string]float64{
				floatMul: 1e-6,
			},
			run: map[string]bool{
				bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
			useGot64: map[string]bool{
				useStr: true,
			},
		},
		"pos": {
			startNum:     0,
			endNum:       math.MaxInt32,
			startDen:     1,
			endDen:       math.MaxInt32,
			stepCountNum: 123,
			stepCountDen: 12321,
			epsilon: map[string]float64{
				// ipB10: epsilon=-0.0001220703125 for 1763380816/871471
				// floatMul: epsilon=-0.03066958300769329 for 17459216/2081593243
				ipB10: 0.00013, floatMul: 0.031, useStr: e23,
			},
			run: map[string]bool{
				bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
			useGot64: map[string]bool{
				ipB10: true, useStr: true,
			},
		},
		"pos_high_numbers": {
			startNum:     math.MaxInt32 - 100,
			endNum:       math.MaxInt32,
			startDen:     math.MaxInt32 - 100,
			endDen:       math.MaxInt32,
			stepCountNum: 100,
			stepCountDen: 100,
			epsilon: map[string]float64{
				// ip: 1e-10, bigRatStr: 1e-10, floatMul: 1e-10,
				floatMul: 1e-7,
			},
			run: map[string]bool{
				bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
		},
		"pos_very_long_run": {
			startNum:     0,
			endNum:       math.MaxInt32,
			startDen:     1,
			endDen:       math.MaxInt32,
			stepCountNum: 123213,
			stepCountDen: 123213,
			epsilon:      map[string]float64{},
			run:          map[string]bool{
				// bitsShift: true, ipB32: true, ipB10: true, floatMul: true, bigRat: true, bigRatStr: true, useStr: true, // all
				// bitsShift: true, ipB32: true
			},
		},
	}
	for tfName, testFunc := range testFuncs {
		for tcName, tc := range tests {
			name := tfName + "_" + tcName
			t.Run(name, func(t *testing.T) {
				if run, ok := tc.run[tfName]; !ok || !run {
					t.Skipf("test case %s skipped intentionally", name)
				}
				epsilon := eSmall
				if e, ok := tc.epsilon[tfName]; ok {
					epsilon = e
				}
				var useGot64 bool
				if u, ok := tc.useGot64[tfName]; ok && u {
					useGot64 = true
				}

				doFractionTest(t, name, tc, testFunc, useGot64, epsilon)
			})
		}
	}
}

func doFractionTest(t *testing.T, name string, tc test, testFunc float32FractionsFunc, useGot64 bool, epsilon float64) {
	t.Helper()

	const msgTemplate = "run %s (want: '%s'(%d/%d), %s: '%s'(%d/%d)"

	// arrange
	numDelta := (tc.endNum - tc.startNum) / tc.stepCountNum
	denDelta := (tc.endDen - tc.startDen) / tc.stepCountDen
	var results []result

	den := tc.startDen
	stepDen := tc.stepCountDen

	start := time.Now()
	for ; stepDen > 0; stepDen-- {
		num := tc.startNum
		stepNum := tc.stepCountNum
		for ; stepNum > 0; stepNum-- {
			want := float32(num) / float32(den)
			// act
			n, d, err := testFunc(want)
			// assert
			// "float32(float64(num)/float64(den))" is especially needed for functions with base10, e.g. for 34918432/174295
			// which results to '200.341' with (20034099/100000) for "float32(n) / float32(d)"
			got := float32(n) / float32(d)
			got64 := float32(float64(n) / float64(d))
			if math.Abs(float64(want-got)) > math.Abs(float64(want-got64)) {
				if !useGot64 {
					msg := fmt.Sprintf(msgTemplate, name, strconv.FormatFloat(float64(want), 'f', -1, 32), num, den,
						"got64", strconv.FormatFloat(float64(got64), 'f', -1, 32), n, d)
					fmt.Printf(">>> got64 automatically used for %s\n", msg)
				}
				got = got64
			}
			require.Equal(t, tc.wantErr, err, "for run %s", name)
			if tc.wantErr == nil {
				msg := fmt.Sprintf(msgTemplate, name, strconv.FormatFloat(float64(want), 'f', -1, 32), num, den,
					"got", strconv.FormatFloat(float64(got), 'f', -1, 32), n, d)
				require.InDelta(t, want, got, epsilon, msg)
			}
			results = append(results, result{num: num, den: den, e: math.Abs(float64(got - want))})

			num = num + numDelta
		}
		den = den + denDelta
	}
	elapse := time.Since(start) / time.Duration(tc.stepCountNum) / time.Duration(tc.stepCountDen)
	sort.Slice(results, func(i, j int) bool {
		return results[i].e < results[j].e
	})

	fmin, fcmin := count(results, 0)
	fmax, fcmax := count(results, len(results)-1)
	fmt.Printf("%s (%d x %s): %d x %v - %d x %v\n", name, len(results), elapse, fcmin, fmin, fcmax, fmax)
}

func count(s []result, idx int) (result, int) {
	res := s[idx]
	var count int
	for _, v := range s {
		if v.e == res.e {
			count++
		}
	}

	return res, count
}

// float32FractionsIntPartBase32 calculates num/den by using float32FractionsIntPartBaseMax(f, MaxInt32).
func float32FractionsIntPartBase32(f float32) (int32, int32, error) {
	return float32FractionsIntPartBaseMax(f, math.MaxInt32)
}

// float32FractionsIntPartBase10 calculates num/den by using float32FractionsIntPartBaseMax(f, 1000000000). The results
// for accuracy are poor for special combinations and are equally regarding the speed.
func float32FractionsIntPartBase10(f float32) (int32, int32, error) {
	const baseMax = 1000000000 // 10 digits, epsilon=-0.0001220703125 for 1763380816/871471
	// const baseMax = 2000000000 // 10.5 digits, epsilon=1.52587890625e-05 for 87296080/348589

	return float32FractionsIntPartBaseMax(f, baseMax)
}

// float32FractionsIntPartBaseMax calculates the denominator "den" from the integer part of a given floating point
// number "f" and returns this denominator together with the resulting nominator "nom", so "f" can be reconstructed by
// "f = num / den".
// All values are in the range MaxInt32: 2147483647, MinInt32: -2147483648. If the given "f" exceeds this range, it is
// not possible anymore to represent "f" with "f = num / 1" and an error will be returned with the nearest values.
// As an exception we define that "den" is always positive, so negative numbers "f" leads always to negative "num".
//
// Sign:
// "abs(MaxInt32) > abs (MinInt32)", the sign can be applied to "den" or "nom", but we already defined "den" as positive
//
// Used formulas:
// For "abs(f)=af < 1" applies the biggest denominator: "den = math.MaxInt32" and "num = af * den". For "af > 0" this
// can be written more generalized when split integer part "ip" from fractional part "fp" with "af = ip + fp":
// "den = math.MaxInt32/(ip + 1)"; "num = af * den"
// very good accuracy can be reached, similar to calculating with "math/big.Rat", but 2-15 times faster:
// max. epsilon = 1.9073486328125e-06 in test for "17459216/697177" on arm64, but better for this example on MCU,
// nrf52840 12.207µs-14.496µs (independent of used base)
func float32FractionsIntPartBaseMax(f float32, baseMax int32) (int32, int32, error) {
	ip, den, err := float32FractionsPreCheck(f)
	if den == 1 || err != nil {
		return ip, den, err
	}

	if f < 0 {
		ip = -ip
	}

	den = baseMax / (ip + 1)
	if den == 0 {
		den = 1
	}

	return int32(float32(den) * f), den, nil
}

// float32FractionsManyMult takes the common approach for dissolving by count of decimals, but do not use strings.
// This comes with the costs of 10 times float multiplication in maximum. This is like the iterative version of
// float32FractionsIntPartBase10.
// nrf52840: 31.28µs-48.828µs (depends on count of iteration)
func float32FractionsManyMult(f float32) (int32, int32, error) {
	if ip, den, err := float32FractionsPreCheck(f); den == 1 || err != nil {
		return ip, den, err
	}

	const accuracy = 10
	den := float32(1)
	num := f

	for i := 0; i < accuracy; i++ {
		if num*den == float32(int32(num*den)) {
			break
		}

		den *= 10.0
	}

	return int32(num * den), int32(den), nil
}

// float32FractionsBigRat uses the big/math go library for splitting the given value in fractions.
// This function produce most accurate results, but is 5-8 times slower than float32FractionsIntPartBaseMax(MaxInt32)
// and ~20 times slower than the fast function using bits.
//
// nrf52840: 74.768µs-106.049µs
func float32FractionsBigRat(f float32) (int32, int32, error) {
	if ip, den, err := float32FractionsPreCheck(f); den == 1 || err != nil {
		return ip, den, err
	}

	r := new(big.Rat).SetFloat64(float64(f))
	d := r.Denom().Int64() // Denom returns the denominator of x; it is always > 0.
	if d > math.MaxInt32 {
		d = math.MaxInt32
	}

	n := f * float32(d)
	if n > math.MaxInt32 {
		println("n ex+", n)
		return math.MaxInt32, int32(float32(math.MaxInt32) / f), nil
	} else if n < math.MinInt32 {
		println("n ex-", n)
		return math.MinInt32, int32(float32(math.MinInt32) / f), nil
	}

	//nolint:gosec // ok here
	return int32(n), int32(d), nil
}

// float32FractionsBigRatStringSplit takes the simplest approach using "big/Rat". The accuracy of course is very good,
// but is ~65 times slower than the fast function using bits.
//
// nrf52840: 225.067µs-334.167µs
func float32FractionsBigRatStringSplit(f float32) (int32, int32, error) {
	if ip, den, err := float32FractionsPreCheck(f); den == 1 || err != nil {
		return ip, den, err
	}

	rat := new(big.Rat).SetFloat64(float64(f))
	s := rat.RatString()
	parts := strings.Split(s, "/")
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	if len(parts) == 1 {
		//nolint:gosec // ok here
		return int32(num), 1, nil
	}

	den, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	//nolint:gosec // ok here
	return int32(num), int32(den), nil
}

// float32FractionsString uses the common method for dissolving by count of decimals. The accuracy is very good, but
// is 15 times slower than the fast function using bits.
//
// max. epsilon = 1.9073486328125e-06 in test for "17459216/697177" on arm64, but better for this example on MCU,
//
// nrf52840: 55.695µs-83.16µs
func float32FractionsString(f float32) (int32, int32, error) {
	// reduce to 9 + 1 digits, according 2147483647 (-2147483648) of 32 bit integer
	const accuracy = 9
	ip, den, err := float32FractionsPreCheck(f)
	if den == 1 || err != nil {
		return ip, den, err
	}

	strDecimal := strconv.FormatFloat(float64(f), 'f', -1, 32)
	parts := strings.Split(strDecimal, ".")

	if len(parts) < 2 {
		// fmt.Errorf("integer detected (%e), pre check was wrong (%d/%d)", f, ip, den)
		// fmt.Printf("integer detected (%e), pre check was wrong (%d/%d)\n", f, ip, den)
		return int32(f), 1, nil
	}

	numeratorStr := parts[0] + parts[1] // "f" without decimal point
	// now for  0.0756: part[0]="0",  part[1]="0756", numeratorStr="00756"
	// now for -0.0756: part[0]="-0", part[1]="0756", numeratorStr="-00756"

	var sign int
	if f < 0 {
		sign = 1
	}

	if len(numeratorStr) > (accuracy + sign) {
		// although when catch-ed by "strconv.FormatFloat" this occurs, but very rare
		parts[1] = parts[1][:(accuracy+sign)-len(parts[0])]
		numeratorStr = parts[0] + parts[1]
	}

	num, err := strconv.Atoi(numeratorStr) // ignores leading zeros automatically and considers the sign
	if err != nil {
		return int32(f), 1, errors.New(numeratorStr + " for numerator is no integer")
	}

	den = 1
	for i := 0; i < len(parts[1]); i++ {
		den *= 10
	}
	// for 0.0756: num=756; for -0.0756: num=-756, den=10000 (len(parts[1])=4)

	//nolint:gosec // ok here
	return int32(num), den, nil
}

func float32FractionsPreCheck(f float32) (int32, int32, error) {
	if f > math.MaxInt32 {
		return math.MaxInt32, 1, errors.New("input value exceeds +int32 range")
	}

	if f < math.MinInt32 {
		return math.MinInt32, 1, errors.New("input value exceeds -int32 range")
	}

	integerPart := int32(f)
	if float32(integerPart) == f {
		// float is an integer
		return integerPart, 1, nil
	}

	return integerPart, 0, nil
}
