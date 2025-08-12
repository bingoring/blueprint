import {
  BellOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  CompassOutlined,
  DollarOutlined,
  EditOutlined,
  EyeOutlined,
  MailOutlined,
  MoonOutlined,
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
  List,
  Progress,
  Rate,
  Row,
  Space,
  Spin,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
  type MenuProps,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTheme } from "../hooks/useTheme";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { ProfileResponse } from "../types";
import LanguageSwitcher from "./LanguageSwitcher";

const { Header, Content } = Layout;
const { Title, Text, Paragraph } = Typography;
const { Search } = Input;

// Mock data - 멘토링과 투자 데이터 (아직 미구현)
const mockMentoringData = [
  {
    id: 1,
    projectTitle: "AI 스타트업 창업",
    mentee: "김영훈",
    status: "completed",
    rating: 5,
    review: "Elon님 덕분에 로켓이 폭발하지 않았어요! 최고의 멘토입니다.",
    completedDate: "2024-02-15",
  },
  {
    id: 2,
    projectTitle: "전기차 프로토타입 개발",
    mentee: "박민수",
    status: "active",
    rating: null,
    review: null,
    startedDate: "2024-03-01",
  },
];

const mockInvestmentData = [
  {
    id: 1,
    projectTitle: "화성 탐사 로봇",
    milestone: "프로토타입 완성",
    option: "성공",
    amount: 50000,
    result: "진행중",
    profit: 0,
    date: "2024-03-15",
  },
  {
    id: 2,
    projectTitle: "AI 의료진단 시스템",
    milestone: "임상시험 통과",
    option: "성공",
    amount: 25000,
    result: "성공",
    profit: 12500,
    date: "2024-02-20",
  },
];

const ProfilePage: React.FC = () => {
  const navigate = useNavigate();
  const { username } = useParams<{ username: string }>();
  const { user, logout } = useAuthStore();
  const { isDark, toggleTheme } = useTheme();
  const [activeTab, setActiveTab] = useState("overview");
  const [profileData, setProfileData] = useState<ProfileResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const isOwnProfile = user?.username === username;

  useEffect(() => {
    const loadProfileData = async () => {
      if (!username) return;

      try {
        setLoading(true);
        setError(null);

        console.log("🔄 프로필 데이터 로딩 중...", username);
        const response = await apiClient.getUserProfile(username);

        if (response.success && response.data) {
          console.log("✅ 프로필 데이터 로딩 성공:", response.data);
          setProfileData(response.data);
        } else {
          console.error("❌ 프로필 데이터 로딩 실패:", response);
          setError("프로필을 불러올 수 없습니다.");
        }
      } catch (error) {
        console.error("❌ 프로필 API 호출 오류:", error);
        setError("프로필을 불러오는 중 오류가 발생했습니다.");
      } finally {
        setLoading(false);
      }
    };

    loadProfileData();
  }, [username]);

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

  const renderOverviewTab = () => (
    <Row gutter={[24, 24]}>
      <Col xs={24} lg={12}>
        {/* 진행 중인 프로젝트 */}
        <Card title="🚀 진행 중인 프로젝트" style={{ marginBottom: 24 }}>
          <Space direction="vertical" size="middle" style={{ width: "100%" }}>
            {(profileData?.currentProjects || []).map((project) => (
              <Card
                key={project.id}
                size="small"
                style={{ backgroundColor: "var(--bg-secondary)" }}
              >
                <Row justify="space-between" align="middle">
                  <Col>
                    <Space direction="vertical" size="small">
                      <Text strong>{project.title}</Text>
                      <Tag color="blue">{project.category}</Tag>
                    </Space>
                  </Col>
                  <Col>
                    <Progress
                      type="circle"
                      percent={project.progress}
                      size={50}
                      strokeColor="#52c41a"
                    />
                  </Col>
                </Row>
              </Card>
            ))}
            {(!profileData?.currentProjects ||
              profileData.currentProjects.length === 0) && (
              <Text type="secondary">진행 중인 프로젝트가 없습니다.</Text>
            )}
          </Space>
        </Card>

        {/* 대표 프로젝트 */}
        <Card title="🏆 대표 프로젝트">
          <Space direction="vertical" size="middle" style={{ width: "100%" }}>
            {(profileData?.featuredProjects || []).map((project) => (
              <Card
                key={project.id}
                size="small"
                style={{ backgroundColor: "var(--bg-secondary)" }}
              >
                <Row justify="space-between" align="middle">
                  <Col flex="auto">
                    <Space direction="vertical" size="small">
                      <Text strong>{project.title}</Text>
                      <Text type="secondary">{project.description}</Text>
                      <Space>
                        <Tag color="green">완료</Tag>
                        <Text>
                          <DollarOutlined />{" "}
                          {(project.investment / 100).toLocaleString()} USDC
                        </Text>
                        <Text>성공률 {project.successRate}%</Text>
                      </Space>
                    </Space>
                  </Col>
                </Row>
              </Card>
            ))}
            {(!profileData?.featuredProjects ||
              profileData.featuredProjects.length === 0) && (
              <Text type="secondary">완료된 프로젝트가 없습니다.</Text>
            )}
          </Space>
        </Card>
      </Col>

      <Col xs={24} lg={12}>
        {/* 최근 활동 */}
        <Card title="📈 최근 활동">
          <List
            dataSource={profileData?.recentActivities || []}
            renderItem={(activity) => (
              <List.Item style={{ border: "none", padding: "8px 0" }}>
                <Space>
                  <Avatar size="small" style={{ backgroundColor: "#1890ff" }}>
                    {activity.type === "investment" && <DollarOutlined />}
                    {activity.type === "milestone" && <CheckCircleOutlined />}
                    {activity.type === "project" && <ProjectOutlined />}
                    {activity.type === "mentoring" && <TeamOutlined />}
                  </Avatar>
                  <div>
                    <Text>{activity.description}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {activity.timestamp}
                    </Text>
                  </div>
                </Space>
              </List.Item>
            )}
          />
          {(!profileData?.recentActivities ||
            profileData.recentActivities.length === 0) && (
            <Text type="secondary">최근 활동이 없습니다.</Text>
          )}
        </Card>
      </Col>
    </Row>
  );

  const renderProjectsTab = () => (
    <div>
      <Row gutter={[16, 16]}>
        {[
          ...(profileData?.currentProjects || []),
          ...(profileData?.featuredProjects || []),
        ].map((project) => (
          <Col xs={24} sm={12} lg={8} key={project.id}>
            <Card
              hoverable
              style={{ height: "100%" }}
              bodyStyle={{ padding: 16 }}
            >
              <Space
                direction="vertical"
                size="small"
                style={{ width: "100%" }}
              >
                <Text strong>{project.title}</Text>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {"description" in project
                    ? project.description
                    : `진행률 ${project.progress}%`}
                </Text>
                <div>
                  <Tag
                    color={project.status === "completed" ? "green" : "blue"}
                  >
                    {project.status === "completed" ? "완료" : "진행중"}
                  </Tag>
                  {"category" in project && <Tag>{project.category}</Tag>}
                </div>
                {"progress" in project && (
                  <Progress percent={project.progress} size="small" />
                )}
              </Space>
            </Card>
          </Col>
        ))}
      </Row>
      {!profileData?.currentProjects?.length &&
        !profileData?.featuredProjects?.length && (
          <div style={{ textAlign: "center", padding: "40px 0" }}>
            <Text type="secondary">프로젝트가 없습니다.</Text>
          </div>
        )}
    </div>
  );

  const renderMentoringTab = () => (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" ghost>
          전체
        </Button>
        <Button>진행 중</Button>
        <Button>완료</Button>
      </Space>
      <Row gutter={[16, 16]}>
        {mockMentoringData.map((item) => (
          <Col xs={24} lg={12} key={item.id}>
            <Card>
              <Space
                direction="vertical"
                size="middle"
                style={{ width: "100%" }}
              >
                <div
                  style={{ display: "flex", justifyContent: "space-between" }}
                >
                  <Text strong>{item.projectTitle}</Text>
                  <Tag color={item.status === "completed" ? "green" : "blue"}>
                    {item.status === "completed" ? "완료" : "진행중"}
                  </Tag>
                </div>
                <div>
                  <Text type="secondary">멘티: </Text>
                  <Text>{item.mentee}</Text>
                </div>
                {item.rating && (
                  <>
                    <Rate
                      disabled
                      value={item.rating}
                      style={{ fontSize: 16 }}
                    />
                    <Paragraph
                      style={{
                        background: "var(--bg-tertiary)",
                        padding: 12,
                        borderRadius: 8,
                        margin: 0,
                        fontStyle: "italic",
                      }}
                    >
                      "{item.review}"
                    </Paragraph>
                  </>
                )}
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {item.status === "completed"
                    ? `완료일: ${item.completedDate}`
                    : `시작일: ${item.startedDate}`}
                </Text>
              </Space>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );

  const renderInvestmentTab = () => {
    const columns = [
      {
        title: "프로젝트/마일스톤",
        dataIndex: "projectTitle",
        key: "project",
        render: (text: string, record: (typeof mockInvestmentData)[0]) => (
          <div>
            <div style={{ fontWeight: 600 }}>{text}</div>
            <div style={{ fontSize: 12, color: "var(--text-secondary)" }}>
              {record.milestone}
            </div>
          </div>
        ),
      },
      {
        title: "베팅 옵션",
        dataIndex: "option",
        key: "option",
        render: (option: string) => (
          <Tag color={option === "성공" ? "green" : "red"}>{option}</Tag>
        ),
      },
      {
        title: "베팅액",
        dataIndex: "amount",
        key: "amount",
        render: (amount: number) => `$${(amount / 100).toFixed(2)}`,
      },
      {
        title: "결과",
        dataIndex: "result",
        key: "result",
        render: (result: string) => (
          <Tag
            color={
              result === "성공" ? "green" : result === "실패" ? "red" : "blue"
            }
          >
            {result}
          </Tag>
        ),
      },
      {
        title: "수익(손실)",
        dataIndex: "profit",
        key: "profit",
        render: (profit: number) => (
          <Text
            style={{
              color:
                profit > 0
                  ? "var(--green)"
                  : profit < 0
                  ? "var(--red)"
                  : "var(--text-primary)",
            }}
          >
            {profit > 0 ? "+" : ""}${(profit / 100).toFixed(2)}
          </Text>
        ),
      },
      {
        title: "날짜",
        dataIndex: "date",
        key: "date",
      },
    ];

    return (
      <div>
        <Space
          style={{
            marginBottom: 16,
            justifyContent: "space-between",
            width: "100%",
          }}
        >
          <div>
            <Text type="secondary">
              <EyeOutlined /> 이 정보는 본인만 볼 수 있습니다
            </Text>
          </div>
          <Button size="small">공개 설정 변경</Button>
        </Space>
        <Table
          columns={columns}
          dataSource={mockInvestmentData}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </div>
    );
  };

  const tabItems = [
    {
      key: "overview",
      label: "개요",
      children: renderOverviewTab(),
    },
    {
      key: "projects",
      label: "프로젝트",
      children: renderProjectsTab(),
    },
    {
      key: "mentoring",
      label: "멘토링",
      children: renderMentoringTab(),
    },
    {
      key: "investment",
      label: "투자 내역",
      children: renderInvestmentTab(),
    },
  ];

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

      <Content
        style={{ padding: "40px 24px", background: "var(--bg-primary)" }}
      >
        <div style={{ maxWidth: 1200, margin: "0 auto" }}>
          {loading ? (
            // 로딩 상태
            <div
              style={{
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                minHeight: "400px",
              }}
            >
              <Spin size="large" />
              <Text style={{ marginLeft: 16 }}>프로필을 불러오는 중...</Text>
            </div>
          ) : error ? (
            // 에러 상태
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                alignItems: "center",
                minHeight: "400px",
              }}
            >
              <Text type="danger" style={{ fontSize: 16, marginBottom: 16 }}>
                {error}
              </Text>
              <Button onClick={() => window.location.reload()}>
                다시 시도
              </Button>
            </div>
          ) : !profileData ? (
            // 데이터 없음
            <div
              style={{
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                minHeight: "400px",
              }}
            >
              <Text>프로필 데이터를 찾을 수 없습니다.</Text>
            </div>
          ) : (
            // 정상 데이터 표시
            <>
              {/* Profile Header */}
              <Card
                style={{
                  marginBottom: 24,
                  background:
                    "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
                  border: "none",
                  borderRadius: 16,
                  color: "white",
                }}
                bodyStyle={{ padding: 40 }}
              >
                <Row align="middle" gutter={40}>
                  <Col>
                    <Avatar
                      size={120}
                      src={`https://api.dicebear.com/6.x/avataaars/svg?seed=${profileData.username}`}
                    />
                  </Col>
                  <Col flex="auto">
                    <Space direction="vertical" size="middle">
                      <div>
                        <Title level={1} style={{ color: "white", margin: 0 }}>
                          {profileData.displayName}
                        </Title>
                        <Text
                          style={{
                            color: "rgba(255,255,255,0.8)",
                            fontSize: 16,
                          }}
                        >
                          @{profileData.username}
                        </Text>
                      </div>
                      <Paragraph
                        style={{
                          color: "rgba(255,255,255,0.9)",
                          fontSize: 16,
                          margin: 0,
                        }}
                      >
                        {profileData.bio}
                      </Paragraph>
                      <Text
                        style={{ color: "rgba(255,255,255,0.7)", fontSize: 14 }}
                      >
                        <CalendarOutlined /> {profileData.joinedDate}부터 활동
                      </Text>
                    </Space>
                  </Col>
                  <Col>
                    <Space direction="vertical" size="large" align="center">
                      {/* 평판 지표 */}
                      <Row gutter={32}>
                        <Col span={12}>
                          <div style={{ textAlign: "center" }}>
                            <div
                              style={{
                                fontSize: 24,
                                fontWeight: 600,
                                color: "white",
                              }}
                            >
                              {profileData.stats.projectSuccessRate}%
                            </div>
                            <div
                              style={{
                                fontSize: 12,
                                color: "rgba(255,255,255,0.8)",
                              }}
                            >
                              프로젝트 성공률
                            </div>
                          </div>
                        </Col>
                        <Col span={12}>
                          <div style={{ textAlign: "center" }}>
                            <div
                              style={{
                                fontSize: 24,
                                fontWeight: 600,
                                color: "white",
                              }}
                            >
                              {profileData.stats.mentoringSuccessRate}%
                            </div>
                            <div
                              style={{
                                fontSize: 12,
                                color: "rgba(255,255,255,0.8)",
                              }}
                            >
                              멘토링 성공률
                            </div>
                          </div>
                        </Col>
                      </Row>
                      <Row gutter={32}>
                        <Col span={12}>
                          <div style={{ textAlign: "center" }}>
                            <div
                              style={{
                                fontSize: 18,
                                fontWeight: 600,
                                color: "white",
                              }}
                            >
                              $
                              {(
                                (profileData.stats.totalInvestment || 0) /
                                100 /
                                1000
                              ).toFixed(1)}
                              K
                            </div>
                            <div
                              style={{
                                fontSize: 12,
                                color: "rgba(255,255,255,0.8)",
                              }}
                            >
                              총 투자액
                            </div>
                          </div>
                        </Col>
                        <Col span={12}>
                          <div style={{ textAlign: "center" }}>
                            <div
                              style={{
                                fontSize: 18,
                                fontWeight: 600,
                                color: "white",
                              }}
                            >
                              <TrophyOutlined /> {profileData.stats.sbtCount}
                            </div>
                            <div
                              style={{
                                fontSize: 12,
                                color: "rgba(255,255,255,0.8)",
                              }}
                            >
                              획득 SBT
                            </div>
                          </div>
                        </Col>
                      </Row>

                      {/* 액션 버튼 */}
                      <div>
                        {isOwnProfile ? (
                          <Button
                            size="large"
                            icon={<EditOutlined />}
                            className="btn-secondary"
                            style={{
                              height: "40px",
                              fontSize: "14px",
                              fontWeight: "500",
                              backgroundColor: "rgba(255,255,255,0.9)",
                              color: "#1890ff",
                              border: "1px solid rgba(255,255,255,0.3)",
                            }}
                            onClick={() => navigate("/settings")}
                          >
                            프로필 편집
                          </Button>
                        ) : (
                          <Button
                            size="large"
                            icon={<MailOutlined />}
                            className="btn-secondary"
                            style={{
                              height: "40px",
                              fontSize: "14px",
                              fontWeight: "500",
                              backgroundColor: "rgba(255,255,255,0.9)",
                              color: "#1890ff",
                              border: "1px solid rgba(255,255,255,0.3)",
                            }}
                            onClick={() =>
                              message.info(
                                "메시지 기능은 곧 출시될 예정입니다!"
                              )
                            }
                          >
                            메시지 보내기
                          </Button>
                        )}
                      </div>
                    </Space>
                  </Col>
                </Row>
              </Card>

              {/* Tab Content */}
              <Card style={{ borderRadius: 16 }}>
                <Tabs
                  activeKey={activeTab}
                  onChange={setActiveTab}
                  items={tabItems}
                  size="large"
                />
              </Card>
            </>
          )}
        </div>
      </Content>
    </Layout>
  );
};

export default ProfilePage;
