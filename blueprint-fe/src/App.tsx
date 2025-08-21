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

// New Blueprint Components
import AccountSettingsPage from "./components/AccountSettingsPage";
import CreateProjectPage from "./components/CreateProjectPage";
import EditProjectPage from "./components/EditProjectPage";
import ExplorePage from "./components/ExplorePage";
import NewDashboardPage from "./components/NewDashboardPage";
import ProfilePage from "./components/ProfilePage";
import ProjectDetailPage from "./components/ProjectDetailPage";

// Legacy Components (for compatibility)
import NewHomePage from "./components/NewHomePage"; // Landing page for non-authenticated users
import PolymarketTradingPage from "./pages/PolymarketTradingPage";

// CSS imports
import "./index.css";
import "./styles/polymarket.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

// Import ActivityPage
import ActivityPage from "./components/ActivityPage";
import GovernancePage from "./components/GovernancePage";
import MentoringPage from "./components/MentoringPage";

import HallOfFamePage from "./components/HallOfFamePage";

const NotificationsPage = () => (
  <div style={{ padding: "100px", textAlign: "center" }}>
    <h2>알림</h2>
    <p>곧 출시 예정입니다!</p>
  </div>
);

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
              {/* 홈 라우트 - 인증 상태에 따라 분기 */}
              <Route
                path="/"
                element={
                  isAuthenticated && user ? (
                    <NewDashboardPage />
                  ) : (
                    <NewHomePage />
                  )
                }
              />

              {/* Blueprint 메인 네비게이션 페이지들 */}
              <Route path="/dashboard" element={<NewDashboardPage />} />
              <Route path="/explore" element={<ExplorePage />} />
              <Route path="/activity" element={<ActivityPage />} />
              <Route path="/mentoring" element={<MentoringPage />} />
              <Route path="/governance" element={<GovernancePage />} />
              <Route path="/hall-of-fame" element={<HallOfFamePage />} />

              {/* 프로젝트 관련 라우트 */}
              <Route path="/projects/new" element={<CreateProjectPage />} />
              <Route path="/project/:id" element={<ProjectDetailPage />} />
              <Route path="/project/:id/edit" element={<EditProjectPage />} />
              <Route
                path="/project/:id/update"
                element={<div>진행 상황 업데이트 (준비중)</div>}
              />

              {/* 거래 페이지 */}
              <Route
                path="/trade/:projectId/:milestoneId"
                element={<PolymarketTradingPage />}
              />

              {/* 사용자 관련 라우트 */}
              <Route path="/profile" element={<ProfilePage />} />
              <Route path="/profile/:username" element={<ProfilePage />} />
              <Route path="/settings" element={<AccountSettingsPage />} />
              <Route path="/notifications" element={<NotificationsPage />} />

              {/* 레거시 라우트 (호환성을 위해 유지) */}
              <Route
                path="/create-project"
                element={<Navigate to="/projects/new" replace />}
              />
              <Route
                path="/edit-project/:id"
                element={<Navigate to="/project/:id/edit" replace />}
              />

              {/* 404 처리 */}
              <Route
                path="*"
                element={
                  <div style={{ padding: "100px", textAlign: "center" }}>
                    <h2>페이지를 찾을 수 없습니다</h2>
                    <p>
                      <a href="/">홈으로 돌아가기</a>
                    </p>
                  </div>
                }
              />
            </Routes>
          </Router>
        </ThemeProvider>
      </AntApp>
    </QueryClientProvider>
  );
}

export default App;
