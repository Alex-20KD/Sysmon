// El paquete main define el punto de entrada de nuestro ejecutable.
// En Go, cualquier programa ejecutable debe comenzar en el paquete "main"
// y definir una función "main()" que no recibe argumentos y no retorna nada.
package main

import (
	// fmt (format) se usa para formatear e imprimir texto en la consola.
	"fmt"
	
	// Importamos nuestro paquete local stats.
	// Al inicializar el módulo como "sysmon" (go mod init sysmon),
	// todas las rutas de importación internas comienzan con "sysmon/".
	"sysmon/stats"
)

func main() {
	// Imprimimos un mensaje de bienvenida para asegurar que todo compila.
	fmt.Println("Monitoreo de Sistema - Listo para iniciar")
	_ = stats.StatsPlaceholder // Evita error de importación no usada por ahora
}
