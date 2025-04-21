import { apiClient } from './apiClient';

export interface Space {
  id: string;
  name: string;
  description: string;
  creator_id: string;
  is_public: boolean;
  created_at: string;
  member_count?: number;
}

export interface CreateSpaceRequest {
  name: string;
  description: string;
  is_public: boolean;
}

// Space services
export const spaceService = {
  // Get all spaces for the current user
  getSpaces: async (): Promise<Space[]> => {
    console.log('Fetching spaces...');
    try {
      const response = await apiClient.get('/spaces');
      console.log('Spaces fetched successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch spaces:', error);
      throw error;
    }
  },

  // Get a specific space by ID
  getSpace: async (id: string): Promise<Space> => {
    console.log(`Fetching space with ID: ${id}`);
    try {
      const response = await apiClient.get(`/spaces/${id}`);
      console.log('Space fetched successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error(`Failed to fetch space with ID ${id}:`, error);
      throw error;
    }
  },

  // Create a new space
  createSpace: async (data: CreateSpaceRequest): Promise<Space> => {
    console.log('Creating space with data:', data);
    try {
      const response = await apiClient.post('/spaces', data);
      console.log('Space created successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('Failed to create space:', error);
      throw error;
    }
  },

  // Join a public space
  joinSpace: async (id: string): Promise<void> => {
    console.log(`Joining space with ID: ${id}`);
    try {
      await apiClient.post(`/spaces/${id}/join`);
      console.log('Joined space successfully');
    } catch (error) {
      console.error(`Failed to join space with ID ${id}:`, error);
      throw error;
    }
  },

  // Invite a user to a space
  inviteToSpace: async (id: string, email: string): Promise<void> => {
    console.log(`Inviting user ${email} to space with ID: ${id}`);
    try {
      await apiClient.post(`/spaces/${id}/invite`, { email });
      console.log('User invited successfully');
    } catch (error) {
      console.error(`Failed to invite user to space with ID ${id}:`, error);
      throw error;
    }
  }
}; 