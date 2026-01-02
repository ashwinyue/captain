/**
 * Mock UI Widget 数据
 * 用于开发模式下测试各种 UI Widget 组件
 */

import type {
  OrderWidget,
  LogisticsWidget,
  ProductWidget,
  ProductListWidget,
  PriceComparisonWidget,
} from '@/types/ui-widget';

/**
 * Mock 订单数据
 */
export const mockOrderWidget: OrderWidget = {
  type: 'order',
  order_id: 'ORD-2024-001234',
  status: 'shipped',
  status_text: '已发货',
  created_at: '2024-12-01 10:30:00',
  paid_at: '2024-12-01 10:35:00',
  shipped_at: '2024-12-02 14:00:00',
  items: [
    {
      name: 'Apple iPhone 15 Pro Max 256GB 原色钛金属',
      sku: 'IPHONE15PM-256-NAT',
      quantity: 1,
      unit_price: 9999,
      total_price: 9999,
      image: {
        url: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/iphone-15-pro-finish-select-202309-6-1inch-naturaltitanium?wid=200&hei=200&fmt=jpeg',
        alt: 'iPhone 15 Pro Max',
      },
      attributes: {
        颜色: '原色钛金属',
        容量: '256GB',
      },
    },
    {
      name: 'Apple 20W USB-C 电源适配器',
      sku: 'MHJA3CH/A',
      quantity: 2,
      unit_price: 149,
      total_price: 298,
      image: {
        url: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/MHJA3?wid=200&hei=200&fmt=jpeg',
        alt: '20W 充电器',
      },
    },
  ],
  subtotal: 10297,
  shipping_fee: 0,
  discount: 200,
  total: 10097,
  currency: '¥',
  receiver_name: '张三',
  receiver_phone: '138****8888',
  shipping_address: '上海市浦东新区陆家嘴环路1000号 恒生银行大厦 18层',
  tracking_number: 'SF1234567890',
  carrier: '顺丰速运',
  actions: [
    { label: '查看物流', action: 'msg://帮我查询订单 ORD-2024-001234 的物流信息', style: 'primary' },
    { label: '复制单号', action: 'copy://SF1234567890', style: 'default' },
    { label: '联系客服', action: 'msg://我想咨询订单 ORD-2024-001234 的问题', style: 'link' },
  ],
};

/**
 * Mock 待支付订单
 */
export const mockPendingOrderWidget: OrderWidget = {
  type: 'order',
  order_id: 'ORD-2024-005678',
  status: 'pending',
  status_text: '待支付',
  created_at: '2024-12-06 15:20:00',
  items: [
    {
      name: 'Sony WH-1000XM5 无线降噪头戴式耳机 黑色',
      quantity: 1,
      unit_price: 2499,
      total_price: 2499,
      image: {
        url: 'https://images.unsplash.com/photo-1618366712010-f4ae9c647dcb?w=200&h=200&fit=crop',
        alt: 'Sony 耳机',
      },
      attributes: {
        颜色: '黑色',
      },
    },
  ],
  subtotal: 2499,
  shipping_fee: 10,
  total: 2509,
  currency: '¥',
  receiver_name: '李四',
  receiver_phone: '139****9999',
  shipping_address: '北京市朝阳区建国路88号 SOHO现代城 A座 2001',
  actions: [
    { label: '立即支付', action: 'url://https://example.com/pay/ORD-2024-005678', style: 'primary' },
    { label: '取消订单', action: 'msg://我想取消订单 ORD-2024-005678', style: 'danger' },
  ],
};

/**
 * Mock 物流数据
 */
export const mockLogisticsWidget: LogisticsWidget = {
  type: 'logistics',
  tracking_number: 'SF1234567890',
  carrier: '顺丰速运',
  carrier_phone: '95338',
  status: 'in_transit',
  status_text: '运输中',
  estimated_delivery: '预计12月8日 18:00前送达',
  receiver: '张三',
  receiver_address: '上海市浦东新区陆家嘴环路1000号',
  receiver_phone: '138****8888',
  courier_name: '王师傅',
  courier_phone: '131****1234',
  timeline: [
    {
      time: '2024-12-06 14:30',
      status: 'in_transit',
      description: '快件已到达【上海浦东集散中心】',
      location: '上海市浦东新区',
    },
    {
      time: '2024-12-06 08:15',
      status: 'in_transit',
      description: '快件已从【杭州转运中心】发出，下一站【上海浦东集散中心】',
      location: '浙江省杭州市',
    },
    {
      time: '2024-12-05 22:00',
      status: 'in_transit',
      description: '快件已到达【杭州转运中心】',
      location: '浙江省杭州市',
    },
    {
      time: '2024-12-05 16:30',
      status: 'picked_up',
      description: '快件已揽收',
      location: '浙江省杭州市西湖区',
      operator: '陈师傅',
      phone: '132****5678',
    },
    {
      time: '2024-12-05 15:00',
      status: 'pending',
      description: '商家已通知快递公司揽件',
    },
  ],
  order_id: 'ORD-2024-001234',
  actions: [
    { label: '复制单号', action: 'copy://SF1234567890', style: 'primary' },
    { label: '联系客服', action: 'msg://我想咨询快递单号 SF1234567890 的配送问题', style: 'default' },
  ],
};

/**
 * Mock 产品数据
 */
export const mockProductWidget: ProductWidget = {
  type: 'product',
  product_id: 'PROD-001',
  name: 'Apple MacBook Pro 14英寸 M3 Pro芯片 18GB+512GB',
  description: '配备 M3 Pro 芯片，11核中央处理器，14核图形处理器。18GB 统一内存，512GB 固态硬盘。深空黑色。',
  brand: 'Apple',
  category: '电脑/笔记本',
  thumbnail: {
    url: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/mbp14-m3-pro-max-spaceblack-select-202310?wid=400&hei=400&fmt=jpeg',
    alt: 'MacBook Pro 14',
  },
  price: 16999,
  original_price: 18999,
  currency: '¥',
  discount_label: '限时优惠',
  in_stock: true,
  stock_quantity: 50,
  stock_status: '有货，48小时内发货',
  specs: [
    { name: '芯片', value: 'Apple M3 Pro' },
    { name: '内存', value: '18GB 统一内存' },
    { name: '存储', value: '512GB 固态硬盘' },
    { name: '屏幕', value: '14.2英寸 Liquid Retina XDR' },
  ],
  rating: 4.9,
  review_count: 2856,
  tags: ['新品', '热卖', '官方认证'],
  actions: [
    { label: '立即购买', action: 'url://https://www.apple.com.cn/shop/buy-mac/macbook-pro', style: 'primary' },
    { label: '咨询客服', action: 'msg://我想了解 MacBook Pro 14英寸 M3 Pro 的更多信息', style: 'default' },
    { label: '查看详情', action: 'url://https://www.apple.com.cn/shop/buy-mac/macbook-pro', style: 'link' },
  ],
};

/**
 * Mock 产品列表数据
 */
export const mockProductListWidget: ProductListWidget = {
  type: 'product_list',
  title: '为您推荐',
  subtitle: '根据您的浏览记录推荐以下商品',
  products: [
    {
      product_id: 'PROD-101',
      name: 'AirPods Pro (第二代)',
      price: 1899,
      thumbnail: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/MQD83?wid=200&hei=200&fmt=jpeg',
      rating: 4.8,
    },
    {
      product_id: 'PROD-102',
      name: 'Apple Watch Series 9',
      price: 3299,
      thumbnail: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/watch-s9-702702-702?wid=200&hei=200&fmt=jpeg',
      rating: 4.7,
    },
    {
      product_id: 'PROD-103',
      name: 'iPad Pro 11英寸',
      price: 6999,
      thumbnail: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/ipad-pro-11-select-cell-spacegray-202405?wid=200&hei=200&fmt=jpeg',
      rating: 4.9,
    },
    {
      product_id: 'PROD-104',
      name: 'Magic Keyboard 妙控键盘',
      price: 749,
      thumbnail: 'https://store.storeimages.cdn-apple.com/8756/as-images.apple.com/is/MK2C3CH?wid=200&hei=200&fmt=jpeg',
      rating: 4.6,
    },
  ],
  total_count: 128,
  page: 1,
  page_size: 4,
  has_more: true,
  actions: [
    { label: '查看更多', action: 'msg://帮我推荐更多类似的商品', style: 'primary' },
  ],
};

/**
 * Mock 价格对比数据
 */
export const mockPriceComparisonWidget: PriceComparisonWidget = {
  type: 'price_comparison',
  title: 'iPhone 15 系列价格对比',
  columns: ['型号', '存储', '价格', '优惠后'],
  items: [
    { '型号': 'iPhone 15', '存储': '128GB', '价格': '¥5999', '优惠后': '¥5799' },
    { '型号': 'iPhone 15 Plus', '存储': '128GB', '价格': '¥6999', '优惠后': '¥6799' },
    { '型号': 'iPhone 15 Pro', '存储': '256GB', '价格': '¥8999', '优惠后': '¥8699' },
    { '型号': 'iPhone 15 Pro Max', '存储': '256GB', '价格': '¥9999', '优惠后': '¥9699' },
  ],
  recommended_index: 2,
  recommendation_reason: '根据您的使用需求，iPhone 15 Pro 256GB 版本性价比最高，既有专业级摄像头，又有充足存储空间。',
  actions: [
    { label: '查看推荐型号', action: 'url://https://www.apple.com.cn/shop/buy-iphone/iphone-15-pro', style: 'primary' },
    { label: '咨询购买', action: 'msg://我想购买 iPhone 15 Pro 256GB，有什么优惠吗？', style: 'default' },
  ],
};

/**
 * 所有 Mock Widget 的映射
 */
export const mockWidgets = {
  order: mockOrderWidget,
  order_pending: mockPendingOrderWidget,
  logistics: mockLogisticsWidget,
  product: mockProductWidget,
  product_list: mockProductListWidget,
  price_comparison: mockPriceComparisonWidget,
} as const;

/**
 * Widget 类型描述
 */
export const widgetTypeDescriptions: Record<keyof typeof mockWidgets, string> = {
  order: '订单卡片（已发货）',
  order_pending: '订单卡片（待支付）',
  logistics: '物流追踪',
  product: '商品详情',
  product_list: '商品列表',
  price_comparison: '价格对比',
};

/**
 * 生成包含 Widget 的 Markdown 消息
 */
export function generateMockWidgetMessage(widgetKey: keyof typeof mockWidgets): string {
  const widget = mockWidgets[widgetKey];
  
  const prefix = getWidgetMessagePrefix(widgetKey);
  const suffix = getWidgetMessageSuffix(widgetKey);
  
  const widgetJson = JSON.stringify(widget, null, 2);
  
  // Use array join to avoid template literal escaping issues with backticks
  const codeBlockStart = '```tgo-ui-widget';
  const codeBlockEnd = '```';
  
  return [
    prefix,
    '',
    codeBlockStart,
    widgetJson,
    codeBlockEnd,
    '',
    suffix
  ].join('\n');
}

/**
 * 获取 Widget 消息前缀
 */
function getWidgetMessagePrefix(widgetKey: keyof typeof mockWidgets): string {
  const prefixes: Record<keyof typeof mockWidgets, string> = {
    order: '已为您查询到订单信息：',
    order_pending: '您有一笔待支付的订单：',
    logistics: '以下是您的物流追踪信息：',
    product: '为您找到以下商品：',
    product_list: '根据您的需求，为您推荐以下商品：',
    price_comparison: '以下是不同型号的价格对比：',
  };
  return prefixes[widgetKey];
}

/**
 * 获取 Widget 消息后缀
 */
function getWidgetMessageSuffix(widgetKey: keyof typeof mockWidgets): string {
  const suffixes: Record<keyof typeof mockWidgets, string> = {
    order: '如有任何问题，请随时告诉我。',
    order_pending: '订单将在30分钟后自动取消，请尽快完成支付。',
    logistics: '如需帮助，可以联系快递员或客服。',
    product: '需要了解更多详情或有其他问题吗？',
    product_list: '点击商品可查看详情，有任何问题随时问我。',
    price_comparison: '如需帮助选择，请告诉我您的具体需求。',
  };
  return suffixes[widgetKey];
}

