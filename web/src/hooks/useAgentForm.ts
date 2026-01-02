import { useCallback, useState } from 'react';

export interface AgentFormState {
  name: string;
  profession: string;
  description: string;
  llmModel: string;
  mcpTools: string[];
  mcpToolConfigs: Record<string, Record<string, any>>;
  knowledgeBases: string[];
}

export interface UseAgentFormOptions {
  initial?: Partial<AgentFormState>;
  // Controlled mode (e.g., Create modal uses store-backed form)
  controlledFormData?: AgentFormState;
  onFormDataChange?: (update: Partial<AgentFormState>) => void;
}

export interface UseAgentFormResult {
  formData: AgentFormState;
  setFormData: (update: Partial<AgentFormState>) => void;
  handleInputChange: (field: keyof AgentFormState, value: any) => void;
  // Tools
  removeTool: (toolId: string) => void;
  setToolConfig: (toolId: string, config: Record<string, any>) => void;
  // Knowledge bases
  removeKnowledgeBase: (kbId: string) => void;
  // Reset
  reset: (next?: Partial<AgentFormState>) => void;
}

const defaultForm: AgentFormState = {
  name: '',
  profession: '',
  description: '',
  llmModel: 'gemini-1.5-pro',
  mcpTools: [],
  mcpToolConfigs: {},
  knowledgeBases: [],
};

export function useAgentForm(options: UseAgentFormOptions = {}): UseAgentFormResult {
  const {
    initial,
    controlledFormData,
    onFormDataChange,
  } = options;

  // Form data: controlled or uncontrolled
  const [uncontrolledForm, setUncontrolledForm] = useState<AgentFormState>({
    ...defaultForm,
    ...(initial || {}),
  });

  const formData = (controlledFormData || uncontrolledForm);

  const setFormData = useCallback((update: Partial<AgentFormState>) => {
    if (onFormDataChange) {
      onFormDataChange(update);
    } else {
      setUncontrolledForm(prev => ({ ...prev, ...update }));
    }
  }, [onFormDataChange]);

  const handleInputChange = useCallback((field: keyof AgentFormState, value: any) => {
    setFormData({ [field]: value } as Partial<AgentFormState>);
  }, [setFormData]);

  const removeTool = useCallback((toolId: string) => {
    // remove from list
    const newTools = (formData.mcpTools || []).filter(id => id !== toolId);
    setFormData({ mcpTools: newTools });
  }, [formData.mcpTools, setFormData]);

  const setToolConfig = useCallback((toolId: string, config: Record<string, any>) => {
    setFormData({
      mcpToolConfigs: {
        ...formData.mcpToolConfigs,
        [toolId]: config,
      },
    });
  }, [formData.mcpToolConfigs, setFormData]);

  // Knowledge bases
  const removeKnowledgeBase = useCallback((kbId: string) => {
    const newKBs = (formData.knowledgeBases || []).filter(id => id !== kbId);
    setFormData({ knowledgeBases: newKBs });
  }, [formData.knowledgeBases, setFormData]);

  // Reset API
  const reset = useCallback((next?: Partial<AgentFormState>) => {
    if (next) {
      if (onFormDataChange) {
        onFormDataChange(next);
      } else {
        setUncontrolledForm({ ...defaultForm, ...next });
      }
    } else if (!onFormDataChange) {
      setUncontrolledForm({ ...defaultForm });
    }
  }, [onFormDataChange]);

  return {
    formData,
    setFormData,
    handleInputChange,
    removeTool,
    setToolConfig,
    removeKnowledgeBase,
    reset,
  };
}

export default useAgentForm;
