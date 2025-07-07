// Teggo — JSX‑like Components for Go
// discover.go – utilitario para localizar archivos de plantilla de manera simple.
// -----------------------------------------------------------------------------
// Ejemplo de uso:
//   files := teggo.Discover("examples", "*.html", "*.gotpl")
// Retorna la lista ordenada (alfabéticamente) y sin duplicados de todos los
// archivos que coincidan con los patrones dados (glob) dentro del directorio
// raíz indicado.  Los patrones se aplican sobre el nombre **base** del archivo.
// -----------------------------------------------------------------------------

package teggo

import (
	"io/fs"
	"path/filepath"
	"sort"
)

// Discover recorre dirRoot recursivamente y devuelve los archivos cuyo nombre
// coincida con alguno de los sufijos (glob).  Si no se pasa sufijo se devuelve
// todo.
func Discover(dirRoot string, suffixes ...string) []string {
	// normaliza sufijos; si viene vacío ponemos "*" (cualquier cosa)
	if len(suffixes) == 0 {
		suffixes = []string{"*"}
	}

	seen := make(map[string]struct{})
	var out []string

	_ = filepath.WalkDir(dirRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil // ignorar errores de permisos y directorios
		}
		base := filepath.Base(path)
		for _, pat := range suffixes {
			if ok, _ := filepath.Match(pat, base); ok {
				if _, dup := seen[path]; !dup {
					seen[path] = struct{}{}
					out = append(out, path)
				}
				break
			}
		}
		return nil
	})

	sort.Strings(out)
	return out
}
