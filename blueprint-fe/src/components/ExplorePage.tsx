import {
  ClockCircleOutlined,
  SearchOutlined,
  SortAscendingOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Card,
  Col,
  Empty,
  Input,
  Pagination,
  Row,
  Select,
  Space,
  Tag,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import GlobalNavbar from "./GlobalNavbar";
import { CompassIcon, MilestoneIcon, PathIcon } from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

interface ExploreProject {
  id: number;
  title: string;
  description: string;
  creator: {
    username: string;
    avatar?: string;
  };
  currentMilestone: {
    title: string;
    daysLeft: number;
  };
  marketData: {
    successPrice: number;
    tvl: number;
    volume24h: number;
    investorCount: number;
  };
  category: string;
  status: "funding" | "active" | "completed";
  isHot: boolean;
  createdAt: string;
}

const ExplorePage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [projects, setProjects] = useState<ExploreProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 12;

  // 필터 및 정렬 상태
  const [searchTerm, setSearchTerm] = useState(
    searchParams.get("search") || ""
  );
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");
  const [sortBy, setSortBy] = useState<string>("latest");

  // 카테고리 및 정렬 옵션
  const categories = [
    { value: "all", label: "전체" },
    { value: "it", label: "IT/개발" },
    { value: "startup", label: "창업" },
    { value: "lifestyle", label: "라이프스타일" },
    { value: "education", label: "교육" },
    { value: "health", label: "건강" },
    { value: "finance", label: "금융" },
  ];

  const statusOptions = [
    { value: "all", label: "전체" },
    { value: "funding", label: "펀딩 중" },
    { value: "active", label: "진행 중" },
    { value: "completed", label: "완료됨" },
  ];

  const sortOptions = [
    { value: "latest", label: "최신순" },
    { value: "tvl", label: "총 투자액(TVL) 순" },
    { value: "deadline", label: "마감 임박 순" },
    { value: "volume", label: "인기순(24h 거래량)" },
  ];

  // Mock data
  useEffect(() => {
    const loadMockData = () => {
      const mockProjects: ExploreProject[] = [
        {
          id: 1,
          title: "AI 기반 피트니스 앱 개발",
          description:
            "개인 맞춤형 운동 프로그램을 제공하는 AI 피트니스 앱을 개발합니다. 사용자의 체력 수준과 목표에 맞는 운동을 추천하고, 실시간 자세 교정 기능을 제공합니다.",
          creator: {
            username: "김개발",
            avatar: undefined,
          },
          currentMilestone: {
            title: "MVP 개발 완료",
            daysLeft: 7,
          },
          marketData: {
            successPrice: 0.68,
            tvl: 45000,
            volume24h: 8500,
            investorCount: 89,
          },
          category: "it",
          status: "active",
          isHot: true,
          createdAt: "2024-01-15",
        },
        {
          id: 2,
          title: "친환경 배달 서비스 플랫폼",
          description:
            "전기차와 자전거를 활용한 친환경 배달 서비스 플랫폼을 구축합니다. 탄소발자국을 줄이면서도 효율적인 배달 시스템을 만들어갑니다.",
          creator: {
            username: "박창업",
          },
          currentMilestone: {
            title: "사업자 등록 및 초기 자금 확보",
            daysLeft: 14,
          },
          marketData: {
            successPrice: 0.72,
            tvl: 32000,
            volume24h: 6200,
            investorCount: 67,
          },
          category: "startup",
          status: "funding",
          isHot: true,
          createdAt: "2024-01-12",
        },
        {
          id: 3,
          title: "개인 브랜딩 마스터 과정",
          description:
            "6개월 만에 개인 브랜드를 구축하고 수익화하는 프로젝트입니다. SNS 전략부터 콘텐츠 제작, 수익 모델 개발까지 체계적으로 진행합니다.",
          creator: {
            username: "이브랜딩",
          },
          currentMilestone: {
            title: "SNS 팔로워 1만명 달성",
            daysLeft: 21,
          },
          marketData: {
            successPrice: 0.55,
            tvl: 28000,
            volume24h: 4100,
            investorCount: 52,
          },
          category: "lifestyle",
          status: "active",
          isHot: false,
          createdAt: "2024-01-10",
        },
        {
          id: 4,
          title: "블록체인 기반 투표 시스템",
          description:
            "투명하고 안전한 블록체인 기반 전자투표 시스템을 개발합니다. 선거의 공정성과 투명성을 보장하는 차세대 투표 플랫폼을 만들어갑니다.",
          creator: {
            username: "최블록",
          },
          currentMilestone: {
            title: "스마트 컨트랙트 개발",
            daysLeft: 18,
          },
          marketData: {
            successPrice: 0.43,
            tvl: 52000,
            volume24h: 7800,
            investorCount: 134,
          },
          category: "it",
          status: "active",
          isHot: false,
          createdAt: "2024-01-08",
        },
        {
          id: 5,
          title: "온라인 요리 클래스 플랫폼",
          description:
            "전문 셰프들과 함께하는 실시간 온라인 요리 클래스 플랫폼을 구축합니다. 재료 배송부터 실시간 레슨까지 원스톱 서비스를 제공합니다.",
          creator: {
            username: "정요리",
          },
          currentMilestone: {
            title: "플랫폼 베타 테스트",
            daysLeft: 12,
          },
          marketData: {
            successPrice: 0.76,
            tvl: 18500,
            volume24h: 2900,
            investorCount: 38,
          },
          category: "education",
          status: "active",
          isHot: false,
          createdAt: "2024-01-05",
        },
        {
          id: 6,
          title: "AI 주식 투자 봇 개발",
          description:
            "머신러닝을 활용한 자동 주식 투자 봇을 개발합니다. 시장 데이터 분석부터 자동 매매까지 AI가 담당하는 스마트 투자 시스템입니다.",
          creator: {
            username: "한AI",
          },
          currentMilestone: {
            title: "백테스팅 완료",
            daysLeft: 9,
          },
          marketData: {
            successPrice: 0.62,
            tvl: 73000,
            volume24h: 12400,
            investorCount: 156,
          },
          category: "finance",
          status: "active",
          isHot: true,
          createdAt: "2024-01-03",
        },
      ];

      // 필터링 및 정렬 로직
      let filteredProjects = mockProjects;

      // 검색어 필터
      if (searchTerm) {
        filteredProjects = filteredProjects.filter(
          (project) =>
            project.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
            project.description.toLowerCase().includes(searchTerm.toLowerCase())
        );
      }

      // 카테고리 필터
      if (selectedCategory !== "all") {
        filteredProjects = filteredProjects.filter(
          (project) => project.category === selectedCategory
        );
      }

      // 상태 필터
      if (selectedStatus !== "all") {
        filteredProjects = filteredProjects.filter(
          (project) => project.status === selectedStatus
        );
      }

      // 정렬
      switch (sortBy) {
        case "tvl":
          filteredProjects.sort((a, b) => b.marketData.tvl - a.marketData.tvl);
          break;
        case "deadline":
          filteredProjects.sort(
            (a, b) => a.currentMilestone.daysLeft - b.currentMilestone.daysLeft
          );
          break;
        case "volume":
          filteredProjects.sort(
            (a, b) => b.marketData.volume24h - a.marketData.volume24h
          );
          break;
        default: // latest
          filteredProjects.sort(
            (a, b) =>
              new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
          );
      }

      setProjects(filteredProjects);
      setTotal(filteredProjects.length);
      setLoading(false);
    };

    setLoading(true);
    setTimeout(loadMockData, 300);
  }, [searchTerm, selectedCategory, selectedStatus, sortBy]);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR").format(amount);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "funding":
        return "orange";
      case "active":
        return "green";
      case "completed":
        return "blue";
      default:
        return "default";
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "funding":
        return "펀딩 중";
      case "active":
        return "진행 중";
      case "completed":
        return "완료됨";
      default:
        return status;
    }
  };

  const handleSearch = (value: string) => {
    setSearchTerm(value);
    if (value) {
      setSearchParams({ search: value });
    } else {
      setSearchParams({});
    }
  };

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
              <CompassIcon size={32} color="var(--primary-color)" />
              <div>
                <Title
                  level={2}
                  style={{ margin: 0, color: "var(--text-primary)" }}
                >
                  프로젝트 탐색
                </Title>
                <Text type="secondary" style={{ fontSize: "16px" }}>
                  투자하고 멘토링할 프로젝트를 찾아보세요
                </Text>
              </div>
            </Space>
          </div>

          {/* 필터 및 검색 */}
          <Card style={{ marginBottom: "24px" }}>
            <Row gutter={[16, 16]} align="middle">
              <Col span={8}>
                <Input
                  placeholder="프로젝트 검색..."
                  prefix={<SearchOutlined />}
                  value={searchTerm}
                  onChange={(e) => handleSearch(e.target.value)}
                  allowClear
                  size="large"
                />
              </Col>
              <Col span={4}>
                <Select
                  value={selectedCategory}
                  onChange={setSelectedCategory}
                  style={{ width: "100%" }}
                  size="large"
                  placeholder="카테고리"
                >
                  {categories.map((category) => (
                    <Option key={category.value} value={category.value}>
                      {category.label}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col span={4}>
                <Select
                  value={selectedStatus}
                  onChange={setSelectedStatus}
                  style={{ width: "100%" }}
                  size="large"
                  placeholder="상태"
                >
                  {statusOptions.map((status) => (
                    <Option key={status.value} value={status.value}>
                      {status.label}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col span={4}>
                <Select
                  value={sortBy}
                  onChange={setSortBy}
                  style={{ width: "100%" }}
                  size="large"
                  placeholder="정렬"
                  suffixIcon={<SortAscendingOutlined />}
                >
                  {sortOptions.map((option) => (
                    <Option key={option.value} value={option.value}>
                      {option.label}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col span={4}>
                <Text type="secondary">
                  총 {formatCurrency(total)}개 프로젝트
                </Text>
              </Col>
            </Row>
          </Card>

          {/* 프로젝트 목록 */}
          {loading ? (
            <div style={{ textAlign: "center", padding: "60px" }}>
              <Text>프로젝트를 불러오는 중...</Text>
            </div>
          ) : projects.length === 0 ? (
            <Empty
              description="검색 조건에 맞는 프로젝트가 없습니다"
              style={{ padding: "60px" }}
            />
          ) : (
            <>
              <Row gutter={[24, 24]}>
                {projects
                  .slice((currentPage - 1) * pageSize, currentPage * pageSize)
                  .map((project) => (
                    <Col span={8} key={project.id}>
                      <Card
                        hoverable
                        onClick={() => navigate(`/project/${project.id}`)}
                        style={{ height: "100%" }}
                        cover={
                          <div
                            style={{
                              padding: "20px",
                              background: "var(--bg-secondary)",
                            }}
                          >
                            <Space>
                              <PathIcon
                                size={24}
                                color="var(--primary-color)"
                              />
                              <Text strong style={{ fontSize: "16px" }}>
                                {project.title}
                              </Text>
                              {project.isHot && <Tag color="red">HOT</Tag>}
                            </Space>
                          </div>
                        }
                      >
                        <div
                          style={{
                            height: "240px",
                            display: "flex",
                            flexDirection: "column",
                          }}
                        >
                          {/* 프로젝트 정보 */}
                          <div style={{ flex: 1 }}>
                            <Paragraph
                              ellipsis={{ rows: 3 }}
                              style={{
                                marginBottom: "12px",
                                minHeight: "66px",
                              }}
                            >
                              {project.description}
                            </Paragraph>

                            <Space
                              size="small"
                              style={{ marginBottom: "12px" }}
                            >
                              <UserOutlined />
                              <Text type="secondary">
                                @{project.creator.username}
                              </Text>
                            </Space>

                            <div style={{ marginBottom: "12px" }}>
                              <Space>
                                <Tag color={getStatusColor(project.status)}>
                                  {getStatusText(project.status)}
                                </Tag>
                                <Tag>
                                  {
                                    categories.find(
                                      (c) => c.value === project.category
                                    )?.label
                                  }
                                </Tag>
                              </Space>
                            </div>

                            {/* 현재 마일스톤 */}
                            <div style={{ marginBottom: "16px" }}>
                              <Space size={4}>
                                <MilestoneIcon
                                  size={14}
                                  color="var(--primary-color)"
                                />
                                <Text style={{ fontSize: "12px" }}>
                                  {project.currentMilestone.title}
                                </Text>
                              </Space>
                              <div style={{ marginTop: "4px" }}>
                                <Tag color="blue">
                                  <ClockCircleOutlined /> D-
                                  {project.currentMilestone.daysLeft}
                                </Tag>
                              </div>
                            </div>
                          </div>

                          {/* 시장 데이터 */}
                          <div
                            style={{
                              borderTop: "1px solid var(--border-color)",
                              paddingTop: "12px",
                            }}
                          >
                            <Row gutter={16}>
                              <Col span={12}>
                                <div style={{ textAlign: "center" }}>
                                  <Text
                                    strong
                                    style={{
                                      color: "#52c41a",
                                      fontSize: "18px",
                                    }}
                                  >
                                    ${project.marketData.successPrice}
                                  </Text>
                                  <div
                                    style={{
                                      fontSize: "11px",
                                      color: "var(--text-secondary)",
                                    }}
                                  >
                                    성공 가격
                                  </div>
                                </div>
                              </Col>
                              <Col span={12}>
                                <div style={{ textAlign: "center" }}>
                                  <Text strong style={{ fontSize: "14px" }}>
                                    {formatCurrency(project.marketData.tvl)}
                                  </Text>
                                  <div
                                    style={{
                                      fontSize: "11px",
                                      color: "var(--text-secondary)",
                                    }}
                                  >
                                    TVL (USDC)
                                  </div>
                                </div>
                              </Col>
                            </Row>
                            <Row gutter={16} style={{ marginTop: "8px" }}>
                              <Col span={12}>
                                <div style={{ textAlign: "center" }}>
                                  <Text style={{ fontSize: "12px" }}>
                                    {formatCurrency(
                                      project.marketData.volume24h
                                    )}
                                  </Text>
                                  <div
                                    style={{
                                      fontSize: "10px",
                                      color: "var(--text-secondary)",
                                    }}
                                  >
                                    24h 거래량
                                  </div>
                                </div>
                              </Col>
                              <Col span={12}>
                                <div style={{ textAlign: "center" }}>
                                  <Text style={{ fontSize: "12px" }}>
                                    {project.marketData.investorCount}명
                                  </Text>
                                  <div
                                    style={{
                                      fontSize: "10px",
                                      color: "var(--text-secondary)",
                                    }}
                                  >
                                    투자자 수
                                  </div>
                                </div>
                              </Col>
                            </Row>
                          </div>
                        </div>
                      </Card>
                    </Col>
                  ))}
              </Row>

              {/* 페이지네이션 */}
              <div style={{ textAlign: "center", marginTop: "48px" }}>
                <Pagination
                  current={currentPage}
                  total={total}
                  pageSize={pageSize}
                  onChange={setCurrentPage}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(total, range) =>
                    `${range[0]}-${range[1]} / 총 ${total}개`
                  }
                />
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default ExplorePage;
