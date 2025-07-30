import { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuthStore } from './stores/useAuthStore';
import Dashboard from './components/Dashboard';
import AuthPage from './components/AuthPage';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  const { isAuthenticated, getCurrentUser, user } = useAuthStore();

  useEffect(() => {
    // 앱 시작 시 현재 사용자 정보 가져오기
    if (isAuthenticated && !user) {
      getCurrentUser();
    }
  }, [isAuthenticated, user, getCurrentUser]);

  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen bg-background">
        {isAuthenticated ? <Dashboard /> : <AuthPage />}
      </div>
    </QueryClientProvider>
  );
}

export default App;
