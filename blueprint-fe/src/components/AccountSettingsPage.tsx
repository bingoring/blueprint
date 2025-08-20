import {
  BellOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  GithubOutlined,
  LinkedinOutlined,
  PhoneOutlined,
  SafetyCertificateOutlined,
  SaveOutlined,
  SecurityScanOutlined,
  TwitterOutlined,
  UploadOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  Row,
  Space,
  Switch,
  Tabs,
  Tag,
  Typography,
  Upload,
  message,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { SettingsAggregateResponse } from "../types";
import GlobalNavbar from "./GlobalNavbar";
import { ConnectionIcon } from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

const AccountSettingsPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuthStore();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<SettingsAggregateResponse | null>(
    null
  );
  const [activeTab, setActiveTab] = useState("profile");

  // Load user settings
  useEffect(() => {
    const loadSettings = async () => {
      if (!isAuthenticated) {
        navigate("/login");
        return;
      }

      try {
        setLoading(true);
        const response = await apiClient.getMySettings();

        if (response.success && response.data) {
          setSettings(response.data);

          // 폼 초기값 설정
          form.setFieldsValue({
            displayName:
              response.data.profile?.display_name || user?.displayName || "",
            email: response.data.user?.email || user?.email || "",
            bio: response.data.profile?.bio || "",
            location: response.data.profile?.location || "",
            website: response.data.profile?.website || "",
            github: response.data.profile?.github_link || "",
            linkedin: response.data.profile?.linkedin_link || "",
            twitter: response.data.profile?.twitter_link || "",
            emailNotifications:
              response.data.profile?.email_notifications ?? true,
            pushNotifications:
              response.data.profile?.push_notifications ?? false,
            marketingNotifications:
              response.data.profile?.marketing_notifications ?? false,
          });
        }
      } catch (error) {
        console.error("설정 로드 실패:", error);
        message.error("설정을 불러오는데 실패했습니다.");
      } finally {
        setLoading(false);
      }
    };

    loadSettings();
  }, [isAuthenticated, navigate, form, user]);

  // Save profile settings
  const handleSaveProfile = async (values: {
    displayName: string;
    bio: string;
    location: string;
    website: string;
    github: string;
    linkedin: string;
    twitter: string;
  }) => {
    try {
      setSaving(true);

      const profileData = {
        display_name: values.displayName,
        bio: values.bio,
        location: values.location,
        website: values.website,
        github_link: values.github,
        linkedin_link: values.linkedin,
        twitter_link: values.twitter,
      };

      const response = await apiClient.updateMyProfile(profileData);

      if (response.success) {
        message.success("프로필이 저장되었습니다.");
      } else {
        message.error("프로필 저장에 실패했습니다.");
      }
    } catch (error) {
      console.error("프로필 저장 실패:", error);
      message.error("프로필 저장에 실패했습니다.");
    } finally {
      setSaving(false);
    }
  };

  // Save notification settings
  const handleSaveNotifications = async (values: {
    emailNotifications: boolean;
    pushNotifications: boolean;
    marketingNotifications: boolean;
  }) => {
    try {
      setSaving(true);

      const response = await apiClient.updatePreferences({
        email_notifications: values.emailNotifications,
        push_notifications: values.pushNotifications,
        marketing_notifications: values.marketingNotifications,
      });

      if (response.success) {
        message.success("알림 설정이 저장되었습니다.");
      } else {
        message.error("알림 설정 저장에 실패했습니다.");
      }
    } catch (error) {
      console.error("알림 설정 저장 실패:", error);
      message.error("알림 설정 저장에 실패했습니다.");
    } finally {
      setSaving(false);
    }
  };

  // Avatar upload
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const handleAvatarUpload = async (_file: File) => {
    // Mock upload
    message.success("프로필 사진이 업로드되었습니다.");
    return false; // Prevent default upload
  };

  // Email verification
  const handleEmailVerification = async () => {
    try {
      message.info("인증 이메일을 발송했습니다.");
    } catch {
      message.error("이메일 인증 발송에 실패했습니다.");
    }
  };

  // Phone verification
  const handlePhoneVerification = async () => {
    try {
      message.info("인증 문자를 발송했습니다.");
    } catch {
      message.error("휴대폰 인증 발송에 실패했습니다.");
    }
  };

  if (!isAuthenticated) {
    return null;
  }

  const tabItems = [
    {
      key: "profile",
      label: (
        <Space>
          <UserOutlined />
          프로필 설정
        </Space>
      ),
      children: (
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveProfile}
          style={{ maxWidth: 800 }}
        >
          <Card title="기본 정보" style={{ marginBottom: 24 }}>
            <Row gutter={24}>
              <Col span={8}>
                <div style={{ textAlign: "center" }}>
                  <Avatar
                    size={120}
                    src={
                      settings?.profile?.avatar ||
                      `https://api.dicebear.com/6.x/avataaars/svg?seed=${user?.username}`
                    }
                    icon={<UserOutlined />}
                    style={{ marginBottom: 16 }}
                  />
                  <div>
                    <Upload
                      showUploadList={false}
                      beforeUpload={handleAvatarUpload}
                      accept="image/*"
                    >
                      <Button icon={<UploadOutlined />} size="small">
                        사진 변경
                      </Button>
                    </Upload>
                  </div>
                </div>
              </Col>
              <Col span={16}>
                <Form.Item
                  label="표시 이름"
                  name="displayName"
                  rules={[
                    { required: true, message: "표시 이름을 입력해주세요" },
                    { min: 2, message: "최소 2자 이상 입력해주세요" },
                  ]}
                >
                  <Input placeholder="다른 사용자에게 보여질 이름" />
                </Form.Item>

                <Form.Item
                  label="이메일 주소"
                  name="email"
                  rules={[
                    { type: "email", message: "올바른 이메일 형식이 아닙니다" },
                  ]}
                >
                  <Input placeholder="your@email.com" disabled />
                </Form.Item>

                <Form.Item label="자기소개" name="bio">
                  <TextArea
                    rows={3}
                    placeholder="자신을 소개해주세요 (최대 200자)"
                    maxLength={200}
                    showCount
                  />
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Card title="추가 정보" style={{ marginBottom: 24 }}>
            <Form.Item label="위치" name="location">
              <Input placeholder="서울, 대한민국" />
            </Form.Item>

            <Form.Item label="웹사이트" name="website">
              <Input placeholder="https://example.com" />
            </Form.Item>

            <Divider>소셜 미디어</Divider>

            <Row gutter={16}>
              <Col span={8}>
                <Form.Item
                  label={
                    <>
                      <GithubOutlined /> GitHub
                    </>
                  }
                  name="github"
                >
                  <Input placeholder="github.com/username" />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  label={
                    <>
                      <LinkedinOutlined /> LinkedIn
                    </>
                  }
                  name="linkedin"
                >
                  <Input placeholder="linkedin.com/in/username" />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  label={
                    <>
                      <TwitterOutlined /> Twitter
                    </>
                  }
                  name="twitter"
                >
                  <Input placeholder="twitter.com/username" />
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={saving}
              icon={<SaveOutlined />}
              size="large"
            >
              프로필 저장
            </Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: "security",
      label: (
        <Space>
          <SafetyCertificateOutlined />
          보안 설정
        </Space>
      ),
      children: (
        <div style={{ maxWidth: 800 }}>
          <Card title="신원 인증" style={{ marginBottom: 24 }}>
            <Space direction="vertical" size="large" style={{ width: "100%" }}>
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong>이메일 인증</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    {settings?.user.email}
                  </Text>
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  {settings?.verification?.email_verified ? (
                    <Tag color="green" icon={<CheckCircleOutlined />}>
                      인증 완료
                    </Tag>
                  ) : (
                    <>
                      <Tag color="orange" icon={<ExclamationCircleOutlined />}>
                        미인증
                      </Tag>
                      <Button size="small" onClick={handleEmailVerification}>
                        인증하기
                      </Button>
                    </>
                  )}
                </div>
              </div>

              <Divider style={{ margin: "12px 0" }} />

              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong>휴대폰 인증</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    휴대폰 번호를 등록하여 계정을 보호하세요
                  </Text>
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  {settings?.verification?.phone_verified ? (
                    <Tag color="green" icon={<CheckCircleOutlined />}>
                      인증 완료
                    </Tag>
                  ) : (
                    <>
                      <Tag color="orange" icon={<ExclamationCircleOutlined />}>
                        미인증
                      </Tag>
                      <Button size="small" onClick={handlePhoneVerification}>
                        인증하기
                      </Button>
                    </>
                  )}
                </div>
              </div>

              {!settings?.verification?.phone_verified && (
                <div>
                  <Form layout="inline" style={{ width: "100%" }}>
                    <Form.Item style={{ flex: 1 }}>
                      <Input
                        placeholder="휴대폰 번호를 입력하세요"
                        prefix={<PhoneOutlined />}
                      />
                    </Form.Item>
                    <Form.Item>
                      <Button onClick={handlePhoneVerification}>
                        번호 등록
                      </Button>
                    </Form.Item>
                  </Form>
                </div>
              )}

              <Divider style={{ margin: "12px 0" }} />

              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong>신원 증명</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    정부 발급 신분증을 통한 본인 확인
                  </Text>
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  {settings?.verification?.professional_status === "approved" ||
                  settings?.verification?.education_status === "approved" ? (
                    <Tag color="green" icon={<CheckCircleOutlined />}>
                      인증 완료
                    </Tag>
                  ) : (
                    <>
                      <Tag color="orange" icon={<ExclamationCircleOutlined />}>
                        미인증
                      </Tag>
                      <Button size="small" disabled>
                        준비 중
                      </Button>
                    </>
                  )}
                </div>
              </div>
            </Space>
          </Card>

          <Card title="보안 강화" style={{ marginBottom: 24 }}>
            <Space direction="vertical" size="middle" style={{ width: "100%" }}>
              <Paragraph type="secondary">
                <SecurityScanOutlined /> 인증 완료 시 혜택:
              </Paragraph>
              <ul style={{ paddingLeft: 20, margin: 0 }}>
                <li>프로필에 인증 배지 표시</li>
                <li>투자 한도 증가 ($10,000 → $50,000)</li>
                <li>멘토 자격 획득</li>
                <li>우선 고객 지원</li>
              </ul>
            </Space>
          </Card>
        </div>
      ),
    },
    {
      key: "notifications",
      label: (
        <Space>
          <BellOutlined />
          알림 설정
        </Space>
      ),
      children: (
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveNotifications}
          style={{ maxWidth: 800 }}
        >
          <Card title="알림 설정" style={{ marginBottom: 24 }}>
            <Space direction="vertical" size="large" style={{ width: "100%" }}>
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong>이메일 알림</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    중요한 업데이트를 이메일로 받습니다
                  </Text>
                </div>
                <Form.Item
                  name="emailNotifications"
                  valuePropName="checked"
                  style={{ margin: 0 }}
                >
                  <Switch />
                </Form.Item>
              </div>

              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong>푸시 알림</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    브라우저 푸시 알림을 받습니다
                  </Text>
                </div>
                <Form.Item
                  name="pushNotifications"
                  valuePropName="checked"
                  style={{ margin: 0 }}
                >
                  <Switch />
                </Form.Item>
              </div>

              <Divider />

              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong>마케팅 알림</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: "12px" }}>
                    새로운 기능 및 프로모션 정보를 받습니다
                  </Text>
                </div>
                <Form.Item
                  name="marketingNotifications"
                  valuePropName="checked"
                  style={{ margin: 0 }}
                >
                  <Switch />
                </Form.Item>
              </div>
            </Space>
          </Card>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={saving}
              icon={<SaveOutlined />}
              size="large"
            >
              알림 설정 저장
            </Button>
          </Form.Item>
        </Form>
      ),
    },
  ];

  return (
    <div style={{ background: "var(--bg-primary)", minHeight: "100vh" }}>
      <GlobalNavbar />

      <div style={{ paddingTop: "64px" }}>
        <div
          style={{
            maxWidth: "1200px",
            margin: "0 auto",
            padding: "32px 24px",
          }}
        >
          {/* 헤더 */}
          <div style={{ marginBottom: "32px" }}>
            <Space align="start" size={16}>
              <ConnectionIcon size={32} color="var(--primary-color)" />
              <div>
                <Title
                  level={2}
                  style={{ margin: 0, color: "var(--text-primary)" }}
                >
                  계정 설정
                </Title>
                <Text type="secondary" style={{ fontSize: "16px" }}>
                  프로필 정보와 보안 설정을 관리하세요
                </Text>
              </div>
            </Space>
          </div>

          {loading ? (
            <div style={{ textAlign: "center", padding: "100px" }}>
              <Text>설정을 불러오는 중...</Text>
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

export default AccountSettingsPage;
