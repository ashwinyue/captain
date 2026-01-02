package orchestration

import (
	"regexp"
	"strings"
)

// IntentType 意图类型
type IntentType string

const (
	IntentUnknown   IntentType = ""          // 需要 LLM 分析
	IntentHuman     IntentType = "human"     // 转人工
	IntentOrder     IntentType = "order"     // 订单查询
	IntentGreeting  IntentType = "greeting"  // 打招呼
	IntentKnowledge IntentType = "knowledge" // 知识库查询
)

// IntentRule 意图规则
type IntentRule struct {
	Intent   IntentType
	Keywords []string         // 关键词匹配
	Patterns []*regexp.Regexp // 正则匹配
}

// IntentRouter 意图前置路由器
// 使用关键词和正则快速匹配意图，减少 LLM 调用
type IntentRouter struct {
	rules []IntentRule
}

// NewIntentRouter 创建意图路由器
func NewIntentRouter() *IntentRouter {
	return &IntentRouter{
		rules: []IntentRule{
			// 转人工意图
			{
				Intent: IntentHuman,
				Keywords: []string{
					"转人工", "人工客服", "人工服务", "找人工",
					"真人", "真人客服", "活人", "转接人工",
					"联系客服", "客服电话", "人工坐席",
				},
			},
			// 订单查询意图
			{
				Intent: IntentOrder,
				Keywords: []string{
					"订单", "订单号", "我的订单", "查订单",
					"退款", "退货", "发货", "物流", "快递",
					"支付", "付款", "付款状态",
				},
				Patterns: []*regexp.Regexp{
					regexp.MustCompile(`#?\d{6,20}`),        // 订单号格式
					regexp.MustCompile(`(?i)order[_\-]?id`), // order_id
				},
			},
			// 打招呼意图
			{
				Intent: IntentGreeting,
				Keywords: []string{
					"你好", "您好", "hi", "hello", "嗨",
					"早上好", "下午好", "晚上好", "早",
				},
			},
		},
	}
}

// QuickMatch 快速意图匹配
// 返回匹配到的意图类型，如果无法确定则返回 IntentUnknown
func (r *IntentRouter) QuickMatch(query string) IntentType {
	queryLower := strings.ToLower(query)

	for _, rule := range r.rules {
		// 关键词匹配
		for _, keyword := range rule.Keywords {
			if strings.Contains(queryLower, strings.ToLower(keyword)) {
				return rule.Intent
			}
		}

		// 正则匹配
		for _, pattern := range rule.Patterns {
			if pattern.MatchString(query) {
				return rule.Intent
			}
		}
	}

	return IntentUnknown
}

// QuickMatchResult 快速匹配结果
type QuickMatchResult struct {
	Intent  IntentType
	Matched bool
	SkipLLM bool   // 是否跳过 LLM 分析
	Reason  string // 匹配原因
}

// Match 执行意图匹配，返回详细结果
func (r *IntentRouter) Match(query string) *QuickMatchResult {
	intent := r.QuickMatch(query)

	switch intent {
	case IntentHuman:
		return &QuickMatchResult{
			Intent:  IntentHuman,
			Matched: true,
			SkipLLM: true,
			Reason:  "关键词匹配：转人工",
		}
	case IntentOrder:
		return &QuickMatchResult{
			Intent:  IntentOrder,
			Matched: true,
			SkipLLM: false, // 订单查询仍需 LLM 判断具体操作
			Reason:  "关键词匹配：订单相关",
		}
	case IntentGreeting:
		return &QuickMatchResult{
			Intent:  IntentGreeting,
			Matched: true,
			SkipLLM: true,
			Reason:  "关键词匹配：打招呼",
		}
	default:
		return &QuickMatchResult{
			Intent:  IntentUnknown,
			Matched: false,
			SkipLLM: false,
			Reason:  "需要 LLM 分析",
		}
	}
}

// AddRule 添加自定义规则
func (r *IntentRouter) AddRule(rule IntentRule) {
	r.rules = append(r.rules, rule)
}

// AddKeywords 为指定意图添加关键词
func (r *IntentRouter) AddKeywords(intent IntentType, keywords ...string) {
	for i := range r.rules {
		if r.rules[i].Intent == intent {
			r.rules[i].Keywords = append(r.rules[i].Keywords, keywords...)
			return
		}
	}
	// 如果意图不存在，创建新规则
	r.rules = append(r.rules, IntentRule{
		Intent:   intent,
		Keywords: keywords,
	})
}
