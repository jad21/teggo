package teggo

import (
	"strings"
	"testing"
)

// Utilidad para limpiar espacios para comparar resultados
func clean(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

// 1. Componente con slot anónimo
func TestParseTagsToGoTpl_SlotAnon(t *testing.T) {
	src := `
{{tag Button}}
<button class="{{.Class}}">
  {{slot}}
</button>
{{end}}
`
	expected := `
{{define "Button"}}
<button class="{{.Class}}">
  {{template "Slot" .}}
</button>
{{end}}
`
	got := ParseTagsToGoTpl(src, "Button", "Button")
	if clean(got) != clean(expected) {
		t.Errorf("Slot anónimo:\n--- Got ---\n%s\n--- Want ---\n%s\n", got, expected)
	}
}

// 2. Componente con slot nombrado
func TestParseTagsToGoTpl_SlotNamed(t *testing.T) {
	src := `
{{tag Card}}
<div>
  {{slot name="Header"}}
  <section>{{slot}}</section>
  <footer>{{slot name="Footer"}}</footer>
</div>
{{end}}
`
	expected := `
{{define "Card"}}
<div>
  {{template "Header" .}}
  <section>{{template "Slot" .}}</section>
  <footer>{{template "Footer" .}}</footer>
</div>
{{end}}
`
	got := ParseTagsToGoTpl(src, "Card", "Card")
	if clean(got) != clean(expected) {
		t.Errorf("Slot nombrado:\n--- Got ---\n%s\n--- Want ---\n%s\n", got, expected)
	}
}

// 3. Página usando slots y componentes
func TestParseTagsToGoTpl_PageWithComponents(t *testing.T) {
	src := `
<Card>
  <h1>Hola</h1>
  <Button>Click aquí</Button>
  <div slot="Footer">Fin</div>
</Card>
`
	// (expected eliminado porque no se usa directamente)
	got := ParseTagsToGoTpl(src, "Home", "pages.Home")
	if !strings.Contains(clean(got), clean(`{{partial "Card"`)) ||
		!strings.Contains(clean(got), clean(`{{define "__slot__1"}}`)) {
		t.Errorf("Página con componentes y slots parece mal parseada:\n--- Got ---\n%s\n", got)
	}
}

// 4. Componente con comentario de props incluyendo "end"
func TestParseTagsToGoTpl_CommentWithEnd(t *testing.T) {
	src := `
{{tag UserCard}}
{{/* props: Este componente maneja el fin de sesión (end), etc. */}}
<div class="usercard">
  <h3>{{.Name}}</h3>
  {{slot}}
</div>
{{end}}
`
	expected := `
{{define "UserCard"}}
{{/* props: Este componente maneja el fin de sesión (end), etc. */}}
<div class="usercard">
  <h3>{{.Name}}</h3>
  {{template "Slot" .}}
</div>
{{end}}
`
	got := ParseTagsToGoTpl(src, "UserCard", "UserCard")
	if clean(got) != clean(expected) {
		t.Errorf("Comentario con 'end' dentro:\n--- Got ---\n%s\n--- Want ---\n%s\n", got, expected)
	}
}

// 5. Bloques anidados de componentes
func TestParseTagsToGoTpl_NestedComponents(t *testing.T) {
	src := `
{{tag Parent}}
<div>
  <Child>
    <Button>OK</Button>
  </Child>
</div>
{{end}}
`
	got := ParseTagsToGoTpl(src, "Parent", "Parent")
	if !strings.Contains(got, `{{partial "Child"`) {
		t.Errorf("No encontró parcial de Child dentro de Parent:\n--- Got ---\n%s\n", got)
	}
}

// 6. Self-closing tags
func TestParseTagsToGoTpl_SelfClosing(t *testing.T) {
	src := `
{{tag Icon}}
<span class="icon"/>
{{end}}
`
	expected := `
{{define "Icon"}}
<span class="icon"/>
{{end}}
`
	got := ParseTagsToGoTpl(src, "Icon", "Icon")
	if clean(got) != clean(expected) {
		t.Errorf("Self-closing tag:\n--- Got ---\n%s\n--- Want ---\n%s\n", got, expected)
	}
}
