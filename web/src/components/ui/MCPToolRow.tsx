import React from 'react';
import { XCircle } from 'lucide-react';
import { generateDefaultAvatar } from '@/utils/avatarUtils';
import type { MCPTool } from '@/types';

interface MCPToolRowProps {
  tool: MCPTool;
  onRemove: () => void;
  disabled?: boolean;
}

const MCPToolRow: React.FC<MCPToolRowProps> = ({
  tool,
  onRemove,
  disabled = false,
}) => {
  const displayName = tool.title || tool.name;
  const toolAvatar = generateDefaultAvatar(displayName, tool.id);

  return (
    <div className="flex items-center justify-between p-3 rounded-md border border-gray-200 bg-white transition-colors">
      <div className="flex items-center space-x-3 min-w-0 flex-1">
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold text-sm ${toolAvatar.colorClass}`}>
          {toolAvatar.letter}
        </div>
        <div className="min-w-0">
          <div className="text-sm font-semibold text-gray-800 truncate">{displayName}</div>
          <div className="text-xs text-gray-500 truncate">{tool.short_no || tool.author}</div>
          <div className="text-xs text-gray-500 mt-1 leading-snug line-clamp-2">{tool.description}</div>
        </div>
      </div>
      <div className="flex items-center shrink-0">
        <button
          type="button"
          onClick={onRemove}
          className="text-gray-400 hover:text-red-500 transition-colors"
          disabled={disabled}
          title="移除工具"
        >
          <XCircle className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
};

export default MCPToolRow;
