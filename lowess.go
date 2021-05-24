package lowess

import (
	"errors"
	"math"
	"sort"
)

type Coord struct {
	X float64
	Y float64
}

type CoordSlice []Coord

func (s CoordSlice) Len() int {
	return len(s)
}

func (s CoordSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s CoordSlice) Less(i, j int) bool {
	return s[i].X < s[j].X
}

type coordDist struct {
	coord Coord
	dist  float64
}

type coordDistSlice []coordDist

func (s coordDistSlice) Len() int {
	return len(s)
}

func (s coordDistSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s coordDistSlice) Less(i, j int) bool {
	return s[i].dist < s[j].dist
}

func CoordsToArrays(coords []Coord) ([]float64, []float64) {
	var xCoords []float64
	var yCoords []float64

	for i := 0; i < len(coords); i++ {
		xCoords = append(xCoords, coords[i].X)
		yCoords = append(yCoords, coords[i].Y)
	}
	return xCoords, yCoords
}

func findDist(a float64, b float64) float64 {
	return math.Abs(a - b)
}

func findMax(values []float64) float64 {
	max := 0.0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func findNearest(sortedCoords CoordSlice, targetX float64, bandwidth float64) (coordDistSlice, error) {

	if bandwidth <= 0 || bandwidth > 1 {
		return nil, errors.New("findnearest: the bandwidth must be >0 and <=1")
	}

	sort.Sort(sortedCoords)
	totalWidth := sortedCoords[len(sortedCoords)-1].X - sortedCoords[0].X
	windowWidth := bandwidth * totalWidth
	minX := targetX - windowWidth/2
	maxX := targetX + windowWidth/2

	var distances coordDistSlice
	for i := 0; i < len(sortedCoords); i++ {
		if sortedCoords[i].X >= minX && sortedCoords[i].X <= maxX {
			distances = append(distances, coordDist{
				coord: sortedCoords[i],
				dist:  findDist(targetX, sortedCoords[i].X),
			},
			)
		}
	}

	sort.Sort(distances)

	return distances, nil
}

func tricubeWeightFunction(sortedCoordDists coordDistSlice) []float64 {
	//https://uk.mathworks.com/help/curvefit/smoothing-data.html

	weights := make([]float64, len(sortedCoordDists))
	maxDist := sortedCoordDists[len(sortedCoordDists)-1].dist

	for i := 0; i < len(sortedCoordDists); i++ {
		weights[i] = math.Pow(1-math.Pow(math.Abs(sortedCoordDists[i].dist/maxDist), 3), 3)
	}
	return weights
}

func weightedMean(values []float64, weights []float64) (float64, error) {
	if len(weights) != len(values) {
		return 0, errors.New("regression: weighted mean requires equal length weight and value slices")
	}

	var sumWeights float64
	for i := 0; i < len(weights); i++ {
		sumWeights = sumWeights + weights[i]
	}

	var sumWeightedValues float64
	for i := 0; i < len(values); i++ {
		sumWeightedValues = sumWeightedValues + values[i]*weights[i]
	}

	return sumWeightedValues / sumWeights, nil
}

func wLSRegression(coordinates coordDistSlice, weights []float64) (float64, float64, error) {
	if len(weights) != len(coordinates) {
		return 0, 0, errors.New("regression: wls regressions requires coordinate and weight slices of equal length")
	}

	var xCoords []float64
	var yCoords []float64

	for i := 0; i < len(coordinates); i++ {
		xCoords = append(xCoords, coordinates[i].coord.X)
		yCoords = append(yCoords, coordinates[i].coord.Y)
	}

	weightedMeanX, err := weightedMean(xCoords, weights)
	if err != nil {
		return 0, 0, err
	}
	weightedMeanY, err := weightedMean(yCoords, weights)
	if err != nil {
		return 0, 0, err
	}

	var sumNumerator float64
	var sumDenominator float64
	for i := 0; i < len(xCoords); i++ {
		sumNumerator = sumNumerator + weights[i]*(xCoords[i]-weightedMeanX)*(yCoords[i]-weightedMeanY)
		sumDenominator = sumDenominator + weights[i]*math.Pow(xCoords[i]-weightedMeanX, 2)
	}

	// This deals with flat lines caused by identical Y values or insufficiently wide bandwidth.
	var slope float64
	if sumDenominator == 0{
		slope = 0.0
	}else{
		slope = sumNumerator / sumDenominator
	}
	var intercept = weightedMeanY - slope*weightedMeanX

	return slope, intercept, nil
}

func CalcLOESS(estimationPoints []float64, coordinates []Coord, bandwidth float64) ([]Coord, error) {
	var loessPoints []Coord

	if bandwidth <= 0 || bandwidth > 1 {
		return nil, errors.New("CalcLOESS: the bandwidth must be >0 and <=1")
	}

	// For each estimation point, calculate WLS regression line from nearest coordinates, then evaluate.
	for i := 0; i < len(estimationPoints); i++ {
		var widthCoords coordDistSlice

		// Capture coordinates within the width
		widthCoords, err := findNearest(coordinates, estimationPoints[i], bandwidth)
		if err != nil {
			return []Coord{}, err
		}


		weights := tricubeWeightFunction(widthCoords)
		slope, intercept, err := wLSRegression(widthCoords, weights)
		if err != nil {
			return []Coord{}, err
		}

		//fmt.Println("Weights:", weights)
		//fmt.Println("Slope: ", slope)
		//fmt.Println("Intercept: ", intercept)

		estimatedValue := slope*estimationPoints[i] + intercept
		//fmt.Println("\033[0;92m", estimatedValue, "\033[0m")
		loessPoints = append(loessPoints, Coord{
			X: estimationPoints[i],
			Y: estimatedValue,
		})
	}

	return loessPoints, nil
}
