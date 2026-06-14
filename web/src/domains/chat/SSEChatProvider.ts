import { AbstractChatProvider, XRequest, type XRequestOptions } from '@ant-design/x-sdk';
import { BACKEND_URL } from '../../config';

const CURRENT_ENTERPRISE_STORAGE_KEY = 'dotblue_current_enterprise_id';

export interface ToolCallItem {
  tool: string;
  emoji: string;
  label: string;
  status: string;
}

export interface MessagePart {
  type: 'text' | 'image' | 'file';
  text?: string;
  fileId?: string;
  name?: string;
  mimeType?: string;
  size?: number;
  previewUrl?: string;
  downloadUrl?: string;
  width?: number;
  height?: number;
}

export interface AttachmentItem {
  id?: string;
  fileId: string;
  kind: 'image' | 'file';
  name: string;
  mimeType: string;
  size: number;
  previewUrl?: string;
  downloadUrl?: string;
  width?: number;
  height?: number;
  status?: 'uploaded' | 'processing' | 'failed';
}

export interface SSEChunk {
  content?: string;
  thinking?: string;
  toolCall?: ToolCallItem;
  conversationId?: string;
  title?: string;
  parts?: MessagePart[];
  attachments?: AttachmentItem[];
  status: string;
}

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  parts?: MessagePart[];
  attachments?: AttachmentItem[];
  thinking?: string;
  toolCalls?: ToolCallItem[];
  createdAt?: string;
  status?: 'pending' | 'streaming' | 'done' | 'error';
}

export interface ChatInput {
  content: string;
  agentId: string;
  conversationId: string;
  parts?: MessagePart[];
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
    const parts = requestParams.parts || buildTextParts(requestParams.content || '');
    return {
      role: 'user',
      content: requestParams.content || '',
      parts,
      attachments: partsToAttachments(parts),
      createdAt: new Date().toISOString(),
      status: 'pending',
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
        ...(chunks.flatMap((c) => c.parts || []).length > 0 ? { parts: coalesceParts(chunks, originMessage) } : {}),
        ...(chunks.flatMap((c) => c.attachments || []).length > 0 ? { attachments: coalesceAttachments(chunks, originMessage) } : {}),
        ...(thinking ? { thinking } : {}),
        ...(toolCalls.length > 0 ? { toolCalls: finalizeToolCalls(toolCalls) } : {}),
        createdAt: originMessage?.createdAt || new Date().toISOString(),
        status: 'done',
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
        ...(chunk.parts || prev.parts ? { parts: chunk.parts || prev.parts } : {}),
        ...(chunk.attachments || prev.attachments ? { attachments: chunk.attachments || prev.attachments } : {}),
        ...(thinking ? { thinking } : {}),
        ...(toolCalls.length > 0 ? { toolCalls } : {}),
        createdAt: prev.createdAt || new Date().toISOString(),
        status: chunk.status === 'error' ? 'error' : 'streaming',
      };
    }

    return originMessage || { role: 'assistant', content: '' };
  }
}

function buildTextParts(content: string): MessagePart[] {
  const text = content.trim();
  if (!text) return [];
  return [{ type: 'text', text }];
}

function partsToAttachments(parts: MessagePart[]): AttachmentItem[] {
  return parts
    .filter((part) => !!part.fileId && (part.type === 'image' || part.type === 'file'))
    .map((part) => ({
      fileId: part.fileId!,
      kind: part.type === 'image' ? 'image' : 'file',
      name: part.name || '',
      mimeType: part.mimeType || '',
      size: part.size || 0,
      previewUrl: part.previewUrl,
      downloadUrl: part.downloadUrl,
      width: part.width,
      height: part.height,
      status: 'uploaded',
    }));
}

function coalesceParts(chunks: SSEChunk[], originMessage?: ChatMessage): MessagePart[] | undefined {
  for (let i = chunks.length - 1; i >= 0; i -= 1) {
    if (chunks[i].parts?.length) return chunks[i].parts;
  }
  return originMessage?.parts;
}

function coalesceAttachments(chunks: SSEChunk[], originMessage?: ChatMessage): AttachmentItem[] | undefined {
  for (let i = chunks.length - 1; i >= 0; i -= 1) {
    if (chunks[i].attachments?.length) return chunks[i].attachments;
  }
  return originMessage?.attachments;
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

// Cache providers per request signature so member mode and public mode do not
// accidentally share the same streaming client for the same agent id.
const providerCache = new Map<string, SSEChatProvider>();

function defaultAuthHeaderProvider(): Record<string, string> {
  const jwt = localStorage.getItem('casdoor_token');
  const enterpriseId = localStorage.getItem(CURRENT_ENTERPRISE_STORAGE_KEY)?.trim();
  const headers: Record<string, string> = {};
  if (jwt) headers.Authorization = `Bearer ${jwt}`;
  if (enterpriseId) headers['X-Enterprise-ID'] = enterpriseId;
  return headers;
}

function createAuthFetch(getHeaders?: () => Record<string, string>): typeof fetch {
  return (url: Parameters<typeof fetch>[0], options?: RequestInit): Promise<Response> => {
    const headers = new Headers(options?.headers);
    const latestHeaders = (getHeaders || defaultAuthHeaderProvider)();
    Object.entries(latestHeaders).forEach(([key, value]) => {
      if (value) headers.set(key, value);
    });
    headers.set('Content-Type', 'application/json');
    const method = (options?.method || 'GET').toUpperCase();
    const bodyLen = typeof options?.body === 'string' ? options.body.length : undefined;
    sseLog('fetch', { method, url, bodyLen });
    return fetch(url, { ...options, headers })
      .then(async (res) => {
        sseLog('fetch_res', { status: res.status, ok: res.ok, contentType: res.headers.get('content-type') });
        if (!res.ok) {
          let errMsg = `Request failed with status ${res.status}`;
          try {
            const body = await res.text();
            const parsed = JSON.parse(body);
            if (parsed?.error || parsed?.message) {
              errMsg = parsed.error || parsed.message;
            } else if (body) {
              errMsg = body.slice(0, 200);
            }
          } catch {
            // ignore parse errors, use default message
          }
          sseLog('fetch_http_error', { status: res.status, message: errMsg });
          return new Response(
            `event: error\ndata: ${JSON.stringify({ content: errMsg, status: 'error' })}\n\nevent: message\ndata: [DONE]\n\n`,
            { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
          );
        }
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
  };
}

interface ProviderOptions {
  cacheKey?: string;
  requestUrl?: string;
  getHeaders?: () => Record<string, string>;
}

export function getOrCreateProvider(
  agentId: string,
  events: SSEProviderEvents,
  options: ProviderOptions = {},
): SSEChatProvider {
  const cacheKey = options.cacheKey || agentId;
  let provider = providerCache.get(cacheKey);
  if (!provider) {
    const request = XRequest(options.requestUrl || `${BACKEND_URL}/api/chat/completions`, {
      manual: true,
      fetch: createAuthFetch(options.getHeaders),
      // Reuse the provider, but create a fresh TransformStream for each request.
      transformStream: () => createSSETransformStream(events),
    });
    provider = new SSEChatProvider({ request });
    providerCache.set(cacheKey, provider);
  }
  return provider;
}

export function resetProvider(cacheKey: string): void {
  providerCache.delete(cacheKey);
}
