/**
 * Models Store
 * Manages AI models state using Zustand
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { AIModelsApiService, ModelTransformUtils } from '@/services/aiModelsApi';
import type { ModelInfo, ModelQueryParams } from '@/types';

interface ModelsState {
  // State
  models: ModelInfo[];
  isLoading: boolean;
  error: string | null;
  lastFetched: Date | null;

  // Actions
  loadModels: (params?: ModelQueryParams) => Promise<void>;
  refreshModels: (params?: ModelQueryParams) => Promise<void>;
  clearError: () => void;
  reset: () => void;

  // Getters
  getActiveModels: () => ModelInfo[];
  getChatModels: () => ModelInfo[];
  getModelsByProvider: (provider: string) => ModelInfo[];
  getModelById: (id: string) => ModelInfo | null;
  getDefaultModel: () => ModelInfo | null;
  validateModelId: (id: string) => boolean;
}

const initialState = {
  models: [],
  isLoading: false,
  error: null,
  lastFetched: null,
};

export const useModelsStore = create<ModelsState>()(
  devtools(
    (set, get) => ({
      ...initialState,

      // Actions
      loadModels: async (params) => {
        const currentState = get();
        
        // Prevent multiple concurrent API calls
        if (currentState.isLoading) {
          console.log('loadModels: Already loading, skipping duplicate call');
          return;
        }

        set({ isLoading: true, error: null }, false, 'loadModels:start');

        try {
          // Default to active chat models for agent use
          const queryParams: ModelQueryParams = {
            status: 'active',
            type: 'chat',
            limit: 100, // Get all available models
            ...params,
          };

          const response = await AIModelsApiService.getModels(queryParams);
          
          set({
            models: response.data,
            isLoading: false,
            error: null,
            lastFetched: new Date(),
          }, false, 'loadModels:success');
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to load models';
          set({
            isLoading: false,
            error: errorMessage,
          }, false, 'loadModels:error');
          console.error('Failed to load models:', error);
          throw error;
        }
      },

      refreshModels: async (params) => {
        // Force refresh by clearing cache
        set({ lastFetched: null }, false, 'refreshModels:start');
        await get().loadModels(params);
      },

      clearError: () => {
        set({ error: null }, false, 'clearError');
      },

      reset: () => {
        set(initialState, false, 'reset');
      },

      // Getters
      getActiveModels: () => {
        const { models } = get();
        return ModelTransformUtils.filterModelsForAgents(models);
      },

      getChatModels: () => {
        const { models } = get();
        return models.filter(model => model.model_type === 'chat');
      },

      getModelsByProvider: (provider: string) => {
        const { models } = get();
        return models.filter(model => 
          model.provider.toLowerCase() === provider.toLowerCase()
        );
      },

      getModelById: (id: string) => {
        const { models } = get();
        return ModelTransformUtils.findModelById(id, models);
      },

      getDefaultModel: () => {
        const { models } = get();
        return ModelTransformUtils.getDefaultModel(models);
      },

      validateModelId: (id: string) => {
        const { models } = get();
        return ModelTransformUtils.validateModelId(id, models);
      },
    }),
    {
      name: 'models-store',
      partialize: (state: ModelsState) => ({
        models: state.models,
        lastFetched: state.lastFetched,
      }),
    }
  )
);

// Selectors for optimized component subscriptions
export const modelsSelectors = {
  models: (state: ModelsState) => state.models,
  isLoading: (state: ModelsState) => state.isLoading,
  error: (state: ModelsState) => state.error,
  activeModels: (state: ModelsState) => state.getActiveModels(),
  chatModels: (state: ModelsState) => state.getChatModels(),
} as const;

// Export store instance for direct access if needed
export default useModelsStore;
