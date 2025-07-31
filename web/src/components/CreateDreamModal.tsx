import { useState, useEffect } from 'react';
import { apiClient } from '../lib/api';
import { useAuthStore } from '../stores/useAuthStore';
import type {
  CreateDreamRequest,
  CreateMilestoneRequest,
  GoalCategoryOption,
  GoalCategory,
  Goal
} from '../types';

interface CreateDreamModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: (dream: Goal) => void;
  onLoginRequired?: () => void; // 로그인이 필요할 때 호출
}

export default function CreateDreamModal({ isOpen, onClose, onSuccess, onLoginRequired }: CreateDreamModalProps) {
  const { isAuthenticated } = useAuthStore();
  const [categories, setCategories] = useState<GoalCategoryOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 폼 상태
  const [formData, setFormData] = useState<Omit<CreateDreamRequest, 'milestones'>>({
    title: '',
    description: '',
    category: 'career',
    target_date: '',
    budget: 0,
    priority: 3,
    is_public: true,
    tags: [],
  });

  // 마일스톤 상태 (최대 5개)
  const [milestones, setMilestones] = useState<CreateMilestoneRequest[]>([
    { title: '', description: '', order: 1, target_date: '' }
  ]);

  // 현재 단계 (1: 꿈 정보, 2: 마일스톤 설정)
  const [currentStep, setCurrentStep] = useState(1);

  // 카테고리 로드
  useEffect(() => {
    if (isOpen) {
      loadCategories();
    }
  }, [isOpen]);

  const loadCategories = async () => {
    try {
      const response = await apiClient.getGoalCategories();
      if (response.success && response.data) {
        setCategories(response.data);
      }
    } catch (err) {
      console.error('카테고리 로드 실패:', err);
    }
  };

  const addMilestone = () => {
    if (milestones.length < 5) {
      setMilestones([
        ...milestones,
        {
          title: '',
          description: '',
          order: milestones.length + 1,
          target_date: ''
        }
      ]);
    }
  };

  const removeMilestone = (index: number) => {
    if (milestones.length > 1) {
      const newMilestones = milestones.filter((_, i) => i !== index);
      // order 재정렬
      const reorderedMilestones = newMilestones.map((milestone, i) => ({
        ...milestone,
        order: i + 1
      }));
      setMilestones(reorderedMilestones);
    }
  };

  const updateMilestone = (index: number, field: keyof CreateMilestoneRequest, value: string) => {
    const newMilestones = [...milestones];
    newMilestones[index] = { ...newMilestones[index], [field]: value };
    setMilestones(newMilestones);
  };

    const handleSubmit = async () => {
    // 인증 체크
    if (!isAuthenticated) {
      onLoginRequired?.();
      onClose();
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // 빈 마일스톤 제거
      const validMilestones = milestones.filter(m => m.title.trim() !== '');

      if (validMilestones.length === 0) {
        setError('최소 1개의 마일스톤이 필요합니다');
        setLoading(false);
        return;
      }

      const dreamData: CreateDreamRequest = {
        ...formData,
        milestones: validMilestones
      };

      const response = await apiClient.createDream(dreamData);

      if (response.success && response.data) {
        onSuccess?.(response.data);
        onClose();
        resetForm();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '꿈 등록에 실패했습니다');
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      title: '',
      description: '',
      category: 'career',
      target_date: '',
      budget: 0,
      priority: 3,
      is_public: true,
      tags: [],
    });
    setMilestones([
      { title: '', description: '', order: 1, target_date: '' }
    ]);
    setCurrentStep(1);
    setError(null);
  };

  const nextStep = () => {
    if (formData.title.trim() === '') {
      setError('꿈의 제목을 입력해주세요');
      return;
    }
    setCurrentStep(2);
    setError(null);
  };

  const prevStep = () => {
    setCurrentStep(1);
    setError(null);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          {/* 헤더 */}
          <div className="flex justify-between items-center mb-6">
            <div>
              <h2 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                ✨ 새로운 꿈 등록하기
              </h2>
              <p className="text-gray-600 mt-1">
                {currentStep === 1 ? '당신의 꿈을 들려주세요' : '성공을 위한 마일스톤을 설정해보세요'}
              </p>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 text-2xl"
            >
              ✕
            </button>
          </div>

          {/* 프로그레스 바 */}
          <div className="mb-8">
            <div className="flex items-center">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                currentStep >= 1 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500'
              }`}>
                1
              </div>
              <div className={`flex-1 h-2 mx-2 rounded ${
                currentStep >= 2 ? 'bg-blue-600' : 'bg-gray-200'
              }`}></div>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                currentStep >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500'
              }`}>
                2
              </div>
            </div>
            <div className="flex justify-between mt-2 text-sm text-gray-600">
              <span>꿈 정보</span>
              <span>마일스톤</span>
            </div>
          </div>

          {/* 에러 메시지 */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
              {error}
            </div>
          )}

          {/* Step 1: 꿈 정보 */}
          {currentStep === 1 && (
            <div className="space-y-6">
              {/* 제목 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  🎯 꿈의 제목 *
                </label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  placeholder="예: 3년 내 카페 사장되기"
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* 카테고리 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  🏷️ 카테고리
                </label>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {categories.map((category) => (
                    <button
                      key={category.value}
                      onClick={() => setFormData({ ...formData, category: category.value as GoalCategory })}
                      className={`p-4 text-left border rounded-lg transition-colors ${
                        formData.category === category.value
                          ? 'border-blue-500 bg-blue-50 text-blue-700'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <div className="font-medium">{category.label}</div>
                      {category.description && (
                        <div className="text-sm text-gray-500 mt-1">{category.description}</div>
                      )}
                    </button>
                  ))}
                </div>
              </div>

              {/* 설명 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  📝 꿈에 대한 설명
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="당신의 꿈에 대해 자세히 설명해보세요..."
                  rows={4}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* 목표 날짜 & 예산 */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    📅 목표 날짜
                  </label>
                  <input
                    type="date"
                    value={formData.target_date}
                    onChange={(e) => setFormData({ ...formData, target_date: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    💰 예산 (만원)
                  </label>
                  <input
                    type="number"
                    value={formData.budget}
                    onChange={(e) => setFormData({ ...formData, budget: parseInt(e.target.value) || 0 })}
                    placeholder="0"
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* 우선순위 & 공개 설정 */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    ⭐ 우선순위
                  </label>
                  <select
                    value={formData.priority}
                    onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value={1}>1 (낮음)</option>
                    <option value={2}>2</option>
                    <option value={3}>3 (보통)</option>
                    <option value={4}>4</option>
                    <option value={5}>5 (높음)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    👥 공개 설정
                  </label>
                  <div className="flex items-center space-x-4 mt-3">
                    <label className="flex items-center">
                      <input
                        type="radio"
                        checked={formData.is_public}
                        onChange={() => setFormData({ ...formData, is_public: true })}
                        className="mr-2"
                      />
                      공개 (다른 사용자가 볼 수 있음)
                    </label>
                    <label className="flex items-center">
                      <input
                        type="radio"
                        checked={!formData.is_public}
                        onChange={() => setFormData({ ...formData, is_public: false })}
                        className="mr-2"
                      />
                      비공개
                    </label>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Step 2: 마일스톤 설정 */}
          {currentStep === 2 && (
            <div className="space-y-6">
              <div className="text-center mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                  🛤️ 마일스톤 설정
                </h3>
                <p className="text-gray-600">
                  꿈을 이루기 위한 중간 단계들을 설정해보세요 (최대 5개)
                </p>
              </div>

              {milestones.map((milestone, index) => (
                <div key={index} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-3">
                    <span className="text-sm font-medium text-gray-700">
                      마일스톤 {index + 1}
                    </span>
                    {milestones.length > 1 && (
                      <button
                        onClick={() => removeMilestone(index)}
                        className="text-red-500 hover:text-red-700 text-sm"
                      >
                        삭제
                      </button>
                    )}
                  </div>

                  <div className="space-y-3">
                    <input
                      type="text"
                      value={milestone.title}
                      onChange={(e) => updateMilestone(index, 'title', e.target.value)}
                      placeholder="마일스톤 제목 (예: 바리스타 자격증 취득)"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />

                    <textarea
                      value={milestone.description}
                      onChange={(e) => updateMilestone(index, 'description', e.target.value)}
                      placeholder="세부 설명..."
                      rows={2}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />

                    <input
                      type="date"
                      value={milestone.target_date}
                      onChange={(e) => updateMilestone(index, 'target_date', e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>
                </div>
              ))}

              {milestones.length < 5 && (
                <button
                  onClick={addMilestone}
                  className="w-full py-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-blue-500 hover:text-blue-600 transition-colors"
                >
                  + 마일스톤 추가 ({milestones.length}/5)
                </button>
              )}
            </div>
          )}

          {/* 액션 버튼 */}
          <div className="flex justify-between mt-8">
            {currentStep === 1 ? (
              <div className="flex space-x-3 ml-auto">
                <button
                  onClick={onClose}
                  className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
                >
                  취소
                </button>
                <button
                  onClick={nextStep}
                  className="px-6 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-700 hover:to-purple-700"
                >
                  다음 단계
                </button>
              </div>
            ) : (
              <div className="flex justify-between w-full">
                <button
                  onClick={prevStep}
                  className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
                >
                  이전 단계
                </button>
                <button
                  onClick={handleSubmit}
                  disabled={loading}
                  className={`px-6 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-700 hover:to-purple-700 ${
                    loading ? 'opacity-50 cursor-not-allowed' : ''
                  }`}
                >
                  {loading ? '등록 중...' : '✨ 꿈 등록하기'}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
