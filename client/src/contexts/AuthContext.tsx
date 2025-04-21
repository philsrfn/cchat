import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import AuthService, { UserResponse, LoginRequest, RegisterRequest } from '../services/auth';
import { messageService } from '../services/messageService';

interface AuthContextType {
  user: UserResponse | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => void;
  error: string | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// Helper function to get cookie
const getCookie = (name: string): string | null => {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(';').shift() || null;
  return null;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<UserResponse | null>(null);
  const [token, setToken] = useState<string | null>(getCookie('token'));
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  // Ensure WebSocket connection is established and maintained
  const ensureWebSocketConnection = (token: string | null) => {
    if (token) {
      console.log("Ensuring WebSocket connection with token");
      // Connect to WebSocket with token
      messageService.connectToWebSocket(token);
    } else {
      console.log("No token available, disconnecting WebSocket");
      // Disconnect from WebSocket if no token
      messageService.disconnectWebSocket();
    }
  };

  // Connect to WebSocket when token changes
  useEffect(() => {
    console.log("Token changed, updating WebSocket connection");
    ensureWebSocketConnection(token);

    // Set up ping interval to keep connection alive
    const pingInterval = setInterval(() => {
      if (token) {
        console.log("Checking WebSocket connection...");
        ensureWebSocketConnection(token);
      }
    }, 30000); // Check every 30 seconds

    return () => {
      // Clean up WebSocket connection and ping interval when component unmounts
      clearInterval(pingInterval);
      messageService.disconnectWebSocket();
    };
  }, [token]);

  // Check if user is already logged in on initial load
  useEffect(() => {
    const initializeAuth = async () => {
      setIsLoading(true);
      try {
        // Try to validate the current session with the server
        console.log("Validating session...");
        const validatedUser = await AuthService.validateSession();
        if (validatedUser) {
          console.log("Session valid, user:", validatedUser);
          setUser(validatedUser);
          const currentToken = getCookie('token');
          setToken(currentToken);
        } else {
          console.log("Session invalid, logging out");
          // Clear any invalid session
          AuthService.logout();
          setToken(null);
        }
      } catch (error) {
        console.error("Authentication error:", error);
        setError(error instanceof Error ? error.message : 'Authentication error');
        // Clear any invalid session
        AuthService.logout();
        setToken(null);
      } finally {
        setIsLoading(false);
      }
    };

    initializeAuth();
  }, []);

  const login = async (credentials: LoginRequest) => {
    try {
      setIsLoading(true);
      setError(null);
      console.log("Logging in...");
      const loggedInUser = await AuthService.login(credentials);
      console.log("Login successful, user:", loggedInUser);
      setUser(loggedInUser);
      const currentToken = getCookie('token');
      console.log("Setting token:", currentToken?.substring(0, 10) + "...");
      setToken(currentToken);
      navigate('/'); // Redirect to home page after login
    } catch (err) {
      console.error("Login error:", err);
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (userData: RegisterRequest) => {
    try {
      setIsLoading(true);
      setError(null);
      await AuthService.register(userData);
      navigate('/login'); // Redirect to login page after registration
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    console.log("Logging out");
    AuthService.logout();
    setUser(null);
    setToken(null);
    navigate('/login');
  };

  const value = {
    user,
    token,
    isAuthenticated: !!user,
    isLoading,
    login,
    register,
    logout,
    error
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthProvider; 