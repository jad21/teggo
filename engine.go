// engine.go
// Paquete teggo — Núcleo de compilación y renderizado de templates tipo JSX/Go.
// -----------------------------------------------------------------------------
// Proporciona la estructura Engine, compilador y renderizador thread-safe.

package teggo

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Engine mantiene el set de templates compilado y la bandera de debug.
type Engine struct {
	base  *template.Template // Set base, nunca ejecutar ni clonar luego de ejecutar.
	debug bool
}

// NewEngine compila todos los archivos indicados en paths en un set lógico único.
// Registra nombres lógicos tipo «pages.Home». Devuelve el engine listo para renderizar.
func NewEngine(paths []string, debug bool) (*Engine, error) {
	e := &Engine{debug: debug}
	rootDir := commonDir(paths)

	var sb strings.Builder
	for _, absPath := range paths {
		srcBytes, err := os.ReadFile(absPath)
		if err != nil {
			return nil, err
		}

		// Nombre lógico tipo "pages.Home"
		rel, _ := filepath.Rel(rootDir, absPath)
		logical := strings.TrimSuffix(rel, filepath.Ext(rel))
		mainName := strings.ReplaceAll(logical, string(os.PathSeparator), ".")
		base := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))

		converted := ParseTagsToGoTpl(string(srcBytes), base, mainName)
		sb.WriteString(converted + "\n")
	}

	fmt.Println(sb.String())

	// Crea el set principal (root)
	root := template.New("root")
	root.Funcs(e.funcMap(root))
	baseSet, err := root.Parse(sb.String())
	if err != nil {
		return nil, err
	}
	e.base = baseSet // Nunca ejecutar ni clonar luego de ejecución.
	return e, nil
}

// Render clona el set y ejecuta el template indicado, seguro para concurrencia.
func (e *Engine) Render(name string, data any, w io.Writer) error {
	execSet, err := e.base.Clone()
	if err != nil {
		return fmt.Errorf("teggo: unable to clone templates: %w", err)
	}
	execSet.Funcs(e.funcMap(execSet))
	return execSet.ExecuteTemplate(w, name, data)
}

// TemplateNames retorna la lista de templates lógicos ordenados.
func (e *Engine) TemplateNames() []string {
	execSet, err := e.base.Clone()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(execSet.Templates()))
	for _, t := range execSet.Templates() {
		names = append(names, t.Name())
	}
	sort.Strings(names)
	return names
}

// FuncMap retorna el mapa de funciones helper para templates, incluyendo partial.
func (e *Engine) FuncMap() template.FuncMap {
	return e.funcMap(e.base)
}

// funcMap produce el mapa de funciones enlazado al set indicado.
// Incluye partial seguro (slots), helpers puros, etc.
func (e *Engine) funcMap(set *template.Template) template.FuncMap {
	return template.FuncMap{
		"dict":  Dict,
		"merge": Merge,
		"cat":   Cat,
		"partial": func(name string, props map[string]interface{}) template.HTML {
			return e.safePartial(set, name, props)
		},
	}
}

// safePartial ejecuta un componente en aislamiento (slots).
func (e *Engine) safePartial(_ *template.Template, name string, props map[string]interface{}) template.HTML {
	var buf bytes.Buffer

	// Clonar para evitar colisiones de definiciones dinámicas
	sub, err := e.base.Clone()
	if err != nil {
		return e.report(fmt.Errorf("clone error: %w", err))
	}

	// Inyecta slots: props con clave mayúscula.
	for k, v := range props {
		if len(k) > 0 && unicode.IsUpper(rune(k[0])) {
			if _, perr := sub.New(k).Parse(fmt.Sprint(v)); perr != nil {
				return e.report(fmt.Errorf("parse slot %q: %w", k, perr))
			}
		}
	}

	if err := sub.ExecuteTemplate(&buf, name, props); err != nil {
		return e.report(err)
	}
	return template.HTML(buf.String())
}

// report imprime el error (debug) y retorna un HTML vacío (producción).
func (e *Engine) report(err error) template.HTML {
	log.Printf("Teggo ▶ %v", err)
	// if e.debug {
	// 	return template.HTML("<!-- Teggo ERROR: " + template.HTMLEscapeString(err.Error()) + " -->")
	// }
	return "" // En producción devuelve vacío
}
