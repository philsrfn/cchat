import { apiClient } from './apiClient';
import { UserResponse } from './auth';

// User services
export const userService = {
  // Get all users
  getUsers: async (): Promise<UserResponse[]> => {
    console.log('Fetching users...');
    try {
      const response = await apiClient.get('/users');
      console.log('Users fetched successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch users:', error);
      throw error;
    }
  },

  // Get the current user profile
  getCurrentUser: async (): Promise<UserResponse> => {
    console.log('Fetching current user profile...');
    try {
      const response = await apiClient.get('/users/profile');
      console.log('User profile fetched successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch user profile:', error);
      throw error;
    }
  }
}; 