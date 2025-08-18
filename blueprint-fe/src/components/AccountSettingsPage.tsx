import {
  BellOutlined,
  BookOutlined,
  CheckCircleOutlined,
  CompassOutlined,
  DollarOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  GithubOutlined,
  LinkedinOutlined,
  MailOutlined,
  MoonOutlined,
  PhoneOutlined,
  ProjectOutlined,
  RocketOutlined,
  SafetyCertificateOutlined,
  SaveOutlined,
  SearchOutlined,
  SecurityScanOutlined,
  StarOutlined,
  SunOutlined,
  TeamOutlined,
  TrophyOutlined,
  TwitterOutlined,
  UploadOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Badge,
  Button,
  Card,
  Divider,
  Dropdown,
  Form,
  Input,
  Layout,
  List,
  Progress,
  Space,
  Switch,
  Tabs,
  Tag,
  Typography,
  message,
  notification,
  type MenuProps,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTheme } from "../hooks/useTheme";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { SettingsAggregateResponse } from "../types";
import LanguageSwitcher from "./LanguageSwitcher";

const { Header, Content } = Layout;
const { Title, Text, Paragraph } = Typography;
const { Search } = Input;
const { TextArea } = Input;
const { TabPane } = Tabs;

// 인증 상태 타입 정의 (UI 전용)
interface VerificationStatusUI {
  level1: { email: boolean; phone: boolean };
  level2: {
    linkedin: boolean;
    github: boolean;
    twitter: boolean;
    workEmail: boolean;
    workEmailCompany?: string;
  };
  level3: {
    professional: boolean;
    education: boolean;
    professionalTitle?: string;
    educationDegree?: string;
  };
  level4: {
    sbtCount: number;
    projectSuccessRate: number;
    mentoringSuccessRate: number;
  };
}

const AccountSettingsPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const { isDark, toggleTheme } = useTheme();
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState("profile");
  const [loading, setLoading] = useState(false);

  // notification API 사용
  const [api, contextHolder] = notification.useNotification();

  // 토글 상태 (서버 동기화)
  const [emailNoti, setEmailNoti] = useState<boolean>(true);
  const [pushNoti, setPushNoti] = useState<boolean>(false);
  const [marketingNoti, setMarketingNoti] = useState<boolean>(false);
  const [profilePublic, setProfilePublic] = useState<boolean>(true);
  const [investmentPublic, setInvestmentPublic] = useState<boolean>(false);

  // 신원 증명 상태 (UI 전용)
  const [verificationStatus, setVerificationStatus] =
    useState<VerificationStatusUI>({
      level1: { email: false, phone: false },
      level2: {
        linkedin: false,
        github: false,
        twitter: false,
        workEmail: false,
      },
      level3: { professional: false, education: false },
      level4: { sbtCount: 0, projectSuccessRate: 0, mentoringSuccessRate: 0 },
    });

  // 설정 로드
  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const res = await apiClient.getMySettings();
        if (res.success && res.data) {
          const data = res.data as SettingsAggregateResponse;
          console.log("📧 Settings API response:", data);
          console.log("📧 User email from API:", data.user?.email);
          console.log("📧 User email from store:", user?.email);

          // 프로필 폼 초기화
          form.setFieldsValue({
            displayName: data.profile?.display_name || user?.displayName || "",
            email: data.user?.email || user?.email || "",
            bio: data.profile?.bio || "",
          });

          setEmailNoti(!!data.profile?.email_notifications);
          setPushNoti(!!data.profile?.push_notifications);
          setMarketingNoti(!!data.profile?.marketing_notifications);
          setProfilePublic(!!data.profile?.profile_public);
          setInvestmentPublic(!!data.profile?.investment_public);

          // 검증 상태 매핑
          setVerificationStatus((prev) => ({
            ...prev,
            level1: {
              email: !!data.verification?.email_verified,
              phone: !!data.verification?.phone_verified,
            },
            level2: {
              linkedin: !!data.verification?.linkedin_connected,
              github: !!data.verification?.github_connected,
              twitter: !!data.verification?.twitter_connected,
              workEmail: !!data.verification?.work_email_verified,
              workEmailCompany: data.verification?.work_email_company,
            },
            level3: {
              professional:
                data.verification?.professional_status === "approved",
              education: data.verification?.education_status === "approved",
              professionalTitle: data.verification?.professional_title,
              educationDegree: data.verification?.education_degree,
            },
          }));
        }
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [form, user]);

  // URL 파라미터 확인 - OAuth 연결 완료/실패 처리
  const urlParams = new URLSearchParams(window.location.search);
  const connectedProvider = urlParams.get("connected");
  const connectedName = urlParams.get("name");
  const errorType = urlParams.get("error");
  const errorProvider = urlParams.get("provider");

  if (connectedProvider && connectedName) {
    api.success({
      message: `${connectedProvider.toUpperCase()} 연결 완료`,
      description: `${connectedName}님의 ${connectedProvider.toUpperCase()} 계정이 성공적으로 연결되었습니다. 신뢰도가 향상되었습니다.`,
      placement: "topRight",
      duration: 4,
    });

    // 연결 상태 업데이트
    setVerificationStatus((prev) => ({
      ...prev,
      level2: {
        ...prev.level2,
        [connectedProvider]: true,
      },
    }));

    // URL에서 파라미터 제거
    const newUrl = window.location.origin + window.location.pathname;
    window.history.replaceState({}, document.title, newUrl);
  } else if (errorType) {
    const errorMessage = "소셜 계정 연결 실패";
    let errorDescription = "알 수 없는 오류가 발생했습니다.";

    switch (errorType) {
      case "oauth_failed":
        errorDescription = `${errorProvider} OAuth 인증에 실패했습니다. 다시 시도해주세요.`;
        break;
      case "connection_failed":
        errorDescription = `${errorProvider} 계정 연결 중 오류가 발생했습니다.`;
        break;
      case "no_code":
        errorDescription = "인증 코드를 받지 못했습니다. 다시 시도해주세요.";
        break;
    }

    api.error({
      message: errorMessage,
      description: errorDescription,
      placement: "topRight",
      duration: 5,
    });

    // URL에서 파라미터 제거
    const newUrl = window.location.origin + window.location.pathname;
    window.history.replaceState({}, document.title, newUrl);
  }

  // User dropdown menu
  const userMenuItems: MenuProps["items"] = [
    {
      key: "profile-header",
      label: (
        <div
          style={{
            padding: "8px 0",
            borderBottom: "1px solid var(--border-color)",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <Avatar
              size={40}
              src={`https://api.dicebear.com/6.x/avataaars/svg?seed=${user?.username}`}
            />
            <div>
              <div style={{ fontWeight: 600, fontSize: 14 }}>
                {user?.username}
              </div>
              <div style={{ color: "var(--text-secondary)", fontSize: 12 }}>
                @{user?.username}
              </div>
            </div>
          </div>
        </div>
      ),
      disabled: true,
    },
    {
      key: "my-profile",
      icon: <RocketOutlined />,
      label: "내 프로필",
      onClick: () => navigate(`/profile/${user?.username}`),
    },
    {
      key: "settings",
      icon: <UserOutlined />,
      label: "계정 설정",
      onClick: () => navigate("/settings"),
    },
    {
      key: "theme",
      icon: isDark ? <SunOutlined /> : <MoonOutlined />,
      label: isDark ? "라이트 모드" : "다크 모드",
      onClick: toggleTheme,
    },
    {
      type: "divider",
    },
    {
      key: "wallet-header",
      label: (
        <div
          style={{
            color: "var(--text-secondary)",
            fontSize: 12,
            fontWeight: 600,
          }}
        >
          지갑 현황
        </div>
      ),
      disabled: true,
    },
    {
      key: "wallet-usdc",
      icon: <DollarOutlined style={{ color: "var(--green)" }} />,
      label: (
        <div>
          <div>USDC 잔액</div>
          <div style={{ fontSize: 12, color: "var(--text-secondary)" }}>
            $1,520.50
          </div>
        </div>
      ),
      disabled: true,
    },
    {
      key: "wallet-blueprint",
      icon: <StarOutlined style={{ color: "var(--gold)" }} />,
      label: (
        <div>
          <div>BLUEPRINT 토큰</div>
          <div style={{ fontSize: 12, color: "var(--text-secondary)" }}>
            25,000
          </div>
        </div>
      ),
      disabled: true,
    },
    {
      type: "divider",
    },
    {
      key: "logout",
      icon: <UserOutlined />,
      label: "로그아웃",
      onClick: logout,
      style: { color: "var(--red)" },
    },
  ];

  const handleSave = async (values: {
    displayName: string;
    email: string;
    bio: string;
  }) => {
    try {
      setLoading(true);
      await apiClient.updateMyProfile({
        display_name: values.displayName,
        bio: values.bio,
      });

      api.success({
        message: "프로필 저장 완료",
        description: "프로필 정보가 성공적으로 업데이트되었습니다.",
        placement: "topRight",
        duration: 3,
      });

      message.success("설정이 저장되었습니다!");
    } catch {
      api.error({
        message: "프로필 저장 실패",
        description:
          "프로필 정보 업데이트 중 오류가 발생했습니다. 다시 시도해주세요.",
        placement: "topRight",
        duration: 4,
      });

      message.error("설정 저장에 실패했습니다.");
    } finally {
      setLoading(false);
    }
  };

  // 토글 핸들러
  const handleToggle = async (
    key:
      | "email_notifications"
      | "push_notifications"
      | "marketing_notifications"
      | "profile_public"
      | "investment_public",
    value: boolean
  ) => {
    try {
      await apiClient.updatePreferences({ [key]: value });

      const settingLabels = {
        email_notifications: "이메일 알림",
        push_notifications: "푸시 알림",
        marketing_notifications: "마케팅 알림",
        profile_public: "프로필 공개",
        investment_public: "투자 내역 공개",
      };

      switch (key) {
        case "email_notifications":
          setEmailNoti(value);
          break;
        case "push_notifications":
          setPushNoti(value);
          break;
        case "marketing_notifications":
          setMarketingNoti(value);
          break;
        case "profile_public":
          setProfilePublic(value);
          break;
        case "investment_public":
          setInvestmentPublic(value);
          break;
      }

      notification.success({
        message: "설정 변경 완료",
        description: `${settingLabels[key]}이(가) ${
          value ? "활성화" : "비활성화"
        }되었습니다.`,
        placement: "topRight",
        duration: 2,
      });

      message.success("변경되었습니다");
    } catch {
      notification.error({
        message: "설정 변경 실패",
        description: "설정 변경 중 오류가 발생했습니다. 다시 시도해주세요.",
        placement: "topRight",
        duration: 3,
      });

      message.error("변경 실패");
    }
  };

  // Verification actions
  const onRequestEmailVerify = async () => {
    console.log("🔔 이메일 인증 요청 시작");

    try {
      await apiClient.requestVerifyEmail();
      console.log("✅ API 호출 성공");

      api.success({
        message: "이메일 인증 요청 완료",
        description: "인증 메일이 발송되었습니다. 받은 편지함을 확인해주세요.",
        placement: "topRight",
        duration: 4,
      });

      message.success("인증 메일을 발송했습니다");
    } catch (error) {
      console.error("❌ API 호출 실패:", error);

      api.error({
        message: "이메일 인증 요청 실패",
        description:
          "인증 메일 발송에 실패했습니다. 이메일 주소를 확인하고 다시 시도해주세요.",
        placement: "topRight",
        duration: 4,
      });

      message.error("요청 실패");
    }
  };

  const onRequestPhoneVerify = async () => {
    try {
      await apiClient.requestVerifyPhone();

      notification.success({
        message: "휴대폰 인증 요청 완료",
        description: "인증 SMS가 발송되었습니다. 메시지를 확인해주세요.",
        placement: "topRight",
        duration: 4,
      });

      message.success("인증 SMS를 발송했습니다");
    } catch {
      notification.error({
        message: "휴대폰 인증 요청 실패",
        description:
          "인증 SMS 발송에 실패했습니다. 전화번호를 확인하고 다시 시도해주세요.",
        placement: "topRight",
        duration: 4,
      });

      message.error("요청 실패");
    }
  };
  const onConnect = async (provider: "linkedin" | "github" | "twitter") => {
    try {
      if (provider === "linkedin") {
        // LinkedIn OAuth 플로우 시작
        const response = await apiClient.connectLinkedIn();
        if (response.success && response.data?.auth_url) {
          // 팝업 창에서 OAuth 페이지 열기
          const popup = window.open(
            response.data.auth_url,
            "linkedin_oauth",
            "width=600,height=700,scrollbars=yes,resizable=yes"
          );

          api.info({
            message: "LinkedIn 연결 중",
            description:
              "LinkedIn 로그인 창에서 인증을 완료해주세요. 완료되면 자동으로 연결됩니다.",
            placement: "topRight",
            duration: 5,
          });

          // 팝업이 닫히는 것을 감지하는 인터벌
          const checkClosed = setInterval(() => {
            if (popup?.closed) {
              clearInterval(checkClosed);
              // 페이지를 다시 로드하여 연결 상태 확인
              setTimeout(() => {
                window.location.reload();
              }, 1000);
            }
          }, 1000);
        }
      } else {
        // GitHub, Twitter는 아직 미구현
        api.info({
          message: `${provider} 연결`,
          description: `${provider} 연결 기능은 곧 출시될 예정입니다.`,
          placement: "topRight",
          duration: 3,
        });
      }
    } catch (error) {
      console.error(`${provider} 연결 실패:`, error);
      api.error({
        message: "소셜 계정 연결 실패",
        description:
          "소셜 계정 연결 중 오류가 발생했습니다. 다시 시도해주세요.",
        placement: "topRight",
        duration: 4,
      });
    }
  };

  const onVerifyWorkEmail = async () => {
    try {
      const res = await apiClient.verifyWorkEmail("Company");
      if (res.success) {
        setVerificationStatus((v) => ({
          ...v,
          level2: {
            ...v.level2,
            workEmail: true,
            workEmailCompany:
              res.data?.work_email_company || v.level2.workEmailCompany,
          },
        }));

        notification.success({
          message: "직장 이메일 인증 완료",
          description:
            "직장 이메일이 성공적으로 인증되었습니다. 전문 신원이 확인되었습니다.",
          placement: "topRight",
          duration: 3,
        });

        message.success("직장 이메일 인증 완료");
      }
    } catch {
      notification.error({
        message: "직장 이메일 인증 실패",
        description:
          "직장 이메일 인증 중 오류가 발생했습니다. 이메일 주소를 확인하고 다시 시도해주세요.",
        placement: "topRight",
        duration: 4,
      });

      message.error("인증 실패");
    }
  };

  // 신원 증명 탭 렌더링
  const currentLevel = (() => {
    const level1Complete =
      verificationStatus.level1.email && verificationStatus.level1.phone;
    const level2Count = [
      verificationStatus.level2.linkedin,
      verificationStatus.level2.github,
      verificationStatus.level2.twitter,
      verificationStatus.level2.workEmail,
    ].filter(Boolean).length;
    const level3Complete =
      verificationStatus.level3.professional ||
      verificationStatus.level3.education;
    if (level3Complete) return 4;
    if (level2Count >= 2) return 3;
    if (level1Complete) return 2;
    if (verificationStatus.level1.email) return 1;
    return 0;
  })();

  const renderVerificationTab = () => (
    <div style={{ maxWidth: 800, margin: "0 auto" }}>
      <Card style={{ marginBottom: 24 }} loading={loading}>
        <div style={{ textAlign: "center", marginBottom: 16 }}>
          <Title level={4}>신뢰도 레벨 {currentLevel}/4</Title>
          <Progress
            percent={(currentLevel / 4) * 100}
            strokeColor={{ "0%": "#108ee9", "100%": "#87d068" }}
            style={{ marginBottom: 8 }}
          />
          <Text type="secondary">
            높은 신뢰도는 더 많은 투자자와 멘티들의 관심을 받습니다
          </Text>
        </div>
      </Card>

      <Card
        title={
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <CheckCircleOutlined
              style={{
                color: verificationStatus.level1.email
                  ? "var(--green)"
                  : "var(--text-secondary)",
              }}
            />
            <span>레벨 1: 기본 계정 인증 (필수)</span>
            {verificationStatus.level1.email &&
              verificationStatus.level1.phone && <Tag color="green">완료</Tag>}
          </div>
        }
        style={{ marginBottom: 16 }}
        loading={loading}
      >
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          유령/스팸 계정이 아닌, 실제 사용하는 계정임을 증명합니다.
        </Paragraph>
        <List
          dataSource={[
            {
              title: "이메일 주소 인증",
              description: `내 이메일: ${user?.email?.replace(
                /(.{2}).*(@.*)/,
                "$1***$2"
              )}`,
              status: verificationStatus.level1.email,
              action: verificationStatus.level1.email ? null : "인증하기",
              statusText: verificationStatus.level1.email
                ? "인증 완료"
                : "미인증",
              icon: <MailOutlined />,
            },
            {
              title: "휴대폰 본인인증",
              description: "대한민국 휴대폰 번호로 본인 인증을 진행합니다.",
              status: verificationStatus.level1.phone,
              action: verificationStatus.level1.phone ? null : "본인인증",
              statusText: verificationStatus.level1.phone
                ? "인증 완료"
                : "미인증",
              icon: <PhoneOutlined />,
            },
          ]}
          renderItem={(item) => (
            <List.Item
              actions={[
                item.action ? (
                  <Button
                    type="primary"
                    size="small"
                    onClick={() => {
                      console.log("🔔 버튼 클릭:", item.title);
                      if (item.title === "이메일 주소 인증") {
                        onRequestEmailVerify();
                      } else if (item.title === "휴대폰 본인인증") {
                        onRequestPhoneVerify();
                      }
                    }}
                  >
                    {item.action}
                  </Button>
                ) : (
                  <Tag color="green" icon={<CheckCircleOutlined />}>
                    {item.statusText}
                  </Tag>
                ),
              ]}
            >
              <List.Item.Meta
                avatar={item.icon}
                title={item.title}
                description={
                  <div>
                    <div>{item.description}</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      상태: {item.statusText}
                    </Text>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Card>

      <Card
        title={
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <SafetyCertificateOutlined
              style={{
                color:
                  currentLevel >= 2 ? "var(--blue)" : "var(--text-secondary)",
              }}
            />
            <span>레벨 2: 소셜 & 커리어 증명 (선택)</span>
            {currentLevel >= 2 && <Tag color="blue">진행 중</Tag>}
          </div>
        }
        style={{ marginBottom: 16 }}
        loading={loading}
      >
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          사용자의 온라인 상 페르소나와 직업적 신원을 연결하여 신뢰도를
          높입니다.
        </Paragraph>
        <div style={{ marginBottom: 24 }}>
          <Title level={5}>소셜 미디어 연결</Title>
          <List
            dataSource={[
              {
                title: "LinkedIn 연결",
                description: "가장 강력한 커리어 증명 수단입니다.",
                status: verificationStatus.level2.linkedin,
                icon: <LinkedinOutlined style={{ color: "#0077b5" }} />,
                buttonIcon: (
                  <LinkedinOutlined style={{ color: "currentColor" }} />
                ),
                buttonText: "Connect with LinkedIn",
                onClick: () => onConnect("linkedin"),
              },
              {
                title: "GitHub 연결",
                description: "개발자들에게 필수적인 증명 수단입니다.",
                status: verificationStatus.level2.github,
                icon: <GithubOutlined />,
                buttonIcon: (
                  <GithubOutlined style={{ color: "currentColor" }} />
                ),
                buttonText: "Connect with GitHub",
                onClick: () => onConnect("github"),
              },
              {
                title: "X (Twitter) 연결",
                description: "개인 브랜딩 및 영향력을 보여줍니다.",
                status: verificationStatus.level2.twitter,
                icon: <TwitterOutlined style={{ color: "#1da1f2" }} />,
                buttonIcon: (
                  <TwitterOutlined style={{ color: "currentColor" }} />
                ),
                buttonText: "Connect with X",
                onClick: () => onConnect("twitter"),
              },
            ]}
            renderItem={(item) => (
              <List.Item
                actions={[
                  item.status ? (
                    <Tag color="green" icon={<CheckCircleOutlined />}>
                      연결됨
                    </Tag>
                  ) : (
                    <Button
                      icon={item.buttonIcon}
                      size="small"
                      onClick={item.onClick}
                      style={{
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
                      {item.buttonText}
                    </Button>
                  ),
                ]}
              >
                <List.Item.Meta
                  avatar={item.icon}
                  title={item.title}
                  description={item.description}
                />
              </List.Item>
            )}
          />
        </div>
        <div style={{ marginBottom: 16 }}>
          <Title level={5}>직장 이메일 인증</Title>
          <div
            style={{
              padding: 16,
              background: "var(--bg-secondary)",
              borderRadius: 8,
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div style={{ fontWeight: 600, marginBottom: 4 }}>
                회사 이메일 주소를 통해 재직 사실을 인증하세요
              </div>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {verificationStatus.level2.workEmail
                  ? `인증 완료: ${
                      verificationStatus.level2.workEmailCompany || "회사"
                    }`
                  : "상태: 미인증"}
              </Text>
            </div>
            {!verificationStatus.level2.workEmail && (
              <Button type="primary" size="small" onClick={onVerifyWorkEmail}>
                인증하기
              </Button>
            )}
          </div>
        </div>
      </Card>

      <Card
        title={
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <SecurityScanOutlined
              style={{
                color:
                  currentLevel >= 3 ? "var(--gold)" : "var(--text-secondary)",
              }}
            />
            <span>레벨 3: 전문 자격 & 학력 증명 (선택)</span>
            {currentLevel >= 3 && <Tag color="gold">인증됨</Tag>}
          </div>
        }
        style={{ marginBottom: 16 }}
        loading={loading}
      >
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          변호사, 의사, 공인회계사 등 특정 자격이나 학위가 중요한 전문가들을
          위한 최고 수준의 신뢰 증명입니다.
        </Paragraph>
        <List
          dataSource={[
            {
              title: "전문 자격증/면허증 인증",
              description:
                "자격증, 면허증 등 전문성을 증명하는 서류를 업로드하세요.",
              status: verificationStatus.level3.professional,
              statusText: verificationStatus.level3.professional
                ? `승인 완료: ${
                    verificationStatus.level3.professionalTitle || "전문가 자격"
                  }`
                : "미인증",
              icon: <SafetyCertificateOutlined />,
              action: async () => {
                try {
                  await apiClient.submitProfessionalDoc();

                  notification.success({
                    message: "전문 자격증 제출 완료",
                    description:
                      "전문 자격증 서류가 성공적으로 제출되었습니다. 검토 후 승인 여부를 알려드리겠습니다.",
                    placement: "topRight",
                    duration: 4,
                  });

                  message.success("제출되었습니다");
                } catch {
                  notification.error({
                    message: "전문 자격증 제출 실패",
                    description:
                      "서류 제출 중 오류가 발생했습니다. 파일을 확인하고 다시 시도해주세요.",
                    placement: "topRight",
                    duration: 4,
                  });

                  message.error("제출 실패");
                }
              },
            },
            {
              title: "학위 증명",
              description: "졸업 증명서, 학위기 등을 통해 학력을 인증하세요.",
              status: verificationStatus.level3.education,
              statusText: verificationStatus.level3.education
                ? `승인 완료: ${
                    verificationStatus.level3.educationDegree || "학위"
                  }`
                : "미인증",
              icon: <BookOutlined />,
              action: async () => {
                try {
                  await apiClient.submitEducationDoc();

                  notification.success({
                    message: "학위 증명 제출 완료",
                    description:
                      "학위 증명 서류가 성공적으로 제출되었습니다. 검토 후 승인 여부를 알려드리겠습니다.",
                    placement: "topRight",
                    duration: 4,
                  });

                  message.success("제출되었습니다");
                } catch {
                  notification.error({
                    message: "학위 증명 제출 실패",
                    description:
                      "서류 제출 중 오류가 발생했습니다. 파일을 확인하고 다시 시도해주세요.",
                    placement: "topRight",
                    duration: 4,
                  });

                  message.error("제출 실패");
                }
              },
            },
          ]}
          renderItem={(item) => (
            <List.Item
              actions={[
                item.status ? (
                  <Tag color="green" icon={<CheckCircleOutlined />}>
                    인증 완료
                  </Tag>
                ) : (
                  <Button
                    icon={<UploadOutlined />}
                    size="small"
                    onClick={item.action}
                  >
                    파일 업로드
                  </Button>
                ),
              ]}
            >
              <List.Item.Meta
                avatar={item.icon}
                title={item.title}
                description={
                  <div>
                    <div>{item.description}</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      상태: {item.statusText}
                    </Text>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Card>

      <Card
        title={
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <TrophyOutlined style={{ color: "var(--purple)" }} />
            <span>레벨 4: 온체인 평판 (자동 획득)</span>
            <Tag color="purple">자동 생성</Tag>
          </div>
        }
        loading={loading}
      >
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          The Blueprint 내에서의 활동을 통해 쌓이는, 가장 본질적이고 위조
          불가능한 신뢰도입니다.
        </Paragraph>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: 16,
          }}
        >
          <Card size="small" style={{ textAlign: "center" }}>
            <TrophyOutlined
              style={{ fontSize: 24, color: "var(--gold)", marginBottom: 8 }}
            />
            <div style={{ fontWeight: 600 }}>획득한 SBT</div>
            <div style={{ fontSize: 20, color: "var(--blue)" }}>
              {verificationStatus.level4.sbtCount}개
            </div>
          </Card>
          <Card size="small" style={{ textAlign: "center" }}>
            <CheckCircleOutlined
              style={{ fontSize: 24, color: "var(--green)", marginBottom: 8 }}
            />
            <div style={{ fontWeight: 600 }}>프로젝트 성공률</div>
            <div style={{ fontSize: 20, color: "var(--green)" }}>
              {verificationStatus.level4.projectSuccessRate}%
            </div>
          </Card>
          <Card size="small" style={{ textAlign: "center" }}>
            <TeamOutlined
              style={{ fontSize: 24, color: "var(--blue)", marginBottom: 8 }}
            />
            <div style={{ fontWeight: 600 }}>멘토링 성공률</div>
            <div style={{ fontSize: 20, color: "var(--blue)" }}>
              {verificationStatus.level4.mentoringSuccessRate}%
            </div>
          </Card>
        </div>
      </Card>
    </div>
  );

  // 기존 프로필 설정 탭 렌더링
  const renderProfileTab = () => (
    <div style={{ maxWidth: 600, margin: "0 auto" }}>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSave}
        initialValues={{
          displayName: user?.displayName || "",
          email: user?.email || "",
          bio: user?.bio || "",
        }}
      >
        <Card title="프로필 정보" style={{ marginBottom: 24 }}>
          <Form.Item
            label="표시 이름"
            name="displayName"
            rules={[{ required: true, message: "표시 이름을 입력해주세요" }]}
          >
            <Input placeholder="다른 사용자들에게 보여질 이름" />
          </Form.Item>
          <Form.Item
            label="이메일 주소"
            name="email"
            rules={[
              { required: true, message: "이메일을 입력해주세요" },
              { type: "email", message: "올바른 이메일 형식이 아닙니다" },
            ]}
          >
            <Input placeholder="your@email.com" disabled />
          </Form.Item>
          <Form.Item
            label="자기소개"
            name="bio"
            rules={[
              { max: 500, message: "자기소개는 500자 이내로 작성해주세요" },
            ]}
          >
            <TextArea
              rows={4}
              placeholder="당신의 목표나 전문 분야에 대해 간단히 소개해주세요"
              showCount
              maxLength={500}
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={loading}
              disabled={loading}
              className="btn-primary"
              style={{
                height: "40px",
                fontSize: "14px",
                fontWeight: "500",
                padding: "0 16px",
              }}
            >
              {loading ? "저장 중..." : "변경사항 저장"}
            </Button>
          </Form.Item>
        </Card>
      </Form>

      <Card title="알림 설정" style={{ marginBottom: 24 }}>
        <Space direction="vertical" style={{ width: "100%" }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div style={{ fontWeight: 600 }}>이메일 알림</div>
              <Text type="secondary">
                프로젝트 업데이트 및 중요한 소식을 이메일로 받습니다
              </Text>
            </div>
            <Switch
              checked={emailNoti}
              onChange={(v) => handleToggle("email_notifications", v)}
            />
          </div>
          <Divider style={{ margin: "12px 0" }} />
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div style={{ fontWeight: 600 }}>푸시 알림</div>
              <Text type="secondary">실시간 알림을 브라우저에서 받습니다</Text>
            </div>
            <Switch
              checked={pushNoti}
              onChange={(v) => handleToggle("push_notifications", v)}
            />
          </div>
          <Divider style={{ margin: "12px 0" }} />
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div style={{ fontWeight: 600 }}>마케팅 알림</div>
              <Text type="secondary">
                새로운 기능 및 이벤트 소식을 받습니다
              </Text>
            </div>
            <Switch
              checked={marketingNoti}
              onChange={(v) => handleToggle("marketing_notifications", v)}
            />
          </div>
        </Space>
      </Card>

      <Card title="프라이버시 설정" style={{ marginBottom: 24 }}>
        <Space direction="vertical" style={{ width: "100%" }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div style={{ fontWeight: 600 }}>프로필 공개</div>
              <Text type="secondary">
                다른 사용자들이 내 프로필을 볼 수 있습니다
              </Text>
            </div>
            <Switch
              checked={profilePublic}
              onChange={(v) => handleToggle("profile_public", v)}
            />
          </div>
          <Divider style={{ margin: "12px 0" }} />
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div style={{ fontWeight: 600 }}>투자 내역 공개</div>
              <Text type="secondary">
                내 투자 내역을 다른 사용자들에게 공개합니다
              </Text>
            </div>
            <Switch
              checked={investmentPublic}
              onChange={(v) => handleToggle("investment_public", v)}
            />
          </div>
        </Space>
      </Card>

      <Card title="계정 관리">
        <Space direction="vertical" style={{ width: "100%" }}>
          <Button icon={<EditOutlined />}>비밀번호 변경</Button>
          <Button icon={<UserOutlined />}>연결된 계정 관리</Button>
          <Divider />
          <Button danger icon={<ExclamationCircleOutlined />}>
            계정 삭제
          </Button>
          <Text type="secondary" style={{ fontSize: 12 }}>
            계정을 삭제하면 모든 데이터가 영구적으로 삭제되며 복구할 수
            없습니다.
          </Text>
        </Space>
      </Card>
    </div>
  );

  return (
    <Layout style={{ minHeight: "100vh", background: "var(--bg-primary)" }}>
      {/* Header - 기존 네비게이션과 동일 */}
      <Header
        style={{
          background: "var(--bg-secondary)",
          boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
          padding: "0 24px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          borderBottom: "1px solid var(--border-color)",
          position: "sticky",
          top: 0,
          zIndex: 1000,
        }}
      >
        {/* Logo */}
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <Title level={3} style={{ margin: 0, color: "var(--blue)" }}>
            <RocketOutlined /> The Blueprint
          </Title>
        </div>

        {/* Central Search */}
        <div style={{ flex: 1, maxWidth: 500, margin: "0 40px" }}>
          <Search
            placeholder="프로젝트, 마일스톤, 멘토 검색..."
            allowClear
            size="large"
            prefix={<SearchOutlined />}
            style={{ width: "100%" }}
          />
        </div>

        {/* Right Navigation */}
        <Space size="middle">
          <Button
            type="primary"
            icon={<RocketOutlined />}
            size="large"
            onClick={() => navigate("/create-project")}
            style={{
              background: "linear-gradient(135deg, #1890ff 0%, #722ed1 100%)",
              border: "none",
              borderRadius: 8,
            }}
          >
            새 프로젝트 시작
          </Button>

          <Space size="large">
            <Button
              type="text"
              icon={<ProjectOutlined />}
              onClick={() => navigate("/dashboard")}
            >
              내 프로젝트
            </Button>
            <Button
              type="text"
              icon={<CompassOutlined />}
              onClick={() => navigate("/explore")}
            >
              프로젝트 탐색
            </Button>
            <Button type="text" icon={<TeamOutlined />}>
              멘토링
            </Button>
            <Button type="text" icon={<TrophyOutlined />}>
              명예의 전당
            </Button>
          </Space>

          <Space>
            <Badge count={3} size="small">
              <Button type="text" icon={<BellOutlined />} size="large" />
            </Badge>
            <LanguageSwitcher />
            <Dropdown
              menu={{ items: userMenuItems }}
              placement="bottomRight"
              trigger={["click"]}
            >
              <Avatar
                src={`https://api.dicebear.com/6.x/avataaars/svg?seed=${user?.username}`}
                style={{ cursor: "pointer" }}
              />
            </Dropdown>
          </Space>
        </Space>
      </Header>

      {/* Main Content */}
      <Content
        style={{
          padding: "24px",
          maxWidth: 1200,
          margin: "0 auto",
          width: "100%",
        }}
      >
        <div style={{ marginBottom: 24 }}>
          <Title level={2}>계정 설정</Title>
          <Text type="secondary">
            프로필 정보와 신원 증명을 관리하여 신뢰도를 높이세요
          </Text>
        </div>

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          size="large"
          style={{
            background: "var(--bg-secondary)",
            padding: 16,
            borderRadius: 8,
          }}
        >
          <TabPane
            tab={
              <span>
                <UserOutlined />
                프로필 설정
              </span>
            }
            key="profile"
          >
            {renderProfileTab()}
          </TabPane>
          <TabPane
            tab={
              <span>
                <SecurityScanOutlined />
                신원 증명
              </span>
            }
            key="verification"
          >
            {renderVerificationTab()}
          </TabPane>
        </Tabs>
      </Content>
      {contextHolder}
    </Layout>
  );
};

export default AccountSettingsPage;
