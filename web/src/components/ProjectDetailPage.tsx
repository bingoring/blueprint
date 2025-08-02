import React, { useState, useEffect } from 'react';
import {
  Layout,
  Card,
  Row,
  Col,
  Progress,
  Button,
  Typography,
  Space,
  Avatar,
  List,
  Badge,
  Statistic,
  Tag,
  Divider,
  Modal,
  Form,
  InputNumber,
  Radio,
  message,
  Timeline,
  Empty,
  Spin,
  Breadcrumb
} from 'antd';
import {
  LeftOutlined,
  DollarOutlined,
  TeamOutlined,
  CalendarOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/useAuthStore';
import type {
  Project,
  ProjectCategory
} from '../types';
import dayjs from 'dayjs';

const { Content } = Layout;
const { Title, Paragraph, Text } = Typography;

// 투자 모달에서 사용할 인터페이스
interface InvestmentOption {
  milestone_id: number;
  option: string;
  amount: number;
}

interface Investor {
  id: number;
  username: string;
  avatar?: string;
  amount: number;
  invested_at: string;
  milestone_bets: {
    milestone_id: number;
    option: string;
    amount: number;
  }[];
}

interface ProjectStats {
  total_investment: number;
  total_investors: number;
  completion_rate: number;
  expected_return: number;
}

const ProjectDetailPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { isAuthenticated, user } = useAuthStore();

  // 상태 관리
  const [loading, setLoading] = useState(true);
  const [project, setProject] = useState<Project | null>(null);
  const [projectStats, setProjectStats] = useState<ProjectStats | null>(null);
  const [investors, setInvestors] = useState<Investor[]>([]);
  const [isOwner, setIsOwner] = useState(false);

  // 투자 모달 상태
  const [investModalVisible, setInvestModalVisible] = useState(false);
  const [investmentForm] = Form.useForm();
  const [totalInvestmentAmount, setTotalInvestmentAmount] = useState(0);
  const [milestoneInvestments, setMilestoneInvestments] = useState<InvestmentOption[]>([]);
  const [investmentLoading, setInvestmentLoading] = useState(false);

  // 프로젝트 데이터 로드
  useEffect(() => {
    if (!id) {
      message.error('프로젝트 ID가 필요합니다');
      navigate('/');
      return;
    }

    loadProjectData();
  }, [id, navigate]);

  const loadProjectData = async () => {
    try {
      setLoading(true);

      // 임시 목 데이터 - 실제로는 API 호출
      const mockProject: Project = {
        id: parseInt(id!),
        user_id: 1,
        title: "3년 내 카페 창업 프로젝트",
        description: "서울 강남구에서 독립 카페를 창업하는 것이 목표입니다. 특별한 원두와 독특한 인테리어로 차별화된 카페를 만들고 싶습니다. 지역 주민들에게 사랑받는 공간이 되는 것이 꿈입니다.",
        category: "business" as ProjectCategory,
        status: "active",
        target_date: "2027-12-31",
        budget: 50000000,
        priority: 3,
        is_public: true,
        tags: '{"location": "강남구", "type": "카페", "concept": "독립카페"}',
        metrics: "월 매출 500만원 이상 달성",
        created_at: "2024-01-15T10:00:00Z",
        updated_at: "2024-01-15T10:00:00Z",
        milestones: [
          {
            title: "사업 계획서 작성 및 승인",
            description: "상세한 사업 계획서를 작성하고 전문가 검토를 받아 승인받기",
            target_date: "2025-03-31",
            order: 1,
            betting_type: "simple",
            betting_options: []
          },
          {
            title: "창업 자금 조달",
            description: "카페 창업에 필요한 자금 5000만원을 조달하기",
            target_date: "2025-12-31",
            order: 2,
            betting_type: "custom",
            betting_options: ["3000만원 달성", "5000만원 달성", "7000만원 이상 달성"]
          },
          {
            title: "매장 임대 및 인테리어",
            description: "강남구 내 적절한 위치의 매장을 임대하고 인테리어 완료",
            target_date: "2026-06-30",
            order: 3,
            betting_type: "custom",
            betting_options: ["6개월 내 완료", "1년 내 완료", "1년 초과"]
          },
          {
            title: "카페 오픈 및 운영",
            description: "카페를 정식 오픈하고 안정적인 운영 궤도에 진입",
            target_date: "2027-12-31",
            order: 4,
            betting_type: "simple",
            betting_options: []
          }
        ]
      };

      const mockStats: ProjectStats = {
        total_investment: 1250000,
        total_investors: 23,
        completion_rate: 65,
        expected_return: 15.2
      };

      const mockInvestors: Investor[] = [
        {
          id: 1,
          username: "김투자",
          avatar: undefined,
          amount: 100000,
          invested_at: "2024-01-20T10:00:00Z",
          milestone_bets: [
            { milestone_id: 2, option: "5000만원 달성", amount: 50000 },
            { milestone_id: 3, option: "6개월 내 완료", amount: 50000 }
          ]
        },
        {
          id: 2,
          username: "박서포터",
          avatar: undefined,
          amount: 75000,
          invested_at: "2024-01-22T14:30:00Z",
          milestone_bets: [
            { milestone_id: 2, option: "7000만원 이상 달성", amount: 75000 }
          ]
        },
        {
          id: 3,
          username: "이응원",
          avatar: undefined,
          amount: 50000,
          invested_at: "2024-01-25T09:15:00Z",
          milestone_bets: [
            { milestone_id: 3, option: "1년 내 완료", amount: 50000 }
          ]
        }
      ];

      setProject(mockProject);
      setProjectStats(mockStats);
      setInvestors(mockInvestors);
      setIsOwner(Number(user?.id) === mockProject.user_id);

    } catch (error) {
      console.error('프로젝트 데이터 로드 실패:', error);
      message.error('프로젝트를 불러오는데 실패했습니다');
      navigate('/');
    } finally {
      setLoading(false);
    }
  };

  // 투자 모달 열기
  const openInvestModal = () => {
    if (!isAuthenticated) {
      message.warning('로그인이 필요합니다');
      return;
    }

    if (isOwner) {
      message.warning('자신의 프로젝트에는 투자할 수 없습니다');
      return;
    }

    setInvestModalVisible(true);

    // 마일스톤별 투자 옵션 초기화
    const initialInvestments = project?.milestones?.map(milestone => ({
      milestone_id: milestone.order,
      option: milestone.betting_type === 'simple' ? '성공' : (milestone.betting_options?.[0] || ''),
      amount: 0
    })) || [];

    setMilestoneInvestments(initialInvestments);
  };

  // 투자 옵션 업데이트
  const updateMilestoneInvestment = (milestoneId: number, field: 'option' | 'amount', value: string | number) => {
    setMilestoneInvestments(prev =>
      prev.map(investment =>
        investment.milestone_id === milestoneId
          ? { ...investment, [field]: value }
          : investment
      )
    );
  };

  // 총 투자 금액 계산
  useEffect(() => {
    const total = milestoneInvestments.reduce((sum, investment) => sum + investment.amount, 0);
    setTotalInvestmentAmount(total);
  }, [milestoneInvestments]);

  // 투자 실행
  const handleInvest = async () => {
    try {
      setInvestmentLoading(true);

      const validInvestments = milestoneInvestments.filter(inv => inv.amount > 0);

      if (validInvestments.length === 0) {
        message.warning('투자할 마일스톤을 선택하고 금액을 입력해주세요');
        return;
      }

      if (totalInvestmentAmount < 1000) {
        message.warning('최소 투자 금액은 1,000원입니다');
        return;
      }

      // 실제로는 API 호출
      console.log('투자 데이터:', {
        project_id: project?.id,
        total_amount: totalInvestmentAmount,
        milestone_bets: validInvestments
      });

      message.success(`총 ${totalInvestmentAmount.toLocaleString()}원이 투자되었습니다! 🎉`);
      setInvestModalVisible(false);

      // 프로젝트 데이터 새로고침
      await loadProjectData();

    } catch (error) {
      console.error('투자 실패:', error);
      message.error('투자에 실패했습니다');
    } finally {
      setInvestmentLoading(false);
    }
  };

  // 뒤로가기
  const handleBack = () => {
    navigate('/dashboard');
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <Spin size="large" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <Empty description="프로젝트를 찾을 수 없습니다" />
      </div>
    );
  }

  const projectTags = project.tags ? JSON.parse(project.tags) : {};

  return (
    <Layout className="min-h-screen bg-gray-50">
      <Content className="p-6">
        <div className="max-w-6xl mx-auto">
          {/* 브레드크럼 & 뒤로가기 */}
          <div className="mb-6">
            <Button
              icon={<LeftOutlined />}
              onClick={handleBack}
              className="mb-4"
            >
              대시보드로 돌아가기
            </Button>

            <Breadcrumb>
              <Breadcrumb.Item>홈</Breadcrumb.Item>
              <Breadcrumb.Item>프로젝트</Breadcrumb.Item>
              <Breadcrumb.Item>{project.title}</Breadcrumb.Item>
            </Breadcrumb>
          </div>

          <Row gutter={[24, 24]}>
            {/* 왼쪽: 프로젝트 메인 정보 */}
            <Col md={16} span={24}>
              {/* 프로젝트 헤더 */}
              <Card className="mb-6">
                <div className="flex justify-between items-start mb-4">
                  <div className="flex-1">
                    <Title level={2} className="mb-2">
                      {project.title}
                    </Title>
                    <Space size="middle" className="mb-3">
                      <Tag color="blue">
                        {t(`category.${project.category}`)}
                      </Tag>
                      <Badge
                        status={project.status === 'active' ? 'processing' : 'success'}
                        text={t(`status.${project.status}`)}
                      />
                      <Text type="secondary">
                        <CalendarOutlined className="mr-1" />
                        목표일: {dayjs(project.target_date).format('YYYY-MM-DD')}
                      </Text>
                    </Space>
                  </div>

                  {!isOwner && (
                    <Button
                      type="primary"
                      size="large"
                      icon={<DollarOutlined />}
                      onClick={openInvestModal}
                    >
                      투자하기
                    </Button>
                  )}
                </div>

                {/* 진행률 */}
                <div className="mb-4">
                  <div className="flex justify-between items-center mb-2">
                    <Text strong>프로젝트 진행률</Text>
                    <Text className="text-lg font-bold">
                      {projectStats?.completion_rate}%
                    </Text>
                  </div>
                  <Progress
                    percent={projectStats?.completion_rate}
                    strokeColor={{
                      '0%': '#108ee9',
                      '100%': '#87d068',
                    }}
                    size="default"
                  />
                </div>

                {/* 프로젝트 설명 */}
                <Paragraph className="text-gray-700 leading-relaxed">
                  {project.description}
                </Paragraph>

                {/* 태그들 */}
                {Object.keys(projectTags).length > 0 && (
                  <div className="mt-4">
                    <Text strong className="mr-3">태그:</Text>
                    <Space wrap>
                      {Object.entries(projectTags).map(([key, value]) => (
                        <Tag key={key} color="geekblue">
                          {key}: {value as string}
                        </Tag>
                      ))}
                    </Space>
                  </div>
                )}

                {project.metrics && (
                  <div className="mt-4 p-3 bg-blue-50 rounded-lg">
                    <Text strong className="text-blue-800">
                      <TrophyOutlined className="mr-2" />
                      성공 지표: {project.metrics}
                    </Text>
                  </div>
                )}
              </Card>

              {/* 마일스톤 타임라인 */}
              <Card title="📋 프로젝트 마일스톤">
                <Timeline>
                  {project.milestones?.map((milestone, index) => (
                    <Timeline.Item
                      key={milestone.order}
                      dot={
                        index < 2 ? <CheckCircleOutlined className="text-green-500" /> :
                        index === 2 ? <ClockCircleOutlined className="text-blue-500" /> :
                        <ExclamationCircleOutlined className="text-gray-400" />
                      }
                    >
                      <div className="pb-4">
                        <div className="flex justify-between items-start mb-2">
                          <Title level={5} className="mb-1">
                            {milestone.title}
                          </Title>
                          <Tag color={index < 2 ? 'green' : index === 2 ? 'blue' : 'default'}>
                            {index < 2 ? '완료' : index === 2 ? '진행중' : '대기'}
                          </Tag>
                        </div>

                        <Paragraph className="text-gray-600 mb-2">
                          {milestone.description}
                        </Paragraph>

                        <div className="flex justify-between items-center">
                          <Text type="secondary">
                            목표일: {dayjs(milestone.target_date).format('YYYY-MM-DD')}
                          </Text>

                          <div className="text-right">
                            <Text strong>
                              투자 옵션: {milestone.betting_type === 'simple' ? '📍 단순' : '🎯 사용자 정의'}
                            </Text>
                            {milestone.betting_type === 'custom' && (
                              <div className="mt-1">
                                <Space size="small" wrap>
                                  {milestone.betting_options?.map((option, optionIndex) => (
                                    <Tag key={optionIndex} color="purple">
                                      {option}
                                    </Tag>
                                  ))}
                                </Space>
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </Card>
            </Col>

            {/* 오른쪽: 투자 정보 */}
            <Col md={8} span={24}>
              {/* 투자 통계 */}
              <Card className="mb-6">
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Statistic
                      title="총 투자금액"
                      value={projectStats?.total_investment}
                      formatter={(value) => `₩${value?.toLocaleString()}`}
                      prefix={<DollarOutlined />}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="투자자 수"
                      value={projectStats?.total_investors}
                      suffix="명"
                      prefix={<TeamOutlined />}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="예상 수익률"
                      value={projectStats?.expected_return}
                      suffix="%"
                      prefix={<TrophyOutlined />}
                      valueStyle={{ color: '#3f8600' }}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="진행률"
                      value={projectStats?.completion_rate}
                      suffix="%"
                      prefix={<ClockCircleOutlined />}
                    />
                  </Col>
                </Row>
              </Card>

              {/* 최근 투자자들 */}
              <Card
                title="💰 최근 투자자"
                extra={
                  <Text type="secondary">
                    총 {projectStats?.total_investors}명
                  </Text>
                }
              >
                <List
                  itemLayout="horizontal"
                  dataSource={investors.slice(0, 5)}
                  renderItem={(investor: Investor) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={
                          <Avatar
                            src={investor.avatar}
                            icon={!investor.avatar ? <TeamOutlined /> : undefined}
                          />
                        }
                        title={investor.username}
                        description={
                          <div>
                            <div>₩{investor.amount.toLocaleString()}</div>
                            <div className="text-xs text-gray-400">
                              {dayjs(investor.invested_at).format('MM-DD HH:mm')}
                            </div>
                          </div>
                        }
                      />
                    </List.Item>
                  )}
                  locale={{
                    emptyText: "아직 투자자가 없습니다"
                  }}
                />

                {investors.length > 5 && (
                  <div className="text-center mt-3">
                    <Button type="link" size="small">
                      전체 투자자 보기 ({investors.length}명)
                    </Button>
                  </div>
                )}
              </Card>
            </Col>
          </Row>

          {/* 투자 모달 */}
          <Modal
            title="💰 프로젝트 투자하기"
            open={investModalVisible}
            onCancel={() => setInvestModalVisible(false)}
            footer={[
              <Button key="cancel" onClick={() => setInvestModalVisible(false)}>
                취소
              </Button>,
              <Button
                key="invest"
                type="primary"
                loading={investmentLoading}
                onClick={handleInvest}
                disabled={totalInvestmentAmount === 0}
              >
                ₩{totalInvestmentAmount.toLocaleString()} 투자하기
              </Button>
            ]}
            width={800}
          >
            <Form form={investmentForm} layout="vertical">
              <div className="mb-4 p-4 bg-blue-50 rounded-lg">
                <Text strong className="text-blue-800">
                  💡 투자 안내: 각 마일스톤별로 결과를 예측하고 투자하세요.
                  정확한 예측에 대한 보상을 받을 수 있습니다.
                </Text>
              </div>

              <div className="space-y-6">
                {project.milestones?.map((milestone) => (
                  <Card key={milestone.order} size="small" className="border-gray-200">
                    <div className="mb-3">
                      <Title level={5} className="mb-1">
                        🎯 마일스톤 {milestone.order}: {milestone.title}
                      </Title>
                      <Text type="secondary" className="text-sm">
                        목표일: {dayjs(milestone.target_date).format('YYYY-MM-DD')}
                      </Text>
                    </div>

                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Form.Item label="예측 선택">
                          <Radio.Group
                            value={milestoneInvestments.find(inv => inv.milestone_id === milestone.order)?.option}
                            onChange={(e) => updateMilestoneInvestment(milestone.order, 'option', e.target.value)}
                          >
                            <Space direction="vertical">
                              {milestone.betting_type === 'simple' ? (
                                <>
                                  <Radio value="성공">✅ 성공</Radio>
                                  <Radio value="실패">❌ 실패</Radio>
                                </>
                              ) : (
                                milestone.betting_options?.map((option, optionIndex) => (
                                  <Radio key={optionIndex} value={option}>
                                    {option}
                                  </Radio>
                                ))
                              )}
                            </Space>
                          </Radio.Group>
                        </Form.Item>
                      </Col>

                      <Col span={12}>
                        <Form.Item label="투자 금액 (원)">
                          <InputNumber
                            style={{ width: '100%' }}
                            min={0}
                            max={1000000}
                            step={1000}
                            placeholder="투자할 금액"
                            formatter={value => `₩ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                            parser={value => Number(value!.replace(/₩\s?|(,*)/g, '')) || 0}
                            value={milestoneInvestments.find(inv => inv.milestone_id === milestone.order)?.amount}
                            onChange={(value) => updateMilestoneInvestment(milestone.order, 'amount', value || 0)}
                          />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                ))}
              </div>

              <Divider />

              <div className="text-center p-4 bg-green-50 rounded-lg">
                <Statistic
                  title="총 투자 금액"
                  value={totalInvestmentAmount}
                  formatter={(value) => `₩${value?.toLocaleString()}`}
                  valueStyle={{ color: '#3f8600', fontSize: '24px' }}
                />
                <Text type="secondary" className="text-sm">
                  최소 투자 금액: ₩1,000 | 최대 투자 금액: ₩1,000,000
                </Text>
              </div>
            </Form>
          </Modal>
        </div>
      </Content>
    </Layout>
  );
};

export default ProjectDetailPage;
