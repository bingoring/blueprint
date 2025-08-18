import {
  BellOutlined,
  LogoutOutlined,
  SearchOutlined,
  SettingOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Badge,
  Button,
  Dropdown,
  Input,
  Layout,
  Menu,
  Space,
  Typography,
} from "antd";
import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuthStore } from "../stores/useAuthStore";
import {
  CompassIcon,
  ConnectionIcon,
  DashboardIcon,
  MentoringIcon,
  PortfolioIcon,
  RocketIcon,
  TrophyIcon,
} from "./icons/BlueprintIcons";

const { Header } = Layout;
const { Text } = Typography;

interface GlobalNavbarProps {
  className?: string;
}

const GlobalNavbar: React.FC<GlobalNavbarProps> = ({ className = "" }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();

  // 현재 경로에 따른 활성 메뉴 키 결정
  const getActiveKey = () => {
    const pathname = location.pathname;
    if (pathname === "/" || pathname === "/dashboard") return "home";
    if (pathname.startsWith("/explore")) return "explore";
    if (pathname.startsWith("/activity")) return "activity";
    if (pathname.startsWith("/mentoring")) return "mentoring";
    if (pathname.startsWith("/hall-of-fame")) return "hall-of-fame";
    return "";
  };

  // 메뉴 아이템 클릭 핸들러
  const handleMenuClick = (key: string) => {
    switch (key) {
      case "home":
        navigate("/");
        break;
      case "explore":
        navigate("/explore");
        break;
      case "activity":
        navigate("/activity");
        break;
      case "mentoring":
        navigate("/mentoring");
        break;
      case "hall-of-fame":
        navigate("/hall-of-fame");
        break;
    }
  };

  // 프로필 드롭다운 메뉴
  const profileMenuItems = [
    {
      key: "profile",
      icon: <UserOutlined />,
      label: "내 프로필",
      onClick: () => navigate("/profile"),
    },
    {
      key: "settings",
      icon: <SettingOutlined />,
      label: "계정 설정",
      onClick: () => navigate("/settings"),
    },
    {
      type: "divider" as const,
    },
    {
      key: "logout",
      icon: <LogoutOutlined />,
      label: "로그아웃",
      onClick: () => {
        logout();
        navigate("/");
      },
    },
  ];

  // 메인 네비게이션 메뉴 아이템
  const mainMenuItems = [
    {
      key: "home",
      icon: <DashboardIcon size={20} />,
      label: "홈",
    },
    {
      key: "explore",
      icon: <CompassIcon size={20} />,
      label: "프로젝트 탐색",
    },
    {
      key: "activity",
      icon: <PortfolioIcon size={20} />,
      label: "내 활동",
    },
    {
      key: "mentoring",
      icon: <MentoringIcon size={20} />,
      label: "멘토링",
    },
    {
      key: "hall-of-fame",
      icon: <TrophyIcon size={20} />,
      label: "명예의 전당",
    },
  ];

  return (
    <Header
      className={`${className} fixed top-0 left-0 right-0 z-50`}
      style={{
        background: "var(--bg-primary)",
        borderBottom: "1px solid var(--border-color)",
        padding: "0 24px",
        height: "64px",
        display: "flex",
        alignItems: "center",
        boxShadow: "0 2px 8px rgba(0, 0, 0, 0.06)",
      }}
    >
      <div
        style={{
          width: "100%",
          maxWidth: "1400px",
          margin: "0 auto",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        {/* 로고 및 서비스명 */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            cursor: "pointer",
          }}
          onClick={() => navigate("/")}
        >
          <ConnectionIcon
            size={32}
            color="var(--primary-color)"
            className="mr-3"
          />
          <Text
            style={{
              fontSize: "20px",
              fontWeight: "bold",
              color: "var(--text-primary)",
              letterSpacing: "-0.5px",
            }}
          >
            Blueprint
          </Text>
        </div>

        {/* 메인 네비게이션 메뉴 */}
        <Menu
          mode="horizontal"
          selectedKeys={[getActiveKey()]}
          items={mainMenuItems}
          onClick={({ key }) => handleMenuClick(key)}
          style={{
            background: "transparent",
            border: "none",
            fontSize: "14px",
            fontWeight: "500",
            minWidth: "500px",
            justifyContent: "center",
          }}
          className="blueprint-main-menu"
        />

        {/* 우측 액션 영역 */}
        <div style={{ display: "flex", alignItems: "center", gap: "16px" }}>
          {/* 새 프로젝트 시작 버튼 */}
          <Button
            type="primary"
            icon={<RocketIcon size={16} />}
            onClick={() => navigate("/projects/new")}
            style={{
              background: "var(--primary-color)",
              borderColor: "var(--primary-color)",
              height: "40px",
              borderRadius: "8px",
              fontWeight: "600",
              display: "flex",
              alignItems: "center",
              gap: "8px",
            }}
            className="blueprint-cta-button"
          >
            새 프로젝트 시작
          </Button>

          {/* 검색창 */}
          <Input
            placeholder="프로젝트 검색..."
            prefix={
              <SearchOutlined style={{ color: "var(--text-secondary)" }} />
            }
            style={{
              width: "240px",
              height: "36px",
              borderRadius: "6px",
              backgroundColor: "var(--bg-secondary)",
              border: "1px solid var(--border-color)",
            }}
            onPressEnter={(e) => {
              const value = (e.target as HTMLInputElement).value;
              if (value.trim()) {
                navigate(`/explore?search=${encodeURIComponent(value.trim())}`);
              }
            }}
          />

          {/* 알림 */}
          <Badge count={3} size="small">
            <Button
              type="text"
              icon={<BellOutlined />}
              style={{
                width: "36px",
                height: "36px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "var(--text-secondary)",
              }}
              onClick={() => navigate("/notifications")}
            />
          </Badge>

          {/* 프로필 드롭다운 */}
          <Dropdown
            menu={{ items: profileMenuItems }}
            placement="bottomRight"
            arrow
          >
            <Space
              style={{
                cursor: "pointer",
                padding: "4px 8px",
                borderRadius: "8px",
                transition: "background-color 0.2s",
              }}
              className="blueprint-profile-menu"
            >
              <Avatar
                size={32}
                icon={<UserOutlined />}
                style={{
                  backgroundColor: "var(--primary-color)",
                }}
              />
              <Text
                style={{
                  color: "var(--text-primary)",
                  fontSize: "14px",
                  fontWeight: "500",
                }}
              >
                {user?.username || "사용자"}
              </Text>
            </Space>
          </Dropdown>
        </div>
      </div>

      {/* 커스텀 스타일 */}
      <style>{`
        .blueprint-main-menu .ant-menu-item {
          color: var(--text-secondary) !important;
          border-bottom: 2px solid transparent !important;
          margin: 0 8px !important;
          padding: 12px 16px !important;
          height: 44px !important;
          display: flex !important;
          align-items: center !important;
          gap: 8px !important;
          transition: all 0.2s ease !important;
        }

        .blueprint-main-menu .ant-menu-item:hover {
          color: var(--primary-color) !important;
          background: var(--bg-hover) !important;
          border-radius: 8px !important;
        }

        .blueprint-main-menu .ant-menu-item-selected {
          color: var(--primary-color) !important;
          border-bottom: 2px solid var(--primary-color) !important;
          background: var(--primary-color-light) !important;
          border-radius: 8px !important;
        }

        .blueprint-cta-button:hover {
          transform: translateY(-1px) !important;
          box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3) !important;
        }

        .blueprint-profile-menu:hover {
          background: var(--bg-hover) !important;
        }

        @media (max-width: 1200px) {
          .blueprint-main-menu {
            min-width: 400px !important;
          }
        }

        @media (max-width: 768px) {
          .blueprint-main-menu .ant-menu-item span {
            display: none !important;
          }
          .blueprint-main-menu {
            min-width: 250px !important;
          }
        }
      `}</style>
    </Header>
  );
};

export default GlobalNavbar;
