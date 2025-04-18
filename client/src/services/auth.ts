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

// Cookie helper functions
const setCookie = (name: string, value: string, days = 7) => {
  const date = new Date();
  date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
  const expires = `expires=${date.toUTCString()}`;
  document.cookie = `${name}=${value};${expires};path=/;SameSite=Strict`;
};

const getCookie = (name: string): string | null => {
  const cookies = document.cookie.split(';');
  for (let i = 0; i < cookies.length; i++) {
    const cookie = cookies[i].trim();
    if (cookie.startsWith(name + '=')) {
      return cookie.substring(name.length + 1);
    }
  }
  return null;
};

const removeCookie = (name: string) => {
  document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
};

// Auth service
const AuthService = {
  // Get current auth token
  getToken: (): string | null => {
    return getCookie('auth_token');
  },

  // Get current user from cookie
  getCurrentUser: (): UserResponse | null => {
    const userJson = getCookie('user');
    return userJson ? JSON.parse(userJson) : null;
  },

  // Check if user is authenticated
  isAuthenticated: (): boolean => {
    return !!AuthService.getToken();
  },

  // Login
  login: async (credentials: LoginRequest): Promise<UserResponse> => {
    try {
      const response = await axios.post<ApiResponse<LoginResponse>>('/api/auth/login', credentials);
      
      if (response.data.success && response.data.data) {
        // Save auth token and user info in cookies
        setCookie('auth_token', response.data.data.token);
        setCookie('user', JSON.stringify(response.data.data.user));
        
        // Set default Authorization header for future requests
        axios.defaults.headers.common['Authorization'] = `Bearer ${response.data.data.token}`;
        
        return response.data.data.user;
      } else {
        throw new Error(response.data.error || 'Login failed');
      }
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
      const response = await axios.post<ApiResponse<UserResponse>>('/api/auth/register', userData);
      
      if (response.data.success && response.data.data) {
        return response.data.data;
      } else {
        throw new Error(response.data.error || 'Registration failed');
      }
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
      await axios.post('/api/auth/logout');
    } catch (error) {
      console.error('Error during logout:', error);
    } finally {
      // Remove cookies even if the server call fails
      removeCookie('auth_token');
      removeCookie('user');
      
      // Remove Authorization header
      delete axios.defaults.headers.common['Authorization'];
    }
  },

  // Validate the current session
  validateSession: async (): Promise<UserResponse | null> => {
    try {
      const response = await axios.get<ApiResponse<UserResponse>>('/api/auth/validate');
      
      if (response.data.success && response.data.data) {
        // Update the user cookie with the latest info
        setCookie('user', JSON.stringify(response.data.data));
        return response.data.data;
      }
      return null;
    } catch (error) {
      // If validation fails, clear the cookies and headers
      AuthService.logout();
      return null;
    }
  }
};

// Set auth header if token exists on app initialization
const token = AuthService.getToken();
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
}

export default AuthService; 