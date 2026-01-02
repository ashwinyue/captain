import React from 'react';
import EditableField from '../ui/EditableField';
import CustomAttributeManager from '../ui/CustomAttributeManager';
import type { VisitorBasicInfo } from '@/data/mockVisitor';
import { useTranslation } from 'react-i18next';
import { formatOnlineDuration } from '@/utils/dateUtils';

interface BasicInfoSectionProps {
  basicInfo: VisitorBasicInfo;
  isOnline?: boolean;
  onUpdateBasicInfo: (
    field: 'name' | 'nickname' | 'email' | 'phone' | 'note',
    value: string
  ) => void;
  onAddCustomAttribute: (key: string, value: string) => void;
  onUpdateCustomAttribute: (id: string, key: string, value: string) => void;
  onDeleteCustomAttribute: (id: string) => void;
  className?: string;
}



/**
 * 基本信息模块组件
 */
const BasicInfoSection: React.FC<BasicInfoSectionProps> = ({
  basicInfo,
  isOnline = false,
  onUpdateBasicInfo,
  onAddCustomAttribute,
  onUpdateCustomAttribute,
  onDeleteCustomAttribute,
  className = ''
}) => {
  const { t } = useTranslation();
  const validateEmail = (email: string): string | null => {
    if (!email.trim()) return null;
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email) ? null : t('visitor.validation.invalidEmail', '\u8bf7\u8f93\u5165\u6709\u6548\u7684\u90ae\u7bb1\u5730\u5740');
  };
  const validatePhone = (phone: string): string | null => {
    if (!phone.trim()) return null;
    const phoneRegex = /^[\d\s\-\+\(\)]{7,}$/;
    return phoneRegex.test(phone) ? null : t('visitor.validation.invalidPhone', '\u8bf7\u8f93\u5165\u6709\u6548\u7684\u7535\u8bdd\u53f7\u7801');
  };
  return (
    <div className={className}>
      <h4 className="text-xs font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-3">{t('visitor.sections.basicInfo', '基本信息')}</h4>
      <div className="space-y-1.5 text-[13px] leading-5">
        <EditableField
          label={t('visitor.fields.name', '姓名')}
          value={basicInfo.name}
          onSave={(value) => onUpdateBasicInfo('name', value)}
          placeholder="-"
        />
        <EditableField
          label={t('visitor.fields.nickname', '昵称')}
          value={basicInfo.nickname || ''}
          onSave={(value) => onUpdateBasicInfo('nickname', value)}
          placeholder="-"
        />
        <EditableField
          label={t('visitor.fields.email', '邮箱')}
          value={basicInfo.email}
          onSave={(value) => onUpdateBasicInfo('email', value)}
          placeholder="-"
          type="email"
          validate={validateEmail}
        />
        <EditableField
          label={t('visitor.fields.phone', '电话')}
          value={basicInfo.phone}
          onSave={(value) => onUpdateBasicInfo('phone', value)}
          placeholder="-"
          type="tel"
          validate={validatePhone}
        />
        <div className="flex justify-between items-center py-1">
          <span className="text-gray-500 dark:text-gray-400">{t('visitor.table.lastOnline', '最后在线')}</span>
          <span className="text-gray-800 dark:text-gray-200 font-medium">
            {formatOnlineDuration(basicInfo.lastOnlineDurationMinutes, isOnline)}
          </span>
        </div>
        <EditableField
          label={t('visitor.fields.note', '备注')}
          value={basicInfo.note || ''}
          onSave={(value) => onUpdateBasicInfo('note', value)}
          placeholder="-"
        />

        {/* 自定义属性 */}
        <CustomAttributeManager
          attributes={basicInfo.customAttributes || []}
          onAdd={onAddCustomAttribute}
          onUpdate={onUpdateCustomAttribute}
          onDelete={onDeleteCustomAttribute}
          className="pt-2 space-y-1.5"
        />
      </div>
    </div>
  );
};

export default BasicInfoSection;
