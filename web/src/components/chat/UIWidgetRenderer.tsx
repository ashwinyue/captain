/**
 * TGO UI Widget 渲染器
 * 根据 widget type 分发到对应的 UI 组件
 */

import React from 'react';
import {
  Package,
  Truck,
  ShoppingBag,
  Phone,
  MapPin,
  Star,
  Clock,
  CheckCircle,
  AlertCircle,
  XCircle,
  ExternalLink,
} from 'lucide-react';
import type {
  UIWidget,
  OrderWidget,
  LogisticsWidget,
  ProductWidget,
  ProductListWidget,
  PriceComparisonWidget,
  ActionButton,
  OrderStatus,
  LogisticsStatus,
} from '@/types/ui-widget';

interface UIWidgetRendererProps {
  widget: UIWidget;
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}

/**
 * UI Widget 渲染器 - 根据 type 分发到对应组件
 */
export const UIWidgetRenderer: React.FC<UIWidgetRendererProps> = ({
  widget,
  onAction,
}) => {
  switch (widget.type) {
    case 'order':
      return <OrderCard data={widget} onAction={onAction} />;
    case 'logistics':
      return <LogisticsCard data={widget} onAction={onAction} />;
    case 'product':
      return <ProductCard data={widget} onAction={onAction} />;
    case 'product_list':
      return <ProductListCard data={widget} onAction={onAction} />;
    case 'price_comparison':
      return <PriceComparisonCard data={widget} onAction={onAction} />;
    default:
      return <UnknownWidget type={(widget as UIWidget).type} />;
  }
};

/**
 * 操作按钮组件
 */
const ActionButtons: React.FC<{
  actions?: ActionButton[];
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}> = ({ actions, onAction }) => {
  if (!actions || actions.length === 0) return null;

  return (
    <div className="mt-4 flex flex-wrap gap-2">
      {actions.map((action, index) => (
        <button
          key={index}
          onClick={() => {
            if (action.url) {
              window.open(action.url, '_blank', 'noopener,noreferrer');
            } else {
              onAction?.(action.action, action.payload);
            }
          }}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            action.style === 'primary'
              ? 'bg-blue-600 text-white hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600'
              : action.style === 'danger'
                ? 'bg-red-500 text-white hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700'
                : action.style === 'link'
                  ? 'text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 underline'
                  : 'border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
          }`}
        >
          {action.label}
          {action.url && <ExternalLink className="inline-block w-3 h-3 ml-1" />}
        </button>
      ))}
    </div>
  );
};

/**
 * 订单状态配色映射
 */
const orderStatusConfig: Record<OrderStatus, { bg: string; text: string; icon: React.ReactNode }> = {
  pending: { bg: 'bg-yellow-100 dark:bg-yellow-900/30', text: 'text-yellow-800 dark:text-yellow-300', icon: <Clock className="w-4 h-4" /> },
  paid: { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-800 dark:text-blue-300', icon: <CheckCircle className="w-4 h-4" /> },
  processing: { bg: 'bg-indigo-100 dark:bg-indigo-900/30', text: 'text-indigo-800 dark:text-indigo-300', icon: <Package className="w-4 h-4" /> },
  shipped: { bg: 'bg-purple-100 dark:bg-purple-900/30', text: 'text-purple-800 dark:text-purple-300', icon: <Truck className="w-4 h-4" /> },
  delivered: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-800 dark:text-green-300', icon: <CheckCircle className="w-4 h-4" /> },
  completed: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-800 dark:text-green-300', icon: <CheckCircle className="w-4 h-4" /> },
  cancelled: { bg: 'bg-gray-100 dark:bg-gray-700', text: 'text-gray-800 dark:text-gray-300', icon: <XCircle className="w-4 h-4" /> },
  refunded: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-800 dark:text-red-300', icon: <AlertCircle className="w-4 h-4" /> },
};

/**
 * 订单卡片组件
 */
const OrderCard: React.FC<{
  data: OrderWidget;
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}> = ({ data, onAction }) => {
  const statusStyle = orderStatusConfig[data.status] || orderStatusConfig.pending;
  const currency = data.currency || '¥';

  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-xl p-5 shadow-sm bg-white dark:bg-gray-800 my-3">
      {/* 头部 */}
      <div className="flex justify-between items-start mb-4">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-blue-50 dark:bg-blue-900/30 rounded-lg">
            <Package className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <span className="text-xs text-gray-500 dark:text-gray-400">订单号</span>
            <p className="font-mono font-semibold text-gray-900 dark:text-gray-100">{data.order_id}</p>
          </div>
        </div>
        <span className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium ${statusStyle.bg} ${statusStyle.text}`}>
          {statusStyle.icon}
          {data.status_text || data.status}
        </span>
      </div>

      {/* 商品列表 */}
      <div className="border-t border-b border-gray-100 dark:border-gray-700 py-3 my-3 space-y-3">
        {data.items.map((item, index) => (
          <div key={index} className="flex justify-between items-center">
            <div className="flex items-center gap-3">
              {item.image && (
                <img
                  src={item.image.url}
                  alt={item.image.alt || item.name}
                  className="w-14 h-14 object-cover rounded-lg border border-gray-100 dark:border-gray-700"
                />
              )}
              <div>
                <p className="font-medium text-gray-900 dark:text-gray-100">{item.name}</p>
                {item.attributes && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {Object.entries(item.attributes).map(([k, v]) => `${k}: ${v}`).join(' | ')}
                  </p>
                )}
                {item.sku && (
                  <p className="text-xs text-gray-400 dark:text-gray-500">SKU: {item.sku}</p>
                )}
              </div>
            </div>
            <div className="text-right">
              <p className="text-sm text-gray-500 dark:text-gray-400">×{item.quantity}</p>
              <p className="font-medium text-gray-900 dark:text-gray-100">{currency}{item.total_price.toFixed(2)}</p>
            </div>
          </div>
        ))}
      </div>

      {/* 金额信息 */}
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-500 dark:text-gray-400">商品小计</span>
          <span className="text-gray-700 dark:text-gray-300">{currency}{data.subtotal.toFixed(2)}</span>
        </div>
        {data.shipping_fee !== undefined && data.shipping_fee > 0 && (
          <div className="flex justify-between">
            <span className="text-gray-500 dark:text-gray-400">运费</span>
            <span className="text-gray-700 dark:text-gray-300">{currency}{data.shipping_fee.toFixed(2)}</span>
          </div>
        )}
        {data.discount !== undefined && data.discount > 0 && (
          <div className="flex justify-between text-red-500 dark:text-red-400">
            <span>优惠</span>
            <span>-{currency}{data.discount.toFixed(2)}</span>
          </div>
        )}
        <div className="flex justify-between font-semibold text-base pt-2 border-t border-gray-100 dark:border-gray-700">
          <span className="text-gray-900 dark:text-gray-100">合计</span>
          <span className="text-red-500 dark:text-red-400">{currency}{data.total.toFixed(2)}</span>
        </div>
      </div>

      {/* 收货信息 */}
      {data.shipping_address && (
        <div className="mt-4 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg text-sm">
          <div className="flex items-start gap-2">
            <MapPin className="w-4 h-4 text-gray-400 mt-0.5 flex-shrink-0" />
            <div>
              <p className="font-medium text-gray-900 dark:text-gray-100">
                {data.receiver_name} {data.receiver_phone}
              </p>
              <p className="text-gray-600 dark:text-gray-400">{data.shipping_address}</p>
            </div>
          </div>
        </div>
      )}

      {/* 物流信息 */}
      {data.tracking_number && (
        <div className="mt-3 flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
          <Truck className="w-4 h-4" />
          <span>{data.carrier}</span>
          <span className="font-mono">{data.tracking_number}</span>
        </div>
      )}

      {/* 操作按钮 */}
      <ActionButtons actions={data.actions} onAction={onAction} />
    </div>
  );
};

/**
 * 物流状态配色映射
 */
const logisticsStatusConfig: Record<LogisticsStatus, { color: string; bgColor: string }> = {
  pending: { color: 'text-gray-600 dark:text-gray-400', bgColor: 'bg-gray-400' },
  picked_up: { color: 'text-blue-600 dark:text-blue-400', bgColor: 'bg-blue-500' },
  in_transit: { color: 'text-purple-600 dark:text-purple-400', bgColor: 'bg-purple-500' },
  out_for_delivery: { color: 'text-orange-600 dark:text-orange-400', bgColor: 'bg-orange-500' },
  delivered: { color: 'text-green-600 dark:text-green-400', bgColor: 'bg-green-500' },
  exception: { color: 'text-red-600 dark:text-red-400', bgColor: 'bg-red-500' },
  returned: { color: 'text-gray-600 dark:text-gray-400', bgColor: 'bg-gray-500' },
};

/**
 * 物流卡片组件
 */
const LogisticsCard: React.FC<{
  data: LogisticsWidget;
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}> = ({ data, onAction }) => {
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-xl p-5 shadow-sm bg-white dark:bg-gray-800 my-3">
      {/* 头部 */}
      <div className="flex justify-between items-start mb-4">
        <div className="flex items-center gap-3">
          {data.carrier_logo ? (
            <img src={data.carrier_logo} alt={data.carrier} className="w-10 h-10 rounded-lg" />
          ) : (
            <div className="p-2 bg-purple-50 dark:bg-purple-900/30 rounded-lg">
              <Truck className="w-5 h-5 text-purple-600 dark:text-purple-400" />
            </div>
          )}
          <div>
            <h3 className="font-semibold text-gray-900 dark:text-gray-100">{data.carrier}</h3>
            <p className="text-sm text-gray-500 dark:text-gray-400 font-mono">{data.tracking_number}</p>
          </div>
        </div>
        <span className="px-3 py-1.5 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 rounded-full text-sm font-medium">
          {data.status_text || data.status}
        </span>
      </div>

      {/* 预计送达 */}
      {data.estimated_delivery && (
        <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400 mb-4 p-2 bg-green-50 dark:bg-green-900/20 rounded-lg">
          <Clock className="w-4 h-4" />
          <span>预计送达: {data.estimated_delivery}</span>
        </div>
      )}

      {/* 配送员信息 */}
      {data.courier_name && (
        <div className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg mb-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-600 dark:text-gray-400">
              配送员: <span className="font-medium text-gray-900 dark:text-gray-100">{data.courier_name}</span>
            </span>
            {data.courier_phone && (
              <a
                href={`tel:${data.courier_phone}`}
                className="inline-flex items-center gap-1 text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
              >
                <Phone className="w-4 h-4" />
                <span className="text-sm">{data.courier_phone}</span>
              </a>
            )}
          </div>
        </div>
      )}

      {/* 时间线 */}
      <div className="space-y-0">
        {data.timeline.map((event, index) => {
          const eventStatusStyle = event.status
            ? logisticsStatusConfig[event.status]
            : (index === 0 ? { bgColor: 'bg-blue-500' } : { bgColor: 'bg-gray-300 dark:bg-gray-600' });

          return (
            <div key={index} className="flex gap-3">
              <div className="flex flex-col items-center">
                <div className={`w-3 h-3 rounded-full ${eventStatusStyle.bgColor} ring-4 ring-white dark:ring-gray-800`} />
                {index < data.timeline.length - 1 && (
                  <div className="w-0.5 flex-1 bg-gray-200 dark:bg-gray-700 my-1 min-h-[24px]" />
                )}
              </div>
              <div className="flex-1 pb-4">
                <p className="text-xs text-gray-500 dark:text-gray-400">{event.time}</p>
                <p className={`text-sm ${index === 0 ? 'font-medium text-gray-900 dark:text-gray-100' : 'text-gray-700 dark:text-gray-300'}`}>
                  {event.description}
                </p>
                {event.location && (
                  <p className="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1 mt-0.5">
                    <MapPin className="w-3 h-3" />
                    {event.location}
                  </p>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* 操作按钮 */}
      <ActionButtons actions={data.actions} onAction={onAction} />
    </div>
  );
};

/**
 * 产品卡片组件
 */
const ProductCard: React.FC<{
  data: ProductWidget;
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}> = ({ data, onAction }) => {
  const currency = data.currency || '¥';
  const hasDiscount = data.original_price && data.original_price > data.price;

  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm bg-white dark:bg-gray-800 my-3">
      {/* 图片 */}
      {data.thumbnail && (
        <div className="relative">
          <img
            src={data.thumbnail.url}
            alt={data.thumbnail.alt || data.name}
            className="w-full h-48 object-cover"
          />
          {data.discount_label && (
            <span className="absolute top-2 left-2 px-2 py-1 bg-red-500 text-white text-xs font-medium rounded">
              {data.discount_label}
            </span>
          )}
        </div>
      )}

      <div className="p-4">
        {/* 标签 */}
        {data.tags && data.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {data.tags.slice(0, 3).map((tag, index) => (
              <span
                key={index}
                className="px-2 py-0.5 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-xs rounded"
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* 名称 */}
        <h3 className="font-semibold text-lg text-gray-900 dark:text-gray-100 line-clamp-2">{data.name}</h3>

        {/* 品牌 */}
        {data.brand && (
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{data.brand}</p>
        )}

        {/* 描述 */}
        {data.description && (
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-2 line-clamp-2">{data.description}</p>
        )}

        {/* 价格 */}
        <div className="mt-3 flex items-baseline gap-2">
          <span className="text-2xl text-red-500 dark:text-red-400 font-bold">
            {currency}{data.price.toFixed(2)}
          </span>
          {hasDiscount && (
            <span className="text-sm text-gray-400 dark:text-gray-500 line-through">
              {currency}{data.original_price!.toFixed(2)}
            </span>
          )}
        </div>

        {/* 评分 */}
        {data.rating !== undefined && (
          <div className="mt-2 flex items-center gap-1.5 text-sm">
            <Star className="w-4 h-4 text-yellow-400 fill-yellow-400" />
            <span className="font-medium text-gray-900 dark:text-gray-100">{data.rating.toFixed(1)}</span>
            {data.review_count !== undefined && (
              <span className="text-gray-400 dark:text-gray-500">({data.review_count}条评价)</span>
            )}
          </div>
        )}

        {/* 库存状态 */}
        <p className={`mt-2 text-sm font-medium ${data.in_stock !== false ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'}`}>
          {data.in_stock !== false ? (data.stock_status || '有货') : '暂时缺货'}
        </p>

        {/* 规格 */}
        {data.specs && data.specs.length > 0 && (
          <div className="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700 space-y-1">
            {data.specs.slice(0, 4).map((spec, index) => (
              <div key={index} className="flex text-sm">
                <span className="text-gray-500 dark:text-gray-400 w-20 flex-shrink-0">{spec.name}</span>
                <span className="text-gray-700 dark:text-gray-300">{spec.value}</span>
              </div>
            ))}
          </div>
        )}

        {/* 操作按钮 */}
        <ActionButtons actions={data.actions} onAction={onAction} />
      </div>
    </div>
  );
};

/**
 * 产品列表卡片组件
 */
const ProductListCard: React.FC<{
  data: ProductListWidget;
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}> = ({ data, onAction }) => {
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-xl p-5 shadow-sm bg-white dark:bg-gray-800 my-3">
      {/* 标题 */}
      {(data.title || data.subtitle) && (
        <div className="mb-4">
          {data.title && (
            <h3 className="font-semibold text-lg text-gray-900 dark:text-gray-100 flex items-center gap-2">
              <ShoppingBag className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              {data.title}
            </h3>
          )}
          {data.subtitle && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{data.subtitle}</p>
          )}
        </div>
      )}

      {/* 产品网格 */}
      <div className="grid grid-cols-2 gap-3">
        {data.products.map((product, index) => (
          <div
            key={index}
            className="border border-gray-100 dark:border-gray-700 rounded-lg p-3 hover:shadow-md transition-shadow cursor-pointer bg-gray-50 dark:bg-gray-700/50"
          >
            {product.thumbnail && (
              <img
                src={product.thumbnail}
                alt={product.name}
                className="w-full h-24 object-cover rounded mb-2"
              />
            )}
            <p className="font-medium text-sm text-gray-900 dark:text-gray-100 truncate">{product.name}</p>
            <div className="flex items-center justify-between mt-1">
              <span className="text-red-500 dark:text-red-400 font-semibold">¥{product.price}</span>
              {product.rating !== undefined && (
                <span className="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-0.5">
                  <Star className="w-3 h-3 text-yellow-400 fill-yellow-400" />
                  {product.rating}
                </span>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* 分页信息 */}
      {data.total_count !== undefined && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-3 text-center">
          共 {data.total_count} 件商品
          {data.has_more && ' · 还有更多'}
        </p>
      )}

      {/* 操作按钮 */}
      <ActionButtons actions={data.actions} onAction={onAction} />
    </div>
  );
};

/**
 * 价格对比卡片组件
 */
const PriceComparisonCard: React.FC<{
  data: PriceComparisonWidget;
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
}> = ({ data, onAction }) => {
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-xl p-5 shadow-sm bg-white dark:bg-gray-800 my-3">
      {/* 标题 */}
      {data.title && (
        <h3 className="font-semibold text-lg text-gray-900 dark:text-gray-100 mb-4">{data.title}</h3>
      )}

      {/* 表格 */}
      <div className="overflow-x-auto -mx-2">
        <table className="w-full border-collapse min-w-full">
          <thead>
            <tr>
              {data.columns.map((col, i) => (
                <th
                  key={i}
                  className="px-3 py-2 text-left text-sm font-semibold text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-700 border border-gray-200 dark:border-gray-600"
                >
                  {col}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.items.map((item, i) => (
              <tr
                key={i}
                className={`${
                  i === data.recommended_index
                    ? 'bg-green-50 dark:bg-green-900/20'
                    : ''
                }`}
              >
                {data.columns.map((col, j) => (
                  <td
                    key={j}
                    className={`px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 ${
                      i === data.recommended_index
                        ? 'text-green-800 dark:text-green-300'
                        : 'text-gray-700 dark:text-gray-300'
                    }`}
                  >
                    {item[col]}
                    {i === data.recommended_index && j === 0 && (
                      <span className="ml-2 inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 dark:bg-green-800 text-green-700 dark:text-green-200">
                        推荐
                      </span>
                    )}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* 推荐原因 */}
      {data.recommendation_reason && (
        <p className="mt-3 text-sm text-green-600 dark:text-green-400 flex items-start gap-1.5">
          <span className="text-base">💡</span>
          <span>{data.recommendation_reason}</span>
        </p>
      )}

      {/* 操作按钮 */}
      <ActionButtons actions={data.actions} onAction={onAction} />
    </div>
  );
};

/**
 * 未知类型组件
 */
const UnknownWidget: React.FC<{ type: string }> = ({ type }) => (
  <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded-lg my-3">
    <p className="text-yellow-800 dark:text-yellow-300 flex items-center gap-2">
      <AlertCircle className="w-5 h-5" />
      未知的 UI 组件类型: {type}
    </p>
  </div>
);

export default UIWidgetRenderer;

