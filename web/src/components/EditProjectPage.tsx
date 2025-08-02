import React, { useState, useEffect, useRef } from 'react';
import {
  Steps,
  Card,
  Form,
  Input,
  Select,
  DatePicker,
  Button,
  Space,
  Typography,
  Row,
  Col,
  Collapse,
  InputNumber,
  Tag,
  Divider,
  message,
  Tooltip,
  Switch,
  Alert,
  Radio,
  Spin
} from 'antd';
import type { InputRef } from 'antd';
import {
  ProjectOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  PlusOutlined,
  RobotOutlined,
  SettingOutlined,
  DollarOutlined,
  TagsOutlined,
  InfoCircleOutlined,
  LeftOutlined,
  EditOutlined,
  LockOutlined
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuthStore } from '../stores/useAuthStore';
import { apiClient } from '../lib/api';
import type {
  CreateProjectWithMilestonesRequest,
  AIMilestoneResponse,
  AIUsageInfo,
  ProjectMilestone,
  Project,
  AIMilestone,
  ProjectCategory
} from '../types';
import dayjs from 'dayjs';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Panel } = Collapse;
const { Step } = Steps;

interface TagPair {
  key: string;
  value: string;
}

interface CustomBettingOptionsProps {
  milestoneIndex: number;
  milestone: ProjectMilestone;
  onAddOption: (milestoneIndex: number, option: string) => void;
  onRemoveOption: (milestoneIndex: number, optionIndex: number) => void;
  disabled?: boolean;
}

// 사용자 정의 투자 옵션 컴포넌트
const CustomBettingOptions: React.FC<CustomBettingOptionsProps> = ({
  milestoneIndex,
  milestone,
  onAddOption,
  onRemoveOption,
  disabled = false
}) => {
  const [newOption, setNewOption] = useState('');

  const handleAddOption = () => {
    const trimmedOption = newOption.trim();
    if (!trimmedOption) return;

    // 중복 체크
    const existingOptions = milestone.betting_options || [];
    if (existingOptions.some(option => option.toLowerCase() === trimmedOption.toLowerCase())) {
      message.warning(`"${trimmedOption}" 옵션이 이미 존재합니다`);
      return;
    }

    onAddOption(milestoneIndex, trimmedOption);
    setNewOption('');
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleAddOption();
    }
  };

  return (
    <div className="space-y-3">
      <div>
        <Text type="secondary" className="text-sm">
          투자자들이 선택할 수 있는 옵션들을 추가하세요. 예: "1년 내 완료", "2년 내 완료", "3년 내 완료"
        </Text>
      </div>

      {!disabled && (
        <Row gutter={[8, 8]}>
          <Col span={16}>
            <Input
              placeholder="새 투자 옵션 입력 (예: 1년 내 완료)"
              value={newOption}
              onChange={(e) => setNewOption(e.target.value)}
              onKeyPress={handleKeyPress}
            />
          </Col>
          <Col span={8}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAddOption}
              disabled={!newOption.trim()}
              block
            >
              옵션 추가
            </Button>
          </Col>
        </Row>
      )}

      {milestone.betting_options && milestone.betting_options.length > 0 && (
        <div className="space-y-2">
          <Text strong className="text-sm">투자 옵션 목록:</Text>
          <div className="space-y-1">
            {milestone.betting_options.map((option, optionIndex) => (
              <Tag
                key={optionIndex}
                closable={!disabled}
                onClose={() => !disabled && onRemoveOption(milestoneIndex, optionIndex)}
                color="blue"
                className="mb-1"
              >
                {option}
              </Tag>
            ))}
          </div>
          {milestone.betting_options.length === 0 && (
            <Text type="secondary" className="text-sm">
              아직 옵션이 추가되지 않았습니다.
            </Text>
          )}
        </div>
      )}
    </div>
  );
};

const EditProjectPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { isAuthenticated } = useAuthStore();

  // 기본 상태
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [projectData, setProjectData] = useState<Project | null>(null);
  const [hasInvestors, setHasInvestors] = useState(false);

  // 폼과 단계 관리
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);

  // 프로젝트 데이터
  const [milestones, setMilestones] = useState<ProjectMilestone[]>([]);
  const [tags, setTags] = useState<TagPair[]>([]);
  const [isPublic, setIsPublic] = useState(true);

  // AI 관련
  const [aiLoading, setAiLoading] = useState(false);
  const [aiUsageInfo, setAiUsageInfo] = useState<AIUsageInfo | null>(null);
  const [aiSuggestions, setAiSuggestions] = useState<AIMilestoneResponse | null>(null);

  // 고급 옵션
  const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);

  // 태그 입력
  const [currentTagKey, setCurrentTagKey] = useState('');
  const [currentTagValue, setCurrentTagValue] = useState('');
  const [tagInputMode, setTagInputMode] = useState<'key' | 'value'>('key');
  const valueInputRef = useRef<InputRef>(null);

  // 인증 체크 및 프로젝트 로드
  useEffect(() => {
    if (!isAuthenticated) {
      message.error('로그인이 필요합니다');
      navigate('/');
      return;
    }

    if (!id) {
      message.error('프로젝트 ID가 필요합니다');
      navigate('/dashboard');
      return;
    }

    loadProject();
    loadAIUsageInfo();
  }, [isAuthenticated, navigate, id]);

  // 프로젝트 데이터 로드
  const loadProject = async () => {
    try {
      setLoading(true);

      // 임시 목 데이터 (실제로는 API 호출)
      const mockProject: Project = {
        id: parseInt(id!),
        user_id: 1,
        title: "3년 내 카페 창업 프로젝트",
        description: "서울 강남구에서 독립 카페 창업을 목표로 하는 프로젝트입니다.",
        category: "business" as ProjectCategory,
        status: "active",
        target_date: "2027-12-31",
        budget: 50000000,
        priority: 3,
        is_public: true,
        tags: '{"location": "강남구", "type": "카페"}',
        metrics: "월 매출 500만원 이상",
        created_at: "2024-01-15T10:00:00Z",
        updated_at: "2024-01-15T10:00:00Z",
        milestones: [
          {
            title: "사업 계획서 작성",
            description: "상세한 사업 계획서를 작성하고 검토받기",
            target_date: "2025-03-31",
            order: 1,
            betting_type: "simple",
            betting_options: []
          },
          {
            title: "자금 조달",
            description: "창업 자금 5000만원 조달하기",
            target_date: "2025-12-31",
            order: 2,
            betting_type: "custom",
            betting_options: ["3000만원 조달", "5000만원 조달", "7000만원 이상 조달"]
          }
        ]
      };

      // 투자자 존재 여부 체크 (임시 데이터)
      const hasInvestorsData = mockProject.id === 1; // ID 1인 프로젝트는 투자자 있음

      setProjectData(mockProject);
      setHasInvestors(hasInvestorsData);

      // 폼 데이터 설정
      form.setFieldsValue({
        title: mockProject.title,
        description: mockProject.description,
        category: mockProject.category,
        target_date: mockProject.target_date ? dayjs(mockProject.target_date) : null,
        budget: mockProject.budget,
        priority: mockProject.priority,
        metrics: mockProject.metrics
      });

      // 마일스톤 설정
      setMilestones(mockProject.milestones || []);
      setIsPublic(mockProject.is_public);

      // 태그 파싱
      if (mockProject.tags) {
        try {
          const tagsObject = JSON.parse(mockProject.tags);
          const parsedTags = Object.entries(tagsObject).map(([key, value]) => ({
            key,
            value: value as string
          }));
          setTags(parsedTags);
        } catch (e) {
          console.error('태그 파싱 실패:', e);
        }
      }

      // 고급 옵션 표시 여부
      if (mockProject.budget || mockProject.priority !== 1 || (mockProject.tags && mockProject.tags !== '{}')) {
        setShowAdvancedOptions(true);
      }

    } catch (error) {
      console.error('프로젝트 로드 실패:', error);
      message.error('프로젝트를 불러오는데 실패했습니다');
      navigate('/dashboard');
    } finally {
      setLoading(false);
    }
  };

  // AI 사용량 정보 로드
  const loadAIUsageInfo = async () => {
    try {
      const response = await apiClient.getAIUsageInfo();
      setAiUsageInfo(response.data || null);
    } catch (error) {
      console.error('AI 사용량 정보 로드 실패:', error);
    }
  };

  // 단계 이동
  const nextStep = () => {
    if (currentStep < 2) {
      setCurrentStep(currentStep + 1);
    }
  };

  const prevStep = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  // 마일스톤 관리
  const addMilestone = () => {
    if (hasInvestors) {
      message.warning('투자자가 있는 프로젝트는 마일스톤을 추가할 수 없습니다');
      return;
    }

    if (milestones.length >= 5) {
      message.warning('최대 5개의 마일스톤까지 추가할 수 있습니다');
      return;
    }

    setMilestones([
      ...milestones,
      {
        title: '',
        description: '',
        target_date: '',
        order: milestones.length + 1,
        betting_type: 'simple',
        betting_options: []
      }
    ]);
  };

  const removeMilestone = (index: number) => {
    if (hasInvestors) {
      message.warning('투자자가 있는 프로젝트는 마일스톤을 삭제할 수 없습니다');
      return;
    }

    const newMilestones = milestones.filter((_, i) => i !== index);
    const reorderedMilestones = newMilestones.map((milestone, i) => ({
      ...milestone,
      order: i + 1
    }));
    setMilestones(reorderedMilestones);
  };

  const updateMilestone = (index: number, field: keyof ProjectMilestone, value: string | string[]) => {
    if (hasInvestors && (field === 'betting_type' || field === 'betting_options')) {
      message.warning('투자자가 있는 프로젝트는 투자 옵션을 변경할 수 없습니다');
      return;
    }

    const newMilestones = [...milestones];
    newMilestones[index] = { ...newMilestones[index], [field]: value };
    setMilestones(newMilestones);
  };

  // 마일스톤 투자 옵션 관리
  const addBettingOption = (milestoneIndex: number, option: string) => {
    if (hasInvestors) {
      message.warning('투자자가 있는 프로젝트는 투자 옵션을 변경할 수 없습니다');
      return;
    }

    const newMilestones = [...milestones];
    const currentOptions = newMilestones[milestoneIndex].betting_options || [];
    newMilestones[milestoneIndex] = {
      ...newMilestones[milestoneIndex],
      betting_options: [...currentOptions, option]
    };
    setMilestones(newMilestones);
  };

  const removeBettingOption = (milestoneIndex: number, optionIndex: number) => {
    if (hasInvestors) {
      message.warning('투자자가 있는 프로젝트는 투자 옵션을 변경할 수 없습니다');
      return;
    }

    const newMilestones = [...milestones];
    const currentOptions = newMilestones[milestoneIndex].betting_options || [];
    newMilestones[milestoneIndex] = {
      ...newMilestones[milestoneIndex],
      betting_options: currentOptions.filter((_, i) => i !== optionIndex)
    };
    setMilestones(newMilestones);
  };

  // 태그 관리
  const handleTagKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && currentTagKey.trim()) {
      setTagInputMode('value');
      // value input에 포커스 이동
      setTimeout(() => {
        if (valueInputRef.current) {
          valueInputRef.current.focus();
        }
      }, 100);
    }
  };

  const handleTagValuePress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && currentTagValue.trim()) {
      addTag();
    }
  };

  const addTag = () => {
    if (currentTagKey.trim() && currentTagValue.trim()) {
      const newTag: TagPair = {
        key: currentTagKey.trim(),
        value: currentTagValue.trim()
      };

      // 중복 키 체크
      if (tags.some(tag => tag.key === newTag.key)) {
        message.warning(`"${newTag.key}" 키가 이미 존재합니다`);
        return;
      }

      setTags([...tags, newTag]);
      setCurrentTagKey('');
      setCurrentTagValue('');
      setTagInputMode('key');
    }
  };

  const removeTag = (index: number) => {
    setTags(tags.filter((_, i) => i !== index));
  };

  // AI 제안 받기
  const handleAISuggestions = async () => {
    try {
      setAiLoading(true);

      // 필수 필드들만 먼저 검증
      const requiredFields = ['title', 'description', 'category', 'target_date'];
      const formValues = await form.validateFields(requiredFields);

      // 필수 필드 체크
      if (!formValues.title?.trim()) {
        message.warning('프로젝트 제목을 입력해주세요');
        return;
      }

      if (!formValues.description?.trim()) {
        message.warning('프로젝트 설명을 입력해주세요');
        return;
      }

      if (!formValues.category) {
        message.warning('카테고리를 선택해주세요');
        return;
      }

      if (!formValues.target_date) {
        message.warning('목표 완료일을 선택해주세요');
        return;
      }

      const formatTargetDate = (dateString?: string) => {
        if (!dateString) return undefined;
        return dayjs(dateString).format('YYYY-MM-DDTHH:mm:ss') + 'Z';
      };

      const projectData: CreateProjectWithMilestonesRequest = {
        title: formValues.title.trim(),
        description: formValues.description.trim(),
        category: formValues.category,
        target_date: formatTargetDate(formValues.target_date),
        budget: formValues.budget || 0,
        priority: formValues.priority || 1,
        is_public: isPublic,
        tags: [],
        metrics: formValues.metrics || '',
        milestones: []
      };

      const response = await apiClient.generateAIMilestones(projectData);
      setAiSuggestions(response.data || null);
      message.success('AI 제안을 받았습니다! 🤖');

    } catch (error: any) {
      console.error('AI 제안 요청 실패:', error);

      if (error.message?.includes('validation')) {
        message.error('프로젝트 정보를 모두 입력한 후 AI 제안을 받아주세요');
      } else {
        message.error('AI 제안 요청에 실패했습니다');
      }
    } finally {
      setAiLoading(false);
    }
  };

  // AI 제안 적용
  const applyAISuggestions = () => {
    if (hasInvestors) {
      message.warning('투자자가 있는 프로젝트는 AI 제안을 적용할 수 없습니다');
      return;
    }

    if (!aiSuggestions?.milestones) return;

    const aiMilestones = aiSuggestions.milestones.map((milestone, index: number) => ({
      title: milestone.title,
      description: milestone.description,
      target_date: '',
      order: milestones.length + index + 1,
      betting_type: 'simple' as const,
      betting_options: []
    }));

    setMilestones([...milestones, ...aiMilestones]);
    message.success('AI 마일스톤 제안이 적용되었습니다!');
  };

  // 프로젝트 수정 저장
  const handleSubmit = async () => {
    try {
      setSaving(true);

      const formValues = await form.validateFields();

      const formatTargetDate = (dateString?: string) => {
        if (!dateString) return undefined;
        return dayjs(dateString).format('YYYY-MM-DDTHH:mm:ss') + 'Z';
      };

      const formattedMilestones = milestones
        .filter(milestone => milestone.title && milestone.description)
        .map(milestone => ({
          ...milestone,
          target_date: formatTargetDate(milestone.target_date)
        }));

      // Tags를 JSON 문자열로 변환
      const tagsObject = tags.reduce((acc, tag) => {
        acc[tag.key] = tag.value;
        return acc;
      }, {} as Record<string, string>);

      const projectData: CreateDreamRequest = {
        ...formValues,
        target_date: formatTargetDate(formValues.target_date),
        milestones: formattedMilestones,
        is_public: isPublic,
        tags: Object.keys(tagsObject).length > 0 ? [JSON.stringify(tagsObject)] : []
      };

      // 실제로는 PUT /api/projects/:id API 호출
      console.log('프로젝트 수정 데이터:', projectData);

      message.success('프로젝트가 성공적으로 수정되었습니다! ✅');
      navigate('/dashboard');

    } catch (error: any) {
      console.error('프로젝트 수정 실패:', error);
      message.error('프로젝트 수정에 실패했습니다');
    } finally {
      setSaving(false);
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

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4">
        {/* 헤더 */}
        <div className="mb-8">
          <Button
            icon={<LeftOutlined />}
            onClick={handleBack}
            className="mb-4"
          >
            대시보드로 돌아가기
          </Button>

          <div className="text-center">
            <Title level={2}>
              <EditOutlined className="mr-3" />
              프로젝트 수정
            </Title>
            <Paragraph className="text-gray-600">
              {hasInvestors ? (
                <Alert
                  message="⚠️ 투자자가 있는 프로젝트"
                  description="이 프로젝트에는 투자자가 있어 마일스톤과 투자 옵션은 수정할 수 없습니다."
                  type="warning"
                  showIcon
                  className="mb-4"
                />
              ) : (
                "프로젝트 정보를 수정할 수 있습니다."
              )}
            </Paragraph>
          </div>
        </div>

        {/* 단계 표시 */}
        <Card className="mb-6">
          <Steps current={currentStep} className="mb-0">
            <Step
              title="프로젝트 정보"
              icon={<ProjectOutlined />}
              description="기본 정보 수정"
            />
            <Step
              title="마일스톤 설정"
              icon={<CalendarOutlined />}
              description="단계별 목표 수정"
            />
            <Step
              title="최종 검토"
              icon={<CheckCircleOutlined />}
              description="검토 및 저장"
            />
          </Steps>
        </Card>

        <Form form={form} layout="vertical">
          {/* 1단계: 프로젝트 기본 정보 */}
          {currentStep === 0 && (
            <Card title="📋 프로젝트 기본 정보 수정">
              <Row gutter={[24, 24]}>
                <Col span={24}>
                  <Form.Item
                    name="title"
                    label="프로젝트 제목"
                    rules={[{ required: true, message: '프로젝트 제목을 입력해주세요' }]}
                  >
                    <Input
                      placeholder="멋진 프로젝트 제목을 입력하세요"
                      size="large"
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="description"
                    label="프로젝트 설명"
                    rules={[{ required: true, message: '프로젝트 설명을 입력해주세요' }]}
                  >
                    <TextArea
                      rows={4}
                      placeholder="프로젝트에 대해 자세히 설명해주세요."
                    />
                  </Form.Item>
                </Col>

                <Col md={12} span={24}>
                  <Form.Item
                    name="category"
                    label="카테고리"
                    rules={[{ required: true, message: '카테고리를 선택해주세요' }]}
                  >
                    <Select size="large" placeholder="카테고리 선택">
                      <Select.Option value="career">💼 Career</Select.Option>
                      <Select.Option value="business">🚀 Business</Select.Option>
                      <Select.Option value="education">📚 Education</Select.Option>
                      <Select.Option value="personal">🌱 Personal</Select.Option>
                      <Select.Option value="life">🏡 Life</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>

                <Col md={12} span={24}>
                  <Form.Item
                    name="target_date"
                    label="목표 완료일"
                    rules={[{ required: true, message: '목표 완료일을 선택해주세요' }]}
                  >
                    <DatePicker
                      size="large"
                      style={{ width: '100%' }}
                      placeholder="완료 목표일 선택"
                      disabledDate={(current) => current && current < dayjs().endOf('day')}
                    />
                  </Form.Item>
                </Col>
              </Row>

              {/* 고급 옵션 */}
              <Divider />
              <div className="text-center mb-4">
                <Button
                  type="link"
                  icon={<SettingOutlined />}
                  onClick={() => setShowAdvancedOptions(!showAdvancedOptions)}
                >
                  고급 옵션 {showAdvancedOptions ? '접기' : '펼치기'}
                </Button>
              </div>

              {showAdvancedOptions && (
                <div className="bg-gray-50 p-4 rounded-lg">
                  <Row gutter={[24, 24]}>
                    <Col md={12} span={24}>
                      <Form.Item
                        name="budget"
                        label={
                          <Space>
                            <DollarOutlined />
                            예산 (선택사항)
                          </Space>
                        }
                      >
                        <InputNumber
                          size="large"
                          style={{ width: '100%' }}
                          placeholder="예상 예산 (원)"
                          formatter={value => `₩ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                          parser={value => value!.replace(/₩\s?|(,*)/g, '')}
                        />
                      </Form.Item>
                    </Col>

                    <Col span={24}>
                      <Form.Item
                        label={
                          <Space>
                            <TagsOutlined />
                            프로젝트 태그 (Key-Value)
                            <Tooltip title="키를 입력하고 엔터를 누른 후, 값을 입력하세요">
                              <InfoCircleOutlined />
                            </Tooltip>
                          </Space>
                        }
                      >
                        <div className="space-y-3">
                          <Row gutter={[8, 8]}>
                            <Col span={8}>
                              <Input
                                placeholder={tagInputMode === 'key' ? '키 입력 후 엔터' : '키 입력됨'}
                                value={currentTagKey}
                                onChange={(e) => setCurrentTagKey(e.target.value)}
                                onKeyPress={handleTagKeyPress}
                                disabled={tagInputMode === 'value'}
                              />
                            </Col>
                            <Col span={8}>
                              <Input
                                ref={valueInputRef}
                                placeholder={tagInputMode === 'value' ? '값 입력 후 엔터' : '먼저 키를 입력하세요'}
                                value={currentTagValue}
                                onChange={(e) => setCurrentTagValue(e.target.value)}
                                onKeyPress={handleTagValuePress}
                                disabled={tagInputMode === 'key'}
                              />
                            </Col>
                            <Col span={8}>
                              <Button
                                icon={<PlusOutlined />}
                                onClick={addTag}
                                disabled={!currentTagKey.trim() || !currentTagValue.trim()}
                              >
                                추가
                              </Button>
                            </Col>
                          </Row>

                          {tags.length > 0 && (
                            <div className="space-y-2">
                              <div className="text-sm text-gray-600">추가된 태그:</div>
                              <div className="space-x-2">
                                {tags.map((tag, index) => (
                                  <Tag
                                    key={index}
                                    closable
                                    onClose={() => removeTag(index)}
                                    color="blue"
                                  >
                                    {tag.key}: {tag.value}
                                  </Tag>
                                ))}
                              </div>
                            </div>
                          )}
                        </div>
                      </Form.Item>
                    </Col>
                  </Row>
                </div>
              )}

              <div className="text-right mt-6">
                <Button type="primary" size="large" onClick={nextStep}>
                  다음 단계 <CalendarOutlined />
                </Button>
              </div>
            </Card>
          )}

          {/* 2단계: 마일스톤 설정 */}
          {currentStep === 1 && (
            <Card title="🎯 마일스톤 설정">
              <div className="space-y-6">
                {hasInvestors && (
                  <Alert
                    message="투자자가 있는 프로젝트"
                    description="마일스톤과 투자 옵션은 투자자 보호를 위해 수정할 수 없습니다. 기본 정보만 수정 가능합니다."
                    type="warning"
                    showIcon
                    icon={<LockOutlined />}
                  />
                )}

                {/* AI 제안 섹션 */}
                {!hasInvestors && (
                  <Card size="small" className="bg-blue-50 border-blue-200">
                    <div className="flex justify-between items-center">
                      <div>
                        <Title level={5} className="mb-1">
                          <RobotOutlined className="mr-2" />
                          AI 마일스톤 제안받기
                        </Title>
                        <Paragraph className="mb-0 text-sm text-gray-600">
                          수정된 프로젝트 정보로 추가 마일스톤을 제안받을 수 있습니다
                        </Paragraph>
                      </div>
                      <Button
                        type="primary"
                        icon={<RobotOutlined />}
                        loading={aiLoading}
                        onClick={handleAISuggestions}
                      >
                        AI 제안받기
                      </Button>
                    </div>

                    {aiUsageInfo && (
                      <div className="mt-3 text-xs text-gray-500">
                        AI 사용량: {aiUsageInfo.used}/{aiUsageInfo.limit}회 사용
                      </div>
                    )}
                  </Card>
                )}

                {/* AI 제안 결과 */}
                {aiSuggestions && !hasInvestors && (
                  <Card
                    size="small"
                    title="🤖 AI 제안 마일스톤"
                    extra={
                      <Button size="small" type="primary" onClick={applyAISuggestions}>
                        제안 적용하기
                      </Button>
                    }
                  >
                    <div className="space-y-2">
                      {aiSuggestions.milestones?.map((milestone: AIMilestone, index: number) => (
                        <div key={index} className="p-2 bg-gray-50 rounded">
                          <div className="font-medium">{milestone.title}</div>
                          <div className="text-sm text-gray-600">{milestone.description}</div>
                        </div>
                      ))}
                    </div>
                  </Card>
                )}

                {/* 마일스톤 추가 버튼 */}
                {!hasInvestors && (
                  <div className="text-center">
                    <Button
                      type="dashed"
                      size="large"
                      icon={<PlusOutlined />}
                      onClick={addMilestone}
                      disabled={milestones.length >= 5}
                    >
                      마일스톤 추가 (최대 5개)
                    </Button>
                  </div>
                )}

                {/* 마일스톤 목록 */}
                <div className="space-y-4">
                  {milestones.map((milestone, index) => (
                    <Card
                      key={index}
                      size="small"
                      title={`🎯 마일스톤 ${index + 1}`}
                      extra={
                        !hasInvestors && (
                          <Button
                            size="small"
                            danger
                            onClick={() => removeMilestone(index)}
                          >
                            삭제
                          </Button>
                        )
                      }
                    >
                      <Row gutter={[12, 12]}>
                        <Col span={24}>
                          <Input
                            placeholder="마일스톤 제목"
                            value={milestone.title}
                            onChange={(e) => updateMilestone(index, 'title', e.target.value)}
                          />
                        </Col>
                        <Col span={16}>
                          <TextArea
                            rows={2}
                            placeholder="마일스톤 설명"
                            value={milestone.description}
                            onChange={(e) => updateMilestone(index, 'description', e.target.value)}
                          />
                        </Col>
                        <Col span={8}>
                          <DatePicker
                            style={{ width: '100%' }}
                            placeholder="목표일"
                            value={milestone.target_date ? dayjs(milestone.target_date) : null}
                            onChange={(date) => updateMilestone(index, 'target_date', date ? date.format('YYYY-MM-DD') : '')}
                          />
                        </Col>
                      </Row>

                      {/* 투자 옵션 설정 */}
                      <Divider className="!my-4" />
                      <div className="space-y-3">
                        <div>
                          <Typography.Text strong>💰 투자 옵션 타입</Typography.Text>
                          <Radio.Group
                            value={milestone.betting_type || 'simple'}
                            onChange={(e) => updateMilestone(index, 'betting_type', e.target.value)}
                            className="ml-3"
                            disabled={hasInvestors}
                          >
                            <Radio value="simple">📍 단순 (성공/실패)</Radio>
                            <Radio value="custom">🎯 사용자 정의</Radio>
                          </Radio.Group>
                        </div>

                        {milestone.betting_type === 'custom' && (
                          <CustomBettingOptions
                            milestoneIndex={index}
                            milestone={milestone}
                            onAddOption={addBettingOption}
                            onRemoveOption={removeBettingOption}
                            disabled={hasInvestors}
                          />
                        )}
                      </div>
                    </Card>
                  ))}

                  {milestones.length === 0 && (
                    <div className="text-center py-8 text-gray-500">
                      <Paragraph>
                        {hasInvestors ? '기존 마일스톤이 없습니다' : '마일스톤을 추가하거나 AI 제안을 받아보세요'}
                      </Paragraph>
                    </div>
                  )}
                </div>

                <div className="flex justify-between mt-6">
                  <Button size="large" onClick={prevStep}>
                    이전 단계
                  </Button>
                  <Button
                    type="primary"
                    size="large"
                    onClick={nextStep}
                  >
                    다음 단계 <CheckCircleOutlined />
                  </Button>
                </div>
              </div>
            </Card>
          )}

          {/* 3단계: 최종 검토 */}
          {currentStep === 2 && (
            <Card title="✅ 최종 검토 및 저장">
              <div className="space-y-6">
                <Alert
                  message="프로젝트 수정 전 최종 확인"
                  description="변경사항을 확인하고 저장하세요."
                  type="info"
                  showIcon
                />

                {/* 공개 설정 */}
                <Card size="small" title="🌍 공개 설정">
                  <div className="flex justify-between items-center">
                    <div>
                      <div className="font-medium">
                        {isPublic ? '🌍 공개 프로젝트' : '🔒 비공개 프로젝트'}
                      </div>
                      <div className="text-sm text-gray-600">
                        {isPublic
                          ? '모든 사용자가 보고 투자할 수 있습니다'
                          : '나만 볼 수 있고, 링크를 공유한 사람만 접근 가능합니다'
                        }
                      </div>
                    </div>
                    <Switch
                      checked={isPublic}
                      onChange={setIsPublic}
                      checkedChildren="공개"
                      unCheckedChildren="비공개"
                    />
                  </div>
                </Card>

                {/* 프로젝트 미리보기 */}
                <Card size="small" title="📋 수정된 프로젝트 미리보기">
                  <div className="space-y-4">
                    <div>
                      <div className="font-medium text-lg">{form.getFieldValue('title') || '프로젝트 제목'}</div>
                      <div className="text-gray-600 mt-1">{form.getFieldValue('description') || '프로젝트 설명'}</div>
                    </div>

                    <Row gutter={[16, 16]}>
                      <Col span={8}>
                        <div className="text-sm text-gray-500">카테고리</div>
                        <div>{form.getFieldValue('category') || '-'}</div>
                      </Col>
                      <Col span={8}>
                        <div className="text-sm text-gray-500">목표일</div>
                        <div>{form.getFieldValue('target_date') ? dayjs(form.getFieldValue('target_date')).format('YYYY-MM-DD') : '-'}</div>
                      </Col>
                      <Col span={8}>
                        <div className="text-sm text-gray-500">마일스톤</div>
                        <div>{milestones.length}개</div>
                      </Col>
                    </Row>

                    {/* 마일스톤 상세 정보 */}
                    {milestones.length > 0 && (
                      <div>
                        <div className="text-sm text-gray-500 mb-2">마일스톤 상세</div>
                        <div className="space-y-2">
                          {milestones.map((milestone, index) => (
                            <div key={index} className="p-3 bg-gray-50 rounded-lg">
                              <div className="font-medium text-sm">{milestone.title || `마일스톤 ${index + 1}`}</div>
                              <div className="text-xs text-gray-600 mt-1">
                                투자 옵션: {milestone.betting_type === 'simple' ? '📍 단순 (성공/실패)' : `🎯 사용자 정의 (${milestone.betting_options?.length || 0}개 옵션)`}
                              </div>
                              {milestone.betting_type === 'custom' && milestone.betting_options && milestone.betting_options.length > 0 && (
                                <div className="mt-2">
                                  <div className="flex flex-wrap gap-1">
                                    {milestone.betting_options.map((option, optionIndex) => (
                                      <Tag key={optionIndex} color="blue">
                                        {option}
                                      </Tag>
                                    ))}
                                  </div>
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {form.getFieldValue('budget') && (
                      <div>
                        <div className="text-sm text-gray-500">예산</div>
                        <div>₩ {form.getFieldValue('budget')?.toLocaleString()}</div>
                      </div>
                    )}

                    {tags.length > 0 && (
                      <div>
                        <div className="text-sm text-gray-500 mb-2">태그</div>
                        <div className="space-x-2">
                          {tags.map((tag, index) => (
                            <Tag key={index} color="blue">
                              {tag.key}: {tag.value}
                            </Tag>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </Card>

                <div className="flex justify-between mt-6">
                  <Button size="large" onClick={prevStep}>
                    이전 단계
                  </Button>
                  <Button
                    type="primary"
                    size="large"
                    loading={saving}
                    onClick={handleSubmit}
                  >
                    💾 변경사항 저장하기
                  </Button>
                </div>
              </div>
            </Card>
          )}
        </Form>
      </div>
    </div>
  );
};

export default EditProjectPage;
