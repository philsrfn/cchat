import React, { useState, useEffect, useRef } from 'react';
import { Message, messageService, SendMessageRequest } from '../services/messageService';
import { Space } from '../services/spaceService';
import { useAuth } from '../contexts/AuthContext';
import '../styles/Chat.css';

interface ChatProps {
  space?: Space;
  recipientId?: string;
  recipientUsername?: string;
}

const Chat: React.FC<ChatProps> = ({ space, recipientId, recipientUsername }) => {
  const { user } = useAuth();
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const isDirect = !space && recipientId;
  const chatName = space ? space.name : recipientUsername || 'Direct Message';

  // Function to fetch messages
  const fetchMessages = async () => {
    if (!space && !recipientId) return;

    setLoading(true);
    setError(null);
    
    try {
      let fetchedMessages: Message[];
      
      if (space) {
        fetchedMessages = await messageService.getSpaceMessages(space.id);
        // Subscribe to space messages via WebSocket
        messageService.subscribeToSpace(space.id);
      } else if (recipientId) {
        fetchedMessages = await messageService.getDirectMessages(recipientId);
      } else {
        fetchedMessages = [];
      }
      
      setMessages(fetchedMessages);
    } catch (err) {
      setError('Failed to load messages');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Load messages when the selected space or recipient changes
  useEffect(() => {
    fetchMessages();

    // Clean up WebSocket subscription when switching spaces
    return () => {
      if (space) {
        messageService.unsubscribeFromSpace(space.id);
      }
    };
  }, [space, recipientId]);

  // Scroll to the bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Set up WebSocket listener for new messages
  useEffect(() => {
    const handleNewMessage = (message: Message) => {
      // Only add message if it's for the current space or direct conversation
      if (space && message.space_id === space.id) {
        setMessages(prev => [...prev, message]);
      } else if (
        isDirect && message.is_direct_message && 
        ((message.sender_id === user?.id && message.recipient_id === recipientId) ||
         (message.sender_id === recipientId && message.recipient_id === user?.id))
      ) {
        setMessages(prev => [...prev, message]);
      }
    };

    messageService.addMessageListener(handleNewMessage);

    return () => {
      messageService.removeMessageListener(handleNewMessage);
    };
  }, [space, recipientId, isDirect, user?.id]);

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!newMessage.trim() || (!space && !recipientId)) {
      return;
    }

    const messageContent = newMessage.trim();
    setNewMessage(''); // Clear input immediately for better UX

    const messageData: SendMessageRequest = {
      content: messageContent,
      ...(space ? { space_id: space.id } : {}),
      ...(recipientId ? { recipient_id: recipientId } : {})
    };

    try {
      // Optimistically add the message to the UI before API response
      const optimisticMessage: Message = {
        id: `temp-${Date.now()}`,
        content: messageContent,
        sender_id: user?.id || '',
        sender_username: user?.username || '',
        space_id: space?.id,
        recipient_id: recipientId,
        is_direct_message: !space,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        is_edited: false
      };
      
      setMessages(prev => [...prev, optimisticMessage]);
      
      // Send to server in background
      await messageService.sendMessage(messageData);
      // No need to refetch all messages - the websocket will handle real updates
    } catch (err) {
      setError('Failed to send message');
      console.error(err);
    }
  };

  if (!space && !recipientId) {
    return (
      <div className="chat-container empty-chat">
        <p>Select a space or user to start chatting</p>
      </div>
    );
  }

  return (
    <div className="chat-container">
      <div className="chat-header">
        <h2>{chatName}</h2>
      </div>
      
      {error && <div className="chat-error">{error}</div>}
      
      <div className="messages-container">
        {loading ? (
          <div className="loading-messages">Loading messages...</div>
        ) : messages.length === 0 ? (
          <div className="no-messages">No messages yet. Start the conversation!</div>
        ) : (
          <div className="messages-list">
            {messages.map(message => (
              <div 
                key={message.id} 
                className={`message ${message.sender_id === user?.id ? 'own-message' : 'other-message'}`}
              >
                <div className="message-header">
                  <span className="message-sender">{message.sender_username || 'Unknown'}</span>
                  <span className="message-time">
                    {new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
                <div className="message-content">{message.content}</div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>
      
      <form className="message-form" onSubmit={handleSendMessage}>
        <input
          type="text"
          value={newMessage}
          onChange={e => setNewMessage(e.target.value)}
          placeholder="Type a message..."
          disabled={loading}
        />
        <button type="submit" disabled={!newMessage.trim() || loading}>Send</button>
      </form>
    </div>
  );
}

export default Chat; 