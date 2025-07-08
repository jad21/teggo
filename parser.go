// parser.go
// Paquete teggo — Parser JSX-like a Go html/template compatible.
// -----------------------------------------------------------------------------
// Convierte sintaxis JSX-like con componentes en plantillas Go estándar.

package teggo

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// -----------------------------------------------------------------------------
// ComponentRegistry — Global compartido con el Engine.
// -----------------------------------------------------------------------------
var componentRegistry = map[string]struct{}{}

func SetComponentRegistry(reg map[string]struct{}) {
	componentRegistry = reg
}

// -----------------------------------------------------------------------------
// Patrones comunes
// -----------------------------------------------------------------------------
var (
	tagPattern       = regexp.MustCompile(`{{\s*tag\s+(\w+)\s*}}`)
	slotNamedPattern = regexp.MustCompile(`{{\s*slot\s+name\s*=\s*"(.*?)"\s*}}`)
	slotAnonPattern  = regexp.MustCompile(`{{\s*slot\s*}}`)
	mustacheBlock    = regexp.MustCompile(`{{.*?}}`)
)

// -----------------------------------------------------------------------------
// Entrada principal
// -----------------------------------------------------------------------------
func ParseTagsToGoTpl(source, base, logicalName string) string {
	if hasTagDirective(source) {
		return parseComponent(source)
	}
	return parsePage(source, logicalName)
}

// Detecta si es un componente con {{tag Name}}
func hasTagDirective(source string) bool {
	return tagPattern.MatchString(source)
}

// -----------------------------------------------------------------------------
// Conversión de definición de componente
// -----------------------------------------------------------------------------
func parseComponent(source string) string {
	out := tagPattern.ReplaceAllStringFunc(source, func(m string) string {
		match := tagPattern.FindStringSubmatch(m)
		if len(match) > 1 {
			return fmt.Sprintf(`{{define "%s"}}`, match[1])
		}
		return m
	})
	out = slotNamedPattern.ReplaceAllString(out, `{{template "$1" .}}`)
	out = slotAnonPattern.ReplaceAllString(out, `{{template "slot" .}}`)
	return out
}

// -----------------------------------------------------------------------------
// Conversión de página (uso de componentes en JSX-like)
// -----------------------------------------------------------------------------
func parsePage(source, logicalName string) string {
	// 1️⃣ Extraer y proteger bloques GoTpl
	cleanSrc, blocks := extractTemplateBlocks(source)

	// 2️⃣ Marcar componentes registrados con teggo-component
	markedSrc := markComponentTags(cleanSrc)

	// 3️⃣ Parsear como HTML
	node, err := html.Parse(strings.NewReader(markedSrc))
	if err != nil {
		return wrapAsDefine(logicalName, source)
	}

	// 4️⃣ Buscar contenido real (body)
	body := findBody(node)
	if body == nil {
		return wrapAsDefine(logicalName, source)
	}

	// 5️⃣ Procesar nodos
	slotCounter := 0
	var slotDefs []string
	var buf bytes.Buffer

	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := walkNode(&buf, c, logicalName, &slotCounter, &slotDefs, blocks); err != nil {
			return wrapAsDefine(logicalName, source)
		}
	}

	// 6️⃣ Generar define principal
	var final bytes.Buffer
	final.WriteString(fmt.Sprintf(`{{define "%s"}}`, logicalName))
	final.WriteString("\n")
	final.WriteString(buf.String())
	final.WriteString("\n{{end}}\n")

	// 7️⃣ Adjuntar defines de slots hijos
	for _, def := range slotDefs {
		final.WriteString(def)
		final.WriteString("\n")
	}

	fmt.Println(final.String())

	return final.String()
}

// -----------------------------------------------------------------------------
// Reemplazo previo: Marcar componentes Teggo válidos
// -----------------------------------------------------------------------------
func markComponentTags(input string) string {
	for comp := range componentRegistry {
		// Reemplazar apertura
		openTag := regexp.MustCompile(fmt.Sprintf(`<%s(\b|>)`, regexp.QuoteMeta(comp)))
		input = openTag.ReplaceAllString(input, fmt.Sprintf(`<teggo-component teggo:name="%s"$1`, comp))

		// Reemplazar cierre
		closeTag := regexp.MustCompile(fmt.Sprintf(`</%s>`, regexp.QuoteMeta(comp)))
		input = closeTag.ReplaceAllString(input, `</teggo-component>`)
	}
	return input
}

// -----------------------------------------------------------------------------
// Helpers para parseo y reemplazo
// -----------------------------------------------------------------------------
func wrapAsDefine(name, body string) string {
	return fmt.Sprintf(`{{define "%s"}}
%s
{{end}}`, name, body)
}

func extractTemplateBlocks(input string) (string, []string) {
	blocks := []string{}
	output := mustacheBlock.ReplaceAllStringFunc(input, func(m string) string {
		idx := len(blocks)
		blocks = append(blocks, m)
		return fmt.Sprintf("__TPL_%d__", idx)
	})
	return output, blocks
}

func restoreTemplateBlocks(input string, blocks []string) string {
	out := input
	for i, block := range blocks {
		out = strings.ReplaceAll(out, fmt.Sprintf("__TPL_%d__", i), block)
	}
	return out
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if res := findBody(c); res != nil {
			return res
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Caminar el árbol HTML parseado
// -----------------------------------------------------------------------------
func walkNode(buf *bytes.Buffer, n *html.Node, logicalPath string, counter *int, slotDefs *[]string, blocks []string) error {
	switch n.Type {
	case html.TextNode:
		buf.WriteString(restoreTemplateBlocks(n.Data, blocks))

	case html.ElementNode:
		if strings.HasPrefix(n.Data, "__TPL_") {
			buf.WriteString(n.Data)
			return nil
		}

		// Teggo-component marcado
		if n.Data == "teggo-component" {
			var compName string
			props := []html.Attribute{}
			for _, a := range n.Attr {
				if a.Key == "teggo:name" {
					compName = a.Val
				} else {
					props = append(props, a)
				}
			}

			if compName != "" && isRegisteredComponent(compName) {
				return renderComponent(buf, &html.Node{
					Type:       html.ElementNode,
					Data:       compName,
					Attr:       props,
					FirstChild: n.FirstChild,
				}, compName, logicalPath, counter, slotDefs, blocks)
			}
		}

		// Tag HTML normal
		buf.WriteString("<" + n.Data)
		for _, attr := range n.Attr {
			buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
		}
		buf.WriteString(">")

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walkNode(buf, c, logicalPath, counter, slotDefs, blocks); err != nil {
				return err
			}
		}
		buf.WriteString("</" + n.Data + ">")
	}
	return nil
}

// Verifica si es un componente Teggo registrado
func isRegisteredComponent(name string) bool {
	_, ok := componentRegistry[name]
	return ok
}

// Renderiza la llamada al template GoTpl
func renderComponent(buf *bytes.Buffer, n *html.Node, componentName, logicalPath string, counter *int, slotDefs *[]string, blocks []string) error {
	// Props
	props := map[string]string{}
	for _, attr := range n.Attr {
		props[attr.Key] = attr.Val
	}

	// Slots
	childSlots := map[string]string{}
	var anonSlotContent bytes.Buffer

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "slot" {
			// Slot nombrado o anónimo
			nameAttr := "slot"
			for _, a := range c.Attr {
				if a.Key == "name" {
					nameAttr = a.Val
					break
				}
			}
			var slotBuf bytes.Buffer
			for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
				walkNode(&slotBuf, gc, logicalPath, counter, slotDefs, blocks)
			}
			slotName := slotDefineName(logicalPath, componentName, nameAttr, *counter)
			*slotDefs = append(*slotDefs, fmt.Sprintf(`{{define "%s"}}%s{{end}}`, slotName, slotBuf.String()))
			childSlots[nameAttr] = slotName
			*counter++
		} else {
			// Slot anónimo
			walkNode(&anonSlotContent, c, logicalPath, counter, slotDefs, blocks)
		}
	}

	if strings.TrimSpace(anonSlotContent.String()) != "" {
		slotName := slotDefineName(logicalPath, componentName, "slot", *counter)
		*slotDefs = append(*slotDefs, fmt.Sprintf(`{{define "%s"}}%s{{end}}`, slotName, anonSlotContent.String()))
		childSlots["slot"] = slotName
		*counter++
	}

	// Generar llamada GoTpl
	buf.WriteString(`{{template "` + componentName + `" dict `)
	for k, v := range props {
		fmt.Fprintf(buf, `"%s" "%s" `, k, v)
	}
	for k, v := range childSlots {
		fmt.Fprintf(buf, `"%s" (template "%s" .) `, k, v)
	}
	buf.WriteString(`}}`)
	return nil
}

func slotDefineName(logicalPath, component, slotName string, counter int) string {
	return fmt.Sprintf("%s__%s__%s__%d", logicalPath, component, slotName, counter)
}
