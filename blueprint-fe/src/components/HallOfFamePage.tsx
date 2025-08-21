import {
  CrownOutlined,
  DollarOutlined,
  LineChartOutlined,
  RiseOutlined,
  StarOutlined,
  TeamOutlined,
  TrophyOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Card,
  Col,
  Divider,
  List,
  Progress,
  Rate,
  Row,
  Select,
  Space,
  Statistic,
  Tabs,
  Tag,
  Typography,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import GlobalNavbar from "./GlobalNavbar";
import { TrophyIcon } from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

interface SuccessfulProject {
  id: number;
  title: string;
  description: string;
  category: string;
  creator: {
    username: string;
    avatar?: string;
  };
  completedAt: string;
  duration: number; // days
  stats: {
    totalInvestment: number;
    finalROI: number;
    investorCount: number;
    mentorCount: number;
    averageRating: number;
    totalMilestones: number;
  };
  achievements: string[];
  story: string;
  images: string[];
  featured: boolean;
}

interface LeaderboardEntry {
  rank: number;
  projectId: number;
  projectTitle: string;
  creator: string;
  value: number;
  category: string;
  completedAt: string;
}

interface HallOfFameStats {
  totalSuccessfulProjects: number;
  totalValueCreated: number;
  averageROI: number;
  totalParticipants: number;
  categoriesRepresented: number;
  averageCompletionTime: number;
}

interface HallOfFameData {
  stats: HallOfFameStats;
  featuredProjects: SuccessfulProject[];
  projects: SuccessfulProject[];
  leaderboards: {
    highestROI: LeaderboardEntry[];
    mostInvestors: LeaderboardEntry[];
    highestRating: LeaderboardEntry[];
    largestInvestment: LeaderboardEntry[];
  };
}

const HallOfFamePage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [hallOfFameData, setHallOfFameData] = useState<HallOfFameData | null>(
    null
  );
  const [activeTab, setActiveTab] = useState("gallery");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [selectedYear, setSelectedYear] = useState<string>("all");

  // Mock data for development
  useEffect(() => {
    const loadMockData = () => {
      const mockData: HallOfFameData = {
        stats: {
          totalSuccessfulProjects: 127,
          totalValueCreated: 2850000,
          averageROI: 156,
          totalParticipants: 5430,
          categoriesRepresented: 8,
          averageCompletionTime: 142, // days
        },
        featuredProjects: [
          {
            id: 1,
            title: "AI 기반 의료진단 시스템",
            description:
              "딥러닝을 활용한 의료 영상 진단 AI 시스템 개발 및 상용화",
            category: "IT/개발",
            creator: {
              username: "medical_ai_pioneer",
              avatar: undefined,
            },
            completedAt: "2024-02-15",
            duration: 186,
            stats: {
              totalInvestment: 85000,
              finalROI: 340,
              investorCount: 156,
              mentorCount: 8,
              averageRating: 4.9,
              totalMilestones: 7,
            },
            achievements: [
              "최고 수익률",
              "가장 많은 투자자",
              "사회적 영향력 대상",
            ],
            story:
              "의료진단의 정확도를 95%까지 향상시킨 혁신적인 AI 시스템입니다. 3개 대형병원에서 도입되어 실제 환자 진료에 활용되고 있으며, 의료진의 업무 효율성을 크게 개선했습니다.",
            images: ["/api/placeholder/400/300"],
            featured: true,
          },
          {
            id: 2,
            title: "친환경 패키징 스타트업",
            description:
              "100% 생분해성 포장재를 개발하여 환경 문제 해결에 기여",
            category: "창업",
            creator: {
              username: "eco_packaging",
            },
            completedAt: "2024-01-28",
            duration: 203,
            stats: {
              totalInvestment: 62000,
              finalROI: 285,
              investorCount: 134,
              mentorCount: 6,
              averageRating: 4.7,
              totalMilestones: 6,
            },
            achievements: ["환경 혁신상", "지속가능성 대상"],
            story:
              "기존 플라스틱 포장재를 완전히 대체할 수 있는 혁신적인 소재를 개발했습니다. 현재 50여 개 기업에서 채택하여 연간 1,000톤의 플라스틱 사용량을 줄이는 데 기여하고 있습니다.",
            images: ["/api/placeholder/400/300"],
            featured: true,
          },
        ],
        projects: [
          {
            id: 3,
            title: "블록체인 기반 투표 시스템",
            description: "투명하고 안전한 전자투표 플랫폼",
            category: "IT/개발",
            creator: {
              username: "blockchain_democracy",
            },
            completedAt: "2024-03-10",
            duration: 165,
            stats: {
              totalInvestment: 45000,
              finalROI: 220,
              investorCount: 89,
              mentorCount: 5,
              averageRating: 4.6,
              totalMilestones: 5,
            },
            achievements: ["보안 혁신상"],
            story: "선거의 투명성과 보안을 획기적으로 개선한 시스템입니다.",
            images: ["/api/placeholder/400/300"],
            featured: false,
          },
          {
            id: 4,
            title: "온라인 교육 플랫폼",
            description: "AI 맞춤형 학습 시스템",
            category: "교육",
            creator: {
              username: "education_future",
            },
            completedAt: "2024-02-20",
            duration: 178,
            stats: {
              totalInvestment: 38000,
              finalROI: 195,
              investorCount: 67,
              mentorCount: 4,
              averageRating: 4.5,
              totalMilestones: 6,
            },
            achievements: ["교육 혁신상"],
            story: "개인 맞춤형 AI 튜터로 학습 효율성을 2배 향상시켰습니다.",
            images: ["/api/placeholder/400/300"],
            featured: false,
          },
          {
            id: 5,
            title: "스마트 농업 IoT 시스템",
            description: "농업 생산성 향상을 위한 IoT 솔루션",
            category: "IT/개발",
            creator: {
              username: "smart_farmer",
            },
            completedAt: "2024-01-15",
            duration: 194,
            stats: {
              totalInvestment: 52000,
              finalROI: 178,
              investorCount: 78,
              mentorCount: 6,
              averageRating: 4.4,
              totalMilestones: 7,
            },
            achievements: ["농업 혁신상"],
            story: "농작물 수확량을 30% 증가시킨 혁신적인 농업 기술입니다.",
            images: ["/api/placeholder/400/300"],
            featured: false,
          },
        ],
        leaderboards: {
          highestROI: [
            {
              rank: 1,
              projectId: 1,
              projectTitle: "AI 기반 의료진단 시스템",
              creator: "medical_ai_pioneer",
              value: 340,
              category: "IT/개발",
              completedAt: "2024-02-15",
            },
            {
              rank: 2,
              projectId: 2,
              projectTitle: "친환경 패키징 스타트업",
              creator: "eco_packaging",
              value: 285,
              category: "창업",
              completedAt: "2024-01-28",
            },
            {
              rank: 3,
              projectId: 3,
              projectTitle: "블록체인 기반 투표 시스템",
              creator: "blockchain_democracy",
              value: 220,
              category: "IT/개발",
              completedAt: "2024-03-10",
            },
            {
              rank: 4,
              projectId: 4,
              projectTitle: "온라인 교육 플랫폼",
              creator: "education_future",
              value: 195,
              category: "교육",
              completedAt: "2024-02-20",
            },
            {
              rank: 5,
              projectId: 5,
              projectTitle: "스마트 농업 IoT 시스템",
              creator: "smart_farmer",
              value: 178,
              category: "IT/개발",
              completedAt: "2024-01-15",
            },
          ],
          mostInvestors: [
            {
              rank: 1,
              projectId: 1,
              projectTitle: "AI 기반 의료진단 시스템",
              creator: "medical_ai_pioneer",
              value: 156,
              category: "IT/개발",
              completedAt: "2024-02-15",
            },
            {
              rank: 2,
              projectId: 2,
              projectTitle: "친환경 패키징 스타트업",
              creator: "eco_packaging",
              value: 134,
              category: "창업",
              completedAt: "2024-01-28",
            },
            {
              rank: 3,
              projectId: 3,
              projectTitle: "블록체인 기반 투표 시스템",
              creator: "blockchain_democracy",
              value: 89,
              category: "IT/개발",
              completedAt: "2024-03-10",
            },
            {
              rank: 4,
              projectId: 5,
              projectTitle: "스마트 농업 IoT 시스템",
              creator: "smart_farmer",
              value: 78,
              category: "IT/개발",
              completedAt: "2024-01-15",
            },
            {
              rank: 5,
              projectId: 4,
              projectTitle: "온라인 교육 플랫폼",
              creator: "education_future",
              value: 67,
              category: "교육",
              completedAt: "2024-02-20",
            },
          ],
          highestRating: [
            {
              rank: 1,
              projectId: 1,
              projectTitle: "AI 기반 의료진단 시스템",
              creator: "medical_ai_pioneer",
              value: 4.9,
              category: "IT/개발",
              completedAt: "2024-02-15",
            },
            {
              rank: 2,
              projectId: 2,
              projectTitle: "친환경 패키징 스타트업",
              creator: "eco_packaging",
              value: 4.7,
              category: "창업",
              completedAt: "2024-01-28",
            },
            {
              rank: 3,
              projectId: 3,
              projectTitle: "블록체인 기반 투표 시스템",
              creator: "blockchain_democracy",
              value: 4.6,
              category: "IT/개발",
              completedAt: "2024-03-10",
            },
            {
              rank: 4,
              projectId: 4,
              projectTitle: "온라인 교육 플랫폼",
              creator: "education_future",
              value: 4.5,
              category: "교육",
              completedAt: "2024-02-20",
            },
            {
              rank: 5,
              projectId: 5,
              projectTitle: "스마트 농업 IoT 시스템",
              creator: "smart_farmer",
              value: 4.4,
              category: "IT/개발",
              completedAt: "2024-01-15",
            },
          ],
          largestInvestment: [
            {
              rank: 1,
              projectId: 1,
              projectTitle: "AI 기반 의료진단 시스템",
              creator: "medical_ai_pioneer",
              value: 85000,
              category: "IT/개발",
              completedAt: "2024-02-15",
            },
            {
              rank: 2,
              projectId: 2,
              projectTitle: "친환경 패키징 스타트업",
              creator: "eco_packaging",
              value: 62000,
              category: "창업",
              completedAt: "2024-01-28",
            },
            {
              rank: 3,
              projectId: 5,
              projectTitle: "스마트 농업 IoT 시스템",
              creator: "smart_farmer",
              value: 52000,
              category: "IT/개발",
              completedAt: "2024-01-15",
            },
            {
              rank: 4,
              projectId: 3,
              projectTitle: "블록체인 기반 투표 시스템",
              creator: "blockchain_democracy",
              value: 45000,
              category: "IT/개발",
              completedAt: "2024-03-10",
            },
            {
              rank: 5,
              projectId: 4,
              projectTitle: "온라인 교육 플랫폼",
              creator: "education_future",
              value: 38000,
              category: "교육",
              completedAt: "2024-02-20",
            },
          ],
        },
      };

      setHallOfFameData(mockData);
      setLoading(false);
    };

    setTimeout(loadMockData, 500);
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ko-KR").format(amount);
  };

  const getRankIcon = (rank: number) => {
    switch (rank) {
      case 1:
        return <CrownOutlined style={{ color: "#FFD700", fontSize: "18px" }} />;
      case 2:
        return (
          <TrophyOutlined style={{ color: "#C0C0C0", fontSize: "16px" }} />
        );
      case 3:
        return (
          <TrophyOutlined style={{ color: "#CD7F32", fontSize: "16px" }} />
        );
      default:
        return (
          <span style={{ fontSize: "14px", fontWeight: "bold" }}>{rank}</span>
        );
    }
  };

  const categories = [
    { value: "all", label: "전체 카테고리" },
    { value: "IT/개발", label: "IT/개발" },
    { value: "창업", label: "창업" },
    { value: "교육", label: "교육" },
    { value: "헬스케어", label: "헬스케어" },
    { value: "라이프스타일", label: "라이프스타일" },
  ];

  const years = [
    { value: "all", label: "전체 기간" },
    { value: "2024", label: "2024년" },
    { value: "2023", label: "2023년" },
    { value: "2022", label: "2022년" },
  ];

  const filteredProjects =
    hallOfFameData?.projects.filter((project) => {
      const categoryMatch =
        selectedCategory === "all" || project.category === selectedCategory;
      const yearMatch =
        selectedYear === "all" ||
        new Date(project.completedAt).getFullYear().toString() === selectedYear;
      return categoryMatch && yearMatch;
    }) || [];

  const tabItems = [
    {
      key: "gallery",
      label: (
        <Space>
          <TrophyIcon size={16} />
          명예의 갤러리
        </Space>
      ),
      children: (
        <div>
          {/* 필터 */}
          <Row gutter={16} style={{ marginBottom: "24px" }}>
            <Col span={6}>
              <Select
                value={selectedCategory}
                onChange={setSelectedCategory}
                style={{ width: "100%" }}
                placeholder="카테고리 선택"
              >
                {categories.map((category) => (
                  <Option key={category.value} value={category.value}>
                    {category.label}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={6}>
              <Select
                value={selectedYear}
                onChange={setSelectedYear}
                style={{ width: "100%" }}
                placeholder="연도 선택"
              >
                {years.map((year) => (
                  <Option key={year.value} value={year.value}>
                    {year.label}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={12}>
              <Text type="secondary">
                총 {filteredProjects.length}개의 성공한 프로젝트
              </Text>
            </Col>
          </Row>

          {/* 특별 전시 프로젝트 */}
          <div style={{ marginBottom: "32px" }}>
            <Title level={4}>🌟 특별 전시</Title>
            <Row gutter={[24, 24]}>
              {hallOfFameData?.featuredProjects.map((project) => (
                <Col span={12} key={project.id}>
                  <Card
                    hoverable
                    onClick={() => navigate(`/project/${project.id}`)}
                    cover={
                      <div
                        style={{
                          height: "200px",
                          background:
                            "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          color: "white",
                          fontSize: "18px",
                          fontWeight: "bold",
                        }}
                      >
                        <TrophyIcon size={48} color="white" />
                      </div>
                    }
                    extra={<Tag color="gold">특별 전시</Tag>}
                  >
                    <Space direction="vertical" style={{ width: "100%" }}>
                      <div>
                        <Text strong style={{ fontSize: "18px" }}>
                          {project.title}
                        </Text>
                        <br />
                        <Text type="secondary">
                          @{project.creator.username}
                        </Text>
                      </div>

                      <Paragraph ellipsis={{ rows: 2 }}>
                        {project.description}
                      </Paragraph>

                      <Row gutter={16}>
                        <Col span={8}>
                          <Statistic
                            title="ROI"
                            value={project.stats.finalROI}
                            suffix="%"
                            valueStyle={{ fontSize: "16px", color: "#52c41a" }}
                          />
                        </Col>
                        <Col span={8}>
                          <Statistic
                            title="투자자"
                            value={project.stats.investorCount}
                            suffix="명"
                            valueStyle={{ fontSize: "16px" }}
                          />
                        </Col>
                        <Col span={8}>
                          <div style={{ textAlign: "center" }}>
                            <div
                              style={{ fontSize: "16px", fontWeight: "bold" }}
                            >
                              {project.stats.averageRating}
                            </div>
                            <Rate
                              disabled
                              value={project.stats.averageRating}
                              style={{ fontSize: "12px" }}
                            />
                          </div>
                        </Col>
                      </Row>

                      <Space wrap>
                        {project.achievements.map((achievement, index) => (
                          <Tag key={index} color="gold">
                            🏆 {achievement}
                          </Tag>
                        ))}
                      </Space>

                      <Paragraph
                        ellipsis={{ rows: 2 }}
                        style={{
                          fontStyle: "italic",
                          background: "var(--bg-secondary)",
                          padding: "8px",
                          borderRadius: "4px",
                          margin: 0,
                        }}
                      >
                        {project.story}
                      </Paragraph>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>

          {/* 일반 성공 프로젝트 */}
          <div>
            <Title level={4}>🎯 모든 성공 프로젝트</Title>
            <Row gutter={[16, 16]}>
              {filteredProjects.map((project) => (
                <Col span={8} key={project.id}>
                  <Card
                    hoverable
                    onClick={() => navigate(`/project/${project.id}`)}
                  >
                    <Space direction="vertical" style={{ width: "100%" }}>
                      <div
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                          alignItems: "center",
                        }}
                      >
                        <Text strong style={{ fontSize: "16px" }}>
                          {project.title}
                        </Text>
                        <Tag color="green">완료</Tag>
                      </div>

                      <div
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 8,
                        }}
                      >
                        <Avatar size="small">
                          {project.creator.username[0].toUpperCase()}
                        </Avatar>
                        <Text type="secondary">
                          @{project.creator.username}
                        </Text>
                      </div>

                      <Text type="secondary" style={{ fontSize: "12px" }}>
                        {project.description}
                      </Text>

                      <Divider style={{ margin: "8px 0" }} />

                      <Row gutter={8}>
                        <Col span={12}>
                          <div style={{ textAlign: "center" }}>
                            <Text
                              strong
                              style={{ color: "#52c41a", fontSize: "14px" }}
                            >
                              {project.stats.finalROI}%
                            </Text>
                            <div
                              style={{
                                fontSize: "10px",
                                color: "var(--text-secondary)",
                              }}
                            >
                              ROI
                            </div>
                          </div>
                        </Col>
                        <Col span={12}>
                          <div style={{ textAlign: "center" }}>
                            <Text strong style={{ fontSize: "14px" }}>
                              {formatCurrency(project.stats.totalInvestment)}
                            </Text>
                            <div
                              style={{
                                fontSize: "10px",
                                color: "var(--text-secondary)",
                              }}
                            >
                              투자액
                            </div>
                          </div>
                        </Col>
                      </Row>

                      <div
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                        }}
                      >
                        <Text style={{ fontSize: "11px" }}>
                          투자자: {project.stats.investorCount}명
                        </Text>
                        <Text style={{ fontSize: "11px" }}>
                          ⭐ {project.stats.averageRating}
                        </Text>
                      </div>

                      <Text type="secondary" style={{ fontSize: "10px" }}>
                        완료일:{" "}
                        {new Date(project.completedAt).toLocaleDateString(
                          "ko-KR"
                        )}
                      </Text>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>
        </div>
      ),
    },
    {
      key: "leaderboard",
      label: (
        <Space>
          <CrownOutlined />
          리더보드
        </Space>
      ),
      children: (
        <div>
          <Row gutter={[24, 24]}>
            {/* 최고 수익률 */}
            <Col span={12}>
              <Card
                title="🚀 최고 수익률 (ROI)"
                extra={<RiseOutlined style={{ color: "#52c41a" }} />}
              >
                <List
                  dataSource={hallOfFameData?.leaderboards.highestROI}
                  renderItem={(item) => (
                    <List.Item
                      onClick={() => navigate(`/project/${item.projectId}`)}
                      style={{ cursor: "pointer" }}
                    >
                      <List.Item.Meta
                        avatar={
                          <div
                            style={{
                              width: "32px",
                              height: "32px",
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "center",
                            }}
                          >
                            {getRankIcon(item.rank)}
                          </div>
                        }
                        title={
                          <div
                            style={{
                              display: "flex",
                              justifyContent: "space-between",
                            }}
                          >
                            <Text strong>{item.projectTitle}</Text>
                            <Text
                              style={{ color: "#52c41a", fontWeight: "bold" }}
                            >
                              {item.value}%
                            </Text>
                          </div>
                        }
                        description={
                          <Space>
                            <Text type="secondary">@{item.creator}</Text>
                            <Tag>{item.category}</Tag>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            </Col>

            {/* 최다 투자자 */}
            <Col span={12}>
              <Card title="👥 최다 투자자" extra={<TeamOutlined />}>
                <List
                  dataSource={hallOfFameData?.leaderboards.mostInvestors}
                  renderItem={(item) => (
                    <List.Item
                      onClick={() => navigate(`/project/${item.projectId}`)}
                      style={{ cursor: "pointer" }}
                    >
                      <List.Item.Meta
                        avatar={
                          <div
                            style={{
                              width: "32px",
                              height: "32px",
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "center",
                            }}
                          >
                            {getRankIcon(item.rank)}
                          </div>
                        }
                        title={
                          <div
                            style={{
                              display: "flex",
                              justifyContent: "space-between",
                            }}
                          >
                            <Text strong>{item.projectTitle}</Text>
                            <Text
                              style={{ color: "#1890ff", fontWeight: "bold" }}
                            >
                              {item.value}명
                            </Text>
                          </div>
                        }
                        description={
                          <Space>
                            <Text type="secondary">@{item.creator}</Text>
                            <Tag>{item.category}</Tag>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            </Col>

            {/* 최고 평점 */}
            <Col span={12}>
              <Card
                title="⭐ 최고 평점"
                extra={<StarOutlined style={{ color: "#faad14" }} />}
              >
                <List
                  dataSource={hallOfFameData?.leaderboards.highestRating}
                  renderItem={(item) => (
                    <List.Item
                      onClick={() => navigate(`/project/${item.projectId}`)}
                      style={{ cursor: "pointer" }}
                    >
                      <List.Item.Meta
                        avatar={
                          <div
                            style={{
                              width: "32px",
                              height: "32px",
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "center",
                            }}
                          >
                            {getRankIcon(item.rank)}
                          </div>
                        }
                        title={
                          <div
                            style={{
                              display: "flex",
                              justifyContent: "space-between",
                            }}
                          >
                            <Text strong>{item.projectTitle}</Text>
                            <Space>
                              <Text
                                style={{ color: "#faad14", fontWeight: "bold" }}
                              >
                                {item.value}
                              </Text>
                              <Rate
                                disabled
                                value={item.value}
                                style={{ fontSize: "12px" }}
                              />
                            </Space>
                          </div>
                        }
                        description={
                          <Space>
                            <Text type="secondary">@{item.creator}</Text>
                            <Tag>{item.category}</Tag>
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            </Col>

            {/* 최대 투자액 */}
            <Col span={12}>
              <Card
                title="💰 최대 투자액"
                extra={<DollarOutlined style={{ color: "#52c41a" }} />}
              >
                <List
                  dataSource={hallOfFameData?.leaderboards.largestInvestment}
                  renderItem={(item) => (
                    <List.Item
                      onClick={() => navigate(`/project/${item.projectId}`)}
                      style={{ cursor: "pointer" }}
                    >
                      <List.Item.Meta
                        avatar={
                          <div
                            style={{
                              width: "32px",
                              height: "32px",
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "center",
                            }}
                          >
                            {getRankIcon(item.rank)}
                          </div>
                        }
                        title={
                          <div
                            style={{
                              display: "flex",
                              justifyContent: "space-between",
                            }}
                          >
                            <Text strong>{item.projectTitle}</Text>
                            <Text
                              style={{ color: "#52c41a", fontWeight: "bold" }}
                            >
                              {formatCurrency(item.value)} USDC
                            </Text>
                          </div>
                        }
                        description={
                          <Space>
                            <Text type="secondary">@{item.creator}</Text>
                            <Tag>{item.category}</Tag>
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
      ),
    },
    {
      key: "stats",
      label: (
        <Space>
          <LineChartOutlined />
          통계 대시보드
        </Space>
      ),
      children: (
        <div>
          <Row gutter={[24, 24]} style={{ marginBottom: "32px" }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 성공 프로젝트"
                  value={hallOfFameData?.stats.totalSuccessfulProjects || 0}
                  prefix={<TrophyIcon size={16} />}
                  valueStyle={{ color: "var(--primary-color)" }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 가치 창출"
                  value={hallOfFameData?.stats.totalValueCreated || 0}
                  suffix="USDC"
                  prefix={<DollarOutlined />}
                  valueStyle={{ color: "#52c41a" }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="평균 ROI"
                  value={hallOfFameData?.stats.averageROI || 0}
                  suffix="%"
                  prefix={<RiseOutlined />}
                  valueStyle={{ color: "#722ed1" }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 참여자"
                  value={hallOfFameData?.stats.totalParticipants || 0}
                  suffix="명"
                  prefix={<TeamOutlined />}
                  valueStyle={{ color: "#fa8c16" }}
                />
              </Card>
            </Col>
          </Row>

          <Row gutter={[24, 24]}>
            <Col span={12}>
              <Card title="📊 성공 요인 분석">
                <Space direction="vertical" style={{ width: "100%" }}>
                  <div>
                    <Text>평균 완료 기간</Text>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Progress
                        percent={70}
                        showInfo={false}
                        strokeColor="#52c41a"
                        style={{ flex: 1, marginRight: "16px" }}
                      />
                      <Text strong>
                        {hallOfFameData?.stats.averageCompletionTime || 0}일
                      </Text>
                    </div>
                  </div>

                  <div>
                    <Text>카테고리 다양성</Text>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Progress
                        percent={85}
                        showInfo={false}
                        strokeColor="#1890ff"
                        style={{ flex: 1, marginRight: "16px" }}
                      />
                      <Text strong>
                        {hallOfFameData?.stats.categoriesRepresented || 0}개
                        분야
                      </Text>
                    </div>
                  </div>

                  <div>
                    <Text>커뮤니티 참여도</Text>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Progress
                        percent={92}
                        showInfo={false}
                        strokeColor="#722ed1"
                        style={{ flex: 1, marginRight: "16px" }}
                      />
                      <Text strong>매우 높음</Text>
                    </div>
                  </div>
                </Space>
              </Card>
            </Col>

            <Col span={12}>
              <Card title="🎯 성공 프로젝트 특징">
                <Space
                  direction="vertical"
                  size="middle"
                  style={{ width: "100%" }}
                >
                  <div
                    style={{
                      background: "var(--bg-secondary)",
                      padding: "12px",
                      borderRadius: "8px",
                    }}
                  >
                    <Text strong>💡 혁신성</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      성공한 프로젝트의 89%가 기존 솔루션을 혁신적으로 개선
                    </Text>
                  </div>

                  <div
                    style={{
                      background: "var(--bg-secondary)",
                      padding: "12px",
                      borderRadius: "8px",
                    }}
                  >
                    <Text strong>🤝 협력성</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      평균 5명 이상의 멘토와 활발한 소통을 유지
                    </Text>
                  </div>

                  <div
                    style={{
                      background: "var(--bg-secondary)",
                      padding: "12px",
                      borderRadius: "8px",
                    }}
                  >
                    <Text strong>🎯 명확성</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      구체적이고 측정 가능한 마일스톤 설정
                    </Text>
                  </div>

                  <div
                    style={{
                      background: "var(--bg-secondary)",
                      padding: "12px",
                      borderRadius: "8px",
                    }}
                  >
                    <Text strong>🌍 사회적 영향</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: "12px" }}>
                      76%가 사회 문제 해결에 기여하는 프로젝트
                    </Text>
                  </div>
                </Space>
              </Card>
            </Col>
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
              <TrophyIcon size={32} color="var(--primary-color)" />
              <div>
                <Title
                  level={2}
                  style={{ margin: 0, color: "var(--text-primary)" }}
                >
                  명예의 전당
                </Title>
                <Text type="secondary" style={{ fontSize: "16px" }}>
                  성공한 프로젝트들의 영광스러운 역사를 기록합니다
                </Text>
              </div>
            </Space>
          </div>

          {loading ? (
            <div style={{ textAlign: "center", padding: "100px" }}>
              <Text>명예의 기록을 불러오는 중...</Text>
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

export default HallOfFamePage;
