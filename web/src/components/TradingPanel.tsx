import { Button, message } from "antd";
import React, { useState } from "react";
import type { OrderSide, OrderType } from "../types";

interface TradingPanelProps {
  selectedOption: string;
  onSubmitOrder: (order: {
    side: OrderSide;
    type: OrderType;
    quantity: number;
    price: number; // Always required (0.01-0.99)
  }) => Promise<void>;
  loading?: boolean;
  userBalance?: number;
}

const TradingPanel: React.FC<TradingPanelProps> = ({
  selectedOption,
  onSubmitOrder,
  loading = false,
  userBalance = 0,
}) => {
  const [activeTab, setActiveTab] = useState<OrderSide>("buy");
  const [orderType, setOrderType] = useState<OrderType>("limit");
  const [quantity, setQuantity] = useState<number>(10);
  const [price, setPrice] = useState<number>(0.5); // 0.01-0.99 범위의 확률값

  const handleSubmit = async () => {
    if (!quantity || quantity <= 0) {
      message.error("Please enter a valid quantity");
      return;
    }

    if (orderType === "limit" && (!price || price < 0.01 || price > 0.99)) {
      message.error("Please enter a valid price between 0.01 and 0.99");
      return;
    }

    const effectivePrice = orderType === "limit" ? price : 0.5; // Market orders use mid-price
    const totalCost = quantity * effectivePrice * 100; // Convert to USDC cents (0.5 probability = 50 cents)

    if (activeTab === "buy" && totalCost > userBalance) {
      message.error("Insufficient balance");
      return;
    }

    try {
      await onSubmitOrder({
        side: activeTab,
        type: orderType,
        quantity,
        price: orderType === "limit" ? price : 0.5, // Market orders use mid-price
      });

      // Reset form on success
      setQuantity(10);
      if (orderType === "limit") {
        setPrice(0.5);
      }

      message.success(
        `${activeTab === "buy" ? "Buy" : "Sell"} order submitted successfully`
      );
    } catch (error) {
      console.error("Order submission failed:", error);
      message.error("Failed to submit order");
    }
  };

  const calculateTotal = () => {
    if (orderType === "market") {
      return quantity * 0.5 * 100; // Market price estimate: 0.5 probability = 50 USDC cents
    }
    return quantity * price * 100; // Convert probability to USDC cents
  };

  return (
    <div className="trading-panel">
      {/* Trade Type Tabs */}
      <div className="trading-tabs">
        <button
          className={`trading-tab buy ${activeTab === "buy" ? "active" : ""}`}
          onClick={() => setActiveTab("buy")}
        >
          Buy {selectedOption}
        </button>
        <button
          className={`trading-tab sell ${activeTab === "sell" ? "active" : ""}`}
          onClick={() => setActiveTab("sell")}
        >
          Sell {selectedOption}
        </button>
      </div>

      <div className="trading-form">
        {/* Order Type */}
        <div className="form-group">
          <div className="form-label">Order Type</div>
          <div style={{ display: "flex", gap: "8px" }}>
            <button
              className={`trading-tab ${orderType === "limit" ? "active" : ""}`}
              style={{ flex: 1, fontSize: "12px", padding: "6px" }}
              onClick={() => setOrderType("limit")}
            >
              Limit
            </button>
            <button
              className={`trading-tab ${
                orderType === "market" ? "active" : ""
              }`}
              style={{ flex: 1, fontSize: "12px", padding: "6px" }}
              onClick={() => setOrderType("market")}
            >
              Market
            </button>
          </div>
        </div>

        {/* Price (for limit orders) */}
        {orderType === "limit" && (
          <div className="form-group">
            <label className="form-label">Price (Probability)</label>
            <input
              type="number"
              className="form-input"
              value={price}
              onChange={(e) => setPrice(Number(e.target.value))}
              placeholder="0.50"
              step="0.01"
              min="0.01"
              max="0.99"
            />
            <div
              style={{
                fontSize: "11px",
                color: "var(--text-secondary)",
                marginTop: "4px",
              }}
            >
              Range: 0.01 - 0.99 (probability)
            </div>
          </div>
        )}

        {/* Quantity */}
        <div className="form-group">
          <label className="form-label">Quantity</label>
          <input
            type="number"
            className="form-input"
            value={quantity}
            onChange={(e) => setQuantity(Number(e.target.value))}
            placeholder="0"
            min="1"
          />
        </div>

        {/* Quick Amount Buttons */}
        <div className="form-group">
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(3, 1fr)",
              gap: "8px",
            }}
          >
            {[10, 50, 100].map((amount) => (
              <button
                key={amount}
                className="trading-tab"
                style={{ fontSize: "12px", padding: "6px" }}
                onClick={() => setQuantity(amount)}
              >
                {amount}
              </button>
            ))}
          </div>
        </div>

        {/* Total */}
        <div className="form-group">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "8px",
            }}
          >
            <span className="form-label">Total Cost</span>
            <span style={{ fontSize: "14px", fontWeight: "600" }}>
              ${(calculateTotal() / 100).toFixed(2)}
            </span>
          </div>
        </div>

        {/* Balance */}
        <div className="form-group">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "16px",
            }}
          >
            <span className="form-label">Available Balance</span>
            <span style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
              {userBalance.toLocaleString()} pts
            </span>
          </div>
        </div>

        {/* Submit Button */}
        <Button
          className={`btn-trade ${activeTab}`}
          onClick={handleSubmit}
          loading={loading}
          disabled={loading || !quantity || (orderType === "limit" && !price)}
          style={{
            width: "100%",
            height: "40px",
            border: "none",
            borderRadius: "8px",
            fontWeight: "600",
            fontSize: "14px",
          }}
        >
          {loading
            ? "Submitting..."
            : `${activeTab === "buy" ? "Buy" : "Sell"} ${selectedOption}`}
        </Button>
      </div>
    </div>
  );
};

export default TradingPanel;
