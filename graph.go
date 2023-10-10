package main
import (
   "fmt"
)

const n = 4

// Graph represents a graph using an adjacency matrix
type Graph struct {
   AdMatrix [n][n]bool
}

func main() {
   // Create graph
   graph := Graph{}

   // Connect nodes
   graph.AdMatrix[0][1] = true
   graph.AdMatrix[0][2] = true
   graph.AdMatrix[1][2] = true

   // Print graph
   fmt.Println("The graph is printed as follows using adjacency matrix:")
   fmt.Println("Adjacency Matrix:")
   for i := 0; i < n; i++ {
      for j := 0; j < n; j++ {
         fmt.Printf("%t ", graph.AdMatrix[i][j])
      }
      fmt.Println()
   }
}
