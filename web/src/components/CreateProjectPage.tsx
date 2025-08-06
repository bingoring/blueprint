import {
  CalendarOutlined,
  CheckCircleOutlined,
  DollarOutlined,
  InfoCircleOutlined,
  LeftOutlined,
  PlusOutlined,
  ProjectOutlined,
  RobotOutlined,
  SettingOutlined,
  TagsOutlined,
} from "@ant-design/icons";
import type { InputRef } from "antd";
import {
  Alert,
  Button,
  Card,
  Col,
  DatePicker,
  Divider,
  Form,
  Input,
  InputNumber,
  Radio,
  Row,
  Select,
  Space,
  Steps,
  Switch,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import dayjs from "dayjs";
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate } from "react-router-dom";
import { MessageHelpers, VALIDATION_MESSAGES } from "../constants/messages";
import { useNotification } from "../hooks/useNotification";
import type { ValidationRule } from "../hooks/useValidation";
import { ValidationRules } from "../hooks/useValidation";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type {
  AIMilestone,
  AIMilestoneResponse,
  AIUsageInfo,
  CreateProjectWithMilestonesRequest,
  ProjectMilestone,
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

// 사용자 정의 투자 옵션 컴포넌트
const CustomBettingOptions: React.FC<CustomBettingOptionsProps> = ({
  milestoneIndex,
  milestone,
  onAddOption,
  onRemoveOption,
}) => {
  const [newOption, setNewOption] = useState("");
  const { showSuccess } = useNotification();

  // Validation 규칙 정의 (newOption과 milestone.betting_options 변경 시에만 재계산)
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

      {/* 기존 옵션들 표시 */}
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

const CreateProjectPage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();

  // 폼과 단계 관리
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);

  // 프로젝트 데이터
  const [milestones, setMilestones] = useState<ProjectMilestone[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [isPublic, setIsPublic] = useState(true);

  // AI 관련
  const [aiLoading, setAiLoading] = useState(false);
  const [aiUsageInfo, setAiUsageInfo] = useState<AIUsageInfo | null>(null);
  const [aiSuggestions, setAiSuggestions] =
    useState<AIMilestoneResponse | null>(null);

  // 고급 옵션
  const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);

  // Notification hook
  const { showError, showWarning, showSuccess } = useNotification();

  // 태그 입력
  const [currentTag, setCurrentTag] = useState("");
  const tagInputRef = useRef<InputRef>(null);

  // 로켓 발사 애니메이션
  const [isHovered, setIsHovered] = useState(false);
  const [isLaunching, setIsLaunching] = useState(false);
  const [showLaunchSuccess, setShowLaunchSuccess] = useState(false);

  // 인증 체크
  useEffect(() => {
    if (!isAuthenticated) {
      showError("로그인이 필요합니다");
      navigate("/");
      return;
    }
    loadAIUsageInfo();
  }, [isAuthenticated, navigate]);

  // AI 사용량 정보 로드
  const loadAIUsageInfo = async () => {
    try {
      const response = await apiClient.getAIUsageInfo();
      setAiUsageInfo(response.data || null);
    } catch (error) {
      console.error("AI 사용량 정보 로드 실패:", error);
    }
  };

  // 단계 이동
  const nextStep = async () => {
    if (currentStep === 0) {
      try {
        // 1단계: 프로젝트 기본 정보 validation
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
        // Ant Design Form이 자동으로 에러를 표시하므로 추가 처리 불필요
      }
    } else if (currentStep === 1) {
      // 2단계: 마일스톤 설정 validation
      if (milestones.length === 0) {
        showError("최소 1개의 마일스톤을 추가해주세요.");
        return;
      }

      // 각 마일스톤 validation 체크
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

  // 마일스톤 날짜 선택 제약 함수
  const getMilestoneDisabledDate = useCallback(
    (milestoneIndex: number) => {
      return (current: dayjs.Dayjs | null) => {
        if (!current) return false;

        const today = dayjs().startOf("day");
        const projectEndDate = form.getFieldValue("target_date");

        // 오늘 이전 날짜는 선택 불가
        if (current.isBefore(today)) {
          return true;
        }

        // 프로젝트 완료일이 설정되어 있고, 그 이후 날짜는 선택 불가
        if (
          projectEndDate &&
          current.isAfter(dayjs(projectEndDate).endOf("day"))
        ) {
          return true;
        }

        // 다른 마일스톤에서 이미 선택된 날짜는 선택 불가
        const selectedDates = milestones
          .map((milestone, index) => {
            if (index !== milestoneIndex && milestone.target_date) {
              return dayjs(milestone.target_date).format("YYYY-MM-DD");
            }
            return null;
          })
          .filter(Boolean);

        if (selectedDates.includes(current.format("YYYY-MM-DD"))) {
          return true;
        }

        return false;
      };
    },
    [form, milestones]
  );

  // 마지막 마일스톤과 프로젝트 완료일 동기화 함수
  const syncLastMilestoneWithProjectEnd = useCallback(
    (milestonesArray: ProjectMilestone[]) => {
      const projectEndDate = form.getFieldValue("target_date");
      if (milestonesArray.length > 0 && projectEndDate) {
        const updatedMilestones = [...milestonesArray];
        const lastIndex = updatedMilestones.length - 1;
        updatedMilestones[lastIndex] = {
          ...updatedMilestones[lastIndex],
          target_date: dayjs(projectEndDate).format("YYYY-MM-DD"),
        };
        return updatedMilestones;
      }
      return milestonesArray;
    },
    [form]
  );

  // 프로젝트 완료일 변경 시 마지막 마일스톤 자동 업데이트
  const handleProjectDateChange = useCallback(
    (date: dayjs.Dayjs | null) => {
      if (date && milestones.length > 0) {
        const updatedMilestones = [...milestones];
        const lastIndex = updatedMilestones.length - 1;
        updatedMilestones[lastIndex] = {
          ...updatedMilestones[lastIndex],
          target_date: date.format("YYYY-MM-DD"),
        };
        setMilestones(updatedMilestones);
      }
    },
    [milestones]
  );

  // 마일스톤 관리
  const addMilestone = () => {
    if (milestones.length >= 5) {
      showWarning("최대 5개의 마일스톤까지 추가할 수 있습니다");
      return;
    }

    const newMilestones = [
      ...milestones,
      {
        title: "",
        description: "",
        target_date: "",
        order: milestones.length + 1,
        betting_type: "simple" as const,
        betting_options: [], // 기본값 제거 - 빈 배열로 시작
      },
    ];

    // 마지막 마일스톤을 프로젝트 완료일과 동기화
    const syncedMilestones = syncLastMilestoneWithProjectEnd(newMilestones);
    setMilestones(syncedMilestones);
  };

  const removeMilestone = (index: number) => {
    const newMilestones = milestones.filter((_, i) => i !== index);
    const reorderedMilestones = newMilestones.map((milestone, i) => ({
      ...milestone,
      order: i + 1,
    }));

    // 마지막 마일스톤을 프로젝트 완료일과 동기화
    const syncedMilestones =
      syncLastMilestoneWithProjectEnd(reorderedMilestones);
    setMilestones(syncedMilestones);
  };

  const updateMilestone = (
    index: number,
    field: keyof ProjectMilestone,
    value: string | string[]
  ) => {
    const newMilestones = [...milestones];
    newMilestones[index] = { ...newMilestones[index], [field]: value };

    // betting_type이 custom으로 변경될 때 빈 배열로 시작 (기본값 없음)
    if (field === "betting_type" && value === "custom") {
      newMilestones[index].betting_options = [];
    }
    // betting_type이 simple로 변경될 때도 빈 배열로 초기화
    else if (field === "betting_type" && value === "simple") {
      newMilestones[index].betting_options = [];
    }

    setMilestones(newMilestones);
  };

  // 마일스톤 투자 옵션 관리
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
      // 중복 태그 체크
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

  // AI 제안 받기
  const handleAISuggestions = async () => {
    try {
      setAiLoading(true);

      // 필수 필드들만 먼저 검증
      const requiredFields = [
        "title",
        "description",
        "category",
        "target_date",
      ];
      const formValues = await form.validateFields(requiredFields);

      // 필수 필드 체크
      if (!formValues.title?.trim()) {
        showWarning("프로젝트 제목을 입력해주세요");
        return;
      }

      if (!formValues.description?.trim()) {
        showWarning("프로젝트 설명을 입력해주세요");
        return;
      }

      if (!formValues.category) {
        showWarning("카테고리를 선택해주세요");
        return;
      }

      if (!formValues.target_date) {
        showWarning("목표 완료일을 선택해주세요");
        return;
      }

      const formatTargetDate = (dateString?: string) => {
        if (!dateString) return undefined;
        return dayjs(dateString).format("YYYY-MM-DDTHH:mm:ss") + "Z";
      };

      const projectData: CreateProjectWithMilestonesRequest = {
        title: formValues.title.trim(),
        description: formValues.description.trim(),
        category: formValues.category,
        target_date: formatTargetDate(formValues.target_date),
        budget: formValues.budget || 0,
        priority: formValues.priority || 1, // 기본값 설정
        is_public: isPublic,
        tags: [], // AI 제안 시에는 빈 배열
        metrics: formValues.metrics || "",
        milestones: [],
      };

      const response = await apiClient.generateAIMilestones(projectData);
      setAiSuggestions(response.data || null);
      showSuccess("AI 제안을 받았습니다! 🤖");
    } catch (error: unknown) {
      console.error("AI 제안 요청 실패:", error);

      if (error instanceof Error && error.message?.includes("validation")) {
        showError("프로젝트 정보를 모두 입력한 후 AI 제안을 받아주세요");
      } else {
        showError("AI 제안 요청에 실패했습니다");
      }
    } finally {
      setAiLoading(false);
    }
  };

  // AI 제안 적용
  const applyAISuggestions = () => {
    if (!aiSuggestions?.milestones) return;

    const aiMilestones = aiSuggestions.milestones.map(
      (milestone, index: number) => ({
        title: milestone.title,
        description: milestone.description,
        target_date: "",
        order: index + 1,
        betting_type: "simple" as const,
        betting_options: [],
      })
    );

    setMilestones(aiMilestones);
    showSuccess("AI 마일스톤 제안이 적용되었습니다!");
  };

  // 로켓 발사 애니메이션과 함께 프로젝트 생성
  const handleLaunchProject = async () => {
    if (loading || isLaunching) return; // 중복 클릭 방지

    // 발사 애니메이션 시작 (hover에서 preparing 상태를 이어받음)
    setIsLaunching(true);
    setShowLaunchSuccess(false);

    // 0.8초 후 실제 제출 시작 (로켓이 화면을 벗어나기 전에)
    setTimeout(async () => {
      await handleSubmit();
    }, 800);
  };

  // 프로젝트 생성
  const handleSubmit = async () => {
    try {
      setLoading(true);

      // 필수 필드들만 검증 (Form에 정의된 것들)
      const requiredFields = [
        "title",
        "description",
        "category",
        "target_date",
      ];
      const formValues = await form.validateFields(requiredFields);

      // 선택적 필드들 직접 가져오기
      const budget = form.getFieldValue("budget") || 0;

      console.log("🔍 Debug formValues:", formValues);
      console.log("🔍 Debug budget:", budget);
      console.log("🔍 Debug tags:", tags);
      console.log("🔍 Debug milestones:", milestones);

      const formatTargetDate = (dateString?: string) => {
        if (!dateString) return undefined;
        return dayjs(dateString).format("YYYY-MM-DDTHH:mm:ss") + "Z";
      };

      const formattedMilestones = milestones
        .filter((milestone) => milestone.title && milestone.description)
        .map((milestone) => ({
          ...milestone,
          target_date: formatTargetDate(milestone.target_date),
        }));

      // Tags는 이미 string 배열이므로 그대로 사용

      const projectData: CreateProjectWithMilestonesRequest = {
        title: formValues.title?.trim(),
        description: formValues.description?.trim() || "",
        category: formValues.category || "personal",
        target_date: formatTargetDate(formValues.target_date),
        budget: budget,
        priority: 1, // 기본값 (Form 필드 없음)
        is_public: isPublic,
        tags: tags,
        metrics: "", // 기본값 (Form 필드 없음)
        milestones: formattedMilestones,
      };

      console.log("🚀 Final projectData:", projectData);

      const response = await apiClient.createProject(projectData);

      if (response.success) {
        setShowLaunchSuccess(true);
        setIsHovered(false); // hover 상태 초기화
        showSuccess("프로젝트가 성공적으로 생성되었습니다! 🎉");

        // 성공 메시지 표시 후 대시보드로 이동
        setTimeout(() => {
          navigate("/dashboard");
        }, 2000);
      } else {
        showError("프로젝트 생성에 실패했습니다");
        setIsLaunching(false); // 실패 시 애니메이션 리셋
        setIsHovered(false); // hover 상태 초기화
      }
    } catch (error: unknown) {
      console.error("프로젝트 생성 실패:", error);
      showError("프로젝트 생성에 실패했습니다");
      setIsLaunching(false); // 에러 시 애니메이션 리셋
      setIsHovered(false); // hover 상태 초기화
    } finally {
      setLoading(false);
    }
  };

  // 뒤로가기
  const handleBack = () => {
    navigate("/");
  };

  return (
    <div
      className="min-h-screen py-8"
      style={{
        backgroundColor: "var(--bg-primary)",
      }}
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
            홈으로 돌아가기
          </Button>

          <div className="text-center">
            <Title level={2} style={{ color: "var(--text-primary)" }}>
              <ProjectOutlined className="mr-3" />새 프로젝트 만들기
            </Title>
            <Paragraph style={{ color: "var(--text-secondary)" }}>
              당신의 아이디어를 현실로 만들어보세요! 투자자들과 함께 목표를
              달성하세요.
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
              description="기본 정보 입력"
            />
            <Step
              title="마일스톤 설정"
              icon={<CalendarOutlined />}
              description="단계별 목표 설정"
            />
            <Step
              title="최종 검토"
              icon={<CheckCircleOutlined />}
              description="검토 및 발행"
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
                      placeholder="프로젝트에 대해 자세히 설명해주세요. 무엇을 만들고, 왜 중요한지 알려주세요."
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
                      onChange={handleProjectDateChange}
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
                  style={{
                    color: "var(--blue)",
                  }}
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
                {/* AI 제안 섹션 */}
                <Card
                  size="small"
                  style={{
                    backgroundColor: "var(--bg-tertiary)",
                    borderColor: "var(--border-color)",
                  }}
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <Title
                        level={5}
                        className="mb-1"
                        style={{ color: "var(--text-primary)" }}
                      >
                        <RobotOutlined className="mr-2" />
                        AI 마일스톤 제안받기
                      </Title>
                      <Paragraph
                        className="mb-0 text-sm"
                        style={{ color: "var(--text-secondary)" }}
                      >
                        AI가 프로젝트에 맞는 단계별 마일스톤을 제안해드립니다
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
                    <div
                      className="mt-3 text-xs"
                      style={{ color: "var(--text-secondary)" }}
                    >
                      AI 사용량: {aiUsageInfo.used}/{aiUsageInfo.limit}회 사용
                    </div>
                  )}
                </Card>

                {/* AI 제안 결과 */}
                {aiSuggestions && (
                  <Card
                    size="small"
                    title="🤖 AI 제안 마일스톤"
                    style={{
                      backgroundColor: "var(--bg-secondary)",
                      borderColor: "var(--border-color)",
                    }}
                    headStyle={{
                      backgroundColor: "var(--bg-secondary)",
                      borderBottomColor: "var(--border-color)",
                      color: "var(--text-primary)",
                    }}
                    bodyStyle={{
                      backgroundColor: "var(--bg-secondary)",
                    }}
                    extra={
                      <Button
                        size="small"
                        type="primary"
                        onClick={applyAISuggestions}
                      >
                        제안 적용하기
                      </Button>
                    }
                  >
                    <div className="space-y-2">
                      {aiSuggestions.milestones?.map(
                        (milestone: AIMilestone, index: number) => (
                          <div
                            key={index}
                            className="p-2 rounded"
                            style={{
                              backgroundColor: "var(--bg-tertiary)",
                              border: "1px solid var(--border-color)",
                            }}
                          >
                            <div
                              className="font-medium"
                              style={{ color: "var(--text-primary)" }}
                            >
                              {milestone.title}
                            </div>
                            <div
                              className="text-sm"
                              style={{ color: "var(--text-secondary)" }}
                            >
                              {milestone.description}
                            </div>
                          </div>
                        )
                      )}
                    </div>
                  </Card>
                )}

                {/* 마일스톤 추가 버튼 */}
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

                {/* 마일스톤 목록 */}
                <div className="space-y-4">
                  {milestones.map((milestone, index) => (
                    <Card
                      key={index}
                      size="small"
                      title={`🎯 마일스톤 ${index + 1}`}
                      extra={
                        <Button
                          size="small"
                          danger
                          onClick={() => removeMilestone(index)}
                        >
                          삭제
                        </Button>
                      }
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
                              ValidationRules.uniqueMilestoneTitle(
                                index,
                                MessageHelpers.getDuplicateMilestoneMessage(
                                  milestone.title || ""
                                )
                              ),
                            ]}
                            validationContext={milestones}
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
                            placeholder={
                              index === milestones.length - 1
                                ? "프로젝트 완료일과 자동 동기화"
                                : form.getFieldValue("target_date")
                                ? `목표일 (${dayjs(
                                    form.getFieldValue("target_date")
                                  ).format("MM/DD")} 이전)`
                                : "먼저 프로젝트 완료일을 설정하세요"
                            }
                            value={
                              milestone.target_date
                                ? dayjs(milestone.target_date)
                                : null
                            }
                            onChange={(date) => {
                              // 마지막 마일스톤이 아닌 경우에만 수동 변경 허용
                              if (index !== milestones.length - 1) {
                                updateMilestone(
                                  index,
                                  "target_date",
                                  date ? date.format("YYYY-MM-DD") : ""
                                );
                              }
                            }}
                            disabledDate={getMilestoneDisabledDate(index)}
                            disabled={
                              !form.getFieldValue("target_date") ||
                              index === milestones.length - 1
                            }
                          />
                          {form.getFieldValue("target_date") &&
                            index === milestones.length - 1 && (
                              <div
                                className="text-xs mt-1"
                                style={{ color: "var(--blue)" }}
                              >
                                🔗 마지막 마일스톤은 프로젝트 완료일과 자동
                                동기화됩니다
                              </div>
                            )}
                          {form.getFieldValue("target_date") &&
                            index !== milestones.length - 1 && (
                              <div
                                className="text-xs mt-1"
                                style={{ color: "var(--text-secondary)" }}
                              >
                                다른 마일스톤과 다른 날짜를 선택하세요
                              </div>
                            )}
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
                    </Card>
                  ))}

                  {milestones.length === 0 && (
                    <div className="text-center py-8 text-gray-500">
                      <Paragraph>
                        마일스톤을 추가하거나 AI 제안을 받아보세요
                      </Paragraph>
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
              title="✅ 최종 검토 및 발행"
              style={{
                backgroundColor: "var(--bg-secondary)",
                borderColor: "var(--border-color)",
              }}
            >
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  gap: "24px",
                }}
              >
                <Alert
                  message="프로젝트 발행 전 최종 확인"
                  description="아래 정보를 확인하고 프로젝트를 발행하세요. 발행 후에도 수정이 가능합니다."
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
                  <div
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      gap: "16px",
                    }}
                  >
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
                        <div
                          style={{
                            display: "flex",
                            flexDirection: "column",
                            gap: "8px",
                          }}
                        >
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
                              {milestone.betting_type === "custom" &&
                                (milestone.betting_options || []).length >
                                  0 && (
                                  <div style={{ marginTop: "8px" }}>
                                    <div
                                      style={{
                                        display: "flex",
                                        flexWrap: "wrap",
                                        gap: "4px",
                                      }}
                                    >
                                      {(milestone.betting_options || []).map(
                                        (
                                          option: string,
                                          optionIndex: number
                                        ) => (
                                          <Tag key={optionIndex} color="blue">
                                            {option}
                                          </Tag>
                                        )
                                      )}
                                    </div>
                                  </div>
                                )}
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
                        <div
                          style={{
                            display: "flex",
                            flexWrap: "wrap",
                            gap: "8px",
                          }}
                        >
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
                    disabled={loading || isLaunching}
                    onClick={handleLaunchProject}
                    onMouseEnter={() => setIsHovered(true)}
                    onMouseLeave={() => !isLaunching && setIsHovered(false)}
                    className="rocket-launch-button"
                    style={{
                      background:
                        "linear-gradient(135deg, var(--blue) 0%, #9333ea 100%)",
                      borderColor: "var(--blue)",
                      minWidth: "200px",
                      height: "48px",
                    }}
                  >
                    <span
                      className={`button-content ${
                        isHovered ? "preparing" : ""
                      } ${isLaunching ? "launching" : ""}`}
                    >
                      🚀 프로젝트 발행하기
                    </span>

                    {/* 로켓 애니메이션 요소들 */}
                    <div
                      className={`rocket-container ${
                        isHovered ? "preparing" : ""
                      } ${isLaunching ? "launching" : ""}`}
                    >
                      🚀
                    </div>

                    <div
                      className={`rocket-trail ${
                        isLaunching ? "launching" : ""
                      }`}
                    ></div>

                    <div className={`sparks ${isLaunching ? "launching" : ""}`}>
                      <div
                        className="spark"
                        style={{ top: "40%", left: "45%" }}
                      ></div>
                      <div
                        className="spark"
                        style={{ top: "60%", left: "55%" }}
                      ></div>
                      <div
                        className="spark"
                        style={{ top: "45%", left: "35%" }}
                      ></div>
                      <div
                        className="spark"
                        style={{ top: "55%", left: "65%" }}
                      ></div>
                      <div
                        className="spark"
                        style={{ top: "50%", left: "50%" }}
                      ></div>
                    </div>

                    <div
                      className={`launch-success ${
                        showLaunchSuccess ? "show" : ""
                      }`}
                    >
                      🎯 발사 완료!
                    </div>
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

export default CreateProjectPage;
