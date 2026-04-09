package game

// Rotation states for a 2-character block (clockwise cycle):
//
//	0: horizontal, normal order    [chars[0] chars[1]]
//	1: vertical, normal order      chars[0] / chars[1]
//	2: horizontal, reversed order  [chars[1] chars[0]]
//	3: vertical, reversed order    chars[1] / chars[0]
type Block struct {
	Chars    []rune
	X, Y     int
	Rotation int
}

type BlockPos struct {
	X    int
	Y    int
	Char rune
}

func NewBlock(chars []rune) *Block {
	return &Block{
		Chars:    chars,
		X:        0,
		Y:        0,
		Rotation: 0,
	}
}

func (b *Block) GetPositions() []BlockPos {
	positions := make([]BlockPos, len(b.Chars))
	if len(b.Chars) == 1 {
		positions[0] = BlockPos{X: b.X, Y: b.Y, Char: b.Chars[0]}
	} else if len(b.Chars) == 2 {
		switch b.Rotation % 4 {
		case 0: // horizontal, normal
			positions[0] = BlockPos{X: b.X, Y: b.Y, Char: b.Chars[0]}
			positions[1] = BlockPos{X: b.X + 1, Y: b.Y, Char: b.Chars[1]}
		case 1: // vertical, normal
			positions[0] = BlockPos{X: b.X, Y: b.Y, Char: b.Chars[0]}
			positions[1] = BlockPos{X: b.X, Y: b.Y + 1, Char: b.Chars[1]}
		case 2: // horizontal, reversed
			positions[0] = BlockPos{X: b.X, Y: b.Y, Char: b.Chars[1]}
			positions[1] = BlockPos{X: b.X + 1, Y: b.Y, Char: b.Chars[0]}
		case 3: // vertical, reversed
			positions[0] = BlockPos{X: b.X, Y: b.Y, Char: b.Chars[1]}
			positions[1] = BlockPos{X: b.X, Y: b.Y + 1, Char: b.Chars[0]}
		}
	}
	return positions
}

func (b *Block) Rotate() *Block {
	if len(b.Chars) != 2 {
		return b
	}
	return &Block{
		Chars:    b.Chars,
		X:        b.X,
		Y:        b.Y,
		Rotation: (b.Rotation + 1) % 4,
	}
}

func (b *Block) SetRotation(r int) {
	b.Rotation = r % 4
}

func (b *Block) SetPosition(x, y int) {
	b.X = x
	b.Y = y
}

func (b *Block) Clone() *Block {
	charsCopy := make([]rune, len(b.Chars))
	copy(charsCopy, b.Chars)
	return &Block{
		Chars:    charsCopy,
		X:        b.X,
		Y:        b.Y,
		Rotation: b.Rotation,
	}
}
