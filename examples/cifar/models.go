package cifar

import (
	"github.com/gomlx/gomlx/graph"
	"github.com/gomlx/gomlx/ml/context"
	"github.com/gomlx/gomlx/ml/layers"
	"github.com/gomlx/gomlx/ml/layers/activations"
	"github.com/gomlx/gomlx/ml/layers/batchnorm"
	"github.com/gomlx/gomlx/ml/layers/fnn"
	"github.com/gomlx/gomlx/ml/layers/kan"
)

// C10PlainModelGraph implements train.ModelFn, and returns the logit Node, given the input image.
// It's a basic FNN (Feedforward Neural Network), so no convolutions. It is meant only as an example.
func C10PlainModelGraph(ctx *context.Context, spec any, inputs []*graph.Node) []*graph.Node {
	ctx = ctx.In("model")
	batchedImages := inputs[0]
	batchSize := batchedImages.Shape().Dimensions[0]
	logits := graph.Reshape(batchedImages, batchSize, -1)
	numClasses := len(C10Labels)
	modelType := context.GetParamOr(ctx, "model", C10ValidModels[0])
	if modelType == "kan" {
		// Configuration of the KAN layer(s) use the context hyperparameters.
		logits = kan.New(ctx, logits, numClasses).Done()
	} else {
		// Configuration of the FNN layer(s) use the context hyperparameters.
		logits = fnn.New(ctx, logits, numClasses).Done()
	}
	logits.AssertDims(batchSize, numClasses)
	return []*graph.Node{logits}
}

// C10ConvolutionModelGraph implements train.ModelFn and returns the logit Node, given the input image.
// It's a straight forward CNN (Convolution Neural Network) model.
//
// This is modeled after the Keras example in Kaggle:
// https://www.kaggle.com/code/ektasharma/simple-cifar10-cnn-keras-code-with-88-accuracy
// (Thanks @ektasharma)
func C10ConvolutionModelGraph(ctx *context.Context, spec any, inputs []*graph.Node) []*graph.Node {
	ctx = ctx.In("model")
	batchedImages := inputs[0]
	g := batchedImages.Graph()
	dtype := batchedImages.DType()
	batchSize := batchedImages.Shape().Dimensions[0]
	logits := batchedImages

	layerIdx := 0
	nextCtx := func(name string) *context.Context {
		newCtx := ctx.Inf("%03d_%s", layerIdx, name)
		layerIdx++
		return newCtx
	}

	logits = layers.Convolution(nextCtx("conv"), logits).Filters(32).KernelSize(3).PadSame().Done()
	logits.AssertDims(batchSize, 32, 32, 32)
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = layers.Convolution(nextCtx("conv"), logits).Filters(32).KernelSize(3).PadSame().Done()
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = graph.MaxPool(logits).Window(2).Done()
	logits = layers.DropoutNormalize(nextCtx("dropout"), logits, graph.Scalar(g, dtype, 0.3), true)
	logits.AssertDims(batchSize, 16, 16, 32)

	logits = layers.Convolution(nextCtx("conv"), logits).Filters(64).KernelSize(3).PadSame().Done()
	logits.AssertDims(batchSize, 16, 16, 64)
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = layers.Convolution(nextCtx("conv"), logits).Filters(64).KernelSize(3).PadSame().Done()
	logits.AssertDims(batchSize, 16, 16, 64)
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = graph.MaxPool(logits).Window(2).Done()
	logits = layers.DropoutNormalize(nextCtx("dropout"), logits, graph.Scalar(g, dtype, 0.5), true)
	logits.AssertDims(batchSize, 8, 8, 64)

	logits = layers.Convolution(nextCtx("conv"), logits).Filters(128).KernelSize(3).PadSame().Done()
	logits.AssertDims(batchSize, 8, 8, 128)
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = layers.Convolution(nextCtx("conv"), logits).Filters(128).KernelSize(3).PadSame().Done()
	logits.AssertDims(batchSize, 8, 8, 128)
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = graph.MaxPool(logits).Window(2).Done()
	logits = layers.DropoutNormalize(nextCtx("dropout"), logits, graph.Scalar(g, dtype, 0.5), true)
	logits.AssertDims(batchSize, 4, 4, 128)

	// Flatten logits, and we can use the usual FNN/KAN.
	logits = graph.Reshape(logits, batchSize, -1)
	logits = layers.Dense(nextCtx("dense"), logits, true, 128)
	logits = activations.Relu(logits)
	logits = batchnorm.New(nextCtx("batchnorm"), logits, -1).Done()
	logits = layers.DropoutNormalize(nextCtx("dropout"), logits, graph.Scalar(g, dtype, 0.5), true)
	numClasses := len(C10Labels)
	logits = layers.Dense(nextCtx("dense"), logits, true, numClasses)
	return []*graph.Node{logits}
}