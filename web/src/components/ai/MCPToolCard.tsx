import React, { useState } from 'react';
import { Edit, Trash2, Copy, Check, Globe, Zap, Radio } from 'lucide-react';
import ConfirmDialog from '../ui/ConfirmDialog';
import type { MCPTool } from '@/types';
import { useTranslation } from 'react-i18next';


interface MCPToolCardProps {
  tool: MCPTool;
  onAction: (actionType: string, tool: MCPTool) => void;
  onShowToast?: (type: 'success' | 'error' | 'warning' | 'info', title: string, message?: string) => void;
}

/**
 * MCP Tool card component
 */
const MCPToolCard: React.FC<MCPToolCardProps> = ({ tool, onAction, onShowToast }) => {
  // 状态管理
  const [isLoading, setIsLoading] = useState<{ [key: string]: boolean }>({});
  const [copiedEndpoint, setCopiedEndpoint] = useState(false);
  const [showConfirmDialog, setShowConfirmDialog] = useState<{
    isOpen: boolean;
    type: string;
    title: string;
    message: string;
    confirmText: string;
    variant: 'danger' | 'primary';
  }>({
    isOpen: false,
    type: '',
    title: '',
    message: '',
    confirmText: '',
    variant: 'primary'
  });

  const { t } = useTranslation();

  // Extract tool data from config
  const toolType = tool.config?.tool_type || 'MCP';
  const transportType = tool.config?.transport_type || 'http';
  const endpoint = tool.config?.endpoint || tool.endpoint || '';

  // 设置加载状态
  const setLoadingState = (actionType: string, loading: boolean) => {
    setIsLoading(prev => ({ ...prev, [actionType]: loading }));
  };

  // 显示Toast提示
  const showToast = (type: 'success' | 'error' | 'warning' | 'info', title: string, message?: string) => {
    onShowToast?.(type, title, message);
  };

  // 显示确认对话框
  const showConfirm = (type: string, title: string, message: string, confirmText: string, variant: 'danger' | 'primary' = 'primary') => {
    setShowConfirmDialog({
      isOpen: true,
      type,
      title,
      message,
      confirmText,
      variant
    });
  };

  // 处理确认对话框确认
  const handleConfirm = async () => {
    const { type } = showConfirmDialog;
    setLoadingState(type, true);

    try {
      if (type === 'delete') {
        // 通知父组件执行删除操作
        await onAction?.(type, tool);
      } else {
        await handleAction(type);
      }
      setShowConfirmDialog(prev => ({ ...prev, isOpen: false }));
    } catch (error) {
      console.error(`Action ${type} failed:`, error);
      showToast(
        'error',
        t('mcp.tools.actionFailed.title', '操作失败'),
        error instanceof Error ? error.message : t('mcp.tools.actionFailed.message', { type, defaultValue: `执行 ${type} 操作时发生错误` })
      );
    } finally {
      setLoadingState(type, false);
    }
  };

  // 处理确认对话框取消
  const handleCancel = () => {
    setShowConfirmDialog(prev => ({ ...prev, isOpen: false }));
  };

  // 主要动作处理函数
  const handleAction = async (actionType: string): Promise<void> => {
    setLoadingState(actionType, true);

    try {
      switch (actionType) {
        case 'delete':
          showConfirm(
            'delete',
            t('mcp.tools.confirmDelete.title', '确认删除'),
            t('mcp.tools.confirmDelete.message', { name: tool.name, defaultValue: `确定要删除工具 "${tool.name}" 吗？此操作无法撤销。` }),
            t('common.delete', '删除'),
            'danger'
          );
          return; // 不设置loading为false，等待确认
        case 'edit':
          // 通知父组件打开编辑模态框
          onAction?.(actionType, tool);
          break;
        default:
          console.warn(`Unknown action type: ${actionType}`);
          showToast('warning', t('mcp.tools.unknownAction.title', '未知操作'), t('mcp.tools.unknownAction.message', { type: actionType, defaultValue: `不支持的操作类型: ${actionType}` }));
      }
    } catch (error) {
      console.error(`Action ${actionType} failed:`, error);
      showToast('error', t('mcp.tools.actionFailed.title', '操作失败'), t('mcp.tools.actionFailed.message', { type: actionType, defaultValue: `执行 ${actionType} 操作时发生错误` }));
    } finally {
      setLoadingState(actionType, false);
    }
  };

  // 复制端点地址
  const handleCopyEndpoint = async () => {
    if (!endpoint) return;

    try {
      await navigator.clipboard.writeText(endpoint);
      setCopiedEndpoint(true);
      showToast('success', t('common.copied', '已复制'), t('mcp.tools.endpointCopied', '端点地址已复制到剪贴板'));
      setTimeout(() => setCopiedEndpoint(false), 2000);
    } catch (error) {
      console.error('Failed to copy endpoint:', error);
      showToast('error', t('common.copyFailed', '复制失败'), t('mcp.tools.copyEndpointFailed', '无法复制端点地址'));
    }
  };

  // 获取传输类型的图标和颜色
  const getTransportTypeConfig = (type: string) => {
    const lowerType = type.toLowerCase();
    switch (lowerType) {
      case 'http':
      case 'https':
        return {
          icon: Globe,
          color: 'text-blue-600 bg-blue-50 border-blue-200 dark:text-blue-400 dark:bg-blue-900/30 dark:border-blue-700',
          label: 'HTTP'
        };
      case 'sse':
        return {
          icon: Radio,
          color: 'text-purple-600 bg-purple-50 border-purple-200 dark:text-purple-400 dark:bg-purple-900/30 dark:border-purple-700',
          label: 'SSE'
        };
      case 'stdio':
        return {
          icon: Zap,
          color: 'text-green-600 bg-green-50 border-green-200 dark:text-green-400 dark:bg-green-900/30 dark:border-green-700',
          label: 'STDIO'
        };
      default:
        return {
          icon: Globe,
          color: 'text-gray-600 bg-gray-50 border-gray-200 dark:text-gray-400 dark:bg-gray-700 dark:border-gray-600',
          label: type.toUpperCase()
        };
    }
  };

  // 脱敏显示端点地址（隐藏敏感信息如 API Key）
  const maskSensitiveInfo = (url: string): string => {
    if (!url) return '';

    try {
      // 匹配常见的 API Key 参数模式
      const maskedUrl = url.replace(
        /([?&])(key|token|apikey|api_key|secret|password|auth)=([^&]+)/gi,
        (_match, prefix, paramName, value) => {
          // 保留前3个字符，其余用 * 替代
          const maskedValue = value.length > 3
            ? value.substring(0, 3) + '*'.repeat(Math.min(value.length - 3, 10))
            : '***';
          return `${prefix}${paramName}=${maskedValue}`;
        }
      );
      return maskedUrl;
    } catch (error) {
      return url;
    }
  };

  // 截断长文本
  const truncateText = (text: string, maxLength: number): string => {
    if (!text || text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

  const transportConfig = getTransportTypeConfig(transportType);
  const TransportIcon = transportConfig.icon;
  const maskedEndpoint = maskSensitiveInfo(endpoint);
  const displayEndpoint = truncateText(maskedEndpoint, 50);



  return (
    <>
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600 hover:shadow-lg transition-all duration-200 overflow-hidden flex flex-col h-full">
        {/* Header Section */}
        <div className="p-4 border-b border-gray-100 dark:border-gray-700">
          <div className="flex items-start justify-between mb-3">
            {/* Tool Name */}
            <div className="flex-1 min-w-0">
              <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 truncate" title={tool.name}>
                {tool.name}
              </h3>
            </div>

            {/* Status Indicator */}
            <div className="flex items-center ml-2">
              <span
                className={`w-2 h-2 rounded-full flex-shrink-0 ${
                  tool.status === 'active' ? 'bg-green-500' :
                  tool.status === 'error' ? 'bg-red-500' : 'bg-gray-400'
                }`}
                title={
                  tool.status === 'active' ? t('common.active', '活跃') :
                  tool.status === 'error' ? t('common.error', '错误') : t('common.inactive', '未激活')
                }
              />
            </div>
          </div>

          {/* Tags: Tool Type & Transport Type */}
          <div className="flex items-center gap-2 flex-wrap">
            {/* Tool Type Badge */}
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-indigo-50 text-indigo-700 border border-indigo-200 dark:bg-indigo-900/30 dark:text-indigo-300 dark:border-indigo-700">
              {toolType}
            </span>

            {/* Transport Type Badge */}
            <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium border ${transportConfig.color}`}>
              <TransportIcon className="w-3 h-3" />
              {transportConfig.label}
            </span>
          </div>
        </div>

        {/* Body Section */}
        <div className="p-4 flex-1 flex flex-col">
          {/* Description */}
          <div className="mb-3">
            <p className="text-sm text-gray-600 dark:text-gray-300 line-clamp-2" title={tool.description}>
              {tool.description || t('mcp.tools.noDescription', '暂无描述')}
            </p>
          </div>

          {/* Endpoint Section */}
          {endpoint && (
            <div className="mt-auto">
              <div className="flex items-start justify-between gap-2 p-2 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600">
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">{t('mcp.tools.endpoint', '端点地址')}</p>
                  <p
                    className="text-xs font-mono text-gray-700 dark:text-gray-200 break-all"
                    title={maskedEndpoint}
                  >
                    {displayEndpoint}
                  </p>
                </div>
                <button
                  onClick={handleCopyEndpoint}
                  className="flex-shrink-0 p-1.5 text-gray-400 dark:text-gray-500 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded transition-colors"
                  title={t('common.copy', '复制')}
                >
                  {copiedEndpoint ? (
                    <Check className="w-3.5 h-3.5 text-green-600 dark:text-green-400" />
                  ) : (
                    <Copy className="w-3.5 h-3.5" />
                  )}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Footer Section - Actions */}
        <div className="px-4 py-3 bg-gray-50 dark:bg-gray-900 border-t border-gray-100 dark:border-gray-700 flex justify-end gap-2">
          <button
            className={`inline-flex items-center gap-1 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 hover:text-green-600 dark:hover:text-green-400 hover:bg-green-50 dark:hover:bg-green-900/30 rounded transition-colors ${
              isLoading.edit ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            title={t('common.edit', '编辑')}
            onClick={() => handleAction('edit')}
            disabled={isLoading.edit}
          >
            <Edit className="w-4 h-4" />
            <span className="hidden sm:inline">{t('common.edit', '编辑')}</span>
          </button>
          <button
            className={`inline-flex items-center gap-1 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-200 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded transition-colors ${
              isLoading.delete ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            title={t('common.delete', '删除')}
            onClick={() => handleAction('delete')}
            disabled={isLoading.delete}
          >
            <Trash2 className="w-4 h-4" />
            <span className="hidden sm:inline">{t('common.delete', '删除')}</span>
          </button>
        </div>
      </div>

      {/* 确认对话框 */}
      <ConfirmDialog
        isOpen={showConfirmDialog.isOpen}
        title={showConfirmDialog.title}
        message={showConfirmDialog.message}
        confirmText={showConfirmDialog.confirmText}
        cancelText={t('common.cancel', '取消')}
        confirmVariant={showConfirmDialog.variant}
        onConfirm={handleConfirm}
        onCancel={handleCancel}
        isLoading={isLoading[showConfirmDialog.type]}
      />
    </>
  );
};

export default MCPToolCard;
