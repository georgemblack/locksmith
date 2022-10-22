package main

import (
	"github.com/georgemblack/locksmith"
)

const baseURL = "http://localhost:8200"

func main() {
	locksmith.GetRekeyStatus(baseURL)
}
