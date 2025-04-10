// TODO: REFACTOR ME!!!
package mygenetics

import (
	"fmt"

	"github.com/muzykantov/health-gpt/genetics"
)

type Reccommendation struct {
	Nutrition  []string // Dietary recommendations
	Additional []string // Additional lifestyle recommendations
	Checklist  []string // Action items to be completed
}

func (c *Client) parseWithConclusion(
	featureName string,
	featureValue map[string]any,
) (genetics.Feature, error) {
	feature := genetics.Feature{
		Name: featureName,
	}

	conclusions, ok := featureValue["conclusion"].(map[string]any)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			featureValue["conclusion"],
		)
	}

	fullConclusion, ok := conclusions["conclusion"].([]any)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			conclusions["conclusion"],
		)
	}

	for _, conclusionItem := range fullConclusion {
		conclusionStr, ok := conclusionItem.(string)
		if !ok {
			return genetics.Feature{}, fmt.Errorf(
				"%w: %T",
				ErrUnexpectedType,
				conclusionItem,
			)
		}

		if conclusionStr == "" {
			continue
		}

		feature.Conclusions = append(feature.Conclusions, conclusionStr)
	}

	genes, ok := featureValue["genes"].(map[string]any)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			featureValue["genes"],
		)
	}

	var err error

	if feature.Genes, err = c.parseGenes(genes); err != nil {
		return genetics.Feature{}, err
	}

	recommendation, ok := featureValue["recommendation"].(map[string]any)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			featureValue["recommendation"],
		)
	}

	r, err := c.parseRecommendation(
		recommendation,
		false,
	)
	if err != nil {
		return genetics.Feature{}, err
	}

	feature.Nutrition = r.Nutrition
	feature.Additional = r.Additional
	feature.Checklist = r.Checklist

	return feature, nil
}

func (c *Client) parseWithoutConclusion(
	featureName string,
	featureValue map[string]any,
) (genetics.Feature, error) {
	feature := genetics.Feature{
		Name: featureName,
	}

	conclusion, ok := featureValue["conclusion"].(string)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			featureValue["conclusion"],
		)
	}

	if conclusion != "" {
		if conclusionRisk, ok := featureValue["conclusion_risk"]; ok {
			if conclusionRiskStr, ok := conclusionRisk.(string); ok {
				conclusion = fmt.Sprintf("%s %s", conclusion, conclusionRiskStr)
			}
		}

		feature.Conclusions = append(feature.Conclusions, conclusion)
	}

	genes, ok := featureValue["genes"].(map[string]any)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			featureValue["genes"],
		)
	}

	var err error

	if feature.Genes, err = c.parseGenes(genes); err != nil {
		return genetics.Feature{}, err
	}

	recommendation, ok := featureValue["recommendation"].(map[string]any)
	if !ok {
		return genetics.Feature{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			featureValue["recommendation"],
		)
	}

	r, err := c.parseRecommendation(
		recommendation,
		true,
	)
	if err != nil {
		return genetics.Feature{}, err
	}

	feature.Nutrition = r.Nutrition
	feature.Additional = r.Additional
	feature.Checklist = r.Checklist

	return feature, nil
}

func (c *Client) parseGenes(genes map[string]any) ([]genetics.Gene, error) {
	parsedGenes := make([]genetics.Gene, 0, len(genes))

	for geneName, gene := range genes {
		geneValue, ok := gene.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: %T", ErrUnexpectedType, gene)
		}

		interpretations, ok := geneValue["interpretation"].([]any)
		if !ok {
			return nil, fmt.Errorf(
				"%w: %T",
				ErrUnexpectedType,
				geneValue["interpretation"],
			)
		}

		interpretationsStr := make([]string, 0, len(interpretations))
		for _, interpretationItem := range interpretations {
			item, ok := interpretationItem.(string)
			if !ok {
				return nil, fmt.Errorf(
					"%w: %T",
					ErrUnexpectedType,
					interpretationItem,
				)
			}

			if item == "" {
				continue
			}

			interpretationsStr = append(interpretationsStr, item)
		}

		if len(interpretationsStr) == 0 {
			continue
		}

		parsedGenes = append(parsedGenes, genetics.Gene{
			Name:            geneName,
			Interpretations: interpretationsStr,
		})
	}

	return parsedGenes, nil
}

func (c *Client) parseRecommendation(
	recommendation map[string]any,
	skipChecklist bool,
) (Reccommendation, error) {
	var featureRecommendation Reccommendation

	nutrition, ok := recommendation["nutrition"].([]any)
	if !ok {
		return Reccommendation{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			recommendation["nutrition"],
		)
	}

	for _, nutritionItem := range nutrition {
		nutritionStr, ok := nutritionItem.(string)
		if !ok {
			return Reccommendation{}, fmt.Errorf(
				"%w: %T",
				ErrUnexpectedType,
				nutritionItem,
			)
		}

		if nutritionStr == "" {
			continue
		}

		featureRecommendation.Nutrition = append(
			featureRecommendation.Nutrition,
			nutritionStr,
		)
	}

	additional, ok := recommendation["additional"].([]any)
	if !ok {
		return Reccommendation{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			recommendation["additional"],
		)
	}

	for _, additionalItem := range additional {
		additionalStr, ok := additionalItem.(string)
		if !ok {
			return Reccommendation{}, fmt.Errorf(
				"%w: %T",
				ErrUnexpectedType,
				additionalItem,
			)
		}

		if additionalStr == "" {
			continue
		}

		featureRecommendation.Additional = append(
			featureRecommendation.Additional,
			additionalStr,
		)
	}

	if skipChecklist {
		return featureRecommendation, nil
	}

	checklist, ok := recommendation["checklist"].([]any)
	if !ok {
		return Reccommendation{}, fmt.Errorf(
			"%w: %T",
			ErrUnexpectedType,
			recommendation["checklist"],
		)
	}

	for _, checklistItem := range checklist {
		checklistStr, ok := checklistItem.(string)
		if !ok {
			return Reccommendation{}, fmt.Errorf(
				"%w: %T",
				ErrUnexpectedType,
				checklistItem,
			)
		}

		if checklistStr == "" {
			continue
		}

		featureRecommendation.Checklist = append(
			featureRecommendation.Checklist,
			checklistStr,
		)
	}

	return featureRecommendation, nil
}
