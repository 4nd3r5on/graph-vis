package vec

import (
	"hash/fnv"
	"math"
)

type Vec struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// New creates a new Vec
func New(x, y float64) Vec {
	return Vec{X: x, Y: y}
}

// Add returns the sum of vectors
func Add(vecs ...Vec) Vec {
	result := Vec{}
	for _, v := range vecs {
		result.X += v.X
		result.Y += v.Y
	}
	return result
}

// Sub returns the difference of two vectors (a - b)
func Sub(a, b Vec) Vec {
	return Vec{
		X: a.X - b.X,
		Y: a.Y - b.Y,
	}
}

// Mul returns the vector multiplied by a scalar
func Mul(v Vec, scalar float64) Vec {
	return Vec{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

// Div returns the vector divided by a scalar
func Div(v Vec, scalar float64) Vec {
	if scalar == 0 {
		return Vec{}
	}
	return Vec{
		X: v.X / scalar,
		Y: v.Y / scalar,
	}
}

// Mag returns the magnitude (length) of the vector
func Mag(v Vec) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// MagSq returns the squared magnitude (avoids sqrt for performance)
func MagSq(v Vec) float64 {
	return v.X*v.X + v.Y*v.Y
}

// Dist returns the distance between two vectors
func Dist(a, b Vec) float64 {
	return Mag(Sub(a, b))
}

// DistSq returns the squared distance (avoids sqrt for performance)
func DistSq(a, b Vec) float64 {
	return MagSq(Sub(a, b))
}

// Unit returns the unit vector (normalized to length 1)
// Returns zero vector if magnitude is 0
func Unit(v Vec) Vec {
	mag := Mag(v)
	if mag == 0 {
		return Vec{}
	}
	return Div(v, mag)
}

// Dot returns the dot product of two vectors
func Dot(a, b Vec) float64 {
	return a.X*b.X + a.Y*b.Y
}

// IsZero checks if the vector is a zero vector
func IsZero(v Vec) bool {
	return v.X == 0 && v.Y == 0
}

// Boundaries creates boundary vectors based on node count and padding
// Returns rightBottom (side, side)
// where side = ceil(sqrt(nodeCount * padding))
func Boundaries(nodeCount int, padding float64) (rightBottom Vec) {
	area := float64(nodeCount) * padding
	side := math.Ceil(math.Sqrt(area))

	return New(side, side)
}

// Random returns a pseudo-random Vec within boundaries using FNV hash
// seed can be any byte slice (e.g., []byte("node-id"))
// leftTop is the top-left corner, rightBottom is the bottom-right corner
func Random(seed []byte, leftTop, rightBottom Vec) Vec {
	// Create FNV hash
	h := fnv.New64a()
	h.Write(seed)
	hash1 := h.Sum64()

	// Create second hash by appending to seed
	h.Reset()
	h.Write(append(seed, 0xFF))
	hash2 := h.Sum64()

	// Convert hash to float64 in range [0, 1)
	normX := float64(hash1) / float64(^uint64(0))
	normY := float64(hash2) / float64(^uint64(0))

	// Map to boundaries
	width := rightBottom.X - leftTop.X
	height := rightBottom.Y - leftTop.Y

	return Vec{
		X: leftTop.X + normX*width,
		Y: leftTop.Y + normY*height,
	}
}
