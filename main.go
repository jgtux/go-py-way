package main

import (
	"fmt"
	"github.com/jgtux/go-py-way/core_funcs"
)

// For tests
func main() {
	keys := map[string]interface{}{
		"data": [][]float64{
			{1.0, 2.0},
			{1.5, 1.8},
			{5.0, 8.0},
			{8.0, 8.0},
			{1.0, 0.6},
			{9.0, 11.0},
		},
		"centroids": [][]float64{},
		"labels":    []int{},
	}

	recipe := `
         import numpy as np
         from sklearn.cluster import KMeans

         data_np = np.array(data)

         kmeans = KMeans(n_clusters=2, random_state=0).fit(data_np)

         centroids = kmeans.cluster_centers_.tolist()
         labels = kmeans.labels_.tolist()
       `


	mutableKeys := []string{"centroids", "labels"}

	err := core_funcs.PyWayRecipe(recipe, keys, mutableKeys)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Println("Centróides:", keys["centroids"])
	fmt.Println("Labels:", keys["labels"])
}


