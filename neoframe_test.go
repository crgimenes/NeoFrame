package main

import (
	"testing"
)

func TestRGBAstrToColor(t *testing.T) {
	tests := []struct {
		in         string
		r, g, b, a uint8
		wantErr    bool
	}{
		{in: "ff0000", r: 0xff, a: 0xff},
		{in: "#00ff00", g: 0xff, a: 0xff},
		{in: "0000ff80", b: 0xff, a: 0x80},
		{in: "#12345678", r: 0x12, g: 0x34, b: 0x56, a: 0x78},
		{in: "00000000"}, // the eraser color: fully transparent
		{in: "fff", wantErr: true},
		{in: "zzzzzz", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tc := range tests {
		r, g, b, a, err := RGBAstrToColor(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("RGBAstrToColor(%q): want error, got none", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("RGBAstrToColor(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if r != tc.r || g != tc.g || b != tc.b || a != tc.a {
			t.Errorf("RGBAstrToColor(%q) = %d,%d,%d,%d, want %d,%d,%d,%d",
				tc.in, r, g, b, a, tc.r, tc.g, tc.b, tc.a)
		}
	}
}
