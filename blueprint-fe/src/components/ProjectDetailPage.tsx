import {
  ArrowLeftOutlined,
  BookOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  FallOutlined,
  HistoryOutlined,
  LineChartOutlined,
  LockOutlined,
  RiseOutlined,
  TeamOutlined,
  TrophyOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Button,
  Card,
  Col,
  Input,
  Layout,
  List,
  Progress,
  Row,
  Space,
  Spin,
  Statistic,
  Tabs,
  Tag,
  Typography,
  message,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { Milestone, Project } from "../types";
import ThemeToggle from "./ThemeToggle";

const { Content } = Layout;
const { Title, Text } = Typography;
const { TabPane } = Tabs;

// Mock data for development (API 연동 전)
const mockMarketData = {
  yesPrice: 0.72,
  noPrice: 0.28,
  priceChange: +0.05,
  totalVolume: 45320,
  totalTVL: 125000,
};

const ProjectDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [project, setProject] = useState<Project | null>(null);
  const [selectedMilestone, setSelectedMilestone] = useState<Milestone | null>(
    null
  );
  const [activeTab, setActiveTab] = useState("proof");
  const [tradeAmount, setTradeAmount] = useState<number>(100);
  const [tradeType, setTradeType] = useState<"yes" | "no">("yes");

  const loadProject = async () => {
    if (!id) return;

    try {
      setLoading(true);
      const response = await apiClient.getProject(parseInt(id));

      if (response.success && response.data) {
        setProject(response.data);
        // 첫 번째 활성 마일스톤을 기본 선택
        if (response.data.milestones && response.data.milestones.length > 0) {
          const activeMilestone = response.data.milestones.find(
            (m) => m.status === "pending"
          );
          if (activeMilestone) {
            setSelectedMilestone(activeMilestone);
          } else {
            setSelectedMilestone(response.data.milestones[0]);
          }
        }
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

  // 마일스톤 상태별 아이콘 및 색상
  const getMilestoneIcon = (status?: string) => {
    switch (status) {
      case "completed":
        return <CheckCircleOutlined style={{ color: "#52c41a" }} />;
      case "pending":
        return <ClockCircleOutlined style={{ color: "#1890ff" }} />;
      default:
        return <LockOutlined style={{ color: "#d9d9d9" }} />;
    }
  };

  const getMilestoneStatus = (status?: string) => {
    switch (status) {
      case "completed":
        return { text: "완료", color: "#52c41a" };
      case "pending":
        return { text: "진행중", color: "#faad14" };
      default:
        return { text: "예정", color: "#d9d9d9" };
    }
  };

  // 프로젝트 소유자 여부 확인 (향후 사용 예정)
  // const isOwner = user && Number(user.id) === project.user_id;
  const totalMilestones = project.milestones?.length || 0;
  const completedMilestones =
    project.milestones?.filter((m) => m.status === "completed").length || 0;
  const progressPercent =
    totalMilestones > 0 ? (completedMilestones / totalMilestones) * 100 : 0;

  return (
    <Layout style={{ minHeight: "100vh", background: "var(--bg-primary)" }}>
      <Content style={{ background: "var(--bg-primary)" }}>
        {/* 상단 프로젝트 헤더 */}
        <div
          style={{
            background: "var(--bg-secondary)",
            borderBottom: "1px solid var(--border-color)",
            padding: "16px 24px",
          }}
        >
          <div style={{ maxWidth: "1400px", margin: "0 auto" }}>
            <Row justify="space-between" align="middle">
              <Col>
                <Button
                  type="link"
                  icon={<ArrowLeftOutlined />}
                  onClick={() => navigate("/")}
                  style={{ padding: 0, color: "var(--text-primary)" }}
                >
                  프로젝트 목록으로 돌아가기
                </Button>
              </Col>
              <Col>
                <ThemeToggle />
              </Col>
            </Row>

            <Row gutter={24} align="middle" style={{ marginTop: 16 }}>
              <Col flex="auto">
                <Space direction="vertical" size={8}>
                  <Title
                    level={3}
                    style={{ margin: 0, color: "var(--text-primary)" }}
                  >
                    {project.title}
                  </Title>
                  <Space>
                    <UserOutlined style={{ color: "var(--text-secondary)" }} />
                    <Text style={{ color: "var(--text-secondary)" }}>
                      프로젝트 생성자 #{project.user_id}
                    </Text>
                    <Tag color="blue">{project.category}</Tag>
                  </Space>
                </Space>
              </Col>
              <Col>
                <Space size="large">
                  <div>
                    <Text
                      style={{
                        color: "var(--text-secondary)",
                        fontSize: "12px",
                      }}
                    >
                      전체 진행률
                    </Text>
                    <Progress
                      percent={Math.round(progressPercent)}
                      size="small"
                      style={{ minWidth: 150 }}
                    />
                  </div>
                  <Statistic
                    title="총 TVL"
                    value={mockMarketData.totalTVL}
                    prefix={<DollarOutlined />}
                    suffix="USDC"
                    valueStyle={{
                      fontSize: "18px",
                      color: "var(--text-primary)",
                    }}
                  />
                </Space>
              </Col>
            </Row>
          </div>
        </div>

        {/* 메인 2단 레이아웃 */}
        <div style={{ maxWidth: "1400px", margin: "0 auto", padding: "24px" }}>
          <Row gutter={24} style={{ minHeight: "calc(100vh - 200px)" }}>
            {/* A. 좌측: 마일스톤 네비게이터 */}
            <Col span={8}>
              <Card
                title={
                  <Space>
                    <TrophyOutlined />
                    마일스톤 목록
                  </Space>
                }
                style={{ height: "100%" }}
                bodyStyle={{ padding: 0 }}
              >
                <div
                  style={{
                    maxHeight: "calc(100vh - 300px)",
                    overflowY: "auto",
                  }}
                >
                  {project.milestones?.map((milestone) => {
                    const status = getMilestoneStatus(milestone.status);
                    const isSelected = selectedMilestone?.id === milestone.id;

                    return (
                      <div
                        key={milestone.id}
                        onClick={() => setSelectedMilestone(milestone)}
                        style={{
                          padding: "16px 20px",
                          borderBottom: "1px solid var(--border-color)",
                          cursor: "pointer",
                          background: isSelected
                            ? "var(--primary-color-light)"
                            : "transparent",
                          transition: "all 0.2s ease",
                        }}
                        onMouseEnter={(e) => {
                          if (!isSelected) {
                            e.currentTarget.style.background =
                              "var(--bg-hover)";
                          }
                        }}
                        onMouseLeave={(e) => {
                          if (!isSelected) {
                            e.currentTarget.style.background = "transparent";
                          }
                        }}
                      >
                        <Row justify="space-between" align="top">
                          <Col span={18}>
                            <Space direction="vertical" size={4}>
                              <Space>
                                {getMilestoneIcon(milestone.status)}
                                <Text
                                  strong
                                  style={{
                                    color: isSelected
                                      ? "var(--primary-color)"
                                      : "var(--text-primary)",
                                    fontSize: "14px",
                                  }}
                                >
                                  {milestone.title}
                                </Text>
                              </Space>
                              <Text
                                type="secondary"
                                style={{ fontSize: "12px", display: "block" }}
                              >
                                {(
                                  milestone.description || "설명이 없습니다"
                                ).slice(0, 60)}
                                ...
                              </Text>
                              <Tag color={status.color}>{status.text}</Tag>
                            </Space>
                          </Col>
                          <Col span={6} style={{ textAlign: "right" }}>
                            {milestone.status === "pending" && (
                              <div>
                                <Text
                                  strong
                                  style={{ color: "#52c41a", fontSize: "12px" }}
                                >
                                  ${mockMarketData.yesPrice}
                                </Text>
                                <div>
                                  <Tag color="green">LIVE</Tag>
                                </div>
                              </div>
                            )}
                          </Col>
                        </Row>
                      </div>
                    );
                  })}
                </div>
              </Card>
            </Col>

            {/* B. 우측: 메인 컨텐츠 영역 */}
            <Col span={16}>
              {selectedMilestone ? (
                <Space direction="vertical" size={24} style={{ width: "100%" }}>
                  {/* B-1. 핵심 거래 인터페이스 */}
                  <Card>
                    <Space
                      direction="vertical"
                      size={16}
                      style={{ width: "100%" }}
                    >
                      {/* 마일스톤 정보 */}
                      <div>
                        <Title level={4} style={{ margin: 0 }}>
                          {selectedMilestone.title}
                        </Title>
                        <Text type="secondary">
                          {selectedMilestone.description}
                        </Text>
                        <div style={{ marginTop: 8 }}>
                          <Space>
                            <CalendarOutlined />
                            <Text>
                              목표 날짜: {selectedMilestone.target_date}
                            </Text>
                          </Space>
                        </div>
                      </div>

                      {/* 핵심 가격 디스플레이 */}
                      <Row gutter={16}>
                        <Col span={12}>
                          <Card
                            size="small"
                            style={{
                              background:
                                tradeType === "yes" ? "#f6ffed" : "#fafafa",
                              border:
                                tradeType === "yes"
                                  ? "2px solid #52c41a"
                                  : "1px solid #d9d9d9",
                              cursor: "pointer",
                            }}
                            onClick={() => setTradeType("yes")}
                          >
                            <Row justify="space-between" align="middle">
                              <Col>
                                <Space direction="vertical" size={4}>
                                  <Text strong style={{ fontSize: "16px" }}>
                                    성공 YES
                                  </Text>
                                  <Text
                                    style={{
                                      fontSize: "20px",
                                      fontWeight: "bold",
                                      color: "#52c41a",
                                    }}
                                  >
                                    ${mockMarketData.yesPrice}
                                  </Text>
                                </Space>
                              </Col>
                              <Col>
                                <RiseOutlined
                                  style={{ fontSize: "24px", color: "#52c41a" }}
                                />
                              </Col>
                            </Row>
                          </Card>
                        </Col>
                        <Col span={12}>
                          <Card
                            size="small"
                            style={{
                              background:
                                tradeType === "no" ? "#fff1f0" : "#fafafa",
                              border:
                                tradeType === "no"
                                  ? "2px solid #ff4d4f"
                                  : "1px solid #d9d9d9",
                              cursor: "pointer",
                            }}
                            onClick={() => setTradeType("no")}
                          >
                            <Row justify="space-between" align="middle">
                              <Col>
                                <Space direction="vertical" size={4}>
                                  <Text strong style={{ fontSize: "16px" }}>
                                    실패 NO
                                  </Text>
                                  <Text
                                    style={{
                                      fontSize: "20px",
                                      fontWeight: "bold",
                                      color: "#ff4d4f",
                                    }}
                                  >
                                    ${mockMarketData.noPrice}
                                  </Text>
                                </Space>
                              </Col>
                              <Col>
                                <FallOutlined
                                  style={{ fontSize: "24px", color: "#ff4d4f" }}
                                />
                              </Col>
                            </Row>
                          </Card>
                        </Col>
                      </Row>

                      {/* 거래창 */}
                      <Card title="거래하기" size="small">
                        <Row gutter={16}>
                          <Col span={8}>
                            <Text>투자 금액</Text>
                            <Input
                              type="number"
                              value={tradeAmount}
                              onChange={(e) =>
                                setTradeAmount(Number(e.target.value))
                              }
                              suffix="USDC"
                              style={{ marginTop: 4 }}
                            />
                          </Col>
                          <Col span={8}>
                            <Text>예상 수익</Text>
                            <div style={{ marginTop: 8 }}>
                              <Text
                                strong
                                style={{ fontSize: "16px", color: "#52c41a" }}
                              >
                                ${(tradeAmount * 0.4).toFixed(2)}
                              </Text>
                            </div>
                          </Col>
                          <Col span={8}>
                            <Text>수익률</Text>
                            <div style={{ marginTop: 8 }}>
                              <Text
                                strong
                                style={{ fontSize: "16px", color: "#52c41a" }}
                              >
                                +40%
                              </Text>
                            </div>
                          </Col>
                        </Row>
                        <Button
                          type="primary"
                          size="large"
                          style={{
                            width: "100%",
                            marginTop: 16,
                            background:
                              tradeType === "yes" ? "#52c41a" : "#ff4d4f",
                            borderColor:
                              tradeType === "yes" ? "#52c41a" : "#ff4d4f",
                          }}
                          disabled={!isAuthenticated}
                        >
                          {tradeType === "yes"
                            ? "성공에 베팅하기"
                            : "실패에 베팅하기"}
                        </Button>
                      </Card>
                    </Space>
                  </Card>

                  {/* B-2. 상세 데이터 */}
                  <Card title="시장 데이터">
                    <Row gutter={24}>
                      <Col span={12}>
                        <Card size="small" title="가격 차트">
                          <div
                            style={{
                              textAlign: "center",
                              padding: "40px",
                              color: "#999",
                            }}
                          >
                            <LineChartOutlined style={{ fontSize: "48px" }} />
                            <div>차트 데이터 준비 중...</div>
                          </div>
                        </Card>
                      </Col>
                      <Col span={12}>
                        <Card size="small" title="최근 거래">
                          <List
                            size="small"
                            dataSource={[
                              {
                                id: 1,
                                type: "YES",
                                price: 0.72,
                                amount: 100,
                                time: "2분 전",
                              },
                              {
                                id: 2,
                                type: "NO",
                                price: 0.28,
                                amount: 50,
                                time: "5분 전",
                              },
                              {
                                id: 3,
                                type: "YES",
                                price: 0.71,
                                amount: 200,
                                time: "8분 전",
                              },
                            ]}
                            renderItem={(item: {
                              id: number;
                              type: string;
                              price: number;
                              amount: number;
                              time: string;
                            }) => (
                              <List.Item>
                                <Space>
                                  <Tag
                                    color={
                                      item.type === "YES" ? "green" : "red"
                                    }
                                  >
                                    {item.type}
                                  </Tag>
                                  <Text>${item.price}</Text>
                                  <Text type="secondary">
                                    {item.amount} USDC
                                  </Text>
                                  <Text type="secondary">{item.time}</Text>
                                </Space>
                              </List.Item>
                            )}
                          />
                        </Card>
                      </Col>
                    </Row>
                  </Card>

                  {/* B-3. 관련 정보 탭 */}
                  <Card>
                    <Tabs activeKey={activeTab} onChange={setActiveTab}>
                      <TabPane
                        tab={
                          <span>
                            <BookOutlined />
                            증명
                          </span>
                        }
                        key="proof"
                      >
                        <div
                          style={{
                            textAlign: "center",
                            padding: "40px",
                            color: "#999",
                          }}
                        >
                          <BookOutlined style={{ fontSize: "48px" }} />
                          <div>아직 제출된 증명 자료가 없습니다.</div>
                        </div>
                      </TabPane>
                      <TabPane
                        tab={
                          <span>
                            <TeamOutlined />
                            멘토
                          </span>
                        }
                        key="mentors"
                      >
                        <List
                          header={<div>리드 멘토 (베팅액 순)</div>}
                          dataSource={[
                            {
                              id: 1,
                              name: "Mentor #123",
                              amount: 5000,
                              isLead: true,
                            },
                            {
                              id: 2,
                              name: "Mentor #456",
                              amount: 3200,
                              isLead: true,
                            },
                            {
                              id: 3,
                              name: "Mentor #789",
                              amount: 2800,
                              isLead: false,
                            },
                          ]}
                          renderItem={(item: {
                            id: number;
                            name: string;
                            amount: number;
                            isLead: boolean;
                          }) => (
                            <List.Item>
                              <List.Item.Meta
                                avatar={<Avatar icon={<UserOutlined />} />}
                                title={
                                  <Space>
                                    {item.name}
                                    {item.isLead && (
                                      <Tag color="gold">리드 멘토</Tag>
                                    )}
                                  </Space>
                                }
                                description={`베팅 금액: ${item.amount} USDC`}
                              />
                            </List.Item>
                          )}
                        />
                      </TabPane>
                      <TabPane
                        tab={
                          <span>
                            <HistoryOutlined />
                            활동
                          </span>
                        }
                        key="activity"
                      >
                        <div
                          style={{
                            textAlign: "center",
                            padding: "40px",
                            color: "#999",
                          }}
                        >
                          <HistoryOutlined style={{ fontSize: "48px" }} />
                          <div>활동 내역이 없습니다.</div>
                        </div>
                      </TabPane>
                    </Tabs>
                  </Card>
                </Space>
              ) : (
                <Card style={{ height: "100%" }}>
                  <div
                    style={{
                      textAlign: "center",
                      padding: "60px",
                      color: "#999",
                    }}
                  >
                    <TrophyOutlined style={{ fontSize: "64px" }} />
                    <Title level={4} type="secondary">
                      마일스톤을 선택해주세요
                    </Title>
                    <Text type="secondary">
                      좌측에서 마일스톤을 선택하면 상세 정보가 표시됩니다.
                    </Text>
                  </div>
                </Card>
              )}
            </Col>
          </Row>
        </div>
      </Content>
    </Layout>
  );
};

export default ProjectDetailPage;
