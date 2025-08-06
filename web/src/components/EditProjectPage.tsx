import React from "react";

const EditProjectPage: React.FC = () => {
  return (
    <div
      style={{
        padding: "40px",
        textAlign: "center",
        backgroundColor: "var(--bg-primary)",
        color: "var(--text-primary)",
        minHeight: "100vh",
      }}
    >
      <h2 style={{ fontSize: "24px", marginBottom: "16px" }}>
        🚧 프로젝트 수정 페이지 업데이트 중
      </h2>
      <p style={{ fontSize: "16px", color: "var(--text-secondary)" }}>
        프로젝트 태그 시스템이 개선되고 있습니다. 곧 새로운 형태로
        돌아오겠습니다.
      </p>
      <div style={{ marginTop: "24px" }}>
        <button
          style={{
            padding: "12px 24px",
            backgroundColor: "var(--blue)",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
          }}
          onClick={() => window.history.back()}
        >
          이전 페이지로 돌아가기
        </button>
      </div>
    </div>
  );
};

export default EditProjectPage;
