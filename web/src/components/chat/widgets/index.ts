/**
 * UI Widget 系统入口
 * 
 * 如何添加新的 Widget：
 * 1. 在 types.ts 中定义 Widget 数据类型（如 MyWidgetData）
 * 2. 创建 Widget 组件文件（如 MyWidget.tsx）
 * 3. 在组件文件中导出 WidgetDefinition
 * 4. 在本文件中导入并注册 Widget
 * 
 * 示例：
 * ```typescript
 * // MyWidget.tsx
 * export const myWidgetDefinition: WidgetDefinition<MyWidgetData> = {
 *   type: 'my_widget',
 *   displayName: '我的组件',
 *   component: MyWidgetComponent,
 *   icon: <MyIcon />,
 * };
 * 
 * // index.ts
 * import { myWidgetDefinition } from './MyWidget';
 * registerWidget(myWidgetDefinition);
 * ```
 */

// 导出类型
export * from './types';

// 导出注册表函数
export {
  registerWidget,
  registerWidgets,
  getWidgetDefinition,
  getRegisteredTypes,
  getAllWidgetDefinitions,
  isWidgetRegistered,
} from './registry';

// 导出共享组件
export * from './shared';

// 导出渲染器
export { default as WidgetRenderer } from './WidgetRenderer';

// ============================================
// 注册内置 Widget
// ============================================

import { registerWidget } from './registry';

// 导入内置 Widget 定义
import { orderWidgetDefinition } from './OrderWidget';
import { logisticsWidgetDefinition } from './LogisticsWidget';
import { productWidgetDefinition } from './ProductWidget';
import { productListWidgetDefinition } from './ProductListWidget';
import { priceComparisonWidgetDefinition } from './PriceComparisonWidget';

// 注册内置 Widget
registerWidget(orderWidgetDefinition);
registerWidget(logisticsWidgetDefinition);
registerWidget(productWidgetDefinition);
registerWidget(productListWidgetDefinition);
registerWidget(priceComparisonWidgetDefinition);

// 导出内置 Widget 定义（供开发模式工具栏使用）
export const builtinWidgetDefinitions = [
  orderWidgetDefinition,
  logisticsWidgetDefinition,
  productWidgetDefinition,
  productListWidgetDefinition,
  priceComparisonWidgetDefinition,
];

