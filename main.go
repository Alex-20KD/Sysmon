package main

import (
	"fmt"
	"log"
	"sysmon/stats"
	"time"
)

func main() {
	// 1. Snapshot inicial de la CPU (antes de entrar al bucle)
	prev, err := stats.ParseCPURaw()
	if err != nil {
		log.Fatalf("Error al inicializar el monitor de CPU: %v", err)
	}

	// 2. Crear el ticker para que itere cada 2 segundos exactos
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 3. Loop infinito de monitoreo en tiempo real
	for range ticker.C {
		// a. Nuevo snapshot de CPU
		curr, err := stats.ParseCPURaw()
		if err != nil {
			log.Printf("Error al parsear CPU raw: %v", err)
			continue
		}

		// b. Leer memoria RAM
		mem, err := stats.ParseMemInfo()
		if err != nil {
			log.Printf("Error al leer memoria: %v", err)
			continue
		}

		// c. Leer espacio en disco
		disk, err := stats.GetDiskStats("/")
		if err != nil {
			log.Printf("Error al leer disco: %v", err)
			continue
		}

		// d. Limpiar pantalla ANTES de imprimir los nuevos datos
		fmt.Print("\033[H\033[2J")

		// e. Calcular % CPU usando las deltas de tiempo
		cpuPercent := stats.CalcCPUPercent(prev, curr)

		// f. Imprimir el panel de control unificado
		fmt.Println("========================================")
		fmt.Println("      MONITOR DE SISTEMA EN TIEMPO REAL ")
		fmt.Println("========================================")

		fmt.Printf(" CPU \n  Uso:        %.2f%%\n\n", cpuPercent)

		fmt.Printf(" MEMORIA RAM \n")
		fmt.Printf("  Total:      %.2f MB\n", mem.MemTotal)
		fmt.Printf("  Usada:      %.2f MB\n", mem.MemUsed)
		fmt.Printf("  Libre:      %.2f MB\n", mem.MemFree)
		fmt.Printf("  Uso:        %.2f%%\n\n", mem.MemUsedPercent)

		fmt.Printf(" DISCO (/) \n")
		fmt.Printf("  Total:      %.2f GB\n", disk.DskTotal)
		fmt.Printf("  Usado:      %.2f GB\n", disk.DskUsed)
		fmt.Printf("  Libre:      %.2f GB\n", disk.DskFree)
		fmt.Printf("  Uso:        %.2f%%\n", disk.DskUsedPercent)
		fmt.Println("========================================")

		// g. Relevo de guardia: el presente se convierte en el pasado para el próximo tick
		prev = curr
	}
}
