// Contains tables and encodings necessary for key generation.
package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

// MaskTable maps one byte to a block, according to an input or output mask.
type MaskTable struct {
	Mask     matrix.Matrix
	Position int
}

func (mt MaskTable) Get(i byte) (out [16]byte) {
	r := make([]byte, 16)
	r[mt.Position] = i

	res := mt.Mask.Mul(matrix.Row(r))
	copy(out[:], res)

	return
}

// A MB^(-1) Table inverts the mixing bijection on the Tyi Table.
type MBInverseTable struct {
	MBInverse matrix.Matrix
	Row       uint
}

func (mbinv MBInverseTable) Get(i byte) (out [4]byte) {
	r := matrix.Row{0, 0, 0, 0}
	r[mbinv.Row] = i

	res := mbinv.MBInverse.Mul(r)
	copy(out[:], res)

	return
}

// An XOR Table computes the XOR of two nibbles.
type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

// Encodes the output of an input/output mask.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the mask's output for this byte, counted in nibbles.
func MaskEncoding(seed []byte, position, subPosition int, surface common.Surface) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'E', byte(position), byte(subPosition), byte(surface)

	return common.GetShuffle(seed, label)
}

func BlockMaskEncoding(seed []byte, position int, surface common.Surface, shift func(int) int) encoding.Block {
	out := encoding.ConcatenatedBlock{}

	for i := 0; i < 16; i++ {
		out[i] = encoding.ConcatenatedByte{
			MaskEncoding(seed, position, 2*i+0, surface),
			MaskEncoding(seed, position, 2*i+1, surface),
		}

		if surface == common.Inside {
			out[i] = encoding.ComposedBytes{
				encoding.ByteLinear{common.MixingBijection(seed, 8, -1, shift(i)), nil},
				out[i],
			}
		}
	}

	return out
}

// Abstraction over the Tyi and MB^(-1) encodings, to match the pattern of the XOR and round encodings.
func StepEncoding(seed []byte, round, position, subPosition int, surface common.Surface) encoding.Nibble {
	if surface == common.Inside {
		return TyiEncoding(seed, round, position, subPosition)
	} else {
		return MBInverseEncoding(seed, round, position, subPosition)
	}
}

func WordStepEncoding(seed []byte, round, position int, surface common.Surface) encoding.Word {
	out := encoding.ConcatenatedWord{}

	for i := 0; i < 4; i++ {
		out[i] = encoding.ConcatenatedByte{
			StepEncoding(seed, round, position, 2*i+0, surface),
			StepEncoding(seed, round, position, 2*i+1, surface),
		}
	}

	return out
}

// Encodes the output of a T-Box/Tyi Table / the input of an XOR Table.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the T-Box/Tyi Table's ouptput for this byte, counted in nibbles.
func TyiEncoding(seed []byte, round, position, subPosition int) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3] = 'T', byte(round), byte(position), byte(subPosition)

	return common.GetShuffle(seed, label)
}

// Encodes the output of a MB^(-1) Table / the input of an XOR Table.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the MB^(-1) Table's ouptput for this byte, counted in nibbles.
func MBInverseEncoding(seed []byte, round, position, subPosition int) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'I', byte(round), byte(position), byte(subPosition)

	return common.GetShuffle(seed, label)
}

// Encodes intermediate results between each successive XOR.
//
// position: Position in the state array, counted in nibbles.
//     gate: The gate number, or, the number of XORs we've computed so far.
//  surface: Location relative to the round structure. Inside or Outside.
func XOREncoding(seed []byte, round, position, gate int, surface common.Surface) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'X', byte(round), byte(position), byte(gate), byte(surface)

	return common.GetShuffle(seed, label)
}

// Encodes the output of an Expand->Squash round. Two Expand->Squash rounds make up one AES round.
//
// position: Position in the state array, counted in nibbles.
//  surface: Location relative to the AES round structure. Inside or Outside.
func RoundEncoding(seed []byte, round, position int, surface common.Surface) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3] = 'R', byte(round), byte(position), byte(surface)

	return common.GetShuffle(seed, label)
}

func ByteRoundEncoding(seed []byte, round, position int, surface common.Surface) encoding.Byte {
	return encoding.ConcatenatedByte{
		RoundEncoding(seed, round, 2*position+0, surface),
		RoundEncoding(seed, round, 2*position+1, surface),
	}
}
