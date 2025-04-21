import React, { useState, useEffect } from 'react';
import { userService } from '../services/userService';
import { useAuth } from '../contexts/AuthContext';
import '../styles/UsersList.css';
import { UserResponse } from '../services/auth';

// Add back the props interface
interface UsersListProps {
  onSelectUser: (userId: string, username: string) => void;
  selectedUserId?: string;
}

// Update the component to accept props
const UsersList: React.FC<UsersListProps> = ({ onSelectUser, selectedUserId }) => {
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { user: currentUser } = useAuth();

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        setLoading(true);
        const usersData = await userService.getUsers();
        setUsers(usersData);
        setError(null);
      } catch (err) {
        console.error('Error fetching users:', err);
        setError('Failed to load users');
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  if (loading) {
    return <div className="users-list-loading">Loading users...</div>;
  }

  if (error) {
    return <div className="users-list-error">{error}</div>;
  }

  return (
    <div className="users-list-container">
      <div className="users-list-header">
        <h2>Direct Messages</h2>
      </div>
      {users.length === 0 ? (
        <div className="no-users">No users available</div>
      ) : (
        <ul className="users-list">
          {users.map(user => (
            <li 
              key={user.id} 
              className={user.id === selectedUserId ? 'user-item selected' : 'user-item'}
              onClick={() => onSelectUser(user.id, user.username)}
            >
              <div className="user-avatar">{user.username.charAt(0).toUpperCase()}</div>
              <div className="user-info">
                <span className="user-name">{user.username}</span>
                <span className="user-email">{user.email}</span>
                {user.id === currentUser?.id && (
                  <span className="current-user-badge">You</span>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default UsersList; 