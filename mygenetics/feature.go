package mygenetics

import (
	"fmt"
	"strings"
)

type Gene struct {
	// Name представляет название гена.
	Name string

	// Interpretation представляет интерпретацию гена.
	Interpretation []string
}

// Reccommendation представляет рекомендацию.
type Reccommendation struct {
	// Nutrition представляет рекомендации по питанию.
	Nutrition []string

	// Additional представляет дополнительные рекомендации.
	Additional []string

	// Checklist представляет чеклист.
	Checklist []string
}

// Feature представляет признак на котором строится заключение.
type Feature struct {
	// Name представляет название признака.
	Name string

	// Conclusions представляет заключения по признаку.
	Conclusions []string

	// Genes представляет список генов в признаке.
	Genes []Gene

	// Recommendations представляет рекомендации по признаку.
	Recommendation Reccommendation
}

func (f Feature) String() string {
	sb := new(strings.Builder)

	// Заголовок
	sb.WriteString(fmt.Sprintf("<b>🧬 Признак: %s</b>\n\n", f.Name))

	// Гены
	if len(f.Genes) > 0 {
		sb.WriteString("<b>🔬 Входящие гены:</b>\n")
		for _, gene := range f.Genes {
			sb.WriteString(fmt.Sprintf("• <b>%s</b>\n", gene.Name))
			if len(gene.Interpretation) > 0 {
				sb.WriteString(
					fmt.Sprintf(
						"  <i>%s</i>\n",
						strings.Join(gene.Interpretation, " "),
					),
				)
			}
		}
		sb.WriteString("\n")
	}

	// Рекомендации

	if len(f.Recommendation.Nutrition) > 0 ||
		len(f.Recommendation.Additional) > 0 ||
		len(f.Recommendation.Checklist) > 0 {
		sb.WriteString("<b>💡 Рекомендации:</b>\n")
	}

	if len(f.Recommendation.Nutrition) > 0 {
		sb.WriteString("<b>🍎 Питание:</b>\n")
		for _, n := range f.Recommendation.Nutrition {
			sb.WriteString(fmt.Sprintf("• %s\n", n))
		}
		sb.WriteString("\n")
	}

	if len(f.Recommendation.Additional) > 0 {
		sb.WriteString("<b>📌 Дополнительные рекомендации:</b>\n")
		for _, a := range f.Recommendation.Additional {
			sb.WriteString(fmt.Sprintf("• %s\n", a))
		}
		sb.WriteString("\n")
	}

	if len(f.Recommendation.Checklist) > 0 {
		sb.WriteString("<b>✅ Чеклист:</b>\n")
		for _, c := range f.Recommendation.Checklist {
			sb.WriteString(fmt.Sprintf("• %s\n", c))
		}
		sb.WriteString("\n")
	}

	// Заключения
	if len(f.Conclusions) > 0 {
		sb.WriteString("<b>📋 Заключения:</b>\n")
		for _, c := range f.Conclusions {
			sb.WriteString(fmt.Sprintf("• %s\n", c))
		}
	}

	return sb.String()
}
