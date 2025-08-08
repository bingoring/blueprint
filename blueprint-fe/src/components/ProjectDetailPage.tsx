import {
  ArrowLeftOutlined,
  CalendarOutlined,
  FlagOutlined,
  ProjectOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Button,
  Card,
  Col,
  Layout,
  message,
  Row,
  Space,
  Spin,
  Tag,
  Timeline,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { Milestone, Project } from "../types";
import ThemeToggle from "./ThemeToggle";

const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;

const ProjectDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [project, setProject] = useState<Project | null>(null);

  const loadProject = async () => {
    if (!id) return;

    try {
      setLoading(true);
      const response = await apiClient.getProject(parseInt(id));

      if (response.success && response.data) {
        setProject(response.data);
      } else {
        message.error(response.error || "프로젝트를 불러올 수 없습니다");
      }
    } catch (error) {
      console.error("프로젝트 로드 실패:", error);
      message.error("프로젝트를 불러오는 중 오류가 발생했습니다");
      navigate("/");
    } finally {
      setLoading(false);
    }
  };

  // 투자 모달 열기
  const openInvestModal = (milestone: Milestone) => {
    if (!milestone.id) return;

    // 폴리마켓 스타일 거래 페이지로 이동
    navigate(`/trade/${project?.id}/${milestone.id}`);
  };

  useEffect(() => {
    loadProject();
  }, [id]);

  if (loading) {
    return (
      <Layout style={{ minHeight: "100vh", background: "var(--bg-primary)" }}>
        <Content
          style={{
            padding: "50px",
            textAlign: "center",
            background: "var(--bg-primary)",
          }}
        >
          <Spin size="large" />
          <div style={{ marginTop: 20, color: "var(--text-primary)" }}>
            프로젝트 정보를 로딩 중...
          </div>
        </Content>
      </Layout>
    );
  }

  if (!project) {
    return (
      <Layout style={{ minHeight: "100vh", background: "var(--bg-primary)" }}>
        <Content
          style={{
            padding: "50px",
            textAlign: "center",
            background: "var(--bg-primary)",
          }}
        >
          <div style={{ color: "var(--text-primary)" }}>
            프로젝트를 찾을 수 없습니다.
          </div>
          <Button
            type="primary"
            onClick={() => navigate("/")}
            style={{ marginTop: 20 }}
          >
            홈으로 돌아가기
          </Button>
        </Content>
      </Layout>
    );
  }

  // 프로젝트 소유자 여부 확인
  const isOwner = user && Number(user.id) === project.user_id;

  return (
    <Layout style={{ minHeight: "100vh", background: "var(--bg-primary)" }}>
      <Content style={{ padding: "24px", background: "var(--bg-primary)" }}>
        <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
          {/* 상단 헤더 바 */}
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 24,
              padding: "12px 24px",
              background: "var(--bg-secondary)",
              borderRadius: "8px",
              border: "1px solid var(--border-color)",
            }}
          >
            <Button
              type="link"
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate("/")}
              style={{ padding: 0, color: "var(--text-primary)" }}
            >
              프로젝트 목록으로 돌아가기
            </Button>
            <ThemeToggle />
          </div>

          {/* 프로젝트 헤더 */}
          <div style={{ marginBottom: 32 }}>
            <Card>
              <Row gutter={[24, 24]}>
                <Col span={18}>
                  <Space
                    direction="vertical"
                    size={16}
                    style={{ width: "100%" }}
                  >
                    <div>
                      <Space>
                        <ProjectOutlined
                          style={{ fontSize: 24, color: "#1890ff" }}
                        />
                        <Title level={2} style={{ margin: 0 }}>
                          {project.title}
                        </Title>
                        <Tag
                          color={
                            project.status === "active" ? "green" : "orange"
                          }
                        >
                          {project.status === "active" ? "진행중" : "계획중"}
                        </Tag>
                        <Tag color="blue">{project.category}</Tag>
                      </Space>
                    </div>

                    <Paragraph style={{ fontSize: 16, marginBottom: 0 }}>
                      {project.description}
                    </Paragraph>

                    <Row gutter={16}>
                      <Col>
                        <Space>
                          <CalendarOutlined />
                          <Text>
                            목표 날짜: {project.target_date || "설정되지 않음"}
                          </Text>
                        </Space>
                      </Col>
                      <Col>
                        <Space>
                          <FlagOutlined />
                          <Text>우선순위: {project.priority}</Text>
                        </Space>
                      </Col>
                    </Row>
                  </Space>
                </Col>

                <Col span={6}>
                  <Space
                    direction="vertical"
                    size={16}
                    style={{ width: "100%", textAlign: "center" }}
                  >
                    <div>
                      <UserOutlined
                        style={{ fontSize: 48, color: "#1890ff" }}
                      />
                    </div>
                    <div>
                      <Text strong style={{ fontSize: 18 }}>
                        프로젝트 소유자
                      </Text>
                      <div style={{ marginTop: 8 }}>
                        <Text>사용자 #{project.user_id}</Text>
                      </div>
                    </div>

                    {!isOwner && isAuthenticated && (
                      <Button
                        type="primary"
                        size="large"
                        onClick={() =>
                          openInvestModal(
                            project?.milestones?.[0] || ({} as Milestone)
                          )
                        }
                        style={{ width: "100%" }}
                      >
                        💰 투자하기
                      </Button>
                    )}
                  </Space>
                </Col>
              </Row>
            </Card>
          </div>

          {/* 마일스톤 타임라인 */}
          <Card title="🎯 마일스톤 진행 상황" style={{ marginBottom: 24 }}>
            {project.milestones && project.milestones.length > 0 ? (
              <Timeline
                mode="left"
                items={project.milestones
                  .sort((a, b) => a.order - b.order)
                  .map((milestone) => ({
                    key: milestone.id,
                    color:
                      milestone.status === "completed"
                        ? "green"
                        : milestone.status === "failed"
                        ? "red"
                        : "blue",
                    label: milestone.target_date,
                    children: (
                      <Card size="small" style={{ marginBottom: 16 }}>
                        <Row justify="space-between" align="middle">
                          <Col span={16}>
                            <Space direction="vertical" size={4}>
                              <Text strong style={{ fontSize: 16 }}>
                                {milestone.title}
                              </Text>
                              <Text type="secondary">
                                {milestone.description}
                              </Text>
                              <Space>
                                <Tag
                                  color={
                                    milestone.status === "completed"
                                      ? "green"
                                      : milestone.status === "failed"
                                      ? "red"
                                      : "blue"
                                  }
                                >
                                  {milestone.status === "completed"
                                    ? "완료"
                                    : milestone.status === "failed"
                                    ? "실패"
                                    : "진행중"}
                                </Tag>
                                {milestone.betting_type && (
                                  <Tag color="purple">
                                    {milestone.betting_type === "simple"
                                      ? "단순 베팅"
                                      : "사용자 정의"}
                                  </Tag>
                                )}
                              </Space>
                            </Space>
                          </Col>
                          <Col span={8} style={{ textAlign: "right" }}>
                            <Space direction="vertical" size={8}>
                              <Text type="secondary" style={{ fontSize: 12 }}>
                                순서: {milestone.order}
                              </Text>
                              {!isOwner &&
                                isAuthenticated &&
                                milestone.status === "pending" && (
                                  <Button
                                    type="primary"
                                    size="small"
                                    onClick={() => openInvestModal(milestone)}
                                  >
                                    💰 투자하기
                                  </Button>
                                )}
                            </Space>
                          </Col>
                        </Row>
                      </Card>
                    ),
                  }))}
              />
            ) : (
              <div style={{ textAlign: "center", padding: "40px 0" }}>
                <Text type="secondary">아직 등록된 마일스톤이 없습니다.</Text>
              </div>
            )}
          </Card>
        </div>
      </Content>
    </Layout>
  );
};

export default ProjectDetailPage;
