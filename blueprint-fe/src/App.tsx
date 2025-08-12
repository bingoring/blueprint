import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App as AntApp } from "antd";
import { useEffect } from "react";
import {
  Navigate,
  Route,
  BrowserRouter as Router,
  Routes,
} from "react-router-dom";

import { ThemeProvider } from "./contexts/ThemeContext";
import { apiClient } from "./lib/api";
import { useAuthStore } from "./stores/useAuthStore";

// Components
import AccountSettingsPage from "./components/AccountSettingsPage";
import CreateProjectPage from "./components/CreateProjectPage";
import EditProjectPage from "./components/EditProjectPage";
import NewDashboard from "./components/NewDashboard";
import NewHomePage from "./components/NewHomePage"; // 실제로는 ProjectExplorePage
import ProfilePage from "./components/ProfilePage";
import ProjectDetailPage from "./components/ProjectDetailPage";
import PolymarketTradingPage from "./pages/PolymarketTradingPage";

// CSS imports
import "./index.css";
import "./styles/polymarket.css";

// HomePage alias
const HomePage = NewHomePage;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  const { isAuthenticated, getCurrentUser, user } = useAuthStore();

  // Google OAuth 콜백 처리 (URL 파라미터에서 토큰 확인)
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get("token");
    const userId = urlParams.get("user_id");

    if (token && userId && !isAuthenticated) {
      console.log("🔑 Google OAuth 토큰 발견, 로그인 처리 중...");

      // apiClient에 토큰 설정
      apiClient.setToken(token);

      // 사용자 정보 가져오기
      getCurrentUser()
        .then(() => {
          console.log("✅ Google OAuth 로그인 성공");
          // URL에서 토큰 파라미터 제거
          const newUrl = window.location.origin + window.location.pathname;
          window.history.replaceState({}, document.title, newUrl);
        })
        .catch((error) => {
          console.error("❌ Google OAuth 로그인 실패:", error);
          // 토큰 제거
          apiClient.clearToken();
        });
    }
  }, []); // 의존성 배열을 비워서 최초 1회만 실행

  return (
    <QueryClientProvider client={queryClient}>
      <AntApp>
        <ThemeProvider>
          <Router>
            <Routes>
              <Route
                path="/"
                element={
                  isAuthenticated && user ? (
                    <Navigate to="/dashboard" replace />
                  ) : (
                    <HomePage />
                  )
                }
              />
              <Route path="/explore" element={<HomePage />} />
              <Route path="/create-project" element={<CreateProjectPage />} />
              <Route path="/edit-project/:id" element={<EditProjectPage />} />
              <Route path="/project/:id" element={<ProjectDetailPage />} />
              <Route
                path="/trade/:projectId/:milestoneId"
                element={<PolymarketTradingPage />}
              />
              <Route path="/dashboard" element={<NewDashboard />} />
              <Route path="/profile/:username" element={<ProfilePage />} />
              <Route path="/settings" element={<AccountSettingsPage />} />
            </Routes>
          </Router>
        </ThemeProvider>
      </AntApp>
    </QueryClientProvider>
  );
}

export default App;
