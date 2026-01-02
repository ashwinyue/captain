import React, { useState, useMemo, useEffect, useRef } from 'react';
import { X, Save, RotateCcw, Bot, Wrench, FolderOpen, XCircle, User, Briefcase } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import ModalFooter from '@/components/ui/ModalFooter';
// import { getIconComponent, getIconColor } from '@/components/knowledge/IconPicker';

import { useAIStore } from '@/stores';
import { useKnowledgeStore } from '@/stores';
import { useProjectToolsStore } from '@/stores/projectToolsStore';

import { useToast } from '@/hooks/useToast';
import { transformAiToolResponseList } from '@/utils/projectToolsTransform';
// import { generateDefaultAvatar } from '@/utils/avatarUtils';
import SectionCard from '@/components/ui/SectionCard';
import SectionHeader from '@/components/ui/SectionHeader';

import MCPToolSelectionModal from './MCPToolSelectionModal';
import KnowledgeBaseSelectionModal from './KnowledgeBaseSelectionModal';
import type { MCPTool } from '@/types';
import AgentToolsSection from '@/components/ui/AgentToolsSection';
import AgentKnowledgeBasesSection from '@/components/ui/AgentKnowledgeBasesSection';
import { useAgentForm } from '@/hooks/useAgentForm';
import { useProvidersStore } from '@/stores/providersStore';
import AIProvidersApiService from '@/services/aiProvidersApi';
import ProjectConfigApiService from '@/services/projectConfigApi';
import { useAuthStore } from '@/stores/authStore';


/**
 * AI员工创建模态框组件
 */
const CreateAgentModal: React.FC = () => {
  const {
    showCreateAgentModal,
    createAgentFormData,
    createAgentErrors,
    isCreatingAgent,

    setShowCreateAgentModal,
    setCreateAgentFormData,
    resetCreateAgentForm,
    validateCreateAgentForm,
    createAgent,
    refreshAgents,
    setAgentsError
  } = useAIStore();

  const { knowledgeBases } = useKnowledgeStore();
  const { aiTools } = useProjectToolsStore();
  const { t } = useTranslation();
  const { showToast } = useToast();

  // Refs for auto-scrolling to the first invalid field on validation failure
  const nameInputRef = useRef<HTMLInputElement | null>(null);
  const professionInputRef = useRef<HTMLInputElement | null>(null);
  const llmSelectRef = useRef<HTMLSelectElement | null>(null);
  const descriptionTextareaRef = useRef<HTMLTextAreaElement | null>(null);
  // Providers store and model options (aligned with Settings → Model Providers)
  const { providers, loadProviders } = useProvidersStore();
  const enabledProviderKeys = useMemo(() => {
    const enabled = (providers || []).filter((p) => p.enabled);
    return new Set(enabled.map((p) => AIProvidersApiService.kindToProviderKey(p.kind)));
  }, [providers]);

  const [llmOptions, setLlmOptions] = useState<Array<{ value: string; label: string }>>([]);
  const [llmLoading, setLlmLoading] = useState(false);
  const [llmError, setLlmError] = useState<string | null>(null);


  const projectId = useAuthStore((s) => s.user?.project_id);

  // Ensure providers are loaded when create modal is open
  useEffect(() => {
    if (!showCreateAgentModal) return;
    if ((providers || []).length === 0) {
      loadProviders().catch(() => {});
    }
  }, [showCreateAgentModal, providers?.length, loadProviders]);

  // Fetch chat models from /v1/ai/providers with model_type=chat and filter by enabled providers
  useEffect(() => {
    if (!showCreateAgentModal) return;
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
  }, [showCreateAgentModal, enabledProviderKeys, t]);

  // Preselect default model from project AI config if available (only when create modal is open)
  useEffect(() => {
    if (!showCreateAgentModal) return;
    let cancelled = false;
    (async () => {
      if (!projectId) return;
      try {
        const svc = new ProjectConfigApiService();
        const conf = await svc.getAIConfig(projectId);
        if (cancelled) return;
        const uiValue = conf.default_chat_provider_id && conf.default_chat_model ? `${conf.default_chat_provider_id}:${conf.default_chat_model}` : '';
        if (uiValue && !createAgentFormData.llmModel) {
          // Only set when form has no value yet; ensure the option exists or still set for persistence
          setCreateAgentFormData({ llmModel: uiValue });
        }
      } catch (_) {
        // ignore
      }
    })();
    return () => { cancelled = true; };
  }, [showCreateAgentModal, projectId, createAgentFormData.llmModel, setCreateAgentFormData]);


  // Tool selection modal state
  const [showToolSelectionModal, setShowToolSelectionModal] = useState(false);

  // Knowledge base selection modal state
  const [showKnowledgeBaseSelectionModal, setShowKnowledgeBaseSelectionModal] = useState(false);

  // Shared form logic (controlled by store formData)
  const {
    handleInputChange,
    removeTool,
    removeKnowledgeBase,
  } = useAgentForm({
    controlledFormData: createAgentFormData,
    onFormDataChange: setCreateAgentFormData,
  });







  // 已添加的MCP工具列表 - 显示所有已添加的工具（包括启用和禁用的）
  const addedMCPTools = useMemo(() => {
    // Transform AI tools (from NEW /v1/ai/tools API) to MCPTool format
    const mcpTools = transformAiToolResponseList(aiTools);
    return mcpTools.filter((tool: MCPTool) => createAgentFormData.mcpTools.includes(tool.id));
  }, [aiTools, createAgentFormData.mcpTools]);

  // 知识库启用状态：后端暂不支持单独启用/禁用，选中即关联

  // 已添加的知识库列表 - 显示所有已添加的知识库（包括启用和禁用的）
  const addedKnowledgeBases = useMemo(() => {
    return knowledgeBases.filter(kb => createAgentFormData.knowledgeBases.includes(kb.id));
  }, [knowledgeBases, createAgentFormData.knowledgeBases]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const isValid = validateCreateAgentForm();
    if (!isValid) {
      // Global feedback
      showToast(
        'error',
        t('agents.messages.validationFailed', '验证失败'),
        t('agents.messages.fillRequired', '请填写所有必填字段')
      );

      // Scroll to the first invalid field and focus it
      const scrollAndFocus = (el: HTMLElement | null) => {
        if (!el) return;
        el.scrollIntoView({ behavior: 'smooth', block: 'center' });
        el.focus({ preventScroll: true });
      };

      const trimmedName = createAgentFormData.name.trim();
      const trimmedProfession = createAgentFormData.profession.trim();
      const trimmedDescription = createAgentFormData.description.trim();
      const hasModel = !!createAgentFormData.llmModel;

      if (!trimmedName) {
        scrollAndFocus(nameInputRef.current);
      } else if (!trimmedProfession) {
        scrollAndFocus(professionInputRef.current);
      } else if (!hasModel) {
        scrollAndFocus(llmSelectRef.current);
      } else if (!trimmedDescription) {
        scrollAndFocus(descriptionTextareaRef.current);
      }

      return;
    }

    try {
      // Clear any previous errors
      setAgentsError(null);

      // Preflight: ensure every selected tool has name
      const byId = new Map(aiTools.map(ts => [ts.id, ts] as const));
      const missing = createAgentFormData.mcpTools.filter(id => {
        const s = byId.get(id);
        return !s || !s.name;
      });
      if (missing.length > 0) {
        const names = missing.join('、');
        showToast(
          'error',
          t('agents.create.tools.missingTitle', '工具信息缺失'),
          t('agents.create.tools.missingDesc', '以下工具缺少名称：{{names}}，请重新选择后重试', { names })
        );
        return;
      }

      // Validate selected model and keep UI value (providerId:modelName)
      const uiModel = createAgentFormData.llmModel;
      if (!uiModel || !uiModel.includes(':')) {
        showToast(
          'error',
          t('agents.create.models.selectPlaceholder', '请选择模型'),
          t('agents.create.models.invalid', '请选择一个有效的模型（需包含提供商）')
        );
        // Ensure the user sees the model field
        if (llmSelectRef.current) {
          llmSelectRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
          llmSelectRef.current.focus({ preventScroll: true });
        }
        return;
      }
      // Pass UI model value directly; transform will extract ai_provider_id and pure model name
      // Convert aiTools to ToolSummary format for compatibility with createAgent
      const toolSummaries = aiTools.map(aiTool => ({
        id: aiTool.id,
        name: aiTool.name,
        title: aiTool.name,
        description: aiTool.description || null,
        version: '1.0.0',
        category: null,
        tags: [],
        status: 'ACTIVE' as const,
        tool_source_type: 'MCP_SERVER' as const,
        execution_count: null,
        created_at: aiTool.created_at,
        mcp_server_id: null,
        input_schema: {},
        output_schema: null,
        short_no: null,
        is_installed: undefined,
      }));
      await createAgent({ ...createAgentFormData, llmModel: uiModel }, toolSummaries);
      // Refresh list to ensure the just-created agent shows tools/collections immediately
      await refreshAgents();

      // Show success toast
      showToast(
        'success',
        t('agents.messages.createSuccess', '创建成功'),
        t('agents.messages.createSuccessDesc', 'AI员工 "{name}" 已成功创建', { name: createAgentFormData.name })
      );
    } catch (error) {
      console.error('创建AI员工失败:', error);
      const errorMessage = error instanceof Error
        ? error.message
        : t('agents.messages.createFailed', '创建AI员工时发生未知错误');
      showToast('error', t('agents.messages.createFailed', '创建失败'), errorMessage);
    }
  };

  const handleClose = () => {
    setShowCreateAgentModal(false);
    resetCreateAgentForm();
  };

  // 处理工具移除
  const handleToolRemove = (toolId: string) => {
    removeTool(toolId);
  };

  // 处理知识库移除
  const handleKnowledgeBaseRemove = (kbId: string) => {
    removeKnowledgeBase(kbId);
  };

  // 知识库图标颜色由通用组件处理

  if (!showCreateAgentModal) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col min-h-0">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-3">
            <Bot className="w-6 h-6 text-blue-600 dark:text-blue-400" />
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-100">{t('agents.modal.create.title', '创建AI员工')}</h2>
          </div>
          <button
            onClick={handleClose}
            className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            disabled={isCreatingAgent}
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Content */}
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
                    <span>{t('agents.form.name', 'AI员工名称')} <span className="text-red-500 dark:text-red-400">*</span></span>
                  </label>
                  <input
                    ref={nameInputRef}
                    type="text"
                    value={createAgentFormData.name}
                    onChange={(e) => handleInputChange('name', e.target.value)}
                    className={`w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all duration-200 dark:bg-gray-700 dark:text-gray-100 ${
                      createAgentErrors.name ? 'border-red-500 bg-red-50 dark:bg-red-950 dark:border-red-700' : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
                    }`}
                    placeholder={t('agents.create.placeholders.name', '请输入AI员工名称')}
                    disabled={isCreatingAgent}
                  />
                  {createAgentErrors.name && (
                    <p className="mt-1 text-sm text-red-500 dark:text-red-400 flex items-center space-x-1">
                      <XCircle className="w-4 h-4" />
                      <span>{createAgentErrors.name}</span>
                    </p>
                  )}
                </div>

                {/* 职业/角色 */}
                <div>
                  <label className="flex items-center space-x-2 text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
                    <Briefcase className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                    <span>{t('agents.form.profession', '职业/角色')} <span className="text-red-500 dark:text-red-400">*</span></span>
                  </label>
                  <input
                    ref={professionInputRef}
                    type="text"
                    value={createAgentFormData.profession}
                    onChange={(e) => handleInputChange('profession', e.target.value)}
                    className={`w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all duration-200 dark:bg-gray-700 dark:text-gray-100 ${
                      createAgentErrors.profession ? 'border-red-500 bg-red-50 dark:bg-red-950 dark:border-red-700' : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
                    }`}
                    placeholder={t('agents.create.placeholders.profession', '例如：客服专员、技术支持、销售顾问')}
                    disabled={isCreatingAgent}
                  />
                  {createAgentErrors.profession && (
                    <p className="mt-1 text-sm text-red-500 dark:text-red-400 flex items-center space-x-1">
                      <XCircle className="w-4 h-4" />
                      <span>{createAgentErrors.profession}</span>
                    </p>
                  )}
                </div>
              </div>
            </SectionCard>

            {/* 模型配置 */}
            <SectionCard variant="purple">
              <SectionHeader icon={<Bot className="w-5 h-5 text-purple-600" />} title={t('agents.create.sections.modelConfig', '模型配置')} />

              <div>
                <label className="flex items-center space-x-2 text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
                  <span>{t('agents.form.llmModel', 'LLM模型')} <span className="text-red-500 dark:text-red-400">*</span></span>
                </label>
                <select
                  ref={llmSelectRef}
                  value={createAgentFormData.llmModel}
                  onChange={(e) => handleInputChange('llmModel', e.target.value)}
                  className={`w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 transition-all duration-200 dark:bg-gray-700 dark:text-gray-100 ${
                    createAgentErrors.llmModel ? 'border-red-500 bg-red-50 dark:bg-red-950 dark:border-red-700' : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
                  }`}
                  disabled={isCreatingAgent || llmLoading}
                >
                  {llmLoading ? (
                    <option value="">{t('agents.create.models.loading', '正在加载模型...')}</option>
                  ) : llmError ? (
                    <option value="">{t('agents.create.models.error', '加载模型失败')}</option>
                  ) : llmOptions.length === 0 ? (
                    <option value="">{t('agents.create.models.empty', '暂无可用模型')}</option>
                  ) : (
                    <>
                      {!createAgentFormData.llmModel && (
                        <option value="">{t('agents.create.models.selectPlaceholder', '请选择模型')}</option>
                      )}
                      {llmOptions.map(option => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </>
                  )}
                </select>
                {createAgentErrors.llmModel && (
                  <p className="mt-1 text-sm text-red-500 dark:text-red-400 flex items-center space-x-1">
                    <XCircle className="w-4 h-4" />
                    <span>{createAgentErrors.llmModel}</span>
                  </p>
                )}
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
                  <span>{t('agents.form.detailedDescription', '详细描述')} <span className="text-red-500 dark:text-red-400">*</span></span>
                </label>
                <textarea
                  ref={descriptionTextareaRef}
                  value={createAgentFormData.description}
                  onChange={(e) => handleInputChange('description', e.target.value)}
                  rows={4}
                  className={`w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500 transition-all duration-200 resize-none dark:bg-gray-700 dark:text-gray-100 ${
                    createAgentErrors.description ? 'border-red-500 bg-red-50 dark:bg-red-950 dark:border-red-700' : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
                  }`}
                  placeholder={t('agents.create.placeholders.description', '请详细描述AI员工的功能、职责和特点，例如：负责处理客户咨询，提供产品信息和技术支持...')}
                  disabled={isCreatingAgent}
                />
                {createAgentErrors.description && (
                  <p className="mt-1 text-sm text-red-500 dark:text-red-400 flex items-center space-x-1">
                    <XCircle className="w-4 h-4" />
                    <span>{createAgentErrors.description}</span>
                  </p>
                )}
              </div>
            </SectionCard>

            {/* MCP工具选择 */}
            <SectionCard variant="orange">
              <SectionHeader icon={<Wrench className="w-5 h-5 text-orange-600" />} title={t('agents.create.sections.mcpTools', 'MCP工具')} />

              <AgentToolsSection
                tools={addedMCPTools}
                toolConfigs={createAgentFormData.mcpToolConfigs}
                onAdd={() => setShowToolSelectionModal(true)}
                onRemove={handleToolRemove}
                disabled={isCreatingAgent}
              />
            </SectionCard>

            {/* 知识库选择 */}
            <SectionCard variant="teal">
              <SectionHeader icon={<FolderOpen className="w-5 h-5 text-teal-600" />} title={t('agents.form.knowledgeBases', '知识库')} />

              <AgentKnowledgeBasesSection
                items={addedKnowledgeBases}
                onAdd={() => setShowKnowledgeBaseSelectionModal(true)}
                onRemove={handleKnowledgeBaseRemove}
                disabled={isCreatingAgent}
              />
            </SectionCard>

            </div>
          </div>

          {/* Footer */}
          <ModalFooter>
              <button
                type="button"
                onClick={handleClose}
                className="inline-flex h-9 items-center px-4 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 transition-colors"
                disabled={isCreatingAgent}
              >
                {t('common.cancel', '取消')}
              </button>
              <button
                type="button"
                onClick={resetCreateAgentForm}
                className="inline-flex h-9 items-center px-4 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 transition-colors"
                disabled={isCreatingAgent}
              >
                <RotateCcw className="w-4 h-4 mr-2" />
                {t('common.reset', '重置')}
              </button>
              <button
                type="submit"
                className="inline-flex h-9 items-center px-5 text-sm font-medium bg-blue-600 dark:bg-blue-700 text-white border border-transparent rounded-md hover:bg-blue-700 dark:hover:bg-blue-800 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={isCreatingAgent}
              >
                {isCreatingAgent ? (
                  <>
                    <div className="w-4 h-4 mr-2 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                    {t('common.creating', '创建中...')}
                  </>
                ) : (
                  <>
                    <Save className="w-4 h-4 mr-2" />
                    {t('agents.actions.create', '创建AI员工')}
                  </>
                )}
              </button>
          </ModalFooter>
        </form>
      </div>

      {/* MCP Tool Selection Modal */}
      <MCPToolSelectionModal
        isOpen={showToolSelectionModal}
        onClose={() => setShowToolSelectionModal(false)}
        selectedTools={createAgentFormData.mcpTools}
        toolConfigs={createAgentFormData.mcpToolConfigs}
        onConfirm={(selectedToolIds, toolConfigs) => {
          setCreateAgentFormData({
            mcpTools: selectedToolIds,
            mcpToolConfigs: toolConfigs
          });
          setShowToolSelectionModal(false);
        }}
      />


      {/* Knowledge Base Selection Modal */}
      <KnowledgeBaseSelectionModal
        isOpen={showKnowledgeBaseSelectionModal}
        onClose={() => setShowKnowledgeBaseSelectionModal(false)}
        selectedKnowledgeBases={createAgentFormData.knowledgeBases}
        onConfirm={(selectedKnowledgeBaseIds) => {
          setCreateAgentFormData({ knowledgeBases: selectedKnowledgeBaseIds });
        }}
      />
    </div>
  );
};

export default CreateAgentModal;
