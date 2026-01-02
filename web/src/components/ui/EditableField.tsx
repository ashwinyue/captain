import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import { Check, X, Edit3 } from 'lucide-react';

interface EditableFieldProps {
  label: string;
  value: string;
  onSave: (newValue: string) => Promise<void> | void;
  onCancel?: () => void;
  placeholder?: string;
  type?: 'text' | 'email' | 'tel';
  validate?: (value: string) => string | null;
  className?: string;
  disabled?: boolean;
}

/**
 * Inline editable field component
 */
const EditableField: React.FC<EditableFieldProps> = ({
  label,
  value,
  onSave,
  onCancel,
  placeholder,
  type = 'text',
  validate,
  className = '',
  disabled = false
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  useEffect(() => {
    setEditValue(value);
  }, [value]);

  const handleEdit = () => {
    if (disabled) return;
    setIsEditing(true);
    setError(null);
  };

  const handleSave = async () => {
    if (validate) {
      const validationError = validate(editValue);
      if (validationError) {
        setError(validationError);
        return;
      }
    }

    setIsLoading(true);
    setError(null);

    try {
      await onSave(editValue);
      setIsEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('common.saveFailed', '保存失败'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    setEditValue(value);
    setError(null);
    setIsEditing(false);
    onCancel?.();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSave();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      handleCancel();
    }
  };

  const handleBlur = () => {
    // Auto-save on blur if value changed
    if (editValue !== value && !error) {
      handleSave();
    } else {
      handleCancel();
    }
  };

  if (isEditing) {
    return (
      <div className={`flex justify-between items-start ${className}`}>
        <span className="text-gray-500 dark:text-gray-400 flex-shrink-0 text-[13px] leading-5 pt-1.5">{label}:</span>
        <div className="flex-1 ml-2 min-w-0">
          <div className="flex items-center space-x-1.5">
            <input
              ref={inputRef}
              type={type}
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onKeyDown={handleKeyDown}
              onBlur={handleBlur}
              placeholder={placeholder}
              disabled={isLoading}
              className={`flex-1 min-w-0 max-w-[140px] px-2.5 py-1.5 text-[13px] leading-5 text-gray-800 dark:text-gray-200 bg-white dark:bg-gray-700 border rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-blue-500 dark:focus:border-blue-400 transition-colors ${
                error ? 'border-red-300 dark:border-red-600 focus:ring-red-500 focus:border-red-500' : 'border-gray-300 dark:border-gray-600'
              } ${isLoading ? 'opacity-50' : ''}`}
            />
            <button
              onClick={handleSave}
              disabled={isLoading}
              className="p-1.5 text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 hover:bg-green-50 dark:hover:bg-green-900/30 rounded-md transition-colors disabled:opacity-50 flex-shrink-0"
              title={t('common.save', '保存')}
            >
              <Check className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={handleCancel}
              disabled={isLoading}
              className="p-1.5 text-gray-500 dark:text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-md transition-colors disabled:opacity-50 flex-shrink-0"
              title={t('common.cancel', '取消')}
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
          {error && (
            <div className="mt-1.5 text-[11px] leading-4 text-red-600 dark:text-red-400 font-medium">{error}</div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={`flex justify-between items-start group ${className}`}>
      <span className="text-gray-500 dark:text-gray-400 text-[13px] leading-5 flex-shrink-0 pt-0.5">{label}:</span>
      <div className="flex items-start space-x-1.5 flex-1 min-w-0 ml-2">
        <span
          className={`text-gray-800 dark:text-gray-200 font-medium text-[13px] leading-5 flex-1 min-w-0 text-right ${disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer hover:text-gray-900 dark:hover:text-gray-100'} line-clamp-2`}
          onClick={handleEdit}
          title={value || placeholder || '-'}
          style={{
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            wordBreak: 'break-all'
          }}
        >
          {value || placeholder || '-'}
        </span>
        {!disabled && (
          <button
            onClick={handleEdit}
            className="opacity-0 group-hover:opacity-100 p-1 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 rounded transition-all flex-shrink-0 mt-0.5"
            title={t('common.edit', '编辑')}
          >
            <Edit3 className="w-3 h-3" />
          </button>
        )}
      </div>
    </div>
  );
};

export default EditableField;
