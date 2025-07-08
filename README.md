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


## ¿Cómo usar?

1. Instala Teggo en tu proyecto (próximamente: `go get github.com/jad21/teggo`)

```go
package main

import (
    "github.com/jad21/teggo"
    "html/template"
    "os"
)

func main() {
    // Lee y convierte el template tipo tag
    files := teggo.Discover("./examples", "*.html") // helper opcional
    // 2. Inicializa el engine Teggo
    engine, err := teggo.NewEngine(files, true)
    if err != nil {
      panic(fmt.Sprintf("Error inicializando Teggo: %v", err))
    }
    err = engine.Render("pages.Home", data, os.Stdout)
    if err != nil {
      panic(fmt.Sprintf("Error renderizando: %v", err))
    }
}
```

---

## Características principales

* Sintaxis tipo tag para componentes (`<Card Title="...">...</Card>`)
* Soporte para slots y slots nombrados
* Props normales
* Helpers: `partial`, `dict`, `merge`, `cat`
* Modular, fácil de extender

---

## Ejemplo avanzado: Props slots

```html
<UserCard ShowActions="true">
  <Button>Editar</Button>
</UserCard>
```

---

## Roadmap
* [ ] Lógica spread (`<UserCard {...User} />`)
* [ ] Lógica condicional y repetición tipo tag (`<If>`, `<For>`)
* [ ] Validación de props y valores por defecto
* [ ] Hot reload en desarrollo
* [ ] Ejemplos y documentación avanzada

---

## Lema

> **Escribe Go. Piensa en componentes. Renderiza con Teggo.**

---

**Hecho con convicción, claridad y respeto por la simplicidad.**

