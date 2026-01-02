import React, { useState, useEffect, useMemo } from 'react';
import { X, Save, RotateCcw, Bot, Wrench, FolderOpen, XCircle, User, Briefcase, Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
// import { getIconComponent, getIconColor } from '@/components/knowledge/IconPicker';
import { useAIStore } from '@/stores';
import { useKnowledgeStore } from '@/stores';

import { useToast } from '@/hooks/useToast';
// import { transformToolSummaryList } from '@/utils/mcpToolsTransform';
import { transformAiToolResponseList } from '@/utils/projectToolsTransform';
import { TransformUtils } from '@/utils/base/BaseTransform';

import ModalFooter from '@/components/ui/ModalFooter';
// import { generateDefaultAvatar } from '@/utils/avatarUtils';
import SectionCard from '@/components/ui/SectionCard';
import SectionHeader from '@/components/ui/SectionHeader';
import { AIAgentsApiService, AIAgentsTransformUtils } from '@/services/aiAgentsApi';
import MCPToolSelectionModal from './MCPToolSelectionModal';
import KnowledgeBaseSelectionModal from './KnowledgeBaseSelectionModal';
import type { Agent, MCPTool, AgentToolResponse, KnowledgeBaseItem, AgentToolDetailed, AgentToolUnion, ToolSummary } from '@/types';
import AgentToolsSection from '@/components/ui/AgentToolsSection';
import AgentKnowledgeBasesSection from '@/components/ui/AgentKnowledgeBasesSection';
import { useAgentForm } from '@/hooks/useAgentForm';
import { useProjectToolsStore } from '@/stores/projectToolsStore';
import { useProvidersStore } from '@/stores/providersStore';
import AIProvidersApiService from '@/services/aiProvidersApi';

interface EditAgentModalProps {
  agentId: string | null;
  isOpen: boolean;
  onClose: () => void;
}

/**
 * AI员工编辑模态框组件
 */
const EditAgentModal: React.FC<EditAgentModalProps> = ({
  agentId,
  isOpen,
  onClose
}) => {
  const { updateAgent, refreshAgents } = useAIStore();
  const { knowledgeBases, fetchKnowledgeBases } = useKnowledgeStore();
  const { aiTools, loadMcpTools } = useProjectToolsStore();
  const { showToast } = useToast();
  const { t } = useTranslation();

  // Providers + model options consistent with Settings → Model Providers
  const { providers, loadProviders } = useProvidersStore();
  const enabledProviderKeys = useMemo(() => {
    const enabled = (providers || []).filter((p) => p.enabled);
    return new Set(enabled.map((p) => AIProvidersApiService.kindToProviderKey(p.kind)));
  }, [providers]);

  const [llmOptions, setLlmOptions] = useState<Array<{ value: string; label: string }>>([]);
  const [llmLoading, setLlmLoading] = useState(false);
  const [llmError, setLlmError] = useState<string | null>(null);


  useEffect(() => {
    if (!isOpen) return;
    if ((providers || []).length === 0) {
      loadProviders().catch(() => {});
    }
  }, [isOpen, providers?.length, loadProviders]);

  useEffect(() => {
    if (!isOpen) return;
    let cancelled = false;
    const fetchChatOptions = async () => {
      if (enabledProviderKeys.size === 0) return;
      setLlmLoading(true);
      setLlmError(null);
      try {
        const svc = new AIProvidersApiService();
        const res = await svc.listProviders({ is_active: true, model_type: 'chat', limit: 100, offset: 0 });
        const options = (res.data || [])
          .filter((p: any) => enabledProviderKeys.has(p.provider) && Array.isArray(p.available_models) && p.available_models.length > 0)
          .flatMap((p: any) => (p.available_models || []).map((m: string) => {
            const ui = `${p.id}:${m}`;
            return { value: ui, label: `${m} · ${p.name || p.provider}` };
          }));
        if (!cancelled) {
          setLlmOptions(options);
        }
      } catch (e: any) {
        if (!cancelled) setLlmError(e?.message || t('agents.create.models.error', '加载模型失败'));
      } finally {
        if (!cancelled) setLlmLoading(false);
      }
    };
    fetchChatOptions();
    return () => { cancelled = true; };
  }, [isOpen, enabledProviderKeys, t]);

  // Agent data state
  const [agent, setAgent] = useState<Agent | null>(null);
  const [isLoadingAgent, setIsLoadingAgent] = useState(false);
  const [agentError, setAgentError] = useState<string | null>(null);

  // 表单状态（通过通用 hook 管理）
  const {
    formData,
    setFormData,
    handleInputChange,
    removeTool,
    removeKnowledgeBase,
    reset,
  } = useAgentForm();

  const [isUpdating, setIsUpdating] = useState(false);
  const [showToolSelectionModal, setShowToolSelectionModal] = useState(false);
  const [showKnowledgeBaseSelectionModal, setShowKnowledgeBaseSelectionModal] = useState(false);

  // Fetch agent data from API
  const fetchAgent = async (id: string): Promise<void> => {
    setIsLoadingAgent(true);
    setAgentError(null);

    try {
      const apiResponse = await AIAgentsApiService.getAgent(id);
      const agentData = AIAgentsTransformUtils.transformApiAgentToAgent(apiResponse);
      setAgent(agentData);
    } catch (error) {
      console.error('Failed to fetch agent:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to load agent details';
      setAgentError(errorMessage);
      showToast('error', t('common.loadFailed', '加载失败'), t('agents.edit.loadAgentError', '无法加载AI员工详情，请稍后重试'));
    } finally {
      setIsLoadingAgent(false);
    }
  };

  // Fetch agent when modal opens or agentId changes
  useEffect(() => {
    if (isOpen && agentId) {
      fetchAgent(agentId);
    } else if (!isOpen) {
      // Reset state when modal closes
      setAgent(null);
      setAgentError(null);
    }
  }, [isOpen, agentId]);

  // 不再在编辑弹窗内主动拉取工具列表，直接使用 agent.tools 进行展示

  // Ensure knowledge bases are loaded when modal opens (for display of selected items)
  useEffect(() => {
    if (isOpen && knowledgeBases.length === 0) {
      try {
        fetchKnowledgeBases();
      } catch (e) {
        console.warn('Failed to load knowledge bases for EditAgentModal:', e);
      }
    }
  }, [isOpen]);

  // 初始化表单数据（直接基于 agent.tools，无需依赖工具商店列表）
  useEffect(() => {
    if (agent && !isLoadingAgent) {
      // 直接使用 Agent 返回的工具详情
      const toolIds: string[] = (agent.tools || []).map(t => t.id);

      // Handle knowledge bases - use collections if available, fallback to knowledgeBases
      const kbIds = agent.collections?.map(collection => collection.id) || agent.knowledgeBases || [];

      reset({
        name: agent.name,
        profession: agent.role || t('agents.copy.defaultProfession', '专家'),
        description: agent.description,
        llmModel: agent.llmModel || 'gemini-1.5-pro',
        mcpTools: toolIds,
        mcpToolConfigs: agent.mcpToolConfigs || {},
        knowledgeBases: kbIds,
      });
    }
  }, [agent, knowledgeBases, isLoadingAgent]);



  // Ensure AI tools are available for mapping when saving
  useEffect(() => {
    if (aiTools.length === 0) {
      // Load AI tools list (only active MCP tools)
      loadMcpTools(false).catch(() => {});
    }
  }, [aiTools.length, loadMcpTools]);



  // 将 AgentToolResponse 转为用于展示的 MCPTool 结构（最小化依赖）
  const agentToolToMCPTool = (t: AgentToolUnion): MCPTool => {
    // Branch 1: legacy AgentToolResponse shape with tool_name
    if ((t as AgentToolResponse) && typeof (t as AgentToolResponse).tool_name === 'string') {
      const tt = t as AgentToolResponse;
      const namePart = tt.tool_name.includes(':') ? tt.tool_name.split(':').slice(1).join(':') : (tt.tool_name || 'tool');
      const provider = tt.tool_name.includes(':') ? tt.tool_name.split(':')[0] : 'mcp';
      return {
        id: tt.id,
        name: namePart,
        title: namePart,
        description: '',
        category: 'integration',
        status: tt.enabled ? 'active' : 'inactive',
        version: 'v1.0.0',
        author: provider,
        lastUpdated: new Date(tt.updated_at || Date.now()).toISOString().split('T')[0],
        usageCount: 0,
        rating: 0,
        tags: [],
        config: tt.config || undefined,
        short_no: provider,
      } as MCPTool;
    }

    // Branch 2: detailed tool object from new API
    const tool = (t as AgentToolDetailed) || {};
    const statusMapped = TransformUtils.transformToolStatus((tool.status || 'ACTIVE') as any);
    const categoryMapped = TransformUtils.transformCategory(tool.category || 'integration');
    const shortNo = tool.mcp_server?.short_no;
    const author = shortNo || tool.mcp_server?.name || tool.tool_source_type || 'mcp';

    return {
      id: tool.id,
      name: tool.title || tool.name || 'tool',
      title: tool.title || tool.name,
      description: tool.description || '',
      category: categoryMapped,
      status: statusMapped,
      version: tool.version || 'v1.0.0',
      author,
      lastUpdated: (tool.updated_at || tool.created_at || new Date().toISOString()).split('T')[0],
      usageCount: 0,
      rating: 0,
      tags: Array.isArray(tool.tags) ? tool.tags : [],
      config: tool.meta_data || undefined,
      input_schema: tool.input_schema,
      short_no: shortNo,
    } as MCPTool;
  };

  // 已添加的MCP工具列表 - 结合 agent.tools 和 AI tools 数据
  const addedMCPTools = useMemo(() => {
    const byId = new Map<string, AgentToolUnion>();
    ((agent?.tools as AgentToolUnion[]) || []).forEach((t) => byId.set(t.id, t));

    const toolsFromAgent = formData.mcpTools
      .map(id => byId.get(id))
      .filter(Boolean) as AgentToolUnion[];

    // 对于不在 agent.tools 中的工具ID，从 AI tools 获取完整信息
    const missingIds = formData.mcpTools.filter(id => !byId.has(id));
    const mcpToolsFromStore = transformAiToolResponseList(aiTools);
    const toolsFromStore = mcpToolsFromStore.filter((tool: MCPTool) => missingIds.includes(tool.id));

    // 对于既不在 agent.tools 也不在工具商店中的工具ID，创建占位符
    const foundStoreIds = toolsFromStore.map(t => t.id);
    const stillMissingIds = missingIds.filter(id => !foundStoreIds.includes(id));
    const placeholderTools: MCPTool[] = stillMissingIds.map(id => ({
      id,
      name: id,
      description: t('agents.edit.tools.pendingNewTool', '待保存的新工具'),
      category: 'integration',
      status: 'active',
      version: 'v1.0.0',
      author: 'mcp',
      lastUpdated: new Date().toISOString().split('T')[0],
      usageCount: 0,
      rating: 0,
      tags: [],
    } as MCPTool));

    return [
      ...toolsFromAgent.map(agentToolToMCPTool),
      ...toolsFromStore,
      ...placeholderTools,
    ];
  }, [agent?.tools, formData.mcpTools, aiTools]);

  // 知识库启用状态：后端暂不支持单独启用/禁用，选中即关联

  // 已添加的知识库列表 - 显示所有已添加的知识库（包括启用和禁用的）
  const addedKnowledgeBases = useMemo(() => {
    return (knowledgeBases as KnowledgeBaseItem[]).filter((kb: KnowledgeBaseItem) => formData.knowledgeBases.includes(kb.id));
  }, [knowledgeBases, formData.knowledgeBases]);

  // 输入处理由 hook 提供

  // 处理工具移除
  const handleToolRemove = (toolId: string) => { removeTool(toolId); };

  // 处理知识库移除
  const handleKnowledgeBaseRemove = (kbId: string) => { removeKnowledgeBase(kbId); };

  // 知识库图标颜色由通用组件处理



  const resolvedLlmModel = useMemo(() => {
    const raw = formData.llmModel;
    if (!raw) return '';
    if (raw.includes(':')) return raw;
    const match = llmOptions.find(option => option.value.split(':').slice(1).join(':') === raw);
    return match ? match.value : '';
  }, [formData.llmModel, llmOptions]);


  const handleSubmit = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    if (!agent || !agentId) return;

    setIsUpdating(true);
    try {
      // 合并可用工具来源：AI tools + 当前已添加/选择的工具（包含 short_no）
      const extraSummaries: ToolSummary[] = addedMCPTools.map((t) => ({
        id: t.id,
        name: t.name,
        title: t.title || t.name,
        description: t.description || null,
        version: t.version || '1.0.0',
        category: typeof t.category === 'string' ? t.category : (t.category as any) || null,
        tags: t.tags || [],
        status: 'ACTIVE',
        tool_source_type: 'MCP_SERVER',
        execution_count: null,
        created_at: new Date().toISOString(),
        mcp_server_id: null,
        input_schema: t.input_schema || {},
        output_schema: null,
        short_no: t.short_no || null,
        is_installed: undefined,
      }));

      // Convert AI tools to ToolSummary format for compatibility
      const aiToolSummaries: ToolSummary[] = aiTools.map((aiTool) => ({
        id: aiTool.id,
        name: aiTool.name,
        title: aiTool.name,
        description: aiTool.description || null,
        version: '1.0.0',
        category: null,
        tags: [],
        status: 'ACTIVE',
        tool_source_type: 'MCP_SERVER',
        execution_count: null,
        created_at: aiTool.created_at,
        mcp_server_id: null,
        input_schema: {},
        output_schema: null,
        short_no: null,
        is_installed: undefined,
      }));

      const byId = new Map<string, ToolSummary>();
      aiToolSummaries.forEach(ts => byId.set(ts.id, ts));
      extraSummaries.forEach(ts => byId.set(ts.id, ts));
      const mergedAvailable = Array.from(byId.values());

      // Preflight: ensure every selected tool has name
      const missing = formData.mcpTools.filter(id => {
        const s = byId.get(id);
        return !s || !s.name;
      });
      if (missing.length > 0) {
        // Try to resolve readable names from addedMCPTools
        const nameMap = new Map(addedMCPTools.map(t => [t.id, t.title || t.name || t.id] as const));
        const missingNames = missing.map(id => nameMap.get(id) || id).join('、');
        showToast('error', t('agents.edit.tools.missingTitle', '工具信息缺失'), t('agents.edit.tools.missingDesc', '以下工具缺少名称：{{names}}，请重新选择后重试', { names: missingNames }));
        setIsUpdating(false);
        return;
      }

      // Validate selected model and keep UI value (providerId:modelName)
      const uiModel = formData.llmModel;
      let normalized = uiModel;
      if (!uiModel || !uiModel.includes(':')) {
        // Try to resolve by matching model name to current options
        const match = llmOptions.find(o => o.value.split(':').slice(1).join(':') === uiModel);
        if (match) {
          normalized = match.value;
        } else {
          showToast('error', t('agents.create.models.selectPlaceholder', '请选择模型'), t('agents.edit.modelProviderRequired', '请重新选择一个有效的模型（需要包含提供商）'));
          setIsUpdating(false);
          return;
        }
      }

      await updateAgent(agentId, {
        name: formData.name,
        description: formData.description,
        llmModel: normalized,
        role: formData.profession,
        mcpTools: formData.mcpTools,
        mcpToolConfigs: formData.mcpToolConfigs,
        knowledgeBases: formData.knowledgeBases,
      }, mergedAvailable);
      // 强制刷新列表，确保卡片立即展示最新 tools/collections
      await refreshAgents();
      showToast('success', t('agents.messages.updateSuccess', '更新成功'), t('agents.messages.updateSuccessDesc', 'AI员工 "{{name}}" 已成功更新', { name: formData.name }));
      onClose();
    } catch (error) {
      console.error('Failed to update agent:', error);
      const errorMessage = error instanceof Error ? error.message : t('agents.edit.updateFailedUnknown', '更新AI员工时发生未知错误');
      showToast('error', t('agents.messages.updateFailed', '更新失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleReset = (): void => {
    if (agent && !isLoadingAgent) {
      // Directly reset based on current agent.tools
      const toolIds: string[] = (agent.tools || []).map((t: any) => t.id);

      // Handle knowledge bases - use collections if available, fallback to knowledgeBases
      const kbIds = agent.collections?.map(collection => collection.id) || agent.knowledgeBases || [];

      reset({
        name: agent.name,
        profession: agent.role || t('agents.copy.defaultProfession', '专家'),
        description: agent.description,
        llmModel: agent.llmModel || 'gemini-1.5-pro',
        mcpTools: toolIds,
        mcpToolConfigs: agent.mcpToolConfigs || {},
        knowledgeBases: kbIds,
      });
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col min-h-0">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center space-x-3">
              <Bot className="w-6 h-6 text-blue-600 dark:text-blue-400" />
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">{t('agents.modal.edit.title', '编辑AI员工')}</h2>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              disabled={isUpdating || isLoadingAgent}
            >
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* Loading State */}
          {isLoadingAgent && (
            <div className="flex items-center justify-center py-12">
              <div className="flex items-center space-x-3">
                <Loader2 className="w-6 h-6 animate-spin text-blue-600 dark:text-blue-400" />
                <span className="text-gray-600 dark:text-gray-300">{t('agents.edit.loading', '加载AI员工详情...')}</span>
              </div>
            </div>
          )}

          {/* Error State */}
          {agentError && !isLoadingAgent && (
            <div className="flex flex-col items-center justify-center py-12 px-6">
              <div className="text-red-500 dark:text-red-400 mb-4">
                <XCircle className="w-12 h-12" />
              </div>
              <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-2">{t('common.loadFailed', '加载失败')}</h3>
              <p className="text-gray-600 dark:text-gray-300 text-center mb-4">{agentError}</p>
              <button
                onClick={() => agentId && fetchAgent(agentId)}
                className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
              >
                {t('common.retry', '重试')}
              </button>
            </div>
          )}

          {/* Form Content - Only show when agent is loaded */}
          {agent && !isLoadingAgent && !agentError && (
            <form onSubmit={handleSubmit} className="flex-1 flex flex-col min-h-0">
              <div className="flex-1 overflow-y-auto min-h-0 dark:bg-gray-900">
                <div className="p-6 space-y-6">
          {/* 基本信息 */}
          <SectionCard variant="blue">
            <SectionHeader icon={<Bot className="w-5 h-5 text-blue-600" />} title={t('agents.detail.basicInfo', '基本信息')} />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* AI员工名称 */}
              <div>
                <label className="flex items-center space-x-2 text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
                  <User className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                  <span>{t('agents.form.name', 'AI员工名称')}</span>
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all duration-200 hover:border-gray-400 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 dark:hover:border-gray-500"
                  placeholder={t('agents.create.placeholders.name', '请输入AI员工名称')}
                  required
                  disabled={isUpdating}
                />
              </div>

              {/* 职业/角色 */}
              <div>
                <label className="flex items-center space-x-2 text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
                  <Briefcase className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                  <span>{t('agents.form.profession', '职业/角色')}</span>
                </label>
                <input
                  type="text"
                  value={formData.profession}
                  onChange={(e) => handleInputChange('profession', e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all duration-200 hover:border-gray-400 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 dark:hover:border-gray-500"
                  placeholder={t('agents.create.placeholders.profession', '例如：客服专员、技术支持、销售顾问')}
                  required
                  disabled={isUpdating}
                />
              </div>
            </div>
          </SectionCard>

          {/* 模型配置 */}
          <SectionCard variant="purple">
            <SectionHeader icon={<Bot className="w-5 h-5 text-purple-600" />} title={t('agents.create.sections.modelConfig', '模型配置')} />

            <div>
              <label className="flex items-center space-x-2 text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
                <span>{t('agents.form.llmModel', 'LLM模型')}</span>
              </label>
              <select
                value={resolvedLlmModel}
                onChange={(e) => handleInputChange('llmModel', e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 transition-all duration-200 hover:border-gray-400 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 dark:hover:border-gray-500"
                disabled={isUpdating || llmLoading}
              >
                {llmLoading ? (
                  <option value="">{t('agents.create.models.loading', '正在加载模型...')}</option>
                ) : llmError ? (
                  <option value="">{t('agents.create.models.error', '加载模型失败')}</option>
                ) : llmOptions.length === 0 ? (
                  <option value="">{t('agents.create.models.empty', '暂无可用模型')}</option>
                ) : (
                  <>
                    {!resolvedLlmModel && (
                      <option value="">{t('agents.create.models.selectPlaceholder', '请选择模型')}</option>
                    )}
                    {/* Fallback option in case current value is a providerKey:model not present in options */}
                    {formData.llmModel && !llmOptions.some(o => o.value === formData.llmModel) && formData.llmModel.includes(':') && (
                      <option value={formData.llmModel}>
                        {formData.llmModel.split(':').slice(1).join(':')} · {formData.llmModel.split(':')[0]}
                      </option>
                    )}
                    {llmOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </>
                )}
              </select>
              {llmError && (
                <p className="mt-1 text-sm text-red-500 dark:text-red-400 flex items-center space-x-1">
                  <XCircle className="w-4 h-4" />
                  <span>{t('agents.create.models.loadFailedInline', '模型加载失败: {{error}}', { error: llmError })}</span>
                </p>
              )}
            </div>
          </SectionCard>


          {/* AI员工描述 */}
          <SectionCard variant="green">
            <SectionHeader icon={<FolderOpen className="w-5 h-5 text-green-600" />} title={t('agents.create.sections.description', 'AI员工描述')} />

            <div>
              <label className="flex items-center space-x-2 text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
                <span>{t('agents.form.detailedDescription', '详细描述')}</span>
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={4}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500 transition-all duration-200 resize-none hover:border-gray-400 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 dark:hover:border-gray-500"
                placeholder={t('agents.create.placeholders.description', '请详细描述AI员工的功能、职责和特点，例如：负责处理客户咨询，提供产品信息和技术支持...')}
                required
                disabled={isUpdating}
              />
            </div>
          </SectionCard>

          {/* MCP工具选择 */}
          <SectionCard variant="orange">
            <SectionHeader icon={<Wrench className="w-5 h-5 text-orange-600" />} title={t('agents.create.sections.mcpTools', 'MCP工具')} />

            <AgentToolsSection
              tools={addedMCPTools}
              toolConfigs={formData.mcpToolConfigs}
              onAdd={() => setShowToolSelectionModal(true)}
              onRemove={handleToolRemove}
              disabled={isUpdating}
            />
          </SectionCard>

          {/* 知识库选择 */}
          <SectionCard variant="teal">
            <SectionHeader icon={<FolderOpen className="w-5 h-5 text-teal-600" />} title={t('agents.form.knowledgeBases', '知识库')} />

            <AgentKnowledgeBasesSection
              items={addedKnowledgeBases}
              onAdd={() => setShowKnowledgeBaseSelectionModal(true)}
              onRemove={handleKnowledgeBaseRemove}
              disabled={isUpdating}
            />
          </SectionCard>

                </div>
              </div>

          {/* Footer */}
          <ModalFooter>
            <button
              type="button"
              onClick={onClose}
              className="inline-flex h-9 items-center px-4 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 transition-colors"
              disabled={isUpdating}
            >
              {t('common.cancel', '取消')}
            </button>
            <button
              type="button"
              onClick={handleReset}
              className="inline-flex h-9 items-center px-4 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 transition-colors"
              disabled={isUpdating}
            >
              <RotateCcw className="w-4 h-4 mr-2" />
              {t('common.reset', '重置')}
            </button>
            <button
              type="submit"
              className="inline-flex h-9 items-center px-5 text-sm font-medium bg-blue-600 dark:bg-blue-700 text-white border border-transparent rounded-md hover:bg-blue-700 dark:hover:bg-blue-800 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={isUpdating}
            >
              {isUpdating ? (
                <>
                  <div className="w-4 h-4 mr-2 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                  {t('common.saving', '保存中...')}
                </>
              ) : (
                <>
                  <Save className="w-4 h-4 mr-2" />
                  {t('common.save', '保存')}
                </>
              )}
            </button>
          </ModalFooter>
            </form>
          )}
        </div>
      </div>

      {/* Modals - Always available regardless of loading state */}
      <MCPToolSelectionModal
        isOpen={showToolSelectionModal}
        onClose={() => setShowToolSelectionModal(false)}
        selectedTools={formData.mcpTools}
        toolConfigs={formData.mcpToolConfigs}
        onConfirm={(selectedToolIds, toolConfigs) => {
          setFormData({
            mcpTools: selectedToolIds,
            mcpToolConfigs: toolConfigs,
          });
        }}
      />

      <KnowledgeBaseSelectionModal
        isOpen={showKnowledgeBaseSelectionModal}
        onClose={() => setShowKnowledgeBaseSelectionModal(false)}
        selectedKnowledgeBases={formData.knowledgeBases}
        onConfirm={(selectedKnowledgeBaseIds) => {
          setFormData({ knowledgeBases: selectedKnowledgeBaseIds });
        }}
      />
    </>
  );
};

export default EditAgentModal;
