package Funciones

import (
	"math/rand"
	"time"
)

func RandomInRange(min int32, max int32) int32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int31n(max-min+1) + min
}