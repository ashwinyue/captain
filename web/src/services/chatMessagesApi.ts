import { BaseApiService } from './base/BaseApiService';

// Request type based on OpenAPI docs#/components/schemas/StaffSendPlatformMessageRequest
export interface StaffSendPlatformMessageRequest {
  channel_id: string;
  channel_type: number; // WuKongIM channel type, customer service chat uses 251
  payload: Record<string, any>; // Platform Service message payload
  client_msg_no?: string | null; // Optional idempotency key
}

// Response type is not specified in OpenAPI schema (empty). Use unknown for now.
export type StaffSendPlatformMessageResponse = unknown;

// Request type for staff-to-team/agent chat based on OpenAPI docs
// Either team_id or agent_id must be provided (exactly one)
export interface StaffTeamChatRequest {
  team_id?: string | null; // AI Team ID to chat with (UUID format)
  agent_id?: string | null; // AI Agent ID to chat with (UUID format)
  message: string; // Message content to send
  system_message?: string | null; // Optional system message/prompt
  expected_output?: string | null; // Optional expected output format
  timeout_seconds?: number | null; // Timeout in seconds (1-600, default 120)
}

// Response type for staff-to-team/agent chat
export interface StaffTeamChatResponse {
  success: boolean; // Whether the chat completed successfully
  message: string; // Status message
  client_msg_no: string; // Message correlation ID for tracking
}

// SSE Event types from aicenter
export interface SSEEvent {
  type?: string;
  agent_name?: string;
  content?: string;
}

class ChatMessagesApiService extends BaseApiService {
  protected readonly apiVersion = 'v1';
  protected readonly endpoints = {
    sendPlatformMessage: '/v1/chat/messages/send',
    teamChat: '/v1/chat/team',
    teamChatStream: '/v1/chat/team/stream',
  } as const;

  /**
   * Forward a staff-authenticated outbound message to the Platform Service.
   * This must be called before sending via WebSocket for non-website platforms.
   */
  async staffSendPlatformMessage(
    data: StaffSendPlatformMessageRequest
  ): Promise<StaffSendPlatformMessageResponse> {
    return this.post<StaffSendPlatformMessageResponse>(this.endpoints.sendPlatformMessage, data);
  }

  /**
   * Staff chat with AI team or agent.
   * This endpoint allows authenticated staff members to chat with AI teams or agents.
   * Either team_id or agent_id must be provided (exactly one).
   * The AI response is delivered via WuKongIM to the client.
   */
  async staffTeamChat(
    data: StaffTeamChatRequest
  ): Promise<StaffTeamChatResponse> {
    return this.post<StaffTeamChatResponse>(this.endpoints.teamChat, data);
  }

  /**
   * Staff chat with AI team or agent (streaming).
   * This endpoint returns SSE stream for real-time AI responses.
   * @param data Request data
   * @param onEvent Callback for each SSE event
   * @param onComplete Callback when stream completes
   * @param onError Callback for errors
   */
  async staffTeamChatStream(
    data: StaffTeamChatRequest,
    onEvent: (event: SSEEvent) => void,
    onComplete?: (finalContent: string) => void,
    onError?: (error: Error) => void
  ): Promise<void> {
    const token = localStorage.getItem('tgo-auth-token');
    // Use runtime config (window.ENV) with fallback to build-time config
    const baseUrl = (window as any).ENV?.VITE_API_BASE_URL || import.meta.env.VITE_API_BASE_URL || '';
    
    try {
      const response = await fetch(`${baseUrl}${this.endpoints.teamChatStream}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : '',
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('No response body');
      }

      const decoder = new TextDecoder();
      let buffer = '';
      let finalContent = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('data:')) {
            const jsonStr = line.slice(5).trim();
            if (jsonStr && jsonStr !== '') {
              try {
                const event = JSON.parse(jsonStr) as SSEEvent;
                onEvent(event);
                // Collect exit content as final response
                if (event.type === 'exit' && event.content) {
                  finalContent = event.content;
                }
              } catch (e) {
                console.warn('Failed to parse SSE event:', jsonStr);
              }
            }
          }
        }
      }

      onComplete?.(finalContent);
    } catch (error) {
      onError?.(error as Error);
    }
  }
}

export const chatMessagesApiService = new ChatMessagesApiService();
export default chatMessagesApiService;

