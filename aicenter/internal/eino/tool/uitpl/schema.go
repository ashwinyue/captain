package uitpl

import (
	"encoding/json"
	"fmt"
)

// TemplateType represents available UI template types
type TemplateType string

const (
	TemplateOrder           TemplateType = "order"
	TemplateProduct         TemplateType = "product"
	TemplateProductList     TemplateType = "product_list"
	TemplateLogistics       TemplateType = "logistics"
	TemplatePriceComparison TemplateType = "price_comparison"
)

// AllTemplateTypes returns all available template types
func AllTemplateTypes() []TemplateType {
	return []TemplateType{
		TemplateOrder,
		TemplateProduct,
		TemplateProductList,
		TemplateLogistics,
		TemplatePriceComparison,
	}
}

// Template is the interface all UI templates must implement
type Template interface {
	GetType() TemplateType
	GetDescription() string
	GetExample() map[string]interface{}
	ToMarkdown() string
}

// BaseTemplate contains common fields for all templates
type BaseTemplate struct {
	Type    TemplateType `json:"type"`
	Version string       `json:"version,omitempty"`
}

// MoneyAmount represents a monetary amount with currency
type MoneyAmount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency,omitempty"`
}

func (m MoneyAmount) String() string {
	symbols := map[string]string{"CNY": "¥", "USD": "$", "EUR": "€", "GBP": "£"}
	symbol := symbols[m.Currency]
	if symbol == "" {
		symbol = m.Currency
	}
	return fmt.Sprintf("%s%.2f", symbol, m.Amount)
}

// ImageInfo represents image information for display
type ImageInfo struct {
	URL    string `json:"url"`
	Alt    string `json:"alt,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// ActionProtocol represents action URI protocol types
type ActionProtocol string

const (
	ActionURL  ActionProtocol = "url"  // External link: url://https://example.com
	ActionMsg  ActionProtocol = "msg"  // Send message: msg://查看订单
	ActionCopy ActionProtocol = "copy" // Copy to clipboard: copy://SF1234567890
)

// ButtonStyle represents button style types
type ButtonStyle string

const (
	ButtonDefault ButtonStyle = "default"
	ButtonPrimary ButtonStyle = "primary"
	ButtonDanger  ButtonStyle = "danger"
	ButtonLink    ButtonStyle = "link"
	ButtonGhost   ButtonStyle = "ghost"
)

// ActionButton represents an interactive button
type ActionButton struct {
	Label  string      `json:"label"`
	Action string      `json:"action"`
	Style  ButtonStyle `json:"style,omitempty"`
}

// toMarkdown converts any template data to tgo-ui-widget markdown block
func toMarkdown(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return fmt.Sprintf("```tgo-ui-widget\n%s\n```", string(jsonBytes))
}
