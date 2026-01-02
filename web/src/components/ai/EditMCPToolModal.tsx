import React, { useState, useEffect } from 'react';
import { X, Edit, AlertCircle } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useToast } from './MCPToastProvider';
import { useProjectToolsStore } from '@/stores/projectToolsStore';
import type { AiToolResponse, AiToolUpdateRequest } from '@/types';

interface EditMCPToolModalProps {
  isOpen: boolean;
  onClose: () => void;
  tool: AiToolResponse | null;
}

interface FormData {
  name: string;
  description: string;
  transport_type: string;
  endpoint: string;
}

interface FormErrors {
  name?: string;
  description?: string;
  transport_type?: string;
  endpoint?: string;
}

const EditMCPToolModal: React.FC<EditMCPToolModalProps> = ({ isOpen, onClose, tool }) => {
  const { t } = useTranslation();
  const { showToast } = useToast();
  const { updateMcpTool, loadMcpTools } = useProjectToolsStore();

  const [formData, setFormData] = useState<FormData>({
    name: '',
    description: '',
    transport_type: 'http',
    endpoint: '',
  });

  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Populate form when tool changes
  useEffect(() => {
    if (tool) {
      setFormData({
        name: tool.name || '',
        description: tool.description || '',
        transport_type: tool.transport_type || 'http',
        endpoint: tool.endpoint || '',
      });
      setErrors({});
    }
  }, [tool]);

  const handleInputChange = (field: keyof FormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = t('mcp.editMcpToolModal.errors.nameRequired', '请输入工具名称');
    } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.name.trim())) {
      newErrors.name = t('mcp.editMcpToolModal.errors.invalidName', '工具名称只能包含英文字母、数字、下划线和连字符');
    }

    if (!formData.endpoint.trim()) {
      newErrors.endpoint = t('mcp.editMcpToolModal.errors.endpointRequired', '请输入端点地址');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm() || !tool) {
      return;
    }

    setIsSubmitting(true);

    try {
      const updateData: AiToolUpdateRequest = {
        name: formData.name.trim(),
        description: formData.description.trim() || null,
        transport_type: formData.transport_type.trim() || null,
        endpoint: formData.endpoint.trim() || null,
      };

      await updateMcpTool(tool.id, updateData);

      showToast(
        'success',
        t('mcp.editMcpToolModal.success.title', '更新成功'),
        t('mcp.editMcpToolModal.success.message', '{{name}} 已成功更新', { name: formData.name })
      );

      // Refresh tool list
      await loadMcpTools(false);

      onClose();
    } catch (error) {
      console.error('Failed to update MCP tool:', error);
      showToast(
        'error',
        t('mcp.editMcpToolModal.error.title', '更新失败'),
        error instanceof Error ? error.message : t('mcp.editMcpToolModal.error.message', '更新 MCP 工具失败，请稍后重试')
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!isSubmitting) {
      setErrors({});
      onClose();
    }
  };

  if (!isOpen || !tool) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-3">
            <Edit className="w-6 h-6 text-green-600 dark:text-green-400" />
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-100">
              {t('mcp.editMcpToolModal.title', '编辑 MCP 工具')}
            </h2>
          </div>
          <button
            onClick={handleClose}
            disabled={isSubmitting}
            className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors disabled:opacity-50"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-6 dark:bg-gray-900">
          {/* Tool Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
              {t('mcp.editMcpToolModal.fields.name', '工具名称')}
              <span className="text-red-500 dark:text-red-400 ml-1">*</span>
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 dark:bg-gray-700 dark:text-gray-100 ${
                errors.name ? 'border-red-500 dark:border-red-600' : 'border-gray-300 dark:border-gray-600'
              }`}
              placeholder={t('mcp.editMcpToolModal.placeholders.name', '例如：weather-api')}
              disabled={isSubmitting}
            />
            {errors.name && (
              <div className="mt-1 flex items-center text-sm text-red-600 dark:text-red-400">
                <AlertCircle className="w-4 h-4 mr-1" />
                {errors.name}
              </div>
            )}
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
              {t('mcp.editMcpToolModal.fields.description', '工具描述')}
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 dark:bg-gray-700 dark:text-gray-100"
              placeholder={t('mcp.editMcpToolModal.placeholders.description', '简要描述该工具的功能')}
              disabled={isSubmitting}
            />
          </div>

          {/* Transport Type */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
              {t('mcp.editMcpToolModal.fields.transportType', '传输类型')}
            </label>
            <select
              value={formData.transport_type}
              onChange={(e) => handleInputChange('transport_type', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 dark:bg-gray-700 dark:text-gray-100"
              disabled={isSubmitting}
            >
              <option value="http">HTTP</option>
              <option value="sse">SSE</option>
            </select>
          </div>

          {/* Endpoint */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
              {t('mcp.editMcpToolModal.fields.endpoint', '端点地址')}
              <span className="text-red-500 dark:text-red-400 ml-1">*</span>
            </label>
            <input
              type="text"
              value={formData.endpoint}
              onChange={(e) => handleInputChange('endpoint', e.target.value)}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 dark:bg-gray-700 dark:text-gray-100 ${
                errors.endpoint ? 'border-red-500 dark:border-red-600' : 'border-gray-300 dark:border-gray-600'
              }`}
              placeholder={t('mcp.editMcpToolModal.placeholders.endpoint', '例如：https://api.example.com/mcp')}
              disabled={isSubmitting}
            />
            {errors.endpoint && (
              <div className="mt-1 flex items-center text-sm text-red-600 dark:text-red-400">
                <AlertCircle className="w-4 h-4 mr-1" />
                {errors.endpoint}
              </div>
            )}
          </div>

          {/* Tool Type - Read Only */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
              {t('mcp.editMcpToolModal.fields.toolType', '工具类型')}
            </label>
            <input
              type="text"
              value={tool.tool_type}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 cursor-not-allowed"
              disabled
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {t('mcp.editMcpToolModal.hints.toolTypeReadonly', '工具类型不可修改')}
            </p>
          </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end space-x-3 p-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
          <button
            type="button"
            onClick={handleClose}
            disabled={isSubmitting}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50"
          >
            {t('common.cancel', '取消')}
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={isSubmitting}
            className="px-4 py-2 text-sm font-medium text-white bg-green-600 dark:bg-green-700 border border-transparent rounded-md hover:bg-green-700 dark:hover:bg-green-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? t('common.submitting', '提交中...') : t('common.save', '保存')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default EditMCPToolModal;


