import axios from 'axios';

// Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface UserResponse {
  id: string;
  username: string;
  email: string;
  created_at: string;
}

export interface LoginResponse {
  token: string;
  user: UserResponse;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

// Helper to parse user cookie
const getUserFromCookie = (): UserResponse | null => {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; user=`);
  if (parts.length === 2) {
    const userCookie = parts.pop()?.split(';').shift();
    if (userCookie) {
      try {
        return JSON.parse(decodeURIComponent(userCookie));
      } catch (e) {
        console.error("Error parsing user cookie:", e);
      }
    }
  }
  return null;
};

// Auth service
const AuthService = {
  // Get current user
  getCurrentUser: (): UserResponse | null => {
    return getUserFromCookie();
  },

  // Check if user is authenticated
  isAuthenticated: (): boolean => {
    return !!AuthService.getCurrentUser();
  },

  // Login
  login: async (credentials: LoginRequest): Promise<UserResponse> => {
    try {
      const response = await axios.post<any>('/api/auth/login', credentials, {
        withCredentials: true
      });
      
      // Return the user data from the response
      return response.data.user;
    } catch (error) {
      if (axios.isAxiosError(error) && error.response) {
        throw new Error(error.response.data.error || 'Login failed');
      }
      throw new Error('Network error occurred');
    }
  },

  // Register
  register: async (userData: RegisterRequest): Promise<UserResponse> => {
    try {
      const response = await axios.post<any>('/api/auth/register', userData, {
        withCredentials: true
      });
      
      return response.data.user;
    } catch (error) {
      if (axios.isAxiosError(error) && error.response) {
        throw new Error(error.response.data.error || 'Registration failed');
      }
      throw new Error('Network error occurred');
    }
  },

  // Logout
  logout: async (): Promise<void> => {
    try {
      // Call the server logout endpoint to clear server-side cookie
      await axios.post('/api/auth/logout', {}, { withCredentials: true });
    } catch (error) {
      console.error('Error during logout:', error);
    }
  },

  // Validate the current session
  validateSession: async (): Promise<UserResponse | null> => {
    try {
      const response = await axios.get<any>('/api/auth/validate', {
        withCredentials: true
      });
      
      // Updated to match the new backend response format
      if (response.data && response.data.authenticated && response.data.user) {
        return response.data.user;
      }
      return null;
    } catch (error) {
      console.error('Session validation failed:', error);
      return null;
    }
  }
};

// Configure axios to always include credentials
axios.defaults.withCredentials = true;

export default AuthService; 