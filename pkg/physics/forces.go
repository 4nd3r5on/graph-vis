package physics

import (
	"graph-positioner/pkg/vec"
)

const minDist = 1e-6 // prevent div-by-zero
const maxForce = 1e6

func CoolingFunc(initCooling float64, currentIter float64) (cooling float64) {
	return initCooling * (1 - currentIter*currentIter)
}

// RepulsiveForce calculates the repulsive force between two positions
// f_rep = ((springLen^2) / dist) * unitVec
func RepulsiveForce(from, to vec.Vec, springLen float64) vec.Vec {
	diff := vec.Sub(from, to)
	dist := vec.Mag(diff)

	if dist < minDist {
		dist = minDist
	}

	magnitude := (springLen * springLen) / dist
	if magnitude > maxForce {
		magnitude = maxForce
	}

	return vec.Mul(vec.Unit(diff), magnitude)
}

// AttractiveForce calculates the attractive force between two positions
// f_attr = (dist^2 / springLen) * unitVec
func AttractiveForce(from, to vec.Vec, springLen float64) vec.Vec {
	diff := vec.Sub(to, from)
	dist := vec.Mag(diff)

	if springLen <= 0 {
		springLen = minDist // avoid div-by-zero
	}
	if dist < minDist {
		dist = minDist
	}

	magnitude := (dist * dist) / springLen
	if magnitude > maxForce {
		magnitude = maxForce
	}

	return vec.Mul(vec.Unit(diff), magnitude)
}
