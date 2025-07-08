// discover.go
// Paquete teggo — Descubrimiento de archivos de plantilla por patrón.
// -----------------------------------------------------------------------------
// Devuelve todos los archivos coincidentes con los patrones glob dados.

package teggo

import (
	"io/fs"
	"path/filepath"
	"sort"
)

// Discover recorre dirRoot y devuelve archivos cuyo nombre coincida
// con los sufijos dados (*.html, *.gotpl, etc.). Si no hay sufijos, trae todo.
func Discover(dirRoot string, suffixes ...string) []string {
	if len(suffixes) == 0 {
		suffixes = []string{"*"}
	}
	seen := make(map[string]struct{})
	var out []string
	_ = filepath.WalkDir(dirRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
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
