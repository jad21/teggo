// Teggo — JSX‑like Components for Go
// helpers.go – utilidades puras sin dependencias del engine
// -----------------------------------------------------------------------------
// • Dict   → crea rápidamente mapas clave/valor para las props.
// • Merge  → combina dos mapas (m2 pisa claves de m1).
// • Cat    → concatena cualquier número de partes convirtiéndolas a string.
// -----------------------------------------------------------------------------

package teggo

import (
	"bytes"
	"fmt"
	"html/template"
)

// Dict convierte una lista variádica clave‑valor
// Dict("Name", "Jad", "Age", 30) => map[string]interface{}{ "Name": "Jad", "Age": 30 }
func Dict(values ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key := values[i].(string)
		m[key] = values[i+1]
	}
	return m
}

// Merge une dos mapas — las claves de m2 sobrescriben las de m1.
func Merge(m1, m2 map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{}, len(m1)+len(m2))
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}
	return res
}

// Cat concatena todos los argumentos y los devuelve como template.HTML.
//
//	Cat("Hola ", nombre, "!")  -> HTML seguro sin escapar nuevamente.
func Cat(parts ...interface{}) template.HTML {
	var b bytes.Buffer
	for _, p := range parts {
		fmt.Fprint(&b, p)
	}
	return template.HTML(b.String())
}

// -----------------------------
//  FuncMap público
// -----------------------------
//   La función está en engine.go porque requiere acceso al flag debug
//   y a los métodos internos de *Engine*.  Aquí dejamos un envoltorio
//   para quien necesite los helpers sin instanciar Engine.

// BasicFuncMap devuelve solo las utilidades puras (sin partial).
func BasicFuncMap() template.FuncMap {
	return template.FuncMap{
		"dict":  Dict,
		"merge": Merge,
		"cat":   Cat,
	}
}
