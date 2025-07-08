// debug.go
// Paquete teggo — Utilidades de depuración de templates
// -----------------------------------------------------------------------------
// Permite compilar templates individualmente para detectar errores tempranos.

package teggo

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// DebugParseTemplates compila cada archivo individualmente y muestra errores tempranos.
// Si debug está activo, imprime confirmación en consola.
func (e *Engine) DebugParseTemplates(paths []string) error {
	rootDir := commonDir(paths)
	for _, absPath := range paths {
		src, _ := os.ReadFile(absPath)
		rel, _ := filepath.Rel(rootDir, absPath)
		logical := strings.TrimSuffix(rel, filepath.Ext(rel))
		mainName := strings.ReplaceAll(logical, string(os.PathSeparator), ".")
		base := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
		conv := ParseTagsToGoTpl(string(src), base, mainName)

		_, err := template.New(filepath.Base(absPath)).Funcs(e.funcMap(nil)).Parse(conv)
		if err != nil {
			printTemplateError(absPath, conv, err)
			return err
		}
	}
	if e.debug {
		fmt.Println("✅ Todos los templates son válidos.")
	}
	return nil
}

// printTemplateError muestra posición y contexto (±3 líneas).
func printTemplateError(path, tpl string, err error) {
	fmt.Printf("\n❌ Error en %s\n%v\n", path, err)
	type liner interface{ Line() int }
	if e, ok := err.(liner); ok {
		lines := strings.Split(tpl, "\n")
		ln := e.Line() - 1
		start := max(0, ln-2)
		end := min(len(lines), ln+3)
		for i := start; i < end; i++ {
			prefix := "   "
			if i == ln {
				prefix = " ▶ "
			}
			fmt.Printf("%s%3d | %s\n", prefix, i+1, lines[i])
		}
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
