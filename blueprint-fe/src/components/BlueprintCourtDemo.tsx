import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  ExclamationCircleOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Badge,
  Button,
  Card,
  Col,
  Progress,
  Row,
  Space,
  Tag,
  Typography,
} from "antd";
import React, { useState } from "react";
import type { CreateDisputeRequest, TimeRemaining } from "../types";
import GlobalNavbar from "./GlobalNavbar";
import CreateDisputeModal from "./court/CreateDisputeModal";
import DisputeTimer from "./court/DisputeTimer";

const { Title, Text, Paragraph } = Typography;

export const BlueprintCourtDemo: React.FC = () => {
  const [disputeModalVisible, setDisputeModalVisible] = useState(false);

  // Mock data
  const mockTimeRemaining: TimeRemaining = {
    phase: "challenge_window",
    hours: 23,
    minutes: 45,
    seconds: 12,
    is_expired: false,
  };

  const handleCreateDispute = async (request: CreateDisputeRequest) => {
    console.log("Creating dispute:", request);
    // TODO: API call
  };

  return (
    <div
      className="min-h-screen"
      style={{ backgroundColor: "var(--bg-layout)" }}
    >
      <GlobalNavbar />

      <div className="container mx-auto px-4 py-6 max-w-6xl">
        {/* Header */}
        <Card
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border-color)",
          }}
          className="text-center mb-6"
        >
          <div className="py-8">
            <div className="flex justify-center mb-4">
              <div className="p-4 rounded-full bg-blue-100 dark:bg-blue-900">
                <CheckCircleOutlined className="text-4xl text-blue-500" />
              </div>
            </div>
            <Title level={2} style={{ color: "var(--text-primary)" }}>
              Blueprint Court ⚖️
            </Title>
            <Paragraph
              className="text-lg"
              style={{ color: "var(--text-secondary)" }}
            >
              투명하고 공정한 분쟁 해결 시스템
            </Paragraph>
            <Text style={{ color: "var(--text-secondary)" }}>
              커뮤니티의 집단지성을 통해 마일스톤 결과의 객관적 진실을
              판결합니다
            </Text>
          </div>
        </Card>

        {/* 48시간 이의 제기 창 */}
        <Card
          title={
            <div className="flex items-center gap-2">
              <ClockCircleOutlined style={{ color: "#52c41a" }} />
              <Text strong>1단계: 이의 제기 창 (Challenge Window)</Text>
            </div>
          }
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border-color)",
          }}
          className="mb-6"
        >
          <Row gutter={24}>
            <Col span={14}>
              <div className="space-y-4">
                <div
                  className="p-4 rounded-lg"
                  style={{ backgroundColor: "var(--bg-secondary)" }}
                >
                  <div className="flex items-center justify-between mb-3">
                    <Text strong>마일스톤: "앱 정식 출시"</Text>
                    <Badge status="success" text="성공 보고됨" />
                  </div>
                  <Text
                    className="text-sm"
                    style={{ color: "var(--text-secondary)" }}
                  >
                    프로젝트 생성자가 결과를 보고했습니다. 투자자들은 48시간
                    동안 이의를 제기할 수 있습니다.
                  </Text>
                </div>

                <Alert
                  icon={<ExclamationCircleOutlined />}
                  type="info"
                  showIcon
                  message="이의 제기 자격"
                  description="해당 마일스톤에 1 USDC 이상 투자한 참여자만 이의를 제기할 수 있습니다."
                />

                <Button
                  type="primary"
                  danger
                  icon={<ExclamationCircleOutlined />}
                  onClick={() => setDisputeModalVisible(true)}
                >
                  결과에 이의 제기
                </Button>
              </div>
            </Col>
            <Col span={10}>
              <DisputeTimer
                timeRemaining={mockTimeRemaining}
                phase="challenge_window"
              />
            </Col>
          </Row>
        </Card>

        {/* 투표 기간 데모 */}
        <Card
          title={
            <div className="flex items-center gap-2">
              <TeamOutlined style={{ color: "#1890ff" }} />
              <Text strong>2단계: 분쟁 투표 (Voting Period)</Text>
            </div>
          }
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border-color)",
          }}
          className="mb-6"
        >
          <Row gutter={24}>
            <Col span={12}>
              <Card size="small" title="Tier 1: 전문가 판결">
                <div className="space-y-3">
                  <Text className="text-sm">총 투자액 {"<"} $10,000</Text>
                  <Text className="block text-sm">
                    판결단: 상위 투자자 10명
                  </Text>
                  <Progress percent={70} strokeColor="#52c41a" />
                  <Text className="text-sm">투표 참여율: 70% (7/10)</Text>
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card size="small" title="Tier 2: DAO 거버넌스">
                <div className="space-y-3">
                  <Text className="text-sm">총 투자액 ≥ $10,000</Text>
                  <Text className="block text-sm">
                    투표자: 모든 토큰 보유자
                  </Text>
                  <Progress percent={23.4} strokeColor="#1890ff" />
                  <Text className="text-sm">투표 참여율: 23.4% (234/1000)</Text>
                </div>
              </Card>
            </Col>
          </Row>
        </Card>

        {/* 활성 분쟁 예시 */}
        <Card
          title="활성 분쟁 예시"
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border-color)",
          }}
        >
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <Title level={5} className="mb-0">
                "매출 1억 달성" 마일스톤
              </Title>
              <Badge color="#1890ff" text="DAO 거버넌스" />
              <Tag color="processing">투표 중</Tag>
            </div>

            <Text style={{ color: "var(--text-secondary)" }}>
              프로젝트: 블록체인 스타트업
            </Text>

            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <DollarOutlined />
                <Text>총 투자액: $15,000</Text>
              </div>
              <div className="flex items-center gap-2">
                <TeamOutlined />
                <Text>투표 참여: 234/1000</Text>
              </div>
            </div>

            {/* 투표 현황 */}
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <Text>생성자 결과 유지</Text>
                <Text>145표</Text>
              </div>
              <Progress percent={62} strokeColor="#52c41a" showInfo={false} />

              <div className="flex justify-between text-sm">
                <Text>분쟁 제기자 지지</Text>
                <Text>89표</Text>
              </div>
              <Progress percent={38} strokeColor="#ff4d4f" showInfo={false} />
            </div>

            <Space>
              <Button type="primary" size="small">
                결과 유지
              </Button>
              <Button danger size="small">
                이의 지지
              </Button>
            </Space>
          </div>
        </Card>
      </div>

      {/* Create Dispute Modal */}
      <CreateDisputeModal
        visible={disputeModalVisible}
        onClose={() => setDisputeModalVisible(false)}
        onSubmit={handleCreateDispute}
        milestoneId={10}
        milestoneTitle="앱 정식 출시"
        originalResult={true}
        evidenceUrl="https://apps.apple.com/app/example"
        evidenceDescription="앱스토어에 앱이 정식으로 출시되었습니다."
      />
    </div>
  );
};

export default BlueprintCourtDemo;
