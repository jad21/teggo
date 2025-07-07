# Teggo

**JSX-like Component Templates for Go**  
_Simple. Modular. Potente._

---

## ¿Qué es Teggo?

Teggo es una librería que lleva la experiencia de los componentes y sintaxis tipo tag (al estilo JSX/React) al desarrollo de aplicaciones web en Go, usando el motor estándar `html/template` como base.

- Escribe tus vistas como si fueran tags de HTML.
- Composición avanzada con slots y props, igual que en React.
- Reutiliza y organiza tus vistas como componentes.
- Disfruta el rendimiento, seguridad y simplicidad de Go puro.

---

## Filosofía

- **Simplicidad:** Cada vista es un componente, cada componente es un tag.
- **Composición:** Anida y reutiliza tus componentes. Usa slots y props para máxima flexibilidad.
- **Rendimiento:** Teggo transpila a `html/template`. Cacheo y SSR de alto rendimiento.
- **Sin dependencias externas:** No requiere frameworks adicionales. Compatible con cualquier stack Go.

---

## Ejemplo rápido

### **Componente: Button**
```html
<!-- components/Button.gotpl -->
{{!props: Class string, Slot any }}
<button class="{{.Class}}">
  {{template "Slot" .}}
</button>
````

### **Componente: Card**

```html
<!-- components/Card.gotpl -->
{{!props: Title string, Slot any, Footer any }}
<div class="card">
  <h2>{{.Title}}</h2>
  <section>{{template "Slot" .}}</section>
  <footer>{{template "Footer" .}}</footer>
</div>
```

### **Página: Home**

```html
<!-- pages/Home.gotpl -->
<Card Title="Bienvenido a Teggo!">
  <Button Class="btn-primary">Click aquí</Button>
  <Footer slot="Footer">© 2025 Teggo</Footer>
</Card>
```

---

## ¿Cómo usar?

1. Instala Teggo en tu proyecto (próximamente: `go get github.com/jad21/teggo`)
2. Escribe tus componentes y páginas en archivos `.gotpl`
3. Usa la función `ParseTagsToGoTpl` de Teggo para transformar tus vistas a Go templates estándar.
4. Renderiza usando las helpers de Teggo en tu aplicación.

```go
package main

import (
    "github.com/jad21/teggo"
    "html/template"
    "os"
)

func main() {
    // Lee y convierte el template tipo tag
    src := teggo.ReadFile("pages/Home.gotpl")
    tplStr := teggo.ParseTagsToGoTpl(src)
    tmpl, _ := template.New("main").Funcs(teggo.FuncMap()).Parse(tplStr)

    // Renderiza con datos
    tmpl.Execute(os.Stdout, map[string]interface{}{})
}
```

---

## Características principales

* Sintaxis tipo tag para componentes (`<Card Title="...">...</Card>`)
* Soporte para slots y slots nombrados
* Props normales y spread (`<UserCard {...User} />`)
* Helpers: `partial`, `dict`, `merge`, `cat`
* Modular, fácil de extender

---

## Ejemplo avanzado: Props Spread y slots

```html
<UserCard {...UserData} ShowActions="true">
  <Button>Editar</Button>
</UserCard>
```

---

## Roadmap

* [ ] Lógica condicional y repetición tipo tag (`<If>`, `<For>`)
* [ ] Validación de props y valores por defecto
* [ ] Hot reload en desarrollo
* [ ] Ejemplos y documentación avanzada

---

## Lema

> **Escribe Go. Piensa en componentes. Renderiza con Teggo.**

---

**Hecho con convicción, claridad y respeto por la simplicidad.**

