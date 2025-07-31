import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, LoginRequest, RegisterRequest } from '../types';
import { apiClient } from '../lib/api';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  loginWithGoogle: () => Promise<void>;
  handleGoogleCallback: (code: string) => Promise<void>;
  logout: () => void;
  getCurrentUser: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (credentials: LoginRequest) => {
        set({ isLoading: true, error: null });

        try {
          const response = await apiClient.login(credentials);

          if (response.success && response.data) {
            set({
              user: response.data.user,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
          } else {
            set({
              error: response.error || 'Login failed',
              isLoading: false,
            });
          }
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Login failed',
            isLoading: false,
          });
        }
      },

      register: async (userData: RegisterRequest) => {
        set({ isLoading: true, error: null });

        try {
          const response = await apiClient.register(userData);

          if (response.success && response.data) {
            set({
              user: response.data.user,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
          } else {
            set({
              error: response.error || 'Registration failed',
              isLoading: false,
            });
          }
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Registration failed',
            isLoading: false,
          });
        }
      },

                              loginWithGoogle: async () => {
              set({ isLoading: true, error: null });
              try {
                console.log('🔵 Google 로그인 시작...');
                const response = await apiClient.getGoogleAuthUrl();
                console.log('📡 API 응답:', response);
                if (response.success && response.data) {
                  // 같은 창에서 Google OAuth로 리다이렉트
                  console.log('🚀 Google OAuth로 리다이렉트:', response.data.auth_url);
                  window.location.href = response.data.auth_url;
                } else {
                  console.error('❌ Google Auth URL 가져오기 실패:', response);
                  set({
                    isLoading: false,
                    error: response.message || response.error || 'Failed to get Google auth URL',
                  });
                }
              } catch (error) {
                console.error('❌ Google 로그인 에러:', error);
                set({
                  isLoading: false,
                  error: error instanceof Error ? error.message : 'Google login failed',
                });
              }
      },

      handleGoogleCallback: async (code: string) => {
        try {
          const response = await apiClient.handleGoogleCallback(code);
          if (response.success && response.data) {
            // 부모 창에 성공 메시지 전송
            window.opener?.postMessage({
              type: 'GOOGLE_AUTH_SUCCESS',
              user: response.data.user,
              token: response.data.token,
            }, window.location.origin);
            window.close();
          } else {
            // 부모 창에 에러 메시지 전송
            window.opener?.postMessage({
              type: 'GOOGLE_AUTH_ERROR',
              error: response.message || 'Google login failed',
            }, window.location.origin);
            window.close();
          }
        } catch (error) {
          window.opener?.postMessage({
            type: 'GOOGLE_AUTH_ERROR',
            error: error instanceof Error ? error.message : 'Google login failed',
          }, window.location.origin);
          window.close();
        }
      },

      logout: () => {
        apiClient.logout();
        set({
          user: null,
          isAuthenticated: false,
          error: null,
        });
      },

      getCurrentUser: async () => {
        // 토큰이 있으면 인증 시도 (Google OAuth 플로우 지원)
        const hasToken = apiClient.getToken();
        if (!get().isAuthenticated && !hasToken) {
          console.log('🚫 토큰이 없어서 getCurrentUser 건너뜀');
          return;
        }

        set({ isLoading: true });

        try {
          console.log('📡 /api/v1/me 호출 중...');
          const response = await apiClient.getCurrentUser();
          console.log('📡 /api/v1/me 응답:', response);

          if (response.success && response.data) {
            console.log('✅ 사용자 정보 로드 성공:', response.data);
            set({
              user: response.data,
              isAuthenticated: true, // ← 중요: 인증 상태 true로 설정
              isLoading: false,
              error: null,
            });
          } else {
            console.error('❌ 사용자 정보 로드 실패:', response);
            // 토큰이 만료되었거나 유효하지 않음
            get().logout();
          }
        } catch (error) {
          console.error('❌ getCurrentUser 에러:', error);
          get().logout();
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
