package twwidth

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

var benchmarkStrings = map[string]string{
	"SimpleASCII":       "hello world, this is a test string.",
	"ASCIIWithANSI":     "\033[31mhello\033[0m \033[34mworld\033[0m, this is \033[1ma\033[0m test string.",
	"EastAsian":         "こんにちは世界、これはテスト文字列です。",
	"EastAsianWithANSI": "\033[32mこんにちは\033[0m \033[35m世界\033[0m、これは\033[4mテスト\033[0m文字列です。",
	"LongSimpleASCII":   strings.Repeat("abcdefghijklmnopqrstuvwxyz ", 20),
	"LongASCIIWithANSI": strings.Repeat("\033[31ma\033[32mb\033[33mc\033[34md\033[35me\033[36mf\033[0m ", 50),
}

func BenchmarkWidthFunction(b *testing.B) {
	eastAsianSettings := []bool{false, true}

	for name, str := range benchmarkStrings {
		for _, eaSetting := range eastAsianSettings {
			SetEastAsian(eaSetting)

			b.Run(fmt.Sprintf("%s_EA%v_NoCache", name, eaSetting), func(b *testing.B) {
				b.SetBytes(int64(len(str)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = WidthNoCache(str)
				}
			})

			b.Run(fmt.Sprintf("%s_EA%v_CacheMiss", name, eaSetting), func(b *testing.B) {
				b.SetBytes(int64(len(str)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = Width(str + strconv.Itoa(i))
				}
			})
			resetGlobalCache()

			b.Run(fmt.Sprintf("%s_EA%v_CacheHit", name, eaSetting), func(b *testing.B) {
				b.SetBytes(int64(len(str)))
				b.ReportAllocs()
				b.ResetTimer()
				if b.N > 0 {
					_ = Width(str)
				}
				for i := 1; i < b.N; i++ {
					_ = Width(str)
				}
			})
			resetGlobalCache()
		}
	}
}
