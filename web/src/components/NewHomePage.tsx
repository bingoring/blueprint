import React, { useState } from 'react';
import { Layout, Button, Typography, Row, Col, Card, Statistic, Space, Tag } from 'antd';
import {
  ProjectOutlined,
  TrophyOutlined,
  UserOutlined,
  PlusOutlined,
  LoginOutlined
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/useAuthStore';
import AuthModal from './AuthModal';
import LanguageSwitcher from './LanguageSwitcher';

const { Header, Content } = Layout;
const { Title, Paragraph } = Typography;

// 임시 모의 데이터 (확정된 워딩 체계)
const mockProjects = [
  {
    id: 1,
    title: "3년 내 카페 창업 프로젝트",
    description: "서울 강남구에서 독립 카페 창업",
    category: "business",
    probability: 72,
    totalInvestment: 1250000,
    timeLeft: "2년 3개월",
    developer: "김창업",
    investors: 23,
    milestones: 4,
    currentMilestone: 2,
    trending: true
  },
  {
    id: 2,
    title: "AI 개발자 전직 프로젝트",
    description: "웹 개발자에서 AI/ML 엔지니어로 전환",
    category: "career",
    probability: 85,
    totalInvestment: 800000,
    timeLeft: "1년 6개월",
    developer: "이개발",
    investors: 15,
    milestones: 5,
    currentMilestone: 3,
    trending: false
  },
  {
    id: 3,
    title: "요가 강사 자격증 취득",
    description: "국제 요가 강사 자격증 취득 후 스튜디오 개업",
    category: "personal",
    probability: 68,
    totalInvestment: 600000,
    timeLeft: "8개월",
    developer: "박요가",
    investors: 8,
    milestones: 3,
    currentMilestone: 1,
    trending: false
  }
];

interface Project {
  id: number;
  title: string;
  description: string;
  category: string;
  probability: number;
  totalInvestment: number;
  timeLeft: string;
  developer: string;
  investors: number;
  milestones: number;
  currentMilestone: number;
  trending: boolean;
}

const NewHomePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isAuthenticated, user, logout } = useAuthStore();
  const [currentView, setCurrentView] = useState<'home' | 'dashboard'>('home');
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);

  // 대시보드로 전환
  if (currentView === 'dashboard') {
    navigate('/dashboard');
    return null;
  }

  const handleCreateProject = () => {
    if (!isAuthenticated) {
      setIsAuthModalOpen(true);
      return;
    }
    navigate('/create-project');
  };

  const handleLogout = async () => {
    await logout();
  };

  return (
    <Layout className="min-h-screen">
      {/* Header */}
      <Header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto flex items-center justify-between h-full">
          {/* Logo */}
          <div className="flex items-center space-x-2">
            <ProjectOutlined className="text-2xl text-blue-600" />
            <Title level={3} className="!mb-0 !text-blue-600">
              {t('home.title')}
            </Title>
          </div>

          {/* Navigation */}
          <div className="flex items-center space-x-4">
            <LanguageSwitcher size="small" />

            {isAuthenticated ? (
              <Space>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleCreateProject}
                >
                  {t('project.newProject')}
                </Button>
                <Button onClick={() => setCurrentView('dashboard')}>
                  {t('nav.dashboard')}
                </Button>
                <Button onClick={handleLogout}>
                  {t('nav.logout')}
                </Button>
                <span>환영합니다, {user?.username}님</span>
              </Space>
            ) : (
              <Space>
                <Button
                  icon={<LoginOutlined />}
                  onClick={() => setIsAuthModalOpen(true)}
                >
                  {t('auth.login')}
                </Button>
                <Button
                  type="primary"
                  onClick={() => setIsAuthModalOpen(true)}
                >
                  {t('auth.register')}
                </Button>
              </Space>
            )}
          </div>
        </div>
      </Header>

      {/* Content */}
      <Content className="bg-gray-50">
        {/* Hero Section */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white py-20">
          <div className="max-w-7xl mx-auto px-4 text-center">
            <Title level={1} className="!text-white !mb-4">
              {t('home.subtitle')}
            </Title>
            <Paragraph className="!text-blue-100 text-lg mb-8 max-w-2xl mx-auto">
              {t('home.description')}
            </Paragraph>
            <Space size="large">
              <Button
                size="large"
                type="primary"
                ghost
                icon={<PlusOutlined />}
                onClick={handleCreateProject}
              >
                {t('home.startProject')}
              </Button>
              <Button size="large" ghost>
                {t('home.browseProjects')}
              </Button>
            </Space>
          </div>
        </div>

        {/* Stats Bar */}
        <div className="bg-white py-12 shadow-sm">
          <div className="max-w-7xl mx-auto px-4">
            <Row gutter={[32, 16]} className="text-center">
              <Col span={6}>
                <Statistic
                  title={t('home.stats.activeProjects')}
                  value={156}
                  prefix={<ProjectOutlined />}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title={t('home.stats.totalInvestment')}
                  value={45000000}
                  prefix="₩"
                  precision={0}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title={t('home.stats.successRate')}
                  value={78}
                  suffix="%"
                  prefix={<TrophyOutlined />}
                  valueStyle={{ color: '#faad14' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title={t('home.stats.totalUsers')}
                  value={2341}
                  prefix={<UserOutlined />}
                  valueStyle={{ color: '#722ed1' }}
                />
              </Col>
            </Row>
          </div>
        </div>

        {/* Featured Projects */}
        <div className="max-w-7xl mx-auto px-4 py-12">
          <Title level={2} className="mb-8">
            {t('home.featuredProjects')}
          </Title>

          <Row gutter={[24, 24]}>
            {mockProjects.map((project: Project) => (
              <Col key={project.id} xs={24} md={12} lg={8}>
                <Card
                  hoverable
                  className="h-full cursor-pointer"
                  onClick={() => navigate(`/project/${project.id}`)}
                  actions={[
                    <Button
                      type="primary"
                      block
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/project/${project.id}`);
                      }}
                    >
                      {t('investment.invest')}
                    </Button>
                  ]}
                >
                  <div className="flex justify-between items-start mb-3">
                    <Tag color={project.trending ? 'red' : 'blue'}>
                      {project.trending ? '🔥 인기' : t(`categories.${project.category}`)}
                    </Tag>
                    <span className="text-green-600 font-semibold">
                      {project.probability}%
                    </span>
                  </div>

                  <Title level={4} className="mb-2">
                    {project.title}
                  </Title>

                  <Paragraph className="text-gray-600 mb-4">
                    {project.description}
                  </Paragraph>

                  <div className="space-y-2 text-sm text-gray-500">
                    <div className="flex justify-between">
                      <span>{t('project.developer')}:</span>
                      <span>{project.developer}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>{t('investment.totalInvestment')}:</span>
                      <span>₩{project.totalInvestment.toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>{t('investment.investors')}:</span>
                      <span>{project.investors}명</span>
                    </div>
                    <div className="flex justify-between">
                      <span>{t('project.constructionProgress')}:</span>
                      <span>{project.currentMilestone}/{project.milestones} {t('milestone.milestones')}</span>
                    </div>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </Content>

      {/* Modals */}
      <AuthModal
        isOpen={isAuthModalOpen}
        onClose={() => setIsAuthModalOpen(false)}
      />
    </Layout>
  );
};

export default NewHomePage;
