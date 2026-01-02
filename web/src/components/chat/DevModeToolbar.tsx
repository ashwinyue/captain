/**
 * 开发模式工具栏
 * 用于在聊天界面快速发送测试 UI Widget 消息
 */

import React, { useState } from 'react';
import { Code2, ChevronUp, ChevronDown, Sparkles, CreditCard } from 'lucide-react';
import { builtinWidgetDefinitions } from './widgets';
import { mockWidgets, generateMockWidgetMessage } from '@/data/mockUIWidgets';

interface DevModeToolbarProps {
  /** 发送消息的回调函数 */
  onSendMessage: (content: string, isFromAssistant?: boolean) => void;
}

const DevModeToolbar: React.FC<DevModeToolbarProps> = ({ onSendMessage }) => {
  const [isExpanded, setIsExpanded] = useState(true);

  const handleSendWidget = (widgetType: string) => {
    // 查找对应的 mock 数据 key
    const mockKey = widgetType as keyof typeof mockWidgets;
    if (mockWidgets[mockKey]) {
      const message = generateMockWidgetMessage(mockKey);
      onSendMessage(message, true);
    }
  };

  return (
    <div className="border-t border-gray-200 dark:border-gray-700 bg-gradient-to-r from-purple-50 to-indigo-50 dark:from-purple-900/20 dark:to-indigo-900/20">
      {/* 折叠/展开按钮 */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-center gap-2 py-2 text-sm text-purple-600 dark:text-purple-400 hover:bg-purple-100/50 dark:hover:bg-purple-900/30 transition-colors"
      >
        <Code2 className="w-4 h-4" />
        <span className="font-medium">开发模式 - UI Widget 调试</span>
        {isExpanded ? (
          <ChevronDown className="w-4 h-4" />
        ) : (
          <ChevronUp className="w-4 h-4" />
        )}
      </button>

      {/* 工具栏内容 */}
      {isExpanded && (
        <div className="px-4 pb-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-3 flex items-center gap-1">
            <Sparkles className="w-3 h-3" />
            点击按钮发送对应的测试消息（以 AI 助手身份）
          </p>
          
          <div className="flex flex-wrap gap-2">
            {/* 从已注册的 Widget 定义生成按钮 */}
            {builtinWidgetDefinitions.map((def) => (
              <button
                key={def.type}
                onClick={() => handleSendWidget(def.type)}
                className={`
                  inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium
                  transition-all duration-200 border border-transparent
                  ${def.toolbarColor || 'text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-900/30 hover:bg-gray-100'}
                `}
                title={def.description}
              >
                {def.icon}
                <span>{def.displayName}</span>
              </button>
            ))}
            
            {/* 额外的待支付订单按钮（使用不同的 mock 数据） */}
            <button
              onClick={() => {
                const message = generateMockWidgetMessage('order_pending');
                onSendMessage(message, true);
              }}
              className="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 border border-transparent text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/30 hover:bg-yellow-100 dark:hover:bg-yellow-900/50"
              title="显示待支付状态的订单"
            >
              <CreditCard className="w-4 h-4" />
              <span>订单(待支付)</span>
            </button>
          </div>

          <div className="mt-3 pt-3 border-t border-purple-200/50 dark:border-purple-700/50">
            <p className="text-xs text-gray-400 dark:text-gray-500">
              💡 提示：这些按钮会模拟 AI 返回包含 UI Widget 的消息，用于测试渲染效果。
              在设置 → 通用中可关闭开发模式。
            </p>
          </div>
        </div>
      )}
    </div>
  );
};

export default DevModeToolbar;

