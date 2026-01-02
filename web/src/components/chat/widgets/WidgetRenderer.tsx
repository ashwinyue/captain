/**
 * Widget 渲染器
 * 根据 Widget 类型自动选择对应的组件进行渲染
 */

import React from 'react';
import { getWidgetDefinition } from './registry';
import UnknownWidget from './UnknownWidget';
import type { WidgetData } from './types';

export interface WidgetRendererProps {
  /** Widget 数据 */
  data: WidgetData;
  /** 操作按钮点击回调（用于非标准 URI 或向后兼容） */
  onAction?: (action: string, payload?: Record<string, unknown>) => void;
  /** 发送消息回调（用于 msg:// 协议） */
  onSendMessage?: (message: string) => void;
}

/**
 * Widget 渲染器组件
 * 自动根据 data.type 查找已注册的 Widget 组件进行渲染
 */
const WidgetRenderer: React.FC<WidgetRendererProps> = ({ data, onAction, onSendMessage }) => {
  const definition = getWidgetDefinition(data.type);

  if (!definition) {
    return <UnknownWidget data={data} onAction={onAction} onSendMessage={onSendMessage} />;
  }

  const Component = definition.component;
  return <Component data={data} onAction={onAction} onSendMessage={onSendMessage} />;
};

export default WidgetRenderer;

