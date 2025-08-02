import {
  ArrowLeftOutlined,
  CalendarOutlined,
  DollarOutlined,
  EditOutlined,
  LoginOutlined,
  PlusOutlined,
  ProjectOutlined,
  TeamOutlined,
  TrophyOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Empty,
  Layout,
  List,
  Progress,
  Row,
  Space,
  Spin,
  Statistic,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from "antd";
import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type {
  ActivityRecord,
  InvestmentTableRecord,
  Milestone,
  ProjectTableRecord,
} from "../types";

const { Content } = Layout;
const { Title } = Typography;

const NewDashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("projects");

  // 실제 데이터 state
  const [projectTableData, setProjectTableData] = useState<
    ProjectTableRecord[]
  >([]);
  const [statistics, setStatistics] = useState({
    totalProjects: 0,
    totalReceived: 0,
    totalInvestments: 0,
    avgProgress: 0,
  });

  // 임시 투자 데이터 (아직 백엔드 API 없음)
  const mockInvestments: InvestmentTableRecord[] = [
    {
      id: 1,
      projectId: 3,
      projectTitle: "요가 강사 자격증 취득",
      developer: "박요가",
      amount: 50000,
      investedAt: "2024-01-20",
      status: "active",
      progress: 45,
    },
    {
      id: 2,
      projectId: 4,
      projectTitle: "웹툰 작가 데뷔 프로젝트",
      developer: "김웹툰",
      amount: 100000,
      investedAt: "2024-01-10",
      status: "active",
      progress: 30,
    },
  ];

  useEffect(() => {
    if (!isAuthenticated) {
      message.error("로그인이 필요합니다");
      navigate("/");
      return;
    }
    loadUserData();
  }, [isAuthenticated, navigate]);

  const loadUserData = async () => {
    try {
      setLoading(true);
      console.log("🔄 사용자 프로젝트 데이터 로딩 중...");

      // 실제 API 호출: 사용자 프로젝트 목록
      const response = await apiClient.getProjects({
        page: 1,
        limit: 50,
        sort: "created_at",
        order: "desc",
      });

      if (response.success && response.data) {
        const projects = response.data.projects || [];
        // Project 데이터를 ProjectTableRecord 형태로 변환
        const tableData: ProjectTableRecord[] = projects.map((project) => ({
          id: project.id!,
          title: project.title,
          category: project.category,
          status: project.status || "draft",
          progress: calculateProjectProgress(project.milestones || []),
          totalInvestment: 0, // TODO: 투자 데이터 API 연결 후 계산
          investors: 0, // TODO: 투자자 수 API 연결 후 계산
          milestones: project.milestones?.length || 0,
          currentMilestone: getCurrentMilestoneIndex(project.milestones || []),
          createdAt: project.created_at?.split("T")[0] || "",
          targetDate: project.target_date?.split("T")[0] || "",
        }));

        setProjectTableData(tableData);

        // 통계 계산
        const stats = {
          totalProjects: projects.length,
          totalReceived: tableData.reduce(
            (sum, proj) => sum + proj.totalInvestment,
            0
          ),
          totalInvestments: mockInvestments.reduce(
            (sum, inv) => sum + inv.amount,
            0
          ),
          avgProgress:
            tableData.length > 0
              ? tableData.reduce((sum, proj) => sum + proj.progress, 0) /
                tableData.length
              : 0,
        };
        setStatistics(stats);

        console.log("✅ 프로젝트 데이터 로딩 완료:", projects.length, "개");
      } else {
        throw new Error(response.error || "프로젝트 조회 실패");
      }
    } catch (error) {
      console.error("❌ 프로젝트 데이터 로딩 실패:", error);
      message.error("프로젝트 데이터를 불러오는데 실패했습니다");

      // 에러 시 빈 데이터로 초기화
      setProjectTableData([]);
    } finally {
      setLoading(false);
    }
  };

  // 프로젝트 진행률 계산 (완료된 마일스톤 비율)
  const calculateProjectProgress = (milestones: Milestone[]): number => {
    if (!milestones || milestones.length === 0) return 0;

    const completedCount = milestones.filter(
      (milestone) => milestone.status === "completed"
    ).length;

    return Math.round((completedCount / milestones.length) * 100);
  };

  // 현재 진행 중인 마일스톤 인덱스 계산
  const getCurrentMilestoneIndex = (milestones: Milestone[]): number => {
    if (!milestones || milestones.length === 0) return 0;

    const completedCount = milestones.filter(
      (milestone) => milestone.status === "completed"
    ).length;

    return Math.min(completedCount + 1, milestones.length);
  };

  // 프로젝트 테이블 컬럼
  const projectColumns = [
    {
      title: t("project.projectTitle"),
      dataIndex: "title",
      key: "title",
      render: (title: string, record: ProjectTableRecord) => (
        <div>
          <Button
            type="link"
            className="font-medium p-0 h-auto text-left"
            onClick={() => navigate(`/project/${record.id}`)}
          >
            {title}
          </Button>
          <div className="text-sm text-gray-500">
            {t(`categories.${record.category}`)}
          </div>
        </div>
      ),
    },
    {
      title: t("project.status"),
      dataIndex: "status",
      key: "status",
      render: (status: string) => (
        <Badge
          status={status === "completed" ? "success" : "processing"}
          text={t(`status.${status}`)}
        />
      ),
    },
    {
      title: t("project.constructionProgress"),
      dataIndex: "progress",
      key: "progress",
      render: (progress: number) => (
        <Progress
          percent={progress}
          size="small"
          status={progress === 100 ? "success" : "active"}
        />
      ),
    },
    {
      title: t("milestone.milestones"),
      key: "milestones",
      render: (_: unknown, record: ProjectTableRecord) => (
        <span>
          {record.currentMilestone}/{record.milestones}
        </span>
      ),
    },
    {
      title: t("investment.totalInvestment"),
      dataIndex: "totalInvestment",
      key: "totalInvestment",
      render: (amount: number) => `₩${amount.toLocaleString()}`,
    },
    {
      title: t("investment.investors"),
      dataIndex: "investors",
      key: "investors",
      render: (count: number) => (
        <Space>
          <TeamOutlined />
          {count}명
        </Space>
      ),
    },
    {
      title: "작업",
      key: "actions",
      render: (_: unknown, record: ProjectTableRecord) => (
        <Space size="small">
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => navigate(`/edit-project/${record.id}`)}
            disabled={record.investors > 0}
            title={
              record.investors > 0
                ? "투자자가 있는 프로젝트는 수정할 수 없습니다"
                : "프로젝트 수정"
            }
          >
            {record.investors > 0 ? "🔒" : "수정"}
          </Button>
        </Space>
      ),
    },
  ];

  // 투자 테이블 컬럼
  const investmentColumns = [
    {
      title: t("project.projectTitle"),
      dataIndex: "projectTitle",
      key: "projectTitle",
      render: (title: string, record: InvestmentTableRecord) => (
        <div>
          <div className="font-medium">{title}</div>
          <div className="text-sm text-gray-500">{record.developer}</div>
        </div>
      ),
    },
    {
      title: t("investment.investmentAmount"),
      dataIndex: "amount",
      key: "amount",
      render: (amount: number) => `₩${amount.toLocaleString()}`,
    },
    {
      title: t("project.constructionProgress"),
      dataIndex: "progress",
      key: "progress",
      render: (progress: number) => (
        <Progress percent={progress} size="small" />
      ),
    },
    {
      title: "투자일",
      dataIndex: "investedAt",
      key: "investedAt",
    },
    {
      title: t("project.status"),
      dataIndex: "status",
      key: "status",
      render: (status: string) => (
        <Tag color={status === "active" ? "green" : "orange"}>
          {status === "active" ? "진행중" : "완료"}
        </Tag>
      ),
    },
  ];

  return (
    <Layout style={{ minHeight: "100vh", background: "#f5f5f5" }}>
      <Content style={{ padding: "24px" }}>
        <div style={{ maxWidth: 1200, margin: "0 auto" }}>
          {/* 헤더 */}
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 24,
            }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
              <Button
                icon={<ArrowLeftOutlined />}
                onClick={() => navigate("/")}
                type="text"
              >
                홈으로
              </Button>
              <Title level={3} className="!mb-0">
                {t("nav.dashboard")}
              </Title>
            </div>

            <Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => navigate("/create-project")}
              >
                {t("project.newProject")}
              </Button>
              <Button
                icon={<LoginOutlined />}
                onClick={() => useAuthStore.getState().logout()}
              >
                로그아웃
              </Button>
            </Space>
          </div>

          {loading ? (
            <div style={{ textAlign: "center", padding: "50px" }}>
              <Spin size="large" />
              <div style={{ marginTop: 16 }}>데이터를 불러오는 중...</div>
            </div>
          ) : (
            <>
              {/* 통계 카드 */}
              <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title={t("project.myProjects")}
                      value={statistics.totalProjects}
                      prefix={<ProjectOutlined />}
                      valueStyle={{ color: "#1890ff" }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="받은 총 투자금"
                      value={statistics.totalReceived}
                      prefix="₩"
                      precision={0}
                      valueStyle={{ color: "#52c41a" }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title={t("investment.myInvestments")}
                      value={statistics.totalInvestments}
                      prefix="₩"
                      precision={0}
                      valueStyle={{ color: "#faad14" }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="평균 진행률"
                      value={statistics.avgProgress}
                      suffix="%"
                      precision={1}
                      prefix={<TrophyOutlined />}
                      valueStyle={{ color: "#722ed1" }}
                    />
                  </Card>
                </Col>
              </Row>

              {/* 탭 컨텐츠 */}
              <Card>
                <Tabs
                  activeKey={activeTab}
                  onChange={setActiveTab}
                  items={[
                    {
                      key: "projects",
                      label: (
                        <span>
                          <ProjectOutlined />
                          {t("project.myProjects")}
                        </span>
                      ),
                      children: (
                        <Table
                          columns={projectColumns}
                          dataSource={projectTableData}
                          rowKey="id"
                          pagination={false}
                          locale={{
                            emptyText: (
                              <Empty
                                description="아직 등록된 프로젝트가 없습니다"
                                image={Empty.PRESENTED_IMAGE_SIMPLE}
                              >
                                <Button
                                  type="primary"
                                  icon={<PlusOutlined />}
                                  onClick={() => navigate("/create-project")}
                                >
                                  첫 프로젝트 만들기
                                </Button>
                              </Empty>
                            ),
                          }}
                        />
                      ),
                    },
                    {
                      key: "investments",
                      label: (
                        <span>
                          <DollarOutlined />
                          {t("investment.myInvestments")}
                        </span>
                      ),
                      children: (
                        <Table
                          columns={investmentColumns}
                          dataSource={mockInvestments}
                          rowKey="id"
                          pagination={false}
                          locale={{
                            emptyText: (
                              <Empty
                                description="아직 투자한 프로젝트가 없습니다"
                                image={Empty.PRESENTED_IMAGE_SIMPLE}
                              >
                                <Button type="primary">
                                  프로젝트 둘러보기
                                </Button>
                              </Empty>
                            ),
                          }}
                        />
                      ),
                    },
                    {
                      key: "activity",
                      label: (
                        <span>
                          <CalendarOutlined />
                          최근 활동
                        </span>
                      ),
                      children: (
                        <List
                          itemLayout="horizontal"
                          dataSource={
                            [
                              {
                                id: 1,
                                type: "investment" as const,
                                title: "새로운 투자를 받았습니다",
                                description:
                                  "김투자님이 카페 창업 프로젝트에 50,000원을 투자했습니다",
                                time: "2시간 전",
                              },
                              {
                                id: 2,
                                type: "milestone" as const,
                                title: "단계가 완료되었습니다",
                                description:
                                  "AI 개발자 프로젝트의 3단계가 완료되었습니다",
                                time: "1일 전",
                              },
                              {
                                id: 3,
                                type: "project" as const,
                                title: "새 프로젝트를 등록했습니다",
                                description:
                                  "카페 창업 프로젝트를 등록했습니다",
                                time: "3일 전",
                              },
                            ] as ActivityRecord[]
                          }
                          renderItem={(item: ActivityRecord) => (
                            <List.Item>
                              <List.Item.Meta
                                avatar={
                                  <Avatar
                                    icon={
                                      item.type === "investment" ? (
                                        <DollarOutlined />
                                      ) : item.type === "milestone" ? (
                                        <CalendarOutlined />
                                      ) : (
                                        <ProjectOutlined />
                                      )
                                    }
                                  />
                                }
                                title={item.title}
                                description={
                                  <div>
                                    <div>{item.description}</div>
                                    <div className="text-sm text-gray-400 mt-1">
                                      {item.time}
                                    </div>
                                  </div>
                                }
                              />
                            </List.Item>
                          )}
                          locale={{
                            emptyText: (
                              <Empty
                                description="최근 활동이 없습니다"
                                image={Empty.PRESENTED_IMAGE_SIMPLE}
                              />
                            ),
                          }}
                        />
                      ),
                    },
                  ]}
                />
              </Card>
            </>
          )}
        </div>
      </Content>
    </Layout>
  );
};

export default NewDashboard;
