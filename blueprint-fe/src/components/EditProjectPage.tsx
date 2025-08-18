import {
  CalendarOutlined,
  CheckCircleOutlined,
  DollarOutlined,
  InfoCircleOutlined,
  LeftOutlined,
  PlusOutlined,
  ProjectOutlined,
  SafetyOutlined,
  SettingOutlined,
  TagsOutlined,
} from "@ant-design/icons";
import type { InputRef } from "antd";
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Divider,
  Form,
  Input,
  InputNumber,
  Radio,
  Row,
  Select,
  Slider,
  Space,
  Spin,
  Steps,
  Switch,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import dayjs from "dayjs";
import React, { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { VALIDATION_MESSAGES } from "../constants/messages";
import { useNotification } from "../hooks/useNotification";
import type { ValidationRule } from "../hooks/useValidation";
import { ValidationRules } from "../hooks/useValidation";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type {
  CreateProjectWithMilestonesRequest,
  Project,
  ProjectMilestone,
  ProofType,
} from "../types";
import { FormFieldWithValidation } from "./common/FormFieldWithValidation";

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Step } = Steps;

interface CustomBettingOptionsProps {
  milestoneIndex: number;
  milestone: ProjectMilestone;
  onAddOption: (milestoneIndex: number, option: string) => void;
  onRemoveOption: (milestoneIndex: number, optionIndex: number) => void;
}

// 사용자 정의 투자 옵션 컴포넌트 (CreateProjectPage와 동일)
const CustomBettingOptions: React.FC<CustomBettingOptionsProps> = ({
  milestoneIndex,
  milestone,
  onAddOption,
  onRemoveOption,
}) => {
  const [newOption, setNewOption] = useState("");
  const { showSuccess } = useNotification();

  const validationRules: ValidationRule<string>[] = useMemo(
    () => [
      ValidationRules.required(VALIDATION_MESSAGES.BETTING_OPTION_REQUIRED),
      ValidationRules.minLength(
        2,
        VALIDATION_MESSAGES.BETTING_OPTION_MIN_LENGTH
      ),
      ValidationRules.maxLength(
        50,
        VALIDATION_MESSAGES.BETTING_OPTION_MAX_LENGTH
      ),
      ValidationRules.unique(VALIDATION_MESSAGES.DUPLICATE),
    ],
    []
  );

  const handleAddOption = () => {
    onAddOption(milestoneIndex, newOption.trim());
    setNewOption("");
    showSuccess("옵션이 추가되었습니다.");
  };

  const handleRemoveOption = (optionIndex: number) => {
    onRemoveOption(milestoneIndex, optionIndex);
    showSuccess("옵션이 삭제되었습니다.");
  };

  return (
    <div className="space-y-3">
      <div>
        <Text
          type="secondary"
          className="text-sm"
          style={{ color: "var(--text-secondary)" }}
        >
          투자자들이 선택할 수 있는 옵션들을 추가하세요. 예: "1년 내 완료", "2년
          내 완료", "3년 내 완료"
        </Text>
      </div>

      <FormFieldWithValidation
        value={newOption}
        onChange={setNewOption}
        placeholder="새 투자 옵션 입력 (예: 1년 내 완료)"
        validationRules={validationRules}
        validationContext={milestone.betting_options || []}
        actionButton={{
          text: "추가",
          icon: <PlusOutlined />,
          onClick: handleAddOption,
        }}
        onEnter={handleAddOption}
        className="mb-4"
      />

      {milestone.betting_options && milestone.betting_options.length > 0 && (
        <div className="space-y-2">
          <Text
            type="secondary"
            className="text-sm"
            style={{ color: "var(--text-secondary)" }}
          >
            현재 옵션들:
          </Text>
          <div className="flex flex-wrap gap-2">
            {milestone.betting_options.map((option: string, index: number) => (
              <Tag
                key={index}
                closable
                onClose={() => handleRemoveOption(index)}
                style={{
                  backgroundColor: "var(--bg-secondary)",
                  borderColor: "var(--border-color)",
                  color: "var(--text-primary)",
                }}
              >
                {option}
              </Tag>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

const EditProjectPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { isAuthenticated } = useAuthStore();

  // 폼과 단계 관리
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);

  // 프로젝트 데이터
  const [project, setProject] = useState<Project | null>(null);
  const [milestones, setMilestones] = useState<ProjectMilestone[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [isPublic, setIsPublic] = useState(true);

  // 고급 옵션
  const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);

  // Notification hook
  const { showError, showWarning, showSuccess } = useNotification();

  // 태그 입력
  const [currentTag, setCurrentTag] = useState("");
  const tagInputRef = useRef<InputRef>(null);

  // 인증 체크 및 프로젝트 로드
  useEffect(() => {
    if (!isAuthenticated) {
      showError("로그인이 필요합니다");
      navigate("/");
      return;
    }

    if (!id) {
      showError("프로젝트 ID가 필요합니다");
      navigate("/dashboard");
      return;
    }

    loadProject();
  }, [isAuthenticated, navigate, id]);

  // 프로젝트 데이터 로드
  const loadProject = async () => {
    if (!id) return;

    try {
      setInitialLoading(true);
      const response = await apiClient.getProject(parseInt(id));

      if (response.success && response.data) {
        const projectData = response.data;
        setProject(projectData);

        // 폼 데이터 설정
        form.setFieldsValue({
          title: projectData.title,
          description: projectData.description,
          category: projectData.category,
          target_date: projectData.target_date
            ? dayjs(projectData.target_date)
            : null,
          budget: projectData.budget || 0,
        });

        // 마일스톤 데이터 설정
        if (projectData.milestones) {
          const formattedMilestones = projectData.milestones.map(
            (milestone) => ({
              ...milestone,
              target_date: milestone.target_date
                ? dayjs(milestone.target_date).format("YYYY-MM-DD")
                : "",
              // 인증 관련 필드들 기본값 설정
              requires_proof: milestone.requires_proof ?? true,
              proof_types: milestone.proof_types || ["file", "url"],
              min_validators: milestone.min_validators ?? 3,
              min_approval_rate: milestone.min_approval_rate ?? 0.6,
              verification_deadline_days:
                milestone.verification_deadline_days ?? 3,
            })
          );
          setMilestones(formattedMilestones);
        }

        // 태그 데이터 설정
        if (projectData.tags && Array.isArray(projectData.tags)) {
          setTags(projectData.tags);
        }

        // 공개 설정
        setIsPublic(projectData.is_public ?? true);
      } else {
        showError("프로젝트를 찾을 수 없습니다");
        navigate("/dashboard");
      }
    } catch (error) {
      console.error("프로젝트 로드 실패:", error);
      showError("프로젝트 로드에 실패했습니다");
      navigate("/dashboard");
    } finally {
      setInitialLoading(false);
    }
  };

  // 단계 이동
  const nextStep = async () => {
    if (currentStep === 0) {
      try {
        const requiredFields = [
          "title",
          "description",
          "category",
          "target_date",
        ];
        await form.validateFields(requiredFields);
        setCurrentStep(currentStep + 1);
      } catch (error) {
        console.error("Form validation failed:", error);
      }
    } else if (currentStep === 1) {
      if (milestones.length === 0) {
        showError("최소 1개의 마일스톤을 추가해주세요.");
        return;
      }

      for (let i = 0; i < milestones.length; i++) {
        const milestone = milestones[i];

        if (!milestone.title?.trim()) {
          showError(`마일스톤 ${i + 1}의 제목을 입력해주세요.`);
          return;
        }

        if (milestone.betting_type === "custom") {
          if (
            !milestone.betting_options ||
            milestone.betting_options.length < 2
          ) {
            showError(
              `마일스톤 ${
                i + 1
              }의 사용자 정의 옵션은 최소 2개 이상이어야 합니다.`
            );
            return;
          }
        }
      }

      setCurrentStep(currentStep + 1);
    }
  };

  const prevStep = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  // 마일스톤 관리 함수들 (CreateProjectPage와 동일)
  const updateMilestone = (
    index: number,
    field: keyof ProjectMilestone,
    value: string | string[] | number | boolean
  ) => {
    const newMilestones = [...milestones];
    newMilestones[index] = { ...newMilestones[index], [field]: value };

    if (field === "betting_type" && value === "custom") {
      newMilestones[index].betting_options = [];
    } else if (field === "betting_type" && value === "simple") {
      newMilestones[index].betting_options = [];
    }

    setMilestones(newMilestones);
  };

  const addBettingOption = (milestoneIndex: number, option: string) => {
    const newMilestones = [...milestones];
    const currentOptions = newMilestones[milestoneIndex].betting_options || [];
    newMilestones[milestoneIndex] = {
      ...newMilestones[milestoneIndex],
      betting_options: [...currentOptions, option],
    };
    setMilestones(newMilestones);
  };

  const removeBettingOption = (milestoneIndex: number, optionIndex: number) => {
    const newMilestones = [...milestones];
    const currentOptions = newMilestones[milestoneIndex].betting_options || [];
    newMilestones[milestoneIndex] = {
      ...newMilestones[milestoneIndex],
      betting_options: currentOptions.filter(
        (_: string, i: number) => i !== optionIndex
      ),
    };
    setMilestones(newMilestones);
  };

  // 태그 관리
  const handleTagKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && currentTag.trim()) {
      addTag();
    }
  };

  const addTag = () => {
    const trimmedTag = currentTag.trim();
    if (trimmedTag) {
      if (tags.includes(trimmedTag)) {
        showWarning(`"${trimmedTag}" 태그가 이미 존재합니다`);
        return;
      }

      setTags([...tags, trimmedTag]);
      setCurrentTag("");
    }
  };

  const removeTag = (index: number) => {
    setTags(tags.filter((_, i) => i !== index));
  };

  // 프로젝트 업데이트
  const handleSubmit = async () => {
    if (!project) return;

    try {
      setLoading(true);

      const requiredFields = [
        "title",
        "description",
        "category",
        "target_date",
      ];
      const formValues = await form.validateFields(requiredFields);

      const budget = form.getFieldValue("budget") || 0;

      const formatTargetDate = (dateString?: string) => {
        if (!dateString) return undefined;
        return dayjs(dateString).format("YYYY-MM-DDTHH:mm:ss") + "Z";
      };

      // 마일스톤 데이터 준비
      const milestonesData = milestones.map((milestone, index) => ({
        id: milestone.id, // 기존 마일스톤 ID
        title: milestone.title,
        description: milestone.description,
        order: index + 1,
        target_date: milestone.target_date
          ? formatTargetDate(milestone.target_date)
          : undefined,
        betting_type: milestone.betting_type || "simple",
        betting_options: milestone.betting_options || ["success", "fail"],
        // 인증 관련 필드들
        requires_proof: milestone.requires_proof,
        proof_types: milestone.proof_types,
        min_validators: milestone.min_validators,
        min_approval_rate: milestone.min_approval_rate,
        verification_deadline_days: milestone.verification_deadline_days,
      }));

      const projectData: CreateProjectWithMilestonesRequest = {
        title: formValues.title?.trim(),
        description: formValues.description?.trim() || "",
        category: formValues.category || "personal",
        target_date: formatTargetDate(formValues.target_date),
        budget: budget,
        priority: 1,
        is_public: isPublic,
        tags: tags,
        metrics: "",
        milestones: milestonesData,
      };

      const response = await apiClient.updateProjectWithMilestones(
        project.id,
        projectData
      );

      if (response.success) {
        showSuccess("프로젝트가 성공적으로 수정되었습니다! 🎉");
        setTimeout(() => {
          navigate("/dashboard");
        }, 1500);
      } else {
        showError("프로젝트 수정에 실패했습니다");
      }
    } catch (error: unknown) {
      console.error("프로젝트 수정 실패:", error);
      showError("프로젝트 수정에 실패했습니다");
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    navigate("/dashboard");
  };

  if (initialLoading) {
    return (
      <div
        className="min-h-screen flex items-center justify-center"
        style={{ backgroundColor: "var(--bg-primary)" }}
      >
        <Spin size="large" />
      </div>
    );
  }

  if (!project) {
    return (
      <div
        className="min-h-screen flex items-center justify-center"
        style={{ backgroundColor: "var(--bg-primary)" }}
      >
        <div className="text-center">
          <Title level={3} style={{ color: "var(--text-primary)" }}>
            프로젝트를 찾을 수 없습니다
          </Title>
          <Button type="primary" onClick={handleBack}>
            대시보드로 돌아가기
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div
      className="min-h-screen py-8"
      style={{ backgroundColor: "var(--bg-primary)" }}
    >
      <div className="max-w-4xl mx-auto px-4">
        {/* 헤더 */}
        <div className="mb-8">
          <Button
            icon={<LeftOutlined />}
            onClick={handleBack}
            className="mb-4"
            style={{
              backgroundColor: "var(--bg-secondary)",
              borderColor: "var(--border-color)",
              color: "var(--text-primary)",
            }}
          >
            대시보드로 돌아가기
          </Button>

          <div className="text-center">
            <Title level={2} style={{ color: "var(--text-primary)" }}>
              <ProjectOutlined className="mr-3" />
              프로젝트 수정하기
            </Title>
            <Paragraph style={{ color: "var(--text-secondary)" }}>
              프로젝트 정보와 마일스톤을 수정할 수 있습니다.
            </Paragraph>
          </div>
        </div>

        {/* 단계 표시 */}
        <Card
          className="mb-6"
          style={{
            backgroundColor: "var(--bg-secondary)",
            borderColor: "var(--border-color)",
          }}
        >
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
            <Card
              title="📋 프로젝트 기본 정보"
              style={{
                backgroundColor: "var(--bg-secondary)",
                borderColor: "var(--border-color)",
              }}
            >
              <Row gutter={[24, 24]}>
                <Col span={24}>
                  <Form.Item
                    name="title"
                    label="프로젝트 제목"
                    rules={[
                      {
                        required: true,
                        message: "프로젝트 제목을 입력해주세요",
                      },
                      {
                        min: 3,
                        message: "프로젝트 제목은 최소 3글자 이상이어야 합니다",
                      },
                      {
                        max: 200,
                        message: "프로젝트 제목은 최대 200글자까지 가능합니다",
                      },
                    ]}
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
                    rules={[
                      {
                        required: true,
                        message: "프로젝트 설명을 입력해주세요",
                      },
                    ]}
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
                    rules={[
                      { required: true, message: "카테고리를 선택해주세요" },
                    ]}
                  >
                    <Select size="large" placeholder="카테고리 선택">
                      <Select.Option value="career">💼 Career</Select.Option>
                      <Select.Option value="business">
                        🚀 Business
                      </Select.Option>
                      <Select.Option value="education">
                        📚 Education
                      </Select.Option>
                      <Select.Option value="personal">
                        🌱 Personal
                      </Select.Option>
                      <Select.Option value="life">🏡 Life</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>

                <Col md={12} span={24}>
                  <Form.Item
                    name="target_date"
                    label="목표 완료일"
                    rules={[
                      { required: true, message: "목표 완료일을 선택해주세요" },
                    ]}
                  >
                    <DatePicker
                      size="large"
                      style={{ width: "100%" }}
                      placeholder="완료 목표일 선택"
                      disabledDate={(current) =>
                        current && current < dayjs().endOf("day")
                      }
                    />
                  </Form.Item>
                </Col>
              </Row>

              {/* 고급 옵션 */}
              <Divider style={{ borderTopColor: "var(--border-color)" }} />
              <div className="text-center mb-4">
                <Button
                  type="link"
                  icon={<SettingOutlined />}
                  onClick={() => setShowAdvancedOptions(!showAdvancedOptions)}
                  style={{ color: "var(--blue)" }}
                >
                  고급 옵션 {showAdvancedOptions ? "접기" : "펼치기"}
                </Button>
              </div>

              {showAdvancedOptions && (
                <div
                  className="p-4 rounded-lg"
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    border: "1px solid var(--border-color)",
                  }}
                >
                  <Row gutter={[24, 24]}>
                    <Col md={12} span={24}>
                      <Form.Item
                        name="budget"
                        label={
                          <Space style={{ color: "var(--text-primary)" }}>
                            <DollarOutlined />
                            예산 (선택사항)
                          </Space>
                        }
                      >
                        <InputNumber
                          size="large"
                          style={{
                            width: "100%",
                            backgroundColor: "var(--bg-primary)",
                            borderColor: "var(--border-color)",
                            color: "var(--text-primary)",
                          }}
                          placeholder="예상 예산 (원)"
                          formatter={(value) =>
                            `₩ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                          }
                          parser={(value) => value!.replace(/₩\s?|(,*)/g, "")}
                        />
                      </Form.Item>
                    </Col>

                    <Col span={24}>
                      <Form.Item
                        label={
                          <Space style={{ color: "var(--text-primary)" }}>
                            <TagsOutlined />
                            프로젝트 태그
                            <Tooltip title="태그를 입력하고 엔터를 누르거나 추가 버튼을 클릭하세요">
                              <InfoCircleOutlined />
                            </Tooltip>
                          </Space>
                        }
                      >
                        <div className="space-y-3">
                          <Row gutter={[8, 8]}>
                            <Col span={16}>
                              <Input
                                ref={tagInputRef}
                                placeholder="태그 입력 후 엔터"
                                value={currentTag}
                                onChange={(e) => setCurrentTag(e.target.value)}
                                onKeyPress={handleTagKeyPress}
                                style={{
                                  backgroundColor: "var(--bg-primary)",
                                  borderColor: "var(--border-color)",
                                  color: "var(--text-primary)",
                                }}
                              />
                            </Col>
                            <Col span={8}>
                              <Button
                                icon={<PlusOutlined />}
                                onClick={addTag}
                                disabled={!currentTag.trim()}
                                style={{
                                  backgroundColor: "var(--bg-secondary)",
                                  borderColor: "var(--border-color)",
                                  color: "var(--text-primary)",
                                }}
                              >
                                추가
                              </Button>
                            </Col>
                          </Row>

                          {tags.length > 0 && (
                            <div className="space-y-2">
                              <div
                                className="text-sm"
                                style={{ color: "var(--text-secondary)" }}
                              >
                                추가된 태그:
                              </div>
                              <div className="space-x-2">
                                {tags.map((tag, index) => (
                                  <Tag
                                    key={index}
                                    closable
                                    onClose={() => removeTag(index)}
                                    color="blue"
                                  >
                                    {tag}
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
                <Button
                  type="primary"
                  size="large"
                  onClick={nextStep}
                  style={{
                    background:
                      "linear-gradient(135deg, var(--blue) 0%, #9333ea 100%)",
                    borderColor: "var(--blue)",
                  }}
                >
                  다음 단계 <CalendarOutlined />
                </Button>
              </div>
            </Card>
          )}

          {/* 2단계: 마일스톤 설정 */}
          {currentStep === 1 && (
            <Card
              title="🎯 마일스톤 설정"
              style={{
                backgroundColor: "var(--bg-secondary)",
                borderColor: "var(--border-color)",
              }}
            >
              <div className="space-y-6">
                <Alert
                  message="마일스톤 수정 안내"
                  description="기존 마일스톤의 투자 데이터가 있는 경우, 일부 설정 변경이 제한될 수 있습니다."
                  type="info"
                  showIcon
                />

                {/* 마일스톤 목록 */}
                <div className="space-y-4">
                  {milestones.map((milestone, index) => (
                    <Card
                      key={milestone.id || index}
                      size="small"
                      title={`🎯 마일스톤 ${index + 1}`}
                    >
                      <Row gutter={[12, 12]}>
                        <Col span={24}>
                          <FormFieldWithValidation
                            value={milestone.title}
                            onChange={(value) =>
                              updateMilestone(index, "title", value)
                            }
                            placeholder="마일스톤 제목"
                            validationRules={[
                              ValidationRules.required(
                                VALIDATION_MESSAGES.MILESTONE_TITLE_REQUIRED
                              ),
                              ValidationRules.minLength(
                                2,
                                VALIDATION_MESSAGES.MILESTONE_TITLE_MIN_LENGTH
                              ),
                              ValidationRules.maxLength(
                                100,
                                VALIDATION_MESSAGES.MILESTONE_TITLE_MAX_LENGTH
                              ),
                            ]}
                            inputSpan={24}
                            className="milestone-title-field"
                          />
                        </Col>
                        <Col span={16}>
                          <TextArea
                            rows={2}
                            placeholder="마일스톤 설명"
                            value={milestone.description}
                            onChange={(e) =>
                              updateMilestone(
                                index,
                                "description",
                                e.target.value
                              )
                            }
                          />
                        </Col>
                        <Col span={8}>
                          <DatePicker
                            style={{ width: "100%" }}
                            placeholder="목표일 선택"
                            value={
                              milestone.target_date
                                ? dayjs(milestone.target_date)
                                : null
                            }
                            onChange={(date) => {
                              updateMilestone(
                                index,
                                "target_date",
                                date ? date.format("YYYY-MM-DD") : ""
                              );
                            }}
                            disabledDate={(current) =>
                              current && current < dayjs().endOf("day")
                            }
                          />
                        </Col>
                      </Row>

                      {/* 투자 옵션 설정 */}
                      <Divider
                        className="!my-4"
                        style={{ borderTopColor: "var(--border-color)" }}
                      />
                      <div className="space-y-3">
                        <div>
                          <Typography.Text strong>
                            💰 투자 옵션 타입
                          </Typography.Text>
                          <Radio.Group
                            value={milestone.betting_type || "simple"}
                            onChange={(e) =>
                              updateMilestone(
                                index,
                                "betting_type",
                                e.target.value
                              )
                            }
                            className="ml-3"
                          >
                            <Radio value="simple">📍 단순 (성공/실패)</Radio>
                            <Radio value="custom">🎯 사용자 정의</Radio>
                          </Radio.Group>
                        </div>

                        {milestone.betting_type === "custom" && (
                          <CustomBettingOptions
                            milestoneIndex={index}
                            milestone={milestone}
                            onAddOption={addBettingOption}
                            onRemoveOption={removeBettingOption}
                          />
                        )}
                      </div>

                      {/* 🔍 인증 방법 설정 */}
                      <Divider
                        className="!my-4"
                        style={{ borderTopColor: "var(--border-color)" }}
                      />
                      <div className="space-y-4">
                        <div>
                          <Typography.Text strong>
                            <SafetyOutlined className="mr-2" />
                            🔍 인증 방법 설정
                          </Typography.Text>
                          <Typography.Text
                            type="secondary"
                            className="block text-sm mt-1"
                            style={{ color: "var(--text-secondary)" }}
                          >
                            마일스톤 달성 시 어떤 방식으로 증명할지 설정하세요
                          </Typography.Text>
                        </div>

                        {/* 증거 제출 필요 여부 */}
                        <div className="flex items-center justify-between">
                          <div>
                            <Typography.Text>증거 제출 필요</Typography.Text>
                            <Typography.Text
                              type="secondary"
                              className="block text-xs"
                              style={{ color: "var(--text-secondary)" }}
                            >
                              완료 시 증명 자료를 제출하도록 요구
                            </Typography.Text>
                          </div>
                          <Switch
                            checked={milestone.requires_proof !== false}
                            onChange={(checked) =>
                              updateMilestone(index, "requires_proof", checked)
                            }
                            checkedChildren="필요"
                            unCheckedChildren="불필요"
                          />
                        </div>

                        {milestone.requires_proof !== false && (
                          <>
                            {/* 허용되는 증거 타입 */}
                            <div>
                              <Typography.Text className="block mb-2">
                                허용되는 증거 타입
                              </Typography.Text>
                              <Checkbox.Group
                                value={milestone.proof_types || ["file", "url"]}
                                onChange={(values) =>
                                  updateMilestone(
                                    index,
                                    "proof_types",
                                    values as ProofType[]
                                  )
                                }
                                className="w-full"
                              >
                                <Row gutter={[8, 8]}>
                                  <Col span={12}>
                                    <Checkbox value="file">
                                      📁 파일 업로드
                                    </Checkbox>
                                  </Col>
                                  <Col span={12}>
                                    <Checkbox value="url">🔗 웹 링크</Checkbox>
                                  </Col>
                                  <Col span={12}>
                                    <Checkbox value="screenshot">
                                      📸 스크린샷
                                    </Checkbox>
                                  </Col>
                                  <Col span={12}>
                                    <Checkbox value="video">🎥 영상</Checkbox>
                                  </Col>
                                  <Col span={12}>
                                    <Checkbox value="text">
                                      📝 텍스트 설명
                                    </Checkbox>
                                  </Col>
                                  <Col span={12}>
                                    <Checkbox value="certificate">
                                      🏆 인증서
                                    </Checkbox>
                                  </Col>
                                  <Col span={12}>
                                    <Checkbox value="api">🔌 API 연동</Checkbox>
                                  </Col>
                                </Row>
                              </Checkbox.Group>
                            </div>

                            {/* 검증 설정 */}
                            <div>
                              <Typography.Text className="block mb-3">
                                검증 조건 설정
                              </Typography.Text>

                              <div className="space-y-3">
                                {/* 최소 검증인 수 */}
                                <div>
                                  <div className="flex justify-between items-center mb-2">
                                    <Typography.Text className="text-sm">
                                      최소 검증인 수:{" "}
                                      {milestone.min_validators || 3}명
                                    </Typography.Text>
                                  </div>
                                  <Slider
                                    min={1}
                                    max={10}
                                    value={milestone.min_validators || 3}
                                    onChange={(value) =>
                                      updateMilestone(
                                        index,
                                        "min_validators",
                                        value
                                      )
                                    }
                                    marks={{
                                      1: "1명",
                                      3: "3명",
                                      5: "5명",
                                      10: "10명",
                                    }}
                                  />
                                </div>

                                {/* 최소 승인률 */}
                                <div>
                                  <div className="flex justify-between items-center mb-2">
                                    <Typography.Text className="text-sm">
                                      최소 승인률:{" "}
                                      {Math.round(
                                        (milestone.min_approval_rate || 0.6) *
                                          100
                                      )}
                                      %
                                    </Typography.Text>
                                  </div>
                                  <Slider
                                    min={0.5}
                                    max={1.0}
                                    step={0.1}
                                    value={milestone.min_approval_rate || 0.6}
                                    onChange={(value) =>
                                      updateMilestone(
                                        index,
                                        "min_approval_rate",
                                        value
                                      )
                                    }
                                    marks={{
                                      0.5: "50%",
                                      0.6: "60%",
                                      0.8: "80%",
                                      1.0: "100%",
                                    }}
                                  />
                                </div>

                                {/* 검증 기간 */}
                                <div>
                                  <Typography.Text className="block mb-2 text-sm">
                                    검증 기간 (일)
                                  </Typography.Text>
                                  <InputNumber
                                    min={1}
                                    max={14}
                                    value={
                                      milestone.verification_deadline_days || 3
                                    }
                                    onChange={(value) =>
                                      updateMilestone(
                                        index,
                                        "verification_deadline_days",
                                        value || 3
                                      )
                                    }
                                    addonAfter="일"
                                    className="w-full"
                                    placeholder="3"
                                  />
                                  <Typography.Text
                                    type="secondary"
                                    className="text-xs block mt-1"
                                    style={{ color: "var(--text-secondary)" }}
                                  >
                                    증거 제출 후 검증인들이 검토할 수 있는 기간
                                  </Typography.Text>
                                </div>
                              </div>
                            </div>
                          </>
                        )}
                      </div>
                    </Card>
                  ))}

                  {milestones.length === 0 && (
                    <div className="text-center py-8 text-gray-500">
                      <Paragraph>마일스톤이 없습니다.</Paragraph>
                    </div>
                  )}
                </div>

                <div className="flex justify-between mt-6">
                  <Button
                    size="large"
                    onClick={prevStep}
                    style={{
                      backgroundColor: "var(--bg-secondary)",
                      borderColor: "var(--border-color)",
                      color: "var(--text-primary)",
                    }}
                  >
                    이전 단계
                  </Button>
                  <Button
                    type="primary"
                    size="large"
                    onClick={nextStep}
                    disabled={milestones.length === 0}
                    style={{
                      background:
                        "linear-gradient(135deg, var(--blue) 0%, #9333ea 100%)",
                      borderColor: "var(--blue)",
                    }}
                  >
                    다음 단계 <CheckCircleOutlined />
                  </Button>
                </div>
              </div>
            </Card>
          )}

          {/* 3단계: 최종 검토 */}
          {currentStep === 2 && (
            <Card
              title="✅ 최종 검토 및 저장"
              style={{
                backgroundColor: "var(--bg-secondary)",
                borderColor: "var(--border-color)",
              }}
            >
              <div className="space-y-6">
                <Alert
                  message="프로젝트 수정 전 최종 확인"
                  description="아래 정보를 확인하고 프로젝트를 수정하세요."
                  type="info"
                  showIcon
                />

                {/* 공개 설정 */}
                <Card
                  size="small"
                  title="🌍 공개 설정"
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    borderColor: "var(--border-color)",
                  }}
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <div
                        className="font-medium"
                        style={{ color: "var(--text-primary)" }}
                      >
                        {isPublic ? "🌍 공개 프로젝트" : "🔒 비공개 프로젝트"}
                      </div>
                      <div
                        className="text-sm"
                        style={{ color: "var(--text-secondary)" }}
                      >
                        {isPublic
                          ? "모든 사용자가 보고 투자할 수 있습니다"
                          : "나만 볼 수 있고, 링크를 공유한 사람만 접근 가능합니다"}
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
                <Card
                  size="small"
                  title="📋 프로젝트 미리보기"
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    borderColor: "var(--border-color)",
                  }}
                >
                  <div className="space-y-4">
                    <div>
                      <div
                        style={{
                          fontWeight: "500",
                          fontSize: "18px",
                          color: "var(--text-primary)",
                        }}
                      >
                        {form.getFieldValue("title") || "프로젝트 제목"}
                      </div>
                      <div
                        style={{
                          color: "var(--text-secondary)",
                          marginTop: "4px",
                        }}
                      >
                        {form.getFieldValue("description") || "프로젝트 설명"}
                      </div>
                    </div>

                    <Row gutter={[16, 16]}>
                      <Col span={8}>
                        <div
                          style={{
                            fontSize: "14px",
                            color: "var(--text-secondary)",
                          }}
                        >
                          카테고리
                        </div>
                        <div style={{ color: "var(--text-primary)" }}>
                          {form.getFieldValue("category") || "-"}
                        </div>
                      </Col>
                      <Col span={8}>
                        <div
                          style={{
                            fontSize: "14px",
                            color: "var(--text-secondary)",
                          }}
                        >
                          목표일
                        </div>
                        <div style={{ color: "var(--text-primary)" }}>
                          {form.getFieldValue("target_date")
                            ? dayjs(form.getFieldValue("target_date")).format(
                                "YYYY-MM-DD"
                              )
                            : "-"}
                        </div>
                      </Col>
                      <Col span={8}>
                        <div
                          style={{
                            fontSize: "14px",
                            color: "var(--text-secondary)",
                          }}
                        >
                          마일스톤
                        </div>
                        <div style={{ color: "var(--text-primary)" }}>
                          {milestones.length}개
                        </div>
                      </Col>
                    </Row>

                    {/* 마일스톤 상세 정보 */}
                    {milestones.length > 0 && (
                      <div>
                        <div
                          style={{
                            fontSize: "14px",
                            color: "var(--text-secondary)",
                            marginBottom: "8px",
                          }}
                        >
                          마일스톤 상세
                        </div>
                        <div className="space-y-2">
                          {milestones.map((milestone, index) => (
                            <div
                              key={index}
                              style={{
                                padding: "12px",
                                backgroundColor: "var(--bg-tertiary)",
                                borderRadius: "8px",
                                border: "1px solid var(--border-color)",
                              }}
                            >
                              <div
                                style={{
                                  fontWeight: "500",
                                  fontSize: "14px",
                                  color: "var(--text-primary)",
                                }}
                              >
                                {milestone.title || `마일스톤 ${index + 1}`}
                              </div>
                              <div
                                style={{
                                  fontSize: "12px",
                                  color: "var(--text-secondary)",
                                  marginTop: "4px",
                                }}
                              >
                                투자 옵션:{" "}
                                {milestone.betting_type === "simple"
                                  ? "📍 단순 (성공/실패)"
                                  : `🎯 사용자 정의 (${
                                      milestone.betting_options?.length || 0
                                    }개 옵션)`}
                              </div>
                              <div
                                style={{
                                  fontSize: "12px",
                                  color: "var(--text-secondary)",
                                  marginTop: "4px",
                                }}
                              >
                                🔍 인증 방법:{" "}
                                {milestone.requires_proof === false
                                  ? "증거 제출 불필요"
                                  : `증거 필요 (${
                                      milestone.proof_types?.length || 2
                                    }개 타입)`}
                                {milestone.requires_proof !== false && (
                                  <span className="ml-2">
                                    · 검증인 {milestone.min_validators || 3}명
                                    이상 · 승인률{" "}
                                    {Math.round(
                                      (milestone.min_approval_rate || 0.6) * 100
                                    )}
                                    % 이상
                                  </span>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {form.getFieldValue("budget") && (
                      <div>
                        <div
                          className="text-sm"
                          style={{ color: "var(--text-secondary)" }}
                        >
                          예산
                        </div>
                        <div style={{ color: "var(--text-primary)" }}>
                          ₩ {form.getFieldValue("budget")?.toLocaleString()}
                        </div>
                      </div>
                    )}

                    {tags.length > 0 && (
                      <div>
                        <div
                          style={{
                            fontSize: "14px",
                            color: "var(--text-secondary)",
                            marginBottom: "8px",
                          }}
                        >
                          태그
                        </div>
                        <div className="flex flex-wrap gap-2">
                          {tags.map((tag, index) => (
                            <Tag key={index} color="blue">
                              {tag}
                            </Tag>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </Card>

                <div className="flex justify-between mt-6">
                  <Button
                    size="large"
                    onClick={prevStep}
                    style={{
                      backgroundColor: "var(--bg-secondary)",
                      borderColor: "var(--border-color)",
                      color: "var(--text-primary)",
                    }}
                  >
                    이전 단계
                  </Button>
                  <Button
                    type="primary"
                    size="large"
                    loading={loading}
                    onClick={handleSubmit}
                    style={{
                      background:
                        "linear-gradient(135deg, var(--blue) 0%, #9333ea 100%)",
                      borderColor: "var(--blue)",
                      minWidth: "200px",
                      height: "48px",
                    }}
                  >
                    💾 프로젝트 수정하기
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
