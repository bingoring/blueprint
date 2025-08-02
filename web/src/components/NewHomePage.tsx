import {
  LoginOutlined,
  PlusOutlined,
  ProjectOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Button,
  Card,
  Col,
  Layout,
  Row,
  Space,
  Spin,
  Statistic,
  Tag,
  Typography,
  message,
} from "antd";
import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { Project } from "../types";
import AuthModal from "./AuthModal";
import LanguageSwitcher from "./LanguageSwitcher";

const { Header, Content } = Layout;
const { Title, Paragraph } = Typography;

const NewHomePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();

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
    loadPublicProjects();
  }, []);

  const loadPublicProjects = async () => {
    try {
      setLoading(true);
      console.log("🔄 공개 프로젝트 데이터 로딩 중...");

      // 실제 API 호출: 공개 프로젝트 목록
      const response = await apiClient.getProjects({
        page: 1,
        limit: 10,
        sort: "created_at",
        order: "desc",
      });

      if (response.success && response.data) {
        const publicProjects = response.data.projects.filter(
          (project) => project.is_public === true
        );

        setProjects(publicProjects);

        // 통계 계산 (실제로는 별도 API가 있어야 함)
        const totalInvestment = publicProjects.reduce((sum, project) => {
          return (
            sum +
            (project.milestones?.reduce((milestoneSum, milestone) => {
              return milestoneSum + (milestone.total_support || 0);
            }, 0) || 0)
          );
        }, 0);

        const totalInvestors = publicProjects.reduce((sum, project) => {
          return (
            sum +
            (project.milestones?.reduce((milestoneSum, milestone) => {
              return milestoneSum + (milestone.supporter_count || 0);
            }, 0) || 0)
          );
        }, 0);

        setStats({
          totalProjects: publicProjects.length,
          totalInvestors: totalInvestors,
          totalInvestment: totalInvestment,
        });

        console.log("✅ 공개 프로젝트 로딩 완료:", publicProjects.length, "개");
      } else {
        throw new Error(response.error || "프로젝트 조회 실패");
      }
    } catch (error) {
      console.error("❌ 공개 프로젝트 로딩 실패:", error);
      message.error("프로젝트 데이터를 불러오는데 실패했습니다");

      // 에러 시 빈 데이터로 초기화
      setProjects([]);
      setStats({ totalProjects: 0, totalInvestors: 0, totalInvestment: 0 });
    } finally {
      setLoading(false);
    }
  };

  // 프로젝트 카테고리 번역
  const getCategoryLabel = (category: string) => {
    const categoryMap: Record<string, string> = {
      business: "🚀 Business",
      career: "💼 Career",
      education: "📚 Education",
      personal: "🌱 Personal",
      life: "🏡 Life",
    };
    return categoryMap[category] || category;
  };

  // 프로젝트 진행률 계산
  const calculateProgress = (project: Project): number => {
    if (!project.milestones || project.milestones.length === 0) return 0;

    const completedCount = project.milestones.filter(
      (milestone) => milestone.status === "completed"
    ).length;

    return Math.round((completedCount / project.milestones.length) * 100);
  };

  // 총 투자금 계산
  const calculateTotalInvestment = (project: Project): number => {
    if (!project.milestones) return 0;

    return project.milestones.reduce((sum, milestone) => {
      return sum + (milestone.total_support || 0);
    }, 0);
  };

  // 총 투자자 수 계산
  const calculateInvestorCount = (project: Project): number => {
    if (!project.milestones) return 0;

    return project.milestones.reduce((sum, milestone) => {
      return sum + (milestone.supporter_count || 0);
    }, 0);
  };

  // 남은 시간 계산 (간단한 계산)
  const calculateTimeLeft = (targetDate?: string | null): string => {
    if (!targetDate) return "기간 미정";

    const target = new Date(targetDate);
    const now = new Date();
    const diffTime = target.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays < 0) return "기간 만료";
    if (diffDays < 30) return `${diffDays}일`;
    if (diffDays < 365) return `${Math.ceil(diffDays / 30)}개월`;
    return `${Math.ceil(diffDays / 365)}년`;
  };

  return (
    <Layout style={{ minHeight: "100vh" }}>
      {/* Header */}
      <Header
        style={{
          background: "white",
          boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
          padding: "0 24px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <Title level={3} style={{ margin: 0, color: "#1890ff" }}>
            <ProjectOutlined /> Blueprint
          </Title>
        </div>

        <Space size="middle">
          <LanguageSwitcher />
          {isAuthenticated ? (
            <Space>
              <Button onClick={() => navigate("/dashboard")}>대시보드</Button>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate("/create-project")}
              >
                프로젝트 만들기
              </Button>
              <Button
                icon={<LoginOutlined />}
                onClick={() => useAuthStore.getState().logout()}
              >
                로그아웃
              </Button>
            </Space>
          ) : (
            <Button
              type="primary"
              icon={<LoginOutlined />}
              onClick={() => setAuthModalVisible(true)}
            >
              로그인
            </Button>
          )}
        </Space>
      </Header>

      <Content style={{ padding: "40px 24px" }}>
        <div style={{ maxWidth: 1200, margin: "0 auto" }}>
          {/* Hero Section */}
          <div style={{ textAlign: "center", marginBottom: 60 }}>
            <Title level={1} style={{ fontSize: 48, marginBottom: 16 }}>
              당신의 <span style={{ color: "#1890ff" }}>청사진</span>을 현실로
            </Title>
            <Paragraph
              style={{ fontSize: 18, marginBottom: 32, color: "#666" }}
            >
              프로젝트를 공유하고, 투자를 받고, 목표를 달성하세요.
              <br />
              투명한 과정으로 함께 성장하는 플랫폼입니다.
            </Paragraph>

            {!isAuthenticated && (
              <Space size="middle">
                <Button
                  type="primary"
                  size="large"
                  onClick={() => setAuthModalVisible(true)}
                >
                  지금 시작하기
                </Button>
                <Button size="large">더 알아보기</Button>
              </Space>
            )}
          </div>

          {/* Stats Section */}
          <Row gutter={[32, 32]} style={{ marginBottom: 60 }}>
            <Col xs={24} md={8}>
              <Card style={{ textAlign: "center" }}>
                <Statistic
                  title="활성 프로젝트"
                  value={stats.totalProjects}
                  prefix={<ProjectOutlined />}
                  valueStyle={{ color: "#1890ff" }}
                />
              </Card>
            </Col>
            <Col xs={24} md={8}>
              <Card style={{ textAlign: "center" }}>
                <Statistic
                  title="총 투자자"
                  value={stats.totalInvestors}
                  prefix={<UserOutlined />}
                  valueStyle={{ color: "#52c41a" }}
                />
              </Card>
            </Col>
            <Col xs={24} md={8}>
              <Card style={{ textAlign: "center" }}>
                <Statistic
                  title="총 투자금"
                  value={stats.totalInvestment}
                  prefix="₩"
                  precision={0}
                  valueStyle={{ color: "#faad14" }}
                />
              </Card>
            </Col>
          </Row>

          {/* Projects Section */}
          <div style={{ marginBottom: 40 }}>
            <Title level={2} style={{ textAlign: "center", marginBottom: 40 }}>
              🌟 최신 프로젝트
            </Title>

            {loading ? (
              <div style={{ textAlign: "center", padding: "50px" }}>
                <Spin size="large" />
                <div style={{ marginTop: 16 }}>프로젝트를 불러오는 중...</div>
              </div>
            ) : (
              <Row gutter={[24, 24]}>
                {projects.slice(0, 6).map((project) => (
                  <Col xs={24} sm={12} lg={8} key={project.id}>
                    <Card
                      hoverable
                      className="h-full cursor-pointer"
                      onClick={() => navigate(`/project/${project.id}`)}
                      actions={[
                        <Button
                          type="primary"
                          block
                          onClick={(e) => {
                            e.stopPropagation();
                            if (isAuthenticated) {
                              navigate(`/project/${project.id}`);
                            } else {
                              setAuthModalVisible(true);
                            }
                          }}
                        >
                          {t("investment.invest")}
                        </Button>,
                      ]}
                    >
                      <div style={{ marginBottom: 16 }}>
                        <Tag color="blue">
                          {getCategoryLabel(project.category)}
                        </Tag>
                        {calculateProgress(project) > 50 && (
                          <Tag color="green">인기</Tag>
                        )}
                      </div>

                      <Title level={4} style={{ marginBottom: 8 }}>
                        {project.title}
                      </Title>
                      <Paragraph
                        style={{
                          marginBottom: 16,
                          color: "#666",
                          height: "40px",
                          overflow: "hidden",
                        }}
                      >
                        {project.description}
                      </Paragraph>

                      <div style={{ marginBottom: 12 }}>
                        <Row gutter={16}>
                          <Col span={12}>
                            <div style={{ fontSize: 12, color: "#999" }}>
                              성공 확률
                            </div>
                            <div style={{ fontWeight: "bold" }}>
                              {calculateProgress(project)}%
                            </div>
                          </Col>
                          <Col span={12}>
                            <div style={{ fontSize: 12, color: "#999" }}>
                              남은 시간
                            </div>
                            <div style={{ fontWeight: "bold" }}>
                              {calculateTimeLeft(project.target_date)}
                            </div>
                          </Col>
                        </Row>
                      </div>

                      <div>
                        <Row gutter={16}>
                          <Col span={12}>
                            <div style={{ fontSize: 12, color: "#999" }}>
                              총 투자금
                            </div>
                            <div
                              style={{ fontWeight: "bold", color: "#52c41a" }}
                            >
                              ₩
                              {calculateTotalInvestment(
                                project
                              ).toLocaleString()}
                            </div>
                          </Col>
                          <Col span={12}>
                            <div style={{ fontSize: 12, color: "#999" }}>
                              투자자
                            </div>
                            <div style={{ fontWeight: "bold" }}>
                              {calculateInvestorCount(project)}명
                            </div>
                          </Col>
                        </Row>
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            )}
          </div>

          {/* CTA Section */}
          {!isAuthenticated && (
            <div
              style={{
                textAlign: "center",
                padding: "40px 0",
                background: "#f5f5f5",
                borderRadius: 8,
                marginTop: 40,
              }}
            >
              <Title level={3}>당신의 프로젝트도 시작해보세요!</Title>
              <Paragraph style={{ marginBottom: 24 }}>
                지금 가입하고 첫 프로젝트를 만들어보세요.
              </Paragraph>
              <Button
                type="primary"
                size="large"
                onClick={() => setAuthModalVisible(true)}
              >
                무료로 시작하기
              </Button>
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

export default NewHomePage;
