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
  logout: () => Promise<void>; // 비동기로 변경
  getCurrentUser: () => Promise<void>;
  refreshToken: () => Promise<boolean>; // 새로 추가
  clearError: () => void;

  // 세션 관리
  startSessionCheck: () => void;
  stopSessionCheck: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => {
      let sessionCheckInterval: number | null = null;

      return {
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

              // 로그인 성공 시 세션 체크 시작
              get().startSessionCheck();
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

              // 회원가입 성공 시 세션 체크 시작
              get().startSessionCheck();
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

        // 개선된 로그아웃 🚪
        logout: async () => {
          try {
            console.log('🚪 로그아웃 시작...');

            // 세션 체크 중단
            get().stopSessionCheck();

            // 백엔드에 로그아웃 요청 (토큰이 있는 경우에만)
            if (apiClient.getToken()) {
              await apiClient.logout();
              console.log('✅ 서버 로그아웃 완료');
            }

            // 상태 초기화
            set({
              user: null,
              isAuthenticated: false,
              isLoading: false,
              error: null,
            });

            console.log('✅ 로그아웃 완료');
          } catch (error) {
            console.error('❌ 로그아웃 실패:', error);

            // 서버 오류가 있어도 클라이언트에서는 로그아웃 처리
            apiClient.clearToken();
            get().stopSessionCheck();
            set({
              user: null,
              isAuthenticated: false,
              isLoading: false,
              error: null,
            });
          }
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

              // getCurrentUser 성공 시 세션 체크 시작
              get().startSessionCheck();
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

        // 토큰 갱신 🔄
        refreshToken: async () => {
          try {
            console.log('🔄 토큰 갱신 시도...');

            const response = await apiClient.refreshToken();

            if (response.success && response.data) {
              console.log('✅ 토큰 갱신 성공');
              set({
                user: response.data.user,
                isAuthenticated: true,
                error: null,
              });
              return true;
            } else {
              console.log('❌ 토큰 갱신 실패');
              get().logout(); // 갱신 실패 시 자동 로그아웃
              return false;
            }
          } catch (error) {
            console.error('❌ 토큰 갱신 오류:', error);
            get().logout(); // 오류 시 자동 로그아웃
            return false;
          }
        },

        // 세션 체크 시작 ⏰
        startSessionCheck: () => {
          // 기존 인터벌 정리
          if (sessionCheckInterval) {
            clearInterval(sessionCheckInterval);
          }

          // 2분마다 토큰 상태 확인
          sessionCheckInterval = setInterval(async () => {
            const { isAuthenticated } = get();
            if (isAuthenticated) {
              console.log('⏰ 토큰 상태 확인 중...');

              try {
                // 토큰 만료 상태 확인
                const expiryResponse = await apiClient.checkTokenExpiry();

                if (expiryResponse.success && expiryResponse.data) {
                  const { should_refresh, remaining_minutes, is_expired } = expiryResponse.data;

                  console.log(`⏰ 토큰 상태: ${remaining_minutes}분 남음`);

                  if (is_expired) {
                    console.log('❌ 토큰 만료 - 자동 로그아웃');
                    get().logout();
                  } else if (should_refresh) {
                    console.log('🔄 토큰 자동 갱신 시작...');
                    const refreshSuccess = await get().refreshToken();

                    if (refreshSuccess) {
                      console.log('✅ 토큰 자동 갱신 성공');
                    } else {
                      console.log('❌ 토큰 자동 갱신 실패 - 로그아웃');
                    }
                  } else {
                    console.log('✅ 토큰 상태 양호');
                  }
                } else {
                  console.log('❌ 토큰 상태 확인 실패');
                  get().logout();
                }
              } catch {
                console.log('❌ 세션 체크 오류 - 자동 로그아웃');
                get().logout();
              }
            }
          }, 2 * 60 * 1000); // 2분

          console.log('⏰ 세션 체크 시작 (2분 간격, 자동 토큰 갱신 포함)');
        },

        // 세션 체크 중단 ⏹️
        stopSessionCheck: () => {
          if (sessionCheckInterval) {
            clearInterval(sessionCheckInterval);
            sessionCheckInterval = null;
            console.log('⏹️ 세션 체크 중단');
          }
        },

        clearError: () => set({ error: null }),
      };
    },
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
