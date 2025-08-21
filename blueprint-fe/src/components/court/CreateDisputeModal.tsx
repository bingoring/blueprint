import {
  DollarOutlined,
  ExclamationCircleOutlined,
  FileTextOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import { Alert, Button, Divider, Input, Modal, Typography } from "antd";
import React, { useState } from "react";
import type { CreateDisputeRequest } from "../../types";

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

interface CreateDisputeModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (request: CreateDisputeRequest) => Promise<void>;
  milestoneId: number;
  milestoneTitle: string;
  originalResult: boolean; // true=성공, false=실패
  evidenceUrl?: string;
  evidenceDescription?: string;
  loading?: boolean;
}

export const CreateDisputeModal: React.FC<CreateDisputeModalProps> = ({
  visible,
  onClose,
  onSubmit,
  milestoneId,
  milestoneTitle,
  originalResult,
  evidenceUrl,
  evidenceDescription,
  loading = false,
}) => {
  const [disputeReason, setDisputeReason] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (disputeReason.length < 100) {
      return;
    }

    setSubmitting(true);
    try {
      await onSubmit({
        milestone_id: milestoneId,
        dispute_reason: disputeReason,
      });

      // 성공 시 모달 닫기 및 폼 초기화
      setDisputeReason("");
      onClose();
    } catch (error) {
      // 에러는 상위 컴포넌트에서 처리
      console.error("Failed to create dispute:", error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setDisputeReason("");
      onClose();
    }
  };

  const isFormValid = disputeReason.length >= 100;
  const remainingChars = Math.max(0, 100 - disputeReason.length);

  return (
    <Modal
      title={null}
      open={visible}
      onCancel={handleClose}
      footer={null}
      width={600}
      centered
      maskClosable={!submitting}
      closable={!submitting}
      className="dispute-modal"
    >
      <div className="space-y-6">
        {/* Header */}
        <div className="text-center pb-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex justify-center mb-3">
            <div className="p-3 rounded-full bg-red-100 dark:bg-red-900">
              <ExclamationCircleOutlined className="text-2xl text-red-500" />
            </div>
          </div>
          <Title
            level={3}
            className="mb-2"
            style={{ color: "var(--text-primary)" }}
          >
            결과에 이의 제기
          </Title>
          <Text className="text-lg" style={{ color: "var(--text-secondary)" }}>
            마일스톤: "{milestoneTitle}"
          </Text>
        </div>

        {/* Current Result */}
        <div
          className="p-4 rounded-lg"
          style={{ backgroundColor: "var(--bg-secondary)" }}
        >
          <div className="flex items-center gap-3 mb-3">
            <FileTextOutlined style={{ color: "var(--color-primary)" }} />
            <Text strong style={{ color: "var(--text-primary)" }}>
              현재 보고된 결과
            </Text>
          </div>

          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Text style={{ color: "var(--text-secondary)" }}>결과:</Text>
              <Text
                strong
                style={{
                  color: originalResult ? "#52c41a" : "#ff4d4f",
                }}
              >
                {originalResult ? "✅ 성공" : "❌ 실패"}
              </Text>
            </div>

            {evidenceUrl && (
              <div className="space-y-1">
                <Text style={{ color: "var(--text-secondary)" }}>
                  제출된 증거:
                </Text>
                <div className="p-2 bg-gray-50 dark:bg-gray-800 rounded">
                  <Text
                    className="break-all text-sm"
                    style={{ color: "var(--color-primary)" }}
                  >
                    {evidenceUrl}
                  </Text>
                </div>
              </div>
            )}

            {evidenceDescription && (
              <div className="space-y-1">
                <Text style={{ color: "var(--text-secondary)" }}>설명:</Text>
                <Text
                  className="text-sm"
                  style={{ color: "var(--text-primary)" }}
                >
                  {evidenceDescription}
                </Text>
              </div>
            )}
          </div>
        </div>

        {/* Stake Warning */}
        <Alert
          icon={<DollarOutlined />}
          type="warning"
          showIcon
          message="예치금 안내"
          description={
            <div className="space-y-2">
              <Paragraph className="mb-2">
                이의 제기를 위해 <strong>100 $BLUEPRINT</strong> 토큰을 예치해야
                합니다.
              </Paragraph>
              <ul className="list-disc list-inside space-y-1 text-sm">
                <li>
                  <strong>승소 시:</strong> 예치금은 전액 반환됩니다.
                </li>
                <li>
                  <strong>패소 시:</strong> 예치금은 몰수되어 판결자들에게
                  보상으로 지급됩니다.
                </li>
              </ul>
            </div>
          }
          className="border-orange-300"
        />

        {/* Dispute Reason Input */}
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <FileTextOutlined style={{ color: "var(--color-primary)" }} />
            <Text strong style={{ color: "var(--text-primary)" }}>
              이의 제기 사유 <span className="text-red-500">*</span>
            </Text>
          </div>

          <Text
            className="text-sm block"
            style={{ color: "var(--text-secondary)" }}
          >
            왜 이 결과가 부정확하다고 생각하는지 구체적으로 설명해주세요. (최소
            100자)
          </Text>

          <TextArea
            value={disputeReason}
            onChange={(e) => setDisputeReason(e.target.value)}
            placeholder="예: '제출된 앱스토어 링크는 실제 작동하는 앱이 아니라 단순 목업(Mockup) 이미지입니다. 이는 앱 정식 출시라는 성공 조건에 부합하지 않습니다. 실제로 링크를 확인해보면...'"
            rows={6}
            maxLength={1000}
            showCount
            disabled={submitting}
            style={{
              backgroundColor: "var(--bg-input)",
              color: "var(--text-primary)",
              borderColor:
                !isFormValid && disputeReason.length > 0
                  ? "#ff4d4f"
                  : "var(--border-color)",
            }}
          />

          <div className="flex justify-between items-center">
            <Text
              className="text-sm"
              style={{
                color: remainingChars > 0 ? "#ff4d4f" : "#52c41a",
              }}
            >
              {remainingChars > 0
                ? `${remainingChars}자 더 입력해주세요`
                : "✅ 입력 완료"}
            </Text>
            <Text
              className="text-sm"
              style={{ color: "var(--text-secondary)" }}
            >
              {disputeReason.length}/1000
            </Text>
          </div>
        </div>

        {/* System Abuse Warning */}
        <Alert
          icon={<WarningOutlined />}
          type="info"
          showIcon
          message="시스템 남용 방지"
          description="명백히 근거 없는 이의 제기는 시스템 남용으로 간주되며, 예치금 몰수 외에도 계정에 제재가 가해질 수 있습니다."
        />

        <Divider />

        {/* Action Buttons */}
        <div className="flex justify-end gap-3">
          <Button size="large" onClick={handleClose} disabled={submitting}>
            취소
          </Button>
          <Button
            type="primary"
            size="large"
            onClick={handleSubmit}
            loading={submitting}
            disabled={!isFormValid || loading}
            icon={<ExclamationCircleOutlined />}
            danger
            style={{
              backgroundColor: isFormValid ? "#ff4d4f" : undefined,
              borderColor: isFormValid ? "#ff4d4f" : undefined,
            }}
          >
            {submitting ? "이의 제기 중..." : "이의 제기 (100 $BLUEPRINT 예치)"}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default CreateDisputeModal;
