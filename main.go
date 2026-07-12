package main

import (
	"fmt"
	"log"
	"sysmon/stats"
	"time"
)

// ─── COLORES ANSI ───────────────────────────────────────────────────────
// Las terminales modernas interpretan secuencias "\033[Xm" como instrucciones
// de formato. \033 es el carácter ESC, [Xm indica el estilo/color.
// Al final de cada texto coloreado, usamos "\033[0m" (Reset) para volver
// al color normal y no "contaminar" el texto siguiente.
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"

	// Colores de texto (foreground)
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	White  = "\033[37m"

	// Fondo oscuro para la barra de progreso vacía
	BgGray = "\033[48;5;236m"
)

// colorByPercent devuelve un color ANSI según el nivel de uso:
//   - Verde: < 60%  → todo tranquilo
//   - Amarillo: 60-85% → atención
//   - Rojo: > 85%  → alerta
//
// Esto usa la técnica de "early return" que es idiomática en Go:
// en lugar de if/else if/else, cada condición retorna directamente.
func colorByPercent(pct float64) string {
	if pct >= 85 {
		return Red
	}
	if pct >= 60 {
		return Yellow
	}
	return Green
}

// progressBar genera una barra visual de ancho fijo usando caracteres Unicode.
// Ejemplo: "████████░░░░░░░░" (8 llenos de 16 = 50%)
//
// El formato funciona así:
// - Calculamos cuántos bloques "llenar" proporcionalmente al porcentaje.
// - Los llenos (█) se colorean según el nivel de uso.
// - Los vacíos (░) se muestran en gris tenue.
func progressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	// int(pct / 100 * width) calcula cuántos caracteres llenar.
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

// clearScreen envía las secuencias ANSI para limpiar la pantalla
// y mover el cursor a la esquina superior izquierda.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// printHeader imprime el encabezado del monitor con la hora actual.
// time.Now().Format() usa un formato de referencia peculiar de Go:
// en vez de "HH:MM:SS", Go usa la fecha de referencia "Mon Jan 2 15:04:05 2006"
// y tú reemplazas los números con el formato que quieres. Es raro pero poderoso.
func printHeader() {
	now := time.Now().Format("15:04:05")
	fmt.Println()
	fmt.Printf("  %s%s ⚡ SYSMON — Monitor de Sistema %s  %s%s%s\n",
		Bold, Cyan, Cyan, Dim, now, Reset)
	fmt.Printf("  %s─────────────────────────────────────────────%s\n", Dim, Reset)
}

// printCPU muestra la sección de CPU con barra de progreso y color.
// %-10s  → alinea el texto a la izquierda en un campo de 10 caracteres.
// %6.2f  → número con 6 caracteres de ancho total y 2 decimales.
// %%     → imprime un "%" literal (porque % es especial en Printf).
func printCPU(pct float64) {
	color := colorByPercent(pct)
	fmt.Println()
	fmt.Printf("  %s%s  CPU%s\n", Bold, Blue, Reset)
	fmt.Printf("    %s %s%6.2f%%%s\n", progressBar(pct, 30), color, pct, Reset)
}

// printMemory muestra la sección de RAM con detalles alineados en columnas.
// El truco de las columnas:
//   fmt.Printf("  %-12s %8.2f MB", "Etiqueta:", valor)
//   %-12s  → "Etiqueta:" ocupa 12 chars, alineado a la izquierda
//   %8.2f  → el número ocupa 8 chars, alineado a la derecha, 2 decimales
// Esto hace que todos los números queden alineados verticalmente.
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

// printDisk muestra la sección de disco con el mismo formato.
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

// printFooter muestra el pie con instrucciones para el usuario.
func printFooter() {
	fmt.Println()
	fmt.Printf("  %s─────────────────────────────────────────────%s\n", Dim, Reset)
	fmt.Printf("  %s  Ctrl+C para salir  •  Refresco: 2s%s\n", Dim, Reset)
	fmt.Println()
}

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

		// d. Limpiar pantalla ANTES de imprimir
		clearScreen()

		// e. Calcular % CPU
		cpuPercent := stats.CalcCPUPercent(prev, curr)

		// f. Imprimir el panel bonito
		printHeader()
		printCPU(cpuPercent)
		printMemory(mem)
		printDisk(disk)
		printFooter()

		// g. Relevo: el snapshot actual se convierte en el anterior
		prev = curr
	}
}
