/**
 * TGO UI Widget 类型定义
 * 用于解析和渲染 AI Agent 返回的结构化数据
 */

/**
 * UI Widget 基础接口
 */
export interface UIWidgetBase {
  type: string;
  version?: string;
}

/**
 * 订单状态
 */
export type OrderStatus =
  | 'pending'
  | 'paid'
  | 'processing'
  | 'shipped'
  | 'delivered'
  | 'completed'
  | 'cancelled'
  | 'refunded';

/**
 * 订单商品项
 */
export interface OrderItem {
  name: string;
  sku?: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  image?: {
    url: string;
    alt?: string;
  };
  attributes?: Record<string, string>;
}

/**
 * 订单 Widget
 */
export interface OrderWidget extends UIWidgetBase {
  type: 'order';
  order_id: string;
  status: OrderStatus;
  status_text?: string;
  created_at?: string;
  paid_at?: string;
  shipped_at?: string;
  delivered_at?: string;
  items: OrderItem[];
  subtotal: number;
  shipping_fee?: number;
  discount?: number;
  total: number;
  currency?: string;
  shipping_address?: string;
  receiver_name?: string;
  receiver_phone?: string;
  tracking_number?: string;
  carrier?: string;
  actions?: ActionButton[];
}

/**
 * 物流状态
 */
export type LogisticsStatus =
  | 'pending'
  | 'picked_up'
  | 'in_transit'
  | 'out_for_delivery'
  | 'delivered'
  | 'exception'
  | 'returned';

/**
 * 物流事件
 */
export interface LogisticsEvent {
  time: string;
  status?: LogisticsStatus;
  description: string;
  location?: string;
  operator?: string;
  phone?: string;
}

/**
 * 物流 Widget
 */
export interface LogisticsWidget extends UIWidgetBase {
  type: 'logistics';
  tracking_number: string;
  carrier: string;
  carrier_logo?: string;
  carrier_phone?: string;
  status: LogisticsStatus;
  status_text?: string;
  estimated_delivery?: string;
  receiver?: string;
  receiver_address?: string;
  receiver_phone?: string;
  courier_name?: string;
  courier_phone?: string;
  timeline: LogisticsEvent[];
  order_id?: string;
  actions?: ActionButton[];
}

/**
 * 产品 Widget
 */
export interface ProductWidget extends UIWidgetBase {
  type: 'product';
  product_id: string;
  name: string;
  description?: string;
  brand?: string;
  category?: string;
  thumbnail?: {
    url: string;
    alt?: string;
  };
  price: number;
  original_price?: number;
  currency?: string;
  discount_label?: string;
  in_stock?: boolean;
  stock_quantity?: number;
  stock_status?: string;
  specs?: Array<{ name: string; value: string }>;
  rating?: number;
  review_count?: number;
  tags?: string[];
  actions?: ActionButton[];
  url?: string;
}

/**
 * 产品列表 Widget
 */
export interface ProductListWidget extends UIWidgetBase {
  type: 'product_list';
  title?: string;
  subtitle?: string;
  products: Array<{
    product_id: string;
    name: string;
    price: number;
    thumbnail?: string;
    rating?: number;
  }>;
  total_count?: number;
  page?: number;
  page_size?: number;
  has_more?: boolean;
  actions?: ActionButton[];
}

/**
 * 价格对比 Widget
 */
export interface PriceComparisonWidget extends UIWidgetBase {
  type: 'price_comparison';
  title?: string;
  columns: string[];
  items: Array<Record<string, string>>;
  recommended_index?: number;
  recommendation_reason?: string;
  actions?: ActionButton[];
}

/**
 * 操作按钮
 */
export interface ActionButton {
  label: string;
  action: string;
  style?: 'default' | 'primary' | 'danger' | 'link';
  url?: string;
  payload?: Record<string, unknown>;
}

/**
 * 所有 Widget 类型的联合
 */
export type UIWidget =
  | OrderWidget
  | LogisticsWidget
  | ProductWidget
  | ProductListWidget
  | PriceComparisonWidget;

/**
 * 解析结果
 */
export interface ParsedUIBlock {
  type: string;
  data: UIWidget;
  raw: string;
  blockId: string;
  startIndex: number;
  endIndex: number;
}

