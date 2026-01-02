/**
 * 未知类型 Widget 组件
 * 当遇到未注册的 Widget 类型时显示
 */

import React from 'react';
import { AlertCircle } from 'lucide-react';
import type { WidgetComponentProps, WidgetData } from './types';

/**
 * 未知 Widget 组件
 */
const UnknownWidgetComponent: React.FC<WidgetComponentProps<WidgetData>> = ({ data }) => (
  <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded-lg my-3">
    <p className="text-yellow-800 dark:text-yellow-300 flex items-center gap-2">
      <AlertCircle className="w-5 h-5" />
      未知的 UI 组件类型: {data.type}
    </p>
    <details className="mt-2">
      <summary className="text-sm text-yellow-600 dark:text-yellow-400 cursor-pointer">
        查看原始数据
      </summary>
      <pre className="mt-2 p-2 bg-yellow-100 dark:bg-yellow-900/30 rounded text-xs overflow-auto max-h-40">
        {JSON.stringify(data, null, 2)}
      </pre>
    </details>
  </div>
);

export default UnknownWidgetComponent;

