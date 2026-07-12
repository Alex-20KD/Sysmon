package stats

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// En este paquete stats implementaremos toda la lógica de recolección de métricas
// del sistema: lectura de /proc/meminfo, /proc/stat y syscalls de disco.

// ReadLines abre el archivo en la ruta indicada, lee su contenido
// línea por línea y retorna un slice de strings.
// Garantiza el cierre del archivo con defer.

func ReadLines(path string) ([]string, error) {

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
	MemTotal       float64
	MemUsed        float64
	MemFree        float64
	MemUsedPercent float64
}

// CPURaw contiene los contadores acumulados crudos de /proc/stat.
type CPURaw struct {
	User, Nice, System, Idle, IOWait, Irq, SoftIrq, Steal uint64
}

// DiskStats contiene los datos de uso de disco.
type DiskStats struct {
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

	// 🛠️ Aquí es donde se arregló la sintaxis (se quitaron las llaves extra)
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

func ParseCPURaw() (CPURaw, error) {
	lines, err := ReadLines("/proc/stat")
	if err != nil || len(lines) == 0 {
		return CPURaw{}, fmt.Errorf("error al leer /proc/stat: %w", err)
	}

	parts := strings.Fields(lines[0])
	// Validación de seguridad para asegurarnos de que la línea tiene suficientes columnas
	if len(parts) < 9 {
		return CPURaw{}, fmt.Errorf("formato inválido en /proc/stat")
	}

	// Parseamos uno por uno...
	user, _ := strconv.ParseUint(parts[1], 10, 64)
	nice, _ := strconv.ParseUint(parts[2], 10, 64)
	system, _ := strconv.ParseUint(parts[3], 10, 64)
	idle, _ := strconv.ParseUint(parts[4], 10, 64)
	iowait, _ := strconv.ParseUint(parts[5], 10, 64)
	irq, _ := strconv.ParseUint(parts[6], 10, 64)
	softirq, _ := strconv.ParseUint(parts[7], 10, 64)
	steal, _ := strconv.ParseUint(parts[8], 10, 64)

	return CPURaw{
		User: user, Nice: nice, System: system, Idle: idle,
		IOWait: iowait, Irq: irq, SoftIrq: softirq, Steal: steal,
	}, nil
}

func CalcCPUPercent(prev, curr CPURaw) float64 {
	// 1. Sumar todos los contadores de cada snapshot
	totalPrev := prev.User + prev.Nice + prev.System + prev.Idle + prev.IOWait + prev.Irq + prev.SoftIrq + prev.Steal
	totalCurr := curr.User + curr.Nice + curr.System + curr.Idle + curr.IOWait + curr.Irq + curr.SoftIrq + curr.Steal

	// 2. Calcular diferencias
	totalDelta := totalCurr - totalPrev
	idleDelta := (curr.Idle + curr.IOWait) - (prev.Idle + prev.IOWait)
	// 3. Candado de seguridad
	if totalDelta == 0 {
		return 0.0
	}

	// 4. Fórmula: (Total - Idle) / Total * 100
	utilizadoDelta := totalDelta - idleDelta
	return (float64(utilizadoDelta) / float64(totalDelta)) * 100
}

// GetDiskStats obtiene el espacio en disco del filesystem montado en la ruta dada.
// Usa la syscall statfs, que es la misma que usa el comando "df" de Linux.
//
// ¿Cómo funciona?
// El kernel organiza el disco en "bloques" de tamaño fijo (generalmente 4096 bytes).
// statfs nos dice:
//   - Bsize:   tamaño de cada bloque en bytes
//   - Blocks:  número total de bloques en el filesystem
//   - Bfree:   bloques libres (incluye los reservados para root)
//   - Bavail:  bloques disponibles para usuarios normales (sin los reservados)
//
// Usamos Bavail (no Bfree) porque Bfree incluye bloques reservados para el
// superusuario (root) que un programa normal no puede usar. Bavail es lo que
// realmente tienes disponible, igual que lo que muestra "df -h".
func GetDiskStats(path string) (DiskStats, error) {
	// syscall.Statfs_t es un struct que el kernel llena con info del filesystem.
	// Lo declaramos vacío y le pasamos un puntero (&fs) para que el kernel lo llene.
	var fs syscall.Statfs_t

	// syscall.Statfs(path, &fs) hace la llamada al kernel.
	// "path" puede ser cualquier ruta dentro del filesystem que quieres consultar.
	// Normalmente usamos "/" para el filesystem raíz.
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return DiskStats{}, fmt.Errorf("error al consultar disco en %s: %w", path, err)
	}

	// Calculamos los tamaños multiplicando bloques × tamaño de bloque.
	// uint64(fs.Bsize) convierte el tamaño de bloque (que es int64) a uint64
	// para que sea compatible con Blocks/Bavail (que ya son uint64).
	totalBytes := fs.Blocks * uint64(fs.Bsize)  // total del filesystem
	freeBytes := fs.Bavail * uint64(fs.Bsize)   // disponible para el usuario
	usedBytes := totalBytes - freeBytes          // usado

	// Convertimos de bytes a gigabytes (÷ 1024³) para que sea legible.
	const bytesPerGB = 1024.0 * 1024.0 * 1024.0
	totalGB := float64(totalBytes) / bytesPerGB
	freeGB := float64(freeBytes) / bytesPerGB
	usedGB := float64(usedBytes) / bytesPerGB

	// Protección contra división por cero (disco de tamaño 0 sería muy raro).
	var percent float64
	if totalGB > 0 {
		percent = (usedGB / totalGB) * 100
	}

	return DiskStats{
		DskTotal:       totalGB,
		DskUsed:        usedGB,
		DskFree:        freeGB,
		DskUsedPercent: percent,
	}, nil
}
