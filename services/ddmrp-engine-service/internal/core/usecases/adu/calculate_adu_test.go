package adu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSimpleAverage_Success(t *testing.T) {
	uc := &CalculateADUUseCase{}

	givenData := []float64{100, 150, 200, 250, 300}

	result := uc.calculateSimpleAverage(givenData)

	expectedAverage := 200.0

	assert.Equal(t, expectedAverage, result)
}

func TestCalculateSimpleAverage_EmptyData(t *testing.T) {
	uc := &CalculateADUUseCase{}

	givenData := []float64{}

	result := uc.calculateSimpleAverage(givenData)

	expectedAverage := 0.0

	assert.Equal(t, expectedAverage, result)
}

func TestCalculateExponentialSmoothing_Success(t *testing.T) {
	uc := &CalculateADUUseCase{}

	givenData := []float64{100, 110, 105, 115, 120}
	givenAlpha := 0.3

	result := uc.calculateExponentialSmoothing(givenData, givenAlpha)

	assert.Greater(t, result, 100.0)
	assert.Less(t, result, 120.0)
}

func TestCalculateExponentialSmoothing_DefaultAlpha(t *testing.T) {
	uc := &CalculateADUUseCase{}

	givenData := []float64{100, 110, 105, 115, 120}
	givenInvalidAlpha := 0.0

	result := uc.calculateExponentialSmoothing(givenData, givenInvalidAlpha)

	assert.Greater(t, result, 0.0)
}

func TestCalculateWeightedMovingAverage_Success(t *testing.T) {
	uc := &CalculateADUUseCase{}

	givenData := []float64{100, 110, 120, 130, 140}

	result := uc.calculateWeightedMovingAverage(givenData)

	assert.Greater(t, result, 100.0)
	assert.LessOrEqual(t, result, 140.0)
}

func TestCalculateWeightedMovingAverage_RecentValuesWeightedHigher(t *testing.T) {
	uc := &CalculateADUUseCase{}

	givenData := []float64{50, 100, 150, 200, 250}

	result := uc.calculateWeightedMovingAverage(givenData)

	simpleAverage := uc.calculateSimpleAverage(givenData)

	assert.Greater(t, result, simpleAverage)
}
