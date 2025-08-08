import { useState } from "react";
import { useAuthStore } from "../stores/useAuthStore";
import type { LoginRequest, RegisterRequest } from "../types";

export default function AuthPage() {
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
  };

  const handleGoogleLogin = async () => {
    await loginWithGoogle();
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8"
      style={{ backgroundColor: "var(--bg-primary)" }}
    >
      <div className="max-w-md w-full space-y-8">
        <div>
          <h1
            className="text-4xl font-bold mb-2"
            style={{ color: "var(--text-primary)" }}
          >
            Blueprint
          </h1>
          <p style={{ color: "var(--text-secondary)" }}>
            당신의 아이디어를 현실로 만들어보세요
          </p>
        </div>

        <div
          className="flex space-x-1 p-1 rounded-lg mb-6"
          style={{ backgroundColor: "var(--bg-secondary)" }}
        >
          <button
            type="button"
            onClick={() => setIsLogin(true)}
            className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${
              isLogin ? "text-white" : ""
            }`}
            style={{
              backgroundColor: isLogin ? "var(--blue)" : "transparent",
              color: isLogin ? "white" : "var(--text-secondary)",
            }}
            onMouseEnter={(e) => {
              if (!isLogin)
                (e.target as HTMLElement).style.backgroundColor =
                  "var(--bg-tertiary)";
            }}
            onMouseLeave={(e) => {
              if (!isLogin)
                (e.target as HTMLElement).style.backgroundColor = "transparent";
            }}
          >
            로그인
          </button>
          <button
            type="button"
            onClick={() => setIsLogin(false)}
            className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${
              !isLogin ? "text-white" : ""
            }`}
            style={{
              backgroundColor: !isLogin ? "var(--blue)" : "transparent",
              color: !isLogin ? "white" : "var(--text-secondary)",
            }}
            onMouseEnter={(e) => {
              if (isLogin)
                (e.target as HTMLElement).style.backgroundColor =
                  "var(--bg-tertiary)";
            }}
            onMouseLeave={(e) => {
              if (isLogin)
                (e.target as HTMLElement).style.backgroundColor = "transparent";
            }}
          >
            회원가입
          </button>
        </div>

        <div
          className="rounded-lg shadow-md p-6"
          style={{
            backgroundColor: "var(--bg-secondary)",
            borderColor: "var(--border-color)",
          }}
        >
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
                onChange={(e) =>
                  setFormData({ ...formData, email: e.target.value })
                }
                className="w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                style={{
                  backgroundColor: "var(--bg-primary)",
                  borderColor: "var(--border-color)",
                  color: "var(--text-primary)",
                }}
                placeholder="your@email.com"
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
                  onChange={(e) =>
                    setFormData({ ...formData, username: e.target.value })
                  }
                  className="w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  style={{
                    backgroundColor: "var(--bg-primary)",
                    borderColor: "var(--border-color)",
                    color: "var(--text-primary)",
                  }}
                  placeholder="사용자명을 입력하세요"
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
                onChange={(e) =>
                  setFormData({ ...formData, password: e.target.value })
                }
                className="w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                style={{
                  backgroundColor: "var(--bg-primary)",
                  borderColor: "var(--border-color)",
                  color: "var(--text-primary)",
                }}
                placeholder="비밀번호를 입력하세요"
              />
            </div>

            {error && (
              <div
                className="text-sm p-3 rounded-md"
                style={{
                  backgroundColor: "var(--bg-tertiary)",
                  color: "var(--red)",
                  border: "1px solid var(--red)",
                }}
              >
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-2 px-4 rounded-md text-white font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              style={{
                backgroundColor: "var(--blue)",
                borderColor: "var(--blue)",
              }}
              onMouseEnter={(e) => {
                if (!isLoading) (e.target as HTMLElement).style.opacity = "0.9";
              }}
              onMouseLeave={(e) => {
                if (!isLoading) (e.target as HTMLElement).style.opacity = "1";
              }}
            >
              {isLoading ? "처리 중..." : isLogin ? "로그인" : "회원가입"}
            </button>

            <div className="relative">
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

            <button
              type="button"
              onClick={handleGoogleLogin}
              disabled={isLoading}
              className="w-full flex justify-center items-center px-4 py-2 border rounded-md shadow-sm text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              style={{
                backgroundColor: "var(--bg-primary)",
                borderColor: "var(--border-color)",
                color: "var(--text-primary)",
              }}
              onMouseEnter={(e) => {
                if (!isLoading)
                  (e.target as HTMLElement).style.backgroundColor =
                    "var(--bg-tertiary)";
              }}
              onMouseLeave={(e) => {
                if (!isLoading)
                  (e.target as HTMLElement).style.backgroundColor =
                    "var(--bg-primary)";
              }}
            >
              <svg className="w-5 h-5 mr-2" viewBox="0 0 24 24">
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
          </form>

          {isLogin && (
            <p
              className="mt-6 text-sm text-center"
              style={{ color: "var(--text-secondary)" }}
            >
              계정이 없으신가요?{" "}
              <button
                onClick={() => setIsLogin(false)}
                className="font-medium"
                style={{ color: "var(--blue)" }}
                onMouseEnter={(e) =>
                  ((e.target as HTMLElement).style.textDecoration = "underline")
                }
                onMouseLeave={(e) =>
                  ((e.target as HTMLElement).style.textDecoration = "none")
                }
              >
                회원가입하기
              </button>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
