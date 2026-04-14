import { Navigate } from 'react-router-dom';
import { useState, useEffect } from 'react';

export const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const [isAuth, setIsAuth] = useState(
    !!localStorage.getItem('taskflow_token')
  );

  useEffect(() => {
    const handleLogout = () => setIsAuth(false);
    window.addEventListener('unauthorized', handleLogout);
    return () => window.removeEventListener('unauthorized', handleLogout);
  }, []);

  if (!isAuth) return <Navigate to="/login" replace />;
  return <>{children}</>;
};
