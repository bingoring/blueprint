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
  MentoringSession
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

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

  async logout(): Promise<void> {
    this.removeToken();
  }

  // 목표 관리 API
  async getGoals(params?: {
    page?: number;
    limit?: number;
    category?: string;
    status?: string;
  }): Promise<ApiResponse<PaginatedResponse<Goal>>> {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.category) queryParams.append('category', params.category);
    if (params?.status) queryParams.append('status', params.status);

    return this.request<PaginatedResponse<Goal>>(`/goals?${queryParams}`);
  }

  async getGoal(id: string): Promise<ApiResponse<Goal>> {
    return this.request<Goal>(`/goals/${id}`);
  }

  async createGoal(goalData: Partial<Goal>): Promise<ApiResponse<Goal>> {
    return this.request<Goal>('/goals', {
      method: 'POST',
      body: JSON.stringify(goalData),
    });
  }

  async updateGoal(id: string, goalData: Partial<Goal>): Promise<ApiResponse<Goal>> {
    return this.request<Goal>(`/goals/${id}`, {
      method: 'PUT',
      body: JSON.stringify(goalData),
    });
  }

  async deleteGoal(id: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/goals/${id}`, {
      method: 'DELETE',
    });
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
