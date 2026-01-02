package uitpl

// OrderItem represents an item in an order
type OrderItem struct {
	Name     string      `json:"name"`
	Quantity int         `json:"quantity"`
	Price    MoneyAmount `json:"price"`
	Image    *ImageInfo  `json:"image,omitempty"`
}

// OrderTemplate represents order details
type OrderTemplate struct {
	BaseTemplate
	OrderID     string         `json:"order_id"`
	Status      string         `json:"status"`
	TotalAmount MoneyAmount    `json:"total_amount"`
	Items       []OrderItem    `json:"items"`
	CreatedAt   string         `json:"created_at,omitempty"`
	Actions     []ActionButton `json:"actions,omitempty"`
}

func (t *OrderTemplate) GetType() TemplateType { return TemplateOrder }
func (t *OrderTemplate) GetDescription() string {
	return "订单详情模板，展示订单号、状态、金额、商品列表等信息"
}
func (t *OrderTemplate) GetExample() map[string]interface{} {
	return map[string]interface{}{
		"type":     "order",
		"order_id": "ORD-20240101-001",
		"status":   "已发货",
		"total_amount": map[string]interface{}{
			"amount":   299.00,
			"currency": "CNY",
		},
		"items": []map[string]interface{}{
			{
				"name":     "商品名称",
				"quantity": 1,
				"price":    map[string]interface{}{"amount": 299.00, "currency": "CNY"},
			},
		},
		"actions": []map[string]interface{}{
			{"label": "查看物流", "action": "msg://查看这个订单的物流信息", "style": "primary"},
		},
	}
}
func (t *OrderTemplate) ToMarkdown() string { return toMarkdown(t) }

// ProductTemplate represents a single product
type ProductTemplate struct {
	BaseTemplate
	ProductID   string         `json:"product_id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Price       MoneyAmount    `json:"price"`
	Image       *ImageInfo     `json:"image,omitempty"`
	Specs       []string       `json:"specs,omitempty"`
	Actions     []ActionButton `json:"actions,omitempty"`
}

func (t *ProductTemplate) GetType() TemplateType { return TemplateProduct }
func (t *ProductTemplate) GetDescription() string {
	return "单个产品详情模板，展示产品名称、图片、价格、规格等"
}
func (t *ProductTemplate) GetExample() map[string]interface{} {
	return map[string]interface{}{
		"type":        "product",
		"product_id":  "PROD-001",
		"name":        "产品名称",
		"description": "产品描述",
		"price":       map[string]interface{}{"amount": 99.00, "currency": "CNY"},
		"specs":       []string{"规格1", "规格2"},
	}
}
func (t *ProductTemplate) ToMarkdown() string { return toMarkdown(t) }

// ProductListItem represents an item in a product list
type ProductListItem struct {
	ProductID string      `json:"product_id"`
	Name      string      `json:"name"`
	Price     MoneyAmount `json:"price"`
	Image     *ImageInfo  `json:"image,omitempty"`
}

// ProductListTemplate represents a list of products
type ProductListTemplate struct {
	BaseTemplate
	Title    string            `json:"title,omitempty"`
	Products []ProductListItem `json:"products"`
}

func (t *ProductListTemplate) GetType() TemplateType { return TemplateProductList }
func (t *ProductListTemplate) GetDescription() string {
	return "产品列表模板，展示多个产品的概要信息"
}
func (t *ProductListTemplate) GetExample() map[string]interface{} {
	return map[string]interface{}{
		"type":  "product_list",
		"title": "搜索结果",
		"products": []map[string]interface{}{
			{
				"product_id": "PROD-001",
				"name":       "产品1",
				"price":      map[string]interface{}{"amount": 99.00, "currency": "CNY"},
			},
		},
	}
}
func (t *ProductListTemplate) ToMarkdown() string { return toMarkdown(t) }

// LogisticsStep represents a logistics tracking step
type LogisticsStep struct {
	Time        string `json:"time"`
	Status      string `json:"status"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}

// LogisticsTemplate represents logistics tracking information
type LogisticsTemplate struct {
	BaseTemplate
	TrackingNo string          `json:"tracking_no"`
	Carrier    string          `json:"carrier"`
	Status     string          `json:"status"`
	Steps      []LogisticsStep `json:"steps"`
}

func (t *LogisticsTemplate) GetType() TemplateType { return TemplateLogistics }
func (t *LogisticsTemplate) GetDescription() string {
	return "物流跟踪模板，展示快递单号、承运商、物流轨迹等"
}
func (t *LogisticsTemplate) GetExample() map[string]interface{} {
	return map[string]interface{}{
		"type":        "logistics",
		"tracking_no": "SF1234567890",
		"carrier":     "顺丰速运",
		"status":      "派送中",
		"steps": []map[string]interface{}{
			{"time": "2024-01-01 10:00", "status": "已签收", "location": "北京"},
		},
	}
}
func (t *LogisticsTemplate) ToMarkdown() string { return toMarkdown(t) }

// PriceItem represents a price comparison item
type PriceItem struct {
	Source string      `json:"source"`
	Price  MoneyAmount `json:"price"`
	URL    string      `json:"url,omitempty"`
}

// PriceComparisonTemplate represents price comparison data
type PriceComparisonTemplate struct {
	BaseTemplate
	ProductName string      `json:"product_name"`
	Prices      []PriceItem `json:"prices"`
}

func (t *PriceComparisonTemplate) GetType() TemplateType { return TemplatePriceComparison }
func (t *PriceComparisonTemplate) GetDescription() string {
	return "价格对比模板，展示同一商品在不同渠道的价格"
}
func (t *PriceComparisonTemplate) GetExample() map[string]interface{} {
	return map[string]interface{}{
		"type":         "price_comparison",
		"product_name": "iPhone 15",
		"prices": []map[string]interface{}{
			{"source": "京东", "price": map[string]interface{}{"amount": 5999.00, "currency": "CNY"}},
			{"source": "淘宝", "price": map[string]interface{}{"amount": 5899.00, "currency": "CNY"}},
		},
	}
}
func (t *PriceComparisonTemplate) ToMarkdown() string { return toMarkdown(t) }
