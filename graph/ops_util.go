/*
 *	Copyright 2023 Jan Pfeifer
 *
 *	Licensed under the Apache License, Version 2.0 (the "License");
 *	you may not use this file except in compliance with the License.
 *	You may obtain a copy of the License at
 *
 *	http://www.apache.org/licenses/LICENSE-2.0
 *
 *	Unless required by applicable law or agreed to in writing, software
 *	distributed under the License is distributed on an "AS IS" BASIS,
 *	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *	See the License for the specific language governing permissions and
 *	limitations under the License.
 */

package graph

import (
	. "github.com/gomlx/exceptions"
	"github.com/gomlx/gomlx/types/shapes"
	"github.com/gomlx/gomlx/types/xslices"
	"github.com/gomlx/gopjrt/dtypes"
)

// This file contains derived practical calculations that often used.

// Scalar returns a constant scalar with the given value.
func Scalar(g *Graph, dtype dtypes.DType, value float64) *Node {
	return g.getScalarConst(dtype, value)
}

// FillScalar creates a Node with a value with the given shape, filled with the given value.
// It's implemented indirectly using other nodes.
func FillScalar(g *Graph, shape shapes.Shape, value float64) *Node {
	return BroadcastPrefix(Scalar(g, shape.DType, value), shape.Dimensions...)
}

// ScalarZero returns a scalar constant 0 for the given DType.
func ScalarZero(g *Graph, dtype dtypes.DType) *Node {
	return Scalar(g, dtype, 0)
}

// ScalarOne returns a scalar constant 1 for the given DType.
func ScalarOne(g *Graph, dtype dtypes.DType) *Node {
	return Scalar(g, dtype, 1)
}

// MulScalar converts scalar to a constant with x's DType and returns `x * scalar`
// with proper broadcasting.
func MulScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	return Mul(x, Scalar(g, x.DType(), scalar))
}

// DivScalar converts scalar to a constant with x's DType and returns `x / scalar`
// with proper broadcasting.
//
// For float DType's, DivScalar instead uses MulScalar(x, 1/scalar).
func DivScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	if scalar == 0 {
		Panicf("division by zero in DivScalar")
	}
	if x.DType().IsFloat() {
		// Multiply by 1/scalar instead:
		return MulScalar(x, 1.0/float64(scalar))
	}
	return Div(x, Scalar(g, x.DType(), scalar))
}

// PowScalar converts scalar to a constant with x's DType and returns `Pow(x, scalar)` (or `x ** scalar`)
// with proper broadcasting.
func PowScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	return Pow(x, Scalar(g, x.DType(), scalar))
}

// Square returns x^2 point-wise. Same as `Mul(x, x)`.
func Square(x *Node) *Node {
	return Mul(x, x)
}

// AddScalar converts scalar to a constant with x's DType and returns `x + scalar`
// with proper broadcasting.
func AddScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	return Add(x, Scalar(g, x.DType(), scalar))
}

// ModScalar converts scalar to a constant with x's DType and returns `x % scalar`
// with proper broadcasting.
func ModScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	return Mod(x, Scalar(g, x.DType(), scalar))
}

// MaxScalar converts scalar to a constant with x's DType and returns element-wise `Max(x, scalar)`.
func MaxScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	return Max(x, Scalar(g, x.DType(), scalar))
}

// MinScalar converts scalar to a constant with x's DType and returns element-wise `Min(x, scalar)`.
func MinScalar(x *Node, scalar float64) *Node {
	g := x.Graph()
	return Min(x, Scalar(g, x.DType(), scalar))
}

func lowestForDType(g *Graph, dtype dtypes.DType) *Node {
	return Const(g, dtype.LowestValue())
}

// OnesLike returns a tensor with the same shape of x, filled with 1's.
func OnesLike(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	return Ones(g, x.Shape())
}

// Ones creates a computation with the same shape as the input, but with the value 1.
// It's implemented indirectly using other nodes.
func Ones(g *Graph, shape shapes.Shape) *Node {
	g.AssertBuilding()
	scalar := ScalarOne(g, shape.DType)
	if scalar == nil {
		return nil
	}
	return BroadcastPrefix(scalar, shape.Dimensions...)
}

// ZerosLike returns a tensor with the same shape of x, filled with 0's.
func ZerosLike(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	return Zeros(g, x.Shape())
}

// Zeros creates a computation with the same shape as the input, but with the value 0.
// It's implemented indirectly using other nodes.
func Zeros(g *Graph, shape shapes.Shape) *Node {
	g.AssertBuilding()
	return BroadcastPrefix(ScalarZero(g, shape.DType), shape.Dimensions...)
}

// OneMinus returns (1-x).
func OneMinus(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	return Sub(ScalarOne(g, x.DType()), x)
}

// MinusOne returns (x-1).
func MinusOne(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	return Sub(x, ScalarOne(g, x.DType()))
}

// OnePlus returns (1+x).
func OnePlus(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	return Add(ScalarOne(g, x.DType()), x)
}

// Inverse returns (1/x), the multiplicative inverse. Also known as the reciprocal.
func Inverse(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	return Div(ScalarOne(g, x.DType()), x)
}

// SignPlusOrMinus return +1 or -1 whether x >= 0 or x < 0. It's similar to Sign, but
// where 0s are considered positive.
func SignPlusOrMinus(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	half := Scalar(g, x.DType(), 0.5)
	return Sign(Add(Sign(x), half))
}

// PositiveIndicator returns 1 where x >= 0, 0 otherwise. See also StrictlyPositiveIndicator.
// E.g: PositiveIndicator({1.0, 0.0001, 0, -0.2, -3.0}) -> [1, 1, 1, 0, 0], with the same shape/dtype as x.
func PositiveIndicator(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	one := ScalarOne(g, x.DType())
	return Sign(Add(Sign(x), one))
}

// MirroredLog1p is similar to Log1p, but it is mirrored to negative numbers.
// It return Log(Abs(x)+1)*Sign(x).
func MirroredLog1p(x *Node) *Node {
	return Mul(Log1p(Abs(x)), Sign(x))
}

// StrictlyPositiveIndicator returns 1 where x > 0, 0 otherwise.
// E.g: StrictlyPositiveIndicator({1.0, 0.0001, 0, -0.2, -3.0}) -> [1, 1, 0, 0, 0], with the same shape/dtype as x.
func StrictlyPositiveIndicator(x *Node) *Node {
	g := validateBuildingGraphFromInputs(x)
	one := ScalarOne(g, x.DType())
	return Add(Sign(Sub(Sign(x), one)), one)
}

// Clip is a shortcut to `Min(max, Max(x, min))`, which returns the values of x clipped between
// min and max.
func Clip(x, min, max *Node) *Node {
	return Min(max, Max(x, min))
}

// ClipScalar is a shortcut to `Min(max, Max(x, min))`, which returns the values of x clipped between
// min and max. The values min and max are given as scalar values -- the float64 is converted to the
// `DType` of x.
func ClipScalar(x *Node, min, max float64) *Node {
	return MinScalar(MaxScalar(x, min), max)
}

// OneHot converts an integer numbers representing indices to it's "one-hot" representation, that is an expanded
// tensor with the indices position set to 1, and the other positions set to 0. The returned tensor has one extra
// dimension at the end.
// For example `OneHot([][]INT64{1, 0, 3}, 4, types.Float32)` returns  `[][]F32{{0, 1, 0, 0}, {1, 0, 0, 0}, {0, 0, 0, 1}}`
func OneHot(indices *Node, depth int, dtype dtypes.DType) *Node {
	g := indices.Graph()
	if !indices.DType().IsInt() {
		Panicf("invalid indices dtype (%s), it must be integer", indices.DType())
	}

	// Add an expanded dimension at the end, which will contain the one-hot representation.
	// The new last axis will be broadcast to the same dimension as positionIndices
	indices = ExpandDims(indices, -1)

	positionIndicesShape := indices.Shape().Clone()
	for ii := range indices.Rank() - 1 {
		positionIndicesShape.Dimensions[ii] = 1
	}
	positionIndicesShape.Dimensions[positionIndicesShape.Rank()-1] = depth
	positionIndices := Iota(g, positionIndicesShape, -1) // Indices for each "bit" in position.
	return ConvertDType(Equal(indices, positionIndices), dtype)
}

// ReduceAndKeep applies the given reduction function but regenerate the reduced dimensions with size 1.
func ReduceAndKeep(x *Node, reduceFn func(x *Node, reduceAxes ...int) *Node, reduceAxes ...int) *Node {
	_ = validateBuildingGraphFromInputs(x)
	rank := x.Rank()
	reduceAxes = adjustAxesToRankAndSort(rank, reduceAxes, "x")
	reduced := reduceFn(x, reduceAxes...)
	shapeWithRecoveredDims := x.Shape().Clone()
	for ii := 0; ii < rank && len(reduceAxes) > 0; ii++ {
		if ii == reduceAxes[0] {
			shapeWithRecoveredDims.Dimensions[ii] = 1
			reduceAxes = reduceAxes[1:]
		}
	}
	return Reshape(reduced, shapeWithRecoveredDims.Dimensions...)
}

// MaskedReduceAndKeep applies the given masked reduction function but regenerates the reduced
// dimensions with size 1.
func MaskedReduceAndKeep(x, mask *Node, reduceFn func(x, mask *Node, reduceAxes ...int) *Node, reduceAxes ...int) *Node {
	_ = validateBuildingGraphFromInputs(x)
	rank := x.Rank()
	reduceAxes = adjustAxesToRankAndSort(rank, reduceAxes, "x")
	reduced := reduceFn(x, mask, reduceAxes...)
	shapeWithRecoveredDims := x.Shape().Clone()
	for ii := 0; ii < rank && len(reduceAxes) > 0; ii++ {
		if ii == reduceAxes[0] {
			shapeWithRecoveredDims.Dimensions[ii] = 1
			reduceAxes = reduceAxes[1:]
		}
	}
	return Reshape(reduced, shapeWithRecoveredDims.Dimensions...)
}

// ReduceAndKeepMasked is an alias for MaskedReduceAndKeep.
//
// Deprecated: all functions that take mask are prefixed with `Masked...`
var ReduceAndKeepMasked = MaskedReduceAndKeep

// Softmax computes softmax activations. It's the equivalent to
// ```
//
//	Exp(logits) / ExpandDims(ReduceSum(Exp(logits), -1), -1)
//
// ```
//
// But implemented in a numerical stable way.
//
// The list axes defines which axes is it supposed to run the softmax over
// (the axes that will be summed over). If no axes are given, it is assumed to
// be [-1], meaning, the last axes.
func Softmax(logits *Node, axes ...int) *Node {
	_ = validateBuildingGraphFromInputs(logits)
	if !logits.DType().IsFloat() {
		Panicf("invalid logits dtype (%s), it must be float", logits.DType())
	}
	if len(axes) == 0 {
		axes = []int{-1}
	}
	normalizingMax := StopGradient(ReduceAndKeep(logits, ReduceMax, axes...))
	normalizedLogits := Sub(logits, normalizingMax)
	numerator := Exp(normalizedLogits)
	denominator := ReduceAndKeep(numerator, ReduceSum, axes...)
	return Div(numerator, denominator)
}

// MaskedSoftmax computes softmax activations. It's the equivalent to
// ```
//
//	Exp(logits) / ExpandDims(ReduceSum(Exp(logits), -1), -1)
//
// ```
//
// But implemented in a numerical stable way.
//
// It takes a mask that is true on the values to be considered, and false for the values
// not to be considered.
//
// The list axes defines which axes is it supposed to run the softmax over
// (the axes that will be summed over). If no axes are given, it is assumed to
// be [-1], meaning, the last axes.
//
// It ignores values for which the corresponding mask is false, and will return 0 for
// those fields. mask and logits must have the same shape.
func MaskedSoftmax(logits, mask *Node, axes ...int) *Node {
	if mask == nil {
		return Softmax(logits, axes...)
	}
	_ = validateBuildingGraphFromInputs(logits, mask)
	if !logits.DType().IsFloat() {
		Panicf("invalid logits dtype (%s), it must be float", logits.DType())
	}
	if len(axes) == 0 {
		axes = []int{-1}
	}
	normalizingMax := StopGradient(MaskedReduceAndKeep(logits, mask, MaskedReduceMax, axes...))
	zeros := ZerosLike(logits)
	normalizedLogits := Sub(logits, normalizingMax)
	normalizedLogits = Where(mask, normalizedLogits, zeros)
	numerator := Exp(normalizedLogits)
	numerator = Where(mask, numerator, zeros)
	// Apply mask on numerator, setting softmax to zero where masked.
	denominator := ReduceAndKeep(numerator, ReduceSum, axes...)
	result := Div(numerator, denominator)
	result = Where(mask, result, zeros)
	return result
}

// LogSoftmax computes the logarithm of the Softmax function, which rescales
// elements to the range $[-\infty, 0)$.
//
//	$$
//	\mathrm{log\_softmax}(x)_i = \log \left( \frac{\exp(x_i)}{\sum_j \exp(x_j)}
//	\right)
//	$$
//
// The axes define over which axes the LogSoftmax should be computed. If missing it is assumed to be -1.
//
// If any input values are "+inf", the result will be all "NaN": this reflects the
// fact that "inf / inf" is not well-defined in the context of floating-point math.
func LogSoftmax(logits *Node, axes ...int) *Node {
	_ = validateBuildingGraphFromInputs(logits)
	if !logits.DType().IsFloat() {
		Panicf("invalid logits dtype (%s), it must be float", logits.DType())
	}
	if len(axes) == 0 {
		axes = []int{-1}
	}
	adjustedAxes := adjustAxesToRankAndSort(logits.Rank(), axes, "logits")
	normalizingMax := StopGradient(ReduceAndKeep(logits, ReduceMax, adjustedAxes...))
	shiftedLogits := Sub(logits, normalizingMax)
	shiftedLogSumExp := Log(ReduceAndKeep(Exp(shiftedLogits), ReduceSum, adjustedAxes...))
	return Sub(shiftedLogits, shiftedLogSumExp)
}

// MaskedLogSoftmax computes the logarithm of the MaskedSoftmax function, which rescales
// elements to the range $[-\infty, 0)$.
//
// It takes a mask that is true on the values to be considered, and false for the values
// not to be considered.
// If mask is nil, it behaves like LogSoftmask.
//
// See LogSoftmax for details.
func MaskedLogSoftmax(logits, mask *Node, axes ...int) *Node {
	if mask == nil {
		return LogSoftmax(logits, axes...)
	}
	g := validateBuildingGraphFromInputs(logits, mask)
	dtype := logits.DType()
	if !dtype.IsFloat() {
		Panicf("invalid logits dtype (%s), it must be float", logits.DType())
	}
	if len(axes) == 0 {
		axes = []int{-1}
	}
	adjustedAxes := adjustAxesToRankAndSort(logits.Rank(), axes, "logits")
	normalizingMax := StopGradient(MaskedReduceAndKeep(logits, mask, MaskedReduceMax, adjustedAxes...))
	shiftedLogits := Sub(logits, normalizingMax)
	shiftedLogSumExp := Log(MaskedReduceAndKeep(Exp(shiftedLogits), mask, MaskedReduceSum, adjustedAxes...))
	return Where(mask, Sub(shiftedLogits, shiftedLogSumExp), Infinity(g, dtype, -1))
}

// L1Norm returns the L1 norm (same as Manhattan length) of the last axis of x.
// The returned value has the same rank, but the last axes will have dimension 1.
//
// If no axes are given, it returns a scalar.
// Otherwise, the returned value has the same rank as `x`, but the reduce axes will have dimension 1.
func L1Norm(x *Node, reduceAxes ...int) *Node {
	if len(reduceAxes) == 0 {
		return ReduceAllSum(Abs(x))
	}
	return ReduceAndKeep(Abs(x), ReduceSum, -1)
}

// L2NormSquare returns the L2 norm square (same as square of the Euclidean length) over the given axes
// of x (defaults to all).
// Same as `\Sum_{reduceAxes}{x_i^2}`.
//
// If no axes are given, it returns a scalar.
// Otherwise, the returned value has the same rank as `x`, but the reduce axes will have dimension 1.
func L2NormSquare(x *Node, reduceAxes ...int) *Node {
	if len(reduceAxes) == 0 {
		return ReduceAllSum(Square(x))
	}
	return ReduceAndKeep(Square(x), ReduceSum, reduceAxes...)
}

// L2Norm returns the L2 norm (same as Euclidean length) over the given axes of x (defaults to all), given by Sqrt(\Sum{x_i^2}).
//
// If no axes are given, it returns a scalar.
// Otherwise, the returned value has the same rank as `x`, but the reduce axes will have dimension 1.
func L2Norm(x *Node, reduceAxes ...int) *Node {
	return Sqrt(L2NormSquare(x, reduceAxes...))
}

// L2Normalize returns `x/L2Norm(x)` on the given reduce axes, making the last axis a unit-length vector.
//
// It will return `inf` for values of x that are zero-length.
// See L2NormalizeWithEpsilon for a version that adds an epsilon to the denominator to avoid that.
func L2Normalize(x *Node, reduceAxis int, moreReduceAxes ...int) *Node {
	reduceAxes := make([]int, 1, 1+len(moreReduceAxes))
	reduceAxes[0] = reduceAxis
	reduceAxes = append(reduceAxes, moreReduceAxes...)
	return Div(x, L2Norm(x, reduceAxes...))
}

// L2NormalizeWithEpsilon returns `x/(L2Norm(x)+epsilon)` on the last axis, making the last axis a unit-length vector.
func L2NormalizeWithEpsilon(x *Node, epsilon float64, reduceAxis int, moreReduceAxes ...int) *Node {
	reduceAxes := make([]int, 1, 1+len(moreReduceAxes))
	reduceAxes[0] = reduceAxis
	reduceAxes = append(reduceAxes, moreReduceAxes...)
	return Div(x, AddScalar(L2Norm(x, reduceAxes...), epsilon))
}

// LowerTriangular returns a lower-triangular boolean square matrix of shape `[dim, dim]`.
//
// This can be combined with `Where` to select values of any arbitrary other matrix.
func LowerTriangular(g *Graph, dim int) *Node {
	shapeInt := shapes.Make(dtypes.Int64, dim, dim)
	rows := Iota(g, shapeInt, 0)
	cols := Iota(g, shapeInt, 1)
	return LessOrEqual(cols, rows)
}

// UpperTriangular returns an upper-triangular boolean square matrix of shape `[dim, dim]`.
//
// This can be combined with `Where` to select values of any arbitrary other matrix.
func UpperTriangular(g *Graph, dim int) *Node {
	shapeInt := shapes.Make(dtypes.Int64, dim, dim)
	rows := Iota(g, shapeInt, 0)
	cols := Iota(g, shapeInt, 1)
	return GreaterOrEqual(cols, rows)
}

// Diagonal returns a diagonal boolean square matrix of shape `[dim, dim]`.
//
// This can be combined with `Where` to select values of any arbitrary other matrix.
func Diagonal(g *Graph, dim int) *Node {
	shapeInt := shapes.Make(dtypes.Int64, dim, dim)
	rows := Iota(g, shapeInt, 0)
	cols := Iota(g, shapeInt, 1)
	return Equal(cols, rows)
}

// DiagonalWithValue returns a diagonal matrix of shape `[dim, dim]` with
// scalar in the diagonal and zero elsewhere. Set scalar to `ScalarOne()`
// and you get an identity matrix.
func DiagonalWithValue(scalar *Node, dim int) *Node {
	g := scalar.Graph()
	matrix := BroadcastPrefix(scalar, dim, dim)
	return Where(Diagonal(g, dim), matrix, ZerosLike(matrix))
}

// ShiftLeft the last axis of [x] by [n] positions ([n] is a static value) and fill the new value
// with [fill]. The value of [fill] is converted to [x]'s [dtypes.DType]. For boolean dtype, use 1.0 or 0.0.
//
// See [ShiftWithScalar] and [ShiftWithValue] for a more generic shift function.
func ShiftLeft(x *Node, n int, fill float64) *Node {
	return ShiftWithScalar(x, -1, ShiftDirLeft, n, fill)
}

// ShiftRight the last axis of [x] by [n] positions ([n] is a static value) and fill the new value
// with [fill]. The value of [fill] is converted to [x]'s [dtypes.DType]. For boolean dtype, use 1.0 or 0.0.
//
// See [ShiftWithScalar] and [ShiftWithValue] for a more generic shift function.
func ShiftRight(x *Node, n int, fill float64) *Node {
	return ShiftWithScalar(x, -1, ShiftDirRight, n, fill)
}

// ShiftDirection used by [ShiftWithScalar] and [ShiftWithValue]. See [ShiftDirLeft] and [ShiftDirRight].
type ShiftDirection bool

const (
	ShiftDirLeft  ShiftDirection = false
	ShiftDirRight                = true
)

// String implements the stringer interface.
func (s ShiftDirection) String() string {
	if s == ShiftDirRight {
		return "ShiftDirRight"
	}
	return "ShiftDirLeft"
}

// ShiftWithScalar a given [axis] of [x] by [n] positions ([n] is a static value) and fill the new value
// with [fill], a **static** scalar value.
// The [shiftDir] defines the direction: left towards lower values or right towards higher values.
// The value of [fill] is converted to [x]'s [dtypes.DType]. For boolean dtype, use 1.0 or 0.0.
func ShiftWithScalar(x *Node, axis int, shiftDir ShiftDirection, n int, fill float64) *Node {
	return genericShiftImpl(x, axis, shiftDir, n, fill, nil)
}

// ShiftWithValue a given [axis] of [x] by [n] positions ([n] is a static value) and fill the new value
// with a dynamic (graph) [value].
// The [shiftDir] defines the direction: left towards lower values or right towards higher values.
// The filling [value] must be "broadcast-able" (see [BroadcastToDim]) to the space it's going to fill with the shift --
// a scalar can always be broadcast.
func ShiftWithValue(x *Node, axis int, shiftDir ShiftDirection, n int, value *Node) *Node {
	return genericShiftImpl(x, axis, shiftDir, n, 0, value)
}

// Shift a given [axis] of [x] by [n] positions ([n] is a static value).
// The [shiftDir] defines the direction: left towards lower values or right towards higher values.
// The spaces left open keep the edge value. Example:
//
//	Shift([0, 1, 2, 3], axis=-1, ShiftDirLeft, n=2)
//
// Will return `[2, 3, 3, 3]`.
func Shift(x *Node, axis int, shiftDir ShiftDirection, n int) *Node {
	rank := x.Rank()
	dims := x.Shape().Dimensions
	shiftAxis := AdjustAxisToOperandRank(x, axis)

	// Find slice of left-most / right-most values to use for filling.
	axisRanges := make([]SliceAxisSpec, rank)
	for ii := range rank {
		if ii != shiftAxis {
			// Take full axes that are not shifted.
			axisRanges[ii] = AxisRange()
			continue
		}
		if shiftDir == ShiftDirLeft {
			axisRanges[ii] = AxisRange(dims[ii] - 1) // Take last value.
		} else {
			axisRanges[ii] = AxisRange(0, 1) // Take first value.
		}
	}
	fillValues := Slice(x, axisRanges...)
	return ShiftWithValue(x, axis, shiftDir, n, fillValues)
}

// genericShiftImpl implements ShiftWithScalar and GenericShitWithValue.
func genericShiftImpl(x *Node, axis int, shiftDir ShiftDirection, n int, fill float64, value *Node) *Node {
	g := x.Graph()
	dtype := x.DType()
	rank := x.Rank()
	dims := x.Shape().Dimensions
	shiftAxis := AdjustAxisToOperandRank(x, axis)
	if n > dims[shiftAxis] {
		Panicf("cannot shift %d positions for axis %d, x.shape=%s", n, axis, x.Shape())
	}
	if value != nil && value.DType() != dtype {
		Panicf("cannot shift x.shape=%s using value.shape=%s with a different dtype", x.Shape(), value.Shape())
	}
	if n == 0 {
		// Trivial solution.
		return x
	}

	// Slice part of the tensor that stays.
	axisRanges := make([]SliceAxisSpec, rank)
	fillDims := make([]int, rank)
	for ii := range rank {
		if ii != shiftAxis {
			// Take axes that are not shifted and fill the full dimension.
			axisRanges[ii] = AxisRange()
			fillDims[ii] = dims[ii]
			continue
		}
		if shiftDir == ShiftDirLeft {
			axisRanges[ii] = AxisRange(n)
		} else {
			axisRanges[ii] = AxisRange(0, dims[ii]-n)
		}
		fillDims[ii] = n
	}

	xSlice := Slice(x, axisRanges...)
	var xFill *Node
	if value == nil {
		// Fill with given value.
		if fill == 0.0 {
			xFill = Zeros(g, shapes.Make(dtype, fillDims...))
		} else {
			xFill = Ones(g, shapes.Make(dtype, fillDims...))
			if fill != 1.0 {
				xFill = MulScalar(xFill, fill)
			}
		}
	} else {
		// Fill with value broadcast on the required dimensions.
		xFill = BroadcastToDims(value, fillDims...)
	}

	if shiftDir == ShiftDirLeft {
		x = Concatenate([]*Node{xSlice, xFill}, shiftAxis)
	} else {
		x = Concatenate([]*Node{xFill, xSlice}, shiftAxis)
	}
	return x
}

// GrowLeft will grow the dimension of the given axis by concatenating n elements to the left (start).
// Those elements are filled with value (converted to the corresponding dtype).
func GrowLeft(x *Node, axis int, n int, fillValue float64) *Node {
	return growImpl(x, axis, ShiftDirRight, n, fillValue)
}

// GrowRight will grow the dimension of the given axis by concatenating n elements to the left (start).
// Those elements are filled with fillValue (converted to the corresponding dtype).
func GrowRight(x *Node, axis int, n int, fillValue float64) *Node {
	return growImpl(x, axis, ShiftDirLeft, n, fillValue)
}

func growImpl(x *Node, axis int, dir ShiftDirection, n int, fillValue float64) *Node {
	g := x.Graph()
	rank := x.Rank()
	dtype := x.DType()
	growAxis := AdjustAxisToOperandRank(x, axis)
	dims := x.Shape().Dimensions

	// Create slice to be concatenated for our desired growth: the slice is the same, independent of the direction.
	fillDims := make([]int, rank)
	for fillAxis := range rank {
		if fillAxis == growAxis {
			fillDims[fillAxis] = n
		} else {
			fillDims[fillAxis] = dims[fillAxis]
		}
	}

	// Fill slice with given value.
	var fill *Node
	if fillValue == 0 {
		fill = Zeros(g, shapes.Make(dtype, fillDims...))
	} else if fillValue == 1 {
		fill = Ones(g, shapes.Make(dtype, fillDims...))
	} else {
		fill = Scalar(g, dtype, fillValue)
		expandDims := xslices.Iota(int(0), rank)
		fill = ExpandAndBroadcast(fill, fillDims, expandDims)
	}

	if dir == ShiftDirLeft {
		x = Concatenate([]*Node{x, fill}, growAxis)
	} else {
		x = Concatenate([]*Node{fill, x}, growAxis)
	}
	return x
}