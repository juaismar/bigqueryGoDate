package bigqueryGoDate

import (
	"fmt"
	"time"
)

// Evento representa un evento con diferentes tipos de fechas
type Evento struct {
	Fecha Date
}

func main() {
	// Crear un evento de prueba
	evento := Evento{
		Fecha: DateOf(time.Now()),
	}

	fmt.Println("Evento creado:")
	fmt.Printf("Fecha: %s\n", evento.Fecha)

}
