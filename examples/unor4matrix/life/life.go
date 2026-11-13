//go:build arduino_uno_r4_wifi

package main

const (
	width  = 12
	height = 8
)

type board [height][width]bool

func nextGeneration(current board) (next board) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			neighbors := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					if current[(y+dy+height)%height][(x+dx+width)%width] {
						neighbors++
					}
				}
			}
			next[y][x] = neighbors == 3 || current[y][x] && neighbors == 2
		}
	}
	return next
}

func seed() board {
	var cells board
	cells[0][1] = true
	cells[1][2] = true
	cells[2][0] = true
	cells[2][1] = true
	cells[2][2] = true
	return cells
}
