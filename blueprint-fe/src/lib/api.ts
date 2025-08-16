import type {
  AIMilestoneResponse,
  AIUsageInfo,
  ActivityLogParams,
  ActivityLogResponse,
  ActivitySummaryResponse,
  ApiResponse,
  AuthResponse,
  CreateOrderRequest,
  CreateProjectWithMilestonesRequest,
  Expert,
  LogoutResponse,
  MarketStatusResponse,
  MentoringSession,
  Order,
  OrderBookResponse,
  OrderResponse,
  PaginatedResponse,
  Pagination,
  Path,
  PathPrediction,
  Position,
  ProfileResponse,
  Project,
  ProjectCategoryOption,
  ProjectStatus,
  ProjectStatusOption,
  RefreshTokenResponse,
  SettingsAggregateResponse,
  TokenExpiryResponse,
  Trade,
  UpdateProfileRequest,
  UpdateProjectRequest,
  User,
  UserProfileSettings,
  UserVerificationStatus,
  UserWallet,
} from "../types";

const API_BASE_URL = import.meta.env.VITE_API_URL || "/api/v1";

class ApiClient {
  private baseURL: string;
  private token: string | null;
  private isRedirecting = false; // 중복 리다이렉트 방지 플래그

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    this.token = localStorage.getItem("authToken");
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;

    const config: RequestInit = {
      headers: {
        "Content-Type": "application/json",
        ...(this.token && { Authorization: `Bearer ${this.token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      // 401 Unauthorized 에러 시 자동 로그아웃 처리
      if (response.status === 401) {
        // 이미 리다이렉트 중이면 중복 처리 방지
        if (this.isRedirecting) {
          throw new Error("세션이 만료되었습니다. 다시 로그인해주세요.");
        }

        this.isRedirecting = true;
        console.log("🔒 401 인증 오류 - 자동 로그아웃 처리");
        this.clearToken();

        // localStorage와 현재 페이지를 정리하고 홈으로 리다이렉트
        localStorage.removeItem("auth-storage"); // zustand persist 스토리지도 정리

        // 약간의 지연 후 리다이렉트하여 상태 업데이트 시간 확보
        setTimeout(() => {
          if (typeof window !== "undefined") {
            window.location.href = "/";
          }
        }, 100);

        throw new Error("세션이 만료되었습니다. 다시 로그인해주세요.");
      }

      if (!response.ok) {
        throw new Error(data.error || "API request failed");
      }

      return data;
    } catch (error) {
      console.error("API Error:", error);
      throw error;
    }
  }

  setToken(token: string) {
    this.token = token;
    localStorage.setItem("authToken", token);
  }

  getToken(): string | null {
    return this.token;
  }

  removeToken() {
    this.token = null;
    localStorage.removeItem("authToken");
  }

  // 인증 관련 API
  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<User>("/users/me");
  }

  // ===== Account Settings & Verification =====
  async getMySettings(): Promise<ApiResponse<SettingsAggregateResponse>> {
    return this.request<SettingsAggregateResponse>("/users/me/settings");
  }

  // 프로필 업데이트
  async updateMyProfile(
    data: UpdateProfileRequest
  ): Promise<ApiResponse<User>> {
    return this.request("/users/me/profile", {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  // 매직링크 생성
  async createMagicLink(data: {
    email: string;
  }): Promise<
    ApiResponse<{ code: string; expires_in: number; message: string }>
  > {
    return this.request("/auth/magic-link", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // 매직링크 인증
  async verifyMagicLink(data: {
    code: string;
  }): Promise<ApiResponse<{ token: string; user: User }>> {
    return this.request("/auth/verify-magic-link", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updatePreferences(
    data: Partial<UserProfileSettings>
  ): Promise<ApiResponse<UserProfileSettings>> {
    return this.request<UserProfileSettings>("/users/me/preferences", {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async requestVerifyEmail(): Promise<ApiResponse<{ message: string }>> {
    return this.request<{ message: string }>("/users/me/verify/email", {
      method: "POST",
    });
  }

  async requestVerifyPhone(): Promise<ApiResponse<{ message: string }>> {
    return this.request<{ message: string }>("/users/me/verify/phone", {
      method: "POST",
    });
  }

  // LinkedIn OAuth 연결 시작 (새로운 방식)
  async connectLinkedIn(): Promise<ApiResponse<{ auth_url: string }>> {
    return this.request<{ auth_url: string }>("/auth/linkedin/connect", {
      method: "GET",
    });
  }

  // 지원되는 OAuth 제공업체 목록 조회
  async getSupportedProviders(): Promise<
    ApiResponse<{ providers: string[]; count: number }>
  > {
    return this.request<{ providers: string[]; count: number }>(
      "/auth/providers",
      {
        method: "GET",
      }
    );
  }

  // 소셜 미디어 연결 (기존 방식 - 토큰 직접 전송)
  async connectProvider(
    provider: string,
    data: { access_token: string; profile_id?: string }
  ): Promise<ApiResponse<{ success: boolean; message: string }>> {
    return this.request(`/users/me/connect/${provider}`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async verifyWorkEmail(
    company: string
  ): Promise<ApiResponse<UserVerificationStatus>> {
    return this.request<UserVerificationStatus>("/users/me/verify/work-email", {
      method: "POST",
      body: JSON.stringify({ company }),
    });
  }

  async submitProfessionalDoc(): Promise<ApiResponse<{ status: string }>> {
    return this.request<{ status: string }>("/users/me/verify/professional", {
      method: "POST",
    });
  }

  async submitEducationDoc(): Promise<ApiResponse<{ status: string }>> {
    return this.request<{ status: string }>("/users/me/verify/education", {
      method: "POST",
    });
  }

  // Project 관리 API (마일스톤 포함)
  async createProject(
    projectData: CreateProjectWithMilestonesRequest
  ): Promise<ApiResponse<Project>> {
    return this.request("/projects", {
      method: "POST",
      body: JSON.stringify(projectData),
    });
  }

  // 🤖 AI 마일스톤 제안 받기
  async generateAIMilestones(
    projectData: CreateProjectWithMilestonesRequest
  ): Promise<ApiResponse<AIMilestoneResponse>> {
    return this.request("/ai/milestones", {
      method: "POST",
      body: JSON.stringify(projectData),
    });
  }

  // 📊 AI 사용 정보 조회
  async getAIUsageInfo(): Promise<ApiResponse<AIUsageInfo>> {
    return this.request("/ai/usage");
  }

  async getProjects(params?: {
    page?: number;
    limit?: number;
    category?: string;
    status?: string;
    sort?: string;
    order?: "asc" | "desc";
  }): Promise<ApiResponse<{ projects: Project[]; pagination: Pagination }>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append("page", params.page.toString());
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.category) queryParams.append("category", params.category);
    if (params?.status) queryParams.append("status", params.status);
    if (params?.sort) queryParams.append("sort", params.sort);
    if (params?.order) queryParams.append("order", params.order);

    const query = queryParams.toString();
    return this.request(`/projects${query ? `?${query}` : ""}`);
  }

  async getProject(id: number): Promise<ApiResponse<Project>> {
    return this.request(`/projects/${id}`);
  }

  async updateProject(
    id: number,
    projectData: UpdateProjectRequest
  ): Promise<ApiResponse<Project>> {
    return this.request(`/projects/${id}`, {
      method: "PUT",
      body: JSON.stringify(projectData),
    });
  }

  async deleteProject(id: number): Promise<ApiResponse<null>> {
    return this.request(`/projects/${id}`, {
      method: "DELETE",
    });
  }

  async updateProjectStatus(
    id: number,
    status: ProjectStatus
  ): Promise<ApiResponse<{ status: ProjectStatus }>> {
    return this.request(`/projects/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
  }

  async getProjectCategories(): Promise<ApiResponse<ProjectCategoryOption[]>> {
    return this.request("/project-categories");
  }

  async getProjectStatuses(): Promise<ApiResponse<ProjectStatusOption[]>> {
    return this.request("/project-statuses");
  }

  // 🔐 로그아웃
  async logout(): Promise<ApiResponse<LogoutResponse>> {
    const response = (await this.request("/auth/logout", {
      method: "POST",
    })) as ApiResponse<LogoutResponse>;

    // 로그아웃 성공 시 토큰 제거
    if (response.success) {
      this.clearToken();
    }

    return response;
  }

  // 🔄 토큰 갱신
  async refreshToken(): Promise<ApiResponse<RefreshTokenResponse>> {
    const response = (await this.request("/auth/refresh", {
      method: "POST",
    })) as ApiResponse<RefreshTokenResponse>;

    // 토큰 갱신 성공 시 새 토큰 저장
    if (response.success && response.data && "token" in response.data) {
      this.setToken(response.data.token);
    }

    return response;
  }

  // 토큰 제거 메서드
  clearToken(): void {
    localStorage.removeItem("authToken");
    this.token = null;
  }

  // Google OAuth 로그인 URL 가져오기
  async getGoogleAuthUrl(): Promise<ApiResponse<{ auth_url: string }>> {
    return this.request<{ auth_url: string }>("/auth/google/login");
  }

  // Google OAuth 콜백 처리
  async handleGoogleCallback(code: string): Promise<ApiResponse<AuthResponse>> {
    return this.request<AuthResponse>(`/auth/google/callback?code=${code}`);
  }

  // 경로 관리 API
  async getPaths(projectId: string): Promise<ApiResponse<Path[]>> {
    return this.request<Path[]>(`/projects/${projectId}/paths`);
  }

  async getPath(id: string): Promise<ApiResponse<Path>> {
    return this.request<Path>(`/paths/${id}`);
  }

  async createPath(pathData: Partial<Path>): Promise<ApiResponse<Path>> {
    return this.request<Path>("/paths", {
      method: "POST",
      body: JSON.stringify(pathData),
    });
  }

  // 예측 마켓 API
  async getPredictions(pathId: string): Promise<ApiResponse<PathPrediction[]>> {
    return this.request<PathPrediction[]>(`/paths/${pathId}/predictions`);
  }

  async createPrediction(
    predictionData: Partial<PathPrediction>
  ): Promise<ApiResponse<PathPrediction>> {
    return this.request<PathPrediction>("/predictions", {
      method: "POST",
      body: JSON.stringify(predictionData),
    });
  }

  // 전문가 관리 API
  async getExperts(params?: {
    specialty?: string;
    minRating?: number;
    page?: number;
    limit?: number;
  }): Promise<ApiResponse<PaginatedResponse<Expert>>> {
    const queryParams = new URLSearchParams();
    if (params?.specialty) queryParams.append("specialty", params.specialty);
    if (params?.minRating)
      queryParams.append("minRating", params.minRating.toString());
    if (params?.page) queryParams.append("page", params.page.toString());
    if (params?.limit) queryParams.append("limit", params.limit.toString());

    return this.request<PaginatedResponse<Expert>>(`/experts?${queryParams}`);
  }

  async getExpert(id: string): Promise<ApiResponse<Expert>> {
    return this.request<Expert>(`/experts/${id}`);
  }

  // 멘토링 세션 API
  async getMentoringSessions(params?: {
    pathId?: string;
    status?: string;
    page?: number;
    limit?: number;
  }): Promise<ApiResponse<PaginatedResponse<MentoringSession>>> {
    const queryParams = new URLSearchParams();
    if (params?.pathId) queryParams.append("pathId", params.pathId);
    if (params?.status) queryParams.append("status", params.status);
    if (params?.page) queryParams.append("page", params.page.toString());
    if (params?.limit) queryParams.append("limit", params.limit.toString());

    return this.request<PaginatedResponse<MentoringSession>>(
      `/mentoring?${queryParams}`
    );
  }

  async createMentoringSession(
    sessionData: Partial<MentoringSession>
  ): Promise<ApiResponse<MentoringSession>> {
    return this.request<MentoringSession>("/mentoring", {
      method: "POST",
      body: JSON.stringify(sessionData),
    });
  }

  // 📊 마켓 데이터 API
  async getMilestoneMarket(
    milestoneId: number
  ): Promise<ApiResponse<MarketStatusResponse>> {
    return this.request(`/milestones/${milestoneId}/market`);
  }

  // 🎯 AI 마일스톤 생성 API

  // 헬스 체크
  async healthCheck(): Promise<{ status: string; message: string }> {
    const response = await fetch(
      `${this.baseURL.replace("/api/v1", "")}/health`
    );
    return response.json();
  }

  // ✅ 토큰 만료 확인
  async checkTokenExpiry(): Promise<ApiResponse<TokenExpiryResponse>> {
    return this.request("/auth/token-expiry") as Promise<
      ApiResponse<TokenExpiryResponse>
    >;
  }

  // ================================
  // 프로필 관련 API
  // ================================

  // 사용자 프로필 조회 (목데이터와 동일한 구조)
  async getUserProfile(
    username: string
  ): Promise<ApiResponse<ProfileResponse>> {
    return this.request(`/users/${username}/profile`);
  }

  // 사용자 활동 로그 조회
  async getUserActivities(
    params?: ActivityLogParams
  ): Promise<ApiResponse<ActivityLogResponse>> {
    const queryParams = new URLSearchParams();

    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());
    if (params?.types) {
      params.types.forEach((type) => queryParams.append("types", type));
    }
    if (params?.start_date) queryParams.append("start_date", params.start_date);
    if (params?.end_date) queryParams.append("end_date", params.end_date);

    const query = queryParams.toString();
    const endpoint = query
      ? `/users/me/activities?${query}`
      : "/users/me/activities";

    return this.request(endpoint);
  }

  // 활동 요약 조회 (대시보드용)
  async getActivitySummary(): Promise<ApiResponse<ActivitySummaryResponse>> {
    return this.request("/users/me/activities/summary");
  }

  // ================================
  // P2P 거래 관련 API (폴리마켓 스타일)
  // ================================

  // 주문 생성 (매수/매도)
  async createOrder(
    data: CreateOrderRequest
  ): Promise<ApiResponse<OrderResponse>> {
    return this.request("/orders", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // 호가창 조회
  async getOrderBook(
    milestoneId: number,
    optionId: string
  ): Promise<ApiResponse<OrderBookResponse>> {
    return this.request(`/milestones/${milestoneId}/orderbook/${optionId}`);
  }

  // 내 주문 내역 조회
  async getMyOrders(params?: {
    status?: string;
    milestone_id?: number;
    limit?: number;
  }): Promise<ApiResponse<Order[]>> {
    const queryParams = new URLSearchParams();
    if (params?.status) queryParams.append("status", params.status);
    if (params?.milestone_id)
      queryParams.append("milestone_id", params.milestone_id.toString());
    if (params?.limit) queryParams.append("limit", params.limit.toString());

    const url = `/orders/my${
      queryParams.toString() ? `?${queryParams.toString()}` : ""
    }`;
    return this.request(url);
  }

  // 내 거래 내역 조회
  async getMyTrades(params?: {
    milestone_id?: number;
    limit?: number;
  }): Promise<ApiResponse<Trade[]>> {
    const queryParams = new URLSearchParams();
    if (params?.milestone_id)
      queryParams.append("milestone_id", params.milestone_id.toString());
    if (params?.limit) queryParams.append("limit", params.limit.toString());

    const url = `/trades/my${
      queryParams.toString() ? `?${queryParams.toString()}` : ""
    }`;
    return this.request(url);
  }

  // 내 포지션 조회
  async getMyPositions(milestoneId?: number): Promise<ApiResponse<Position[]>> {
    const url = milestoneId
      ? `/positions/my?milestone_id=${milestoneId}`
      : "/positions/my";
    return this.request(url);
  }

  // 특정 포지션 조회
  async getMilestonePosition(
    milestoneId: number,
    optionId: string
  ): Promise<ApiResponse<Position>> {
    return this.request(`/milestones/${milestoneId}/position/${optionId}`);
  }

  // 사용자 지갑 조회 (임시)
  async getUserWallet(): Promise<ApiResponse<UserWallet>> {
    return this.request("/wallet");
  }

  // 주문 취소
  async cancelOrder(orderId: number): Promise<ApiResponse<Order>> {
    return this.request(`/orders/${orderId}`, { method: "DELETE" });
  }

  // 최근 거래 내역 조회 (공개)
  async getRecentTrades(
    milestoneId: number,
    optionId: string,
    limit: number = 50
  ): Promise<ApiResponse<Trade[]>> {
    return this.request(
      `/milestones/${milestoneId}/trades/${optionId}?limit=${limit}`
    );
  }

  // 가격 히스토리 조회
  async getPriceHistory(
    milestoneId: number,
    optionId: string,
    interval: string = "1h",
    limit: number = 100
  ): Promise<ApiResponse<{ data: object[]; interval: string; count: number }>> {
    return this.request<{ data: object[]; interval: string; count: number }>(
      `/milestones/${milestoneId}/price-history/${optionId}?interval=${interval}&limit=${limit}`
    );
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
export default apiClient;
