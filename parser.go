package teggo

import (
	"fmt"
	"regexp"
	"strings"
)

// Regex precompiladas
var (
	tagComponentOpen = regexp.MustCompile(`<([A-Z][A-Za-z0-9]*)\s*([^>]*)>`)
	tagComponentSelf = regexp.MustCompile(`<([A-Z][A-Za-z0-9]*)\s*([^>]*)\/>`)
	propRE           = regexp.MustCompile(`([A-Za-z0-9_]+)=["']([^"']+)["']|\{\.\.\.([A-Za-z0-9_\.]+)\}`)
	slotRE           = regexp.MustCompile(`([A-Za-z][A-Za-z0-9]*)\s+slot=["']([A-Za-z0-9_]+)["']`)
	tagBlockRE       = regexp.MustCompile(`(?s)\{\{tag\s+([A-Za-z][A-Za-z0-9_]*)\}\}(.*?)\n\{\{\s*end\s*\}\}`)
	reNamedSlot      = regexp.MustCompile(`\{\{\s*slot\s+name=['"]([A-Za-z0-9_]+)['"]\s*\}\}`)
	reAnonSlot       = regexp.MustCompile(`\{\{\s*slot\s*\}\}`)
)

type parserState struct {
	fileName string
	mainName string
	slotID   int
	slots    []string // defines de slots acumulados
}

func ParseTagsToGoTpl(src, fileName, mainName string) string {
	var out strings.Builder

	// 1) Procesar bloques {{tag ...}}
	matches := tagBlockRE.FindAllStringSubmatchIndex(src, -1)
	if len(matches) > 0 {
		for _, m := range matches {
			name := src[m[2]:m[3]]
			body := src[m[4]:m[5]]

			ps := &parserState{fileName: fileName, mainName: mainName}
			content := ps.convert(body)
			content = strings.Trim(content, "\n")

			// Primero los defines de slots
			for _, slot := range ps.slots {
				out.WriteString(slot + "\n")
			}
			// Luego el define principal, sin líneas extras
			out.WriteString(fmt.Sprintf("{{define \"%s\"}}\n%s\n{{end}}\n", name, content))
		}
		return out.String()
	}

	// 2) Sin bloques tag: archivo completo
	ps := &parserState{fileName: fileName, mainName: mainName}
	content := ps.convert(src)
	content = strings.Trim(content, "\n")
	for _, slot := range ps.slots {
		out.WriteString(slot + "\n")
	}
	out.WriteString(fmt.Sprintf("{{define \"%s\"}}\n%s\n{{end}}\n", mainName, content))
	return out.String()
}

func (ps *parserState) convert(input string) string {
	output := input

	// Self-closing <Tag .../>
	output = tagComponentSelf.ReplaceAllStringFunc(output, func(m string) string {
		sub := tagComponentSelf.FindStringSubmatch(m)
		tag := sub[1]
		props := ps.parseProps(sub[2])
		return fmt.Sprintf(`{{partial "%s" (%s)}}`, tag, props)
	})

	// Componentes con matching anidado
	for {
		m := tagComponentOpen.FindStringSubmatchIndex(output)
		if m == nil {
			break
		}
		tag := output[m[2]:m[3]]
		if strings.EqualFold(tag, ps.fileName) {
			break
		}
		propsStr := output[m[4]:m[5]]
		openTag := fmt.Sprintf("<%s", tag)
		closeTag := fmt.Sprintf("</%s>", tag)
		start, innerStart := m[0], m[1]
		depth, i := 1, innerStart

		for i < len(output) {
			nextOpen := strings.Index(output[i:], openTag)
			nextClose := strings.Index(output[i:], closeTag)
			if nextClose == -1 {
				break
			}
			if nextOpen != -1 && nextOpen < nextClose {
				depth++
				i += nextOpen + len(openTag)
			} else {
				depth--
				if depth == 0 {
					innerEnd := i + nextClose
					inner := output[innerStart:innerEnd]
					slotName := ""
					if slotRE.MatchString(output[m[0]:m[1]]) {
						slotName = slotRE.FindStringSubmatch(output[m[0]:m[1]])[2]
					}
					props := ps.parseProps(propsStr)
					var replaced string
					if slotName != "" {
						replaced = ps.mergeSlotWithName(props, slotName, ps.convert(inner))
					} else {
						replaced = ps.mergeSlot(props, ps.convert(inner))
					}
					// Componente → sólo el partial, descartamos el cierre HTML
					newBlock := fmt.Sprintf(`{{partial "%s" (%s)}}`, tag, replaced)
					output = output[:start] + newBlock + output[innerEnd+len(closeTag):]
					break
				}
				i += nextClose + len(closeTag)
			}
		}
	}

	// Slots nombrados
	output = reNamedSlot.ReplaceAllStringFunc(output, func(m string) string {
		name := reNamedSlot.FindStringSubmatch(m)[1]
		return fmt.Sprintf(`{{template "%s" .}}`, name)
	})
	// Slot anónimo
	output = reAnonSlot.ReplaceAllString(output, `{{template "Slot" .}}`)

	return output
}

func (ps *parserState) nextSlot(inner string) string {
	ps.slotID++
	name := fmt.Sprintf("__slot_%s_%d", ps.fileName, ps.slotID)
	trimmed := strings.Trim(inner, "\n")
	def := fmt.Sprintf("{{define \"%s\"}}\n%s\n{{end}}", name, trimmed)
	ps.slots = append(ps.slots, def)
	return name
}

func (ps *parserState) parseProps(propStr string) string {
	var dicts, spreads []string
	for _, m := range propRE.FindAllStringSubmatch(propStr, -1) {
		if m[3] != "" {
			spreads = append(spreads, "."+m[3])
		} else {
			dicts = append(dicts, fmt.Sprintf(`"%s" "%s"`, m[1], m[2]))
		}
	}
	dictCode := "dict " + strings.Join(dicts, " ")
	if len(dicts) == 0 && len(spreads) == 1 {
		return spreads[0]
	}
	if len(spreads) == 0 {
		return dictCode
	}
	code := spreads[0]
	for _, s := range spreads[1:] {
		code = fmt.Sprintf("merge %s %s", code, s)
	}
	return fmt.Sprintf("merge %s (%s)", code, dictCode)
}

func (ps *parserState) mergeSlot(dictCode, inner string) string {
	name := ps.nextSlot(inner)
	if strings.Contains(dictCode, "dict") {
		return strings.Replace(dictCode, "dict", fmt.Sprintf(`dict "Slot" (partial "%s" .)`, name), 1)
	}
	return fmt.Sprintf(`merge %s (dict "Slot" (partial "%s" .))`, dictCode, name)
}

func (ps *parserState) mergeSlotWithName(dictCode, slotName, inner string) string {
	name := ps.nextSlot(inner)
	if strings.Contains(dictCode, "dict") {
		return strings.Replace(dictCode, "dict", fmt.Sprintf(`dict "%s" (partial "%s" .)`, slotName, name), 1)
	}
	return fmt.Sprintf(`merge %s (dict "%s" (partial "%s" .))`, dictCode, slotName, name)
}
