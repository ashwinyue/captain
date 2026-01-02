/**
 * Optimized MCP Tools Transformation Utilities
 * Consolidates all MCP tool transformation logic using base patterns
 */

import { BaseTransformerClass, TransformUtils } from './base/BaseTransform';
import type {
  ToolSummary,
  ToolResponse,
  ToolStoreItem,
  MCPTool
} from '@/types';

/**
 * Tool Summary to Store Item Transformer
 */
export class ToolSummaryToStoreItemTransformer extends BaseTransformerClass<ToolSummary, ToolStoreItem> {
  transform(toolSummary: ToolSummary): ToolStoreItem {
    return {
      id: toolSummary.id,
      name: TransformUtils.sanitizeString(toolSummary.name),
      title: TransformUtils.createDisplayName(toolSummary.title, toolSummary.name),
      description: TransformUtils.sanitizeString(toolSummary.description, '暂无描述'),
      version: TransformUtils.sanitizeString(toolSummary.version, '1.0.0'),
      category: TransformUtils.sanitizeString(toolSummary.category, '其他'),
      tags: TransformUtils.extractTags(toolSummary.tags),
      
      // Use short_no as author field for better identification
      author: TransformUtils.sanitizeString(toolSummary.short_no) || 
              TransformUtils.getAuthor(toolSummary.tool_source_type),
      authorHandle: `@${(toolSummary.short_no || TransformUtils.getAuthor(toolSummary.tool_source_type))
        .toLowerCase().replace(/\s+/g, '')}`,
      
      // Generate metrics based on execution count
      rating: TransformUtils.generateRating(toolSummary.execution_count),
      ratingCount: Math.floor((toolSummary.execution_count || 0) / 10) || 1,
      downloads: TransformUtils.sanitizeNumber(toolSummary.execution_count),
      
      // Format dates and status
      lastUpdated: TransformUtils.formatDate(toolSummary.created_at),
      featured: (toolSummary.execution_count || 0) > 100,
      verified: toolSummary.status === 'ACTIVE',
      
      // Default UI fields
      icon: '🔧',
      screenshots: [],
      longDescription: TransformUtils.sanitizeString(toolSummary.description, '暂无描述'),
      requirements: [],
      changelog: '',
      mcpMethods: [],
      
      // Preserve API fields
      isInstalled: toolSummary.is_installed || false,
      input_schema: toolSummary.input_schema,
      short_no: toolSummary.short_no || undefined,
    };
  }
}

/**
 * Store Item to MCP Tool Transformer
 */
export class StoreItemToMCPToolTransformer extends BaseTransformerClass<ToolStoreItem, MCPTool> {
  transform(storeItem: ToolStoreItem): MCPTool {
    // Determine status based on installation and verification
    const status = storeItem.isInstalled ? 'active' : 
                   storeItem.verified ? 'available' : 'inactive';

    return {
      id: storeItem.id,
      name: storeItem.name,
      title: TransformUtils.createDisplayName(storeItem.title, storeItem.name),
      description: storeItem.description,
      category: TransformUtils.transformCategory(storeItem.category),
      status: TransformUtils.transformToolStatus(status as any),
      version: storeItem.version,
      author: storeItem.author,
      lastUpdated: storeItem.lastUpdated,
      usageCount: storeItem.downloads,
      rating: storeItem.rating,
      tags: storeItem.tags,
      
      // Optional fields with reasonable defaults
      capabilities: storeItem.tags.length > 0 ? storeItem.tags : undefined,
      successRate: storeItem.rating >= 4.0 ? 0.95 : 0.85,
      avgResponseTime: TransformUtils.generateAvgResponseTime(),
      config: {
        featured: storeItem.featured,
        verified: storeItem.verified,
        ratingCount: storeItem.ratingCount,
      },
      
      // Schema handling - prioritize API schema, fallback to mcpMethods
      input_schema: storeItem.input_schema || 
                   TransformUtils.createSchemaFromMethods(storeItem.mcpMethods || []),
      
      // Preserve short_no from API
      short_no: storeItem.short_no,
    };
  }
}

/**
 * Tool Response to MCP Tool Transformer
 */
export class ToolResponseToMCPToolTransformer extends BaseTransformerClass<ToolResponse, MCPTool> {
  transform(toolResponse: ToolResponse): MCPTool {
    return {
      id: toolResponse.id,
      name: TransformUtils.sanitizeString(toolResponse.name),
      title: TransformUtils.createDisplayName(toolResponse.title, toolResponse.name),
      description: TransformUtils.sanitizeString(toolResponse.description, '暂无描述'),
      version: TransformUtils.sanitizeString(toolResponse.version, '1.0.0'),
      category: TransformUtils.transformCategory(toolResponse.category),
      tags: TransformUtils.extractTags(toolResponse.tags),
      status: TransformUtils.transformToolStatus(toolResponse.status),
      author: TransformUtils.getAuthor(toolResponse.tool_source_type),
      lastUpdated: TransformUtils.formatDate(toolResponse.updated_at),
      usageCount: TransformUtils.sanitizeNumber(toolResponse.execution_count),
      rating: TransformUtils.generateRating(toolResponse.execution_count),
      
      // Optional fields
      config: toolResponse.meta_data || undefined,
      capabilities: toolResponse.tags || undefined,
      successRate: TransformUtils.generateSuccessRate(toolResponse.execution_count),
      avgResponseTime: TransformUtils.generateAvgResponseTime(),
      input_schema: toolResponse.input_schema,
    };
  }
}

/**
 * MCP Tool to Store Item Transformer (reverse transformation)
 */
export class MCPToolToStoreItemTransformer extends BaseTransformerClass<MCPTool, ToolStoreItem> {
  transform(mcpTool: MCPTool): ToolStoreItem {
    return {
      id: mcpTool.id,
      name: mcpTool.name,
      title: TransformUtils.createDisplayName(mcpTool.title, mcpTool.name),
      description: mcpTool.description,
      author: mcpTool.author,
      authorHandle: `@${mcpTool.author.toLowerCase().replace(/\s+/g, '')}`,
      version: mcpTool.version,
      category: mcpTool.category,
      tags: mcpTool.tags,
      rating: mcpTool.rating,
      ratingCount: Math.floor(mcpTool.usageCount / 10),
      downloads: mcpTool.usageCount,
      lastUpdated: mcpTool.lastUpdated,
      featured: mcpTool.rating >= 4.5,
      verified: mcpTool.status === 'active',
      icon: '',
      screenshots: [],
      longDescription: mcpTool.description,
      requirements: [],
      changelog: '',
      input_schema: mcpTool.input_schema,
      short_no: mcpTool.short_no,
    };
  }
}

// Create transformer instances
const toolSummaryToStoreItemTransformer = new ToolSummaryToStoreItemTransformer();
const storeItemToMCPToolTransformer = new StoreItemToMCPToolTransformer();
const toolResponseToMCPToolTransformer = new ToolResponseToMCPToolTransformer();
const mcpToolToStoreItemTransformer = new MCPToolToStoreItemTransformer();

/**
 * Optimized transformation functions using the new pattern
 */
export const OptimizedTransforms = {
  // Single item transformations
  toolSummaryToStoreItem: (item: ToolSummary): ToolStoreItem => 
    toolSummaryToStoreItemTransformer.transform(item),
  
  storeItemToMCPTool: (item: ToolStoreItem): MCPTool => 
    storeItemToMCPToolTransformer.transform(item),
  
  toolResponseToMCPTool: (item: ToolResponse): MCPTool => 
    toolResponseToMCPToolTransformer.transform(item),
  
  mcpToolToStoreItem: (item: MCPTool): ToolStoreItem => 
    mcpToolToStoreItemTransformer.transform(item),

  // Batch transformations
  toolSummariesToStoreItems: (items: ToolSummary[]): ToolStoreItem[] => 
    toolSummaryToStoreItemTransformer.transformMany(items),
  
  storeItemsToMCPTools: (items: ToolStoreItem[]): MCPTool[] => 
    storeItemToMCPToolTransformer.transformMany(items),
  
  toolResponsesToMCPTools: (items: ToolResponse[]): MCPTool[] => 
    toolResponseToMCPToolTransformer.transformMany(items),
  
  mcpToolsToStoreItems: (items: MCPTool[]): ToolStoreItem[] => 
    mcpToolToStoreItemTransformer.transformMany(items),
};

// Register transformers for global access (using TransformRegistry)
import { TransformRegistry } from './base/BaseTransform';
TransformRegistry.register('toolSummaryToStoreItem', toolSummaryToStoreItemTransformer);
TransformRegistry.register('storeItemToMCPTool', storeItemToMCPToolTransformer);
TransformRegistry.register('toolResponseToMCPTool', toolResponseToMCPToolTransformer);
TransformRegistry.register('mcpToolToStoreItem', mcpToolToStoreItemTransformer);

// Export individual transformers for direct use
export {
  toolSummaryToStoreItemTransformer,
  storeItemToMCPToolTransformer,
  toolResponseToMCPToolTransformer,
  mcpToolToStoreItemTransformer,
};

// Backward compatibility exports (can be removed after migration)
export const transformToolSummaryToStoreItem = OptimizedTransforms.toolSummaryToStoreItem;
export const transformStoreItemToMCPTool = OptimizedTransforms.storeItemToMCPTool;
export const transformToolResponseToMCPTool = OptimizedTransforms.toolResponseToMCPTool;
export const transformMCPToolToStoreItem = OptimizedTransforms.mcpToolToStoreItem;
export const transformToolSummariesToStoreItems = OptimizedTransforms.toolSummariesToStoreItems;

export default OptimizedTransforms;
