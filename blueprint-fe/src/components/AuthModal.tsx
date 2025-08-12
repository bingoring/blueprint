import { CopyOutlined, GoogleOutlined } from "@ant-design/icons";
import {
  Button,
  Divider,
  Input,
  Typography,
  message,
  notification,
} from "antd";
import { useState } from "react";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";

const { Title, Text } = Typography;

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function AuthModal({ isOpen, onClose }: AuthModalProps) {
  const [step, setStep] = useState<"email" | "verify">("email");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [verificationCode, setVerificationCode] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { loginWithGoogle } = useAuthStore();

  const handleEmailSubmit = async () => {
    if (!email) {
      message.error("이메일을 입력해주세요");
      return;
    }

    setIsLoading(true);
    try {
      const response = await apiClient.createMagicLink({ email });

      if (response.success && response.data) {
        setVerificationCode(response.data.code); // 개발/테스트용
        setStep("verify");

        notification.success({
          message: "인증 코드 발송 완료",
          description:
            "이메일로 인증 코드가 발송되었습니다. 받은 편지함을 확인해주세요.",
          placement: "topRight",
          duration: 4,
        });
      }
    } catch (error) {
      console.error("Magic link creation failed:", error);
      notification.error({
        message: "인증 코드 발송 실패",
        description: "이메일 발송 중 오류가 발생했습니다. 다시 시도해주세요.",
        placement: "topRight",
        duration: 4,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCodeSubmit = async () => {
    if (!code) {
      message.error("인증 코드를 입력해주세요");
      return;
    }

    setIsLoading(true);
    try {
      const response = await apiClient.verifyMagicLink({ code });

      if (response.success && response.data) {
        notification.success({
          message: "인증 완료",
          description: "성공적으로 로그인되었습니다. 계정 설정을 완료해주세요.",
          placement: "topRight",
          duration: 3,
        });

        // 토큰 설정 및 사용자 정보 업데이트
        apiClient.setToken(response.data.token);

        onClose();

        // 계정 설정 페이지로 이동
        window.location.href = "/settings";
      }
    } catch (error) {
      console.error("Magic link verification failed:", error);
      notification.error({
        message: "인증 실패",
        description: "인증 코드가 올바르지 않거나 만료되었습니다.",
        placement: "topRight",
        duration: 4,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const copyCode = () => {
    navigator.clipboard.writeText(verificationCode);
    message.success("인증 코드가 복사되었습니다");
  };

  const resetFlow = () => {
    setStep("email");
    setEmail("");
    setCode("");
    setVerificationCode("");
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={onClose}
    >
      <div
        className="rounded-lg max-w-md w-full max-h-[90vh] overflow-y-auto"
        style={{
          backgroundColor: "var(--bg-secondary)",
          border: "1px solid var(--border-color)",
          boxShadow: "0 10px 25px rgba(0, 0, 0, 0.2)",
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          {/* 헤더 */}
          <div className="flex justify-between items-center mb-6">
            <Title
              level={3}
              style={{ margin: 0, color: "var(--text-primary)" }}
            >
              {step === "email" ? "Blueprint에 로그인" : "Check your email"}
            </Title>
            <button
              onClick={onClose}
              className="text-xl transition-colors"
              style={{ color: "var(--text-secondary)" }}
            >
              ✕
            </button>
          </div>

          {step === "email" ? (
            // 이메일 입력 단계
            <div className="space-y-4">
              <div className="text-center mb-6">
                <div
                  className="mb-4"
                  style={{ fontSize: 48, color: "var(--blue)" }}
                >
                  📧
                </div>
                <Text style={{ color: "var(--text-secondary)" }}>
                  이메일 주소를 입력하면 로그인 링크를 보내드립니다
                </Text>
              </div>

              <div>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="your@email.com"
                  size="large"
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    border: "1px solid var(--border-color)",
                    color: "var(--text-primary)",
                  }}
                  onPressEnter={handleEmailSubmit}
                />
              </div>

              <Button
                type="primary"
                size="large"
                loading={isLoading}
                onClick={handleEmailSubmit}
                className="w-full btn-primary"
                style={{
                  height: "48px",
                  fontSize: "16px",
                  fontWeight: "500",
                }}
              >
                {isLoading ? "발송 중..." : "로그인 링크 보내기"}
              </Button>

              <Divider style={{ color: "var(--text-secondary)" }}>또는</Divider>

              <Button
                size="large"
                onClick={loginWithGoogle}
                className="w-full"
                style={{
                  height: "48px",
                  fontSize: "16px",
                  fontWeight: "500",
                  backgroundColor: "#1f2937",
                  color: "white",
                  border: "1px solid #374151",
                  transition: "all 0.2s ease-in-out",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = "#1f2937";
                  e.currentTarget.style.color = "#3b82f6";
                  e.currentTarget.style.borderColor = "#3b82f6";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = "#1f2937";
                  e.currentTarget.style.color = "white";
                  e.currentTarget.style.borderColor = "#374151";
                }}
              >
                <GoogleOutlined style={{ marginRight: 8 }} />
                Google로 계속하기
              </Button>
            </div>
          ) : (
            // 인증 코드 확인 단계
            <div className="space-y-4">
              <div className="text-center mb-6">
                <div
                  className="mb-4"
                  style={{ fontSize: 48, color: "var(--blue)" }}
                >
                  📬
                </div>
                <Text style={{ color: "var(--text-primary)", fontSize: 16 }}>
                  로그인 링크를 다음 주소로 보냈습니다:
                </Text>
                <div
                  style={{
                    color: "var(--blue)",
                    fontSize: 16,
                    fontWeight: 600,
                    marginTop: 8,
                  }}
                >
                  {email}
                </div>
              </div>

              {/* 개발/테스트용 코드 표시 */}
              {verificationCode && (
                <div
                  className="text-center p-4 rounded-lg"
                  style={{ backgroundColor: "var(--bg-tertiary)" }}
                >
                  <Text
                    style={{ color: "var(--text-secondary)", fontSize: 14 }}
                  >
                    개발용 인증 코드:
                  </Text>
                  <div
                    className="text-center cursor-pointer mt-2"
                    onClick={copyCode}
                    style={{
                      fontSize: 24,
                      fontWeight: "bold",
                      color: "var(--text-primary)",
                      padding: "8px",
                      borderRadius: "8px",
                      backgroundColor: "var(--bg-secondary)",
                      border: "2px dashed var(--border-color)",
                    }}
                  >
                    {verificationCode}
                    <CopyOutlined style={{ marginLeft: 8, fontSize: 16 }} />
                  </div>
                </div>
              )}

              <div>
                <Text style={{ color: "var(--text-primary)", fontSize: 14 }}>
                  보안 코드 입력:
                </Text>
                <Input
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="123456"
                  size="large"
                  maxLength={6}
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    border: "1px solid var(--border-color)",
                    color: "var(--text-primary)",
                    fontSize: 18,
                    textAlign: "center",
                    letterSpacing: "4px",
                    marginTop: 8,
                  }}
                  onPressEnter={handleCodeSubmit}
                />
              </div>

              <Button
                type="primary"
                size="large"
                loading={isLoading}
                onClick={handleCodeSubmit}
                className="w-full btn-primary"
                style={{
                  height: "48px",
                  fontSize: "16px",
                  fontWeight: "500",
                }}
              >
                {isLoading ? "인증 중..." : "인증하고 로그인"}
              </Button>

              <div className="text-center">
                <Button
                  type="link"
                  onClick={resetFlow}
                  className="btn-ghost"
                  style={{ color: "var(--text-secondary)" }}
                >
                  다른 이메일로 시도하기
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
