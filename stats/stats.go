package stats

// En este paquete stats implementaremos toda la lógica de recolección de métricas
// del sistema: lectura de /proc/meminfo, /proc/stat y syscalls de disco.

// StatsPlaceholder es una constante temporal para evitar que el compilador de Go
// se queje de importaciones no usadas en main.go mientras construimos las funciones.
const StatsPlaceholder = 42
