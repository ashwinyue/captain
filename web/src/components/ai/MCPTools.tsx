import React, { useState, useMemo, useEffect } from 'react';
import { LuPackage } from 'react-icons/lu';
import { useTranslation } from 'react-i18next';
import MCPSearchBar from './MCPSearchBar';
import MCPToolsList from './MCPToolsList';
import AddMCPToolModal from './AddMCPToolModal';
import EditMCPToolModal from './EditMCPToolModal';
import { MCPToastProvider, useToast } from './MCPToastProvider';
import { useProjectToolsStore } from '@/stores/projectToolsStore';
import { transformAiToolResponseList, searchProjectTools } from '@/utils/projectToolsTransform';
import { ProjectToolsGridSkeleton, ProjectToolsErrorState, ProjectToolsEmptyState } from '@/components/ui/ProjectToolsSkeleton';
import type { MCPTool, AiToolResponse } from '@/types';

/**
 * MCP Tools page component content
 * Updated to use NEW /v1/ai/tools API
 */
const MCPToolsContent: React.FC = () => {
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [showAddMCPToolModal, setShowAddMCPToolModal] = useState<boolean>(false);
  const [showEditMCPToolModal, setShowEditMCPToolModal] = useState<boolean>(false);
  const [selectedToolForEdit, setSelectedToolForEdit] = useState<AiToolResponse | null>(null);
  const { showToast } = useToast();
  const { t } = useTranslation();

  // Project Tools store (now using AI tools)
  const {
    aiTools,
    isLoading,
    error,
    loadMcpTools,
    deleteTool,
    clearError,
  } = useProjectToolsStore();

  // Load MCP tools on component mount (only active tools by default)
  useEffect(() => {
    loadMcpTools(false); // false = don't include deleted tools
  }, [loadMcpTools]);

  // Transform API AI tools to component format
  const mcpTools = useMemo(() => {
    return transformAiToolResponseList(aiTools);
  }, [aiTools]);

  // Filter tools by search query (client-side for better UX)
  const filteredTools = useMemo(() => {
    let filtered = mcpTools;

    // Apply search filter
    if (searchQuery.trim()) {
      filtered = searchProjectTools(filtered, searchQuery);
    }

    return filtered;
  }, [mcpTools, searchQuery]);

  const handleToolAction = async (actionType: string, tool: MCPTool): Promise<void> => {
    try {
      switch (actionType) {
        case 'delete':
        case 'uninstall':
          // Delete AI tool (soft delete)
          await deleteTool(tool.id);
          showToast('success', t('mcp.tools.delete.successTitle', '删除成功'), t('mcp.tools.delete.successMessage', { name: tool.name, defaultValue: `工具 "${tool.name}" 已被删除` }));
          // Reload tools list to reflect the deletion
          await loadMcpTools(false);
          break;
        case 'edit':
          // Find the original AiToolResponse for editing
          const aiTool = aiTools.find(t => t.id === tool.id);
          if (aiTool) {
            setSelectedToolForEdit(aiTool);
            setShowEditMCPToolModal(true);
          } else {
            showToast('error', t('mcp.tools.edit.notFound', '工具未找到'), t('mcp.tools.edit.notFoundMessage', '无法找到该工具的详细信息'));
          }
          break;
        default:
          showToast('warning', t('mcp.tools.unknownAction.title', '未知操作'), t('mcp.tools.unknownAction.message', { type: actionType, defaultValue: `不支持的操作类型: ${actionType}` }));
      }
    } catch (error) {
      console.error(`Tool action ${actionType} failed:`, error);
      const errorMessage = error instanceof Error ? error.message : t('mcp.tools.actionFailed.title', '操作失败');
      showToast('error', t('mcp.tools.actionFailed.title', '操作失败'), errorMessage);
    }
  };

  // Handle retry on error
  const handleRetry = () => {
    clearError();
    loadMcpTools(false);
  };



  const handleSearchChange = (query: string): void => {
    setSearchQuery(query);
  };

  return (
    <main className="flex-grow flex flex-col bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
      {/* Header */}
      <header className="px-6 py-4 border-b border-gray-200/80 dark:border-gray-700 flex justify-between items-center bg-white/60 dark:bg-gray-800/60 backdrop-blur-lg sticky top-0 z-10">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">{t('navigation.mcpTools', 'MCP 工具')}</h2>
        <div className="flex items-center space-x-3">
          <button
            className="flex items-center px-3 py-1.5 bg-green-600 dark:bg-green-700 text-white text-sm rounded-md hover:bg-green-700 dark:hover:bg-green-800 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-1 transition-colors duration-200"
            onClick={() => setShowAddMCPToolModal(true)}
          >
            <LuPackage className="w-4 h-4 mr-1" />
            <span>{t('mcp.tools.addMcpTool', '添加 MCP 工具')}</span>
          </button>
        </div>
      </header>

      {/* Content Area */}
      <div className="flex-grow overflow-y-auto p-6" style={{ height: 0 }}>
        {/* Search Bar */}
        <MCPSearchBar
          value={searchQuery}
          onChange={handleSearchChange}
          className="mb-6"
        />

        {/* Tools List with Loading/Error States */}
        {error ? (
          <ProjectToolsErrorState
            error={error}
            onRetry={handleRetry}
          />
        ) : isLoading ? (
          <ProjectToolsGridSkeleton count={6} />
        ) : filteredTools.length === 0 ? (
          <ProjectToolsEmptyState
            title={t('mcp.tools.installedEmpty.title', '暂无已安装的工具')}
            description={t('mcp.tools.installedEmpty.description', '点击右上角"添加 MCP 工具"按钮创建您的第一个工具')}
          />
        ) : (
          <MCPToolsList
            tools={filteredTools}
            selectedCategory="all"
            onToolAction={handleToolAction}
            onShowToast={showToast}
          />
        )}

      </div>

      {/* 添加 MCP 工具弹窗 */}
      <AddMCPToolModal
        isOpen={showAddMCPToolModal}
        onClose={() => setShowAddMCPToolModal(false)}
      />

      {/* 编辑 MCP 工具弹窗 */}
      <EditMCPToolModal
        isOpen={showEditMCPToolModal}
        onClose={() => {
          setShowEditMCPToolModal(false);
          setSelectedToolForEdit(null);
        }}
        tool={selectedToolForEdit}
      />
    </main>
  );
};

/**
 * MCP Tools page component with Toast provider
 */
const MCPTools: React.FC = () => {
  return (
    <MCPToastProvider>
      <MCPToolsContent />
    </MCPToastProvider>
  );
};

export default MCPTools;
