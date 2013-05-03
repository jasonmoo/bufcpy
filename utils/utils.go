package utils

import (
	"strconv"
	"strings"
)

const divisor float64 = float64(1024)

var (
	SizeNames []string = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
)

func FormatHumanReadableSize(b int64, precision int) string {

	i, negator := 0, int64(1)
	if b < 0 {
		negator = int64(-1)
		b *= negator
	}
	n := float64(b)
	for ; n >= divisor; i, n = i+1, n/divisor {
	}
	return strconv.FormatFloat(n*float64(negator), 'f', precision, 64) + SizeNames[i]

}

func ParseHumanReadableSize(s string) (int64, error) {

	number, m := s[:], int64(1)
	if i := strings.IndexAny(s, "bBkKmMgGtTpPeE"); i > 0 {
		switch s[i] {
		case 'k','K': m <<= 10; break
		case 'm','M': m <<= 20; break
		case 'g','G': m <<= 30; break
		case 't','T': m <<= 40; break
		case 'p','P': m <<= 50; break
		case 'e','E': m <<= 60
		}
		number = s[:i]
	}
	num, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return int64(0), err
	}
	num = num * float64(m)
	return int64(num), err
}

func ZeroBytes(b []byte) {
	for i, l := 0, len(b); i < l; i++ {
		b[i] = byte(0)
	}
}

func FillBytes(b []byte) {
	for i, l := 0, len(b); i < l; i++ {
		b[i] = byte(i)
	}
}

func DiffBytes(a, b []byte) int {
	ct, l := 0, 0
	if len(a) < len(b) {
		l = len(a)
	} else {
		l = len(b)
	}
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			ct++
		}
	}
	return ct
}
