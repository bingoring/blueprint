import {
  ClockCircleOutlined,
  FallOutlined,
  MessageOutlined,
  PlusOutlined,
  RiseOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Button,
  Card,
  Col,
  Divider,
  Empty,
  Progress,
  Row,
  Space,
  Statistic,
  Table,
  Tabs,
  Tag,
  Timeline,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../stores/useAuthStore";
import GlobalNavbar from "./GlobalNavbar";
import {
  InvestmentIcon,
  MentoringIcon,
  MilestoneIcon,
  PathIcon,
  PortfolioIcon,
  TrophyIcon,
} from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;

interface MyProject {
  id: number;
  title: string;
  description: string;
  category: string;
  status: "funding" | "active" | "completed" | "paused";
  progress: number;
  totalMilestones: number;
  completedMilestones: number;
  currentMilestone: string;
  tvl: number;
  investorCount: number;
  createdAt: string;
  targetDate: string;
}

interface MyInvestment {
  id: number;
  projectTitle: string;
  milestoneTitle: string;
  option: "success" | "fail";
  amount: number;
  currentPrice: number;
  quantity: number;
  profitLoss: number;
  profitLossPercent: number;
  status: "active" | "completed" | "failed";
  investedAt: string;
  resolvedAt?: string;
}

interface MyMentoring {
  id: number;
  projectTitle: string;
  menteeUsername: string;
  status: "active" | "completed" | "paused";
  startedAt: string;
  completedAt?: string;
  rating?: number;
  totalSessions: number;
  lastSessionAt: string;
  milestoneTitle: string;
  menteeFeedback?: string;
}

interface ActivityData {
  overview: {
    totalProjects: number;
    activeProjects: number;
    totalInvestment: number;
    currentValue: number;
    profitLoss: number;
    profitLossPercent: number;
    activeMentoring: number;
    completedMentoring: number;
    totalEarnings: number;
  };
  projects: MyProject[];
  investments: MyInvestment[];
  mentoring: MyMentoring[];
  recentActivity: Array<{
    id: number;
    type: "project" | "investment" | "mentoring";
    title: string;
    description: string;
    timestamp: string;
    amount?: number;
  }>;
}

const ActivityPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [activityData, setActivityData] = useState<ActivityData | null>(null);
  const [activeTab, setActiveTab] = useState("dashboard");

  // Mock data for development
  useEffect(() => {
    const loadMockData = () => {
      const mockData: ActivityData = {
        overview: {
          totalProjects: 5,
          activeProjects: 2,
          totalInvestment: 25000,
          currentValue: 28750,
          profitLoss: 3750,
          profitLossPercent: 15,
          activeMentoring: 3,
          completedMentoring: 8,
          totalEarnings: 12500,
        },
        projects: [
          {
            id: 1,
            title: "AI 기반 피트니스 앱",
            description: "개인 맞춤형 운동 프로그램을 제공하는 AI 피트니스 앱",
            category: "IT/개발",
            status: "active",
            progress: 75,
            totalMilestones: 6,
            completedMilestones: 4,
            currentMilestone: "MVP 개발 완료",
            tvl: 45000,
            investorCount: 89,
            createdAt: "2024-01-15",
            targetDate: "2024-06-15",
          },
          {
            id: 2,
            title: "온라인 교육 플랫폼",
            description: "실시간 화상 강의 및 과제 관리 시스템",
            category: "교육",
            status: "completed",
            progress: 100,
            totalMilestones: 5,
            completedMilestones: 5,
            currentMilestone: "서비스 런칭",
            tvl: 32000,
            investorCount: 67,
            createdAt: "2023-10-01",
            targetDate: "2024-03-01",
          },
          {
            id: 3,
            title: "친환경 배달 서비스",
            description: "전기차와 자전거를 활용한 친환경 배달 서비스",
            category: "창업",
            status: "funding",
            progress: 15,
            totalMilestones: 8,
            completedMilestones: 1,
            currentMilestone: "시장 조사 완료",
            tvl: 12000,
            investorCount: 23,
            createdAt: "2024-02-20",
            targetDate: "2024-12-20",
          },
        ],
        investments: [
          {
            id: 1,
            projectTitle: "블록체인 투표 시스템",
            milestoneTitle: "스마트 컨트랙트 개발",
            option: "success",
            amount: 5000,
            currentPrice: 0.72,
            quantity: 6944,
            profitLoss: 1000,
            profitLossPercent: 20,
            status: "active",
            investedAt: "2024-01-20",
          },
          {
            id: 2,
            projectTitle: "AR 기반 인테리어 앱",
            milestoneTitle: "프로토타입 완성",
            option: "success",
            amount: 3000,
            currentPrice: 0.95,
            quantity: 3157,
            profitLoss: 850,
            profitLossPercent: 28.3,
            status: "completed",
            investedAt: "2024-01-10",
            resolvedAt: "2024-02-25",
          },
          {
            id: 3,
            projectTitle: "개인 브랜딩 과정",
            milestoneTitle: "SNS 팔로워 1만명 달성",
            option: "fail",
            amount: 2000,
            currentPrice: 0.25,
            quantity: 8000,
            profitLoss: -1500,
            profitLossPercent: -75,
            status: "failed",
            investedAt: "2024-01-05",
            resolvedAt: "2024-03-10",
          },
        ],
        mentoring: [
          {
            id: 1,
            projectTitle: "스타트업 창업 프로젝트",
            menteeUsername: "창업준비생",
            status: "active",
            startedAt: "2024-02-01",
            totalSessions: 8,
            lastSessionAt: "2024-03-15",
            milestoneTitle: "사업계획서 작성",
            rating: 5,
          },
          {
            id: 2,
            projectTitle: "웹 개발 포트폴리오",
            menteeUsername: "신입개발자",
            status: "completed",
            startedAt: "2024-01-10",
            completedAt: "2024-02-28",
            totalSessions: 12,
            lastSessionAt: "2024-02-28",
            milestoneTitle: "포트폴리오 완성",
            rating: 4.8,
            menteeFeedback:
              "정말 많은 도움이 되었습니다. 포트폴리오를 완성할 수 있었어요!",
          },
          {
            id: 3,
            projectTitle: "AI 모델 개발",
            menteeUsername: "데이터사이언티스트",
            status: "active",
            startedAt: "2024-03-01",
            totalSessions: 4,
            lastSessionAt: "2024-03-20",
            milestoneTitle: "모델 성능 최적화",
          },
        ],
        recentActivity: [
          {
            id: 1,
            type: "investment",
            title: "투자 수익 실현",
            description:
              '"AR 기반 인테리어 앱" 프로젝트에서 28.3% 수익을 기록했습니다.',
            timestamp: "2시간 전",
            amount: 850,
          },
          {
            id: 2,
            type: "mentoring",
            title: "멘토링 세션 완료",
            description:
              '"스타트업 창업 프로젝트" 멘토링 8회차 세션을 완료했습니다.',
            timestamp: "4시간 전",
          },
          {
            id: 3,
            type: "project",
            title: "마일스톤 진행 업데이트",
            description:
              '"AI 기반 피트니스 앱" 프로젝트의 MVP 개발이 75% 완료되었습니다.',
            timestamp: "1일 전",
          },
          {
            id: 4,
            type: "investment",
            title: "새로운 투자",
            description:
              '"친환경 에너지 솔루션" 프로젝트에 3,000 USDC를 투자했습니다.',
            timestamp: "2일 전",
            amount: 3000,
          },
        ],
      };

      setActivityData(mockData);
      setLoading(false);
    };

    setTimeout(loadMockData, 500);
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR").format(amount);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "active":
        return "green";
      case "completed":
        return "blue";
      case "funding":
        return "orange";
      case "paused":
        return "yellow";
      case "failed":
        return "red";
      default:
        return "default";
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "active":
        return "진행 중";
      case "completed":
        return "완료됨";
      case "funding":
        return "펀딩 중";
      case "paused":
        return "일시정지";
      case "failed":
        return "실패";
      default:
        return status;
    }
  };

  const getActivityIcon = (type: string) => {
    switch (type) {
      case "project":
        return <PathIcon size={16} color="#1890ff" />;
      case "investment":
        return <InvestmentIcon size={16} color="#52c41a" />;
      case "mentoring":
        return <MentoringIcon size={16} color="#722ed1" />;
      default:
        return <MilestoneIcon size={16} />;
    }
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
          <PortfolioIcon size={16} />
          대시보드
        </Space>
      ),
      children: (
        <div>
          {/* 요약 통계 */}
          <Row gutter={[24, 24]} style={{ marginBottom: "32px" }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 프로젝트"
                  value={activityData?.overview.totalProjects || 0}
                  prefix={<PathIcon size={16} />}
                  valueStyle={{ color: "var(--primary-color)" }}
                />
                <div style={{ marginTop: "8px" }}>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    활성: {activityData?.overview.activeProjects || 0}개
                  </Text>
                </div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 투자액"
                  value={activityData?.overview.totalInvestment || 0}
                  suffix="USDC"
                  prefix={<InvestmentIcon size={16} />}
                  valueStyle={{ color: "#722ed1" }}
                />
                <div style={{ marginTop: "8px" }}>
                  <Space>
                    {(activityData?.overview.profitLoss || 0) > 0 ? (
                      <RiseOutlined style={{ color: "#52c41a" }} />
                    ) : (
                      <FallOutlined style={{ color: "#ff4d4f" }} />
                    )}
                    <Text
                      style={{
                        color:
                          (activityData?.overview.profitLoss || 0) > 0
                            ? "#52c41a"
                            : "#ff4d4f",
                        fontSize: "12px",
                      }}
                    >
                      {(activityData?.overview.profitLoss || 0) > 0 ? "+" : ""}
                      {formatCurrency(
                        activityData?.overview.profitLoss || 0
                      )}{" "}
                      USDC
                    </Text>
                  </Space>
                </div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="멘토링"
                  value={activityData?.overview.activeMentoring || 0}
                  suffix="건"
                  prefix={<MentoringIcon size={16} />}
                  valueStyle={{ color: "#fa8c16" }}
                />
                <div style={{ marginTop: "8px" }}>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    완료: {activityData?.overview.completedMentoring || 0}건
                  </Text>
                </div>
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 수익"
                  value={activityData?.overview.totalEarnings || 0}
                  suffix="USDC"
                  prefix={<TrophyIcon size={16} />}
                  valueStyle={{ color: "#52c41a" }}
                />
                <div style={{ marginTop: "8px" }}>
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    투자 + 멘토링 수익
                  </Text>
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={[24, 24]}>
            {/* 최근 활동 */}
            <Col span={16}>
              <Card
                title="최근 활동"
                extra={
                  <Button type="link" onClick={() => setActiveTab("projects")}>
                    전체보기
                  </Button>
                }
              >
                <Timeline>
                  {activityData?.recentActivity.map((activity) => (
                    <Timeline.Item
                      key={activity.id}
                      dot={getActivityIcon(activity.type)}
                    >
                      <div>
                        <Text strong>{activity.title}</Text>
                        {activity.amount && (
                          <Tag
                            color={activity.amount > 0 ? "green" : "red"}
                            style={{ marginLeft: "8px" }}
                          >
                            {activity.amount > 0 ? "+" : ""}
                            {formatCurrency(activity.amount)} USDC
                          </Tag>
                        )}
                        <br />
                        <Text type="secondary">{activity.description}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: "12px" }}>
                          <ClockCircleOutlined /> {activity.timestamp}
                        </Text>
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </Card>
            </Col>

            {/* 진행 중인 프로젝트 */}
            <Col span={8}>
              <Card
                title="진행 중인 프로젝트"
                extra={
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => navigate("/projects/new")}
                    size="small"
                  >
                    새 프로젝트
                  </Button>
                }
              >
                <Space direction="vertical" style={{ width: "100%" }}>
                  {activityData?.projects
                    .filter((p) => p.status === "active")
                    .map((project) => (
                      <Card
                        key={project.id}
                        size="small"
                        hoverable
                        onClick={() => navigate(`/project/${project.id}`)}
                      >
                        <Space
                          direction="vertical"
                          size={4}
                          style={{ width: "100%" }}
                        >
                          <Text strong style={{ fontSize: "14px" }}>
                            {project.title}
                          </Text>
                          <Text type="secondary" style={{ fontSize: "12px" }}>
                            {project.currentMilestone}
                          </Text>
                          <div>
                            <Progress
                              percent={project.progress}
                              size="small"
                              strokeColor="var(--primary-color)"
                            />
                          </div>
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
                              투자자: {project.investorCount}명
                            </Text>
                          </div>
                        </Space>
                      </Card>
                    ))}
                  {!activityData?.projects.filter((p) => p.status === "active")
                    .length && (
                    <Empty
                      description="진행 중인 프로젝트가 없습니다"
                      style={{ margin: "20px 0" }}
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
      key: "projects",
      label: (
        <Space>
          <PathIcon size={16} />내 프로젝트
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
              내가 만든 프로젝트
            </Title>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => navigate("/projects/new")}
            >
              새 프로젝트 시작
            </Button>
          </div>

          <Row gutter={[16, 16]}>
            {activityData?.projects.map((project) => (
              <Col span={8} key={project.id}>
                <Card
                  hoverable
                  onClick={() => navigate(`/project/${project.id}`)}
                  extra={
                    <Tag color={getStatusColor(project.status)}>
                      {getStatusText(project.status)}
                    </Tag>
                  }
                >
                  <Space direction="vertical" style={{ width: "100%" }}>
                    <Text strong style={{ fontSize: "16px" }}>
                      {project.title}
                    </Text>
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      {project.description}
                    </Text>
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      카테고리: {project.category}
                    </Text>

                    <Divider style={{ margin: "8px 0" }} />

                    <div>
                      <div
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                          marginBottom: "4px",
                        }}
                      >
                        <Text style={{ fontSize: "12px" }}>전체 진행률</Text>
                        <Text style={{ fontSize: "12px" }}>
                          {project.progress}%
                        </Text>
                      </div>
                      <Progress
                        percent={project.progress}
                        size="small"
                        strokeColor="var(--primary-color)"
                      />
                    </div>

                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Text style={{ fontSize: "11px" }}>
                        마일스톤: {project.completedMilestones}/
                        {project.totalMilestones}
                      </Text>
                      <Text style={{ fontSize: "11px" }}>
                        TVL: {formatCurrency(project.tvl)}
                      </Text>
                    </div>

                    <Text type="secondary" style={{ fontSize: "11px" }}>
                      현재: {project.currentMilestone}
                    </Text>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>

          {!activityData?.projects.length && (
            <Empty
              description="아직 생성한 프로젝트가 없습니다"
              style={{ margin: "60px 0" }}
            >
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate("/projects/new")}
              >
                첫 프로젝트 시작하기
              </Button>
            </Empty>
          )}
        </div>
      ),
    },
    {
      key: "investments",
      label: (
        <Space>
          <InvestmentIcon size={16} />내 투자
        </Space>
      ),
      children: (
        <div>
          <div style={{ marginBottom: "24px" }}>
            <Row gutter={[24, 16]}>
              <Col span={6}>
                <Statistic
                  title="총 투자액"
                  value={activityData?.overview.totalInvestment || 0}
                  suffix="USDC"
                  valueStyle={{ fontSize: "20px" }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="현재 가치"
                  value={activityData?.overview.currentValue || 0}
                  suffix="USDC"
                  valueStyle={{ fontSize: "20px", color: "#52c41a" }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="총 손익"
                  value={activityData?.overview.profitLoss || 0}
                  suffix="USDC"
                  prefix={
                    activityData?.overview.profitLoss &&
                    activityData.overview.profitLoss > 0
                      ? "+"
                      : ""
                  }
                  valueStyle={{
                    fontSize: "20px",
                    color:
                      (activityData?.overview.profitLoss || 0) > 0
                        ? "#52c41a"
                        : "#ff4d4f",
                  }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="수익률"
                  value={activityData?.overview.profitLossPercent || 0}
                  suffix="%"
                  prefix={
                    activityData?.overview.profitLossPercent &&
                    activityData.overview.profitLossPercent > 0
                      ? "+"
                      : ""
                  }
                  valueStyle={{
                    fontSize: "20px",
                    color:
                      (activityData?.overview.profitLossPercent || 0) > 0
                        ? "#52c41a"
                        : "#ff4d4f",
                  }}
                />
              </Col>
            </Row>
          </div>

          <Table
            dataSource={activityData?.investments}
            rowKey="id"
            columns={[
              {
                title: "프로젝트/마일스톤",
                key: "project",
                render: (record: MyInvestment) => (
                  <div>
                    <Text strong>{record.projectTitle}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      {record.milestoneTitle}
                    </Text>
                  </div>
                ),
              },
              {
                title: "옵션",
                dataIndex: "option",
                render: (option: string) => (
                  <Tag color={option === "success" ? "green" : "red"}>
                    {option === "success" ? "성공" : "실패"}
                  </Tag>
                ),
              },
              {
                title: "투자액",
                dataIndex: "amount",
                render: (amount: number) => `${formatCurrency(amount)} USDC`,
              },
              {
                title: "현재가",
                dataIndex: "currentPrice",
                render: (price: number) => `$${price}`,
              },
              {
                title: "보유 수량",
                dataIndex: "quantity",
                render: (quantity: number) => formatCurrency(quantity),
              },
              {
                title: "손익",
                key: "profit",
                render: (record: MyInvestment) => (
                  <Space direction="vertical" size={0}>
                    <Text
                      style={{
                        color: record.profitLoss > 0 ? "#52c41a" : "#ff4d4f",
                        fontWeight: "bold",
                      }}
                    >
                      {record.profitLoss > 0 ? "+" : ""}
                      {formatCurrency(record.profitLoss)} USDC
                    </Text>
                    <Text
                      style={{
                        color:
                          record.profitLossPercent > 0 ? "#52c41a" : "#ff4d4f",
                        fontSize: "12px",
                      }}
                    >
                      ({record.profitLossPercent > 0 ? "+" : ""}
                      {record.profitLossPercent}%)
                    </Text>
                  </Space>
                ),
              },
              {
                title: "상태",
                dataIndex: "status",
                render: (status: string) => (
                  <Tag color={getStatusColor(status)}>
                    {getStatusText(status)}
                  </Tag>
                ),
              },
              {
                title: "투자일",
                dataIndex: "investedAt",
                render: (date: string) =>
                  new Date(date).toLocaleDateString("ko-KR"),
              },
            ]}
            pagination={{ pageSize: 10 }}
          />
        </div>
      ),
    },
    {
      key: "mentoring",
      label: (
        <Space>
          <MentoringIcon size={16} />내 멘토링
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
              멘토링 활동
            </Title>
            <Button
              type="primary"
              icon={<TeamOutlined />}
              onClick={() => navigate("/mentoring")}
            >
              멘토링 신청하기
            </Button>
          </div>

          <Row gutter={[16, 16]}>
            {activityData?.mentoring.map((mentoring) => (
              <Col span={8} key={mentoring.id}>
                <Card
                  hoverable
                  extra={
                    <Tag color={getStatusColor(mentoring.status)}>
                      {getStatusText(mentoring.status)}
                    </Tag>
                  }
                >
                  <Space direction="vertical" style={{ width: "100%" }}>
                    <Text strong style={{ fontSize: "16px" }}>
                      {mentoring.projectTitle}
                    </Text>
                    <div>
                      <Avatar size="small" style={{ marginRight: "8px" }}>
                        {mentoring.menteeUsername[0]}
                      </Avatar>
                      <Text>@{mentoring.menteeUsername}</Text>
                    </div>
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      목표: {mentoring.milestoneTitle}
                    </Text>

                    <Divider style={{ margin: "8px 0" }} />

                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Text style={{ fontSize: "12px" }}>
                        세션 수: {mentoring.totalSessions}회
                      </Text>
                      {mentoring.rating && (
                        <Text style={{ fontSize: "12px" }}>
                          평점: ⭐ {mentoring.rating}
                        </Text>
                      )}
                    </div>

                    <Text type="secondary" style={{ fontSize: "11px" }}>
                      최근 세션:{" "}
                      {new Date(mentoring.lastSessionAt).toLocaleDateString(
                        "ko-KR"
                      )}
                    </Text>

                    {mentoring.menteeFeedback && (
                      <div
                        style={{
                          background: "var(--bg-secondary)",
                          padding: "8px",
                          borderRadius: "6px",
                          marginTop: "8px",
                        }}
                      >
                        <Text style={{ fontSize: "12px", fontStyle: "italic" }}>
                          "{mentoring.menteeFeedback}"
                        </Text>
                      </div>
                    )}

                    <div style={{ marginTop: "12px" }}>
                      <Button
                        size="small"
                        icon={<MessageOutlined />}
                        onClick={() => navigate(`/mentoring/${mentoring.id}`)}
                      >
                        채팅방 입장
                      </Button>
                    </div>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>

          {!activityData?.mentoring.length && (
            <Empty
              description="아직 멘토링 활동이 없습니다"
              style={{ margin: "60px 0" }}
            >
              <Button
                type="primary"
                icon={<TeamOutlined />}
                onClick={() => navigate("/mentoring")}
              >
                멘토링 시작하기
              </Button>
            </Empty>
          )}
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
              <PortfolioIcon size={32} color="var(--primary-color)" />
              <div>
                <Title
                  level={2}
                  style={{ margin: 0, color: "var(--text-primary)" }}
                >
                  내 활동
                </Title>
                <Text type="secondary" style={{ fontSize: "16px" }}>
                  프로젝트, 투자, 멘토링 활동을 관리하세요
                </Text>
              </div>
            </Space>
          </div>

          {loading ? (
            <div style={{ textAlign: "center", padding: "100px" }}>
              <Text>활동 내역을 불러오는 중...</Text>
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
    </div>
  );
};

export default ActivityPage;
