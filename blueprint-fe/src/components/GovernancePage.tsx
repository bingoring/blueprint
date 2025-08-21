import {
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  TeamOutlined,
  TrophyOutlined,
} from "@ant-design/icons";
import {
  Badge,
  Button,
  Card,
  Col,
  Progress,
  Row,
  Statistic,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from "antd";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiClient } from "../lib/api";
import { useAuthStore } from "../stores/useAuthStore";
import type { ActiveDisputesResponse } from "../types";
import GlobalNavbar from "./GlobalNavbar";
import { GovernanceIcon } from "./icons/BlueprintIcons";

const { Title, Text, Paragraph } = Typography;

const GovernancePage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [activeDisputes, setActiveDisputes] =
    useState<ActiveDisputesResponse | null>(null);
  const [activeTab, setActiveTab] = useState("active-disputes");

  const loadActiveDisputes = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getActiveDisputes();
      if (response.success && response.data) {
        setActiveDisputes(response.data);
      } else {
        message.error("분쟁 목록을 불러올 수 없습니다");
      }
    } catch (error) {
      console.error("분쟁 목록 로드 실패:", error);
      message.error("분쟁 목록을 불러오는 중 오류가 발생했습니다");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadActiveDisputes();
  }, []);

  if (!isAuthenticated) {
    navigate("/login");
    return null;
  }

  // Tier 1 분쟁 테이블 컬럼
  const tier1Columns = [
    {
      title: "프로젝트",
      dataIndex: "project_title",
      key: "project_title",
      render: (text: string) => (
        <Text strong style={{ color: "var(--text-primary)" }}>
          {text}
        </Text>
      ),
    },
    {
      title: "마일스톤",
      dataIndex: "milestone_title",
      key: "milestone_title",
    },
    {
      title: "심급",
      dataIndex: "tier",
      key: "tier",
      render: (tier: string) => (
        <Tag color={tier === "expert" ? "blue" : "purple"}>
          {tier === "expert" ? "전문가 판결" : "DAO 거버넌스"}
        </Tag>
      ),
    },
    {
      title: "투자액",
      dataIndex: "total_investment",
      key: "total_investment",
      render: (amount: number) => `$${amount.toLocaleString()}`,
    },
    {
      title: "투표 진행률",
      dataIndex: "voting_stats",
      key: "voting_progress",
      render: (stats: {
        voting_progress: number;
        voted_count: number;
        total_voters: number;
      }) => (
        <div>
          <Progress
            percent={Math.round(stats.voting_progress * 100)}
            size="small"
            strokeColor="#1890ff"
          />
          <Text className="text-sm" style={{ color: "var(--text-secondary)" }}>
            {stats.voted_count}/{stats.total_voters} 투표
          </Text>
        </div>
      ),
    },
    {
      title: "남은 시간",
      dataIndex: "time_remaining",
      key: "time_remaining",
      render: (time: { hours: number; minutes: number }) => (
        <div className="flex items-center gap-1">
          <ClockCircleOutlined />
          <Text>
            {time.hours}시간 {time.minutes}분
          </Text>
        </div>
      ),
    },
    {
      title: "액션",
      key: "action",
      render: (record: { id: number }) => (
        <Button
          type="primary"
          size="small"
          onClick={() => navigate(`/disputes/${record.id}`)}
        >
          투표하기
        </Button>
      ),
    },
  ];

  const tabItems = [
    {
      key: "active-disputes",
      label: (
        <span className="flex items-center gap-2">
          <ExclamationCircleOutlined />
          진행 중인 분쟁
        </span>
      ),
      children: (
        <div className="space-y-6">
          {/* 통계 카드 */}
          <Row gutter={[24, 24]}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="진행 중인 분쟁"
                  value={
                    (activeDisputes?.active_disputes.length || 0) +
                    (activeDisputes?.governance_disputes.length || 0)
                  }
                  prefix={<ExclamationCircleOutlined />}
                  valueStyle={{ color: "#fa8c16" }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="전문가 판결"
                  value={activeDisputes?.active_disputes.length || 0}
                  prefix={<TeamOutlined />}
                  valueStyle={{ color: "#1890ff" }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="DAO 거버넌스"
                  value={activeDisputes?.governance_disputes.length || 0}
                  prefix={<GovernanceIcon size={16} />}
                  valueStyle={{ color: "#722ed1" }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="총 분쟁 해결"
                  value={158} // Mock 데이터
                  prefix={<TrophyOutlined />}
                  valueStyle={{ color: "#52c41a" }}
                />
              </Card>
            </Col>
          </Row>

          {/* Tier 1: 전문가 판결 */}
          <Card
            title={
              <div className="flex items-center gap-2">
                <TeamOutlined style={{ color: "#1890ff" }} />
                <Text strong>Tier 1: 전문가 판결 (투자액 &lt; $10,000)</Text>
              </div>
            }
            style={{
              background: "var(--bg-card)",
              border: "1px solid var(--border-color)",
            }}
          >
            <Table
              columns={tier1Columns}
              dataSource={activeDisputes?.active_disputes || []}
              rowKey="id"
              loading={loading}
              pagination={false}
              locale={{
                emptyText: "현재 진행 중인 전문가 판결이 없습니다.",
              }}
            />
          </Card>

          {/* Tier 2: DAO 거버넌스 */}
          <Card
            title={
              <div className="flex items-center gap-2">
                <GovernanceIcon size={16} color="#722ed1" />
                <Text strong>Tier 2: DAO 거버넌스 (투자액 ≥ $10,000)</Text>
              </div>
            }
            style={{
              background: "var(--bg-card)",
              border: "1px solid var(--border-color)",
            }}
          >
            <Table
              columns={tier1Columns}
              dataSource={activeDisputes?.governance_disputes || []}
              rowKey="id"
              loading={loading}
              pagination={false}
              locale={{
                emptyText: "현재 진행 중인 DAO 거버넌스 분쟁이 없습니다.",
              }}
            />
          </Card>
        </div>
      ),
    },
    {
      key: "how-it-works",
      label: (
        <span className="flex items-center gap-2">
          <GovernanceIcon size={16} />
          작동 방식
        </span>
      ),
      children: (
        <div className="space-y-6">
          <Card
            title="⚖️ Blueprint Court 분쟁 해결 시스템"
            style={{
              background: "var(--bg-card)",
              border: "1px solid var(--border-color)",
            }}
          >
            <div className="space-y-4">
              <Paragraph>
                Blueprint Court는 커뮤니티의 집단지성을 통해 마일스톤 결과의
                객관적 진실을 판결하는 탈중앙화 분쟁 해결 시스템입니다.
              </Paragraph>

              <Row gutter={[24, 24]}>
                <Col span={12}>
                  <Card size="small" title="1단계: 결과 보고">
                    <ul className="list-disc list-inside space-y-2">
                      <li>프로젝트 생성자가 마일스톤 결과 보고</li>
                      <li>48시간 이의 제기 창 오픈</li>
                      <li>투자자들에게 알림 발송</li>
                    </ul>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card size="small" title="2단계: 이의 제기">
                    <ul className="list-disc list-inside space-y-2">
                      <li>1 USDC 이상 투자자만 참여 가능</li>
                      <li>100 $BLUEPRINT 예치 필요</li>
                      <li>최소 100자 이상 사유 작성</li>
                    </ul>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card size="small" title="3단계: 심급 결정">
                    <div className="space-y-2">
                      <Badge color="#1890ff" text="Tier 1: 전문가 판결" />
                      <div>총 투자액 &lt; $10,000</div>
                      <div>상위 투자자 10명이 판결</div>
                    </div>
                    <div className="mt-3 space-y-2">
                      <Badge color="#722ed1" text="Tier 2: DAO 거버넌스" />
                      <div>총 투자액 ≥ $10,000</div>
                      <div>모든 토큰 보유자 투표</div>
                    </div>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card size="small" title="4단계: 최종 판결">
                    <ul className="list-disc list-inside space-y-2">
                      <li>72시간 투표 기간</li>
                      <li>다수결 원칙으로 판결</li>
                      <li>예측 시장 자동 정산</li>
                      <li>예치금 분배/환급</li>
                    </ul>
                  </Card>
                </Col>
              </Row>
            </div>
          </Card>
        </div>
      ),
    },
  ];

  return (
    <div
      className="min-h-screen"
      style={{ backgroundColor: "var(--bg-layout)" }}
    >
      <GlobalNavbar />

      <div
        className="container mx-auto px-4 py-6 max-w-7xl"
        style={{ marginTop: "64px" }}
      >
        {/* Header */}
        <Card
          className="text-center mb-6"
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border-color)",
          }}
        >
          <div className="py-8">
            <div className="flex justify-center mb-4">
              <div className="p-4 rounded-full bg-purple-100 dark:bg-purple-900">
                <GovernanceIcon size={48} color="#722ed1" />
              </div>
            </div>
            <Title level={2} style={{ color: "var(--text-primary)" }}>
              Blueprint 거버넌스
            </Title>
            <Paragraph
              className="text-lg"
              style={{ color: "var(--text-secondary)" }}
            >
              커뮤니티의 집단지성으로 공정한 판결을 내리는 분산 법정
            </Paragraph>
            <Text style={{ color: "var(--text-secondary)" }}>
              투명하고 탈중앙화된 분쟁 해결 시스템에 참여하세요
            </Text>
          </div>
        </Card>

        {/* Tabs Content */}
        <Card
          style={{
            background: "var(--bg-card)",
            border: "1px solid var(--border-color)",
          }}
        >
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={tabItems}
            size="large"
          />
        </Card>
      </div>
    </div>
  );
};

export default GovernancePage;
