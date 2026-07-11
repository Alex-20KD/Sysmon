package stats

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

// En este paquete stats implementaremos toda la lógica de recolección de métricas
// del sistema: lectura de /proc/meminfo, /proc/stat y syscalls de disco.

// StatsPlaceholder es una constante temporal para evitar que el compilador de Go
// se queje de importaciones no usadas en main.go mientras construimos las funciones.
const StatsPlaceholder = 42

// ReadLines abre el archivo en la ruta indicada, lee su contenido
// línea por línea y retorna un slice de strings.
// Garantiza el cierre del archivo con defer.

func ReadLines(path string) ([]string, error) {
	// TODO: tu código aquí
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// MemStats contiene los datos de uso de memoria RAM.
type MemStats struct {
	// TODO: campos para total, usada, libre (en MB) y porcentaje de uso
	MemTotal       float64
	MemUsed        float64
	MemFree        float64
	MemUsedPercent float64
}

// CPURaw contiene los contadores acumulados crudos de /proc/stat.
type CPURaw struct {
	// TODO: campos para user, nice, system, idle, iowait, irq, softirq, steal
	User, Nice, System, Idle, IOWait, Irq, SoftIrq, Steal uint64
}

// DiskStats contiene los datos de uso de disco.
type DiskStats struct {
	// TODO: campos para total, usado, libre (en GB o MB) y porcentaje de uso
	DskTotal       float64
	DskUsed        float64
	DskFree        float64
	DskUsedPercent float64
}

// ParseMemInfo lee /proc/meminfo y retorna un MemStats con los valores en MB.
func ParseMemInfo() (MemStats, error) {
	lines, err := ReadLines("/proc/meminfo")
	if err != nil {
		return MemStats{}, fmt.Errorf("error al leer /proc/meminfo: %w", err)
	}
	var totalKB, freeKB, availableKB float64

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return MemStats{}, fmt.Errorf("error al parsear MemTotal: %w", err)
			}
			totalKB = val
		} else if strings.HasPrefix(line, "MemFree:") {
			parts := strings.Fields(line)
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return MemStats{}, fmt.Errorf("error al parsear MemFree: %w", err)
			}
			freeKB = val
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return MemStats{}, fmt.Errorf("error al parsear MemAvailable: %w", err)
			}
			availableKB = val
		}
	}
	
	totalMB := totalKB / 1024.0
	freeMB := freeKB / 1024.0
	availableMB := availableKB / 1024.0
	usedMB := totalMB - availableMB
	var percent float64
	if totalMB > 0 {
		percent = (usedMB / totalMB) * 100
	}
	return MemStats{
		MemTotal:       totalMB,
		MemUsed:        usedMB,
		MemFree:        freeMB,
		MemUsedPercent: percent,
	}, nil
}
