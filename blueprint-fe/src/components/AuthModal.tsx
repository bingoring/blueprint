import { useState } from "react";
import { useAuthStore } from "../stores/useAuthStore";
import type { LoginRequest, RegisterRequest } from "../types";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function AuthModal({ isOpen, onClose }: AuthModalProps) {
  const [isLogin, setIsLogin] = useState(true);
  const { login, register, loginWithGoogle, isLoading, error, clearError } =
    useAuthStore();

  const [formData, setFormData] = useState({
    email: "",
    username: "",
    password: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();

    try {
      if (isLogin) {
        const credentials: LoginRequest = {
          email: formData.email,
          password: formData.password,
        };
        await login(credentials);
      } else {
        const userData: RegisterRequest = {
          email: formData.email,
          username: formData.username,
          password: formData.password,
        };
        await register(userData);
      }

      // 성공시 모달 닫기
      onClose();
    } catch {
      // 에러는 스토어에서 처리됨
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleGoogleLogin = async () => {
    await loginWithGoogle();
    // Google 로그인은 리다이렉트되므로 모달을 닫지 않음
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div
        className="rounded-lg max-w-md w-full max-h-[90vh] overflow-y-auto"
        style={{
          backgroundColor: "var(--bg-secondary)",
          border: "1px solid var(--border-color)",
          boxShadow: "0 10px 25px rgba(0, 0, 0, 0.2)",
        }}
      >
        <div className="p-6">
          {/* 헤더 */}
          <div className="flex justify-between items-center mb-6">
            <h3
              className="text-xl font-bold"
              style={{ color: "var(--text-primary)" }}
            >
              {isLogin ? "로그인" : "회원가입"}
            </h3>
            <button
              onClick={onClose}
              className="text-xl transition-colors"
              style={{
                color: "var(--text-secondary)",
              }}
              onMouseEnter={(e) =>
                ((e.target as HTMLElement).style.color = "var(--text-primary)")
              }
              onMouseLeave={(e) =>
                ((e.target as HTMLElement).style.color =
                  "var(--text-secondary)")
              }
            >
              ✕
            </button>
          </div>

          {/* 탭 전환 */}
          <div
            className="flex mb-6 rounded-lg p-1"
            style={{ backgroundColor: "var(--bg-tertiary)" }}
          >
            <button
              onClick={() => {
                setIsLogin(true);
                clearError();
              }}
              className="flex-1 py-2 px-4 text-sm font-medium rounded-md transition-all"
              style={{
                backgroundColor: isLogin
                  ? "var(--bg-secondary)"
                  : "transparent",
                color: isLogin ? "var(--blue)" : "var(--text-secondary)",
                boxShadow: isLogin ? "0 1px 3px rgba(0, 0, 0, 0.1)" : "none",
              }}
            >
              로그인
            </button>
            <button
              onClick={() => {
                setIsLogin(false);
                clearError();
              }}
              className="flex-1 py-2 px-4 text-sm font-medium rounded-md transition-all"
              style={{
                backgroundColor: !isLogin
                  ? "var(--bg-secondary)"
                  : "transparent",
                color: !isLogin ? "var(--blue)" : "var(--text-secondary)",
                boxShadow: !isLogin ? "0 1px 3px rgba(0, 0, 0, 0.1)" : "none",
              }}
            >
              회원가입
            </button>
          </div>

          {/* 폼 */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium mb-1"
                style={{ color: "var(--text-primary)" }}
              >
                이메일
              </label>
              <input
                type="email"
                name="email"
                id="email"
                required
                value={formData.email}
                onChange={handleInputChange}
                className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 transition-all"
                style={{
                  backgroundColor: "var(--bg-tertiary)",
                  border: "1px solid var(--border-color)",
                  color: "var(--text-primary)",
                }}
                placeholder="your@email.com"
                onFocus={(e) =>
                  ((e.target as HTMLElement).style.boxShadow =
                    "0 0 0 2px rgba(24, 144, 255, 0.2)")
                }
                onBlur={(e) =>
                  ((e.target as HTMLElement).style.boxShadow = "none")
                }
              />
            </div>

            {!isLogin && (
              <div>
                <label
                  htmlFor="username"
                  className="block text-sm font-medium mb-1"
                  style={{ color: "var(--text-primary)" }}
                >
                  사용자명
                </label>
                <input
                  type="text"
                  name="username"
                  id="username"
                  required
                  value={formData.username}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 transition-all"
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    border: "1px solid var(--border-color)",
                    color: "var(--text-primary)",
                  }}
                  placeholder="사용자명"
                  onFocus={(e) =>
                    ((e.target as HTMLElement).style.boxShadow =
                      "0 0 0 2px rgba(24, 144, 255, 0.2)")
                  }
                  onBlur={(e) =>
                    ((e.target as HTMLElement).style.boxShadow = "none")
                  }
                />
              </div>
            )}

            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium mb-1"
                style={{ color: "var(--text-primary)" }}
              >
                비밀번호
              </label>
              <input
                type="password"
                name="password"
                id="password"
                required
                value={formData.password}
                onChange={handleInputChange}
                className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 transition-all"
                style={{
                  backgroundColor: "var(--bg-tertiary)",
                  border: "1px solid var(--border-color)",
                  color: "var(--text-primary)",
                }}
                placeholder="비밀번호"
                onFocus={(e) =>
                  ((e.target as HTMLElement).style.boxShadow =
                    "0 0 0 2px rgba(24, 144, 255, 0.2)")
                }
                onBlur={(e) =>
                  ((e.target as HTMLElement).style.boxShadow = "none")
                }
              />
            </div>

            {error && (
              <div
                className="px-3 py-2 rounded-md text-sm"
                style={{
                  backgroundColor: "rgba(255, 77, 109, 0.1)",
                  border: "1px solid rgba(255, 77, 109, 0.3)",
                  color: "var(--red)",
                }}
              >
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full font-medium py-2 px-4 rounded-md transition duration-200"
              style={{
                background: isLoading
                  ? "linear-gradient(to right, rgba(24, 144, 255, 0.5), rgba(147, 51, 234, 0.5))"
                  : "linear-gradient(to right, var(--blue), #9333ea)",
                color: "white",
                opacity: isLoading ? 0.7 : 1,
                cursor: isLoading ? "not-allowed" : "pointer",
              }}
              onMouseEnter={(e) => {
                if (!isLoading) {
                  (e.target as HTMLElement).style.transform =
                    "translateY(-1px)";
                  (e.target as HTMLElement).style.boxShadow =
                    "0 4px 12px rgba(24, 144, 255, 0.3)";
                }
              }}
              onMouseLeave={(e) => {
                (e.target as HTMLElement).style.transform = "translateY(0)";
                (e.target as HTMLElement).style.boxShadow = "none";
              }}
            >
              {isLoading ? "처리중..." : isLogin ? "로그인" : "회원가입"}
            </button>
          </form>

          {/* 구분선 */}
          <div className="relative my-4">
            <div className="absolute inset-0 flex items-center">
              <div
                className="w-full border-t"
                style={{ borderColor: "var(--border-color)" }}
              />
            </div>
            <div className="relative flex justify-center text-sm">
              <span
                className="px-2"
                style={{
                  backgroundColor: "var(--bg-secondary)",
                  color: "var(--text-secondary)",
                }}
              >
                또는
              </span>
            </div>
          </div>

          {/* Google 로그인 버튼 */}
          <button
            type="button"
            onClick={handleGoogleLogin}
            disabled={isLoading}
            className="w-full flex justify-center items-center px-4 py-2 rounded-md shadow-sm text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 transition duration-200"
            style={{
              backgroundColor: "var(--bg-tertiary)",
              border: "1px solid var(--border-color)",
              color: "var(--text-primary)",
              opacity: isLoading ? 0.5 : 1,
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
            onMouseEnter={(e) => {
              if (!isLoading) {
                (e.target as HTMLElement).style.backgroundColor =
                  "var(--bg-primary)";
              }
            }}
            onMouseLeave={(e) => {
              (e.target as HTMLElement).style.backgroundColor =
                "var(--bg-tertiary)";
            }}
            onFocus={(e) =>
              ((e.target as HTMLElement).style.boxShadow =
                "0 0 0 2px rgba(24, 144, 255, 0.2)")
            }
            onBlur={(e) => ((e.target as HTMLElement).style.boxShadow = "none")}
          >
            <svg className="w-4 h-4 mr-2" viewBox="0 0 24 24">
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
            Google로 {isLogin ? "로그인" : "회원가입"}
          </button>
        </div>
      </div>
    </div>
  );
}
