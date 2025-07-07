// -----------------------------------------------------------------------------
// Teggo — JSX‑like Components for Go
// engine.go – núcleo de compilación y renderizado de plantillas
// -----------------------------------------------------------------------------
//   - Compila todos los archivos indicados en `paths`.
//   - Registra nombres lógicos «subdir.Base» usando el directorio común
//     como raíz (examples/pages/Home.html → pages.Home).
//   - Clona el set para cada ejecución (thread‑safe).
//   - Expone helpers dict/merge/cat y partial seguro.
//
// -----------------------------------------------------------------------------
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

// Engine mantiene el set de templates compilado y la bandera debug.
type Engine struct {
	base  *template.Template // nunca se ejecuta ni lista nombres
	debug bool
}

// -----------------------------------------------------------------------------
//  Construcción
// -----------------------------------------------------------------------------

// NewEngine compila todos los archivos indicados en paths.
func NewEngine(paths []string, debug bool) (*Engine, error) {
	e := &Engine{debug: debug}
	rootDir := commonDir(paths)

	var sb strings.Builder
	for _, absPath := range paths {
		srcBytes, err := os.ReadFile(absPath)
		if err != nil {
			return nil, err
		}

		// Obtiene nombre lógico para cada template (ej: "pages.Home")
		rel, _ := filepath.Rel(rootDir, absPath)
		logical := strings.TrimSuffix(rel, filepath.Ext(rel))
		mainName := strings.ReplaceAll(logical, string(os.PathSeparator), ".")
		base := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))

		converted := ParseTagsToGoTpl(string(srcBytes), base, mainName)
		sb.WriteString(converted + "\n")
	}

	// Crea set principal (root) y lo llena con todos los templates
	root := template.New("root")
	root.Funcs(e.funcMap(root))
	baseSet, err := root.Parse(sb.String())
	if err != nil {
		return nil, err
	}
	e.base = baseSet // <-- JAMÁS ejecutar ni clonar después de ejecutar

	return e, nil
}

// -----------------------------------------------------------------------------
//  Ejecución
// -----------------------------------------------------------------------------

// Render clona el set y ejecuta el template indicado.
func (e *Engine) Render(name string, data any, w io.Writer) error {
	execSet, err := e.base.Clone()
	if err != nil {
		return fmt.Errorf("teggo: unable to clone templates: %w", err)
	}
	execSet.Funcs(e.funcMap(execSet))
	return execSet.ExecuteTemplate(w, name, data)
}

// TemplateNames devuelve la lista ordenada de nombres lógicos registrados.
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

// -----------------------------------------------------------------------------
//  FuncMap & helpers internos
// -----------------------------------------------------------------------------

func (e *Engine) FuncMap() template.FuncMap {
	return e.funcMap(e.base)
}

// funcMap produce el mapa de funciones enlazado al set indicado.
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

// safePartial ejecuta un componente aislado.
func (e *Engine) safePartial(_ *template.Template, name string, props map[string]interface{}) template.HTML {
	var buf bytes.Buffer

	// clonar para evitar colisiones de definiciones dinámicas
	sub, err := e.base.Clone()
	if err != nil {
		return e.report(fmt.Errorf("clone error: %w", err))
	}

	// inyectar slots (props con clave iniciando en mayúscula)
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

func (e *Engine) report(err error) template.HTML {
	log.Printf("Teggo ▶ %v", err)
	// if e.debug {
	// 	return template.HTML("<!-- Teggo ERROR: " + template.HTMLEscapeString(err.Error()) + " -->")
	// }
	return "" // en producción devolvemos vacío
}
