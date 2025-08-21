import { ArrowLeftOutlined } from "@ant-design/icons";
import { Breadcrumb, message, Spin, Tag, Typography } from "antd";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import OrderBook from "../components/OrderBook";
import ThemeToggle from "../components/ThemeToggle";
import TradingPanel from "../components/TradingPanel";
import { apiClient } from "../lib/api";
import type {
  CreateOrderRequest,
  Milestone,
  OrderBook as OrderBookType,
  OrderSide,
  OrderType,
  Project,
  Trade,
  UserWallet,
} from "../types";

const { Title, Text } = Typography;

interface ChartDataPoint {
  time: string;
  timestamp: number;
  price: number;
  volume: number;
}

interface SSEMessage {
  type:
    | "connection"
    | "market_update"
    | "price_change"
    | "order_update"
    | "orderbook_update"
    | "trade"
    | "ping"
    | "error";
  data?: unknown;
  milestone_id?: number;
  status?: string;
  message?: string;
  timestamp: number;
}

const PolymarketTradingPage: React.FC = () => {
  const { projectId, milestoneId } = useParams<{
    projectId: string;
    milestoneId: string;
  }>();
  const navigate = useNavigate();

  // Core data
  const [project, setProject] = useState<Project | null>(null);
  const [milestone, setMilestone] = useState<Milestone | null>(null);
  const [loading, setLoading] = useState(true);
  const [userWallet, setUserWallet] = useState<UserWallet | null>(null);

  // Trading state
  const [selectedOption, setSelectedOption] = useState<string>("");
  const [orderBook, setOrderBook] = useState<OrderBookType | null>(null);
  const [recentTrades, setRecentTrades] = useState<Trade[]>([]);
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [orderLoading, setOrderLoading] = useState(false);
  const [currentPrice, setCurrentPrice] = useState<number>(0.5); // Current market price for selected option

  // SSE connection
  const [isSSEConnected, setIsSSEConnected] = useState(false);
  const sseRef = useRef<EventSource | null>(null);

  // Load initial data
  useEffect(() => {
    if (!projectId || !milestoneId) {
      navigate("/");
      return;
    }

    loadInitialData();
  }, [projectId, milestoneId, navigate]);

  // SSE connection effect
  useEffect(() => {
    if (milestone && selectedOption) {
      connectSSE();
      return () => {
        const eventSource = sseRef.current;
        if (eventSource) {
          eventSource.close();
          setIsSSEConnected(false);
        }
      };
    }
  }, [milestone, selectedOption]);

  const loadInitialData = async () => {
    try {
      setLoading(true);

      // Load project and milestone
      const [projectRes, walletRes] = await Promise.all([
        apiClient.getProject(parseInt(projectId!)),
        apiClient.getUserWallet().catch(() => null), // Optional
      ]);

      if (projectRes.success && projectRes.data) {
        setProject(projectRes.data);
        const foundMilestone = projectRes.data.milestones?.find(
          (m) => m.id === parseInt(milestoneId!)
        );

        if (foundMilestone) {
          setMilestone(foundMilestone);

          // Set default option to success
          const firstOption = "success";
          setSelectedOption(firstOption);
          console.log("🎯 Selected option set to:", firstOption);

          // Load market data for the selected option
          setTimeout(() => {
            console.log("🔄 Loading market data for option:", firstOption);
            loadMarketDataForOption(firstOption);
          }, 100);
        }
      }

      if (walletRes?.success && walletRes.data) {
        setUserWallet(walletRes.data);
      }
    } catch (error) {
      console.error("Failed to load initial data:", error);
      message.error("Failed to load data");
    } finally {
      setLoading(false);
    }
  };

  const loadMarketData = async () => {
    if (!milestoneId || !selectedOption) return;
    await loadMarketDataForOption(selectedOption);
  };

  const loadMarketDataForOption = async (optionId: string) => {
    if (!milestoneId || !optionId) return;

    try {
      console.log("📊 Loading market data for:", optionId);
      const [orderBookRes, tradesRes, priceHistoryRes] = await Promise.all([
        apiClient
          .getOrderBook(parseInt(milestoneId), optionId)
          .catch(() => ({ success: false, data: null })),
        apiClient
          .getRecentTrades(parseInt(milestoneId), optionId, 20)
          .catch(() => ({ success: false, data: null })),
        apiClient
          .getPriceHistory(parseInt(milestoneId), optionId, "1h", 24)
          .catch(() => ({ success: false, data: null })),
      ]);

      if (orderBookRes.success && orderBookRes.data) {
        setOrderBook(orderBookRes.data.order_book);

        // Extract current market price from order book or price history
        const orderBookData = orderBookRes.data.order_book;
        if (
          orderBookData &&
          (orderBookData.asks.length > 0 || orderBookData.bids.length > 0)
        ) {
          // Use best ask/bid average as current price
          const bestAsk = orderBookData.asks[0]?.price || 0.5;
          const bestBid = orderBookData.bids[0]?.price || 0.5;
          const marketPrice = (bestAsk + bestBid) / 2;
          setCurrentPrice(marketPrice);
          console.log(
            "💰 Market price updated from order book:",
            Math.round(marketPrice * 100),
            "¢"
          );
        }
      }

      if (tradesRes.success && tradesRes.data) {
        const trades = Array.isArray(tradesRes.data)
          ? tradesRes.data
          : (tradesRes.data as { trades: Trade[] }).trades || [];
        setRecentTrades(trades);
      }

      // 가격 히스토리에서 차트 데이터 생성 (우선순위 1)
      if (priceHistoryRes.success && priceHistoryRes.data?.data) {
        const historyData = priceHistoryRes.data.data;
        if (historyData.length > 0) {
          const chartPoints: ChartDataPoint[] = historyData.map((point) => {
            const pointData = point as Record<string, unknown>;
            return {
              time: new Date(pointData.bucket as string).toLocaleTimeString(),
              timestamp: new Date(pointData.bucket as string).getTime(),
              price:
                (pointData.close as number) ||
                (pointData.open as number) ||
                0.5,
              volume: (pointData.volume as number) || 0,
            };
          });
          setChartData(chartPoints);

          // Set current price from latest data point
          const latestPrice = chartPoints[chartPoints.length - 1]?.price || 0.5;
          setCurrentPrice(latestPrice);
          console.log(
            "📈 Chart data loaded from price history:",
            chartPoints.length,
            "points"
          );
          console.log(
            "💰 Current price set from history:",
            Math.round(latestPrice * 100),
            "¢"
          );
        }
      }
      // Fallback: 거래 데이터에서 차트 생성 (우선순위 2)
      else if (tradesRes.success && tradesRes.data) {
        const trades = Array.isArray(tradesRes.data)
          ? tradesRes.data
          : (tradesRes.data as { trades: Trade[] }).trades || [];

        if (trades.length > 0) {
          const chartPoints: ChartDataPoint[] = trades.map((trade: Trade) => ({
            time: new Date(trade.created_at).toLocaleTimeString(),
            timestamp: new Date(trade.created_at).getTime(),
            price: trade.price,
            volume: trade.quantity,
          }));
          setChartData(chartPoints.reverse());
          console.log(
            "📈 Chart data loaded from trades:",
            chartPoints.length,
            "points"
          );
        }
      }
    } catch (error) {
      console.error("Failed to load market data:", error);
      // Don't show error to user - this is background data loading
    }
  };

  const connectSSE = useCallback(() => {
    if (!milestoneId) return;

    try {
      const currentEventSource = sseRef.current;
      if (currentEventSource) {
        currentEventSource.close();
      }

      console.log(`Attempting SSE connection to milestone ${milestoneId}`);

      const eventSource = new EventSource(
        `http://localhost:8080/api/v1/milestones/${milestoneId}/stream`
      );
      sseRef.current = eventSource;

      eventSource.onopen = () => {
        setIsSSEConnected(true);
        console.log("SSE connected successfully");
      };

      eventSource.onmessage = (event) => {
        try {
          const message: SSEMessage = JSON.parse(event.data);
          console.log("SSE message received:", message);
          handleSSEMessage(message);
        } catch (error) {
          console.error("SSE message parsing error:", error);
        }
      };

      eventSource.onerror = (error) => {
        console.error("SSE connection error:", error);
        setIsSSEConnected(false);

        // Close the current connection
        eventSource.close();

        // Load initial data even if SSE fails
        if (milestone) {
          loadMarketData();
        }

        // Don't auto-reconnect aggressively, just show disconnected state
        // The user can still use the app without real-time updates
      };
    } catch (error) {
      console.error("SSE connection setup error:", error);
      setIsSSEConnected(false);
    }
  }, [milestoneId]);

  const handleSSEMessage = (message: SSEMessage) => {
    switch (message.type) {
      case "connection":
        console.log(
          `SSE connected to milestone ${message.milestone_id}, status: ${message.status}`
        );
        break;
      case "ping":
        console.log(`SSE ping received for milestone ${message.milestone_id}`);
        // Just keep the connection alive, no action needed
        break;
      case "error":
        console.error(`SSE error: ${message.message}`);
        setIsSSEConnected(false);
        break;
      case "market_update":
        // Reload order book
        loadMarketData();
        break;
      case "trade":
        // Handle real-time trade updates
        if (message.data) {
          const tradeData = message.data as {
            trade_id: number;
            option_id: string;
            buyer_id: number;
            seller_id: number;
            quantity: number;
            price: number;
            total_amount: number;
            timestamp: number;
          };

          // Update recent trades immediately
          const newTrade: Trade = {
            id: tradeData.trade_id,
            project_id: project?.id || 0,
            milestone_id: milestone?.id || 0,
            option_id: tradeData.option_id,
            buyer_id: tradeData.buyer_id,
            seller_id: tradeData.seller_id,
            buy_order_id: 0,
            sell_order_id: 0,
            quantity: tradeData.quantity,
            price: tradeData.price,
            total_amount: tradeData.total_amount,
            created_at: new Date(tradeData.timestamp * 1000).toISOString(),
            buyer_fee: 0,
            seller_fee: 0,
          };

          // Only show trades for the currently selected option
          if (tradeData.option_id === selectedOption) {
            setRecentTrades((prev) => [newTrade, ...prev.slice(0, 19)]);

            // Update chart data
            const newPoint: ChartDataPoint = {
              time: new Date(newTrade.created_at).toLocaleTimeString(),
              timestamp: new Date(newTrade.created_at).getTime(),
              price: newTrade.price,
              volume: newTrade.quantity,
            };
            setChartData((prev) => [...prev, newPoint].slice(-50));
          }

          console.log(
            `📈 Real-time trade: ${tradeData.quantity}@${(
              tradeData.price * 100
            ).toFixed(0)}¢ for ${tradeData.option_id}`
          );
        }
        break;
      case "orderbook_update":
        // Handle real-time order book updates
        if (message.data) {
          const orderBookData = message.data as {
            milestone_id: number;
            option_id: string;
            buy_orders: Array<{ price: number; quantity: number }>;
            sell_orders: Array<{ price: number; quantity: number }>;
          };

          // Only update if it's for the currently selected option
          if (orderBookData.option_id === selectedOption) {
            setOrderBook({
              milestone_id: orderBookData.milestone_id,
              option_id: orderBookData.option_id,
              bids: orderBookData.buy_orders.map((order) => ({
                price: order.price,
                quantity: order.quantity,
                orders: 1,
              })),
              asks: orderBookData.sell_orders.map((order) => ({
                price: order.price,
                quantity: order.quantity,
                orders: 1,
              })),
              spread: 0,
              last_price: 0,
              volume_24h: 0,
              timestamp: new Date().toISOString(),
            });

            console.log(
              `📊 Real-time order book update for ${orderBookData.option_id}`
            );
          }
        }
        break;
      case "price_change":
        // Handle real-time price changes
        if (message.data) {
          const priceData = message.data as {
            option: string;
            old_price: number;
            new_price: number;
          };

          // Only update if it's for the currently selected option
          if (priceData.option === selectedOption) {
            setCurrentPrice(priceData.new_price);
            console.log(
              `💰 Price change: ${priceData.option} ${(
                priceData.old_price * 100
              ).toFixed(0)}¢ → ${(priceData.new_price * 100).toFixed(0)}¢`
            );
          }
        }
        break;
      default:
        console.log("Unknown SSE message type:", message.type);
        break;
    }
  };

  const handleSubmitOrder = async (order: {
    side: OrderSide;
    type: OrderType;
    quantity: number;
    price: number; // Always required (0.01-0.99)
  }) => {
    if (!projectId || !milestoneId) return;

    setOrderLoading(true);
    try {
      const orderRequest: CreateOrderRequest = {
        project_id: parseInt(projectId),
        milestone_id: parseInt(milestoneId),
        option_id: selectedOption,
        type: order.type,
        side: order.side,
        quantity: order.quantity,
        price: order.price, // Always provided now (0.01-0.99)
        currency: "USDC", // 🔵 항상 USDC 사용
      };

      const response = await apiClient.createOrder(orderRequest);

      if (response.success) {
        message.success("Order submitted successfully!");
        // Reload market data
        await loadMarketData();
        // Reload wallet
        const walletRes = await apiClient.getUserWallet();
        if (walletRes.success && walletRes.data) {
          setUserWallet(walletRes.data);
        }
      }
    } catch (error) {
      console.error("Order submission failed:", error);
      throw error;
    } finally {
      setOrderLoading(false);
    }
  };

  const handlePriceClick = (price: number) => {
    // Auto-fill price when clicking on order book
    // This would be handled by the TradingPanel component
    console.log("Price clicked:", price);
  };

  if (loading) {
    return (
      <div className="trading-layout">
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            alignItems: "center",
            height: "100vh",
            gap: "16px",
          }}
        >
          <Spin size="large" />
          <Text style={{ color: "var(--text-primary)" }}>
            {!milestone
              ? "Loading milestone data..."
              : !selectedOption
              ? "Setting up options..."
              : "Loading trading data..."}
          </Text>
          {milestone && !isSSEConnected && (
            <Text style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
              Real-time updates unavailable, using cached data
            </Text>
          )}
        </div>
      </div>
    );
  }

  if (!project || !milestone) {
    return (
      <div className="trading-layout">
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            alignItems: "center",
            height: "100vh",
            gap: "16px",
          }}
        >
          <Text style={{ color: "var(--text-primary)", fontSize: "18px" }}>
            {!project ? "Project not found" : "Milestone not found"}
          </Text>
          <button
            onClick={() => navigate(-1)}
            style={{
              padding: "8px 16px",
              background: "var(--blue)",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: "pointer",
            }}
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="trading-layout">
      <ThemeToggle />

      {/* Header */}
      <div className="trading-header">
        <div style={{ display: "flex", alignItems: "center", gap: "16px" }}>
          <button
            onClick={() => navigate(`/project/${projectId}`)}
            style={{
              background: "none",
              border: "none",
              color: "var(--text-primary)",
              cursor: "pointer",
              padding: "4px",
            }}
          >
            <ArrowLeftOutlined />
          </button>

          <div>
            <Breadcrumb>
              <Breadcrumb.Item>{project.title}</Breadcrumb.Item>
              <Breadcrumb.Item>{milestone.title}</Breadcrumb.Item>
            </Breadcrumb>
            <Title
              level={3}
              style={{ margin: 0, color: "var(--text-primary)" }}
            >
              {milestone.title}
            </Title>
          </div>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "16px" }}>
          {/* Option selector */}
          <div style={{ display: "flex", gap: "8px" }}>
            {["success", "fail"].map((option) => (
              <button
                key={option}
                className={`trading-tab ${
                  selectedOption === option ? "active" : ""
                }`}
                onClick={() => {
                  console.log("🎯 Option changed to:", option);
                  setSelectedOption(option);
                  // Load market data for new option
                  loadMarketDataForOption(option);
                }}
                style={{ padding: "8px 16px" }}
              >
                {option}
              </button>
            ))}
          </div>

          {/* Connection status */}
          <Tag
            color={isSSEConnected ? "green" : "orange"}
            style={{ borderRadius: "6px" }}
          >
            {isSSEConnected ? (
              <span>
                <span
                  style={{
                    display: "inline-block",
                    width: "8px",
                    height: "8px",
                    backgroundColor: "#52c41a",
                    borderRadius: "50%",
                    marginRight: "6px",
                  }}
                ></span>
                Live
              </span>
            ) : (
              <span>
                <span
                  style={{
                    display: "inline-block",
                    width: "8px",
                    height: "8px",
                    backgroundColor: "#faad14",
                    borderRadius: "50%",
                    marginRight: "6px",
                  }}
                ></span>
                Offline
              </span>
            )}
          </Tag>
        </div>
      </div>

      {/* Main content */}
      <div className="trading-content">
        {/* Left side - Chart and trades */}
        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          {/* Chart */}
          <div className="chart-container">
            <div className="chart-header">
              <Title level={4} style={{ margin: 0 }}>
                Price Chart - {selectedOption || "Loading..."}
              </Title>
              <Text type="secondary">
                Last:{" "}
                {chartData.length > 0
                  ? `${Math.round(
                      chartData[chartData.length - 1].price * 100
                    )}¢`
                  : "--"}
              </Text>
            </div>

            <div style={{ height: "300px" }}>
              {chartData.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData}>
                    <CartesianGrid
                      strokeDasharray="3 3"
                      stroke="var(--border-color)"
                    />
                    <XAxis
                      dataKey="time"
                      stroke="var(--text-secondary)"
                      fontSize={12}
                    />
                    <YAxis
                      stroke="var(--text-secondary)"
                      fontSize={12}
                      domain={[0, 1]}
                      tickFormatter={(value) => `${Math.round(value * 100)}¢`}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: "var(--bg-secondary)",
                        border: "1px solid var(--border-color)",
                        borderRadius: "8px",
                        color: "var(--text-primary)",
                      }}
                      formatter={(value: number) => [
                        `${Math.round(value * 100)}¢`,
                        "Price",
                      ]}
                    />
                    <Line
                      type="monotone"
                      dataKey="price"
                      stroke="var(--blue)"
                      strokeWidth={2}
                      dot={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div
                  style={{
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "center",
                    height: "100%",
                    color: "var(--text-secondary)",
                  }}
                >
                  No price data available
                </div>
              )}
            </div>
          </div>

          {/* Recent Trades */}
          <div className="orderbook">
            <div
              className="orderbook-header"
              style={{ color: "var(--text-primary)" }}
            >
              Recent Trades
            </div>
            <div className="trades-list">
              {recentTrades.length > 0 ? (
                recentTrades.map((trade, index) => (
                  <div
                    key={index}
                    className="trade-row"
                    style={{ color: "var(--text-primary)" }}
                  >
                    <span
                      className={`trade-price ${
                        trade.price > 50 ? "buy" : "sell"
                      }`}
                    >
                      {trade.price.toFixed(2)}
                    </span>
                    <span style={{ color: "var(--text-primary)" }}>
                      {trade.quantity.toLocaleString()}
                    </span>
                    <span style={{ color: "var(--text-secondary)" }}>
                      {new Date(trade.created_at).toLocaleTimeString()}
                    </span>
                  </div>
                ))
              ) : (
                <div
                  style={{
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "center",
                    height: "100px",
                    color: "var(--text-secondary)",
                    fontSize: "12px",
                  }}
                >
                  No recent trades
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right side - Order book and trading */}
        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          {/* Order Book */}
          <OrderBook
            orderBook={orderBook}
            onPriceClick={handlePriceClick}
            loading={false}
          />

          {/* Trading Panel */}
          <TradingPanel
            selectedOption={selectedOption}
            onSubmitOrder={handleSubmitOrder}
            loading={orderLoading}
            userBalance={userWallet?.usdc_balance || 0}
            currentPrice={currentPrice}
          />

          {/* User Stats */}
          {userWallet && (
            <div className="stat-card">
              <div className="stat-label">🔵 USDC Balance</div>
              <div className="stat-value">
                ${(userWallet.usdc_balance / 100).toFixed(2)}
              </div>
              {userWallet.usdc_locked_balance > 0 && (
                <div
                  style={{
                    fontSize: "12px",
                    color: "var(--text-secondary)",
                    marginTop: "4px",
                  }}
                >
                  Locked: ${(userWallet.usdc_locked_balance / 100).toFixed(2)}
                </div>
              )}
              {userWallet.blueprint_balance > 0 && (
                <div
                  style={{
                    fontSize: "11px",
                    color: "var(--blue)",
                    marginTop: "4px",
                  }}
                >
                  🟦 BLUEPRINT: {userWallet.blueprint_balance.toLocaleString()}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default PolymarketTradingPage;
