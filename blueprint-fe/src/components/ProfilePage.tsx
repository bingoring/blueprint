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
import { apiClient } from "../lib/api";
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

  // 실제 프로필 데이터 로드
  useEffect(() => {
    const loadProfile = async () => {
      try {
        setLoading(true);
        
        // 프로필 조회할 사용자명 결정
        const targetUsername = username || user?.username || "default";
        
        // API에서 실제 프로필 데이터 가져오기
        const response = await apiClient.getUserProfile(targetUsername);
        
        if (response.success && response.data) {
          // API 응답 구조를 UI 구조에 맞게 변환
          const apiData = response.data;
          const profileData: UserProfile = {
            id: apiData.username.length, // 임시 ID
            username: apiData.username,
            displayName: apiData.displayName,
            bio: apiData.bio || "소개글이 없습니다.",
            avatar: apiData.avatar,
            joinDate: apiData.joinedDate,
            location: "위치 정보 없음",
            website: "https://example.com",
            stats: {
              totalProjects: apiData.currentProjects.length + apiData.featuredProjects.length,
              completedProjects: apiData.featuredProjects.length,
              totalInvestment: apiData.stats.totalInvestment,
              totalEarnings: apiData.stats.totalInvestment * 0.1, // 임시 계산
              successRate: apiData.stats.projectSuccessRate,
              mentorRating: apiData.stats.mentoringSuccessRate / 20, // 100% -> 5점 변환
            },
            currentProjects: apiData.currentProjects.map(project => ({
              id: project.id,
              title: project.title,
              category: project.category,
              progress: project.progress,
              role: project.status === "active" ? "creator" : "investor", // 임시 매핑
            })),
            completedProjects: apiData.featuredProjects.map(project => ({
              id: project.id,
              title: project.title,
              category: "일반", // API에 카테고리가 없어서 임시값
              successRate: project.successRate,
              investment: project.investment,
              earnings: project.investment * 0.1, // 임시 계산
            })),
            achievements: [
              // 임시 성취 목록 (향후 백엔드에 achievements API 추가 필요)
              {
                id: 1,
                title: "프로필 활성화",
                description: "Blueprint에 프로필을 등록했습니다",
                icon: "🎯",
                earnedAt: apiData.joinedDate,
              },
            ],
          };
          
          setProfileData(profileData);
        } else {
          console.error("프로필 로드 실패:", response.error);
        }
      } catch (error) {
        console.error("프로필 로드 중 오류:", error);
        // 오류 발생 시에도 기본 프로필 표시
        const targetUsername = username || user?.username || "default";
        const fallbackProfile: UserProfile = {
          id: 0,
          username: targetUsername,
          displayName: `${targetUsername}님`,
          bio: "사용자 정보를 불러올 수 없습니다.",
          joinDate: "2024-01-01",
          location: "알 수 없음",
          website: "",
          stats: {
            totalProjects: 0,
            completedProjects: 0,
            totalInvestment: 0,
            totalEarnings: 0,
            successRate: 0,
            mentorRating: 0,
          },
          currentProjects: [],
          completedProjects: [],
          achievements: [],
        };
        setProfileData(fallbackProfile);
      } finally {
        setLoading(false);
      }
    };

    loadProfile();
  }, [username, user]);


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
