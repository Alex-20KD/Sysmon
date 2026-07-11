package main

import (
	"fmt"
	"sysmon/stats"
)

func main() {
	mem, err := stats.ParseMemInfo()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== Memoria RAM ===")
	fmt.Printf("Total:      %.2f MB\n", mem.MemTotal)
	fmt.Printf("Usada:      %.2f MB\n", mem.MemUsed)
	fmt.Printf("Libre:      %.2f MB\n", mem.MemFree)
	fmt.Printf("Uso:        %.2f%%\n", mem.MemUsedPercent)
}
