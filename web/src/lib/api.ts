import type {
  ApiResponse,
  PaginatedResponse,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  User,
  Goal,
  Path,
  PathPrediction,
  Expert,
  MentoringSession,
  CreateGoalRequest,
  UpdateGoalRequest,
  GoalStatus,
  GoalCategoryOption,
  GoalStatusOption,
  Pagination,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

class ApiClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    this.token = localStorage.getItem('authToken');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;

    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(this.token && { Authorization: `Bearer ${this.token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'API request failed');
      }

      return data;
    } catch (error) {
      console.error('API Error:', error);
      throw error;
    }
  }

  setToken(token: string) {
    this.token = token;
    localStorage.setItem('authToken', token);
  }

  getToken(): string | null {
    return this.token;
  }

  removeToken() {
    this.token = null;
    localStorage.removeItem('authToken');
  }

  // 인증 관련 API
  async login(credentials: LoginRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });

    if (response.success && response.data?.token) {
      this.setToken(response.data.token);
    }

    return response;
  }

  async register(userData: RegisterRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    if (response.success && response.data?.token) {
      this.setToken(response.data.token);
    }

    return response;
  }

  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<User>('/me');
  }

  // Goal 관리 API
  async createGoal(goalData: CreateGoalRequest): Promise<ApiResponse<Goal>> {
    return this.request('/goals', {
      method: 'POST',
      body: JSON.stringify(goalData),
    });
  }

  async getGoals(params?: {
    page?: number;
    limit?: number;
    category?: string;
    status?: string;
    sort?: string;
    order?: 'asc' | 'desc';
  }): Promise<ApiResponse<{ goals: Goal[]; pagination: Pagination }>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.category) queryParams.append('category', params.category);
    if (params?.status) queryParams.append('status', params.status);
    if (params?.sort) queryParams.append('sort', params.sort);
    if (params?.order) queryParams.append('order', params.order);

    const query = queryParams.toString();
    return this.request(`/goals${query ? `?${query}` : ''}`);
  }

  async getGoal(id: number): Promise<ApiResponse<Goal>> {
    return this.request(`/goals/${id}`);
  }

  async updateGoal(id: number, goalData: UpdateGoalRequest): Promise<ApiResponse<Goal>> {
    return this.request(`/goals/${id}`, {
      method: 'PUT',
      body: JSON.stringify(goalData),
    });
  }

  async deleteGoal(id: number): Promise<ApiResponse<null>> {
    return this.request(`/goals/${id}`, {
      method: 'DELETE',
    });
  }

  async updateGoalStatus(id: number, status: GoalStatus): Promise<ApiResponse<{ status: GoalStatus }>> {
    return this.request(`/goals/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  }

  async getGoalCategories(): Promise<ApiResponse<GoalCategoryOption[]>> {
    return this.request('/goal-categories');
  }

  async getGoalStatuses(): Promise<ApiResponse<GoalStatusOption[]>> {
    return this.request('/goal-statuses');
  }

  async logout(): Promise<void> {
    this.removeToken();
  }

  // Google OAuth 로그인 URL 가져오기
  async getGoogleAuthUrl(): Promise<ApiResponse<{ auth_url: string }>> {
    return this.request<{ auth_url: string }>('/auth/google/login');
  }

  // Google OAuth 콜백 처리
  async handleGoogleCallback(code: string): Promise<ApiResponse<AuthResponse>> {
    return this.request<AuthResponse>(`/auth/google/callback?code=${code}`);
  }


  // 경로 관리 API
  async getPaths(goalId: string): Promise<ApiResponse<Path[]>> {
    return this.request<Path[]>(`/goals/${goalId}/paths`);
  }

  async getPath(id: string): Promise<ApiResponse<Path>> {
    return this.request<Path>(`/paths/${id}`);
  }

  async createPath(pathData: Partial<Path>): Promise<ApiResponse<Path>> {
    return this.request<Path>('/paths', {
      method: 'POST',
      body: JSON.stringify(pathData),
    });
  }

  // 예측 마켓 API
  async getPredictions(pathId: string): Promise<ApiResponse<PathPrediction[]>> {
    return this.request<PathPrediction[]>(`/paths/${pathId}/predictions`);
  }

  async createPrediction(predictionData: Partial<PathPrediction>): Promise<ApiResponse<PathPrediction>> {
    return this.request<PathPrediction>('/predictions', {
      method: 'POST',
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
    if (params?.specialty) queryParams.append('specialty', params.specialty);
    if (params?.minRating) queryParams.append('minRating', params.minRating.toString());
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());

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
    if (params?.pathId) queryParams.append('pathId', params.pathId);
    if (params?.status) queryParams.append('status', params.status);
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    return this.request<PaginatedResponse<MentoringSession>>(`/mentoring?${queryParams}`);
  }

  async createMentoringSession(sessionData: Partial<MentoringSession>): Promise<ApiResponse<MentoringSession>> {
    return this.request<MentoringSession>('/mentoring', {
      method: 'POST',
      body: JSON.stringify(sessionData),
    });
  }

  // 헬스 체크
  async healthCheck(): Promise<{ status: string; message: string }> {
    const response = await fetch(`${this.baseURL.replace('/api/v1', '')}/health`);
    return response.json();
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
export default apiClient;
