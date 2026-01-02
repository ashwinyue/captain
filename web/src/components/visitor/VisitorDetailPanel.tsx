import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import i18n from '@/i18n';
import { User, AlertCircle, Loader2, X } from 'lucide-react';
import TagManager from '../ui/TagManager';
import VisitorHeader from './VisitorHeader';
import BasicInfoSection from './BasicInfoSection';
import ImageCropModal from '../ui/ImageCropModal';

import AIInsightsSection from './AIInsightsSection';
import SystemInfoSection from './SystemInfoSection';
import RecentActivitySection from './RecentActivitySection';
import { visitorApiService, type VisitorAttributesUpdateRequest, type VisitorResponse } from '@/services/visitorApi';
import { tagsApiService } from '@/services/tagsApi';
import { useChannelStore } from '@/stores/channelStore';
import { useChatStore } from '@/stores/chatStore';
import { getChannelKey } from '@/utils/channelUtils';
import { useToast } from '@/hooks/useToast';
import type { ChannelVisitorExtra } from '@/types';
import { PlatformType } from '@/types';
import { toPlatformType } from '@/utils/platformUtils';
import { buildLastSeenText } from '@/utils/dateUtils';
import type { ExtendedVisitor, CustomAttribute, VisitorTag } from '@/data/mockVisitor';

export interface VisitorDetailPanelProps {
  /** 访客ID (用于独立模式) */
  visitorId?: string;
  /** 频道ID (用于聊天模式) */
  channelId?: string | null;
  /** 频道类型 (用于聊天模式) */
  channelType?: number | null;
  /** 直接传入的访客数据 (可选，用于预填充) */
  visitorData?: VisitorResponse | null;
  /** 是否显示关闭按钮 */
  showCloseButton?: boolean;
  /** 关闭回调 */
  onClose?: () => void;
  /** 访客数据更新后的回调 (用于独立模式，通知父组件刷新数据) */
  onVisitorUpdated?: (visitorId: string) => void;
  /** 面板样式变体 */
  variant?: 'sidebar' | 'drawer';
  /** 自定义类名 */
  className?: string;
}

/**
 * 将频道信息中的访客扩展数据转换为面板使用的扩展结构
 */
const toExtendedVisitorFromChannel = (extra: ChannelVisitorExtra): ExtendedVisitor => {
  const tags = Array.isArray(extra.tags)
    ? extra.tags.map((t, idx) => ({
        id: t.id || `tag_${idx}`,
        name: t.display_name || '',
        display_name: t.display_name || '',
        color: t.color || 'gray',
        weight: typeof t.weight === 'number' ? t.weight : 0
      }))
    : [];

  const customAttrsObj = (extra.custom_attributes || {}) as Record<string, unknown>;
  const customAttributes = Object.entries(customAttrsObj).map(([key, value], index) => ({
    id: `custom_${index}`,
    key,
    value: String(value ?? ''),
    editable: true,
  }));

  return {
    id: extra.id,
    name: extra.name || extra.nickname || i18n.t('visitor.unknown', '未知访客'),
    avatar: extra.avatar_url || '',
    status: extra.is_online ? 'online' : 'offline',
    platform: extra.source || 'website',
    firstVisit: extra.first_visit_time || '',
    visitCount: 1,
    tags,
    basicInfo: {
      name: extra.name || '',
      email: extra.email || '',
      phone: extra.phone_number || '',
      nickname: extra.display_nickname || '',
      company: extra.company || '',
      jobTitle: extra.job_title || '',
      source: extra.source || '',
      note: extra.note || '',
      avatarUrl: extra.avatar_url || '',
      lastOnlineDurationMinutes: (extra as any).last_online_duration_minutes,
      customAttributes,
    },
    aiInsights: {
      satisfaction: 4,
      emotion: { type: 'neutral', icon: 'Meh', label: i18n.t('visitor.ai.emotion.neutral', '中性') },
    },
    systemInfo: {
      firstVisit: extra.first_visit_time || '',
      source: extra.source || i18n.t('visitor.system.defaultSource', '官网'),
      browser: 'Chrome / macOS',
    },
    recentActivity: [],
    relatedTickets: [],
    aiPersonaTags: [],
  };
};

/**
 * 将 VisitorResponse 转换为 ExtendedVisitor
 */
const toExtendedVisitorFromResponse = (response: VisitorResponse): ExtendedVisitor => {
  const tags = Array.isArray(response.tags)
    ? response.tags.map((t, idx) => ({
        id: t.id || `tag_${idx}`,
        name: t.name || '',
        display_name: t.name || '',
        color: t.color || 'gray',
        weight: 0
      }))
    : [];

  const customAttrsObj = (response.custom_attributes || {}) as Record<string, unknown>;
  const customAttributes = Object.entries(customAttrsObj).map(([key, value], index) => ({
    id: `custom_${index}`,
    key,
    value: String(value ?? ''),
    editable: true,
  }));

  return {
    id: response.id,
    name: response.name || response.display_nickname || response.nickname_zh || response.nickname || i18n.t('visitor.unknown', '未知访客'),
    avatar: response.avatar_url || '',
    status: response.is_online ? 'online' : 'offline',
    platform: response.platform_type || 'website',
    firstVisit: response.first_visit_time || '',
    visitCount: 1,
    tags,
    basicInfo: {
      name: response.name || '',
      email: response.email || '',
      phone: response.phone_number || '',
      nickname: response.display_nickname || response.nickname_zh || response.nickname || '',
      company: response.company || '',
      jobTitle: response.job_title || '',
      source: response.source || '',
      note: response.note || '',
      avatarUrl: response.avatar_url || '',
      lastOnlineDurationMinutes: response.last_online_duration_minutes,
      customAttributes,
    },
    aiInsights: {
      satisfaction: 4,
      emotion: { type: 'neutral', icon: 'Meh', label: i18n.t('visitor.ai.emotion.neutral', '中性') },
    },
    systemInfo: {
      firstVisit: response.first_visit_time || '',
      source: response.source || i18n.t('visitor.system.defaultSource', '官网'),
      browser: response.system_info?.browser || '',
    },
    recentActivity: [],
    relatedTickets: [],
    aiPersonaTags: [],
  };
};

/**
 * 公共访客详情面板组件
 * 可用于聊天界面侧边栏和访客管理界面的详情抽屉
 */
const VisitorDetailPanel: React.FC<VisitorDetailPanelProps> = ({
  visitorId,
  channelId,
  channelType,
  visitorData,
  showCloseButton = false,
  onClose,
  onVisitorUpdated,
  variant = 'sidebar',
  className = '',
}) => {
  const [visitor, setVisitor] = useState<ExtendedVisitor | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [updateError, setUpdateError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [hasLoadedInitial, setHasLoadedInitial] = useState(false);
  // Avatar crop modal state
  const [showCropModal, setShowCropModal] = useState(false);
  const [cropImageSrc, setCropImageSrc] = useState<string>('');
  const [cropImageMimeType, setCropImageMimeType] = useState<string>('image/png');
  const { t } = useTranslation();
  const { showToast } = useToast();
  const avatarInputRef = useRef<HTMLInputElement>(null);

  // 是否使用频道模式（从 channelStore 获取数据）
  const useChannelMode = Boolean(channelId && channelType != null);

  const compositeKey = useMemo(() => {
    if (!channelId || channelType == null) return null;
    return getChannelKey(channelId, channelType);
  }, [channelId, channelType]);

  const channelInfo = useChannelStore(state => (compositeKey ? state.channels[compositeKey] : undefined));
  const channelStoreError = useChannelStore(state => (compositeKey ? state.errors[compositeKey] : null));
  const isChannelFetching = useChannelStore(state => (compositeKey ? Boolean(state.inFlight[compositeKey]) : false));
  const ensureChannelInfo = useChannelStore(state => state.ensureChannel);

  // Derive recent activities from channel extra
  const recentActivities = useMemo(() => {
    const extra = channelInfo?.extra as ChannelVisitorExtra | undefined;
    const list = extra?.recent_activities;
    if (!Array.isArray(list)) return [] as NonNullable<ChannelVisitorExtra['recent_activities']>;
    return [...list].sort((a, b) => new Date(b.occurred_at).getTime() - new Date(a.occurred_at).getTime());
  }, [channelInfo]);

  // Derive system info from channel extra
  const systemInfo = useMemo(() => {
    const extra = channelInfo?.extra as ChannelVisitorExtra | undefined;
    return extra?.system_info || null;
  }, [channelInfo]);

  // Derive AI insights from channel extra
  const aiInsights = useMemo(() => {
    const extra = channelInfo?.extra as ChannelVisitorExtra | undefined;
    return extra?.ai_insights ?? null;
  }, [channelInfo]);

  // 频道模式：从 channelStore 获取数据
  useEffect(() => {
    if (!useChannelMode || !channelId) {
      return;
    }

    // Only auto-fetch when we have not fetched yet, nothing is in-flight, and no prior error exists
    if (!channelInfo && !isChannelFetching && !channelStoreError) {
      ensureChannelInfo({ channel_id: channelId, channel_type: channelType ?? 1 }).catch(error => {
        const message = error instanceof Error ? error.message : t('visitor.load.failedDesc', '加载访客数据失败');
        setLoadError(message);
        showToast('error', t('visitor.load.failedTitle', '加载访客信息失败'), message);
      });
    }
  }, [useChannelMode, ensureChannelInfo, isChannelFetching, showToast, channelId, channelType, channelInfo, channelStoreError, t]);

  // 独立模式：从 visitorData 或 API 获取数据
  useEffect(() => {
    if (useChannelMode) return;

    if (visitorData) {
      setVisitor(toExtendedVisitorFromResponse(visitorData));
      setHasLoadedInitial(true);
      setIsLoading(false);
      return;
    }

    if (visitorId) {
      setIsLoading(true);
      setLoadError(null);
      visitorApiService.getVisitor(visitorId)
        .then(response => {
          setVisitor(toExtendedVisitorFromResponse(response));
          setHasLoadedInitial(true);
        })
        .catch(error => {
          const message = error instanceof Error ? error.message : t('visitor.load.failedDesc', '加载访客数据失败');
          setLoadError(message);
        })
        .finally(() => {
          setIsLoading(false);
        });
    }
  }, [useChannelMode, visitorId, visitorData, t]);

  // 频道模式：更新 loading 状态
  useEffect(() => {
    if (!useChannelMode || !channelId) {
      return;
    }

    setIsLoading(!hasLoadedInitial && isChannelFetching && !channelInfo);
  }, [useChannelMode, isChannelFetching, channelId, channelInfo, hasLoadedInitial]);

  // 频道模式：处理错误
  useEffect(() => {
    if (!useChannelMode) return;

    if (channelStoreError) {
      setLoadError(channelStoreError);
    } else {
      setLoadError(null);
    }
  }, [useChannelMode, channelStoreError]);

  // 频道模式：从 channelInfo 转换访客数据
  useEffect(() => {
    if (!useChannelMode) return;

    const extra = channelInfo?.extra as any;
    const isVisitorExtra = extra && typeof extra === 'object' && 'platform_open_id' in extra && !('staff_id' in extra);
    if (isVisitorExtra) {
      setVisitor(toExtendedVisitorFromChannel(extra as ChannelVisitorExtra));
      setHasLoadedInitial(true);
    } else {
      setVisitor(null);
    }
  }, [useChannelMode, channelInfo]);

  // API集成的基本信息更新函数
  const handleUpdateBasicInfo = useCallback(async (
    field: 'name' | 'nickname' | 'email' | 'phone' | 'note',
    value: string
  ) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      const customAttributesForApi = (visitor.basicInfo.customAttributes || []).reduce((acc: Record<string, string | null>, attr: CustomAttribute) => {
        acc[attr.key] = attr.value;
        return acc;
      }, {} as Record<string, string | null>);

      const apiKey = field === 'phone' ? 'phone_number' : field;

      const updateData: VisitorAttributesUpdateRequest = {
        [apiKey]: value,
        custom_attributes: customAttributesForApi
      } as VisitorAttributesUpdateRequest;

      await visitorApiService.updateVisitorAttributes(visitor.id, updateData);
      if (useChannelMode && channelId) {
        await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
      }

      const localKey = field === 'phone' ? 'phone' : field;

      setVisitor((prev: ExtendedVisitor | null) => prev ? {
        ...prev,
        basicInfo: {
          ...prev.basicInfo,
          [localKey]: value
        }
      } : null);

      const labelMap: Record<string, string> = {
        name: t('visitor.fields.name', '姓名'),
        nickname: t('visitor.fields.nickname', '昵称'),
        email: t('visitor.fields.email', '邮箱'),
        phone: t('visitor.fields.phone', '电话'),
        note: t('visitor.fields.note', '备注')
      };
      showToast('success', t('visitor.update.successTitle', '更新成功'), t('visitor.update.fieldUpdated', '{{field}}已更新', { field: labelMap[field] || t('visitor.fields.field', '字段') }));

      // 通知父组件数据已更新
      if (!useChannelMode && onVisitorUpdated && visitor.id) {
        onVisitorUpdated(visitor.id);
      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.update.failedDesc', '更新失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.update.failedTitle', '更新失败'), errorMessage);
      throw error;
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, useChannelMode, channelId, channelType, t, onVisitorUpdated]);

  // API集成的自定义属性管理函数
  const handleAddCustomAttribute = useCallback(async (key: string, value: string) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      const newAttribute: CustomAttribute = {
        id: Date.now().toString(),
        key,
        value,
        editable: true
      };

      const currentAttributes = visitor.basicInfo.customAttributes || [];
      const updatedAttributes = [...currentAttributes, newAttribute];

      const customAttributesForApi = updatedAttributes.reduce((acc, attr) => {
        acc[attr.key] = attr.value;
        return acc;
      }, {} as Record<string, any>);

      await visitorApiService.updateVisitorAttributes(visitor.id, {
        custom_attributes: customAttributesForApi
      });

      if (useChannelMode && channelId) {
        await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
      }

      setVisitor((prev: ExtendedVisitor | null) => prev ? {
        ...prev,
        basicInfo: {
          ...prev.basicInfo,
          customAttributes: updatedAttributes
        }
      } : null);

      showToast('success', t('visitor.customAttr.addSuccessTitle', '添加成功'), t('visitor.customAttr.addSuccessDesc', '自定义属性已添加'));
      
      // 通知父组件数据已更新
      if (!useChannelMode && onVisitorUpdated && visitor.id) {
        onVisitorUpdated(visitor.id);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.customAttr.addFailedDesc', '添加失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.customAttr.addFailedTitle', '添加失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, useChannelMode, channelId, channelType, t, onVisitorUpdated]);

  const handleUpdateCustomAttribute = useCallback(async (id: string, key: string, value: string) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      const updatedAttributes = visitor.basicInfo.customAttributes?.map((attr: CustomAttribute) =>
        attr.id === id ? { ...attr, key, value } : attr
      ) || [];

      const customAttributesForApi = updatedAttributes.reduce((acc: Record<string, string | null>, attr: CustomAttribute) => {
        acc[attr.key] = attr.value;
        return acc;
      }, {} as Record<string, string | null>);

      await visitorApiService.updateVisitorAttributes(visitor.id, {
        custom_attributes: customAttributesForApi
      });

      setVisitor(prev => prev ? {
        ...prev,
        basicInfo: {
          ...prev.basicInfo,
          customAttributes: updatedAttributes
        }
      } : null);

      showToast('success', t('visitor.customAttr.updateSuccessTitle', '更新成功'), t('visitor.customAttr.updateSuccessDesc', '自定义属性已更新'));
      
      // 通知父组件数据已更新
      if (!useChannelMode && onVisitorUpdated && visitor.id) {
        onVisitorUpdated(visitor.id);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.customAttr.updateFailedDesc', '更新失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.customAttr.updateFailedTitle', '更新失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, t, useChannelMode, onVisitorUpdated]);

  const handleDeleteCustomAttribute = useCallback(async (id: string) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      const updatedAttributes = visitor.basicInfo.customAttributes?.filter(attr => attr.id !== id) || [];

      const customAttributesForApi = updatedAttributes.reduce((acc, attr) => {
        acc[attr.key] = attr.value;
        return acc;
      }, {} as Record<string, any>);

      await visitorApiService.updateVisitorAttributes(visitor.id, {
        custom_attributes: customAttributesForApi
      });

      setVisitor(prev => prev ? {
        ...prev,
        basicInfo: {
          ...prev.basicInfo,
          customAttributes: updatedAttributes
        }
      } : null);

      showToast('success', t('visitor.customAttr.deleteSuccessTitle', '删除成功'), t('visitor.customAttr.deleteSuccessDesc', '自定义属性已删除'));
      
      // 通知父组件数据已更新
      if (!useChannelMode && onVisitorUpdated && visitor.id) {
        onVisitorUpdated(visitor.id);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.customAttr.deleteFailedDesc', '删除失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.customAttr.deleteFailedTitle', '删除失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, t, useChannelMode, onVisitorUpdated]);

  // 获取常用标签
  const fetchCommonTags = useCallback(async () => {
    const res = await tagsApiService.listVisitorTags({ limit: 50 });
    return res.data.map(tag => ({
      id: tag.id,
      display_name: tag.display_name,
      name: tag.name,
      color: tag.color || 'gray',
      weight: tag.weight
    }));
  }, []);

  // 选择已有标签
  const handleAssociateExistingTag = useCallback(async (tagId: string) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);
    try {
      await tagsApiService.createVisitorTag({ visitor_id: visitor.id, tag_id: tagId });
      if (useChannelMode && channelId) {
        await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
      }
      showToast('success', t('visitor.tags.addSuccessTitle', '添加成功'), t('visitor.tags.addSuccessDesc', '标签已添加'));
      
      // 通知父组件数据已更新，并刷新本地数据
      if (!useChannelMode && visitor.id) {
        // 重新获取访客详情以更新标签列表
        const updatedVisitor = await visitorApiService.getVisitor(visitor.id);
        setVisitor(toExtendedVisitorFromResponse(updatedVisitor));
        if (onVisitorUpdated) {
          onVisitorUpdated(visitor.id);
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.tags.addFailedDesc', '添加标签失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.tags.addFailedTitle', '添加失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, useChannelMode, channelId, channelType, t, onVisitorUpdated]);

  // 添加新标签
  const handleAddTag = useCallback(async (tagData: Omit<VisitorTag, 'id'>) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      let tagId: string | null = null;
      try {
        const newTagResponse = await tagsApiService.createTag({
          name: tagData.name,
          category: 'visitor',
          weight: tagData.weight || 0,
          color: tagData.color || null,
          description: null
        });
        tagId = newTagResponse.id;
      } catch (createErr: any) {
        try {
          const existing = await tagsApiService.listVisitorTags({ search: tagData.name, limit: 1 });
          const found = existing.data.find(t => t.name.toLowerCase() === tagData.name.toLowerCase());
          if (found) {
            tagId = found.id;
          } else {
            throw createErr;
          }
        } catch (e) {
          throw createErr;
        }
      }

      if (!tagId) throw new Error(t('visitor.tags.errors.noTagId', '无法获取标签ID'));

      await tagsApiService.createVisitorTag({
        visitor_id: visitor.id,
        tag_id: tagId
      });

      if (useChannelMode && channelId) {
        await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
      }

      showToast('success', t('visitor.tags.addSuccessTitle', '添加成功'), t('visitor.tags.addSuccessDesc', '标签已添加'));
      
      // 通知父组件数据已更新，并刷新本地数据
      if (!useChannelMode && visitor.id) {
        const updatedVisitor = await visitorApiService.getVisitor(visitor.id);
        setVisitor(toExtendedVisitorFromResponse(updatedVisitor));
        if (onVisitorUpdated) {
          onVisitorUpdated(visitor.id);
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.tags.addFailedDesc', '添加标签失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.tags.addFailedTitle', '添加失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, useChannelMode, channelId, channelType, t, onVisitorUpdated]);

  const handleUpdateTag = useCallback(async (id: string, updates: Partial<VisitorTag>) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      const tagToUpdate = visitor.tags.find(tag => tag.id === id);
      if (!tagToUpdate) {
        throw new Error(t('visitor.tags.errors.notFound', '标签不存在'));
      }

      const updateData: { weight?: number; color?: string } = {};
      if (updates.weight !== undefined) {
        updateData.weight = updates.weight;
      }
      if (updates.color !== undefined) {
        updateData.color = updates.color;
      }

      if (Object.keys(updateData).length > 0) {
        await tagsApiService.updateTag(tagToUpdate.id, updateData);
      }

      if (useChannelMode && channelId) {
        await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
      }

      showToast('success', t('visitor.tags.updateSuccessTitle', '更新成功'), t('visitor.tags.updateSuccessDesc', '标签已更新'));
      
      // 通知父组件数据已更新
      if (!useChannelMode && onVisitorUpdated && visitor.id) {
        const updatedVisitor = await visitorApiService.getVisitor(visitor.id);
        setVisitor(toExtendedVisitorFromResponse(updatedVisitor));
        onVisitorUpdated(visitor.id);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.tags.updateFailedDesc', '更新标签失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.tags.updateFailedTitle', '更新失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, useChannelMode, channelId, channelType, t, onVisitorUpdated]);

  const handleRemoveTag = useCallback(async (id: string) => {
    if (!visitor) return;

    setIsUpdating(true);
    setUpdateError(null);

    try {
      await tagsApiService.deleteVisitorTagByVisitorAndTag(visitor.id, id);
      if (useChannelMode && channelId) {
        await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
      }
      showToast('success', t('visitor.tags.removeSuccessTitle', '删除成功'), t('visitor.tags.removeSuccessDesc', '标签已删除'));
      
      // 通知父组件数据已更新，并刷新本地数据
      if (!useChannelMode && visitor.id) {
        const updatedVisitor = await visitorApiService.getVisitor(visitor.id);
        setVisitor(toExtendedVisitorFromResponse(updatedVisitor));
        if (onVisitorUpdated) {
          onVisitorUpdated(visitor.id);
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.tags.removeFailedDesc', '删除标签失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.tags.removeFailedTitle', '删除失败'), errorMessage);
    } finally {
      setIsUpdating(false);
    }
  }, [visitor, showToast, useChannelMode, channelId, channelType, t, onVisitorUpdated]);

  // 头像上传处理
  const handleAvatarClick = useCallback(() => {
    avatarInputRef.current?.click();
  }, []);

  const handleAvatarChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !visitor) return;

    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      showToast('error', t('visitor.avatar.invalidTypeTitle', '文件类型错误'), t('visitor.avatar.invalidTypeDesc', '请选择 JPEG、PNG、GIF 或 WebP 格式的图片'));
      return;
    }

    const maxSize = 5 * 1024 * 1024;
    if (file.size > maxSize) {
      showToast('error', t('visitor.avatar.fileTooLargeTitle', '文件过大'), t('visitor.avatar.fileTooLargeDesc', '图片大小不能超过 5MB'));
      return;
    }

    const imageUrl = URL.createObjectURL(file);
    setCropImageSrc(imageUrl);
    setCropImageMimeType(file.type);
    setShowCropModal(true);

    if (avatarInputRef.current) {
      avatarInputRef.current.value = '';
    }
  }, [visitor, showToast, t]);

  const handleCropCancel = useCallback(() => {
    setShowCropModal(false);
    if (cropImageSrc) {
      URL.revokeObjectURL(cropImageSrc);
    }
    setCropImageSrc('');
  }, [cropImageSrc]);

  const handleCropConfirm = useCallback(async (blob: Blob) => {
    if (!visitor) return;

    setShowCropModal(false);
    if (cropImageSrc) {
      URL.revokeObjectURL(cropImageSrc);
    }
    setCropImageSrc('');

    setIsUploadingAvatar(true);
    setUpdateError(null);

    try {
      const fileName = `avatar_${Date.now()}.${cropImageMimeType.split('/')[1] || 'png'}`;
      const file = new File([blob], fileName, { type: cropImageMimeType });

      const response = await visitorApiService.uploadAvatar(visitor.id, file);
      
      const avatarUrlWithCacheBust = `${response.avatar_url}?t=${Date.now()}`;
      setVisitor((prev: ExtendedVisitor | null) => prev ? {
        ...prev,
        avatar: avatarUrlWithCacheBust,
        basicInfo: {
          ...prev.basicInfo,
          avatarUrl: avatarUrlWithCacheBust
        }
      } : null);

      if (useChannelMode && channelId) {
        const syncedInfo = await useChatStore.getState().syncChannelInfoAcrossUI(channelId, channelType ?? 1);
        if (syncedInfo && syncedInfo.avatar) {
          const avatarWithCacheBust = syncedInfo.avatar.includes('?') 
            ? `${syncedInfo.avatar}&t=${Date.now()}` 
            : `${syncedInfo.avatar}?t=${Date.now()}`;
          useChannelStore.getState().updateChannelAvatar(channelId, channelType ?? 1, avatarWithCacheBust);
          useChatStore.getState().applyChannelInfo(channelId, channelType ?? 1, {
            ...syncedInfo,
            avatar: avatarWithCacheBust
          });
        }
      }

      showToast('success', t('visitor.avatar.uploadSuccessTitle', '上传成功'), t('visitor.avatar.uploadSuccessDesc', '访客头像已更新'));
      
      // 通知父组件数据已更新
      if (!useChannelMode && onVisitorUpdated && visitor.id) {
        onVisitorUpdated(visitor.id);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('visitor.avatar.uploadFailedDesc', '上传失败');
      setUpdateError(errorMessage);
      showToast('error', t('visitor.avatar.uploadFailedTitle', '上传失败'), errorMessage);
    } finally {
      setIsUploadingAvatar(false);
    }
  }, [visitor, useChannelMode, channelId, channelType, cropImageSrc, cropImageMimeType, showToast, t, onVisitorUpdated]);

  // 基于 variant 的样式
  const containerClasses = variant === 'drawer'
    ? `w-full h-full bg-white dark:bg-gray-800 flex flex-col ${className}`
    : `w-72 bg-white/80 dark:bg-gray-800/80 backdrop-blur-lg border-l border-gray-200/60 dark:border-gray-700/60 flex flex-col shrink-0 font-sans antialiased ${className}`;

  const emptyContainerClasses = variant === 'drawer'
    ? `w-full h-full bg-white dark:bg-gray-800 flex items-center justify-center ${className}`
    : `w-72 bg-white/80 dark:bg-gray-800/80 backdrop-blur-lg border-l border-gray-200/60 dark:border-gray-700/60 flex items-center justify-center shrink-0 font-sans antialiased ${className}`;

  // 渲染空状态或错误状态
  if (!visitor) {
    return (
      <div className={emptyContainerClasses}>
        <div className="text-center text-gray-500 dark:text-gray-400 px-4">
          {isLoading ? (
            <>
              <Loader2 size={48} className="mx-auto mb-4 text-gray-300 dark:text-gray-600 animate-spin" />
              <p className="text-sm leading-5">{t('visitor.ui.loading', '正在加载访客信息...')}</p>
            </>
          ) : loadError ? (
            <>
              <AlertCircle size={48} className="mx-auto mb-4 text-red-300 dark:text-red-500" />
              <p className="text-sm leading-5 text-red-600 dark:text-red-400 mb-2">{t('visitor.ui.loadFailed', '加载失败')}</p>
              <p className="text-xs text-gray-400 dark:text-gray-500 mb-4">{loadError}</p>
              <button
                onClick={() => window.location.reload()}
                className="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md transition-colors"
              >
                {t('common.retry', '重试')}
              </button>
            </>
          ) : (
            <>
              <User size={48} className="mx-auto mb-4 text-gray-300 dark:text-gray-600" />
              <p className="text-sm leading-5">{t('visitor.ui.noInfo', '暂无访客信息')}</p>
            </>
          )}
        </div>
      </div>
    );
  }

  // 获取频道额外信息用于显示
  const channelExtra = channelInfo?.extra as ChannelVisitorExtra | undefined;

  return (
    <div className={containerClasses}>
      {/* Panel Header */}
      <div className="p-4 border-b border-gray-200/60 dark:border-gray-700/60 sticky top-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-lg z-10">
        {showCloseButton && onClose && (
          <div className="flex justify-end mb-2">
            <button
              onClick={onClose}
              className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition-colors"
            >
              <X size={20} className="text-gray-500 dark:text-gray-400" />
            </button>
          </div>
        )}
        {/* Hidden file input for avatar upload */}
        <input
          ref={avatarInputRef}
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          className="hidden"
          onChange={handleAvatarChange}
        />
        {/* Visitor Avatar and Name */}
        <VisitorHeader
          name={channelExtra?.display_nickname || visitor.name}
          status={visitor.status || 'offline'}
          avatar={visitor.avatar}
          platformType={(() => {
            const fromExtra: PlatformType | undefined = channelExtra?.platform_type as PlatformType | undefined;
            const fallbackPlatform = visitor.platform || '';
            return fromExtra ?? toPlatformType(fallbackPlatform);
          })()}
          lastSeenText={buildLastSeenText(channelExtra?.last_offline_time, channelExtra?.is_online ?? null) || undefined}
          onAvatarClick={handleAvatarClick}
          isUploading={isUploadingAvatar}
          visitorId={channelId || visitor.id}
        />
      </div>

      {/* Panel Content */}
      <div className="relative flex-grow overflow-y-auto p-4 space-y-6" style={{ height: 0 }}>
        {/* Loading/Error State */}
        {isUpdating && (
          <div className="pointer-events-none absolute top-2 left-1/2 -translate-x-1/2 z-20 flex items-center text-xs text-blue-700 dark:text-blue-400 bg-white/80 dark:bg-gray-800/80 backdrop-blur px-2 py-1 rounded border border-blue-100 dark:border-blue-800 shadow-sm">
            <Loader2 className="w-3 h-3 mr-1 animate-spin" />
            {t('common.updating', '更新中...')}
          </div>
        )}

        {(updateError || loadError) && (
          <div className="flex items-center py-2 px-3 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800/50 rounded-md">
            <AlertCircle className="w-4 h-4 mr-2 flex-shrink-0" />
            <span className="flex-1">{updateError || loadError}</span>
          </div>
        )}

        {/* Basic Info */}
        <BasicInfoSection
          basicInfo={visitor.basicInfo}
          isOnline={visitor.status === 'online'}
          onUpdateBasicInfo={handleUpdateBasicInfo}
          onAddCustomAttribute={handleAddCustomAttribute}
          onUpdateCustomAttribute={handleUpdateCustomAttribute}
          onDeleteCustomAttribute={handleDeleteCustomAttribute}
        />

        {/* AI 洞察 */}
        <AIInsightsSection
          satisfactionScore={aiInsights?.satisfaction_score ?? null}
          emotionScore={aiInsights?.emotion_score ?? null}
          intent={aiInsights?.intent ?? null}
          insightSummary={aiInsights?.insight_summary ?? null}
        />

        {/* Tags */}
        <div className="pt-4 space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="text-xs font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">{t('visitor.tags.title', '标签')}</h4>
            {isUpdating && (
              <div className="flex items-center text-xs text-blue-600 dark:text-blue-400">
                <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                {t('common.updating', '更新中')}
              </div>
            )}
          </div>
          <TagManager
            tags={visitor.tags}
            onAddTag={handleAddTag}
            onUpdateTag={handleUpdateTag}
            onRemoveTag={handleRemoveTag}
            fetchCommonTags={fetchCommonTags}
            onAssociateExistingTag={handleAssociateExistingTag}
            maxTags={8}
            className="mt-1"
          />
        </div>

        {/* System Info Section */}
        <SystemInfoSection
          systemInfo={useChannelMode ? systemInfo : (visitorData?.system_info || null)}
          language={channelExtra?.language || visitorData?.language || undefined}
          timezone={channelExtra?.timezone || visitorData?.timezone || undefined}
          ipAddress={channelExtra?.ip_address || visitorData?.ip_address || undefined}
          displayLocation={channelExtra?.display_location || visitorData?.display_location || undefined}
        />

        {/* Recent Activity Section */}
        <RecentActivitySection activities={useChannelMode ? recentActivities : (visitorData?.recent_activities || [])} />
      </div>

      {/* Avatar Crop Modal */}
        <ImageCropModal
        isOpen={showCropModal}
        imageSrc={cropImageSrc}
        aspect={1}
        mimeType={cropImageMimeType}
        title={t('visitor.avatar.cropTitle', '裁剪头像')}
        onCancel={handleCropCancel}
        onConfirm={handleCropConfirm}
      />
    </div>
  );
};

export default VisitorDetailPanel;
