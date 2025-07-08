// doc.go
// Paquete teggo: Componentes tipo JSX para Go
//
// Teggo permite escribir componentes de UI y páginas usando sintaxis de tags similar a JSX/React,
// y los transpila automáticamente a templates estándar de Go para renderizado seguro y eficiente.
//
// Ejemplo básico:
//
//	import (
//		"github.com/jad21/teggo"
//	)
//	files := teggo.Discover("./components", "*.gotpl")
//	engine, err := teggo.NewEngine(files, true)
//	if err != nil { /* manejar error */ }
//
//	// Renderizar template:
//	err = engine.Render("pages.Home", data, os.Stdout)
package teggo
