/**
 * UI Widget 注册表
 * 管理所有已注册的 Widget 类型
 */

import type { WidgetDefinition, WidgetData, WidgetRegistry } from './types';

/**
 * Widget 注册表实例
 */
const widgetRegistry: WidgetRegistry = new Map();

/**
 * 注册一个 Widget
 * @param definition Widget 定义
 */
export function registerWidget<T extends WidgetData>(definition: WidgetDefinition<T>): void {
  if (widgetRegistry.has(definition.type)) {
    console.warn(`[WidgetRegistry] Widget type "${definition.type}" is already registered. Overwriting.`);
  }
  widgetRegistry.set(definition.type, definition as WidgetDefinition);
}

/**
 * 批量注册 Widgets
 * @param definitions Widget 定义数组
 */
export function registerWidgets(definitions: WidgetDefinition[]): void {
  definitions.forEach(def => registerWidget(def));
}

/**
 * 获取 Widget 定义
 * @param type Widget 类型
 */
export function getWidgetDefinition(type: string): WidgetDefinition | undefined {
  return widgetRegistry.get(type);
}

/**
 * 获取所有已注册的 Widget 类型
 */
export function getRegisteredTypes(): string[] {
  return Array.from(widgetRegistry.keys());
}

/**
 * 获取所有 Widget 定义（用于开发模式工具栏）
 */
export function getAllWidgetDefinitions(): WidgetDefinition[] {
  return Array.from(widgetRegistry.values());
}

/**
 * 检查 Widget 类型是否已注册
 * @param type Widget 类型
 */
export function isWidgetRegistered(type: string): boolean {
  return widgetRegistry.has(type);
}

/**
 * 清空注册表（主要用于测试）
 */
export function clearRegistry(): void {
  widgetRegistry.clear();
}

/**
 * 获取注册表大小
 */
export function getRegistrySize(): number {
  return widgetRegistry.size;
}
