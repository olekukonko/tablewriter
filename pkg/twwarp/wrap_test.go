// Copyright 2014 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Table Writer  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems

package twwarp

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter/pkg/twwidth"
	"github.com/olekukonko/tablewriter/tw"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

var (
	text    = "The quick brown fox jumps over the lazy dog."
	testDir = "./_data"
)

// checkEqual compares two values and fails the test if they are not equal
func checkEqual(t *testing.T, got, want interface{}, msgs ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("got:\n[%v]\nwant:\n[%v]\n", got, want))
		for _, v := range msgs {
			buf.WriteString(fmt.Sprint(v))
		}
		t.Errorf(buf.String())
	}
}

func TestWrap(t *testing.T) {
	exp := []string{
		"The", "quick", "brown", "fox",
		"jumps", "over", "the", "lazy", "dog."}

	got, _ := WrapString(text, 6)
	checkEqual(t, len(got), len(exp))
}

func TestWrapOneLine(t *testing.T) {
	exp := "The quick brown fox jumps over the lazy dog."
	words, _ := WrapString(text, 500)
	checkEqual(t, strings.Join(words, string(tw.Space)), exp)

}

func TestUnicode(t *testing.T) {
	input := "Česká řeřicha"
	var wordsUnicode []string
	if runewidth.IsEastAsian() {
		wordsUnicode, _ = WrapString(input, 14)
	} else {
		wordsUnicode, _ = WrapString(input, 13)
	}
	// input contains 13 (or 14 for CJK) runes, so it fits on one line.
	checkEqual(t, len(wordsUnicode), 1)
}

func TestDisplayWidth(t *testing.T) {
	input := "Česká řeřicha"
	want := 13
	if runewidth.IsEastAsian() {
		want = 14
	}
	if n := twwidth.Width(input); n != want {
		t.Errorf("Wants: %d Got: %d", want, n)
	}
	input = "\033[43;30m" + input + "\033[00m"
	checkEqual(t, twwidth.Width(input), want)

	input = "\033]8;;idea://open/?file=/path/somefile.php&line=12\033\\some URL\033]8;;\033\\"
	checkEqual(t, twwidth.Width(input), 8)

}

// WrapString was extremely memory greedy, it performed insane number of
// allocations for what it was doing. See BenchmarkWrapString for details.
func TestWrapStringAllocation(t *testing.T) {
	originalTextBytes, err := os.ReadFile(testDir + "/long-text.txt")
	if err != nil {
		t.Fatal(err)
	}
	originalText := string(originalTextBytes)

	wantWrappedBytes, err := os.ReadFile(testDir + "/long-text-wrapped.txt")
	if err != nil {
		t.Fatal(err)
	}
	wantWrappedText := string(wantWrappedBytes)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	heapAllocBefore := int64(ms.HeapAlloc / 1024 / 1024)

	// When
	gotLines, gotLim := WrapString(originalText, 80)

	// Then
	wantLim := 80
	if gotLim != wantLim {
		t.Errorf("Invalid limit: want=%d, got=%d", wantLim, gotLim)
	}

	gotWrappedText := strings.Join(gotLines, "\n")
	if gotWrappedText != wantWrappedText {
		t.Errorf("Invalid lines: want=\n%s\n got=\n%s", wantWrappedText, gotWrappedText)
	}

	runtime.ReadMemStats(&ms)
	heapAllocAfter := int64(ms.HeapAlloc / 1024 / 1024)
	heapAllocDelta := heapAllocAfter - heapAllocBefore
	if heapAllocDelta > 1 {
		t.Fatalf("heap allocation should not be greater than 1Mb, got=%dMb", heapAllocDelta)
	}
}

// Before optimization:
// BenchmarkWrapString-16    	       1	2490331031 ns/op	2535184104 B/op	50905550 allocs/op
// After optimization:
// BenchmarkWrapString-16    	    1652	    658098 ns/op	    230223 B/op	    5176 allocs/op
func BenchmarkWrapString(b *testing.B) {
	d, err := os.ReadFile(testDir + "/long-text.txt")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		WrapString(string(d), 128)
	}
}

func TestSplitWords(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out []string
	}{{
		in:  "",
		out: []string{},
	}, {
		in:  "a",
		out: []string{"a"},
	}, {
		in:  "a b",
		out: []string{"a", "b"},
	}, {
		in:  "   a   b   ",
		out: []string{"a", "b"},
	}, {
		in:  "\r\na\t\t \r\t b\r\n  ",
		out: []string{"a", "b"},
	}} {
		t.Run(tt.in, func(t *testing.T) {
			got := SplitWords(tt.in)
			if !reflect.DeepEqual(tt.out, got) {
				t.Errorf("want=%s, got=%s", tt.out, got)
			}
		})
	}
}

func TestWrapString(t *testing.T) {
	want := []string{"ああああああああああああああああああああああああ", "あああああああ"}
	got, _ := WrapString("ああああああああああああああああああああああああ あああああああ", 55)
	checkEqual(t, got, want)
}
