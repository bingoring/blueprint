import {
  ArrowLeftOutlined,
  BookOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  FileAddOutlined,
  HistoryOutlined,
  LineChartOutlined,
  LinkOutlined,
  LockOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  PlusOutlined,
  TagOutlined,
  TeamOutlined,
  TrophyOutlined,
  UploadOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Form,
  Input,
  List,
  Modal,
  Progress,
  Row,
  Select,
  Spin,
  Statistic,
  Tabs,
  Tag,
  Tooltip,
  Typography,
  Upload,
  message,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { Milestone, Project } from "../types";
import GlobalNavbar from "./GlobalNavbar";
import { MilestoneIcon, PathIcon } from "./icons/BlueprintIcons";

// const { Content } = Layout;
const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;
const { Option } = Select;

// Mock data for development
const mockMarketData = {
  yesPrice: 0.72,
  noPrice: 0.28,
  priceChange: +0.05,
  totalVolume: 45320,
  totalTVL: 125000,
};

// 게시글 유형 정의
const postTypes = [
  { value: "progress", label: "진행 상황 보고", color: "blue" },
  { value: "evidence", label: "데이터/증거 자료", color: "green" },
  { value: "announcement", label: "중요 공지", color: "orange" },
  { value: "completion", label: "최종 증명 제출", color: "red" },
];

// Mock 진행 기록 데이터
const mockProgressLogs = [
  {
    id: 1,
    type: "completion",
    title: "MVP 개발 완료 - 최종 증명 제출",
    content:
      "3개월간 개발한 MVP가 완성되었습니다. 모든 핵심 기능이 구현되었으며, 베타 테스터 50명을 통한 사용성 테스트도 완료했습니다.",
    attachments: [
      {
        type: "github",
        url: "https://github.com/user/project",
        title: "GitHub 저장소",
      },
      { type: "demo", url: "https://demo.example.com", title: "데모 사이트" },
    ],
    timestamp: "2시간 전",
    likes: 24,
    comments: 8,
  },
  {
    id: 2,
    type: "evidence",
    title: "베타 테스트 결과 보고",
    content:
      "50명의 베타 테스터를 대상으로 한 사용성 테스트 결과입니다. 평균 만족도 4.2/5점, 주요 피드백을 반영하여 UI를 개선했습니다.",
    attachments: [
      {
        type: "file",
        url: "/files/beta-test-report.pdf",
        title: "베타 테스트 보고서.pdf",
      },
    ],
    timestamp: "1일 전",
    likes: 15,
    comments: 3,
  },
  {
    id: 3,
    type: "progress",
    title: "주간 진행 상황 업데이트",
    content:
      "이번 주는 사용자 인터페이스 최적화에 집중했습니다. 로딩 시간을 30% 단축하고, 모바일 반응형 디자인을 완성했습니다.",
    attachments: [],
    timestamp: "3일 전",
    likes: 8,
    comments: 2,
  },
];

const ProjectDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isAuthenticated, user } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [project, setProject] = useState<Project | null>(null);
  const [selectedMilestone, setSelectedMilestone] = useState<Milestone | null>(
    null
  );
  const [activeTab, setActiveTab] = useState("trade");
  const [tradeAmount, setTradeAmount] = useState<number>(100);
  const [tradeType, setTradeType] = useState<"yes" | "no">("yes");

  // 레이아웃 상태
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  // 진행 기록 관련 상태
  const [progressLogs, setProgressLogs] = useState(mockProgressLogs);
  const [showPostModal, setShowPostModal] = useState(false);
  const [postForm] = Form.useForm();

  // ⚖️ Blueprint Court 분쟁 관련 상태
  const [disputeDetail, setDisputeDetail] = useState<any>(null);
  const [disputeTimer, setDisputeTimer] = useState<any>(null);
  const [showDisputeModal, setShowDisputeModal] = useState(false);
  const [disputeLoading, setDisputeLoading] = useState(false);

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

  // ⚖️ Blueprint Court 관련 함수들
  const loadDisputeDetail = async (disputeId: number) => {
    try {
      setDisputeLoading(true);
      const response = await apiClient.getDisputeDetail(disputeId);
      if (response.success && response.data) {
        setDisputeDetail(response.data);
        // 분쟁 타이머도 함께 로드
        loadDisputeTimer(disputeId);
      }
    } catch (error) {
      console.error("분쟁 상세 정보 로드 실패:", error);
      message.error("분쟁 정보를 불러올 수 없습니다");
    } finally {
      setDisputeLoading(false);
    }
  };

  const loadDisputeTimer = async (disputeId: number) => {
    try {
      const response = await apiClient.getDisputeTimer(disputeId);
      if (response.success && response.data) {
        setDisputeTimer(response.data.time_remaining);
      }
    } catch (error) {
      console.error("분쟁 타이머 로드 실패:", error);
    }
  };

  const handleReportResult = async (
    result: boolean,
    evidence?: string,
    description?: string
  ) => {
    if (!selectedMilestone) return;

    try {
      setDisputeLoading(true);
      const response = await apiClient.reportMilestoneResult(
        selectedMilestone.id,
        {
          result,
          evidence_url: evidence || "",
          evidence_files: [],
          description: description || "",
        }
      );

      if (response.success) {
        message.success(
          "마일스톤 결과가 성공적으로 보고되었습니다. 48시간 동안 이의제기 기간입니다."
        );
        // 프로젝트 데이터 다시 로드
        await loadProject();
      } else {
        message.error(response.error || "결과 보고에 실패했습니다");
      }
    } catch (error) {
      console.error("결과 보고 실패:", error);
      message.error("결과 보고 중 오류가 발생했습니다");
    } finally {
      setDisputeLoading(false);
    }
  };

  const handleCreateDispute = async (request: CreateDisputeRequest) => {
    try {
      setDisputeLoading(true);
      const response = await apiClient.createDispute(request);

      if (response.success) {
        message.success(
          "이의 제기가 성공적으로 접수되었습니다. 투표 기간이 시작됩니다."
        );
        setShowDisputeModal(false);
        // 프로젝트 데이터 다시 로드
        await loadProject();
      } else {
        message.error(response.error || "이의 제기에 실패했습니다");
      }
    } catch (error) {
      console.error("이의 제기 실패:", error);
      message.error("이의 제기 중 오류가 발생했습니다");
    } finally {
      setDisputeLoading(false);
    }
  };

  // 게시글 등록 처리
  const handlePostSubmit = async (values: any) => {
    try {
      const newPost = {
        id: Date.now(),
        type: values.type,
        title: values.title,
        content: values.content,
        attachments: values.attachments || [],
        timestamp: "방금 전",
        likes: 0,
        comments: 0,
      };

      setProgressLogs([newPost, ...progressLogs]);
      setShowPostModal(false);
      postForm.resetFields();
      message.success("진행 상황이 성공적으로 등록되었습니다!");

      // 투자자들에게 알림 발송 (실제 구현에서는 API 호출)
      console.log("알림 발송: 새로운 업데이트가 등록되었습니다.");
    } catch (error) {
      message.error("등록 중 오류가 발생했습니다.");
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen" style={{ background: "var(--bg-primary)" }}>
        <GlobalNavbar />
        <div className="flex items-center justify-center h-screen">
          <Spin size="large" />
          <span className="ml-3" style={{ color: "var(--text-secondary)" }}>
            프로젝트 정보를 로딩 중...
          </span>
        </div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen" style={{ background: "var(--bg-primary)" }}>
        <GlobalNavbar />
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <Text style={{ color: "var(--text-secondary)" }}>
              프로젝트를 찾을 수 없습니다.
            </Text>
            <br />
            <Button
              type="primary"
              onClick={() => navigate("/")}
              className="mt-4"
            >
              홈으로 돌아가기
            </Button>
          </div>
        </div>
      </div>
    );
  }

  // 마일스톤 상태별 아이콘 및 색상
  const getMilestoneIcon = (status?: string) => {
    switch (status) {
      case "completed":
        return <CheckCircleOutlined className="text-green-500" />;
      case "pending":
        return <ClockCircleOutlined className="text-blue-500" />;
      default:
        return <LockOutlined className="text-gray-400" />;
    }
  };

  const getMilestoneStatus = (status?: string) => {
    switch (status) {
      case "completed":
        return { text: "완료", color: "green" };
      case "pending":
        return { text: "진행중", color: "blue" };
      default:
        return { text: "예정", color: "default" };
    }
  };

  // 게시글 유형별 아이콘
  const getPostTypeIcon = (type: string) => {
    switch (type) {
      case "progress":
        return <ClockCircleOutlined className="text-blue-500" />;
      case "evidence":
        return <FileAddOutlined className="text-green-500" />;
      case "announcement":
        return <TagOutlined className="text-orange-500" />;
      case "completion":
        return <TrophyOutlined className="text-red-500" />;
      default:
        return <BookOutlined />;
    }
  };

  // 첨부파일 아이콘
  const getAttachmentIcon = (type: string) => {
    switch (type) {
      case "github":
        return "🔗";
      case "demo":
        return "🌐";
      case "file":
        return "📎";
      default:
        return "📎";
    }
  };

  const totalMilestones = project.milestones?.length || 0;
  const completedMilestones =
    project.milestones?.filter((m) => m.status === "completed").length || 0;
  const progressPercent =
    totalMilestones > 0 ? (completedMilestones / totalMilestones) * 100 : 0;

  // 프로젝트 소유자 여부 확인
  const isOwner = user && Number(user.id) === project.user_id;

  return (
    <div className="min-h-screen" style={{ background: "var(--bg-primary)" }}>
      <GlobalNavbar />

      <div className="pt-16">
        {/* 프로젝트 헤더 */}
        <div
          style={{
            background: "var(--bg-secondary)",
            borderBottom: "1px solid var(--border-color)",
          }}
          className="px-6 py-4"
        >
          <div className="max-w-7xl mx-auto">
            <div className="flex items-center justify-between mb-4">
              <Button
                type="text"
                icon={<ArrowLeftOutlined />}
                onClick={() => navigate("/explore")}
                style={{ color: "var(--text-secondary)" }}
              >
                프로젝트 목록으로 돌아가기
              </Button>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-4 mb-2">
                  <PathIcon size={24} color="#3b82f6" />
                  <Title level={2} className="!mb-0">
                    {project.title}
                  </Title>
                  <Tag color="blue">{project.category}</Tag>
                </div>
                <div
                  className="flex items-center gap-4"
                  style={{ color: "var(--text-secondary)" }}
                >
                  <div className="flex items-center gap-2">
                    <UserOutlined />
                    <Text style={{ color: "var(--text-secondary)" }}>
                      프로젝트 생성자 #{project.user_id}
                    </Text>
                  </div>
                  <div className="flex items-center gap-2">
                    <Text style={{ color: "var(--text-secondary)" }}>
                      전체 진행률:
                    </Text>
                    <Progress
                      percent={Math.round(progressPercent)}
                      size="small"
                      className="w-32"
                    />
                  </div>
                </div>
              </div>

              <div className="text-right">
                <Statistic
                  title="총 TVL"
                  value={mockMarketData.totalTVL}
                  prefix={<DollarOutlined />}
                  suffix="USDC"
                  className="text-right"
                />
              </div>
            </div>
          </div>
        </div>

        {/* 메인 컨텐츠 */}
        <div className="max-w-7xl mx-auto p-6">
          <div className="flex gap-6">
            {/* 좌측: 마일스톤 네비게이터 (접을 수 있게) */}
            <div
              className={`transition-all duration-300 ${
                sidebarCollapsed ? "w-14" : "w-80"
              } flex-shrink-0 overflow-hidden`}
            >
              <Card
                className="h-fit"
                style={{
                  background: "var(--bg-card)",
                  border: "1px solid var(--border-color)",
                }}
              >
                <div className="flex items-center justify-between mb-4">
                  {!sidebarCollapsed && (
                    <div className="flex items-center gap-2">
                      <MilestoneIcon size={20} color="var(--primary-color)" />
                      <Text strong style={{ color: "var(--text-primary)" }}>
                        마일스톤 목록
                      </Text>
                    </div>
                  )}
                  <Button
                    type="text"
                    icon={
                      sidebarCollapsed ? (
                        <MenuUnfoldOutlined />
                      ) : (
                        <MenuFoldOutlined />
                      )
                    }
                    onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
                    size="small"
                  />
                </div>

                {sidebarCollapsed ? (
                  // 접힌 상태: 세로 점 네비게이션
                  <div className="flex flex-col items-center justify-center py-4">
                    {project.milestones?.map((milestone, index) => {
                      const isSelected = selectedMilestone?.id === milestone.id;
                      const status = getMilestoneStatus(milestone.status);

                      return (
                        <Tooltip
                          key={milestone.id}
                          title={`${index + 1}. ${milestone.title} - ${
                            status.text
                          }`}
                          placement="right"
                        >
                          <div
                            onClick={() => setSelectedMilestone(milestone)}
                            className="relative cursor-pointer mb-3 last:mb-0 transition-all duration-200 hover:scale-110"
                          >
                            {/* 연결선 (마지막 제외) */}
                            {index < (project.milestones?.length || 0) - 1 && (
                              <div
                                className="absolute left-1/2 top-6 w-0.5 h-6 -translate-x-1/2"
                                style={{
                                  background:
                                    milestone.status === "completed"
                                      ? "var(--color-success)"
                                      : "var(--border-color)",
                                }}
                              ></div>
                            )}

                            {/* 마일스톤 점 */}
                            <div className="relative">
                              {isSelected ? (
                                // 선택된 마일스톤: 큰 채워진 원
                                <div
                                  className="w-6 h-6 rounded-full border-2 flex items-center justify-center shadow-lg"
                                  style={{
                                    background: "var(--primary-color)",
                                    borderColor: "var(--primary-color)",
                                    boxShadow:
                                      "0 0 0 3px rgba(24, 144, 255, 0.2)",
                                  }}
                                >
                                  <div className="w-2 h-2 bg-white rounded-full"></div>
                                </div>
                              ) : milestone.status === "completed" ? (
                                // 완료된 마일스톤: 체크마크
                                <div
                                  className="w-5 h-5 rounded-full flex items-center justify-center"
                                  style={{ background: "var(--color-success)" }}
                                >
                                  <CheckCircleOutlined className="text-white text-xs" />
                                </div>
                              ) : milestone.status === "pending" ? (
                                // 진행 중 마일스톤: 펄싱 효과
                                <div className="relative">
                                  <div
                                    className="w-4 h-4 rounded-full border-2"
                                    style={{
                                      borderColor: "var(--color-warning)",
                                      background: "var(--bg-card)",
                                    }}
                                  ></div>
                                  <div
                                    className="absolute inset-0 w-4 h-4 rounded-full animate-ping"
                                    style={{
                                      background: "var(--color-warning)",
                                      opacity: 0.4,
                                    }}
                                  ></div>
                                </div>
                              ) : (
                                // 대기 중 마일스톤: 빈 원
                                <div
                                  className="w-4 h-4 rounded-full border-2"
                                  style={{
                                    borderColor: "var(--border-color)",
                                    background: "var(--bg-card)",
                                  }}
                                ></div>
                              )}
                            </div>
                          </div>
                        </Tooltip>
                      );
                    })}
                  </div>
                ) : (
                  // 펼쳐진 상태: 기존 목록 형태
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {project.milestones?.map((milestone) => {
                      const status = getMilestoneStatus(milestone.status);
                      const isSelected = selectedMilestone?.id === milestone.id;

                      return (
                        <div
                          key={milestone.id}
                          onClick={() => setSelectedMilestone(milestone)}
                          className="p-3 rounded-lg cursor-pointer transition-all duration-200 border"
                          style={{
                            background: isSelected
                              ? "var(--bg-hover)"
                              : "transparent",
                            borderColor: isSelected
                              ? "var(--primary-color)"
                              : "transparent",
                          }}
                        >
                          <div style={{ minWidth: 0 }}>
                            <div className="flex items-center gap-2 mb-2">
                              {getMilestoneIcon(milestone.status)}
                              <Text
                                strong
                                className="text-sm flex-1 truncate"
                                style={{ color: "var(--text-primary)" }}
                              >
                                {milestone.title}
                              </Text>
                              <Tag color={status.color}>{status.text}</Tag>
                            </div>
                            <Text
                              className="text-xs line-clamp-2"
                              style={{ color: "var(--text-tertiary)" }}
                            >
                              {milestone.description || "설명이 없습니다"}
                            </Text>
                            {milestone.status === "pending" && (
                              <div className="flex items-center justify-between mt-2">
                                <Text
                                  className="text-xs font-bold"
                                  style={{ color: "var(--color-success)" }}
                                >
                                  ${mockMarketData.yesPrice}
                                </Text>
                                <Badge status="processing" text="LIVE" />
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </Card>
            </div>

            {/* 우측: 메인 컨텐츠 영역 (탭 구조) */}
            <div className="flex-1">
              {selectedMilestone ? (
                <div>
                  {/* 선택된 마일스톤 정보 */}
                  <Card
                    className="mb-6"
                    style={{
                      background: "var(--bg-card)",
                      border: "1px solid var(--border-color)",
                    }}
                  >
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <Title level={3} className="!mb-2">
                          {selectedMilestone.title}
                        </Title>
                        <Text style={{ color: "var(--text-secondary)" }}>
                          {selectedMilestone.description}
                        </Text>
                        {selectedMilestone.target_date && (
                          <div className="flex items-center gap-2 mt-2">
                            <CalendarOutlined
                              style={{ color: "var(--text-tertiary)" }}
                            />
                            <Text style={{ color: "var(--text-tertiary)" }}>
                              목표 날짜: {selectedMilestone.target_date}
                            </Text>
                          </div>
                        )}
                      </div>
                      {isOwner && (
                        <Button
                          type="primary"
                          icon={<PlusOutlined />}
                          onClick={() => setShowPostModal(true)}
                        >
                          진행 상황 등록
                        </Button>
                      )}
                    </div>
                  </Card>

                  {/* 탭 컨텐츠 */}
                  <Card
                    style={{
                      background: "var(--bg-card)",
                      border: "1px solid var(--border-color)",
                    }}
                  >
                    <Tabs
                      activeKey={activeTab}
                      onChange={setActiveTab}
                      size="large"
                    >
                      {/* 거래 탭 */}
                      <Tabs.TabPane
                        tab={
                          <span className="flex items-center gap-2">
                            <DollarOutlined />
                            거래 (Trade)
                          </span>
                        }
                        key="trade"
                      >
                        <Row gutter={24}>
                          {/* 좌측: 차트 영역 (메인) */}
                          <Col span={16}>
                            <div className="space-y-4">
                              {/* 가격 차트 */}
                              <Card
                                title="가격 차트"
                                style={{
                                  background: "var(--bg-card)",
                                  border: "1px solid var(--border-color)",
                                }}
                              >
                                <div
                                  className="h-80 flex items-center justify-center"
                                  style={{ color: "var(--text-tertiary)" }}
                                >
                                  <div className="text-center">
                                    <LineChartOutlined className="text-6xl mb-4" />
                                    <div>실시간 가격 차트</div>
                                    <Text
                                      style={{ color: "var(--text-tertiary)" }}
                                      className="text-sm"
                                    >
                                      차트 데이터를 로딩 중입니다...
                                    </Text>
                                  </div>
                                </div>
                              </Card>

                              {/* 호가창 (Order Book) */}
                              <Card
                                title="호가창 (Order Book)"
                                style={{
                                  background: "var(--bg-card)",
                                  border: "1px solid var(--border-color)",
                                }}
                              >
                                <div className="h-96">
                                  {/* 매도 호가 (Ask Orders) */}
                                  <div className="mb-4">
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                      className="text-sm font-medium mb-2 block"
                                    >
                                      매도 (Ask) - NO
                                    </Text>
                                    <div className="space-y-1">
                                      {[
                                        {
                                          price: 0.32,
                                          amount: 450,
                                          total: 144,
                                        },
                                        {
                                          price: 0.31,
                                          amount: 820,
                                          total: 254.2,
                                        },
                                        {
                                          price: 0.3,
                                          amount: 1200,
                                          total: 360,
                                        },
                                        {
                                          price: 0.29,
                                          amount: 950,
                                          total: 275.5,
                                        },
                                        {
                                          price: 0.28,
                                          amount: 1500,
                                          total: 420,
                                        },
                                      ].map((order, idx) => (
                                        <div
                                          key={idx}
                                          className="flex items-center justify-between py-1 px-2 rounded hover:bg-red-50 dark:hover:bg-red-900/10 cursor-pointer"
                                          style={{
                                            background: `linear-gradient(to left, rgba(220, 38, 38, 0.1) ${
                                              (order.amount / 1500) * 100
                                            }%, transparent 0%)`,
                                          }}
                                        >
                                          <div className="flex items-center gap-4 flex-1">
                                            <Text
                                              strong
                                              style={{
                                                color: "var(--color-error)",
                                              }}
                                              className="min-w-16"
                                            >
                                              ${order.price.toFixed(2)}
                                            </Text>
                                            <Text
                                              style={{
                                                color: "var(--text-primary)",
                                              }}
                                              className="min-w-20"
                                            >
                                              {order.amount.toLocaleString()}
                                            </Text>
                                            <Text
                                              style={{
                                                color: "var(--text-secondary)",
                                              }}
                                              className="text-sm"
                                            >
                                              ${order.total.toLocaleString()}
                                            </Text>
                                          </div>
                                        </div>
                                      ))}
                                    </div>
                                  </div>

                                  {/* 현재 가격 */}
                                  <div className="py-3 px-2 text-center border-y border-dashed border-gray-300 dark:border-gray-600 mb-4">
                                    <div className="flex items-center justify-center gap-2">
                                      <Text
                                        strong
                                        className="text-lg"
                                        style={{
                                          color: "var(--color-success)",
                                        }}
                                      >
                                        $0.72
                                      </Text>
                                      <Text
                                        style={{
                                          color: "var(--color-success)",
                                        }}
                                        className="text-sm"
                                      >
                                        ↑ +$0.05
                                      </Text>
                                    </div>
                                    <Text
                                      style={{ color: "var(--text-tertiary)" }}
                                      className="text-xs"
                                    >
                                      현재 시장가 (YES)
                                    </Text>
                                  </div>

                                  {/* 매수 호가 (Bid Orders) */}
                                  <div>
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                      className="text-sm font-medium mb-2 block"
                                    >
                                      매수 (Bid) - YES
                                    </Text>
                                    <div className="space-y-1">
                                      {[
                                        {
                                          price: 0.71,
                                          amount: 2100,
                                          total: 1491,
                                        },
                                        {
                                          price: 0.7,
                                          amount: 1800,
                                          total: 1260,
                                        },
                                        {
                                          price: 0.69,
                                          amount: 1350,
                                          total: 931.5,
                                        },
                                        {
                                          price: 0.68,
                                          amount: 950,
                                          total: 646,
                                        },
                                        {
                                          price: 0.67,
                                          amount: 750,
                                          total: 502.5,
                                        },
                                      ].map((order, idx2) => (
                                        <div
                                          key={idx2}
                                          className="flex items-center justify-between py-1 px-2 rounded hover:bg-green-50 dark:hover:bg-green-900/10 cursor-pointer"
                                          style={{
                                            background: `linear-gradient(to left, rgba(34, 197, 94, 0.1) ${
                                              (order.amount / 2100) * 100
                                            }%, transparent 0%)`,
                                          }}
                                        >
                                          <div className="flex items-center gap-4 flex-1">
                                            <Text
                                              strong
                                              style={{
                                                color: "var(--color-success)",
                                              }}
                                              className="min-w-16"
                                            >
                                              ${order.price.toFixed(2)}
                                            </Text>
                                            <Text
                                              style={{
                                                color: "var(--text-primary)",
                                              }}
                                              className="min-w-20"
                                            >
                                              {order.amount.toLocaleString()}
                                            </Text>
                                            <Text
                                              style={{
                                                color: "var(--text-secondary)",
                                              }}
                                              className="text-sm"
                                            >
                                              ${order.total.toLocaleString()}
                                            </Text>
                                          </div>
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                </div>
                              </Card>

                              {/* 최근 거래 내역 */}
                              <Card
                                title="최근 거래"
                                size="small"
                                style={{
                                  background: "var(--bg-card)",
                                  border: "1px solid var(--border-color)",
                                }}
                              >
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
                                    {
                                      id: 4,
                                      type: "YES",
                                      price: 0.73,
                                      amount: 150,
                                      time: "12분 전",
                                    },
                                  ]}
                                  renderItem={(item) => (
                                    <List.Item className="px-0 py-2">
                                      <div className="flex items-center justify-between w-full">
                                        <div className="flex items-center gap-3">
                                          <Tag
                                            color={
                                              item.type === "YES"
                                                ? "green"
                                                : "red"
                                            }
                                            className="min-w-12 text-center"
                                          >
                                            {item.type}
                                          </Tag>
                                          <Text strong>${item.price}</Text>
                                        </div>
                                        <div
                                          className="flex items-center gap-3"
                                          style={{
                                            color: "var(--text-secondary)",
                                          }}
                                        >
                                          <Text
                                            style={{
                                              color: "var(--text-secondary)",
                                            }}
                                          >
                                            {item.amount} USDC
                                          </Text>
                                          <Text
                                            style={{
                                              color: "var(--text-tertiary)",
                                            }}
                                            className="text-xs"
                                          >
                                            {item.time}
                                          </Text>
                                        </div>
                                      </div>
                                    </List.Item>
                                  )}
                                />
                              </Card>
                            </div>
                          </Col>

                          {/* 우측: 거래 패널 (사이드바) */}
                          <Col span={8}>
                            <div className="space-y-4">
                              {/* YES/NO 버튼들 - 폴리마켓 스타일 */}
                              <div className="space-y-2">
                                <Button
                                  className={`w-full h-10 font-medium transition-all ${
                                    tradeType === "yes"
                                      ? "bg-green-500 hover:bg-green-600 text-white border-green-500"
                                      : "bg-green-50 hover:bg-green-100 text-green-600 border-green-200"
                                  }`}
                                  onClick={() => setTradeType("yes")}
                                >
                                  <div className="flex items-center justify-between w-full">
                                    <span>성공 YES</span>
                                    <span className="font-bold">
                                      ${mockMarketData.yesPrice}
                                    </span>
                                  </div>
                                </Button>
                                <Button
                                  className={`w-full h-10 font-medium transition-all ${
                                    tradeType === "no"
                                      ? "bg-red-500 hover:bg-red-600 text-white border-red-500"
                                      : "bg-red-50 hover:bg-red-100 text-red-600 border-red-200"
                                  }`}
                                  onClick={() => setTradeType("no")}
                                >
                                  <div className="flex items-center justify-between w-full">
                                    <span>실패 NO</span>
                                    <span className="font-bold">
                                      ${mockMarketData.noPrice}
                                    </span>
                                  </div>
                                </Button>
                              </div>

                              {/* 거래 입력 폼 */}
                              <Card
                                size="small"
                                style={{
                                  background: "var(--bg-card)",
                                  border: "1px solid var(--border-color)",
                                }}
                              >
                                <div className="space-y-3">
                                  <div>
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                      className="text-sm block mb-1"
                                    >
                                      투자 금액
                                    </Text>
                                    <Input
                                      type="number"
                                      value={tradeAmount}
                                      onChange={(e) =>
                                        setTradeAmount(Number(e.target.value))
                                      }
                                      suffix="USDC"
                                      className="text-right"
                                    />
                                  </div>

                                  <div className="flex justify-between text-sm">
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                    >
                                      예상 수익:
                                    </Text>
                                    <Text
                                      strong
                                      style={{ color: "var(--color-success)" }}
                                    >
                                      ${(tradeAmount * 0.4).toFixed(2)}
                                    </Text>
                                  </div>

                                  <div className="flex justify-between text-sm">
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                    >
                                      수익률:
                                    </Text>
                                    <Text
                                      strong
                                      style={{ color: "var(--color-success)" }}
                                    >
                                      +40%
                                    </Text>
                                  </div>

                                  <Button
                                    type="primary"
                                    className={`w-full h-9 ${
                                      tradeType === "yes"
                                        ? "bg-green-500 hover:bg-green-600 border-green-500"
                                        : "bg-red-500 hover:bg-red-600 border-red-500"
                                    }`}
                                    disabled={!isAuthenticated}
                                  >
                                    {tradeType === "yes" ? "Buy YES" : "Buy NO"}
                                  </Button>
                                </div>
                              </Card>

                              {/* 시장 정보 */}
                              <Card
                                title="시장 정보"
                                size="small"
                                style={{
                                  background: "var(--bg-card)",
                                  border: "1px solid var(--border-color)",
                                }}
                              >
                                <div className="space-y-2">
                                  <div className="flex justify-between text-sm">
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                    >
                                      총 거래량:
                                    </Text>
                                    <Text strong>
                                      $
                                      {mockMarketData.totalVolume.toLocaleString()}
                                    </Text>
                                  </div>
                                  <div className="flex justify-between text-sm">
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                    >
                                      TVL:
                                    </Text>
                                    <Text strong>
                                      $
                                      {mockMarketData.totalTVL.toLocaleString()}
                                    </Text>
                                  </div>
                                  <div className="flex justify-between text-sm">
                                    <Text
                                      style={{ color: "var(--text-secondary)" }}
                                    >
                                      24h 변화:
                                    </Text>
                                    <Text
                                      strong
                                      style={{ color: "var(--color-success)" }}
                                    >
                                      +
                                      {(
                                        mockMarketData.priceChange * 100
                                      ).toFixed(1)}
                                      %
                                    </Text>
                                  </div>
                                </div>
                              </Card>
                            </div>
                          </Col>
                        </Row>
                      </Tabs.TabPane>

                      {/* 진행 기록 탭 */}
                      <Tabs.TabPane
                        tab={
                          <span className="flex items-center gap-2">
                            <HistoryOutlined />
                            진행 기록 (Log)
                          </span>
                        }
                        key="log"
                      >
                        <div className="space-y-4">
                          {progressLogs.length === 0 ? (
                            <div className="text-center py-12">
                              <HistoryOutlined
                                className="text-5xl mb-4"
                                style={{ color: "var(--text-tertiary)" }}
                              />
                              <div style={{ color: "var(--text-tertiary)" }}>
                                아직 등록된 진행 기록이 없습니다.
                              </div>
                            </div>
                          ) : (
                            progressLogs.map((log) => {
                              const postType = postTypes.find(
                                (t) => t.value === log.type
                              );
                              return (
                                <Card
                                  key={log.id}
                                  style={{
                                    background:
                                      log.type === "completion"
                                        ? "var(--bg-hover)"
                                        : "var(--bg-card)",
                                    border:
                                      log.type === "completion"
                                        ? "2px solid var(--color-error)"
                                        : "1px solid var(--border-color)",
                                  }}
                                >
                                  <div className="flex items-start gap-3">
                                    <div className="flex-shrink-0 mt-1">
                                      {getPostTypeIcon(log.type)}
                                    </div>
                                    <div className="flex-1">
                                      <div className="flex items-center gap-2 mb-2">
                                        <Tag color={postType?.color}>
                                          {postType?.label}
                                        </Tag>
                                        <Text
                                          className="text-sm"
                                          style={{
                                            color: "var(--text-tertiary)",
                                          }}
                                        >
                                          {log.timestamp}
                                        </Text>
                                      </div>
                                      <Title level={5} className="!mb-2">
                                        {log.title}
                                      </Title>
                                      <Paragraph
                                        className="mb-3"
                                        style={{
                                          color: "var(--text-secondary)",
                                        }}
                                      >
                                        {log.content}
                                      </Paragraph>

                                      {/* 첨부파일 */}
                                      {log.attachments.length > 0 && (
                                        <div className="mb-3">
                                          <Text
                                            strong
                                            className="text-sm mb-2 block"
                                            style={{
                                              color: "var(--text-secondary)",
                                            }}
                                          >
                                            첨부 자료:
                                          </Text>
                                          <div className="space-y-1">
                                            {log.attachments.map(
                                              (attachment, index) => (
                                                <div
                                                  key={index}
                                                  className="flex items-center gap-2"
                                                >
                                                  <span>
                                                    {getAttachmentIcon(
                                                      attachment.type
                                                    )}
                                                  </span>
                                                  <a
                                                    href={attachment.url}
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    style={{
                                                      color:
                                                        "var(--primary-color)",
                                                    }}
                                                  >
                                                    {attachment.title}
                                                  </a>
                                                </div>
                                              )
                                            )}
                                          </div>
                                        </div>
                                      )}

                                      {/* 반응 */}
                                      <div
                                        className="flex items-center gap-4 text-sm"
                                        style={{
                                          color: "var(--text-tertiary)",
                                        }}
                                      >
                                        <Button
                                          type="text"
                                          size="small"
                                          className="p-0"
                                        >
                                          👍 {log.likes}
                                        </Button>
                                        <Button
                                          type="text"
                                          size="small"
                                          className="p-0"
                                        >
                                          💬 {log.comments}
                                        </Button>
                                      </div>
                                    </div>
                                  </div>
                                </Card>
                              );
                            })
                          )}
                        </div>
                      </Tabs.TabPane>

                      {/* 멘토 탭 */}
                      <Tabs.TabPane
                        tab={
                          <span className="flex items-center gap-2">
                            <TeamOutlined />
                            멘토
                          </span>
                        }
                        key="mentors"
                      >
                        <div>
                          <Title level={5} className="mb-4">
                            리드 멘토 (베팅액 순)
                          </Title>
                          <List
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
                            renderItem={(item) => (
                              <List.Item>
                                <List.Item.Meta
                                  avatar={<Avatar icon={<UserOutlined />} />}
                                  title={
                                    <div className="flex items-center gap-2">
                                      {item.name}
                                      {item.isLead && (
                                        <Tag color="gold">리드 멘토</Tag>
                                      )}
                                    </div>
                                  }
                                  description={`베팅 금액: ${item.amount.toLocaleString()} USDC`}
                                />
                                <Button size="small">멘토링 요청</Button>
                              </List.Item>
                            )}
                          />
                        </div>
                      </Tabs.TabPane>

                      {/* ⚖️ 분쟁 탭 */}
                      <Tabs.TabPane
                        tab={
                          <span className="flex items-center gap-2">
                            <span>⚖️</span>
                            분쟁 (Dispute)
                          </span>
                        }
                        key="dispute"
                      >
                        <div className="space-y-4">
                          {/* 분쟁 타이머 (분쟁 진행 중일 때만 표시) */}
                          {disputeTimer && (
                            <DisputeTimer
                              timeRemaining={disputeTimer}
                              phase="challenge_window"
                            />
                          )}

                          {/* 마일스톤 상태에 따른 액션 버튼들 */}
                          <Card
                            title="분쟁 해결 시스템"
                            style={{
                              background: "var(--bg-card)",
                              border: "1px solid var(--border-color)",
                            }}
                          >
                            {selectedMilestone && (
                              <div className="space-y-4">
                                {/* 결과 미보고 상태 - 생성자만 결과 보고 가능 */}
                                {!selectedMilestone.result_reported &&
                                  user?.id === project?.creator_id && (
                                    <div className="space-y-3">
                                      <Text className="block">
                                        마일스톤이 완료되면 결과를 보고해주세요.
                                      </Text>
                                      <div className="flex gap-2">
                                        <Button
                                          type="primary"
                                          className="bg-green-500 hover:bg-green-600 border-green-500"
                                          onClick={() =>
                                            handleReportResult(
                                              true,
                                              "",
                                              "마일스톤이 성공적으로 완료되었습니다."
                                            )
                                          }
                                          loading={disputeLoading}
                                        >
                                          성공 보고
                                        </Button>
                                        <Button
                                          danger
                                          onClick={() =>
                                            handleReportResult(
                                              false,
                                              "",
                                              "마일스톤 완료에 실패했습니다."
                                            )
                                          }
                                          loading={disputeLoading}
                                        >
                                          실패 보고
                                        </Button>
                                      </div>
                                    </div>
                                  )}

                                {/* 결과 보고됨, 이의제기 기간 중 - 투자자들이 이의제기 가능 */}
                                {selectedMilestone.result_reported &&
                                  !selectedMilestone.is_in_dispute && (
                                    <div className="space-y-3">
                                      <Text className="block">
                                        결과가 보고되었습니다. 48시간 동안
                                        이의제기가 가능합니다.
                                      </Text>
                                      <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded">
                                        <Text>
                                          보고된 결과:{" "}
                                          {selectedMilestone.completed
                                            ? "✅ 성공"
                                            : "❌ 실패"}
                                        </Text>
                                      </div>
                                      <Button
                                        danger
                                        onClick={() =>
                                          setShowDisputeModal(true)
                                        }
                                        disabled={
                                          user?.id === project?.creator_id
                                        } // 생성자는 이의제기 불가
                                      >
                                        결과에 이의 제기
                                      </Button>
                                    </div>
                                  )}

                                {/* 분쟁 진행 중 */}
                                {selectedMilestone.is_in_dispute && (
                                  <div className="space-y-3">
                                    <Text className="block text-orange-500">
                                      ⚖️ 분쟁이 진행 중입니다. 판결단의 투표를
                                      기다리고 있습니다.
                                    </Text>
                                    {disputeDetail && (
                                      <div className="p-3 bg-orange-50 dark:bg-orange-900 rounded">
                                        <Text className="block font-semibold">
                                          분쟁 사유:
                                        </Text>
                                        <Text>
                                          {disputeDetail.dispute_reason}
                                        </Text>
                                      </div>
                                    )}
                                  </div>
                                )}

                                {/* 분쟁 없음 */}
                                {!selectedMilestone.result_reported &&
                                  !selectedMilestone.is_in_dispute && (
                                    <div className="text-center py-8">
                                      <Text
                                        style={{
                                          color: "var(--text-secondary)",
                                        }}
                                      >
                                        마일스톤 완료 후 결과 보고 시 분쟁 해결
                                        시스템이 작동됩니다.
                                      </Text>
                                    </div>
                                  )}
                              </div>
                            )}
                          </Card>
                        </div>
                      </Tabs.TabPane>
                    </Tabs>
                  </Card>
                </div>
              ) : (
                <Card
                  className="h-96"
                  style={{
                    background: "var(--bg-card)",
                    border: "1px solid var(--border-color)",
                  }}
                >
                  <div className="text-center py-20">
                    <MilestoneIcon size={64} />
                    <Title
                      level={4}
                      className="mt-4"
                      style={{ color: "var(--text-secondary)" }}
                    >
                      마일스톤을 선택해주세요
                    </Title>
                    <Text style={{ color: "var(--text-tertiary)" }}>
                      좌측에서 마일스톤을 선택하면 상세 정보가 표시됩니다.
                    </Text>
                  </div>
                </Card>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 진행 상황 등록 모달 */}
      <Modal
        title="진행 상황 등록"
        open={showPostModal}
        onCancel={() => setShowPostModal(false)}
        footer={null}
        width={800}
      >
        <Form
          form={postForm}
          onFinish={handlePostSubmit}
          layout="vertical"
          className="mt-4"
        >
          <Form.Item
            name="milestone"
            label="마일스톤"
            rules={[{ required: true, message: "마일스톤을 선택해주세요" }]}
          >
            <Select
              placeholder="마일스톤을 선택하세요"
              disabled
              value={selectedMilestone?.title}
            >
              <Option value={selectedMilestone?.id}>
                {selectedMilestone?.title}
              </Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="type"
            label="게시글 유형"
            rules={[{ required: true, message: "게시글 유형을 선택해주세요" }]}
          >
            <Select placeholder="게시글 유형을 선택하세요">
              {postTypes.map((type) => (
                <Option key={type.value} value={type.value}>
                  <Tag color={type.color}>{type.label}</Tag>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="title"
            label="제목"
            rules={[{ required: true, message: "제목을 입력해주세요" }]}
          >
            <Input placeholder="게시글 제목을 입력하세요" />
          </Form.Item>

          <Form.Item
            name="content"
            label="내용"
            rules={[{ required: true, message: "내용을 입력해주세요" }]}
          >
            <TextArea
              rows={6}
              placeholder="진행 상황에 대해 상세히 설명해주세요"
            />
          </Form.Item>

          <Form.Item name="attachments" label="증거 자료 첨부">
            <Upload multiple beforeUpload={() => false} listType="text">
              <Button icon={<UploadOutlined />}>파일 첨부</Button>
            </Upload>
            <div className="mt-2">
              <Input
                placeholder="외부 링크 (GitHub, Figma, YouTube 등)"
                prefix={<LinkOutlined />}
              />
            </div>
          </Form.Item>

          <Form.Item className="mb-0">
            <div className="flex justify-end gap-2">
              <Button onClick={() => setShowPostModal(false)}>취소</Button>
              <Button type="primary" htmlType="submit">
                등록하기
              </Button>
            </div>
          </Form.Item>
        </Form>
      </Modal>

      {/* ⚖️ 분쟁 제기 모달 */}
      <CreateDisputeModal
        visible={showDisputeModal}
        onClose={() => setShowDisputeModal(false)}
        onSubmit={handleCreateDispute}
        milestoneId={selectedMilestone?.id || 0}
        milestoneTitle={selectedMilestone?.title || ""}
        originalResult={selectedMilestone?.completed || false}
      />
    </div>
  );
};

export default ProjectDetailPage;
