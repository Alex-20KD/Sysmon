package main

import (
	"fmt"
	"sysmon/stats"
)

func main() {
	// Prueba rápida: leemos /proc/meminfo con tu función ReadLines
	lines, err := stats.ReadLines("/proc/meminfo")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Mostramos solo las primeras 5 líneas para verificar
	fmt.Println("=== Primeras 5 líneas de /proc/meminfo ===")
	for i, line := range lines {
		if i >= 5 {
			break
		}
		fmt.Println(line)
	}
	fmt.Printf("\nTotal de líneas leídas: %d\n", len(lines))
}
