import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import SpacesList from '../components/SpacesList';
import UsersList from '../components/UsersList';
import ChatArea from '../components/ChatArea';
import { Space } from '../services/spaceService';
import { messageService } from '../services/messageService';
import '../styles/Dashboard.css';

// Define a type for chat target
type ChatTarget = {
  type: 'space' | 'user';
  space?: Space;
  userId?: string;
  username?: string;
};

const Dashboard: React.FC = () => {
  const { user, token, logout } = useAuth();
  const [chatTarget, setChatTarget] = useState<ChatTarget | null>(null);
  const [activeTab, setActiveTab] = useState<'spaces' | 'users'>('spaces');

  // Connect to WebSocket when the dashboard mounts
  useEffect(() => {
    if (token) {
      messageService.connectToWebSocket(token);
    }
    
    // Disconnect when component unmounts
    return () => {
      messageService.disconnectWebSocket();
    };
  }, [token]);

  const handleSelectSpace = (space: Space) => {
    setChatTarget({
      type: 'space',
      space
    });
    setActiveTab('spaces');
  };

  const handleSelectUser = (userId: string, username: string) => {
    setChatTarget({
      type: 'user',
      userId,
      username
    });
    setActiveTab('users');
  };

  return (
    <div className="dashboard">
      <div className="sidebar">
        <div className="user-profile">
          <div className="user-avatar">{user?.username.charAt(0).toUpperCase()}</div>
          <div className="user-name">{user?.username}</div>
          <button onClick={logout} className="logout-button sidebar-logout">Logout</button>
        </div>
        
        <div className="tabs">
          <button 
            className={`tab ${activeTab === 'spaces' ? 'active' : ''}`}
            onClick={() => setActiveTab('spaces')}
          >
            Spaces
          </button>
          <button 
            className={`tab ${activeTab === 'users' ? 'active' : ''}`}
            onClick={() => setActiveTab('users')}
          >
            Direct Messages
          </button>
        </div>
        
        <div className="sidebar-content">
          {activeTab === 'spaces' ? (
            <SpacesList 
              onSelectSpace={handleSelectSpace} 
              selectedSpaceId={chatTarget?.type === 'space' ? chatTarget.space?.id : undefined}
            />
          ) : (
            <UsersList 
              onSelectUser={handleSelectUser} 
              selectedUserId={chatTarget?.type === 'user' ? chatTarget.userId : undefined}
            />
          )}
        </div>
      </div>
      
      <ChatArea chatTarget={chatTarget} />
    </div>
  );
};

export default Dashboard; 