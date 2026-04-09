package game

import (
	"testing"
)

func TestNewBlock(t *testing.T) {
	// Test single character block
	b1 := NewBlock([]rune{'A'})
	if len(b1.Chars) != 1 {
		t.Errorf("Expected 1 char, got %d", len(b1.Chars))
	}
	if b1.Chars[0] != 'A' {
		t.Errorf("Expected char 'A', got %q", b1.Chars[0])
	}
	if b1.X != 0 || b1.Y != 0 {
		t.Error("New block should have position (0, 0)")
	}
	if b1.Rotation != 0 {
		t.Error("New block should not be rotated")
	}

	// Test double character block
	b2 := NewBlock([]rune{'A', 'B'})
	if len(b2.Chars) != 2 {
		t.Errorf("Expected 2 chars, got %d", len(b2.Chars))
	}
	if b2.Chars[0] != 'A' || b2.Chars[1] != 'B' {
		t.Error("Block chars don't match input")
	}
}

func TestBlockGetPositions(t *testing.T) {
	// Test single char block
	b1 := &Block{Chars: []rune{'X'}, X: 3, Y: 5}
	pos1 := b1.GetPositions()
	if len(pos1) != 1 {
		t.Errorf("Expected 1 position for single block, got %d", len(pos1))
	}
	if pos1[0].X != 3 || pos1[0].Y != 5 || pos1[0].Char != 'X' {
		t.Errorf("Position mismatch: got (%d, %d, %q), want (3, 5, 'X')", pos1[0].X, pos1[0].Y, pos1[0].Char)
	}

	// Test unrotated double block (horizontal)
	b2 := &Block{Chars: []rune{'A', 'B'}, X: 2, Y: 4, Rotation: 0}
	pos2 := b2.GetPositions()
	if len(pos2) != 2 {
		t.Errorf("Expected 2 positions for double block, got %d", len(pos2))
	}
	if pos2[0].X != 2 || pos2[0].Y != 4 || pos2[0].Char != 'A' {
		t.Errorf("First position mismatch: got (%d, %d, %q)", pos2[0].X, pos2[0].Y, pos2[0].Char)
	}
	if pos2[1].X != 3 || pos2[1].Y != 4 || pos2[1].Char != 'B' {
		t.Errorf("Second position mismatch: got (%d, %d, %q)", pos2[1].X, pos2[1].Y, pos2[1].Char)
	}

	// Test rotated double block (vertical)
	b3 := &Block{Chars: []rune{'C', 'D'}, X: 1, Y: 2, Rotation: 1}
	pos3 := b3.GetPositions()
	if len(pos3) != 2 {
		t.Errorf("Expected 2 positions for rotated double block, got %d", len(pos3))
	}
	if pos3[0].X != 1 || pos3[0].Y != 2 || pos3[0].Char != 'C' {
		t.Errorf("First position mismatch: got (%d, %d, %q)", pos3[0].X, pos3[0].Y, pos3[0].Char)
	}
	if pos3[1].X != 1 || pos3[1].Y != 3 || pos3[1].Char != 'D' {
		t.Errorf("Second position mismatch: got (%d, %d, %q)", pos3[1].X, pos3[1].Y, pos3[1].Char)
	}
}

func TestBlockRotate(t *testing.T) {
	// Test rotating double block cycles through 4 states
	b := &Block{Chars: []rune{'A', 'B'}, X: 5, Y: 10, Rotation: 0}
	rotated := b.Rotate()

	if rotated.Rotation != 1 {
		t.Errorf("First rotation should be state 1, got %d", rotated.Rotation)
	}
	if rotated.X != 5 || rotated.Y != 10 {
		t.Error("Rotation should preserve position")
	}
	if len(rotated.Chars) != 2 || rotated.Chars[0] != 'A' || rotated.Chars[1] != 'B' {
		t.Error("Rotation should preserve characters")
	}

	// Test full cycle returns to original state
	quadRotated := b.Rotate().Rotate().Rotate().Rotate()
	if quadRotated.Rotation != 0 {
		t.Errorf("Four rotations should return to state 0, got %d", quadRotated.Rotation)
	}

	// Test rotating single block (should return same block)
	bSingle := &Block{Chars: []rune{'X'}, X: 3, Y: 7, Rotation: 0}
	rotatedSingle := bSingle.Rotate()
	if rotatedSingle.Rotation != 0 {
		t.Error("Single block rotation should not change rotation state")
	}
	if len(rotatedSingle.Chars) != 1 {
		t.Error("Single block rotation should preserve chars")
	}
}

func TestBlockSetPosition(t *testing.T) {
	b := NewBlock([]rune{'A', 'B'})
	b.SetPosition(10, 20)

	if b.X != 10 {
		t.Errorf("Expected X = 10, got %d", b.X)
	}
	if b.Y != 20 {
		t.Errorf("Expected Y = 20, got %d", b.Y)
	}
}

func TestBlockClone(t *testing.T) {
	original := &Block{Chars: []rune{'A', 'B', 'C'}, X: 5, Y: 10, Rotation: 2}
	cloned := original.Clone()

	// Verify all fields are copied
	if len(cloned.Chars) != len(original.Chars) {
		t.Error("Clone should have same number of chars")
	}
	for i := range cloned.Chars {
		if cloned.Chars[i] != original.Chars[i] {
			t.Errorf("Char %d mismatch: got %q, want %q", i, cloned.Chars[i], original.Chars[i])
		}
	}
	if cloned.X != original.X || cloned.Y != original.Y {
		t.Error("Clone should have same position")
	}
	if cloned.Rotation != original.Rotation {
		t.Error("Clone should have same rotation state")
	}

	// Verify it's a deep copy (modifying clone shouldn't affect original)
	cloned.Chars[0] = 'Z'
	if original.Chars[0] == 'Z' {
		t.Error("Clone should be a deep copy")
	}

	cloned.X = 99
	if original.X == 99 {
		t.Error("Clone should be a deep copy")
	}
}

func TestBlockPosStruct(t *testing.T) {
	pos := BlockPos{X: 3, Y: 5, Char: 'X'}
	if pos.X != 3 || pos.Y != 5 || pos.Char != 'X' {
		t.Error("BlockPos struct fields not set correctly")
	}
}

func BenchmarkBlockGetPositions(b *testing.B) {
	block := &Block{Chars: []rune{'A', 'B'}, X: 5, Y: 10, Rotation: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.GetPositions()
	}
}

func BenchmarkBlockRotate(b *testing.B) {
	block := &Block{Chars: []rune{'A', 'B'}, X: 5, Y: 10, Rotation: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.Rotate()
	}
}

func BenchmarkBlockClone(b *testing.B) {
	block := &Block{Chars: []rune{'A', 'B', 'C', 'D', 'E'}, X: 5, Y: 10, Rotation: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.Clone()
	}
}
