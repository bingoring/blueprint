import { Spin } from "antd";
import React from "react";
import type { OrderBook as OrderBookType } from "../types";

interface OrderBookProps {
  orderBook: OrderBookType | null;
  loading?: boolean;
  onPriceClick?: (price: number) => void;
}

const OrderBook: React.FC<OrderBookProps> = ({
  orderBook,
  loading = false,
  onPriceClick,
}) => {
  if (loading) {
    return (
      <div className="orderbook">
        <div className="orderbook-header">Order Book</div>
        <div
          style={{
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            height: "200px",
          }}
        >
          <Spin />
        </div>
      </div>
    );
  }

  if (!orderBook) {
    return (
      <div className="orderbook">
        <div className="orderbook-header">Order Book</div>
        <div
          style={{
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            height: "200px",
            color: "var(--text-secondary)",
          }}
        >
          No order book data
        </div>
      </div>
    );
  }

  const asks = orderBook.asks || [];
  const bids = orderBook.bids || [];
  const spread =
    asks.length > 0 && bids.length > 0 ? asks[0].price - bids[0].price : 0;

  const formatNumber = (num: number, decimals = 2) => {
    return num.toFixed(decimals);
  };

  const handlePriceClick = (price: number) => {
    if (onPriceClick) {
      onPriceClick(price);
    }
  };

  return (
    <div className="orderbook">
      <div className="orderbook-header">
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr 1fr",
            fontSize: "11px",
            color: "var(--text-secondary)",
          }}
        >
          <span>Price</span>
          <span style={{ textAlign: "center" }}>Quantity</span>
          <span style={{ textAlign: "right" }}>Total</span>
        </div>
      </div>

      <div className="orderbook-content">
        {/* Asks (매도 주문) - 높은 가격부터 */}
        <div className="orderbook-section">
          {asks
            .slice(0, 10)
            .reverse()
            .map((level, index) => {
              const total = level.price * level.quantity;
              return (
                <div
                  key={`ask-${index}`}
                  className="orderbook-row ask-row"
                  onClick={() => handlePriceClick(level.price)}
                  style={{ cursor: onPriceClick ? "pointer" : "default" }}
                >
                  <span>{formatNumber(level.price)}</span>
                  <span style={{ textAlign: "center" }}>
                    {level.quantity.toLocaleString()}
                  </span>
                  <span style={{ textAlign: "right" }}>
                    {formatNumber(total)}
                  </span>
                </div>
              );
            })}
        </div>

        {/* Spread */}
        {spread > 0 && (
          <div className="spread-row">
            Spread: {formatNumber(spread)} (
            {((spread / ((asks[0]?.price + bids[0]?.price) / 2)) * 100).toFixed(
              2
            )}
            %)
          </div>
        )}

        {/* Bids (매수 주문) - 높은 가격부터 */}
        <div className="orderbook-section">
          {bids.slice(0, 10).map((level, index) => {
            const total = level.price * level.quantity;
            return (
              <div
                key={`bid-${index}`}
                className="orderbook-row bid-row"
                onClick={() => handlePriceClick(level.price)}
                style={{ cursor: onPriceClick ? "pointer" : "default" }}
              >
                <span>{formatNumber(level.price)}</span>
                <span style={{ textAlign: "center" }}>
                  {level.quantity.toLocaleString()}
                </span>
                <span style={{ textAlign: "right" }}>
                  {formatNumber(total)}
                </span>
              </div>
            );
          })}
        </div>

        {/* Empty states */}
        {asks.length === 0 && bids.length === 0 && (
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
            No orders available
          </div>
        )}
      </div>
    </div>
  );
};

export default OrderBook;
