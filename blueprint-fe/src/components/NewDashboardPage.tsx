import {
  ArrowRightOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  FireOutlined,
  PlusOutlined,
  RiseOutlined,
} from "@ant-design/icons";
import {
  Button,
  Card,
  Col,
  Divider,
  List,
  Progress,
  Row,
  Space,
  Statistic,
  Tag,
  Timeline,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../stores/useAuthStore";
import GlobalNavbar from "./GlobalNavbar";
import {
  ConnectionIcon,
  InvestmentIcon,
  MilestoneIcon,
  PathIcon,
  ProgressIcon,
  RocketIcon,
} from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;

interface NextMilestone {
  id: number;
  title: string;
  description: string;
  daysLeft: number;
  progress: number;
  projectTitle: string;
  isOwner: boolean;
}

interface ActivityFeedItem {
  id: number;
  type:
    | "investment"
    | "mentor_feedback"
    | "milestone_update"
    | "new_investment";
  title: string;
  description: string;
  timestamp: string;
  amount?: number;
  projectTitle?: string;
  avatar?: string;
}

interface PortfolioStats {
  totalInvestment: number;
  currentValue: number;
  profitLoss: number;
  profitLossPercent: number;
  totalProjects: number;
  activeProjects: number;
}

interface RecommendedProject {
  id: number;
  title: string;
  description: string;
  creator: string;
  currentPrice: number;
  tvl: number;
  category: string;
  daysLeft: number;
  isHot: boolean;
}

const NewDashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuthStore();

  const [nextMilestone, setNextMilestone] = useState<NextMilestone | null>(
    null
  );
  const [activityFeed, setActivityFeed] = useState<ActivityFeedItem[]>([]);
  const [portfolioStats, setPortfolioStats] = useState<PortfolioStats | null>(
    null
  );
  const [recommendedProjects, setRecommendedProjects] = useState<
    RecommendedProject[]
  >([]);

  // Mock data for development
  useEffect(() => {
    const loadMockData = () => {
      // Mock next milestone
      setNextMilestone({
        id: 1,
        title: "MVP 개발 완료",
        description:
          "기본 기능을 포함한 최소 기능 제품(MVP) 개발을 완료합니다.",
        daysLeft: 7,
        progress: 75,
        projectTitle: "AI 기반 피트니스 앱",
        isOwner: true,
      });

      // Mock activity feed
      setActivityFeed([
        {
          id: 1,
          type: "new_investment",
          title: "새로운 투자 발생",
          description:
            '"블록체인 학습 플랫폼" 프로젝트에 새로운 투자자가 참여했습니다.',
          timestamp: "2시간 전",
          amount: 500,
          projectTitle: "블록체인 학습 플랫폼",
        },
        {
          id: 2,
          type: "mentor_feedback",
          title: "멘토 피드백 도착",
          description: "김민수 멘토님이 코드 리뷰 피드백을 남겨주셨습니다.",
          timestamp: "4시간 전",
          projectTitle: "AI 기반 피트니스 앱",
        },
        {
          id: 3,
          type: "milestone_update",
          title: "마일스톤 진행 업데이트",
          description: "프로토타입 개발이 90% 완료되었습니다.",
          timestamp: "1일 전",
          projectTitle: "AI 기반 피트니스 앱",
        },
        {
          id: 4,
          type: "investment",
          title: "투자 수익 실현",
          description:
            '"온라인 교육 플랫폼" 프로젝트에서 +15% 수익을 기록했습니다.',
          timestamp: "2일 전",
          amount: 750,
          projectTitle: "온라인 교육 플랫폼",
        },
      ]);

      // Mock portfolio stats
      setPortfolioStats({
        totalInvestment: 25000,
        currentValue: 28750,
        profitLoss: 3750,
        profitLossPercent: 15,
        totalProjects: 8,
        activeProjects: 3,
      });

      // Mock recommended projects
      setRecommendedProjects([
        {
          id: 1,
          title: "친환경 배달 서비스 플랫폼",
          description:
            "전기차와 자전거를 활용한 친환경 배달 서비스를 개발합니다.",
          creator: "박지현",
          currentPrice: 0.68,
          tvl: 45000,
          category: "창업",
          daysLeft: 14,
          isHot: true,
        },
        {
          id: 2,
          title: "AR 기반 인테리어 앱",
          description:
            "증강현실을 활용한 가구 배치 시뮬레이션 앱을 제작합니다.",
          creator: "이상민",
          currentPrice: 0.72,
          tvl: 32000,
          category: "IT/개발",
          daysLeft: 8,
          isHot: false,
        },
        {
          id: 3,
          title: "개인 브랜딩 마스터 과정",
          description:
            "6개월 만에 개인 브랜드를 구축하고 수익화하는 프로젝트입니다.",
          creator: "김영희",
          currentPrice: 0.55,
          tvl: 28000,
          category: "라이프스타일",
          daysLeft: 21,
          isHot: false,
        },
      ]);
    };

    setTimeout(loadMockData, 500);
  }, []);

  const getActivityIcon = (type: string) => {
    switch (type) {
      case "investment":
        return <InvestmentIcon size={16} color="#52c41a" />;
      case "mentor_feedback":
        return <ConnectionIcon size={16} color="#1890ff" />;
      case "milestone_update":
        return <MilestoneIcon size={16} color="#faad14" />;
      case "new_investment":
        return <DollarOutlined style={{ color: "#722ed1" }} />;
      default:
        return <PathIcon size={16} />;
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR").format(amount);
  };

  if (!isAuthenticated) {
    navigate("/login");
    return null;
  }

  return (
    <div style={{ background: "var(--bg-primary)", minHeight: "100vh" }}>
      <GlobalNavbar />

      {/* 메인 컨텐츠 */}
      <div style={{ paddingTop: "64px" }}>
        <div
          style={{
            maxWidth: "1400px",
            margin: "0 auto",
            padding: "32px 24px",
          }}
        >
          {/* 환영 헤더 */}
          <div style={{ marginBottom: "32px" }}>
            <Title
              level={2}
              style={{ margin: 0, color: "var(--text-primary)" }}
            >
              안녕하세요, {user?.username || "사용자"}님! 👋
            </Title>
            <Text type="secondary" style={{ fontSize: "16px" }}>
              오늘도 목표를 향해 한 걸음 더 나아가볼까요?
            </Text>
          </div>

          <Row gutter={[24, 24]}>
            {/* 좌측 컬럼 */}
            <Col span={16}>
              {/* 나의 다음 마일스톤 */}
              <Card
                style={{ marginBottom: "24px" }}
                title={
                  <Space>
                    <MilestoneIcon size={20} color="var(--primary-color)" />
                    <span>🎯 나의 다음 마일스톤</span>
                  </Space>
                }
                extra={
                  <Button
                    type="link"
                    icon={<ArrowRightOutlined />}
                    onClick={() => navigate(`/project/${nextMilestone?.id}`)}
                  >
                    상세보기
                  </Button>
                }
              >
                {nextMilestone ? (
                  <div>
                    <Row align="middle" gutter={16}>
                      <Col span={16}>
                        <Space direction="vertical" size={8}>
                          <Title level={4} style={{ margin: 0 }}>
                            {nextMilestone.title}
                          </Title>
                          <Text type="secondary">
                            {nextMilestone.projectTitle}
                          </Text>
                          <Paragraph style={{ margin: 0 }}>
                            {nextMilestone.description}
                          </Paragraph>
                        </Space>
                      </Col>
                      <Col span={8} style={{ textAlign: "center" }}>
                        <ProgressIcon
                          size={80}
                          color="var(--primary-color)"
                          progress={nextMilestone.progress / 100}
                        />
                        <div style={{ marginTop: "8px" }}>
                          <Text
                            strong
                            style={{
                              fontSize: "18px",
                              color: "var(--primary-color)",
                            }}
                          >
                            D-{nextMilestone.daysLeft}
                          </Text>
                        </div>
                      </Col>
                    </Row>
                    <div style={{ marginTop: "16px" }}>
                      <Progress
                        percent={nextMilestone.progress}
                        strokeColor="var(--primary-color)"
                        showInfo={false}
                      />
                      <div style={{ marginTop: "12px" }}>
                        <Button
                          type="primary"
                          icon={<PlusOutlined />}
                          onClick={() =>
                            navigate(`/project/${nextMilestone.id}/update`)
                          }
                        >
                          진행 상황 업데이트
                        </Button>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div style={{ textAlign: "center", padding: "40px" }}>
                    <RocketIcon size={48} color="var(--text-secondary)" />
                    <div style={{ marginTop: "16px" }}>
                      <Text>진행 중인 프로젝트가 없습니다.</Text>
                      <br />
                      <Button
                        type="primary"
                        icon={<RocketIcon size={16} />}
                        onClick={() => navigate("/projects/new")}
                        style={{ marginTop: "12px" }}
                      >
                        새 프로젝트 시작하기
                      </Button>
                    </div>
                  </div>
                )}
              </Card>

              {/* 내 활동 피드 */}
              <Card
                title={
                  <Space>
                    <PathIcon size={20} color="var(--primary-color)" />
                    <span>📋 내 활동 피드</span>
                  </Space>
                }
                extra={
                  <Button
                    type="link"
                    icon={<ArrowRightOutlined />}
                    onClick={() => navigate("/activity")}
                  >
                    전체보기
                  </Button>
                }
              >
                <Timeline>
                  {activityFeed.map((item) => (
                    <Timeline.Item
                      key={item.id}
                      dot={getActivityIcon(item.type)}
                    >
                      <div>
                        <Text strong>{item.title}</Text>
                        {item.amount && (
                          <Tag color="green" style={{ marginLeft: "8px" }}>
                            {item.amount > 0 ? "+" : ""}
                            {formatCurrency(item.amount)} USDC
                          </Tag>
                        )}
                        <br />
                        <Text type="secondary">{item.description}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: "12px" }}>
                          <ClockCircleOutlined /> {item.timestamp}
                        </Text>
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </Card>
            </Col>

            {/* 우측 컬럼 */}
            <Col span={8}>
              {/* 포트폴리오 요약 */}
              <Card
                style={{ marginBottom: "24px" }}
                title={
                  <Space>
                    <InvestmentIcon size={20} color="var(--primary-color)" />
                    <span>💼 포트폴리오 요약</span>
                  </Space>
                }
              >
                {portfolioStats && (
                  <div>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Statistic
                          title="총 투자액"
                          value={portfolioStats.totalInvestment}
                          suffix="USDC"
                          valueStyle={{ fontSize: "16px" }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="현재 가치"
                          value={portfolioStats.currentValue}
                          suffix="USDC"
                          valueStyle={{ fontSize: "16px", color: "#52c41a" }}
                        />
                      </Col>
                    </Row>
                    <Divider />
                    <div style={{ textAlign: "center" }}>
                      <Statistic
                        title="총 수익"
                        value={portfolioStats.profitLoss}
                        suffix="USDC"
                        prefix={portfolioStats.profitLoss > 0 ? "+" : ""}
                        valueStyle={{
                          color:
                            portfolioStats.profitLoss > 0
                              ? "#52c41a"
                              : "#ff4d4f",
                          fontSize: "20px",
                          fontWeight: "bold",
                        }}
                      />
                      <div style={{ marginTop: "8px" }}>
                        <Tag
                          color={
                            portfolioStats.profitLossPercent > 0
                              ? "green"
                              : "red"
                          }
                          style={{ fontSize: "14px" }}
                        >
                          <RiseOutlined />
                          {portfolioStats.profitLossPercent > 0 ? "+" : ""}
                          {portfolioStats.profitLossPercent}%
                        </Tag>
                      </div>
                    </div>
                    <Divider />
                    <Row gutter={16}>
                      <Col span={12}>
                        <Text type="secondary">총 프로젝트</Text>
                        <div style={{ fontSize: "18px", fontWeight: "bold" }}>
                          {portfolioStats.totalProjects}개
                        </div>
                      </Col>
                      <Col span={12}>
                        <Text type="secondary">활성 프로젝트</Text>
                        <div
                          style={{
                            fontSize: "18px",
                            fontWeight: "bold",
                            color: "var(--primary-color)",
                          }}
                        >
                          {portfolioStats.activeProjects}개
                        </div>
                      </Col>
                    </Row>
                  </div>
                )}
              </Card>

              {/* 주목할 만한 프로젝트 */}
              <Card
                title={
                  <Space>
                    <FireOutlined style={{ color: "#ff4d4f" }} />
                    <span>🔥 주목할 만한 프로젝트</span>
                  </Space>
                }
                extra={
                  <Button
                    type="link"
                    icon={<ArrowRightOutlined />}
                    onClick={() => navigate("/explore")}
                  >
                    더보기
                  </Button>
                }
              >
                <List
                  dataSource={recommendedProjects}
                  renderItem={(project) => (
                    <List.Item style={{ padding: "12px 0" }}>
                      <List.Item.Meta
                        title={
                          <div
                            onClick={() => navigate(`/project/${project.id}`)}
                            style={{ cursor: "pointer" }}
                          >
                            <Space>
                              {project.title}
                              {project.isHot && <Tag color="red">HOT</Tag>}
                            </Space>
                          </div>
                        }
                        description={
                          <Space direction="vertical" size={4}>
                            <Text type="secondary" style={{ fontSize: "12px" }}>
                              {project.description.slice(0, 50)}...
                            </Text>
                            <div>
                              <Space>
                                <Text strong style={{ color: "#52c41a" }}>
                                  ${project.currentPrice}
                                </Text>
                                <Text type="secondary">
                                  TVL: {formatCurrency(project.tvl)}
                                </Text>
                              </Space>
                            </div>
                            <div>
                              <Tag>{project.category}</Tag>
                              <Tag color="blue">D-{project.daysLeft}</Tag>
                            </div>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            </Col>
          </Row>
        </div>
      </div>
    </div>
  );
};

export default NewDashboardPage;
