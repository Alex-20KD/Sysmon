package main

import (
	"fmt"
	"sysmon/stats"
	"time"
)

func main() {
	// Prueba de memoria
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

	// Prueba de CPU
	fmt.Println("\n=== CPU (midiendo durante 2 segundos...) ===")
	prev, err := stats.ParseCPURaw()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	time.Sleep(2 * time.Second)
	curr, err := stats.ParseCPURaw()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	cpuPercent := stats.CalcCPUPercent(prev, curr)
	fmt.Printf("Uso CPU:    %.2f%%\n", cpuPercent)

	// Prueba de disco
	disk, err := stats.GetDiskStats("/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\n=== Disco (/) ===")
	fmt.Printf("Total:      %.2f GB\n", disk.DskTotal)
	fmt.Printf("Usado:      %.2f GB\n", disk.DskUsed)
	fmt.Printf("Libre:      %.2f GB\n", disk.DskFree)
	fmt.Printf("Uso:        %.2f%%\n", disk.DskUsedPercent)
}
