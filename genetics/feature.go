package genetics

import (
	"fmt"
	"strings"
)

// Feature represents a trait on which conclusions are based
type Feature struct {
	Name        string   // Name of the feature
	Genes       []Gene   // Genes associated with this feature
	Conclusions []string // Conclusions about the feature
	Nutrition   []string // Dietary recommendations
	Additional  []string // Additional lifestyle recommendations
	Checklist   []string // Action items to be completed
}

// Gene represents a genetic marker with its name and interpretations
type Gene struct {
	Name            string   // Name of the gene
	Interpretations []string // Interpretations of gene variations
}

// ToHTML formats Feature in HTML format with emoji
func (f Feature) ToHTML() string {
	sb := new(strings.Builder)

	// Title
	sb.WriteString(fmt.Sprintf("<b>🧬 Признак: %s</b>\n\n", f.Name))

	// Genes
	if len(f.Genes) > 0 {
		sb.WriteString("<b>🔬 Входящие гены:</b>\n")
		for _, gene := range f.Genes {
			sb.WriteString(fmt.Sprintf("• <b>%s</b>\n", gene.Name))
			if len(gene.Interpretations) > 0 {
				sb.WriteString(
					fmt.Sprintf(
						"  <i>%s</i>\n",
						strings.Join(gene.Interpretations, " "),
					),
				)
			}
		}
		sb.WriteString("\n")
	}

	// Nutrition recommendations
	if len(f.Nutrition) > 0 {
		sb.WriteString("<b>🍎 Питание:</b>\n")
		for _, n := range f.Nutrition {
			sb.WriteString(fmt.Sprintf("• %s\n", n))
		}
		sb.WriteString("\n")
	}

	// Additional recommendations
	if len(f.Additional) > 0 {
		sb.WriteString("<b>📌 Дополнительные рекомендации:</b>\n")
		for _, a := range f.Additional {
			sb.WriteString(fmt.Sprintf("• %s\n", a))
		}
		sb.WriteString("\n")
	}

	// Checklist
	if len(f.Checklist) > 0 {
		sb.WriteString("<b>✅ Чеклист:</b>\n")
		for _, c := range f.Checklist {
			sb.WriteString(fmt.Sprintf("• %s\n", c))
		}
		sb.WriteString("\n")
	}

	// Conclusions
	if len(f.Conclusions) > 0 {
		sb.WriteString("<b>📋 Заключения:</b>\n")
		for _, c := range f.Conclusions {
			sb.WriteString(fmt.Sprintf("• %s\n", c))
		}
	}

	return sb.String()
}
