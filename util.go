package main

import "math/rand"

func randFloat(max float64) float64 { return rand.Float64() * max }
func randIntn(n int) int            { return rand.Intn(n) }
