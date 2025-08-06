import type { InputRef } from "antd";
import { Button, Col, Input, Row, Typography } from "antd";
import React, { useCallback, useEffect, useRef, useState } from "react";
import type { ValidationRule } from "../../hooks/useValidation";
import { useValidation } from "../../hooks/useValidation";

const { Text } = Typography;

export interface FormFieldWithValidationProps {
  // 기본 입력 필드 속성
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;

  // Validation 관련
  validationRules: ValidationRule<string>[];
  validationContext?: unknown;

  // 액션 버튼 (옵션 추가 등)
  actionButton?: {
    text: string;
    icon?: React.ReactNode;
    onClick: () => void;
    disabled?: boolean;
  };

  // 레이아웃
  inputSpan?: number;
  buttonSpan?: number;

  // 이벤트 핸들러
  onEnter?: () => void;
  onValidationChange?: (isValid: boolean, error: string | null) => void;

  // 스타일링
  className?: string;
  showErrorInline?: boolean;
}

/**
 * 재사용 가능한 Validation 기능이 포함된 FormField 컴포넌트
 * - 실시간 validation
 * - 에러 상태 표시
 * - 다크모드 지원
 * - 확장 가능한 구조
 */
export const FormFieldWithValidation: React.FC<
  FormFieldWithValidationProps
> = ({
  value,
  onChange,
  placeholder,
  disabled = false,
  validationRules,
  validationContext,
  actionButton,
  inputSpan = 16,
  buttonSpan = 8,
  onEnter,
  onValidationChange,
  className,
  showErrorInline = true,
}) => {
  const [hasUserInteracted, setHasUserInteracted] = useState(false);
  const inputRef = useRef<InputRef>(null);

  const { error, isValid, validate, clearError } = useValidation({
    rules: validationRules,
    context: validationContext,
  });

  // validation 상태 변경을 부모에게 알림
  useEffect(() => {
    onValidationChange?.(isValid, error);
  }, [isValid, error, onValidationChange]);

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value;
      onChange(newValue);

      // 사용자가 의미 있는 입력을 시작했을 때 interaction 상태 설정
      if (!hasUserInteracted && newValue.length > 0) {
        setHasUserInteracted(true);
      }

      // 사용자가 상호작용한 후에만 실시간 validation 수행
      if (hasUserInteracted || newValue.length > 0) {
        // 실시간으로 validation 수행
        validate(newValue);
      } else if (error) {
        // 에러가 있는 상태에서 첫 입력 시에만 에러 클리어
        clearError();
      }

      // 입력값이 완전히 비어있을 때만 에러 클리어 (사용자가 모든 내용을 삭제한 경우)
      if (newValue === "" && error) {
        clearError();
      }
    },
    [
      onChange,
      hasUserInteracted,
      validate,
      error,
      clearError,
      setHasUserInteracted,
    ]
  );

  const handleKeyPress = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        setHasUserInteracted(true);
        const isValidInput = validate(value);
        if (isValidInput && onEnter) {
          onEnter();
        } else if (!isValidInput) {
          // validation 실패 시 input에 focus 유지
          setTimeout(() => {
            inputRef.current?.focus();
          }, 0);
        }
        // validation 실패 시에도 hasUserInteracted를 true로 설정하여
        // 이후 입력에서 실시간 validation이 작동하도록 함
      }
    },
    [value, validate, onEnter]
  );

  const handleBlur = useCallback(() => {
    setHasUserInteracted(true);
    // blur 시에는 현재 값이 비어있지 않은 경우에만 validation 수행
    if (value && value.trim()) {
      validate(value);
    } else if (error) {
      // 값이 비어있고 에러가 있다면 에러 클리어 (사용자가 입력을 지웠을 수 있음)
      clearError();
    }
  }, [value, validate, error, clearError]);

  const handleActionClick = useCallback(() => {
    setHasUserInteracted(true);
    const isValidInput = validate(value);
    if (isValidInput && actionButton?.onClick) {
      actionButton.onClick();
    }
  }, [value, validate, actionButton]);

  // 에러 표시 여부 결정
  const shouldShowError = hasUserInteracted && error && showErrorInline;
  const inputStatus = shouldShowError ? "error" : "";

  return (
    <div className={className}>
      <Row gutter={[8, 8]}>
        <Col span={inputSpan}>
          <Input
            ref={inputRef}
            value={value}
            onChange={handleInputChange}
            onKeyPress={handleKeyPress}
            onBlur={handleBlur}
            placeholder={placeholder}
            disabled={disabled}
            status={inputStatus}
            style={{
              backgroundColor: disabled
                ? "var(--bg-tertiary)"
                : "var(--bg-primary)",
              borderColor: shouldShowError
                ? "var(--red)"
                : "var(--border-color)",
              color: "var(--text-primary)",
            }}
          />
          {shouldShowError && (
            <Text
              style={{
                color: "var(--red)",
                fontSize: "12px",
                display: "block",
                marginTop: "4px",
                lineHeight: "1.2",
              }}
            >
              {error}
            </Text>
          )}
        </Col>

        {actionButton && (
          <Col span={buttonSpan}>
            <Button
              type="primary"
              icon={actionButton.icon}
              onClick={handleActionClick}
              disabled={
                disabled || actionButton.disabled || !value.trim() || !!error
              }
              block
              style={{
                background:
                  disabled || !!error
                    ? "var(--bg-tertiary)"
                    : "linear-gradient(135deg, var(--blue) 0%, #9333ea 100%)",
                borderColor: "var(--blue)",
                opacity: disabled || !!error ? 0.6 : 1,
              }}
            >
              {actionButton.text}
            </Button>
          </Col>
        )}
      </Row>
    </div>
  );
};
