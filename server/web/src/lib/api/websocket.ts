import { type SearchResults } from './endpoints/schemas';

interface WebSocketConfig {
  reconnectInterval?: number;
  maxReconnectInterval?: number;
  reconnectDecay?: number;
  maxReconnectAttempts?: number;
  timeoutInterval?: number;
}

const DEFAULT_CONFIG: Required<WebSocketConfig> = {
  reconnectInterval: 1000,
  maxReconnectInterval: 30000,
  reconnectDecay: 1.5,
  maxReconnectAttempts: 10,
  timeoutInterval: 2000,
};

interface PendingMessage {
  query: { text: string; [key: string]: unknown };
  callback: (results: SearchResults) => void;
}

class ResilientWebSocket {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketConfig>;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private timeoutTimer: ReturnType<typeof setTimeout> | null = null;
  private forcedClose = false;
  private url: string;

  private callbacks = new Map<string, (results: SearchResults) => void>();
  private pendingMessages: PendingMessage[] = [];

  constructor(url: string, config: WebSocketConfig = {}) {
    this.url = url;
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  connect(): WebSocket {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return this.ws;
    }

    if (this.ws?.readyState === WebSocket.CONNECTING) {
      return this.ws;
    }

    this.ws = new WebSocket(this.url);

    this.timeoutTimer = setTimeout(() => {
      if (this.ws?.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
    }, this.config.timeoutInterval);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.clearReconnectTimer();
      this.clearTimeoutTimer();
      this.reconnectAttempts = 0;
      this.flushPendingMessages();
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as SearchResults;
        const queryText = (data.query as { text: string }).text;
        const callback = this.callbacks.get(queryText);
        if (callback) {
          callback(data);
          this.callbacks.delete(queryText);
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = (event) => {
      console.log('WebSocket disconnected', event.code, event.reason);
      this.ws = null;
      this.clearTimeoutTimer();

      if (!this.forcedClose) {
        this.scheduleReconnect();
      }
    };

    return this.ws;
  }

  private flushPendingMessages(): void {
    while (this.pendingMessages.length > 0) {
      const msg = this.pendingMessages.shift();
      if (msg) {
        this.send(msg.query, msg.callback);
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;

    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    const delay = Math.min(
      this.config.reconnectInterval * Math.pow(this.config.reconnectDecay, this.reconnectAttempts),
      this.config.maxReconnectInterval
    );

    this.reconnectAttempts++;
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private clearTimeoutTimer(): void {
    if (this.timeoutTimer) {
      clearTimeout(this.timeoutTimer);
      this.timeoutTimer = null;
    }
  }

  send(
    query: { text: string; [key: string]: unknown },
    callback: (results: SearchResults) => void
  ): void {
    this.callbacks.set(query.text, callback);

    const socket = this.connect();

    if (socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(query));
    } else if (socket.readyState === WebSocket.CONNECTING) {
      this.pendingMessages.push({ query, callback });

      const originalOnOpen = socket.onopen;
      socket.onopen = (event) => {
        originalOnOpen?.call(socket, event);
        socket.send(JSON.stringify(query));
      };
    } else {
      this.pendingMessages.push({ query, callback });
    }
  }

  close(): void {
    this.forcedClose = true;
    this.clearReconnectTimer();
    this.clearTimeoutTimer();
    this.ws?.close();
    this.ws = null;
    this.callbacks.clear();
    this.pendingMessages = [];
  }

  get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }
}

let resilientWs: ResilientWebSocket | null = null;

export function initWebSocket(): ResilientWebSocket {
  if (!resilientWs) {
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/search`;
    resilientWs = new ResilientWebSocket(wsUrl);
  }
  return resilientWs;
}

export function search(
  query: { text: string; [key: string]: unknown },
  callback: (results: SearchResults) => void
): void {
  const ws = initWebSocket();
  ws.send(query, callback);
}

export function closeWebSocket(): void {
  resilientWs?.close();
  resilientWs = null;
}
