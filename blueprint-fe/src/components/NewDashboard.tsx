import {
  BellOutlined,
  ClockCircleOutlined,
  CompassOutlined,
  DollarOutlined,
  FireOutlined,
  MoonOutlined,
  PlusOutlined,
  ProjectOutlined,
  RiseOutlined,
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
  Divider,
  Dropdown,
  Input,
  Layout,
  List,
  Progress,
  Row,
  Space,
  Statistic,
  Tag,
  Typography,
  type MenuProps,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTheme } from "../hooks/useTheme";
import { useAuthStore } from "../stores/useAuthStore";
import LanguageSwitcher from "./LanguageSwitcher";

const { Header, Content } = Layout;
const { Title, Text, Paragraph } = Typography;
const { Search } = Input;

// Types
interface NextMilestone {
  title: string;
  daysLeft: number;
  progress: number;
  mentorName: string;
  mentorAvatar: string;
}

// Mock data for demonstration
const mockActivityFeed = [
  {
    id: 1,
    type: "mentor_feedback",
    title:
      "멘토 Elon M.님이 '화성 탐사선' 마일스톤에 새로운 피드백을 남겼습니다.",
    time: "2분 전",
    avatar: "https://api.dicebear.com/6.x/avataaars/svg?seed=elon",
  },
  {
    id: 2,
    type: "project_completed",
    title:
      "'카페 창업' 프로젝트가 성공적으로 완료되었습니다! 지금 확인해 보세요.",
    time: "1시간 전",
    avatar: null,
  },
  {
    id: 3,
    type: "probability_change",
    title: "내가 투자한 '앱 개발' 마일스톤의 성공 확률이 65%로 상승했습니다.",
    time: "3시간 전",
    avatar: null,
  },
];

const mockFeaturedProjects = [
  {
    id: 1,
    title: "AI 스타트업 창업",
    creator: "김영훈",
    category: "business",
    currentPrice: 0.45,
    totalInvestment: 125000,
    badge: "펀딩 성공!",
    badgeColor: "green",
  },
  {
    id: 2,
    title: "요가 강사 자격증 취득",
    creator: "박민지",
    category: "personal",
    currentPrice: 0.72,
    totalInvestment: 89000,
    badge: "가파른 성장!",
    badgeColor: "volcano",
  },
  {
    id: 3,
    title: "웹툰 작가 데뷔",
    creator: "이창호",
    category: "career",
    currentPrice: 0.38,
    totalInvestment: 203000,
    badge: "거물의 선택!",
    badgeColor: "gold",
  },
];

const mockTopMentors = [
  {
    id: 1,
    name: "김사업가",
    specialty: "창업·사업",
    successRate: 92,
    avatar: "https://api.dicebear.com/6.x/avataaars/svg?seed=kim",
    achievement: "스타트업 5개 성공 exit",
  },
  {
    id: 2,
    name: "박개발자",
    specialty: "개발·기술",
    successRate: 88,
    avatar: "https://api.dicebear.com/6.x/avataaars/svg?seed=park",
    achievement: "FAANG 시니어 엔지니어",
  },
  {
    id: 3,
    name: "이커리어",
    specialty: "커리어·취업",
    successRate: 95,
    avatar: "https://api.dicebear.com/6.x/avataaars/svg?seed=lee",
    achievement: "헤드헌터 10년 경력",
  },
];

const popularGoalTemplates = [
  { label: "커리어전환", icon: "💼" },
  { label: "창업", icon: "🚀" },
  { label: "사이드프로젝트", icon: "💡" },
  { label: "자격증", icon: "📚" },
  { label: "건강관리", icon: "💪" },
  { label: "투자공부", icon: "💰" },
];

const MissionControlDashboard: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const { isDark, toggleTheme } = useTheme();
  const [nextMilestone, setNextMilestone] = useState<NextMilestone | null>(
    null
  );
  const [portfolio] = useState({
    totalInvested: 5200,
    currentValue: 6150,
    profit: 950,
    profitPercent: 18.2,
    blueprintTokens: 12500,
  });

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      // TODO: Load user's next milestone and portfolio data
      setNextMilestone({
        title: "MVP 개발 완료",
        daysLeft: 35,
        progress: 65,
        mentorName: "박개발자",
        mentorAvatar: "https://api.dicebear.com/6.x/avataaars/svg?seed=mentor",
      });
    } catch (error) {
      console.error("Dashboard data loading failed:", error);
    }
  };

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
            ${(portfolio.totalInvested / 100).toFixed(2)}
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
            {portfolio.blueprintTokens.toLocaleString()}
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
      {/* Mission Control Navigation */}
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
            icon={<PlusOutlined />}
            size="large"
            onClick={() => navigate("/create-project")}
            className="btn-primary"
            style={{
              height: "40px",
              fontSize: "14px",
              fontWeight: "600",
              padding: "0 16px",
            }}
          >
            새 프로젝트 시작
          </Button>

          <Space size="large">
            <Button
              type="text"
              icon={<ProjectOutlined />}
              onClick={() => navigate("/dashboard")}
              className="btn-ghost"
              style={{
                color: "var(--blue)",
                fontWeight: 600,
                fontSize: "14px",
              }}
            >
              내 프로젝트
            </Button>
            <Button
              type="text"
              icon={<CompassOutlined />}
              onClick={() => navigate("/explore")}
              className="btn-ghost"
              style={{
                fontSize: "14px",
                fontWeight: "500",
              }}
            >
              프로젝트 탐색
            </Button>
            <Button
              type="text"
              icon={<TeamOutlined />}
              className="btn-ghost"
              style={{
                fontSize: "14px",
                fontWeight: "500",
              }}
            >
              멘토링
            </Button>
            <Button
              type="text"
              icon={<TrophyOutlined />}
              className="btn-ghost"
              style={{
                fontSize: "14px",
                fontWeight: "500",
              }}
            >
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

      <Content style={{ padding: "24px", background: "var(--bg-primary)" }}>
        <Row gutter={24} style={{ height: "100%" }}>
          {/* Left Column: My Journey */}
          <Col xs={24} lg={14}>
            <Space direction="vertical" size="large" style={{ width: "100%" }}>
              {/* Next Milestone */}
              <Card
                style={{
                  background:
                    "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
                  border: "none",
                  borderRadius: 16,
                  color: "white",
                }}
                bodyStyle={{ padding: 24 }}
              >
                <Row align="middle" gutter={24}>
                  <Col flex="auto">
                    <Space
                      direction="vertical"
                      size="small"
                      style={{ width: "100%" }}
                    >
                      <Text
                        style={{ color: "rgba(255,255,255,0.8)", fontSize: 14 }}
                      >
                        🚀 나의 다음 마일스톤
                      </Text>
                      <Title level={2} style={{ color: "white", margin: 0 }}>
                        {nextMilestone?.title || "설정된 마일스톤이 없습니다"}
                      </Title>
                      {nextMilestone && (
                        <>
                          <Space size="large">
                            <Text style={{ color: "white", fontSize: 16 }}>
                              <ClockCircleOutlined /> D-{nextMilestone.daysLeft}
                            </Text>
                            <Text style={{ color: "white", fontSize: 16 }}>
                              진행률 {nextMilestone.progress}%
                            </Text>
                          </Space>
                          <Progress
                            percent={nextMilestone.progress}
                            strokeColor="white"
                            trailColor="rgba(255,255,255,0.3)"
                            strokeWidth={8}
                            showInfo={false}
                          />
                        </>
                      )}
                    </Space>
                  </Col>
                  {nextMilestone && (
                    <Col>
                      <Space direction="vertical" align="center">
                        <Avatar size={64} src={nextMilestone.mentorAvatar} />
                        <Text
                          style={{
                            color: "rgba(255,255,255,0.9)",
                            fontSize: 12,
                          }}
                        >
                          핵심 멘토
                        </Text>
                        <Text style={{ color: "white", fontWeight: 600 }}>
                          {nextMilestone.mentorName}
                        </Text>
                      </Space>
                    </Col>
                  )}
                </Row>
                <Row gutter={16} style={{ marginTop: 20 }}>
                  <Col span={12}>
                    <Button
                      size="large"
                      block
                      style={{
                        background: "rgba(255,255,255,0.2)",
                        border: "1px solid rgba(255,255,255,0.3)",
                        color: "white",
                      }}
                    >
                      진행 상황 업데이트
                    </Button>
                  </Col>
                  <Col span={12}>
                    <Button
                      size="large"
                      block
                      style={{
                        background: "rgba(255,255,255,0.2)",
                        border: "1px solid rgba(255,255,255,0.3)",
                        color: "white",
                      }}
                    >
                      멘토와 대화하기
                    </Button>
                  </Col>
                </Row>
              </Card>

              {/* Portfolio Summary */}
              <Card
                title={
                  <Space>
                    <DollarOutlined style={{ color: "var(--green)" }} />
                    나의 포트폴리오
                  </Space>
                }
                extra={<Button type="link">자세히 보기 →</Button>}
                style={{ borderRadius: 12 }}
              >
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic
                      title="총 투자액"
                      value={portfolio.totalInvested}
                      prefix="$"
                      suffix="USDC"
                      valueStyle={{ color: "var(--text-primary)" }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="현재 평가액"
                      value={portfolio.currentValue}
                      prefix="$"
                      suffix="USDC"
                      valueStyle={{ color: "var(--green)" }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="수익률"
                      value={portfolio.profitPercent}
                      prefix="+"
                      suffix="%"
                      valueStyle={{ color: "var(--green)" }}
                    />
                  </Col>
                </Row>
                <Divider />
                <Space>
                  <StarOutlined style={{ color: "var(--gold)" }} />
                  <Text strong>보유 토큰:</Text>
                  <Text style={{ color: "var(--blue)", fontWeight: 600 }}>
                    {portfolio.blueprintTokens.toLocaleString()} BLUEPRINT
                  </Text>
                </Space>
              </Card>

              {/* Activity Feed */}
              <Card
                title={
                  <Space>
                    <BellOutlined style={{ color: "var(--orange)" }} />
                    나의 활동 피드
                  </Space>
                }
                style={{ borderRadius: 12 }}
              >
                <List
                  dataSource={mockActivityFeed}
                  renderItem={(item) => (
                    <List.Item style={{ border: "none", padding: "12px 0" }}>
                      <List.Item.Meta
                        avatar={
                          item.avatar ? (
                            <Avatar src={item.avatar} />
                          ) : (
                            <Avatar style={{ background: "var(--blue)" }}>
                              <BellOutlined />
                            </Avatar>
                          )
                        }
                        title={
                          <Text style={{ fontSize: 14, lineHeight: 1.4 }}>
                            {item.title}
                          </Text>
                        }
                        description={
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {item.time}
                          </Text>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            </Space>
          </Col>

          {/* Right Column: Discover Opportunities */}
          <Col xs={24} lg={10}>
            <Space direction="vertical" size="large" style={{ width: "100%" }}>
              {/* Featured Projects */}
              <Card
                title={
                  <Space>
                    <FireOutlined style={{ color: "var(--red)" }} />
                    주목할 만한 프로젝트
                  </Space>
                }
                extra={<Button type="link">더 보기 →</Button>}
                style={{ borderRadius: 12 }}
              >
                <Space
                  direction="vertical"
                  size="middle"
                  style={{ width: "100%" }}
                >
                  {mockFeaturedProjects.map((project) => (
                    <Card
                      key={project.id}
                      size="small"
                      style={{
                        background: "var(--bg-secondary)",
                        border: "1px solid var(--border-color)",
                        borderRadius: 8,
                        cursor: "pointer",
                      }}
                      bodyStyle={{ padding: 16 }}
                      onClick={() => navigate(`/project/${project.id}`)}
                    >
                      <Row justify="space-between" align="middle">
                        <Col flex="auto">
                          <Space direction="vertical" size={4}>
                            <Space>
                              <Text strong>{project.title}</Text>
                              <Tag
                                color={project.badgeColor}
                                style={{ fontSize: 10 }}
                              >
                                {project.badge}
                              </Tag>
                            </Space>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              by {project.creator}
                            </Text>
                            <Space>
                              <Text style={{ fontSize: 12 }}>
                                @ ${project.currentPrice.toFixed(2)} USDC
                              </Text>
                              <Text type="secondary" style={{ fontSize: 12 }}>
                                TVL: $
                                {(project.totalInvestment / 1000).toFixed(0)}K
                              </Text>
                            </Space>
                          </Space>
                        </Col>
                        <Col>
                          <RiseOutlined
                            style={{ color: "var(--green)", fontSize: 16 }}
                          />
                        </Col>
                      </Row>
                    </Card>
                  ))}
                </Space>
              </Card>

              {/* AI Path Recommendation */}
              <Card
                title={
                  <Space>
                    <RocketOutlined style={{ color: "var(--purple)" }} />
                    AI 추천 경로
                  </Space>
                }
                style={{ borderRadius: 12 }}
              >
                <Space
                  direction="vertical"
                  size="middle"
                  style={{ width: "100%" }}
                >
                  <Paragraph
                    style={{ margin: 0, color: "var(--text-secondary)" }}
                  >
                    당신의 다음 목표는 무엇인가요?
                  </Paragraph>
                  <Search
                    placeholder="목표를 입력하면 AI가 경로를 제안해드려요..."
                    enterButton="추천받기"
                    size="large"
                    style={{ width: "100%" }}
                  />
                  <Space wrap>
                    {popularGoalTemplates.map((template) => (
                      <Button
                        key={template.label}
                        size="small"
                        style={{
                          borderRadius: 16,
                          background: "var(--bg-tertiary)",
                          border: "1px solid var(--border-color)",
                        }}
                      >
                        {template.icon} {template.label}
                      </Button>
                    ))}
                  </Space>
                </Space>
              </Card>

              {/* Top Mentors */}
              <Card
                title={
                  <Space>
                    <TrophyOutlined style={{ color: "var(--gold)" }} />
                    이주의 멘토
                  </Space>
                }
                style={{ borderRadius: 12 }}
              >
                <Space
                  direction="vertical"
                  size="middle"
                  style={{ width: "100%" }}
                >
                  {mockTopMentors.map((mentor) => (
                    <Row key={mentor.id} align="middle" gutter={12}>
                      <Col>
                        <Avatar src={mentor.avatar} size={48} />
                      </Col>
                      <Col flex="auto">
                        <Space direction="vertical" size={2}>
                          <Text strong>{mentor.name}</Text>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {mentor.specialty}
                          </Text>
                          <Text style={{ fontSize: 11, color: "var(--green)" }}>
                            성공률 {mentor.successRate}% • {mentor.achievement}
                          </Text>
                        </Space>
                      </Col>
                      <Col>
                        <Button size="small" type="primary" ghost>
                          멘토링
                        </Button>
                      </Col>
                    </Row>
                  ))}
                </Space>
              </Card>
            </Space>
          </Col>
        </Row>
      </Content>
    </Layout>
  );
};

export default MissionControlDashboard;
