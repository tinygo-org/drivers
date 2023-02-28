package textarea

import (
	"image/color"
	"reflect"
	"testing"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/textarea/font"
	"tinygo.org/x/drivers/textarea/font/mocks"
)

func TestNew(t *testing.T) {
	displayer := mocks.NewMockDisplayer(6, 8)
	ft := font.NewFont6x8()
	type args struct {
		displayer drivers.Displayer
		ft        font.Font
	}
	tests := []struct {
		name string
		args args
		want *TextArea
	}{
		{"default", args{displayer, ft}, &TextArea{
			Wrap:      false,
			displayer: displayer,
			ft:        ft,
			cursorX:   0,
			cursorY:   0,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.displayer, tt.args.ft); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextArea_Print(t *testing.T) {
	type fields struct {
		wrap      bool
		displayer drivers.Displayer
		ft        font.Font
		cursorX   int16
		cursorY   int16
	}
	type args struct {
		str string
		c   color.RGBA
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]byte
	}{
		{"A", fields{false, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{"A", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"A\nB", fields{false, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{"A\nB", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"A\rB", fields{false, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{"A\rB", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"ABC - no wrap", fields{false, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{"ABC", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0},
			{0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"ABC - wrap", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{"ABC", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0},
			{0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := &TextArea{
				Wrap:      tt.fields.wrap,
				displayer: tt.fields.displayer,
				ft:        tt.fields.ft,
				cursorX:   tt.fields.cursorX,
				cursorY:   tt.fields.cursorY,
			}
			text.Print(tt.args.str, tt.args.c)
			if got := tt.fields.displayer.(*mocks.MockDisplayer).GetPixels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextArea.Print() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextArea_Reset(t *testing.T) {
	type fields struct {
		Wrap      bool
		displayer drivers.Displayer
		ft        font.Font
		cursorX   int16
		cursorY   int16
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"default", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 7, 9}},
		{"null cursor", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := &TextArea{
				Wrap:      tt.fields.Wrap,
				displayer: tt.fields.displayer,
				ft:        tt.fields.ft,
				cursorX:   tt.fields.cursorX,
				cursorY:   tt.fields.cursorY,
			}
			text.Reset()
			if text.cursorX != 0 || text.cursorY != 0 {
				t.Errorf("TextArea.Reset() - cursor not reset, x=%d, y=%d", text.cursorX, text.cursorY)
			}
		})
	}
}

func TestTextArea_Size(t *testing.T) {
	type fields struct {
		Wrap      bool
		displayer drivers.Displayer
		ft        font.Font
		cursorX   int16
		cursorY   int16
	}
	tests := []struct {
		name   string
		fields fields
		want   int16
		want1  int16
	}{
		{"font multiple", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, 2, 2},
		{"smaller than font", fields{true, mocks.NewMockDisplayer(5, 16), font.NewFont6x8(), 0, 0}, 0, 2},
		{"not font multiple", fields{true, mocks.NewMockDisplayer(15, 19), font.NewFont6x8(), 0, 0}, 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := TextArea{
				Wrap:      tt.fields.Wrap,
				displayer: tt.fields.displayer,
				ft:        tt.fields.ft,
				cursorX:   tt.fields.cursorX,
				cursorY:   tt.fields.cursorY,
			}
			got, got1 := text.Size()
			if got != tt.want {
				t.Errorf("TextArea.Size() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TextArea.Size() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTextArea_PrintAt(t *testing.T) {
	type fields struct {
		Wrap      bool
		displayer drivers.Displayer
		ft        font.Font
		cursorX   int16
		cursorY   int16
	}
	type args struct {
		row int16
		col int16
		str string
		c   color.RGBA
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]byte
	}{
		{"inside", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{0, 0, "AB", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0},
			{0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"outside", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{-1, 0, "AB", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{"partial", fields{true, mocks.NewMockDisplayer(12, 16), font.NewFont6x8(), 0, 0}, args{0, -1, "AB", color.RGBA{0xff, 0xff, 0xff, 0xff}}, [][]byte{
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := &TextArea{
				Wrap:      tt.fields.Wrap,
				displayer: tt.fields.displayer,
				ft:        tt.fields.ft,
				cursorX:   tt.fields.cursorX,
				cursorY:   tt.fields.cursorY,
			}
			text.PrintAt(tt.args.row, tt.args.col, tt.args.str, tt.args.c)
			if got := tt.fields.displayer.(*mocks.MockDisplayer).GetPixels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextArea.PrintAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
