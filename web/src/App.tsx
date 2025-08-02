import { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { useAuthStore } from './stores/useAuthStore';
import { apiClient } from './lib/api';
import HomePage from './components/NewHomePage';
import CreateProjectPage from './components/CreateProjectPage';
import EditProjectPage from './components/EditProjectPage';
import ProjectDetailPage from './components/ProjectDetailPage';
import NewDashboard from './components/NewDashboard';

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

    // URL 파라미터에서 Google 로그인 토큰 확인
  useEffect(() => {
    console.log('🔍 URL 파라미터 확인:', window.location.search);
    console.log('🔍 현재 인증 상태:', { isAuthenticated, user: !!user });

    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    const userId = urlParams.get('user_id');

    console.log('🔍 추출된 파라미터:', {
      hasToken: !!token,
      hasUserId: !!userId,
      isAuthenticated
    });

    if (token && userId && !isAuthenticated) {
      console.log('🔑 Google 로그인 토큰 발견:', {
        token: token.substring(0, 20) + '...',
        userId,
        tokenLength: token.length
      });

      // apiClient에 토큰 설정
      console.log('⚙️ apiClient에 토큰 설정 중...');
      apiClient.setToken(token);

      // 사용자 정보 가져오기
      console.log('👤 사용자 정보 로드 중...');
      getCurrentUser().then(() => {
        console.log('✅ getCurrentUser 완료');
      }).catch((error) => {
        console.error('❌ getCurrentUser 실패:', error);
      });

      // URL에서 토큰 파라미터 제거
      const newUrl = window.location.origin + window.location.pathname;
      window.history.replaceState({}, document.title, newUrl);
      console.log('🧹 URL 파라미터 제거 완료');
    } else if (token && userId && isAuthenticated) {
      console.log('⚠️ 토큰은 있지만 이미 인증됨 - URL 정리만 수행');
      const newUrl = window.location.origin + window.location.pathname;
      window.history.replaceState({}, document.title, newUrl);
    }
  }, [getCurrentUser, isAuthenticated]);

  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/create-project" element={<CreateProjectPage />} />
          <Route path="/edit-project/:id" element={<EditProjectPage />} />
          <Route path="/project/:id" element={<ProjectDetailPage />} />
          <Route path="/dashboard" element={<NewDashboard />} />
        </Routes>
      </Router>
    </QueryClientProvider>
  );
}

export default App;
