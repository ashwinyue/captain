/**
 * MCP Tools Store
 * Zustand store for managing MCP tools state
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { ToolSummary, ToolListResponse, ProjectToolsMeta, ToolResponse } from '@/types';
import { MCPToolsApiService, type MCPToolsQueryParams } from '@/services/mcpToolsApi';

// Store state interface
interface MCPToolsState {
  // Data
  tools: ToolSummary[];
  selectedTool: ToolSummary | null;
  toolDetails: ToolResponse | null;
  meta: ProjectToolsMeta | null;
  
  // Loading states
  isLoading: boolean;
  isLoadingTool: boolean;
  isLoadingToolDetails: boolean;
  isSearching: boolean;
  
  // Error states
  error: string | null;
  toolError: string | null;
  toolDetailsError: string | null;
  
  // Filters and search
  searchQuery: string;
  selectedCategory: string;
  selectedStatus: string;
  currentPage: number;
  pageSize: number;
  
  // Actions
  setTools: (tools: ToolSummary[]) => void;
  setSelectedTool: (tool: ToolSummary | null) => void;
  setToolDetails: (toolDetails: ToolResponse | null) => void;
  setMeta: (meta: ProjectToolsMeta | null) => void;
  setLoading: (loading: boolean) => void;
  setLoadingTool: (loading: boolean) => void;
  setLoadingToolDetails: (loading: boolean) => void;
  setSearching: (searching: boolean) => void;
  setError: (error: string | null) => void;
  setToolError: (error: string | null) => void;
  setToolDetailsError: (error: string | null) => void;
  setSearchQuery: (query: string) => void;
  setSelectedCategory: (category: string) => void;
  setSelectedStatus: (status: string) => void;
  setCurrentPage: (page: number) => void;
  setPageSize: (size: number) => void;
  
  // Async actions
  loadTools: (params?: MCPToolsQueryParams) => Promise<void>;
  loadTool: (id: string) => Promise<void>;
  loadToolDetails: (id: string) => Promise<void>;
  searchTools: (query: string, filters?: { category?: string; status?: string }) => Promise<void>;
  loadToolsByCategory: (category: string) => Promise<void>;
  loadToolsByStatus: (status: string) => Promise<void>;
  loadPage: (page: number) => Promise<void>;
  refreshTools: () => Promise<void>;
  clearError: () => void;
  clearToolError: () => void;
  clearToolDetailsError: () => void;
  reset: () => void;
}

// Initial state
const initialState = {
  tools: [],
  selectedTool: null,
  toolDetails: null,
  meta: null,
  isLoading: false,
  isLoadingTool: false,
  isLoadingToolDetails: false,
  isSearching: false,
  error: null,
  toolError: null,
  toolDetailsError: null,
  searchQuery: '',
  selectedCategory: 'all',
  selectedStatus: 'all',
  currentPage: 1,
  pageSize: 20,
};

// Create the store
export const useMCPToolsStore = create<MCPToolsState>()(
  devtools(
    (set, get) => ({
      ...initialState,

      // Synchronous actions
      setTools: (tools) => set({ tools }, false, 'setTools'),
      setSelectedTool: (tool) => set({ selectedTool: tool }, false, 'setSelectedTool'),
      setToolDetails: (toolDetails) => set({ toolDetails }, false, 'setToolDetails'),
      setMeta: (meta) => set({ meta }, false, 'setMeta'),
      setLoading: (loading) => set({ isLoading: loading }, false, 'setLoading'),
      setLoadingTool: (loading) => set({ isLoadingTool: loading }, false, 'setLoadingTool'),
      setLoadingToolDetails: (loading) => set({ isLoadingToolDetails: loading }, false, 'setLoadingToolDetails'),
      setSearching: (searching) => set({ isSearching: searching }, false, 'setSearching'),
      setError: (error) => set({ error }, false, 'setError'),
      setToolError: (error) => set({ toolError: error }, false, 'setToolError'),
      setToolDetailsError: (error) => set({ toolDetailsError: error }, false, 'setToolDetailsError'),
      setSearchQuery: (query) => set({ searchQuery: query }, false, 'setSearchQuery'),
      setSelectedCategory: (category) => set({ selectedCategory: category }, false, 'setSelectedCategory'),
      setSelectedStatus: (status) => set({ selectedStatus: status }, false, 'setSelectedStatus'),
      setCurrentPage: (page) => set({ currentPage: page }, false, 'setCurrentPage'),
      setPageSize: (size) => set({ pageSize: size }, false, 'setPageSize'),

      // Async actions
      loadTools: async (params) => {
        const { setLoading, setError, setTools, setMeta } = get();

        setLoading(true);
        setError(null);

        try {
          const response: ToolListResponse = await MCPToolsApiService.getTools(params);



          setTools(response.data);
          setMeta(response.meta);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load tools';
          setError(errorMessage);
          console.error('Failed to load MCP tools:', error);


        } finally {
          setLoading(false);
        }
      },

      loadTool: async (id) => {
        const { setLoadingTool, setToolError, setSelectedTool } = get();

        setLoadingTool(true);
        setToolError(null);

        try {
          const tool = await MCPToolsApiService.getTool(id);
          setSelectedTool(tool);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load tool';
          setToolError(errorMessage);
          console.error(`Failed to load MCP tool ${id}:`, error);
        } finally {
          setLoadingTool(false);
        }
      },

      loadToolDetails: async (id) => {
        const { setLoadingToolDetails, setToolDetailsError, setToolDetails } = get();

        setLoadingToolDetails(true);
        setToolDetailsError(null);

        try {
          const toolDetails = await MCPToolsApiService.getToolDetails(id);
          setToolDetails(toolDetails);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load tool details';
          setToolDetailsError(errorMessage);
          console.error(`Failed to load MCP tool details ${id}:`, error);
        } finally {
          setLoadingToolDetails(false);
        }
      },

      searchTools: async (query, filters) => {
        const { setSearching, setError, setTools, setMeta, setSearchQuery } = get();

        setSearching(true);
        setError(null);
        setSearchQuery(query);

        try {
          const response = await MCPToolsApiService.searchTools(query, filters);
          setTools(response.data);
          setMeta(response.meta);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to search tools';
          setError(errorMessage);
          console.error('Failed to search MCP tools:', error);
        } finally {
          setSearching(false);
        }
      },

      loadToolsByCategory: async (category) => {
        const { setLoading, setError, setTools, setMeta, setSelectedCategory } = get();

        setLoading(true);
        setError(null);
        setSelectedCategory(category);

        try {
          const response = await MCPToolsApiService.getToolsByCategory(category);
          setTools(response.data);
          setMeta(response.meta);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load tools by category';
          setError(errorMessage);
          console.error(`Failed to load tools for category ${category}:`, error);
        } finally {
          setLoading(false);
        }
      },

      loadToolsByStatus: async (status) => {
        const { setLoading, setError, setTools, setMeta, setSelectedStatus } = get();

        setLoading(true);
        setError(null);
        setSelectedStatus(status);

        try {
          const response = await MCPToolsApiService.getToolsByStatus(status);
          setTools(response.data);
          setMeta(response.meta);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load tools by status';
          setError(errorMessage);
          console.error(`Failed to load tools with status ${status}:`, error);
        } finally {
          setLoading(false);
        }
      },

      loadPage: async (page) => {
        const { setLoading, setError, setTools, setMeta, setCurrentPage, pageSize, searchQuery, selectedCategory, selectedStatus } = get();

        setLoading(true);
        setError(null);
        setCurrentPage(page);

        try {
          const filters: MCPToolsQueryParams = {};

          if (searchQuery) filters.search = searchQuery;
          if (selectedCategory && selectedCategory !== 'all') filters.category = selectedCategory;
          if (selectedStatus && selectedStatus !== 'all') filters.status = selectedStatus;

          const response = await MCPToolsApiService.getToolsPage(page, pageSize, filters);
          setTools(response.data);
          setMeta(response.meta);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load page';
          setError(errorMessage);
          console.error(`Failed to load page ${page}:`, error);
        } finally {
          setLoading(false);
        }
      },

      refreshTools: async () => {
        const { loadTools, searchQuery, selectedCategory, selectedStatus } = get();
        
        const params: MCPToolsQueryParams = {};
        if (searchQuery) params.search = searchQuery;
        if (selectedCategory && selectedCategory !== 'all') params.category = selectedCategory;
        if (selectedStatus && selectedStatus !== 'all') params.status = selectedStatus;
        
        await loadTools(params);
      },

      clearError: () => set({ error: null }, false, 'clearError'),
      clearToolError: () => set({ toolError: null }, false, 'clearToolError'),
      clearToolDetailsError: () => set({ toolDetailsError: null }, false, 'clearToolDetailsError'),
      
      reset: () => set(initialState, false, 'reset'),
    }),
    {
      name: 'mcp-tools-store',
    }
  )
);

export default useMCPToolsStore;
