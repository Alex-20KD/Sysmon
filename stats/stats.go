package stats

import (
	"bufio"
	"os"
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
