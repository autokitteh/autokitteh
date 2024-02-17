package sheets

import (
	"fmt"
	"testing"
)

func TestA1RangeToSubstrings(t *testing.T) {
	tests := []struct {
		a1Range   string
		wantSheet string
		wantFrom  string
		wantTo    string
		wantErr   bool
	}{
		{
			a1Range:   "Sheet1!A1:B2",
			wantSheet: "Sheet1",
			wantFrom:  "A1",
			wantTo:    "B2",
		},
		{
			a1Range:   "Sheet1!A:A",
			wantSheet: "Sheet1",
			wantFrom:  "A",
			wantTo:    "A",
		},
		{
			a1Range:   "Sheet1!1:2",
			wantSheet: "Sheet1",
			wantFrom:  "1",
			wantTo:    "2",
		},
		{
			a1Range:   "Sheet1!A5:A",
			wantSheet: "Sheet1",
			wantFrom:  "A5",
			wantTo:    "A",
		},
		{
			a1Range:   "A1:B2",
			wantSheet: "",
			wantFrom:  "A1",
			wantTo:    "B2",
		},
		{
			a1Range:   "A:B",
			wantSheet: "",
			wantFrom:  "A",
			wantTo:    "B",
		},
		{
			a1Range:   "1:2",
			wantSheet: "",
			wantFrom:  "1",
			wantTo:    "2",
		},
		{
			a1Range:   "A5:A",
			wantSheet: "",
			wantFrom:  "A5",
			wantTo:    "A",
		},
		{
			a1Range:   "Sheet1",
			wantSheet: "Sheet1",
			wantFrom:  "",
			wantTo:    "",
		},
		{
			a1Range:   "'My Custom Sheet'!A:A",
			wantSheet: "My Custom Sheet",
			wantFrom:  "A",
			wantTo:    "A",
		},
		{
			a1Range:   "'My Custom Sheet'",
			wantSheet: "My Custom Sheet",
			wantFrom:  "",
			wantTo:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.a1Range, func(t *testing.T) {
			got, got1, got2, err := a1RangeToSubstrings(tt.a1Range)
			if (err != nil) != tt.wantErr {
				t.Errorf("a1RangeToSubstrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantSheet {
				t.Errorf("a1RangeToSubstrings() got sheet = %v, want %v", got, tt.wantSheet)
			}
			if got1 != tt.wantFrom {
				t.Errorf("a1RangeToSubstrings() got from = %v, want %v", got1, tt.wantFrom)
			}
			if got2 != tt.wantTo {
				t.Errorf("a1RangeToSubstrings() got to = %v, want %v", got2, tt.wantTo)
			}
		})
	}
}

func TestCellToIndexes(t *testing.T) {
	tests := []struct {
		cell    string
		want    int64
		want1   int64
		wantErr bool
	}{
		{
			cell:  "A1",
			want:  0,
			want1: 0,
		},
		{
			cell:  "A",
			want:  0,
			want1: -1,
		},
		{
			cell:  "1",
			want:  -1,
			want1: 0,
		},
		{
			cell:  "B123",
			want:  1,
			want1: 122,
		},
		{
			cell:  "Z123",
			want:  25,
			want1: 122,
		},
		{
			cell:  "AA123",
			want:  26,
			want1: 122,
		},
		{
			cell:  "AB123",
			want:  27,
			want1: 122,
		},
	}
	for _, tt := range tests {
		t.Run(tt.cell, func(t *testing.T) {
			got, got1, err := cellToIndexes(tt.cell)
			if (err != nil) != tt.wantErr {
				t.Errorf("cellToIndexes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cellToIndexes() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("cellToIndexes() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestIndexToBase26(t *testing.T) {
	tests := []struct {
		i    int
		want string
	}{
		{
			i:    0,
			want: "A",
		},
		{
			i:    1,
			want: "B",
		},
		{
			i:    25,
			want: "Z",
		},
		{
			i:    26,
			want: "AA",
		},
		{
			i:    27,
			want: "AB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := indexToBase26(tt.i)
			if got != tt.want {
				t.Errorf("indexToBase26() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		color     int
		wantAlpha float64
		wantRed   float64
		wantGreen float64
		wantBlue  float64
		wantErr   bool
	}{
		{
			color:     0x000000,
			wantAlpha: 1.0,
			wantRed:   0.0,
			wantGreen: 0.0,
			wantBlue:  0.0,
		},
		{
			color:     0xFFFFFF,
			wantAlpha: 1.0,
			wantRed:   1.0,
			wantGreen: 1.0,
			wantBlue:  1.0,
		},
		{
			color:     0xFFFFFFFF,
			wantAlpha: 1.0,
			wantRed:   1.0,
			wantGreen: 1.0,
			wantBlue:  1.0,
		},
		{
			color:   0x100000000,
			wantErr: true,
		},
		{
			color:   -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%08X", tt.color), func(t *testing.T) {
			got, got1, got2, got3, err := hexToRGB(tt.color)
			if (err != nil) != tt.wantErr {
				t.Errorf("hexToRGB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantAlpha {
				t.Errorf("hexToRGB() got alpha = %v, want %v", got, tt.wantAlpha)
			}
			if got1 != tt.wantRed {
				t.Errorf("hexToRGB() got red = %v, want %v", got1, tt.wantRed)
			}
			if got2 != tt.wantGreen {
				t.Errorf("hexToRGB() got green = %v, want %v", got2, tt.wantGreen)
			}
			if got3 != tt.wantBlue {
				t.Errorf("hexToRGB() got blue = %v, want %v", got3, tt.wantBlue)
			}
		})
	}
}
