import React from 'react';
import Chat from './Chat';
import { Space } from '../services/spaceService';

type ChatTarget = {
  type: 'space' | 'user';
  space?: Space;
  userId?: string;
  username?: string;
};

interface ChatAreaProps {
  chatTarget: ChatTarget | null;
}

const ChatArea: React.FC<ChatAreaProps> = ({ chatTarget }) => {
  return (
    <div className="chat-area">
      {chatTarget ? (
        <Chat 
          space={chatTarget.type === 'space' ? chatTarget.space : undefined}
          recipientId={chatTarget.type === 'user' ? chatTarget.userId : undefined}
          recipientUsername={chatTarget.type === 'user' ? chatTarget.username : undefined}
        />
      ) : (
        <div className="welcome-message">
          <h2>Welcome to GoText!</h2>
          <p>Select a space or user to start chatting</p>
        </div>
      )}
    </div>
  );
};

export default ChatArea; 