import React from 'react';
import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LuBot, LuWrench } from 'react-icons/lu';
import { AI_MENU_ITEMS } from '@/utils/constants';
import type { NavigationItem } from '@/types';
import OnboardingSidebarPanel from '@/components/onboarding/OnboardingSidebarPanel';

interface AIMenuItemProps {
  item: NavigationItem;
}

// Icon mapping for AI menu items using react-icons
const ICON_MAP: Record<string, React.ComponentType<{ className?: string }>> = {
  'Bot': LuBot,
  'Wrench': LuWrench
};

/**
 * AI menu navigation item component using React Router NavLink
 */
const AIMenuItem: React.FC<AIMenuItemProps> = ({ item }) => {
  const { t } = useTranslation();
  const IconComponent = ICON_MAP[item.icon];

  return (
    <NavLink
      to={item.path}
      className={({ isActive }) => `
        flex items-center px-3 py-2 rounded-md text-sm transition-colors w-full text-left
        ${isActive
          ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 font-medium'
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100/70 dark:hover:bg-gray-700/70 hover:text-gray-800 dark:hover:text-gray-200'
        }
      `}
    >
      {IconComponent && <IconComponent className="w-4 h-4 mr-2" />}
      {t(item.title)}
    </NavLink>
  );
};

/**
 * AI feature menu component
 */
const AIMenu: React.FC = () => {
  const { t } = useTranslation();
  return (
    <aside className="w-64 bg-white/90 dark:bg-gray-800/90 backdrop-blur-lg border-r border-gray-200/60 dark:border-gray-700/60 flex flex-col shrink-0">
      {/* Header */}
      <div className="p-4 border-b border-gray-200/60 dark:border-gray-700/60 sticky top-0 bg-white/90 dark:bg-gray-800/90 backdrop-blur-lg z-10">
        <h3 className="text-md font-semibold text-gray-800 dark:text-gray-200 px-1">{t('ai.menu.title', 'AI 功能')}</h3>
      </div>

      {/* Menu Navigation */}
      <nav className="flex-grow overflow-y-auto p-3 space-y-1">
        {AI_MENU_ITEMS.map((item) => (
          <AIMenuItem
            key={item.id}
            item={item}
          />
        ))}
      </nav>

      {/* Onboarding Panel */}
      <OnboardingSidebarPanel />
    </aside>
  );
};

export default AIMenu;
