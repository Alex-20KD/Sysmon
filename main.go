package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sysmon/stats"
	"time"
)

// ─── COLORES ANSI ───────────────────────────────────────────────────────
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	BgGray = "\033[48;5;236m"
)

func colorByPercent(pct float64) string {
	if pct >= 85 {
		return Red
	}
	if pct >= 60 {
		return Yellow
	}
	return Green
}

func progressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := int(pct / 100 * float64(width))
	empty := width - filled

	color := colorByPercent(pct)

	bar := color
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	bar += Reset + Dim
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += Reset

	return bar
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printHeader() {
	now := time.Now().Format("15:04:05")
	fmt.Println()
	fmt.Printf("  %s%s ⚡ SYSMON — Monitor de Sistema %s  %s%s%s\n",
		Bold, Cyan, Cyan, Dim, now, Reset)
	fmt.Printf("  %s─────────────────────────────────────────────%s\n", Dim, Reset)
}

func printCPU(pct float64) {
	color := colorByPercent(pct)
	fmt.Println()
	fmt.Printf("  %s%s  CPU%s\n", Bold, Blue, Reset)
	fmt.Printf("    %s %s%6.2f%%%s\n", progressBar(pct, 30), color, pct, Reset)
}

func printMemory(mem stats.MemStats) {
	color := colorByPercent(mem.MemUsedPercent)
	fmt.Println()
	fmt.Printf("  %s%s  MEMORIA RAM%s\n", Bold, Blue, Reset)
	fmt.Printf("    %s %s%6.2f%%%s\n", progressBar(mem.MemUsedPercent, 30), color, mem.MemUsedPercent, Reset)
	fmt.Printf("    %s%-12s%s %8.2f MB   %s%-12s%s %8.2f MB\n",
		Dim, "Total:", Reset, mem.MemTotal,
		Dim, "Usada:", Reset, mem.MemUsed)
	fmt.Printf("    %s%-12s%s %8.2f MB\n",
		Dim, "Libre:", Reset, mem.MemFree)
}

func printDisk(disk stats.DiskStats) {
	color := colorByPercent(disk.DskUsedPercent)
	fmt.Println()
	fmt.Printf("  %s%s  DISCO (/)%s\n", Bold, Blue, Reset)
	fmt.Printf("    %s %s%6.2f%%%s\n", progressBar(disk.DskUsedPercent, 30), color, disk.DskUsedPercent, Reset)
	fmt.Printf("    %s%-12s%s %8.2f GB   %s%-12s%s %8.2f GB\n",
		Dim, "Total:", Reset, disk.DskTotal,
		Dim, "Usado:", Reset, disk.DskUsed)
	fmt.Printf("    %s%-12s%s %8.2f GB\n",
		Dim, "Libre:", Reset, disk.DskFree)
}

// Modificamos el footer para que refleje dinámicamente el intervalo configurado
func printFooter(intervalSec int) {
	fmt.Println()
	fmt.Printf("  %s─────────────────────────────────────────────%s\n", Dim, Reset)
	fmt.Printf("  %s  Ctrl+C para salir  •  Refresco: %ds%s\n", Dim, intervalSec, Reset)
	fmt.Println()
}

func main() {
	// ─── FLAG DE CLI ──────────────────────────────────────────────────────
	intervalSec := flag.Int("interval", 2, "Intervalo de refresco en segundos")
	flag.Parse()

	interval := time.Duration(*intervalSec) * time.Second

	// 1. Crear el contexto que intercepta Ctrl+C (SIGINT) de forma limpia
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop() // Libera los recursos del listener de señales al salir

	// 2. Snapshot inicial de la CPU
	prev, err := stats.ParseCPURaw()
	if err != nil {
		log.Fatalf("Error al inicializar el monitor de CPU: %v", err)
	}

	// 3. Crear el ticker con el intervalo del flag
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 4. El loop multiplexado con 'select' para escuchar ambos canales
	for {
		select {
		case <-ctx.Done():
			// 🛑 Caso Ctrl+C: Limpiamos pantalla antes de despedirnos de forma ordenada
			clearScreen()
			fmt.Println()
			fmt.Printf("  %s%s👋 Cerrando el monitor limpiamente... ¡Hasta luego!%s\n", Bold, Yellow, Reset)
			fmt.Println()
			return // Finaliza main de inmediato ejecutando los defers en cadena

		case <-ticker.C:
			// ⏱️ Caso Ticker: Recolección y renderizado normal cada N segundos
			curr, err := stats.ParseCPURaw()
			if err != nil {
				log.Printf("Error al parsear CPU raw: %v", err)
				continue
			}

			mem, err := stats.ParseMemInfo()
			if err != nil {
				log.Printf("Error al leer memoria: %v", err)
				continue
			}

			disk, err := stats.GetDiskStats("/")
			if err != nil {
				log.Printf("Error al leer disco: %v", err)
				continue
			}

			// Limpiar pantalla e imprimir toda la UI hermosa con esteroides visuales
			clearScreen()
			cpuPercent := stats.CalcCPUPercent(prev, curr)

			printHeader()
			printCPU(cpuPercent)
			printMemory(mem)
			printDisk(disk)
			printFooter(*intervalSec)

			// Relevo de estados para el siguiente tick
			prev = curr
		}
	}
}
