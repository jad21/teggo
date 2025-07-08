// helpers.go
// Paquete teggo – Helpers puros para manipulación de props y helpers de template.
// -----------------------------------------------------------------------------
// Incluye utilidades para la creación y combinación de mapas de propiedades,
// concatenación segura de HTML, etc.

package teggo

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// Dict crea un mapa a partir de pares clave-valor, útil para pasar props a componentes.
//
//	Dict("Name", "Jad", "Age", 30) => map[string]interface{}{"Name": "Jad", "Age": 30}
func Dict(values ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values)-1; i += 2 {
		key, ok := values[i].(string)
		if !ok {
			continue // ignora claves no string
		}
		m[key] = values[i+1]
	}
	return m
}

// Merge combina dos mapas, donde las claves de m2 sobrescriben las de m1.
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

// Cat concatena todos los argumentos y los devuelve como template.HTML seguro.
//
//	Cat("Hola ", nombre, "!")  -> HTML sin escape adicional.
func Cat(parts ...interface{}) template.HTML {
	var b bytes.Buffer
	for _, p := range parts {
		fmt.Fprint(&b, p)
	}
	return template.HTML(b.String())
}

// BasicFuncMap retorna las funciones puras para uso directo en templates Go.
func BasicFuncMap() template.FuncMap {
	return template.FuncMap{
		"dict":  Dict,
		"merge": Merge,
		"cat":   Cat,
	}
}

// --- helpers de math pequeños ---
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// commonDir retorna el directorio común para varios paths absolutos.
func commonDir(paths []string) string {
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		for !strings.HasPrefix(p, dir) && dir != string(os.PathSeparator) {
			dir = filepath.Dir(dir)
		}
	}
	return dir
}
