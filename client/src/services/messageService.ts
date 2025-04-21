import { apiClient } from './apiClient';

export interface Message {
  id: string;
  content: string;
  sender_id: string;
  sender_username?: string;
  space_id?: string;
  recipient_id?: string;
  is_direct_message: boolean;
  created_at: string;
  updated_at: string;
  is_edited: boolean;
}

export interface SendMessageRequest {
  content: string;
  space_id?: string;
  recipient_id?: string;
}

// WebSocket connection for real-time messaging
let socket: WebSocket | null = null;
let messageListeners: ((message: Message) => void)[] = [];
let reconnectTimer: number | null = null;
let isConnecting = false;

// Message services
export const messageService = {
  // Get messages for a space
  getSpaceMessages: async (spaceId: string): Promise<Message[]> => {
    const response = await apiClient.get(`/messages?space_id=${spaceId}`);
    return response.data;
  },

  // Get direct messages between the current user and another user
  getDirectMessages: async (recipientId: string): Promise<Message[]> => {
    const response = await apiClient.get(`/messages?recipient_id=${recipientId}`);
    return response.data;
  },

  // Send a message to a space or user
  sendMessage: async (data: SendMessageRequest): Promise<Message> => {
    const response = await apiClient.post('/messages', data);
    return response.data;
  },

  // Subscribe to messages via WebSocket
  connectToWebSocket: (token: string): void => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      console.log("WebSocket already connected");
      return;
    }
    
    if (isConnecting) {
      console.log("WebSocket connection already in progress");
      return;
    }
    
    isConnecting = true;
    
    if (socket) {
      socket.close();
      socket = null;
    }

    console.log("Connecting to WebSocket...");
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host || 'localhost:8080';
    const wsUrl = `${protocol}//${host}/ws?token=${token}`;
    
    console.log(`WebSocket URL: ${wsUrl}`);
    socket = new WebSocket(wsUrl);
    
    socket.onopen = () => {
      console.log("WebSocket connection established");
      isConnecting = false;
    };
    
    socket.onmessage = (event) => {
      try {
        console.log("WebSocket message received:", event.data);
        const message = JSON.parse(event.data) as Message;
        console.log("Parsed message:", message);
        messageListeners.forEach(listener => {
          console.log("Notifying listener about new message");
          listener(message);
        });
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };
    
    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      isConnecting = false;
    };
    
    socket.onclose = (event) => {
      console.log(`WebSocket connection closed: ${event.code} ${event.reason}`);
      isConnecting = false;
      
      // Try to reconnect after a delay
      if (reconnectTimer === null) {
        console.log("Scheduling WebSocket reconnection...");
        reconnectTimer = window.setTimeout(() => {
          reconnectTimer = null;
          console.log("Attempting to reconnect WebSocket...");
          messageService.connectToWebSocket(token);
        }, 3000);
      }
    };
  },
  
  // Disconnect WebSocket
  disconnectWebSocket: (): void => {
    console.log("Disconnecting WebSocket");
    
    if (socket) {
      socket.close();
      socket = null;
    }
    
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    
    isConnecting = false;
    messageListeners = [];
  },
  
  // Subscribe to a space (to receive its messages)
  subscribeToSpace: (spaceId: string): void => {
    console.log(`Subscribing to space: ${spaceId}`);
    
    if (socket && socket.readyState === WebSocket.OPEN) {
      const subscribeMsg = JSON.stringify({
        type: 'subscribe',
        space_id: spaceId,
        subscribe: true
      });
      console.log(`Sending subscription message: ${subscribeMsg}`);
      socket.send(subscribeMsg);
    } else {
      console.warn(`Cannot subscribe to space ${spaceId} - WebSocket not open. Current state: ${socket ? socket.readyState : 'no socket'}`);
    }
  },
  
  // Unsubscribe from a space
  unsubscribeFromSpace: (spaceId: string): void => {
    console.log(`Unsubscribing from space: ${spaceId}`);
    
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: 'subscribe',
        space_id: spaceId,
        subscribe: false
      }));
    }
  },
  
  // Add a listener for incoming messages
  addMessageListener: (listener: (message: Message) => void): void => {
    console.log("Adding message listener");
    messageListeners.push(listener);
  },
  
  // Remove a message listener
  removeMessageListener: (listener: (message: Message) => void): void => {
    console.log("Removing message listener");
    messageListeners = messageListeners.filter(l => l !== listener);
  }
}; 