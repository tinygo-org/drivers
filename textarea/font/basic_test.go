package font

import (
	"image/color"
	"reflect"
	"testing"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/textarea/font/mocks"
)

func Test_basicFont_getCharBytes(t *testing.T) {
	type args struct {
		c rune
	}
	tests := []struct {
		name string
		font basicFont
		args args
		want []byte
	}{
		{"6 (54)", newBasicFont(font6x8), args{'6'}, []byte{0x00, 0x3c, 0x4a, 0x49, 0x49, 0x30}},
		{"~ (126)", newBasicFont(font6x8), args{'~'}, []byte{0x00, 0x10, 0x08, 0x10, 0x08, 0x0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.font.getCharBytes(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("basicFont.getCharBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_basicFont_PrintChar(t *testing.T) {
	type args struct {
		displayer drivers.Displayer
		x         int16
		y         int16
		char      rune
		c         color.RGBA
	}
	tests := []struct {
		name string
		font basicFont
		args args
		want [][]byte
	}{
		{"Font6x8 - 6", newBasicFont(font6x8), args{mocks.NewMockDisplayer(6, 8), 0, 0, '6', color.RGBA{0xff, 0x00, 0x00, 0xff}}, [][]byte{
			{0, 0, 0, 1, 1, 0},
			{0, 0, 1, 0, 0, 0},
			{0, 1, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0},
			{0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1},
			{0, 0, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0},
		}},
		{"Font6x8 - @", newBasicFont(font6x8), args{mocks.NewMockDisplayer(6, 8), 0, 0, '@', color.RGBA{0xff, 0x00, 0x00, 0xff}}, [][]byte{
			{0, 0, 1, 1, 1, 0},
			{0, 1, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 1},
			{0, 0, 1, 1, 0, 1},
			{0, 1, 0, 1, 1, 1},
			{0, 1, 0, 0, 0, 1},
			{0, 0, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0},
		}},
		{"Font6x8 - [", newBasicFont(font6x8), args{mocks.NewMockDisplayer(6, 8), 0, 0, '[', color.RGBA{0xff, 0x00, 0x00, 0xff}}, [][]byte{
			{0, 0, 1, 1, 1, 0},
			{0, 0, 1, 0, 0, 0},
			{0, 0, 1, 0, 0, 0},
			{0, 0, 1, 0, 0, 0},
			{0, 0, 1, 0, 0, 0},
			{0, 0, 1, 0, 0, 0},
			{0, 0, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0},
		}},
		{"font6x8 - b", newBasicFont(font6x8), args{mocks.NewMockDisplayer(6, 8), 0, 0, 'b', color.RGBA{0xff, 0x00, 0x00, 0xff}}, [][]byte{
			{0, 1, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 0},
			{0, 1, 0, 1, 1, 0},
			{0, 1, 1, 0, 0, 1},
			{0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1},
			{0, 1, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0},
		}},
		{"font6x8 - b - outo f bounds", newBasicFont(font6x8), args{mocks.NewMockDisplayer(6, 8), 4, 4, 'b', color.RGBA{0xff, 0x00, 0x00, 0xff}}, [][]byte{
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 1},
		}},
		{"font6x8 - ° (invalid)", newBasicFont(font6x8), args{mocks.NewMockDisplayer(6, 8), 0, 0, '°', color.RGBA{0xff, 0x00, 0x00, 0xff}}, [][]byte{
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.font.PrintChar(tt.args.displayer, tt.args.x, tt.args.y, tt.args.char, tt.args.c)
			if got := tt.args.displayer.(*mocks.MockDisplayer).GetPixels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("basicFont.PrintChar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_basicFont_Print(t *testing.T) {
	type args struct {
		displayer drivers.Displayer
		x         int16
		y         int16
		str       string
		c         color.RGBA
	}
	tests := []struct {
		name string
		font basicFont
		args args
		want [][]byte
	}{
		{"Font6x8 - AB", newBasicFont(font6x8), args{mocks.NewMockDisplayer(12, 8), 0, 0, "AB", color.RGBA{0x00, 0x00, 0xFF, 0xFF}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0},
			{0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"Font6x8 - 12", newBasicFont(font6x8), args{mocks.NewMockDisplayer(12, 8), 0, 0, "12", color.RGBA{0x00, 0x00, 0xFF, 0xFF}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 0, 1, 1, 1, 0},
			{0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 1},
			{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1},
			{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0},
			{0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0},
			{0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 1, 1},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"Font6x8 - °A - 2 bytes rune", newBasicFont(font6x8), args{mocks.NewMockDisplayer(12, 8), 0, 0, "°A", color.RGBA{0x00, 0x00, 0xFF, 0xFF}}, [][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1},
			{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.font.Print(tt.args.displayer, tt.args.x, tt.args.y, tt.args.str, tt.args.c)
			if got := tt.args.displayer.(*mocks.MockDisplayer).GetPixels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("basicFont.Print() = %v, want %v", got, tt.want)
			}
		})
	}
}
