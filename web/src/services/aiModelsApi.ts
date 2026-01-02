/**
 * AI Models API Service
 * Handles AI Models API endpoints following the BaseApiService pattern
 */

import BaseApiService from './base/BaseApiService';
import type { 
  ModelListResponse, 
  ModelInfo, 
  ModelQueryParams 
} from '@/types';

/**
 * AI Models API Service Class
 */
export class AIModelsApiService extends BaseApiService {
  protected readonly apiVersion = 'v1';
  protected readonly endpoints = {
    MODELS: `/${this.apiVersion}/ai/models`,
    MODEL_BY_ID: (id: string) => `/${this.apiVersion}/ai/models/${id}`,
  } as const;

  /**
   * Get paginated list of AI models
   */
  static async getModels(params?: ModelQueryParams): Promise<ModelListResponse> {
    const service = new AIModelsApiService();
    return service.get<ModelListResponse>(service.endpoints.MODELS, params);
  }

  /**
   * Get a specific model by ID
   */
  static async getModel(id: string): Promise<ModelInfo> {
    const service = new AIModelsApiService();
    const endpoint = service.endpoints.MODEL_BY_ID(id);
    return service.get<ModelInfo>(endpoint);
  }

  /**
   * Get models by provider
   */
  static async getModelsByProvider(
    provider: string,
    params?: Omit<ModelQueryParams, 'provider'>
  ): Promise<ModelListResponse> {
    const service = new AIModelsApiService();
    return service.get<ModelListResponse>(service.endpoints.MODELS, { ...params, provider });
  }

  /**
   * Get models by type
   */
  static async getModelsByType(
    type: string,
    params?: Omit<ModelQueryParams, 'type'>
  ): Promise<ModelListResponse> {
    const service = new AIModelsApiService();
    return service.get<ModelListResponse>(service.endpoints.MODELS, { ...params, type });
  }

  /**
   * Get active models only
   */
  static async getActiveModels(
    params?: Omit<ModelQueryParams, 'status'>
  ): Promise<ModelListResponse> {
    const service = new AIModelsApiService();
    return service.get<ModelListResponse>(service.endpoints.MODELS, { ...params, status: 'active' });
  }

  /**
   * Get chat models only (most common for agents)
   */
  static async getChatModels(
    params?: Omit<ModelQueryParams, 'type'>
  ): Promise<ModelListResponse> {
    const service = new AIModelsApiService();
    return service.get<ModelListResponse>(service.endpoints.MODELS, { ...params, type: 'chat' });
  }

  /**
   * Get models with pagination
   */
  static async getModelsPage(
    page: number,
    pageSize: number = 50,
    filters?: Omit<ModelQueryParams, 'limit' | 'offset'>
  ): Promise<ModelListResponse> {
    const service = new AIModelsApiService();
    const offset = (page - 1) * pageSize;
    return service.get<ModelListResponse>(service.endpoints.MODELS, {
      ...filters,
      limit: pageSize,
      offset,
    });
  }
}

/**
 * Model transformation utilities
 */
export class ModelTransformUtils {
  /**
   * Transform ModelInfo to dropdown option format
   */
  static transformModelToOption(model: ModelInfo): { value: string; label: string; provider: string; status: string } {
    return {
      value: model.id,
      label: model.display_name,
      provider: model.provider,
      status: model.status,
    };
  }

  /**
   * Transform ModelListResponse to dropdown options
   */
  static transformModelsToOptions(response: ModelListResponse): Array<{ value: string; label: string; provider: string; status: string }> {
    return response.data.map(this.transformModelToOption);
  }

  /**
   * Filter models for agent use (active chat models)
   */
  static filterModelsForAgents(models: ModelInfo[]): ModelInfo[] {
    return models.filter(model => 
      model.status === 'active' && 
      model.model_type === 'chat'
    );
  }

  /**
   * Group models by provider
   */
  static groupModelsByProvider(models: ModelInfo[]): Record<string, ModelInfo[]> {
    return models.reduce((groups, model) => {
      const provider = model.provider;
      if (!groups[provider]) {
        groups[provider] = [];
      }
      groups[provider].push(model);
      return groups;
    }, {} as Record<string, ModelInfo[]>);
  }

  /**
   * Get default model (first active chat model)
   */
  static getDefaultModel(models: ModelInfo[]): ModelInfo | null {
    const chatModels = this.filterModelsForAgents(models);
    return chatModels.length > 0 ? chatModels[0] : null;
  }

  /**
   * Validate if a model ID exists in the available models
   */
  static validateModelId(modelId: string, models: ModelInfo[]): boolean {
    return models.some(model => model.id === modelId);
  }

  /**
   * Find model by ID
   */
  static findModelById(modelId: string, models: ModelInfo[]): ModelInfo | null {
    return models.find(model => model.id === modelId) || null;
  }
}
