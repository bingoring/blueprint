import React, { useState, useEffect } from 'react';
import {
  Layout,
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Table,
  Button,
  Badge,
  Space,
  Typography,
  Tabs,
  Tag,
  Avatar,
  List,
  Empty,
  Spin,
  message
} from 'antd';
import {
  ProjectOutlined,
  DollarOutlined,
  TrophyOutlined,
  UserOutlined,
  ArrowLeftOutlined,
  PlusOutlined,
  CalendarOutlined,
  TeamOutlined,
  EditOutlined
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/useAuthStore';
import type { ProjectTableRecord, InvestmentTableRecord, ActivityRecord } from '../types';

const { Content } = Layout;
const { Title, Paragraph } = Typography;
// TabPane은 deprecated되어 제거됨 (items prop 사용)

// 임시 프로젝트 데이터
const mockUserProjects: ProjectTableRecord[] = [
  {
    id: 1,
    title: "3년 내 카페 창업 프로젝트",
    category: "business",
    status: "active",
    progress: 65,
    totalInvestment: 1250000,
    investors: 23,
    milestones: 4,
    currentMilestone: 2,
    createdAt: "2024-01-15",
    targetDate: "2027-12-31"
  },
  {
    id: 2,
    title: "AI 개발자 전직 프로젝트",
    category: "career",
    status: "active",
    progress: 85,
    totalInvestment: 800000,
    investors: 15,
    milestones: 5,
    currentMilestone: 4,
    createdAt: "2023-08-20",
    targetDate: "2025-06-30"
  },
  {
    id: 3,
    title: "정보처리기사 자격증 취득",
    category: "education",
    status: "active",
    progress: 25,
    totalInvestment: 0,
    investors: 0, // 투자자 없음 - 수정 가능
    milestones: 3,
    currentMilestone: 1,
    createdAt: "2024-12-01",
    targetDate: "2025-06-15"
  }
];

// 임시 투자 데이터
const mockInvestments: InvestmentTableRecord[] = [
  {
    id: 1,
    projectId: 3,
    projectTitle: "요가 강사 자격증 취득",
    developer: "박요가",
    amount: 50000,
    investedAt: "2024-01-20",
    status: "active",
    progress: 45
  },
  {
    id: 2,
    projectId: 4,
    projectTitle: "웹툰 작가 데뷔 프로젝트",
    developer: "김웹툰",
    amount: 100000,
    investedAt: "2024-01-10",
    status: "active",
    progress: 30
  }
];

const NewDashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState("overview");

  useEffect(() => {
    // TODO: 실제 데이터 로드
    loadUserData();
  }, []);

  const loadUserData = async () => {
    setLoading(true);
    try {
      // TODO: API 호출
      await new Promise(resolve => setTimeout(resolve, 1000)); // 임시 로딩
    } catch {
      message.error('데이터 로드에 실패했습니다');
    } finally {
      setLoading(false);
    }
  };

  // 통계 계산
  const totalProjects = mockUserProjects.length;
  const totalInvestments = mockInvestments.reduce((sum, inv) => sum + inv.amount, 0);
  const totalReceived = mockUserProjects.reduce((sum, proj) => sum + proj.totalInvestment, 0);
  const avgProgress = mockUserProjects.reduce((sum, proj) => sum + proj.progress, 0) / totalProjects;

  // 프로젝트 테이블 컬럼
  const projectColumns = [
    {
      title: t('project.projectTitle'),
      dataIndex: 'title',
      key: 'title',
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
      )
    },
    {
      title: t('project.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Badge
          status={status === 'completed' ? 'success' : 'processing'}
          text={t(`status.${status}`)}
        />
      )
    },
    {
      title: t('project.constructionProgress'),
      dataIndex: 'progress',
      key: 'progress',
      render: (progress: number) => (
        <Progress
          percent={progress}
          size="small"
          status={progress === 100 ? 'success' : 'active'}
        />
      )
    },
    {
      title: t('milestone.milestones'),
      key: 'milestones',
      render: (_: unknown, record: ProjectTableRecord) => (
        <span>{record.currentMilestone}/{record.milestones}</span>
      )
    },
    {
      title: t('investment.totalInvestment'),
      dataIndex: 'totalInvestment',
      key: 'totalInvestment',
      render: (amount: number) => `₩${amount.toLocaleString()}`
    },
    {
      title: t('investment.investors'),
      dataIndex: 'investors',
      key: 'investors',
      render: (count: number) => (
        <Space>
          <TeamOutlined />
          {count}명
        </Space>
      )
    },
    {
      title: '작업',
      key: 'actions',
      render: (_: unknown, record: ProjectTableRecord) => (
        <Space size="small">
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => navigate(`/edit-project/${record.id}`)}
            disabled={record.investors > 0}
            title={record.investors > 0 ? '투자자가 있는 프로젝트는 수정할 수 없습니다' : '프로젝트 수정'}
          >
            {record.investors > 0 ? '🔒' : '수정'}
          </Button>
        </Space>
      )
    }
  ];

  // 투자 테이블 컬럼
  const investmentColumns = [
    {
      title: t('project.projectTitle'),
      dataIndex: 'projectTitle',
      key: 'projectTitle',
      render: (title: string, record: InvestmentTableRecord) => (
        <div>
          <div className="font-medium">{title}</div>
          <div className="text-sm text-gray-500">{record.developer}</div>
        </div>
      )
    },
    {
      title: t('investment.investmentAmount'),
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number) => `₩${amount.toLocaleString()}`
    },
    {
      title: t('project.constructionProgress'),
      dataIndex: 'progress',
      key: 'progress',
      render: (progress: number) => (
        <Progress percent={progress} size="small" />
      )
    },
    {
      title: '투자일',
      dataIndex: 'investedAt',
      key: 'investedAt'
    },
    {
      title: t('project.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'orange'}>
          {status === 'active' ? '진행중' : '완료'}
        </Tag>
      )
    }
  ];

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <Layout className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b p-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/')}
              type="text"
            >
              홈으로
            </Button>
            <Title level={3} className="!mb-0">
              {t('nav.dashboard')}
            </Title>
          </div>

          <Space>
            <Button type="primary" icon={<PlusOutlined />}>
              {t('project.newProject')}
            </Button>
          </Space>
        </div>
      </div>

      <Content className="p-6 max-w-7xl mx-auto w-full">
        {/* 사용자 정보 */}
        <Card className="mb-6">
          <Row align="middle" gutter={[16, 16]}>
            <Col>
              <Avatar size={64} icon={<UserOutlined />} />
            </Col>
            <Col flex="auto">
              <Title level={4} className="!mb-1">
                환영합니다, {user?.username}님! 👋
              </Title>
              <Paragraph className="!mb-0 text-gray-600">
                현재 {totalProjects}개의 프로젝트를 진행하고 있습니다.
              </Paragraph>
            </Col>
          </Row>
        </Card>

        {/* 통계 카드 */}
        <Row gutter={[16, 16]} className="mb-6">
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title={t('project.myProjects')}
                value={totalProjects}
                prefix={<ProjectOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="받은 총 투자금"
                value={totalReceived}
                prefix="₩"
                precision={0}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title={t('investment.myInvestments')}
                value={totalInvestments}
                prefix="₩"
                precision={0}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="평균 진행률"
                value={avgProgress}
                suffix="%"
                precision={1}
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#722ed1' }}
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
                key: 'projects',
                label: (
                  <span>
                    <ProjectOutlined />
                    {t('project.myProjects')}
                  </span>
                ),
                children: (
                  <Table
                    columns={projectColumns}
                    dataSource={mockUserProjects}
                    rowKey="id"
                    pagination={false}
                    locale={{
                      emptyText: (
                        <Empty
                          description="아직 등록된 프로젝트가 없습니다"
                          image={Empty.PRESENTED_IMAGE_SIMPLE}
                        >
                          <Button type="primary" icon={<PlusOutlined />}>
                            첫 프로젝트 만들기
                          </Button>
                        </Empty>
                      )
                    }}
                  />
                )
              },
              {
                key: 'investments',
                label: (
                  <span>
                    <DollarOutlined />
                    {t('investment.myInvestments')}
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
                      )
                    }}
                  />
                )
              },
              {
                key: 'activity',
                label: (
                  <span>
                    <CalendarOutlined />
                    최근 활동
                  </span>
                ),
                children: (
                  <List
                    itemLayout="horizontal"
                    dataSource={([
                      {
                        id: 1,
                        type: 'investment' as const,
                        title: '새로운 투자를 받았습니다',
                        description: '김투자님이 카페 창업 프로젝트에 50,000원을 투자했습니다',
                        time: '2시간 전'
                      },
                      {
                        id: 2,
                        type: 'milestone' as const,
                        title: '단계가 완료되었습니다',
                        description: 'AI 개발자 프로젝트의 3단계가 완료되었습니다',
                        time: '1일 전'
                      },
                      {
                        id: 3,
                        type: 'project' as const,
                        title: '새 프로젝트를 등록했습니다',
                        description: '카페 창업 프로젝트를 등록했습니다',
                        time: '3일 전'
                      }
                    ] as ActivityRecord[])}
                    renderItem={(item: ActivityRecord) => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={
                            <Avatar
                              icon={
                                item.type === 'investment' ? <DollarOutlined /> :
                                item.type === 'milestone' ? <CalendarOutlined /> :
                                <ProjectOutlined />
                              }
                            />
                          }
                          title={item.title}
                          description={
                            <div>
                              <div>{item.description}</div>
                              <div className="text-sm text-gray-400 mt-1">{item.time}</div>
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
                      )
                    }}
                  />
                )
              }
            ]}
          />
        </Card>
      </Content>
    </Layout>
  );
};

export default NewDashboard;
