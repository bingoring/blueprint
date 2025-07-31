import React, { useState, useEffect } from 'react';
import { apiClient } from '../lib/api';
import type { Goal, GoalCategoryOption, GoalStatusOption } from '../types';
// import GoalCard from '../components/goals/GoalCard';
// import CreateGoalModal from '../components/goals/CreateGoalModal';

const GoalsPage: React.FC = () => {
  const [goals, setGoals] = useState<Goal[]>([]);
  const [categories, setCategories] = useState<GoalCategoryOption[]>([]);
  const [statuses, setStatuses] = useState<GoalStatusOption[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedStatus, setSelectedStatus] = useState<string>('all');

  // 목표 목록 조회
  const fetchGoals = async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {};
      if (selectedCategory !== 'all') params.category = selectedCategory;
      if (selectedStatus !== 'all') params.status = selectedStatus;

      const response = await apiClient.getGoals(params);
      if (response.success && response.data) {
        setGoals(response.data.goals);
      }
    } catch (error) {
      console.error('목표 목록 조회 실패:', error);
    } finally {
      setLoading(false);
    }
  };

  // 카테고리 및 상태 옵션 조회
  const fetchMetadata = async () => {
    try {
      const [categoriesRes, statusesRes] = await Promise.all([
        apiClient.getGoalCategories(),
        apiClient.getGoalStatuses(),
      ]);

      if (categoriesRes.success && categoriesRes.data) {
        setCategories(categoriesRes.data);
      }
      if (statusesRes.success && statusesRes.data) {
        setStatuses(statusesRes.data);
      }
    } catch (error) {
      console.error('메타데이터 조회 실패:', error);
    }
  };

  useEffect(() => {
    fetchMetadata();
  }, []);

  useEffect(() => {
    fetchGoals();
  }, [selectedCategory, selectedStatus]);

  const handleGoalCreated = () => {
    setIsCreateModalOpen(false);
    fetchGoals();
  };

  const handleGoalUpdated = () => {
    fetchGoals();
  };

  const handleGoalDeleted = () => {
    fetchGoals();
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* 헤더 */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">내 목표</h1>
            <p className="text-gray-600 mt-1">나의 목표들을 관리하고 진행 상황을 확인하세요</p>
          </div>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors"
          >
            + 새 목표 만들기
          </button>
        </div>

        {/* 필터링 */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-8">
          <div className="flex flex-wrap gap-4">
            {/* 카테고리 필터 */}
            <div className="flex-1 min-w-48">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                카테고리
              </label>
              <select
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">전체 카테고리</option>
                {categories.map((category) => (
                  <option key={category.value} value={category.value}>
                    {category.icon} {category.label}
                  </option>
                ))}
              </select>
            </div>

            {/* 상태 필터 */}
            <div className="flex-1 min-w-48">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                상태
              </label>
              <select
                value={selectedStatus}
                onChange={(e) => setSelectedStatus(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="all">전체 상태</option>
                {statuses.map((status) => (
                  <option key={status.value} value={status.value}>
                    {status.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* 목표 목록 */}
        {loading ? (
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span className="ml-3 text-gray-600">목표를 불러오는 중...</span>
          </div>
        ) : goals.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-gray-400 text-6xl mb-4">🎯</div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              아직 목표가 없습니다
            </h3>
            <p className="text-gray-600 mb-6">
              첫 번째 목표를 만들어서 여정을 시작해보세요!
            </p>
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors"
            >
              목표 만들기
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {goals.map((goal) => (
              <div key={goal.id} className="bg-white rounded-lg shadow-sm p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-2">{goal.title}</h3>
                <p className="text-gray-600 text-sm mb-4">{goal.description}</p>
                <div className="flex justify-between items-center">
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                    {goal.category}
                  </span>
                  <span className="px-2 py-1 bg-gray-100 text-gray-800 text-xs rounded-full">
                    {goal.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* 목표 생성 모달 - 임시 구현 */}
        {isCreateModalOpen && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
              <h2 className="text-xl font-bold mb-4">새 목표 만들기</h2>
              <p className="text-gray-600 mb-4">목표 생성 기능을 개발 중입니다...</p>
              <button
                onClick={() => setIsCreateModalOpen(false)}
                className="w-full bg-gray-500 text-white py-2 rounded-lg"
              >
                닫기
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default GoalsPage;
