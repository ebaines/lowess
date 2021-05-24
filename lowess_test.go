package lowess

import (
	"encoding/csv"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"strconv"
	"testing"
)

func TestCoordsToArrays(t *testing.T) {
	coordinates := []Coord{
		{
			X: 0.5578196,
			Y: 18.63654,
		},
		{
			X: 2.0217271,
			Y: 103.49646,
		},
		{
			X: 2.5773252,
			Y: 150.35391,
		},
		{
			X: 3.4140288,
			Y: 190.51031,
		},
		{
			X: 4.3014084,
			Y: 208.70115,
		},
		{
			X: 4.7448394,
			Y: 213.71135,
		},
		{
			X: 5.1073781,
			Y: 228.49353,
		},
	}

	xPoints, yPoints := CoordsToArrays(coordinates)
	correctX := []float64{0.5578196, 2.0217271, 2.5773252, 3.4140288, 4.3014084, 4.7448394, 5.1073781}
	correctY := []float64{18.63654, 103.49646, 150.35391, 190.51031, 208.70115, 213.71135, 228.49353}
	for i := 0; i < len(xPoints); i++ {
		if xPoints[i] != correctX[i] {
			t.Errorf("CoordsToArrays returned %v not %v", xPoints[i], correctX[i])
		}
	}
	for i := 0; i < len(xPoints); i++ {
		if yPoints[i] != correctY[i] {
			t.Errorf("CoordsToArrays returned %v not %v", yPoints[i], correctY[i])
		}
	}
}

func TestFindDistance(t *testing.T) {
	coordinates := []Coord{
		{
			X: 0.5578196,
			Y: 18.63654,
		},
		{
			X: 2.0217271,
			Y: 103.49646,
		},
		{
			X: 2.5773252,
			Y: 150.35391,
		},
		{
			X: 3.4140288,
			Y: 190.51031,
		},
		{
			X: 4.3014084,
			Y: 208.70115,
		},
		{
			X: 4.7448394,
			Y: 213.71135,
		},
		{
			X: 5.1073781,
			Y: 228.49353,
		},
	}

	correctDistances := []float64{
		0.000000,
		1.463908,
		2.019506,
		2.856209,
		3.743589,
		4.187020,
		4.549559,
	}

	xPoints, _ := CoordsToArrays(coordinates)
	for i := 0; i < len(xPoints); i++ {
		distance := findDist(xPoints[0], xPoints[i])
		if math.Round(distance*1000000)/1000000 != correctDistances[i] {
			t.Errorf("findDist returned %v not %v", math.Round(distance*1000000)/1000000, correctDistances[i])
		}
	}

}

func TestFindMax(t *testing.T) {
	values := []float64{82, 82, 82, 82, 82, 82, 82, 81, 81, 83}
	max := findMax(values)

	if max != 83 {
		t.Error()
	}

	values = []float64{-2, -1, 0, 1, 2}
	max = findMax(values)

	if max != 2 {
		t.Error()
	}

	values = []float64{0, 2, -1, 1, -2}
	max = findMax(values)

	if max != 2 {
		t.Error()
	}

	values = []float64{0, -3, -1, -4, -2}
	max = findMax(values)

	if max != 0 {
		t.Error()
	}
}

func TestFindNearest(t *testing.T) {
	assert := assert.New(t)
	_, err := findNearest(CoordSlice{}, 0.0, -0.1)
	assert.NotEqual(nil, err)
}

func TestTricubeWeightFunction(t *testing.T) {
	coordinates := coordDistSlice{
		{
			coord: Coord{
				X: 0.5578196,
				Y: 18.63654,
			},
			dist: 0.,
		},
		{
			coord: Coord{
				X: 2.0217271,
				Y: 103.49646,
			},
			dist: 1.463908,
		},
		{
			coord: Coord{
				X: 2.5773252,
				Y: 150.35391,
			},
			dist: 2.019506,
		},
		{
			coord: Coord{
				X: 3.4140288,
				Y: 190.51031,
			},
			dist: 2.856209,
		},
		{
			coord: Coord{
				X: 4.3014084,
				Y: 208.70115,
			},
			dist: 3.743589,
		},
		{
			coord: Coord{
				X: 4.7448394,
				Y: 213.71135,
			},
			dist: 4.187020,
		},
		{
			coord: Coord{
				X: 5.1073781,
				Y: 228.49353,
			},
			dist: 4.549559,
		},
	}
	correct := []float64{1, 0.903349061506753, 0.7598896621661493, 0.4262173725153165, 0.0868617633725999, 0.01072309994722968, 0}
	weights := tricubeWeightFunction(coordinates)
	for i := 0; i < len(weights); i++ {
		assert.Equal(t, math.Round(correct[i]*10000), math.Round(weights[i]*10000))
	}
}

func TestCalcLOESS(t *testing.T) {

	correctAnswers := make(map[float64][]float64)
	bandwidths := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}

	for i := 0; i < len(bandwidths); i++ {
		correctAnswers[bandwidths[i]] = make([]float64, 0, 100)
	}

	file, err := os.Open("test.csv")
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()

	var x, y []float64

	for i, row := range records {
		if i > 0 {
			float, _ := strconv.ParseFloat(row[1], 64)
			x = append(x, float)
			float, _ = strconv.ParseFloat(row[2], 64)
			y = append(y, float)

			for i := 0; i < len(bandwidths); i++ {
				float, _ = strconv.ParseFloat(row[3+i], 64)
				correctAnswers[bandwidths[i]] = append(correctAnswers[bandwidths[i]], float)
			}
		}
	}

	var coords CoordSlice

	for i := 0; i < len(x); i++ {
		coords = append(coords, Coord{
			X: x[i],
			Y: y[i],
		})
	}

	for _, bandwidth := range bandwidths {

		loessPoints, _ := CalcLOESS(x, coords, bandwidth)
		_, yPoints := CoordsToArrays(loessPoints)

		for i := 0; i < len(yPoints); i++ {
			//if math.Round(yPoints[i]*1000)/1000 != math.Round(correct[i]*1000)/1000 {
			if math.Abs(yPoints[i]-correctAnswers[bandwidth][i]) > 0.01 {
				t.Errorf("Loess returned %v not %v --- %v", yPoints[i], correctAnswers[bandwidth][i], math.Abs(yPoints[i]-correctAnswers[bandwidth][i]))
			}
		}
	}
}
