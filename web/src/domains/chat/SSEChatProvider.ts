import { AbstractChatProvider, XRequest, type XRequestOptions } from '@ant-design/x-sdk';
import { BACKEND_URL } from '../../config';

export interface ToolCallItem {
  tool: string;
  emoji: string;
  label: string;
  status: string;
}

export interface SSEChunk {
  content?: string;
  thinking?: string;
  toolCall?: ToolCallItem;
  conversationId?: string;
  title?: string;
  status: string;
}

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  thinking?: string;
  toolCalls?: ToolCallItem[];
  createdAt?: string;
}

export interface ChatInput {
  content: string;
  agentId: string;
  conversationId: string;
}

export interface SSEProviderEvents {
  onTitleUpdated?: (conversationId: string, title: string) => void;
}

function finalizeToolCalls(toolCalls: ToolCallItem[]): ToolCallItem[] {
  return toolCalls.map((toolCall) => (
    toolCall.status === 'running'
      ? { ...toolCall, status: 'done' }
      : toolCall
  ));
}

function sseDebugEnabled(): boolean {
  if (typeof window !== 'undefined') {
    try {
      if (new URLSearchParams(window.location.search).get('debug_sse') === '1') return true;
    } catch {
    }
  }
  return localStorage.getItem('dotblue_debug_sse') === '1';
}

function sseLog(...args: any[]): void {
  if (!sseDebugEnabled()) return;
  console.info('[sse]', ...args);
}

class SSEChatProvider extends AbstractChatProvider<ChatMessage, ChatInput, SSEChunk> {
  transformParams(
    requestParams: Partial<ChatInput>,
    options: XRequestOptions<ChatInput, SSEChunk, ChatMessage>,
  ): ChatInput {
    return {
      ...(options?.params || {}),
      ...(requestParams || {}),
    } as ChatInput;
  }

  transformLocalMessage(requestParams: Partial<ChatInput>): ChatMessage {
    return {
      role: 'user',
      content: requestParams.content || '',
      createdAt: new Date().toISOString(),
    };
  }

  transformMessage(info: import('@ant-design/x-sdk').TransformMessage<ChatMessage, SSEChunk>): ChatMessage {
    const { chunk, chunks, originMessage } = info;

    // onSuccess: chunks has all accumulated data, no chunk
    if (!chunk && chunks.length > 0) {
      let content = '';
      let thinking = '';
      const toolCalls: ToolCallItem[] = [];
      for (const c of chunks) {
        if (c.content) content += c.content;
        if (c.thinking) thinking += (thinking ? '\n\n' : '') + c.thinking;
        if (c.toolCall) toolCalls.push(c.toolCall);
      }
      return {
        role: 'assistant',
        content,
        ...(thinking ? { thinking } : {}),
        ...(toolCalls.length > 0 ? { toolCalls: finalizeToolCalls(toolCalls) } : {}),
        createdAt: originMessage?.createdAt || new Date().toISOString(),
      };
    }

    // onUpdate: only chunk (latest), chunks is []. Merge with originMessage.
    if (chunk) {
      // Skip meta chunks (title notifications) from message content
      if (chunk.conversationId || chunk.title) {
        return originMessage || { role: 'assistant', content: '' };
      }
      const prev = originMessage || { role: 'assistant' as const, content: '' };
      const content = (prev.content || '') + (chunk.content || '');
      const thinking = [prev.thinking, chunk.thinking].filter(Boolean).join('\n\n') || undefined;
      const toolCalls = [...(prev.toolCalls || []), ...(chunk.toolCall ? [chunk.toolCall] : [])];
      return {
        role: 'assistant',
        content,
        ...(thinking ? { thinking } : {}),
        ...(toolCalls.length > 0 ? { toolCalls } : {}),
        createdAt: prev.createdAt || new Date().toISOString(),
      };
    }

    return originMessage || { role: 'assistant', content: '' };
  }
}

function createSSETransformStream(events: SSEProviderEvents): TransformStream<string, SSEChunk> {
  let buffer = '';
  let streamClosed = false;

  const handlePart = (part: string, controller: TransformStreamDefaultController<SSEChunk>) => {
    if (!part.trim()) return;
    let eventName = '';
    const dataLines: string[] = [];
    for (const rawLine of part.split(/\r?\n/)) {
      const line = rawLine.trimEnd();
      if (line.startsWith('event:')) {
        eventName = line.slice(6).trim();
      }
      if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trimStart());
      }
    }
    const data = dataLines.join('\n').trim();
    if (!data) return;
    if (data === '[DONE]') {
      sseLog('done');
      streamClosed = true;
      return;
    }
    try {
      const chunk: SSEChunk = JSON.parse(data);
      sseLog('event', eventName || '(none)', {
        status: chunk.status,
        hasContent: !!chunk.content,
        contentLen: chunk.content?.length || 0,
        hasThinking: !!chunk.thinking,
        thinkingLen: chunk.thinking?.length || 0,
        hasTool: !!chunk.toolCall,
        hasMeta: !!chunk.conversationId || !!chunk.title,
      });
      if (chunk.conversationId && chunk.title) {
        events.onTitleUpdated?.(chunk.conversationId, chunk.title);
      }
      controller.enqueue(chunk);
    } catch {
      sseLog('parse_failed', eventName || '(none)', data.slice(0, 120));
    }
  };

  return new TransformStream<string, SSEChunk>({
    transform(textChunk, controller) {
      if (streamClosed) return;
      sseLog('recv', { len: textChunk.length });
      buffer += textChunk;
      const parts = buffer.split(/\r?\n\r?\n/);
      buffer = parts.pop() || '';
      for (const part of parts) {
        if (streamClosed) return;
        handlePart(part, controller);
      }
    },
    flush(controller) {
      if (streamClosed || !buffer.trim()) return;
      const parts = buffer.split(/\r?\n\r?\n/);
      for (const part of parts) {
        if (streamClosed) return;
        handlePart(part, controller);
      }
    },
  });
}

// Cache providers per agent
const providerCache = new Map<string, SSEChatProvider>();

// Custom fetch that injects the latest JWT on every request
function authFetch(url: Parameters<typeof fetch>[0], options?: RequestInit): Promise<Response> {
  const jwt = localStorage.getItem('casdoor_token');
  const headers = new Headers(options?.headers);
  if (jwt) headers.set('Authorization', `Bearer ${jwt}`);
  headers.set('Content-Type', 'application/json');
  const method = (options?.method || 'GET').toUpperCase();
  const bodyLen = typeof options?.body === 'string' ? options.body.length : undefined;
  sseLog('fetch', { method, url, bodyLen });
  return fetch(url, { ...options, headers })
    .then((res) => {
      sseLog('fetch_res', { status: res.status, ok: res.ok, contentType: res.headers.get('content-type') });
      return res;
    })
    .catch((err) => {
      sseLog('fetch_err', { name: err?.name, message: err?.message });
      const msg = String(err?.message || err || '');
      if (options?.signal?.aborted || err?.name === 'AbortError' || msg.includes('ERR_ABORTED') || msg.toLowerCase().includes('aborted')) {
        sseLog('fetch_abort_ignored');
        return new Response('data: [DONE]\n\n', {
          status: 200,
          headers: { 'Content-Type': 'text/event-stream' },
        });
      }
      throw err;
    });
}

export function getOrCreateProvider(
  agentId: string,
  events: SSEProviderEvents,
): SSEChatProvider {
  let provider = providerCache.get(agentId);
  if (!provider) {
    const request = XRequest(`${BACKEND_URL}/api/chat/completions`, {
      manual: true,
      fetch: authFetch,
      // Reuse the provider, but create a fresh TransformStream for each request.
      transformStream: () => createSSETransformStream(events),
    });
    provider = new SSEChatProvider({ request });
    providerCache.set(agentId, provider);
  }
  return provider;
}

export function resetProvider(agentId: string): void {
  providerCache.delete(agentId);
}
