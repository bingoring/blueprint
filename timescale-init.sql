-- TimescaleDB Extension 활성화
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 🔥 고성능 시계열 테이블들

-- 가격 히스토리 테이블 (초단위 가격 데이터)
CREATE TABLE price_ticks (
    time TIMESTAMPTZ NOT NULL,
    milestone_id INTEGER NOT NULL,
    option_id TEXT NOT NULL,
    price DECIMAL(18, 8) NOT NULL,
    volume_24h BIGINT DEFAULT 0,
    trades_count INTEGER DEFAULT 0,
    bid_price DECIMAL(18, 8),
    ask_price DECIMAL(18, 8),
    spread DECIMAL(18, 8),
    market_cap DECIMAL(18, 2)
);

-- 하이퍼테이블 생성 (시간 기준 파티셔닝)
SELECT create_hypertable('price_ticks', 'time');

-- 인덱스 생성
CREATE INDEX idx_price_ticks_milestone_option ON price_ticks (milestone_id, option_id, time DESC);
CREATE INDEX idx_price_ticks_time ON price_ticks (time DESC);

-- 거래 데이터 테이블 (개별 거래 기록)
CREATE TABLE trade_events (
    time TIMESTAMPTZ NOT NULL,
    trade_id BIGINT NOT NULL,
    milestone_id INTEGER NOT NULL,
    option_id TEXT NOT NULL,
    buyer_id INTEGER NOT NULL,
    seller_id INTEGER NOT NULL,
    quantity BIGINT NOT NULL,
    price DECIMAL(18, 8) NOT NULL,
    total_amount BIGINT NOT NULL,
    trade_type TEXT NOT NULL, -- 'market', 'limit', 'amm'
    side TEXT NOT NULL, -- 'buy', 'sell'
    fees BIGINT DEFAULT 0
);

-- 하이퍼테이블 생성
SELECT create_hypertable('trade_events', 'time');

-- 인덱스 생성
CREATE INDEX idx_trade_events_milestone_option ON trade_events (milestone_id, option_id, time DESC);
CREATE INDEX idx_trade_events_users ON trade_events (buyer_id, seller_id, time DESC);
CREATE INDEX idx_trade_events_time ON trade_events (time DESC);

-- 시장 통계 테이블 (분단위 집계)
CREATE TABLE market_stats (
    time TIMESTAMPTZ NOT NULL,
    milestone_id INTEGER NOT NULL,
    option_id TEXT NOT NULL,
    open_price DECIMAL(18, 8),
    high_price DECIMAL(18, 8),
    low_price DECIMAL(18, 8),
    close_price DECIMAL(18, 8),
    volume BIGINT DEFAULT 0,
    trades_count INTEGER DEFAULT 0,
    unique_traders INTEGER DEFAULT 0,
    avg_trade_size DECIMAL(18, 8),
    price_volatility DECIMAL(18, 8),
    spread_avg DECIMAL(18, 8)
);

-- 하이퍼테이블 생성
SELECT create_hypertable('market_stats', 'time');

-- 인덱스 생성
CREATE INDEX idx_market_stats_milestone_option ON market_stats (milestone_id, option_id, time DESC);

-- 유저 활동 로그 테이블
CREATE TABLE user_activity (
    time TIMESTAMPTZ NOT NULL,
    user_id INTEGER NOT NULL,
    activity_type TEXT NOT NULL, -- 'login', 'trade', 'view', 'order_place', 'order_cancel'
    milestone_id INTEGER,
    option_id TEXT,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT
);

-- 하이퍼테이블 생성
SELECT create_hypertable('user_activity', 'time');

-- 인덱스 생성
CREATE INDEX idx_user_activity_user_time ON user_activity (user_id, time DESC);
CREATE INDEX idx_user_activity_type ON user_activity (activity_type, time DESC);

-- 📊 실시간 분석을 위한 연속 집계 (Continuous Aggregates)

-- 1분 단위 가격 집계
CREATE MATERIALIZED VIEW price_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS bucket,
    milestone_id,
    option_id,
    first(price, time) AS open_price,
    max(price) AS high_price,
    min(price) AS low_price,
    last(price, time) AS close_price,
    sum(volume_24h) AS volume,
    count(*) AS ticks_count
FROM price_ticks
GROUP BY bucket, milestone_id, option_id;

-- 1시간 단위 가격 집계
CREATE MATERIALIZED VIEW price_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    milestone_id,
    option_id,
    first(price, time) AS open_price,
    max(price) AS high_price,
    min(price) AS low_price,
    last(price, time) AS close_price,
    sum(volume_24h) AS volume,
    count(*) AS ticks_count
FROM price_ticks
GROUP BY bucket, milestone_id, option_id;

-- 1일 단위 가격 집계
CREATE MATERIALIZED VIEW price_1d
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    milestone_id,
    option_id,
    first(price, time) AS open_price,
    max(price) AS high_price,
    min(price) AS low_price,
    last(price, time) AS close_price,
    sum(volume_24h) AS volume,
    count(*) AS ticks_count
FROM price_ticks
GROUP BY bucket, milestone_id, option_id;

-- 거래 통계 연속 집계 (5분 단위)
CREATE MATERIALIZED VIEW trade_stats_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    milestone_id,
    option_id,
    count(*) AS trades_count,
    sum(quantity) AS total_volume,
    avg(price) AS avg_price,
    stddev(price) AS price_volatility,
    count(DISTINCT buyer_id) + count(DISTINCT seller_id) AS unique_traders
FROM trade_events
GROUP BY bucket, milestone_id, option_id;

-- 🔄 자동 정책 설정

-- 오래된 원시 데이터 압축 (7일 후)
SELECT add_compression_policy('price_ticks', INTERVAL '7 days');
SELECT add_compression_policy('trade_events', INTERVAL '7 days');
SELECT add_compression_policy('user_activity', INTERVAL '7 days');

-- 오래된 데이터 삭제 정책 (1년 후)
SELECT add_retention_policy('price_ticks', INTERVAL '1 year');
SELECT add_retention_policy('trade_events', INTERVAL '1 year');
SELECT add_retention_policy('user_activity', INTERVAL '6 months');

-- 연속 집계 새로고침 정책
SELECT add_continuous_aggregate_policy('price_1m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

SELECT add_continuous_aggregate_policy('price_1h',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

SELECT add_continuous_aggregate_policy('price_1d',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');

SELECT add_continuous_aggregate_policy('trade_stats_5m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes');

-- 📈 실시간 분석용 뷰들

-- 현재 가격 뷰 (가장 최근 가격)
CREATE VIEW current_prices AS
SELECT DISTINCT ON (milestone_id, option_id)
    milestone_id,
    option_id,
    price AS current_price,
    time AS last_updated,
    volume_24h,
    trades_count,
    spread
FROM price_ticks
ORDER BY milestone_id, option_id, time DESC;

-- 24시간 통계 뷰
CREATE VIEW daily_stats AS
SELECT
    pt.milestone_id,
    pt.option_id,
    first(pt.price, pt.time) AS open_24h,
    max(pt.price) AS high_24h,
    min(pt.price) AS low_24h,
    last(pt.price, pt.time) AS current_price,
    sum(te.quantity) AS volume_24h,
    count(te.*) AS trades_24h,
    (last(pt.price, pt.time) - first(pt.price, pt.time)) / first(pt.price, pt.time) * 100 AS change_24h_pct
FROM price_ticks pt
LEFT JOIN trade_events te ON pt.milestone_id = te.milestone_id
    AND pt.option_id = te.option_id
    AND te.time >= now() - INTERVAL '24 hours'
WHERE pt.time >= now() - INTERVAL '24 hours'
GROUP BY pt.milestone_id, pt.option_id;

-- 상위 거래량 마켓 뷰
CREATE VIEW top_markets AS
SELECT
    milestone_id,
    option_id,
    sum(quantity) AS volume_24h,
    count(*) AS trades_24h,
    count(DISTINCT buyer_id) + count(DISTINCT seller_id) AS unique_traders_24h,
    avg(price) AS avg_price_24h
FROM trade_events
WHERE time >= now() - INTERVAL '24 hours'
GROUP BY milestone_id, option_id
ORDER BY volume_24h DESC
LIMIT 100;

-- 🚀 성능 최적화를 위한 추가 인덱스

-- 복합 인덱스 (조회 성능 향상)
CREATE INDEX idx_price_ticks_recent ON price_ticks (milestone_id, option_id, time DESC)
WHERE time >= now() - INTERVAL '7 days';

CREATE INDEX idx_trade_events_recent ON trade_events (milestone_id, option_id, time DESC)
WHERE time >= now() - INTERVAL '7 days';

-- 부분 인덱스 (활성 마켓만)
CREATE INDEX idx_price_ticks_active ON price_ticks (time DESC)
WHERE time >= now() - INTERVAL '1 day';

-- BRIN 인덱스 (시계열 데이터 최적화)
CREATE INDEX idx_price_ticks_time_brin ON price_ticks USING BRIN (time);
CREATE INDEX idx_trade_events_time_brin ON trade_events USING BRIN (time);

-- 🎯 유용한 함수들

-- 가격 변화율 계산 함수
CREATE OR REPLACE FUNCTION calculate_price_change(
    p_milestone_id INTEGER,
    p_option_id TEXT,
    p_interval INTERVAL DEFAULT '24 hours'
) RETURNS DECIMAL AS $$
DECLARE
    start_price DECIMAL;
    end_price DECIMAL;
BEGIN
    SELECT price INTO start_price
    FROM price_ticks
    WHERE milestone_id = p_milestone_id
      AND option_id = p_option_id
      AND time <= now() - p_interval
    ORDER BY time DESC
    LIMIT 1;

    SELECT price INTO end_price
    FROM price_ticks
    WHERE milestone_id = p_milestone_id
      AND option_id = p_option_id
    ORDER BY time DESC
    LIMIT 1;

    IF start_price IS NULL OR end_price IS NULL THEN
        RETURN 0;
    END IF;

    RETURN (end_price - start_price) / start_price * 100;
END;
$$ LANGUAGE plpgsql;

COMMENT ON DATABASE timeseries IS 'TimescaleDB for Blueprint trading analytics and time-series data';

-- 초기 샘플 데이터 (테스트용)
-- INSERT INTO price_ticks (time, milestone_id, option_id, price, volume_24h)
-- VALUES (now(), 1, 'success', 0.65, 1000);

NOTICE 'TimescaleDB initialized successfully! 🚀';
NOTICE 'Available views: current_prices, daily_stats, top_markets';
NOTICE 'Available aggregates: price_1m, price_1h, price_1d, trade_stats_5m';
