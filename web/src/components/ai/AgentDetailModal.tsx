import React from 'react';
import { useTranslation } from 'react-i18next';
import { X, Bot, Wrench, FolderOpen, Activity, Clock, MessageSquare, TrendingUp } from 'lucide-react';
import AgentToolsList from '@/components/ui/AgentToolsList';
import MCPToolsTagList from '@/components/ui/MCPToolsList';
import OptimizedTransforms from '@/utils/mcpToolsTransformOptimized';
import { useMCPToolsStore } from '@/stores/mcpToolsStore';
import type { Agent, AgentToolResponse } from '@/types';

interface AgentDetailModalProps {
  agent: Agent | null;
  isOpen: boolean;
  onClose: () => void;
  onToolClick?: (tool: AgentToolResponse) => void;
}

/**
 * AI员工详情模态框组件
 */
const AgentDetailModal: React.FC<AgentDetailModalProps> = ({
  agent,
  isOpen,
  onClose,
  onToolClick
}) => {
  const { t } = useTranslation();
  if (!isOpen || !agent) return null;

  // Build tools map for tag rendering (from marketplace/tools API)
  const { tools: apiTools } = useMCPToolsStore();
  const mcpTools = OptimizedTransforms.storeItemsToMCPTools(
    OptimizedTransforms.toolSummariesToStoreItems(apiTools)
  );
  const toolsMap = React.useMemo(() => Object.fromEntries(mcpTools.map(t => [t.id, t])), [mcpTools]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'text-green-600 bg-green-50 border-green-200 dark:text-green-400 dark:bg-green-900/30 dark:border-green-700';
      case 'inactive':
        return 'text-gray-600 bg-gray-50 border-gray-200 dark:text-gray-400 dark:bg-gray-700 dark:border-gray-600';
      case 'error':
        return 'text-red-600 bg-red-50 border-red-200 dark:text-red-400 dark:bg-red-900/30 dark:border-red-700';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200 dark:text-gray-400 dark:bg-gray-700 dark:border-gray-600';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active':
        return t('agents.card.status.active', '运行中');
      case 'inactive':
        return t('agents.card.status.inactive', '未激活');
      case 'error':
        return t('agents.card.status.error', '错误');
      default:
        return t('agents.card.status.unknown', '未知');
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-4">
            <img
              src={agent.avatar}
              alt={agent.name}
              className="w-12 h-12 rounded-full object-cover border border-gray-200 dark:border-gray-600"
            />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">{agent.name}</h2>
              <p className="text-sm text-gray-600 dark:text-gray-300">{agent.role || t('agents.card.defaultRole', 'AI员工')}</p>
            </div>
            <div className={`flex items-center space-x-2 px-3 py-1 rounded-full border ${getStatusColor(agent.status)}`}>
              <div className={`w-2 h-2 rounded-full ${agent.status === 'active' ? 'bg-green-500' : agent.status === 'error' ? 'bg-red-500' : 'bg-gray-400'}`}></div>
              <span className="text-sm font-medium">{getStatusText(agent.status)}</span>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* 基本信息 */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3 flex items-center">
              <Bot className="w-5 h-5 mr-2" />
              {t('agents.modal.detail.basicInfo', '基本信息')}
            </h3>
            <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 space-y-3">
              <div>
                <label className="text-sm font-medium text-gray-700 dark:text-gray-200">{t('agents.modal.detail.fields.description', '描述')}</label>
                <p className="text-sm text-gray-600 dark:text-gray-300 mt-1">{agent.description}</p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-200">{t('agents.modal.detail.fields.agentId', 'AI员工ID')}</label>
                  <p className="text-sm text-gray-600 dark:text-gray-300 font-mono mt-1">{agent.id}</p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-200">{t('agents.modal.detail.fields.llmModel', 'LLM模型')}</label>
                  <p className="text-sm text-gray-600 dark:text-gray-300 font-mono mt-1">{agent.llmModel || 'gemini-1.5-pro'}</p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-200">{t('agents.modal.detail.fields.endpoint', '端点地址')}</label>
                  <p className="text-sm text-gray-600 dark:text-gray-300 font-mono mt-1">{agent.endpoint || 'N/A'}</p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-200">{t('agents.modal.detail.fields.lastActive', '最后活跃')}</label>
                  <p className="text-sm text-gray-600 dark:text-gray-300 mt-1">{agent.lastActive || 'N/A'}</p>
                </div>
              </div>
            </div>
          </div>

          {/* 性能统计 */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3 flex items-center">
              <TrendingUp className="w-5 h-5 mr-2" />
              {t('agents.modal.detail.performanceStats', '性能统计')}
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="bg-blue-50 dark:bg-blue-900/30 rounded-lg p-4 text-center">
                <MessageSquare className="w-8 h-8 text-blue-600 dark:text-blue-400 mx-auto mb-2" />
                <div className="text-2xl font-bold text-blue-900 dark:text-blue-200">{agent.conversationCount || 0}</div>
                <div className="text-sm text-blue-600 dark:text-blue-300">{t('agents.modal.detail.metrics.conversationCount', '对话次数')}</div>
              </div>
              <div className="bg-green-50 dark:bg-green-900/30 rounded-lg p-4 text-center">
                <Activity className="w-8 h-8 text-green-600 dark:text-green-400 mx-auto mb-2" />
                <div className="text-2xl font-bold text-green-900 dark:text-green-200">{agent.successRate || 0}%</div>
                <div className="text-sm text-green-600 dark:text-green-300">{t('agents.modal.detail.metrics.successRate', '成功率')}</div>
              </div>
              <div className="bg-purple-50 dark:bg-purple-900/30 rounded-lg p-4 text-center">
                <Clock className="w-8 h-8 text-purple-600 dark:text-purple-400 mx-auto mb-2" />
                <div className="text-2xl font-bold text-purple-900 dark:text-purple-200">{agent.responseTime || '0s'}</div>
                <div className="text-sm text-purple-600 dark:text-purple-300">{t('agents.modal.detail.metrics.avgResponseTime', '平均响应时间')}</div>
              </div>
            </div>
          </div>

          {/* MCP工具 */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3 flex items-center">
              <Wrench className="w-5 h-5 mr-2" />
              {t('agents.modal.detail.mcpTools', '关联MCP工具')}
            </h3>
            <AgentToolsList
              tools={agent.tools || []}
              onToolClick={onToolClick}
              size="md"
              showLabel={false}
              className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4"
            />
            {/* Tags based on tool IDs (fallback to marketplace mapping) */}
            {agent.mcpTools && agent.mcpTools.length > 0 && (
              <div className="mt-3">
                <MCPToolsTagList
                  toolIds={agent.mcpTools}
                  toolsMap={toolsMap}
                  size="xs"
                  showLabel={true}
                />
              </div>
            )}
          </div>

          {/* 知识库 */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3 flex items-center">
              <FolderOpen className="w-5 h-5 mr-2" />
              {t('agents.modal.detail.knowledgeBases', '关联知识库')}
            </h3>
            <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
              {agent.knowledgeBases && agent.knowledgeBases.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {agent.knowledgeBases.map((kbId) => (
                    <span
                      key={kbId}
                      className="inline-flex items-center px-3 py-1 rounded-md bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300 text-sm"
                    >
                      <FolderOpen className="w-4 h-4 mr-1" />
                      {t('agents.modal.detail.knowledgeBasePrefix', '知识库 {{id}}', { id: kbId })}
                    </span>
                  ))}
                </div>
              ) : (
                <div className="flex items-center text-gray-500 dark:text-gray-400">
                  <FolderOpen className="w-4 h-4 mr-2 opacity-50" />
                  <span className="text-sm">{t('agents.modal.detail.noKnowledgeBases', '未关联知识库')}</span>
                </div>
              )}
            </div>
          </div>

          {/* 标签 */}
          {agent.tags && agent.tags.length > 0 && (
            <div>
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3">{t('agents.modal.detail.tags', '标签')}</h3>
              <div className="flex flex-wrap gap-2">
                {agent.tags.map((tag, index) => (
                  <span
                    key={index}
                    className="inline-block px-2 py-1 bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300 text-xs rounded-md"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end space-x-3 p-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            {t('agents.modal.detail.closeButton', '关闭')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default AgentDetailModal;
