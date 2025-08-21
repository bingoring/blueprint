import {
  BookOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  EyeOutlined,
  HeartOutlined,
  MessageOutlined,
  PlusOutlined,
  SearchOutlined,
  StarOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Divider,
  Empty,
  Form,
  Input,
  Modal,
  Progress,
  Rate,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tabs,
  Tag,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../stores/useAuthStore";
import GlobalNavbar from "./GlobalNavbar";
import { MentoringIcon } from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;
const { Option } = Select;

interface MentorStats {
  totalSessions: number;
  activeSessions: number;
  completedSessions: number;
  averageRating: number;
  totalEarnings: number;
  successRate: number;
  responseTime: number; // hours
  expertise: string[];
}

interface MentoringRequest {
  id: number;
  projectTitle: string;
  requesterUsername: string;
  requesterAvatar?: string;
  milestoneTitle: string;
  description: string;
  category: string;
  expectedDuration: number; // weeks
  budget: number;
  requestedAt: string;
  urgency: "low" | "medium" | "high";
  status: "pending" | "accepted" | "rejected";
}

interface ActiveSession {
  id: number;
  projectTitle: string;
  menteeUsername: string;
  menteeAvatar?: string;
  milestoneTitle: string;
  startedAt: string;
  totalSessions: number;
  completedSessions: number;
  nextSessionAt?: string;
  lastActivity: string;
  progress: number;
  weeklyHours: number;
}

interface SessionHistory {
  id: number;
  projectTitle: string;
  menteeUsername: string;
  startedAt: string;
  completedAt: string;
  totalSessions: number;
  duration: number; // weeks
  rating: number;
  earnings: number;
  feedback: string;
  outcome: "success" | "incomplete" | "cancelled";
}

interface MentoringData {
  stats: MentorStats;
  requests: MentoringRequest[];
  activeSessions: ActiveSession[];
  history: SessionHistory[];
  availableProjects: Array<{
    id: number;
    title: string;
    description: string;
    category: string;
    creatorUsername: string;
    milestoneTitle: string;
    tvl: number;
    difficulty: "beginner" | "intermediate" | "advanced";
    estimatedHours: number;
  }>;
}

const MentoringPage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [mentoringData, setMentoringData] = useState<MentoringData | null>(
    null
  );
  const [activeTab, setActiveTab] = useState("dashboard");
  const [applicationModalVisible, setApplicationModalVisible] = useState(false);
  const [selectedProject, setSelectedProject] = useState<any>(null);
  const [form] = Form.useForm();

  // Mock data for development
  useEffect(() => {
    const loadMockData = () => {
      const mockData: MentoringData = {
        stats: {
          totalSessions: 45,
          activeSessions: 5,
          completedSessions: 40,
          averageRating: 4.8,
          totalEarnings: 15750,
          successRate: 92,
          responseTime: 2.5,
          expertise: ["스타트업", "IT개발", "마케팅", "프로젝트 관리"],
        },
        requests: [
          {
            id: 1,
            projectTitle: "AI 기반 개인 금융 관리 앱",
            requesterUsername: "fintech_dreamer",
            milestoneTitle: "MVP 프로토타입 개발",
            description:
              "React Native를 사용한 모바일 앱 개발에 대한 멘토링이 필요합니다. 특히 AI 모델 연동 부분에서 도움이 필요합니다.",
            category: "IT/개발",
            expectedDuration: 8,
            budget: 2000,
            requestedAt: "2024-03-20",
            urgency: "high",
            status: "pending",
          },
          {
            id: 2,
            projectTitle: "친환경 배달 서비스 스타트업",
            requesterUsername: "green_delivery",
            milestoneTitle: "시장 검증 및 사업계획서 작성",
            description:
              "스타트업 초기 단계에서 시장 분석과 사업계획서 작성에 대한 멘토링을 요청합니다.",
            category: "창업",
            expectedDuration: 6,
            budget: 1500,
            requestedAt: "2024-03-19",
            urgency: "medium",
            status: "pending",
          },
          {
            id: 3,
            projectTitle: "온라인 교육 플랫폼",
            requesterUsername: "edu_innovator",
            milestoneTitle: "사용자 경험 최적화",
            description:
              "기존 교육 플랫폼의 UX/UI 개선과 사용자 참여도 향상을 위한 전략에 대해 멘토링 받고 싶습니다.",
            category: "교육",
            expectedDuration: 4,
            budget: 1200,
            requestedAt: "2024-03-18",
            urgency: "low",
            status: "pending",
          },
        ],
        activeSessions: [
          {
            id: 1,
            projectTitle: "블록체인 투표 시스템",
            menteeUsername: "blockchain_dev",
            milestoneTitle: "스마트 컨트랙트 개발",
            startedAt: "2024-02-15",
            totalSessions: 12,
            completedSessions: 8,
            nextSessionAt: "2024-03-22T14:00:00",
            lastActivity: "2024-03-20T16:30:00",
            progress: 67,
            weeklyHours: 3,
          },
          {
            id: 2,
            projectTitle: "개인 브랜딩 컨설팅",
            menteeUsername: "personal_brand",
            milestoneTitle: "SNS 전략 수립",
            startedAt: "2024-03-01",
            totalSessions: 8,
            completedSessions: 3,
            nextSessionAt: "2024-03-23T10:00:00",
            lastActivity: "2024-03-19T11:15:00",
            progress: 38,
            weeklyHours: 2,
          },
          {
            id: 3,
            projectTitle: "스타트업 창업 프로젝트",
            menteeUsername: "startup_founder",
            milestoneTitle: "투자 유치 준비",
            startedAt: "2024-01-20",
            totalSessions: 16,
            completedSessions: 14,
            nextSessionAt: "2024-03-24T15:00:00",
            lastActivity: "2024-03-21T09:45:00",
            progress: 88,
            weeklyHours: 4,
          },
        ],
        history: [
          {
            id: 1,
            projectTitle: "웹 개발 포트폴리오",
            menteeUsername: "junior_dev",
            startedAt: "2024-01-10",
            completedAt: "2024-02-28",
            totalSessions: 12,
            duration: 7,
            rating: 5,
            earnings: 1800,
            feedback:
              "정말 체계적이고 실질적인 도움을 받았습니다. 덕분에 포트폴리오를 완성하고 취업에 성공했어요!",
            outcome: "success",
          },
          {
            id: 2,
            projectTitle: "모바일 게임 개발",
            menteeUsername: "game_creator",
            startedAt: "2023-11-15",
            completedAt: "2024-01-05",
            totalSessions: 10,
            duration: 8,
            rating: 4.5,
            earnings: 1500,
            feedback:
              "게임 개발의 전반적인 프로세스를 이해할 수 있었습니다. 매우 만족합니다.",
            outcome: "success",
          },
          {
            id: 3,
            projectTitle: "이커머스 플랫폼",
            menteeUsername: "ecommerce_builder",
            startedAt: "2023-10-01",
            completedAt: "2023-11-10",
            totalSessions: 6,
            duration: 6,
            rating: 4.8,
            earnings: 1200,
            feedback: "실무 경험을 바탕으로 한 조언이 정말 유용했습니다.",
            outcome: "success",
          },
        ],
        availableProjects: [
          {
            id: 1,
            title: "AR 기반 인테리어 앱",
            description: "증강현실 기술을 활용한 가구 배치 시뮬레이션 앱 개발",
            category: "IT/개발",
            creatorUsername: "ar_designer",
            milestoneTitle: "AR 엔진 최적화",
            tvl: 35000,
            difficulty: "advanced",
            estimatedHours: 40,
          },
          {
            id: 2,
            title: "친환경 패션 브랜드",
            description: "지속가능한 패션 브랜드 런칭 프로젝트",
            category: "창업",
            creatorUsername: "eco_fashion",
            milestoneTitle: "브랜드 아이덴티티 구축",
            tvl: 28000,
            difficulty: "intermediate",
            estimatedHours: 25,
          },
          {
            id: 3,
            title: "농업 IoT 시스템",
            description: "스마트 농업을 위한 IoT 센서 네트워크 구축",
            category: "IT/개발",
            creatorUsername: "smart_farmer",
            milestoneTitle: "센서 데이터 분석 시스템",
            tvl: 42000,
            difficulty: "advanced",
            estimatedHours: 35,
          },
        ],
      };

      setMentoringData(mockData);
      setLoading(false);
    };

    setTimeout(loadMockData, 500);
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR").format(amount);
  };

  const getUrgencyColor = (urgency: string) => {
    switch (urgency) {
      case "high":
        return "red";
      case "medium":
        return "orange";
      case "low":
        return "green";
      default:
        return "default";
    }
  };

  const getUrgencyText = (urgency: string) => {
    switch (urgency) {
      case "high":
        return "긴급";
      case "medium":
        return "보통";
      case "low":
        return "여유";
      default:
        return urgency;
    }
  };

  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case "beginner":
        return "green";
      case "intermediate":
        return "orange";
      case "advanced":
        return "red";
      default:
        return "default";
    }
  };

  const getDifficultyText = (difficulty: string) => {
    switch (difficulty) {
      case "beginner":
        return "초급";
      case "intermediate":
        return "중급";
      case "advanced":
        return "고급";
      default:
        return difficulty;
    }
  };

  const handleRequestResponse = (
    requestId: number,
    action: "accept" | "reject"
  ) => {
    // Mock response handling
    console.log(`${action} request ${requestId}`);
  };

  const handleApplyForMentoring = (project: any) => {
    setSelectedProject(project);
    setApplicationModalVisible(true);
  };

  const handleSubmitApplication = async (values: any) => {
    // Mock application submission
    console.log("Application submitted:", values, selectedProject);
    setApplicationModalVisible(false);
    form.resetFields();
  };

  if (!isAuthenticated) {
    navigate("/login");
    return null;
  }

  const tabItems = [
    {
      key: "dashboard",
      label: (
        <Space>
          <MentoringIcon size={16} />
          멘토 대시보드
        </Space>
      ),
      children: (
        <div>
          {/* 통계 카드 */}
          <Row gutter={[24, 24]} style={{ marginBottom: "32px" }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 멘토링 세션"
                  value={mentoringData?.stats.totalSessions || 0}
                  prefix={<TeamOutlined />}
                  valueStyle={{ color: "var(--primary-color)" }}
                />
                <div style={{ marginTop: "8px" }}>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    활성: {mentoringData?.stats.activeSessions || 0}개
                  </Text>
                </div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="평균 평점"
                  value={mentoringData?.stats.averageRating || 0}
                  suffix="/5.0"
                  prefix={<StarOutlined />}
                  valueStyle={{ color: "#faad14" }}
                  precision={1}
                />
                <div style={{ marginTop: "8px" }}>
                  <Rate
                    disabled
                    value={mentoringData?.stats.averageRating || 0}
                    style={{ fontSize: "12px" }}
                  />
                </div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 수익"
                  value={mentoringData?.stats.totalEarnings || 0}
                  suffix="USDC"
                  prefix={<DollarOutlined />}
                  valueStyle={{ color: "#52c41a" }}
                />
                <div style={{ marginTop: "8px" }}>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    성공률: {mentoringData?.stats.successRate || 0}%
                  </Text>
                </div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="응답 시간"
                  value={mentoringData?.stats.responseTime || 0}
                  suffix="시간"
                  prefix={<ClockCircleOutlined />}
                  valueStyle={{ color: "#722ed1" }}
                  precision={1}
                />
                <div style={{ marginTop: "8px" }}>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    평균 응답 속도
                  </Text>
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={[24, 24]}>
            {/* 전문 분야 */}
            <Col span={8}>
              <Card title="전문 분야" style={{ height: "300px" }}>
                <Space wrap>
                  {mentoringData?.stats.expertise.map((skill, index) => (
                    <Tag
                      key={index}
                      color="blue"
                      style={{ marginBottom: "8px" }}
                    >
                      {skill}
                    </Tag>
                  ))}
                </Space>
                <Divider />
                <Button
                  type="dashed"
                  icon={<PlusOutlined />}
                  style={{ width: "100%" }}
                >
                  전문 분야 추가
                </Button>
              </Card>
            </Col>

            {/* 최근 활동 */}
            <Col span={16}>
              <Card title="활성 멘토링 세션" style={{ height: "300px" }}>
                <Space direction="vertical" style={{ width: "100%" }}>
                  {mentoringData?.activeSessions.slice(0, 3).map((session) => (
                    <Card
                      key={session.id}
                      size="small"
                      hoverable
                      onClick={() =>
                        navigate(`/mentoring/session/${session.id}`)
                      }
                    >
                      <Row align="middle">
                        <Col span={16}>
                          <Space direction="vertical" size={2}>
                            <Text strong>{session.projectTitle}</Text>
                            <Text type="secondary" style={{ fontSize: "12px" }}>
                              @{session.menteeUsername} •{" "}
                              {session.milestoneTitle}
                            </Text>
                            <Progress
                              percent={session.progress}
                              size="small"
                              strokeColor="var(--primary-color)"
                            />
                          </Space>
                        </Col>
                        <Col span={8} style={{ textAlign: "right" }}>
                          <div>
                            <Text style={{ fontSize: "12px" }}>
                              {session.completedSessions}/
                              {session.totalSessions} 세션
                            </Text>
                            <br />
                            <Text type="secondary" style={{ fontSize: "11px" }}>
                              주 {session.weeklyHours}시간
                            </Text>
                          </div>
                        </Col>
                      </Row>
                    </Card>
                  ))}
                  {!mentoringData?.activeSessions.length && (
                    <Empty
                      description="활성 멘토링 세션이 없습니다"
                      style={{ margin: "40px 0" }}
                    />
                  )}
                </Space>
              </Card>
            </Col>
          </Row>
        </div>
      ),
    },
    {
      key: "requests",
      label: (
        <Space>
          <HeartOutlined />
          멘토링 요청
          <Badge
            count={
              mentoringData?.requests.filter((r) => r.status === "pending")
                .length || 0
            }
          />
        </Space>
      ),
      children: (
        <div>
          <div
            style={{
              marginBottom: "16px",
              display: "flex",
              justifyContent: "space-between",
            }}
          >
            <Title level={4} style={{ margin: 0 }}>
              받은 멘토링 요청
            </Title>
            <Button
              type="primary"
              onClick={() => setActiveTab("opportunities")}
            >
              새로운 기회 찾기
            </Button>
          </div>

          <Space direction="vertical" style={{ width: "100%" }}>
            {mentoringData?.requests.map((request) => (
              <Card key={request.id}>
                <Row gutter={24}>
                  <Col span={16}>
                    <Space direction="vertical" size={8}>
                      <div
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 12,
                        }}
                      >
                        <Avatar>
                          {request.requesterUsername[0].toUpperCase()}
                        </Avatar>
                        <div>
                          <Text strong style={{ fontSize: "16px" }}>
                            {request.projectTitle}
                          </Text>
                          <br />
                          <Text type="secondary">
                            @{request.requesterUsername}
                          </Text>
                        </div>
                      </div>

                      <Text strong>목표: {request.milestoneTitle}</Text>
                      <Paragraph style={{ margin: 0 }}>
                        {request.description}
                      </Paragraph>

                      <Space wrap>
                        <Tag>{request.category}</Tag>
                        <Tag color={getUrgencyColor(request.urgency)}>
                          {getUrgencyText(request.urgency)}
                        </Tag>
                        <Tag color="purple">
                          {request.expectedDuration}주 예상
                        </Tag>
                      </Space>
                    </Space>
                  </Col>

                  <Col span={8}>
                    <Space
                      direction="vertical"
                      style={{ width: "100%", textAlign: "right" }}
                    >
                      <Statistic
                        title="제안 예산"
                        value={request.budget}
                        suffix="USDC"
                        valueStyle={{ fontSize: "18px", color: "#52c41a" }}
                      />

                      <Text type="secondary" style={{ fontSize: "12px" }}>
                        요청일:{" "}
                        {new Date(request.requestedAt).toLocaleDateString(
                          "ko-KR"
                        )}
                      </Text>

                      <Space>
                        <Button
                          type="primary"
                          onClick={() =>
                            handleRequestResponse(request.id, "accept")
                          }
                        >
                          수락
                        </Button>
                        <Button
                          onClick={() =>
                            handleRequestResponse(request.id, "reject")
                          }
                        >
                          거절
                        </Button>
                      </Space>
                    </Space>
                  </Col>
                </Row>
              </Card>
            ))}

            {!mentoringData?.requests.length && (
              <Empty
                description="새로운 멘토링 요청이 없습니다"
                style={{ margin: "60px 0" }}
              />
            )}
          </Space>
        </div>
      ),
    },
    {
      key: "sessions",
      label: (
        <Space>
          <MessageOutlined />
          활성 세션
          <Badge count={mentoringData?.activeSessions.length || 0} />
        </Space>
      ),
      children: (
        <div>
          <Row gutter={[16, 16]}>
            {mentoringData?.activeSessions.map((session) => (
              <Col span={8} key={session.id}>
                <Card
                  hoverable
                  onClick={() => navigate(`/mentoring/session/${session.id}`)}
                >
                  <Space direction="vertical" style={{ width: "100%" }}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Text strong style={{ fontSize: "16px" }}>
                        {session.projectTitle}
                      </Text>
                      <Tag color="green">진행 중</Tag>
                    </div>

                    <div
                      style={{ display: "flex", alignItems: "center", gap: 8 }}
                    >
                      <Avatar size="small">
                        {session.menteeUsername[0].toUpperCase()}
                      </Avatar>
                      <Text>@{session.menteeUsername}</Text>
                    </div>

                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      목표: {session.milestoneTitle}
                    </Text>

                    <div>
                      <div
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                          marginBottom: "4px",
                        }}
                      >
                        <Text style={{ fontSize: "12px" }}>진행률</Text>
                        <Text style={{ fontSize: "12px" }}>
                          {session.progress}%
                        </Text>
                      </div>
                      <Progress
                        percent={session.progress}
                        size="small"
                        strokeColor="var(--primary-color)"
                      />
                    </div>

                    <Divider style={{ margin: "8px 0" }} />

                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Text style={{ fontSize: "11px" }}>
                        세션: {session.completedSessions}/
                        {session.totalSessions}
                      </Text>
                      <Text style={{ fontSize: "11px" }}>
                        주 {session.weeklyHours}시간
                      </Text>
                    </div>

                    {session.nextSessionAt && (
                      <div
                        style={{
                          background: "var(--bg-secondary)",
                          padding: "6px 8px",
                          borderRadius: "4px",
                          marginTop: "8px",
                        }}
                      >
                        <Text style={{ fontSize: "11px" }}>
                          <CalendarOutlined /> 다음 세션:{" "}
                          {new Date(session.nextSessionAt).toLocaleString(
                            "ko-KR"
                          )}
                        </Text>
                      </div>
                    )}

                    <Button
                      type="primary"
                      icon={<MessageOutlined />}
                      style={{ marginTop: "12px" }}
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/mentoring/session/${session.id}/chat`);
                      }}
                    >
                      채팅방 입장
                    </Button>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>

          {!mentoringData?.activeSessions.length && (
            <Empty
              description="활성 멘토링 세션이 없습니다"
              style={{ margin: "60px 0" }}
            >
              <Button
                type="primary"
                onClick={() => setActiveTab("opportunities")}
              >
                새로운 멘토링 시작하기
              </Button>
            </Empty>
          )}
        </div>
      ),
    },
    {
      key: "history",
      label: (
        <Space>
          <BookOutlined />
          세션 기록
        </Space>
      ),
      children: (
        <div>
          <Table
            dataSource={mentoringData?.history}
            rowKey="id"
            columns={[
              {
                title: "프로젝트",
                key: "project",
                render: (record: SessionHistory) => (
                  <div>
                    <Text strong>{record.projectTitle}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      @{record.menteeUsername}
                    </Text>
                  </div>
                ),
              },
              {
                title: "기간",
                key: "duration",
                render: (record: SessionHistory) => (
                  <div>
                    <Text>{record.duration}주</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "11px" }}>
                      {new Date(record.startedAt).toLocaleDateString("ko-KR")} ~
                      <br />
                      {new Date(record.completedAt).toLocaleDateString("ko-KR")}
                    </Text>
                  </div>
                ),
              },
              {
                title: "세션 수",
                dataIndex: "totalSessions",
                render: (sessions: number) => `${sessions}회`,
              },
              {
                title: "평점",
                key: "rating",
                render: (record: SessionHistory) => (
                  <Space direction="vertical" size={2}>
                    <Rate
                      disabled
                      value={record.rating}
                      style={{ fontSize: "12px" }}
                    />
                    <Text style={{ fontSize: "11px" }}>
                      {record.rating}/5.0
                    </Text>
                  </Space>
                ),
              },
              {
                title: "수익",
                dataIndex: "earnings",
                render: (earnings: number) => (
                  <Text style={{ color: "#52c41a", fontWeight: "bold" }}>
                    {formatCurrency(earnings)} USDC
                  </Text>
                ),
              },
              {
                title: "결과",
                dataIndex: "outcome",
                render: (outcome: string) => {
                  const colors = {
                    success: "green",
                    incomplete: "orange",
                    cancelled: "red",
                  };
                  const texts = {
                    success: "성공 완료",
                    incomplete: "미완료",
                    cancelled: "취소됨",
                  };
                  return (
                    <Tag color={colors[outcome as keyof typeof colors]}>
                      {texts[outcome as keyof typeof texts]}
                    </Tag>
                  );
                },
              },
              {
                title: "액션",
                key: "action",
                render: (record: SessionHistory) => (
                  <Button
                    size="small"
                    icon={<EyeOutlined />}
                    onClick={() =>
                      navigate(`/mentoring/session/${record.id}/review`)
                    }
                  >
                    상세보기
                  </Button>
                ),
              },
            ]}
            pagination={{ pageSize: 10 }}
            expandable={{
              expandedRowRender: (record) => (
                <div style={{ margin: 0 }}>
                  <Paragraph
                    style={{
                      fontStyle: "italic",
                      background: "var(--bg-secondary)",
                      padding: "12px",
                      borderRadius: "6px",
                    }}
                  >
                    "{record.feedback}"
                  </Paragraph>
                </div>
              ),
              rowExpandable: (record) => !!record.feedback,
            }}
          />
        </div>
      ),
    },
    {
      key: "opportunities",
      label: (
        <Space>
          <SearchOutlined />
          새로운 기회
        </Space>
      ),
      children: (
        <div>
          <div style={{ marginBottom: "24px" }}>
            <Title level={4} style={{ marginBottom: "16px" }}>
              멘토링 기회 찾기
            </Title>
            <Search
              placeholder="프로젝트 검색..."
              size="large"
              style={{ maxWidth: 400 }}
            />
          </div>

          <Row gutter={[16, 16]}>
            {mentoringData?.availableProjects.map((project) => (
              <Col span={8} key={project.id}>
                <Card hoverable>
                  <Space direction="vertical" style={{ width: "100%" }}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Text strong style={{ fontSize: "16px" }}>
                        {project.title}
                      </Text>
                      <Tag color={getDifficultyColor(project.difficulty)}>
                        {getDifficultyText(project.difficulty)}
                      </Tag>
                    </div>

                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      {project.description}
                    </Text>

                    <div
                      style={{ display: "flex", alignItems: "center", gap: 8 }}
                    >
                      <Avatar size="small">
                        {project.creatorUsername[0].toUpperCase()}
                      </Avatar>
                      <Text style={{ fontSize: "12px" }}>
                        @{project.creatorUsername}
                      </Text>
                    </div>

                    <Text strong style={{ fontSize: "14px" }}>
                      목표: {project.milestoneTitle}
                    </Text>

                    <Divider style={{ margin: "8px 0" }} />

                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Text style={{ fontSize: "11px" }}>
                        TVL: {formatCurrency(project.tvl)} USDC
                      </Text>
                      <Text style={{ fontSize: "11px" }}>
                        예상: {project.estimatedHours}시간
                      </Text>
                    </div>

                    <Tag color="blue" style={{ alignSelf: "flex-start" }}>
                      {project.category}
                    </Tag>

                    <Button
                      type="primary"
                      onClick={() => handleApplyForMentoring(project)}
                      style={{ marginTop: "8px" }}
                    >
                      멘토링 신청
                    </Button>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      ),
    },
  ];

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
          {/* 헤더 */}
          <div style={{ marginBottom: "32px" }}>
            <Space align="start" size={16}>
              <MentoringIcon size={32} color="var(--primary-color)" />
              <div>
                <Title
                  level={2}
                  style={{ margin: 0, color: "var(--text-primary)" }}
                >
                  멘토링
                </Title>
                <Text type="secondary" style={{ fontSize: "16px" }}>
                  지식을 나누고 성장을 도우며 보상을 받으세요
                </Text>
              </div>
            </Space>
          </div>

          {loading ? (
            <div style={{ textAlign: "center", padding: "100px" }}>
              <Text>멘토링 정보를 불러오는 중...</Text>
            </div>
          ) : (
            <Card>
              <Tabs
                activeKey={activeTab}
                onChange={setActiveTab}
                items={tabItems}
                size="large"
              />
            </Card>
          )}
        </div>
      </div>

      {/* 멘토링 신청 모달 */}
      <Modal
        title="멘토링 신청"
        open={applicationModalVisible}
        onCancel={() => setApplicationModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmitApplication}>
          <Form.Item
            label="신청 메시지"
            name="message"
            rules={[{ required: true, message: "신청 메시지를 입력해주세요" }]}
          >
            <Input.TextArea
              rows={4}
              placeholder="이 프로젝트에 어떤 도움을 제공할 수 있는지 설명해주세요"
            />
          </Form.Item>

          <Form.Item
            label="예상 멘토링 기간"
            name="duration"
            rules={[{ required: true, message: "예상 기간을 선택해주세요" }]}
          >
            <Select placeholder="멘토링 기간 선택">
              <Option value={4}>4주</Option>
              <Option value={6}>6주</Option>
              <Option value={8}>8주</Option>
              <Option value={12}>12주</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="주당 시간"
            name="weeklyHours"
            rules={[{ required: true, message: "주당 시간을 선택해주세요" }]}
          >
            <Select placeholder="주당 멘토링 시간">
              <Option value={2}>주 2시간</Option>
              <Option value={3}>주 3시간</Option>
              <Option value={4}>주 4시간</Option>
              <Option value={5}>주 5시간</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                신청하기
              </Button>
              <Button onClick={() => setApplicationModalVisible(false)}>
                취소
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default MentoringPage;
