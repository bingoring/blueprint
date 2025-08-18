import {
  CalendarOutlined,
  EditOutlined,
  MailOutlined,
  StarOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Button,
  Card,
  Col,
  List,
  Progress,
  Row,
  Space,
  Statistic,
  Tabs,
  Tag,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useAuthStore } from "../stores/useAuthStore";
import GlobalNavbar from "./GlobalNavbar";
import {
  InvestmentIcon,
  MilestoneIcon,
  PathIcon,
  TrophyIcon,
} from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;

interface UserProfile {
  id: number;
  username: string;
  displayName: string;
  bio: string;
  avatar?: string;
  joinDate: string;
  location?: string;
  website?: string;
  stats: {
    totalProjects: number;
    completedProjects: number;
    totalInvestment: number;
    totalEarnings: number;
    successRate: number;
    mentorRating: number;
  };
  currentProjects: Array<{
    id: number;
    title: string;
    category: string;
    progress: number;
    role: "creator" | "investor" | "mentor";
  }>;
  completedProjects: Array<{
    id: number;
    title: string;
    category: string;
    successRate: number;
    investment: number;
    earnings: number;
  }>;
  achievements: Array<{
    id: number;
    title: string;
    description: string;
    icon: string;
    earnedAt: string;
  }>;
}

const ProfilePage: React.FC = () => {
  const navigate = useNavigate();
  const { username } = useParams<{ username: string }>();
  const { user, isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [profileData, setProfileData] = useState<UserProfile | null>(null);
  const [activeTab, setActiveTab] = useState("projects");

  // 본인 프로필인지 확인
  const isOwnProfile = !username || username === user?.username;
  const displayUser = isOwnProfile ? user : null;

  // Mock data for development
  useEffect(() => {
    const loadMockProfile = () => {
      const mockProfile: UserProfile = {
        id: 1,
        username: displayUser?.username || username || "user",
        displayName: displayUser?.displayName || `${username || "user"}님`,
        bio: "혁신적인 아이디어로 세상을 바꾸고 싶은 개발자입니다. AI와 블록체인 기술에 관심이 많으며, 지속가능한 솔루션을 만드는 것이 목표입니다.",
        joinDate: "2024-01-15",
        location: "서울, 대한민국",
        website: "https://example.com",
        stats: {
          totalProjects: 8,
          completedProjects: 5,
          totalInvestment: 25000,
          totalEarnings: 8750,
          successRate: 87,
          mentorRating: 4.8,
        },
        currentProjects: [
          {
            id: 1,
            title: "AI 기반 피트니스 앱",
            category: "IT/개발",
            progress: 75,
            role: "creator",
          },
          {
            id: 2,
            title: "친환경 배달 서비스",
            category: "창업",
            progress: 45,
            role: "investor",
          },
          {
            id: 3,
            title: "블록체인 투표 시스템",
            category: "IT/개발",
            progress: 90,
            role: "mentor",
          },
        ],
        completedProjects: [
          {
            id: 4,
            title: "온라인 교육 플랫폼",
            category: "교육",
            successRate: 100,
            investment: 5000,
            earnings: 2500,
          },
          {
            id: 5,
            title: "스마트 홈 IoT 시스템",
            category: "IT/개발",
            successRate: 95,
            investment: 8000,
            earnings: 3200,
          },
        ],
        achievements: [
          {
            id: 1,
            title: "첫 프로젝트 성공",
            description: "첫 번째 프로젝트를 성공적으로 완료했습니다",
            icon: "🚀",
            earnedAt: "2024-02-01",
          },
          {
            id: 2,
            title: "멘토 마스터",
            description: "10명 이상의 프로젝트 창작자를 멘토링했습니다",
            icon: "👨‍🏫",
            earnedAt: "2024-03-15",
          },
          {
            id: 3,
            title: "투자 고수",
            description: "총 투자액 $20,000를 달성했습니다",
            icon: "💰",
            earnedAt: "2024-04-10",
          },
        ],
      };

      setProfileData(mockProfile);
      setLoading(false);
    };

    setTimeout(loadMockProfile, 500);
  }, [displayUser, username]);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR").format(amount);
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case "creator":
        return "blue";
      case "investor":
        return "green";
      case "mentor":
        return "purple";
      default:
        return "default";
    }
  };

  const getRoleText = (role: string) => {
    switch (role) {
      case "creator":
        return "창작자";
      case "investor":
        return "투자자";
      case "mentor":
        return "멘토";
      default:
        return role;
    }
  };

  // 탭 컨텐츠
  const tabItems = [
    {
      key: "projects",
      label: (
        <Space>
          <PathIcon size={16} />
          진행 중인 프로젝트
        </Space>
      ),
      children: (
        <Row gutter={[16, 16]}>
          {profileData?.currentProjects.map((project) => (
            <Col span={8} key={project.id}>
              <Card
                hoverable
                onClick={() => navigate(`/project/${project.id}`)}
              >
                <Space direction="vertical" style={{ width: "100%" }}>
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <Text strong>{project.title}</Text>
                    <Tag color={getRoleColor(project.role)}>
                      {getRoleText(project.role)}
                    </Tag>
                  </div>
                  <Text type="secondary">{project.category}</Text>
                  <div>
                    <Text style={{ fontSize: "12px", marginBottom: "4px" }}>
                      진행률 {project.progress}%
                    </Text>
                    <Progress
                      percent={project.progress}
                      strokeColor="var(--primary-color)"
                      size="small"
                    />
                  </div>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      ),
    },
    {
      key: "completed",
      label: (
        <Space>
          <TrophyIcon size={16} />
          완료된 프로젝트
        </Space>
      ),
      children: (
        <List
          dataSource={profileData?.completedProjects}
          renderItem={(project) => (
            <List.Item>
              <List.Item.Meta
                title={
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <Text strong>{project.title}</Text>
                    <Space>
                      <Tag color="green">성공률 {project.successRate}%</Tag>
                      <Text style={{ color: "#52c41a" }}>
                        +{formatCurrency(project.earnings)} USDC
                      </Text>
                    </Space>
                  </div>
                }
                description={project.category}
              />
            </List.Item>
          )}
        />
      ),
    },
    {
      key: "investments",
      label: (
        <Space>
          <InvestmentIcon size={16} />
          투자 내역
        </Space>
      ),
      children: (
        <div>
          <Row gutter={[24, 24]} style={{ marginBottom: "24px" }}>
            <Col span={6}>
              <Statistic
                title="총 투자액"
                value={profileData?.stats.totalInvestment || 0}
                suffix="USDC"
                valueStyle={{ color: "#1890ff" }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="총 수익"
                value={profileData?.stats.totalEarnings || 0}
                suffix="USDC"
                valueStyle={{ color: "#52c41a" }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="수익률"
                value={
                  profileData?.stats.totalEarnings &&
                  profileData?.stats.totalInvestment
                    ? (
                        (profileData.stats.totalEarnings /
                          profileData.stats.totalInvestment) *
                        100
                      ).toFixed(1)
                    : 0
                }
                suffix="%"
                valueStyle={{ color: "#52c41a" }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="성공률"
                value={profileData?.stats.successRate || 0}
                suffix="%"
                valueStyle={{ color: "#722ed1" }}
              />
            </Col>
          </Row>
        </div>
      ),
    },
    {
      key: "achievements",
      label: (
        <Space>
          <TrophyIcon size={16} />
          업적
        </Space>
      ),
      children: (
        <Row gutter={[16, 16]}>
          {profileData?.achievements.map((achievement) => (
            <Col span={8} key={achievement.id}>
              <Card>
                <Space
                  direction="vertical"
                  style={{ textAlign: "center", width: "100%" }}
                >
                  <div style={{ fontSize: "32px" }}>{achievement.icon}</div>
                  <Text strong>{achievement.title}</Text>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    {achievement.description}
                  </Text>
                  <Text type="secondary" style={{ fontSize: "11px" }}>
                    {new Date(achievement.earnedAt).toLocaleDateString("ko-KR")}
                  </Text>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      ),
    },
  ];

  if (!isAuthenticated && isOwnProfile) {
    navigate("/login");
    return null;
  }

  return (
    <div style={{ background: "var(--bg-primary)", minHeight: "100vh" }}>
      <GlobalNavbar />

      <div style={{ paddingTop: "64px" }}>
        <div
          style={{
            maxWidth: "1400px",
            margin: "0 auto",
            padding: "32px 24px",
          }}
        >
          {loading ? (
            <div style={{ textAlign: "center", padding: "100px" }}>
              <Text>프로필을 불러오는 중...</Text>
            </div>
          ) : !profileData ? (
            <div style={{ textAlign: "center", padding: "100px" }}>
              <Text>프로필을 찾을 수 없습니다.</Text>
            </div>
          ) : (
            <>
              {/* 프로필 헤더 */}
              <Card style={{ marginBottom: "24px" }}>
                <Row gutter={24} align="middle">
                  <Col span={4}>
                    <div style={{ textAlign: "center" }}>
                      <Avatar
                        size={120}
                        src={
                          profileData.avatar ||
                          `https://api.dicebear.com/6.x/avataaars/svg?seed=${profileData.username}`
                        }
                        icon={<UserOutlined />}
                      />
                    </div>
                  </Col>
                  <Col span={14}>
                    <Space direction="vertical" size="small">
                      <Title level={2} style={{ margin: 0 }}>
                        {profileData.displayName}
                      </Title>
                      <Text type="secondary" style={{ fontSize: "16px" }}>
                        @{profileData.username}
                      </Text>
                      <Paragraph style={{ margin: "12px 0", fontSize: "14px" }}>
                        {profileData.bio}
                      </Paragraph>
                      <Space>
                        <Text type="secondary">
                          <CalendarOutlined />{" "}
                          {new Date(profileData.joinDate).toLocaleDateString(
                            "ko-KR"
                          )}
                          에 가입
                        </Text>
                        {profileData.location && (
                          <Text type="secondary">
                            📍 {profileData.location}
                          </Text>
                        )}
                        {profileData.stats.mentorRating > 0 && (
                          <Text type="secondary">
                            <StarOutlined /> 멘토 평점{" "}
                            {profileData.stats.mentorRating}/5.0
                          </Text>
                        )}
                      </Space>
                    </Space>
                  </Col>
                  <Col span={6}>
                    <div style={{ textAlign: "right" }}>
                      {isOwnProfile ? (
                        <Button
                          type="primary"
                          icon={<EditOutlined />}
                          onClick={() => navigate("/settings")}
                        >
                          프로필 편집
                        </Button>
                      ) : (
                        <Space direction="vertical">
                          <Button type="primary" icon={<MailOutlined />}>
                            메시지 보내기
                          </Button>
                          <Button icon={<TeamOutlined />}>멘토링 요청</Button>
                        </Space>
                      )}
                    </div>
                  </Col>
                </Row>
              </Card>

              {/* 통계 카드 */}
              <Row gutter={[24, 24]} style={{ marginBottom: "24px" }}>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="총 프로젝트"
                      value={profileData.stats.totalProjects}
                      prefix={<PathIcon size={16} />}
                      valueStyle={{ color: "var(--primary-color)" }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="완료 프로젝트"
                      value={profileData.stats.completedProjects}
                      prefix={<TrophyIcon size={16} />}
                      valueStyle={{ color: "#52c41a" }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="총 투자액"
                      value={profileData.stats.totalInvestment}
                      suffix="USDC"
                      prefix={<InvestmentIcon size={16} />}
                      valueStyle={{ color: "#722ed1" }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="성공률"
                      value={profileData.stats.successRate}
                      suffix="%"
                      prefix={<MilestoneIcon size={16} />}
                      valueStyle={{ color: "#fa8c16" }}
                    />
                  </Card>
                </Col>
              </Row>

              {/* 탭 컨텐츠 */}
              <Card>
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
      </div>
    </div>
  );
};

export default ProfilePage;
