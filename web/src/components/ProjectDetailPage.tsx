import {
  ArrowLeftOutlined,
  CalendarOutlined,
  DollarOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Badge,
  Button,
  Card,
  Col,
  Empty,
  Form,
  InputNumber,
  Layout,
  Modal,
  Progress,
  Radio,
  Row,
  Space,
  Spin,
  Statistic,
  Tag,
  Timeline,
  Typography,
  message,
} from "antd";
import dayjs from "dayjs";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type {
  Milestone,
  Project,
  ProjectCategory,
  ProjectStatus,
} from "../types";

const { Content } = Layout;
const { Title, Paragraph, Text } = Typography;

interface Investor {
  id: number;
  username: string;
  amount: number;
  date: string;
  milestone_bets: Array<{
    milestone_id: number;
    option: string;
    amount: number;
  }>;
}

interface ProjectStats {
  completion_rate: number;
  total_investment: number;
  investor_count: number;
  success_probability: number;
  recent_investors: Investor[];
}

const ProjectDetailPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { isAuthenticated, user } = useAuthStore();

  // 상태 관리
  const [loading, setLoading] = useState(true);
  const [project, setProject] = useState<Project | null>(null);
  const [projectStats, setProjectStats] = useState<ProjectStats | null>(null);
  const [isOwner, setIsOwner] = useState(false);
  const [investModalVisible, setInvestModalVisible] = useState(false);
  const [investmentForm] = Form.useForm();
  const [totalInvestmentAmount, setTotalInvestmentAmount] = useState(0);

  useEffect(() => {
    if (!id) {
      message.error("프로젝트 ID가 없습니다");
      navigate("/");
      return;
    }

    loadProjectData();
  }, [id, navigate]);

  const loadProjectData = async () => {
    try {
      setLoading(true);
      console.log("🔄 프로젝트 상세 데이터 로딩 중...", id);

      // 실제 API 호출: 프로젝트 상세 정보
      const response = await apiClient.getProject(parseInt(id!));

      if (response.success && response.data) {
        const projectData = response.data;
        setProject(projectData);

        // 프로젝트 소유자 확인
        if (user && projectData.user_id) {
          setIsOwner(Number(user.id) === projectData.user_id);
        }

        // 프로젝트 통계 계산 (실제로는 별도 API가 있어야 함)
        const stats = calculateProjectStats(projectData);
        setProjectStats(stats);

        console.log("✅ 프로젝트 상세 데이터 로딩 완료:", projectData.title);
      } else {
        throw new Error(response.error || "프로젝트 조회 실패");
      }
    } catch (error) {
      console.error("❌ 프로젝트 데이터 로딩 실패:", error);
      message.error("프로젝트 정보를 불러오는데 실패했습니다");
      navigate("/");
    } finally {
      setLoading(false);
    }
  };

  // 프로젝트 통계 계산
  const calculateProjectStats = (project: Project): ProjectStats => {
    if (!project.milestones || project.milestones.length === 0) {
      return {
        completion_rate: 0,
        total_investment: 0,
        investor_count: 0,
        success_probability: 0,
        recent_investors: [],
      };
    }

    // 완료율 계산
    const completedCount = project.milestones.filter(
      (milestone) => milestone.status === "completed"
    ).length;
    const completion_rate = Math.round(
      (completedCount / project.milestones.length) * 100
    );

    // 총 투자금 계산
    const total_investment = project.milestones.reduce((sum, milestone) => {
      return sum + (milestone.total_support || 0);
    }, 0);

    // 총 투자자 수 계산
    const investor_count = project.milestones.reduce((sum, milestone) => {
      return sum + (milestone.supporter_count || 0);
    }, 0);

    // 성공 확률 계산 (단순히 진행률 기반)
    const success_probability = Math.min(completion_rate + 20, 95);

    // 임시 최근 투자자 (실제로는 별도 API 필요)
    const recent_investors: Investor[] = [];

    return {
      completion_rate,
      total_investment,
      investor_count,
      success_probability,
      recent_investors,
    };
  };

  // 카테고리 번역
  const getCategoryLabel = (category: ProjectCategory) => {
    const categoryMap: Record<ProjectCategory, string> = {
      business: "🚀 Business",
      career: "💼 Career",
      education: "📚 Education",
      personal: "🌱 Personal",
      life: "🏡 Life",
    };
    return categoryMap[category] || category;
  };

  // 상태 번역
  const getStatusBadge = (status: ProjectStatus) => {
    const statusConfig: Record<
      ProjectStatus,
      {
        status: "success" | "processing" | "error" | "default" | "warning";
        text: string;
      }
    > = {
      draft: { status: "default", text: "초안" },
      active: { status: "processing", text: "진행중" },
      completed: { status: "success", text: "완료" },
      cancelled: { status: "error", text: "취소" },
      on_hold: { status: "warning", text: "보류" },
    };

    const config = statusConfig[status] || statusConfig.draft;
    return <Badge status={config.status} text={config.text} />;
  };

  // 투자 모달 열기
  const openInvestModal = () => {
    if (!isAuthenticated) {
      message.warning("투자하려면 로그인이 필요합니다");
      return;
    }
    setInvestModalVisible(true);
  };

  // 투자 금액 계산
  const calculateTotalAmount = () => {
    const values = investmentForm.getFieldsValue();
    let total = 0;

    project?.milestones?.forEach((milestone) => {
      const amount = values[`milestone_${milestone.id}_amount`] || 0;
      total += amount;
    });

    setTotalInvestmentAmount(total);
  };

  // 태그 표시 헬퍼
  const renderProjectTags = (tags: string) => {
    try {
      const parsed = JSON.parse(tags);
      if (typeof parsed === "object" && parsed !== null) {
        return Object.entries(parsed).map(([key, value]) => (
          <Tag key={key} color="geekblue">
            {key}: {String(value)}
          </Tag>
        ));
      }
    } catch {
      // If parsing fails, treat as simple string
    }
    return <Tag color="geekblue">{tags}</Tag>;
  };

  // 베팅 옵션 표시 헬퍼
  const renderBettingOptions = (milestone: Milestone) => {
    if (!milestone.betting_options || milestone.betting_options.length === 0) {
      return null;
    }

    return (
      <span>
        {milestone.betting_options.map((option: string, index: number) => (
          <Tag key={index} color="blue">
            {option}
          </Tag>
        ))}
      </span>
    );
  };

  // 모달용 베팅 옵션 렌더링
  const renderModalBettingOptions = (milestone: Milestone) => {
    if (milestone.betting_type === "simple") {
      return (
        <>
          <Radio value="성공">✅ 성공</Radio>
          <Radio value="실패">❌ 실패</Radio>
        </>
      );
    }

    if (!milestone.betting_options || milestone.betting_options.length === 0) {
      return null;
    }

    return milestone.betting_options.map(
      (option: string, optionIndex: number) => (
        <Radio key={optionIndex} value={option}>
          {option}
        </Radio>
      )
    );
  };

  if (loading) {
    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Content style={{ padding: "50px", textAlign: "center" }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>프로젝트 정보를 불러오는 중...</div>
        </Content>
      </Layout>
    );
  }

  if (!project) {
    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Content style={{ padding: "50px", textAlign: "center" }}>
          <Empty
            description="프로젝트를 찾을 수 없습니다"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type="primary" onClick={() => navigate("/")}>
              홈으로 돌아가기
            </Button>
          </Empty>
        </Content>
      </Layout>
    );
  }

  return (
    <Layout style={{ minHeight: "100vh", background: "#f5f5f5" }}>
      <Content style={{ padding: "24px" }}>
        <div style={{ maxWidth: 1200, margin: "0 auto" }}>
          {/* 뒤로가기 버튼 */}
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate("/")}
            style={{ marginBottom: 24 }}
          >
            홈으로 돌아가기
          </Button>

          {/* 프로젝트 헤더 */}
          <Card style={{ marginBottom: 24 }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "start",
                marginBottom: 16,
              }}
            >
              <div style={{ flex: 1 }}>
                <Title level={2} style={{ marginBottom: 8 }}>
                  {project.title}
                </Title>
                <Space size="middle" style={{ marginBottom: 12 }}>
                  <Tag color="blue">{getCategoryLabel(project.category)}</Tag>
                  {getStatusBadge(project.status || "draft")}
                  <Text type="secondary">
                    <CalendarOutlined style={{ marginRight: 4 }} />
                    목표일:{" "}
                    {project.target_date
                      ? dayjs(project.target_date).format("YYYY-MM-DD")
                      : "미정"}
                  </Text>
                </Space>
              </div>
              {!isOwner && isAuthenticated && (
                <Button
                  type="primary"
                  size="large"
                  icon={<DollarOutlined />}
                  onClick={openInvestModal}
                >
                  투자하기
                </Button>
              )}
            </div>

            {projectStats && (
              <Progress
                percent={projectStats.completion_rate}
                strokeColor="#52c41a"
              />
            )}

            <Paragraph style={{ marginTop: 16, fontSize: 16 }}>
              {project.description}
            </Paragraph>

            {/* 프로젝트 태그 & 메트릭 */}
            {project.tags && (
              <div style={{ marginTop: 16 }}>
                <Text strong>태그: </Text>
                {renderProjectTags(project.tags)}
              </div>
            )}

            {project.metrics && (
              <div style={{ marginTop: 8 }}>
                <Text strong>성공 지표: </Text>
                <Text>{project.metrics}</Text>
              </div>
            )}
          </Card>

          <Row gutter={[24, 24]}>
            {/* 왼쪽: 마일스톤 타임라인 */}
            <Col xs={24} lg={16}>
              <Card title="📋 프로젝트 마일스톤">
                {project.milestones && project.milestones.length > 0 ? (
                  <Timeline>
                    {project.milestones
                      .sort((a, b) => (a.order || 0) - (b.order || 0))
                      .map((milestone) => (
                        <Timeline.Item
                          key={milestone.id}
                          color={
                            milestone.status === "completed" ? "green" : "blue"
                          }
                        >
                          <div>
                            <Title level={5}>
                              🎯 마일스톤 {milestone.order}: {milestone.title}
                            </Title>
                            <Paragraph style={{ color: "#666" }}>
                              {milestone.description}
                            </Paragraph>

                            {milestone.target_date && (
                              <Text type="secondary">
                                <CalendarOutlined style={{ marginRight: 4 }} />
                                목표일:{" "}
                                {dayjs(milestone.target_date).format(
                                  "YYYY-MM-DD"
                                )}
                              </Text>
                            )}

                            {/* 베팅 옵션 표시 */}
                            {milestone.betting_options && (
                              <div style={{ marginTop: 8 }}>
                                <Text strong>투자 옵션: </Text>
                                {milestone.betting_type === "simple" ? (
                                  <span>
                                    <Tag color="green">✅ 성공</Tag>
                                    <Tag color="red">❌ 실패</Tag>
                                  </span>
                                ) : (
                                  renderBettingOptions(milestone)
                                )}
                              </div>
                            )}

                            {/* 투자 정보 */}
                            <div style={{ marginTop: 8 }}>
                              <Space>
                                <Text>
                                  💰 투자금: ₩
                                  {(
                                    milestone.total_support || 0
                                  ).toLocaleString()}
                                </Text>
                                <Text>
                                  👥 투자자: {milestone.supporter_count || 0}명
                                </Text>
                              </Space>
                            </div>
                          </div>
                        </Timeline.Item>
                      ))}
                  </Timeline>
                ) : (
                  <Empty description="마일스톤이 없습니다" />
                )}
              </Card>
            </Col>

            {/* 오른쪽: 투자 통계 & 최근 투자자 */}
            <Col xs={24} lg={8}>
              <Space
                direction="vertical"
                style={{ width: "100%" }}
                size="large"
              >
                {/* 투자 통계 */}
                <Card title="📊 투자 현황">
                  <Row gutter={[16, 16]}>
                    <Col span={12}>
                      <Statistic
                        title="완료율"
                        value={projectStats?.completion_rate || 0}
                        suffix="%"
                        valueStyle={{ color: "#52c41a" }}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="성공 확률"
                        value={projectStats?.success_probability || 0}
                        suffix="%"
                        valueStyle={{ color: "#1890ff" }}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="총 투자금"
                        value={projectStats?.total_investment || 0}
                        prefix="₩"
                        valueStyle={{ color: "#faad14" }}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="투자자 수"
                        value={projectStats?.investor_count || 0}
                        suffix="명"
                        prefix={<TeamOutlined />}
                        valueStyle={{ color: "#722ed1" }}
                      />
                    </Col>
                  </Row>
                </Card>

                {/* 최근 투자자 */}
                <Card title="👥 최근 투자자">
                  {projectStats?.recent_investors &&
                  projectStats.recent_investors.length > 0 ? (
                    <div style={{ maxHeight: 300, overflowY: "auto" }}>
                      {projectStats.recent_investors.map((investor) => (
                        <div
                          key={investor.id}
                          style={{
                            padding: "8px 0",
                            borderBottom: "1px solid #f0f0f0",
                            display: "flex",
                            justifyContent: "space-between",
                            alignItems: "center",
                          }}
                        >
                          <div>
                            <div style={{ fontWeight: "bold" }}>
                              {investor.username}
                            </div>
                            <div style={{ fontSize: "12px", color: "#999" }}>
                              {investor.date}
                            </div>
                          </div>
                          <div style={{ textAlign: "right" }}>
                            <div
                              style={{ fontWeight: "bold", color: "#52c41a" }}
                            >
                              ₩{investor.amount.toLocaleString()}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <Empty
                      description="아직 투자자가 없습니다"
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                    />
                  )}
                </Card>
              </Space>
            </Col>
          </Row>
        </div>
      </Content>

      {/* 투자 모달 */}
      <Modal
        title="💰 프로젝트 투자하기"
        open={investModalVisible}
        onCancel={() => setInvestModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setInvestModalVisible(false)}>
            취소
          </Button>,
          <Button
            key="invest"
            type="primary"
            onClick={() => message.info("투자 기능은 곧 구현 예정입니다!")}
          >
            ₩{totalInvestmentAmount.toLocaleString()} 투자하기
          </Button>,
        ]}
        width={800}
      >
        <Form
          form={investmentForm}
          layout="vertical"
          onChange={calculateTotalAmount}
        >
          {project.milestones?.map((milestone) => (
            <Card key={milestone.id} size="small" style={{ marginBottom: 16 }}>
              <Title level={5}>
                🎯 마일스톤 {milestone.order}: {milestone.title}
              </Title>
              <Paragraph style={{ fontSize: "14px", color: "#666" }}>
                {milestone.description}
              </Paragraph>

              <Form.Item
                label="베팅 옵션 선택"
                name={`milestone_${milestone.id}_option`}
                rules={[{ required: true, message: "옵션을 선택해주세요" }]}
              >
                <Radio.Group>
                  {renderModalBettingOptions(milestone)}
                </Radio.Group>
              </Form.Item>

              <Form.Item
                label="투자 금액 (원)"
                name={`milestone_${milestone.id}_amount`}
                rules={[
                  { required: true, message: "투자 금액을 입력해주세요" },
                ]}
              >
                <InputNumber
                  style={{ width: "100%" }}
                  min={1000}
                  step={1000}
                  formatter={(value) =>
                    `₩ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                  }
                  parser={(value: string | undefined) => {
                    if (!value) return 1000;
                    const parsed = Number(value.replace(/₩\s?|(,*)/g, ""));
                    return isNaN(parsed) ? 1000 : parsed;
                  }}
                  placeholder="최소 1,000원"
                />
              </Form.Item>
            </Card>
          ))}

          <Card
            size="small"
            style={{ backgroundColor: "#f6ffed", border: "1px solid #b7eb8f" }}
          >
            <Statistic
              title="총 투자 금액"
              value={totalInvestmentAmount}
              prefix="₩"
              valueStyle={{ color: "#52c41a", fontSize: "24px" }}
            />
          </Card>
        </Form>
      </Modal>
    </Layout>
  );
};

export default ProjectDetailPage;
