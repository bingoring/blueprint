import { Button, message } from "antd";
import React, { useEffect, useState } from "react";
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
  currentPrice?: number; // Current market price in probability (0.01-0.99)
}

const TradingPanel: React.FC<TradingPanelProps> = ({
  selectedOption,
  onSubmitOrder,
  loading = false,
  userBalance = 0,
  currentPrice = 0.5, // Default fallback
}) => {
  const [activeTab, setActiveTab] = useState<OrderSide>("buy");
  const [orderType, setOrderType] = useState<OrderType>("limit");
  const [quantity, setQuantity] = useState<number>(10);

  // Calculate default price based on current market price
  const defaultPriceCents = Math.round(currentPrice * 100);
  const [price, setPrice] = useState<number>(defaultPriceCents);

  // Update price when currentPrice changes (for different options)
  useEffect(() => {
    const newDefaultPrice = Math.round(currentPrice * 100);
    setPrice(newDefaultPrice);
    console.log("💰 Price updated to market price:", newDefaultPrice, "¢");
  }, [currentPrice]);

  const handleSubmit = async () => {
    if (!quantity || quantity <= 0) {
      message.error("Please enter a valid quantity");
      return;
    }

    if (orderType === "limit" && (!price || price < 1 || price > 99)) {
      message.error("Please enter a valid price between 1¢ and 99¢");
      return;
    }

    const totalCost =
      quantity * (orderType === "limit" ? price : defaultPriceCents); // Direct cents calculation

    if (activeTab === "buy" && totalCost > userBalance) {
      message.error("Insufficient balance");
      return;
    }

    try {
      await onSubmitOrder({
        side: activeTab,
        type: orderType,
        quantity,
        price: orderType === "limit" ? price / 100 : 0.5, // Convert cents to probability for API
      });

      // Reset form on success
      setQuantity(10);
      if (orderType === "limit") {
        setPrice(defaultPriceCents); // Reset to current market price
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
      return quantity * defaultPriceCents; // Market price estimate based on current price
    }
    return quantity * price; // Direct cents calculation
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
            <label className="form-label">Price</label>
            <div style={{ position: "relative" }}>
              <input
                type="number"
                className="form-input"
                value={price}
                onChange={(e) => setPrice(Math.round(Number(e.target.value)))} // 반올림 처리
                placeholder={defaultPriceCents.toString()}
                step="1"
                min="1"
                max="99"
                style={{ paddingRight: "20px" }}
              />
              <span
                style={{
                  position: "absolute",
                  right: "8px",
                  top: "50%",
                  transform: "translateY(-50%)",
                  color: "var(--text-secondary)",
                  fontSize: "12px",
                  pointerEvents: "none",
                }}
              >
                ¢
              </span>
            </div>
            <div
              style={{
                fontSize: "11px",
                color: "var(--text-secondary)",
                marginTop: "4px",
              }}
            >
              Range: 1¢ - 99¢ (probability in cents)
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
