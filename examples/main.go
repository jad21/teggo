package main

import (
	"fmt"
	"os"

	"github.com/jad21/teggo"
)

func main() {
	// 1. Lista de archivos de templates a cargar
	files := teggo.Discover("./examples", "*.html") // helper opcional
	fmt.Println("Templates detectados:")            // ← inspección rápida
	for _, t := range files {                       // ← método helper opcional
		fmt.Println(" •", t)
	}

	// 2. Inicializa el engine Teggo
	engine, err := teggo.NewEngine(files, true)
	if err != nil {
		panic(fmt.Sprintf("Error inicializando Teggo: %v", err))
	}
	// fmt.Println("TemplateNames:")              // ← inspección rápida
	// for _, t := range engine.TemplateNames() { // ← método helper opcional
	// 	fmt.Println(" •", t)
	// }

	// Usa el debugger para parsear e identificar errores
	if err := engine.DebugParseTemplates(files); err != nil {
		panic(fmt.Sprintf("Error renderizando[0]: %v", err))
	}

	// 3. Datos de ejemplo
	data := map[string]interface{}{
		"IsAdmin": true,
		"Users": []map[string]interface{}{
			{"Name": "Jad", "Email": "jad@teggo.com"},
			{"Name": "Ana", "Email": "ana@teggo.com"},
		},
	}

	// 4. Renderiza la página principal (Home)
	err = engine.Render("pages.Home", data, os.Stdout)
	if err != nil {
		panic(fmt.Sprintf("Error renderizando: %v", err))
	}
}
