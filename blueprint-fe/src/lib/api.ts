import type {
  AIMilestoneResponse,
  AIUsageInfo,
  ApiResponse,
  AuthResponse,
  CreateOrderRequest,
  CreateProjectWithMilestonesRequest,
  Expert,
  LoginRequest,
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
  Project,
  ProjectCategoryOption,
  ProjectStatus,
  ProjectStatusOption,
  RefreshTokenResponse,
  RegisterRequest,
  TokenExpiryResponse,
  Trade,
  UpdateProjectRequest,
  User,
  UserWallet,
} from "../types";

const API_BASE_URL = import.meta.env.VITE_API_URL || "/api/v1";

class ApiClient {
  private baseURL: string;
  private token: string | null = null;

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
  async login(credentials: LoginRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(credentials),
    });

    if (response.success && response.data?.token) {
      this.setToken(response.data.token);
    }

    return response;
  }

  async register(
    userData: RegisterRequest
  ): Promise<ApiResponse<AuthResponse>> {
    const response = await this.request<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify(userData),
    });

    if (response.success && response.data?.token) {
      this.setToken(response.data.token);
    }

    return response;
  }

  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<User>("/me");
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
    localStorage.removeItem("auth_token");
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

  // ⏰ 토큰 만료 확인
  async checkTokenExpiry(): Promise<ApiResponse<TokenExpiryResponse>> {
    return this.request("/auth/token-expiry") as Promise<
      ApiResponse<TokenExpiryResponse>
    >;
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
