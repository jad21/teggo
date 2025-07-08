package teggo

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func clean(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

func TestParseTagsToGoTpl_CardWithFooterSlotAndButton_FullRender(t *testing.T) {
	// 1️⃣ Fuente de Card
	cardSrc := `
{{tag Card}}
<div class="card">
  <h2>{{.Title}}</h2>
  <section>
    {{slot}}
  </section>
  <footer>
    {{slot name="Footer"}}
  </footer>
</div>
{{end}}
`

	// 2️⃣ Fuente de Button
	buttonSrc := `
{{tag MyButton}}
<button class="{{.Class}}">
  {{slot}}
</button>
{{end}}
`

	// 3️⃣ Página que usa Card con slot Footer que contiene un Button
	pageSrc := `
<Card Title="Hola">
  contenido
  <slot name="Footer">
       <MyButton Class="success">Guardar</MyButton>
  </slot>
</Card>
`

	// ✅ Simula archivos descubiertos
	files := map[string]string{
		"components/Card.html":   cardSrc,
		"components/Button.html": buttonSrc,
		"pages/Home.html":        pageSrc,
	}

	// ✅ Usa el nuevo Engine que registra componentes y parsea
	eng, err := NewEngineFromSource(files, true)
	if err != nil {
		t.Fatalf("Engine failed to parse generated templates: %v", err)
	}

	// ✅ Intenta renderizar la página
	var out strings.Builder
	err = eng.Render("pages.Home", map[string]interface{}{
		"IsAdmin": true,
		"Users": []map[string]interface{}{
			{"Name": "Alice"},
		},
	}, &out)
	if err != nil {
		t.Fatalf("Engine failed to render: %v", err)
	}

	// ✅ Salida generada
	rendered := out.String()
	if strings.TrimSpace(rendered) == "" {
		t.Errorf("Rendered output is empty. Expected non-empty HTML.")
	}
	os.WriteFile("rendered.html", []byte(rendered), 0644)

	want := `<div class="card">
  <h2>Hola</h2>
  <section>
    contenido
  </section>
  <footer>
    <button class="success">Guardar</button>
  </footer>
</div>`

	fmt.Println(eng.componentRegistry)
	if clean(rendered) != clean(want) {
		t.Errorf("Button Component:\n--- Got ---\n%s\n--- Want ---\n%s\n", rendered, want)
	}

}
