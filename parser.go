package teggo

import (
	"fmt"
	"regexp"
	"strings"
)

type parserState struct {
	fileName string // basename (Home) para slots
	mainName string // nombre lógico completo (pages.Home)
	slotID   int    // contador incremental
	defs     strings.Builder
}

// API pública
func ParseTagsToGoTpl(src, fileName, mainName string) string {
	ps := &parserState{fileName: fileName, mainName: mainName}
	body := ps.convert(src)
	var out strings.Builder
	out.WriteString(ps.defs.String())
	out.WriteString(`{{define "` + mainName + `"}}`)
	out.WriteString(body)
	out.WriteString(`{{end}}`)
	return out.String()
}

// Regex precompiladas
var (
	tagComponentOpen = regexp.MustCompile(`<([A-Z][A-Za-z0-9]*)\s*([^>]*)>`)
	tagComponentSelf = regexp.MustCompile(`<([A-Z][A-Za-z0-9]*)\s*([^>]*)\/>`)
	propRE           = regexp.MustCompile(`([A-Za-z0-9_]+)=["']([^"']+)["']|\{\.\.\.([A-Za-z0-9_\.]+)\}`)
	slotRE           = regexp.MustCompile(`([A-Za-z][A-Za-z0-9]*)\s+slot=["']([A-Za-z0-9_]+)["']`)
)

// Conversión recursiva
func (ps *parserState) convert(input string) string {
	output := input

	// --- self-closing components: <Button ... /> ---
	output = tagComponentSelf.ReplaceAllStringFunc(output, func(m string) string {
		sub := tagComponentSelf.FindStringSubmatch(m)
		tag := sub[1]
		props := ps.parseProps(sub[2])
		return fmt.Sprintf(`{{partial "%s" (%s)}}`, tag, props)
	})

	// --- apertura / cierre de componentes (slots) ---
	for {
		m := tagComponentOpen.FindStringSubmatchIndex(output)
		if m == nil {
			break
		}
		tag := output[m[2]:m[3]]
		propsStr := output[m[4]:m[5]]
		closeTag := fmt.Sprintf("</%s>", tag)
		innerStart := m[1]
		closeIdx := strings.Index(output[innerStart:], closeTag)
		if closeIdx == -1 {
			break
		}
		innerEnd := innerStart + closeIdx
		innerContent := output[innerStart:innerEnd]

		slotName := ""
		if slotRE.MatchString(output[m[0]:m[1]]) {
			slotMatch := slotRE.FindStringSubmatch(output[m[0]:m[1]])
			slotName = slotMatch[2]
		}
		props := ps.parseProps(propsStr)
		if slotName != "" {
			props = ps.mergeSlotWithName(props, slotName, ps.convert(innerContent))
		} else {
			props = ps.mergeSlot(props, ps.convert(innerContent))
		}

		newBlock := fmt.Sprintf(`{{partial "%s" (%s)}}`, tag, props)
		output = output[:m[0]] + newBlock + output[innerEnd+len(closeTag):]
	}

	return output
}

// Helpers

func (ps *parserState) nextSlot(inner string) string {
	ps.slotID++
	name := fmt.Sprintf("__slot_%s_%d", ps.fileName, ps.slotID)
	ps.defs.WriteString(fmt.Sprintf(`{{define "%s"}}%s{{end}}`, name, inner))
	return name
}

// parseProps genera merge/dict robusto para props normales y spread
func (ps *parserState) parseProps(propStr string) string {
	var dicts []string
	var spreads []string
	for _, match := range propRE.FindAllStringSubmatch(propStr, -1) {
		if match[3] != "" {
			spreads = append(spreads, fmt.Sprintf(".%s", match[3]))
		} else {
			dicts = append(dicts, fmt.Sprintf(`"%s" "%s"`, match[1], match[2]))
		}
	}
	dictCode := "dict " + strings.Join(dicts, " ")
	if len(dicts) == 0 && len(spreads) == 1 {
		// solo un spread
		return spreads[0]
	}
	if len(spreads) == 0 {
		return dictCode
	}
	code := spreads[0]
	for i := 1; i < len(spreads); i++ {
		code = fmt.Sprintf("merge %s %s", code, spreads[i])
	}
	code = fmt.Sprintf("merge %s (%s)", code, dictCode)
	return code
}

func (ps *parserState) mergeSlot(dictCode, inner string) string {
	name := ps.nextSlot(inner)
	if strings.Contains(dictCode, "dict") {
		return strings.Replace(dictCode, "dict",
			fmt.Sprintf(`dict "Slot" (partial "%s" .)`, name), 1)
	}
	return fmt.Sprintf(`merge %s (dict "Slot" (partial "%s" .))`, dictCode, name)
}

func (ps *parserState) mergeSlotWithName(dictCode, slotName, inner string) string {
	name := ps.nextSlot(inner)
	if strings.Contains(dictCode, "dict") {
		return strings.Replace(dictCode, "dict",
			fmt.Sprintf(`dict "%s" (partial "%s" .)`, slotName, name),
			1)
	}
	return fmt.Sprintf(`merge %s (dict "%s" (partial "%s" .))`, dictCode, slotName, name)
}
