import {
  BellOutlined,
  CompassOutlined,
  DollarOutlined,
  LoginOutlined,
  MoonOutlined,
  PlusOutlined,
  ProjectOutlined,
  RocketOutlined,
  SearchOutlined,
  StarOutlined,
  SunOutlined,
  TeamOutlined,
  TrophyOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Dropdown,
  Input,
  Layout,
  Row,
  Space,
  Spin,
  Statistic,
  Tag,
  Typography,
  message,
  type MenuProps,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTheme } from "../hooks/useTheme";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { Project } from "../types";
import AuthModal from "./AuthModal";
import LanguageSwitcher from "./LanguageSwitcher";

const { Header, Content } = Layout;
const { Title, Text } = Typography;
const { Search } = Input;

const ProjectExplorePage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated, user, logout } = useAuthStore();
  const { isDark, toggleTheme } = useTheme();

  // 상태 관리
  const [authModalVisible, setAuthModalVisible] = useState(false);
  const [loading, setLoading] = useState(true);
  const [projects, setProjects] = useState<Project[]>([]);
  const [stats, setStats] = useState({
    totalProjects: 0,
    totalInvestors: 0,
    totalInvestment: 0,
  });

  useEffect(() => {
    // 로그인된 사용자에게만 프로젝트 로드
    if (isAuthenticated) {
      loadPublicProjects();
      console.log("🏠 HomePage useEffect 실행됨 - 인증된 사용자");
    } else {
      console.log("🏠 HomePage useEffect 실행됨 - 미인증 사용자");
      setLoading(false); // 미인증 사용자는 로딩 해제
    }
  }, [isAuthenticated]);

  const loadPublicProjects = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getProjects({
        page: 1,
        limit: 20,
        sort: "created_at",
        order: "desc",
      });

      if (response.success && response.data) {
        const projectsData = response.data.projects || [];
        setProjects(projectsData);
        setStats({
          totalProjects: projectsData.length,
          totalInvestors: 150, // Mock data
          totalInvestment: 2500000, // Mock data
        });
      }
    } catch (error) {
      console.error("Failed to load projects:", error);
      message.error("프로젝트를 불러오는데 실패했습니다");
    } finally {
      setLoading(false);
    }
  };

  const getCategoryLabel = (category: string) => {
    const categoryMap: { [key: string]: string } = {
      career: "💼 커리어",
      business: "🚀 비즈니스",
      education: "📚 교육",
      personal: "🌱 개인",
      life: "🏡 라이프",
    };
    return categoryMap[category] || category;
  };

  const calculateProgress = (project: Project): number => {
    if (!project.milestones || project.milestones.length === 0) return 0;
    const completed = project.milestones.filter(
      (m) => m.status === "completed"
    ).length;
    return Math.round((completed / project.milestones.length) * 100);
  };

  const calculateTotalInvestment = (): number => {
    // Mock calculation - replace with actual investment data
    return Math.floor(Math.random() * 500000) + 50000;
  };

  const calculateInvestorCount = (): number => {
    // Mock calculation - replace with actual investor data
    return Math.floor(Math.random() * 50) + 5;
  };

  const calculateTimeLeft = (targetDate?: string | null): string => {
    if (!targetDate) return "목표일 미설정";
    const target = new Date(targetDate);
    const now = new Date();
    const diffTime = target.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays < 0) return "목표일 경과";
    if (diffDays === 0) return "오늘";
    if (diffDays <= 30) return `${diffDays}일 남음`;
    if (diffDays <= 365) return `${Math.ceil(diffDays / 30)}개월 남음`;

    return `${Math.ceil(diffDays / 365)}년`;
  };

  // User menu items for authenticated users
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

  return (
    <Layout style={{ minHeight: "100vh", background: "var(--bg-primary)" }}>
      {/* Header */}
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
          {isAuthenticated ? (
            <Button
              type="primary"
              icon={<PlusOutlined />}
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
          ) : (
            <Button
              type="primary"
              icon={<LoginOutlined />}
              size="large"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setAuthModalVisible(true);
              }}
              style={{
                background: "linear-gradient(135deg, #1890ff 0%, #722ed1 100%)",
                border: "none",
                borderRadius: 8,
              }}
            >
              로그인하여 시작하기
            </Button>
          )}

          <Space size="large">
            {isAuthenticated && (
              <Button
                type="text"
                icon={<ProjectOutlined />}
                onClick={() => navigate("/dashboard")}
              >
                내 프로젝트
              </Button>
            )}
            <Button
              type="text"
              icon={<CompassOutlined />}
              onClick={() => navigate("/explore")}
              style={{ color: "var(--blue)", fontWeight: 600 }}
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
            {isAuthenticated && (
              <Badge count={3} size="small">
                <Button type="text" icon={<BellOutlined />} size="large" />
              </Badge>
            )}

            <LanguageSwitcher />

            {isAuthenticated ? (
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
            ) : (
              <Space>
                <Button
                  type="text"
                  icon={isDark ? <SunOutlined /> : <MoonOutlined />}
                  onClick={toggleTheme}
                >
                  {isDark ? "라이트" : "다크"}
                </Button>
                <Button
                  type="text"
                  icon={<LoginOutlined />}
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setAuthModalVisible(true);
                  }}
                  className="btn-ghost"
                  style={{
                    fontSize: "14px",
                    fontWeight: "500",
                  }}
                >
                  로그인
                </Button>
              </Space>
            )}
          </Space>
        </Space>
      </Header>

      <Content
        style={{ padding: "40px 24px", background: "var(--bg-primary)" }}
      >
        <div style={{ maxWidth: 1200, margin: "0 auto" }}>
          {/* Hero Section */}
          <div style={{ textAlign: "center", marginBottom: 60 }}>
            <Title level={1} style={{ fontSize: 48, marginBottom: 16 }}>
              🚀 프로젝트 탐색
            </Title>
            <Text style={{ fontSize: 18, color: "var(--text-secondary)" }}>
              다양한 목표를 가진 사람들의 여정을 발견하고, 성공에 베팅해보세요
            </Text>
          </div>

          {/* Stats Cards */}
          <Row gutter={[24, 24]} style={{ marginBottom: 40 }}>
            <Col xs={24} sm={8}>
              <Card style={{ textAlign: "center", borderRadius: 12 }}>
                <Statistic
                  title="활성 프로젝트"
                  value={stats.totalProjects}
                  suffix="개"
                  valueStyle={{ color: "var(--blue)" }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card style={{ textAlign: "center", borderRadius: 12 }}>
                <Statistic
                  title="참여한 투자자"
                  value={stats.totalInvestors}
                  suffix="명"
                  valueStyle={{ color: "var(--green)" }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card style={{ textAlign: "center", borderRadius: 12 }}>
                <Statistic
                  title="총 투자금액"
                  value={stats.totalInvestment}
                  prefix="₩"
                  suffix="원"
                  valueStyle={{ color: "var(--purple)" }}
                />
              </Card>
            </Col>
          </Row>

          {/* Projects Grid */}
          {!isAuthenticated ? (
            <div style={{ textAlign: "center", padding: "100px 0" }}>
              <div style={{ marginBottom: 24 }}>
                <RocketOutlined
                  style={{ fontSize: 48, color: "var(--blue)" }}
                />
              </div>
              <Title level={3} style={{ color: "var(--text-primary)" }}>
                프로젝트를 탐색하려면 로그인하세요
              </Title>
              <Text style={{ color: "var(--text-secondary)", fontSize: 16 }}>
                다양한 프로젝트에 참여하고 투자할 수 있습니다
              </Text>
              <div style={{ marginTop: 24 }}>
                <Button
                  type="primary"
                  size="large"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setAuthModalVisible(true);
                  }}
                  className="btn-primary"
                  style={{
                    height: "56px",
                    fontSize: "18px",
                    fontWeight: "600",
                    padding: "0 32px",
                  }}
                >
                  로그인하여 시작하기
                </Button>
              </div>
            </div>
          ) : loading ? (
            <div style={{ textAlign: "center", padding: "100px 0" }}>
              <Spin size="large" />
              <div style={{ marginTop: 16, color: "var(--text-secondary)" }}>
                프로젝트를 불러오는 중...
              </div>
            </div>
          ) : (
            <Row gutter={[24, 24]}>
              {projects.map((project) => {
                const progress = calculateProgress(project);
                const investment = calculateTotalInvestment();
                const investors = calculateInvestorCount();
                const timeLeft = calculateTimeLeft(project.target_date);

                return (
                  <Col xs={24} sm={12} lg={8} key={project.id}>
                    <Card
                      hoverable
                      style={{
                        borderRadius: 16,
                        border: "1px solid var(--border-color)",
                        height: "100%",
                        transition: "all 0.3s ease",
                      }}
                      bodyStyle={{ padding: 20 }}
                      onClick={() => navigate(`/project/${project.id}`)}
                    >
                      <Space
                        direction="vertical"
                        size="middle"
                        style={{ width: "100%" }}
                      >
                        {/* Header */}
                        <div>
                          <Tag color="blue" style={{ marginBottom: 8 }}>
                            {getCategoryLabel(project.category)}
                          </Tag>
                          <Title
                            level={4}
                            style={{ margin: 0, lineHeight: 1.3 }}
                          >
                            {project.title}
                          </Title>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {timeLeft}
                          </Text>
                        </div>

                        {/* Description */}
                        <Text
                          style={{
                            color: "var(--text-secondary)",
                            fontSize: 14,
                            lineHeight: 1.4,
                          }}
                        >
                          {project.description?.substring(0, 80)}
                          {project.description &&
                          project.description.length > 80
                            ? "..."
                            : ""}
                        </Text>

                        {/* Stats */}
                        <Row gutter={16}>
                          <Col span={8}>
                            <Statistic
                              title="진행률"
                              value={progress}
                              suffix="%"
                              valueStyle={{
                                fontSize: 14,
                                color: "var(--success)",
                              }}
                            />
                          </Col>
                          <Col span={8}>
                            <Statistic
                              title="투자자"
                              value={investors}
                              suffix="명"
                              valueStyle={{
                                fontSize: 14,
                                color: "var(--blue)",
                              }}
                            />
                          </Col>
                          <Col span={8}>
                            <Statistic
                              title="투자금"
                              value={investment}
                              prefix="₩"
                              valueStyle={{
                                fontSize: 14,
                                color: "var(--purple)",
                              }}
                            />
                          </Col>
                        </Row>
                      </Space>
                    </Card>
                  </Col>
                );
              })}
            </Row>
          )}

          {!loading && projects.length === 0 && (
            <div style={{ textAlign: "center", padding: "100px 0" }}>
              <Text type="secondary" style={{ fontSize: 16 }}>
                아직 등록된 프로젝트가 없습니다.
              </Text>
              {isAuthenticated && (
                <div style={{ marginTop: 16 }}>
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => navigate("/create-project")}
                  >
                    첫 번째 프로젝트 만들기
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </Content>

      {/* Auth Modal */}
      <AuthModal
        isOpen={authModalVisible}
        onClose={() => setAuthModalVisible(false)}
      />
    </Layout>
  );
};

export default ProjectExplorePage;
