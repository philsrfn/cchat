import React, { useEffect, useState } from 'react';
import { Space, spaceService } from '../services/spaceService';
import '../styles/SpacesList.css';

interface SpacesListProps {
  onSelectSpace: (space: Space) => void;
  selectedSpaceId?: string;
}

const SpacesList: React.FC<SpacesListProps> = ({ onSelectSpace, selectedSpaceId }) => {
  const [spaces, setSpaces] = useState<Space[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreatingSpace, setIsCreatingSpace] = useState(false);
  const [newSpace, setNewSpace] = useState({ name: '', description: '', is_public: true });

  // Fetch spaces on component mount
  useEffect(() => {
    const fetchSpaces = async () => {
      try {
        setLoading(true);
        const spacesData = await spaceService.getSpaces();
        setSpaces(spacesData);
        setError(null);

        // If there's at least one space and none selected, select the first one
        if (spacesData.length > 0 && !selectedSpaceId) {
          onSelectSpace(spacesData[0]);
        }
      } catch (err) {
        setError('Failed to load spaces');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchSpaces();
  }, [onSelectSpace, selectedSpaceId]);

  const handleCreateSpace = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const createdSpace = await spaceService.createSpace(newSpace);
      setSpaces(prev => [...prev, createdSpace]);
      setIsCreatingSpace(false);
      setNewSpace({ name: '', description: '', is_public: true });
      onSelectSpace(createdSpace);
    } catch (err) {
      setError('Failed to create space');
      console.error(err);
    }
  };

  // Commenting out handleJoinSpace as it's not currently used but will be needed in the future
  /* 
  const handleJoinSpace = async (spaceId: string) => {
    try {
      await spaceService.joinSpace(spaceId);
      // Refresh spaces list
      const spacesData = await spaceService.getSpaces();
      setSpaces(spacesData);
      
      // Select the joined space
      const joinedSpace = spacesData.find(s => s.id === spaceId);
      if (joinedSpace) {
        onSelectSpace(joinedSpace);
      }
    } catch (err) {
      setError('Failed to join space');
      console.error(err);
    }
  };
  */

  if (loading) {
    return <div className="spaces-list-loading">Loading spaces...</div>;
  }

  return (
    <div className="spaces-list-container">
      <div className="spaces-list-header">
        <h2>Spaces</h2>
        <button 
          className="create-space-button"
          onClick={() => setIsCreatingSpace(!isCreatingSpace)}
        >
          {isCreatingSpace ? 'Cancel' : 'Create Space'}
        </button>
      </div>

      {error && <div className="spaces-list-error">{error}</div>}

      {isCreatingSpace && (
        <form className="create-space-form" onSubmit={handleCreateSpace}>
          <input
            type="text"
            placeholder="Space name"
            value={newSpace.name}
            onChange={e => setNewSpace({...newSpace, name: e.target.value})}
            required
            minLength={3}
            maxLength={100}
          />
          <textarea
            placeholder="Description"
            value={newSpace.description}
            onChange={e => setNewSpace({...newSpace, description: e.target.value})}
            rows={3}
          />
          <div className="space-privacy">
            <label>
              <input
                type="checkbox"
                checked={newSpace.is_public}
                onChange={e => setNewSpace({...newSpace, is_public: e.target.checked})}
              />
              Public space
            </label>
          </div>
          <button type="submit">Create</button>
        </form>
      )}

      <div className="spaces-list">
        {spaces.length === 0 ? (
          <div className="no-spaces">No spaces available</div>
        ) : (
          <ul>
            {spaces.map(space => (
              <li 
                key={space.id} 
                className={space.id === selectedSpaceId ? 'selected' : ''}
                onClick={() => onSelectSpace(space)}
              >
                <div className="space-item">
                  <div className="space-details">
                    <div className="space-name">{space.name}</div>
                    <div className="space-meta">
                      <span className="space-members">{space.member_count || 0} members</span>
                      {space.is_public && <span className="space-public">Public</span>}
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

export default SpacesList; 