package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type K_Means struct {
	k, max_iter int
	tol         float64
}

var wg sync.WaitGroup
var centroids map[int][]float64
var classifications map[int][][]float64

func min_Pos(slice []float64) int {
	minValue := slice[0]
	minPos := 0
	for i, v := range slice {
		if v < minValue {
			minValue = v
			minPos = i
		}
	}

	return minPos
}

func average(slice [][]float64) []float64 {

	temp := make([]float64, len(slice[0]))

	for _, v := range slice {
		for i := 0; i < len(v); i++ {
			temp[i] = temp[i] + v[i]
		}

	}

	for i := 0; i < len(slice[0]); i++ {
		temp[i] = temp[i] / float64(len(slice))
	}

	return temp

}

func norm(slice []float64) float64 {
	sum := 0.0
	for _, v := range slice {
		sum += math.Pow(v, 2)
	}
	return math.Sqrt(sum)
}

func restar(firstSlice, secondSlice []float64) []float64 {
	temp := make([]float64, len(firstSlice))
	for i := 0; i < len(firstSlice); i++ {
		temp[i] = firstSlice[i] - secondSlice[i]
	}
	return temp
}

func sum(firstSlice, secondSlice []float64) float64 {

	temp := make([]float64, len(firstSlice))

	for i := 0; i < len(firstSlice); i++ {
		temp[i] = firstSlice[i] - secondSlice[i]
	}

	for i := 0; i < len(firstSlice); i++ {
		temp[i] = temp[i] / (secondSlice[i] * 100.0)
	}

	sum := 0.0
	for _, v := range temp {
		sum += v * 10000.0
	}
	return sum
}

func Team(v []float64) {

	defer wg.Done()
	distances := make([]float64, len(centroids))
	for c := 0; c < len(centroids); c++ {
		sleep := rand.Int63n(1000)
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		distances[c] = norm(restar(v, centroids[c]))
	}
	classification := min_Pos(distances)
	classifications[classification] = append(classifications[classification], v)

}

func (KMeans *K_Means) Fit(data [][]float64) {

	centroids = make(map[int][]float64)

	for i := 0; i < KMeans.k; i++ {

		centroids[i] = data[i]
	}

	for i := 0; i < KMeans.max_iter; i++ {

		classifications = make(map[int][][]float64)
		for r := 0; r < KMeans.k; r++ {

			classifications[r] = make([][]float64, 0)
		}

		wg.Add(len(data))

		for _, v := range data {
			go Team(v)
		}

		wg.Wait()

		prev_centroids := make(map[int][]float64)

		for key, value := range centroids {
			prev_centroids[key] = value
		}

		for c := 0; c < len(classifications); c++ {
			centroids[c] = average(classifications[c])
		}

		optimized := true

		for c := 0; c < len(centroids); c++ {
			original_centroid := prev_centroids[c]
			current_centroid := centroids[c]

			if sum(current_centroid, original_centroid) > KMeans.tol {
				optimized = false
			}
		}

		if optimized == true {
			break
		}
	}

}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 << 20)

	Kvalue := r.FormValue("Kvalue")
	intKValue, _ := strconv.Atoi(Kvalue)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	csvFile, _ := handler.Open()
	reader := csv.NewReader(csvFile)
	filedata, _ := reader.ReadAll()
	data := make([][]float64, 0)

	for _, value := range filedata {
		FirstValue, _ := strconv.ParseFloat(value[0], 64)
		SecondValue, _ := strconv.ParseFloat(value[1], 64)
		temporal := make([]float64, 0)
		temporal = append(temporal, FirstValue, SecondValue)
		data = append(data, temporal)
	}

	KMeans := &K_Means{k: intKValue, tol: 0.001, max_iter: 500}
	KMeans.Fit(data)

	for i := 0; i < len(classifications); i++ {
		fmt.Fprintf(w, "Grupo = %d\n", i+1)

		for _, v := range classifications[i] {
			fmt.Fprintln(w, v)
		}
		fmt.Fprintf(w, "\n")
	}

}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	setupRoutes()
}
