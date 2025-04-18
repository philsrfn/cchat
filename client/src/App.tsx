import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import AuthProvider, { useAuth } from './contexts/AuthContext';
import Login from './pages/Login';
import Register from './pages/Register';
import './styles/App.css';

// Placeholder components - these will be implemented later
const Home = () => {
  const { user, logout } = useAuth();
  
  return (
    <div className="page">
      <h1>Home Page</h1>
      <p>Welcome, {user?.username}!</p>
      <button onClick={logout} className="logout-button">Logout</button>
    </div>
  );
};

const NotFound = () => <div className="page">404 - Not Found</div>;

// Protected route wrapper
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <div className="loading">Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return <>{children}</>;
};

// Public route wrapper (redirect if already authenticated)
const PublicRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <div className="loading">Loading...</div>;
  }

  if (isAuthenticated) {
    return <Navigate to="/" />;
  }

  return <>{children}</>;
};

const AppRoutes = () => {
  return (
    <Routes>
      <Route path="/" element={
        <ProtectedRoute>
          <Home />
        </ProtectedRoute>
      } />
      <Route path="/login" element={
        <PublicRoute>
          <Login />
        </PublicRoute>
      } />
      <Route path="/register" element={
        <PublicRoute>
          <Register />
        </PublicRoute>
      } />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <div className="app">
        <header className="app-header">
          <h1>CChat</h1>
        </header>
        <main className="app-content">
          <AppRoutes />
        </main>
        <footer className="app-footer">
          <p>&copy; {new Date().getFullYear()} CChat</p>
        </footer>
      </div>
    </AuthProvider>
  );
};

export default App; 